# План: MCP-сервер внутри TUI (krw)

## 1. Анализ требований — правки и решения

Выбрано: **HTTP-сервер внутри живого TUI-процесса** + набор инструментов
(list/add/delete кластеров, show/select cluster·topic·partition, send) с ответом
`ok/ошибка` по каждой операции + обязательная партиция.

### Правки к требованиям

1. **Нет отдельного бинарника.** stdio и живой BubbleTea несовместимы (оба захватывают
   stdin/stdout). Итог: **один процесс** — TUI рисуется, а в goroutine крутится MCP
   HTTP-сервер. Модуль `internal/mcp` — «отдельный», но внутри того же бинарника.
2. **Kafka-коннектор нельзя использовать как есть** — в нём три блокера (см. Шаг 0).
   Без правок MCP-сервер будет либо умирать, либо душить TUI.
3. **`database.db` удаляется при каждом старте TUI** (`os.Remove` в `CreateOrGetDb`,
   помечен `//TODO delete later`). Это data-loss баг: добавленные кластеры исчезают
   после рестарта. Для MCP критично — чиним.
4. **`localstorage.SaveNewKafkaCluster` застаблена** (`return errors.New("TEST errro")`) —
   `add_cluster` работать не будет. Чиним.
5. **«Показать кластеры (равны логину пользователя)»** — трактуется как: показать все
   кластеры, у каждого поле `username`; плюс опциональный аргумент `username` для фильтра.
6. **Партиция обязательна** — сигнатура `Send(cluster, topic, partition, text)` сохраняется.
7. **Кэширование kafka-соединений** — `Send` переводится с `DialLeader` на `kafka.Writer`
   (пул соединений на кластер, явная партиция поддерживается), dialer кэшируется по
   кластеру, добавляется `Close()` и раскомментируется `closer.Add("кафка")`. Детали —
   Шаг 0.4.

### Альтернативы, которые рассмотрены и отклонены
- **stdio headless (`krw --mcp`)** — рабочий вариант, но это отдельный процесс без
  живого TUI; `show_selection`/`select_*` жили бы только внутри MCP-сессии. Не нужен
  после выбора «HTTP в живом TUI». Можно добавить позже (тот же `mcp.Server`, другой
  транспорт).
- **SSE-транспорт вместо Streamable HTTP** — в SDK есть оба (`NewSSEHTTPHandler` /
  `NewStreamableHTTPHandler`). Берём Streamable HTTP (современный протокол), SSE как
  fallback, если у ИИ-клиента проблемы.
- **Кэш leader-`Conn` по ключу (cluster,topic,partition) (Вариант B)** — отклонён:
  смена лидера (`NotLeaderForPartition`) требует ручного re-dial и учёта FD;
  `kafka.Writer` делает это автоматически (см. Шаг 0.4).

### Советы
- Замкнуть сервер на `127.0.0.1`, опциональный bearer-токен (`KRW_MCP_TOKEN`), т.к.
  из-за plaintext-паролей в БД любой, кто достучится до порта, сможет писать в кластеры.
- **Не логировать** тела сообщений и пароли (сейчас `Update` логирует каждое `tea.Msg`
  целиком — в новых MCP-cases логировать только метаданные).
- Всё состояние держать в event-loop BubbleTea (однопоточная запись), MCP-хендлеры
  общаются с моделью через `program.Send` + reply-канал — отсутствие data race.

---

## 2. Архитектура

```
┌─────────────────────────────────────────────┐
│ процесс krw                                │
│                                             │
│  [BubbleTea Program]  ←tea.Msg←─────┐      │
│   Model (selection, UI)              │      │
│   Update() журнал событий            │ Send │
│                                     │      │
│  [internal/mcp]  ──deps.Deps───────▶│      │
│   mcp.Server (StreamableHTTP        │      │
│      на :18810, goroutine)          │      │
│      ▲  tool call                   │      │
│  ИИ-клиент ← http://127.0.0.1:18810│      │
└─────────────────────────────────────────────┘
```

Ключевые принципы:
- **MCP-хендлер (goroutine HTTP) никогда не читает/не пишет состояние модели
  напрямую.** Для операций над состоянием он посылает
  `contracts.MCP*Request{..., Reply chan MCPResult}` через `program.Send` и блокируется
  на `Reply`. Пишет ответ только `Model.Update` (единственный писатель).
- Быстрые операции над БД (`list/add/delete cluster`) выполняются прямо в хендлере —
  БД защищена собственным мьютексом; после add/delete хендлер шлёт уведомление
  `ClusterStoreChangedEvent`, чтобы TUI обновил список.
- Зависимости MCP-пакета охвачены интерфейсом `mcp.Deps`, чтобы `internal/mcp` не
  импортировал `app`/`windows` (без циклических импортов).

---

## 3. Пошаговая спецификация изменений

### Шаг 0. Блокеры (без них MCP не работает)

**0.1 `internal/database/database.go:58`** — убрать `os.Remove(filePath)` (баг потери
данных). Добавить устойчивость к конкурентному доступу при открытии:
```go
sql.Open("sqlite3", databaseUrl+"?_journal_mode=WAL&_busy_timeout=5000")
```

**0.2 `internal/kafka/connection/connection.go:153-182` (`Send`)** — заменить
`log.Fatal(...)` на `return err` (иначе при Authorization-ошибке умрёт весь процесс:
TUI + MCP). Добавить `defer conn.Close()` (сейчас conn-на-успех течёт).

**0.3 `internal/localstorage/localstorage.go:47-55`** — развернуть стабы:
```go
func (ls *LocalStorage) SaveNewKafkaCluster(...) error { return ls.db.SaveNewKafkaCluster(...) }
```
`SaveEditedKafkaCluster` тоже развернуть **и** починить SQL в `database.go:281-298`
(лишняя `)` после `SET list` — сейчас `UPDATE` сломан).

**0.4 Кэширование kafka-соединений — рефакторинг `KafkaConnectorProvider` (Вариант C)**

Зачем (это тоже блокер скорости/корректности):
- `Send` делает `DialLeader` — новое TCP+TLS-соединение на **каждое** сообщение.
- `createDialer` перечитывает PEM-сертификат при каждой операции (пересборка TLS).
- Кэш `connections` в `GetTopicsByClusterName` (`connection.go:42-49`) **мёртвый**: значение
  из кэша сразу перезаписывается новым `createConnection(...)`; старые соединения текут.
- Карты `topics/partitions/connections` без мьютекса, а `Send`/`loadTopics`/`loadPartitions`
  уже вызываются из tea.Cmd-горутин (`model.go:543-554, 281-292, 1108-1118`) → data race
  есть уже сейчас; MCP-хендлеры (HTTP-goroutine) сделают его гарантированным.

Изменения — **только внутри** `internal/kafka/connection/connection.go`: интерфейс
`KafkaProvider` (`model.go:79-84`) и сигнатура `Send(cluster, topic, partition, text)`
не меняются:
```go
type KafkaConnectorProvider struct {
    mu         sync.Mutex               // защита карт: Send конкурентен из tea.Cmd + MCP
    dialers    map[string]*kafka.Dialer // по Title кластера: SASL+TLS собираются 1 раз
    writers    map[string]*kafka.Writer // по Title кластера: пул соединений на кластер
    topics     map[string][]string
    partitions map[string]map[string][]int
}
```
1. `dialerFor(cluster) (*kafka.Dialer, error)` — ленивое создание; ошибки чтения
   сертификата **возвращать**, а не `panic`.
2. `Send` → `writerFor(cluster).WriteMessages(ctx, kafka.Message{Topic, Partition: part,
   Value: []byte(text)})`. `Partition >= 0` — балансер не участвует, пишем строго в
   указанную партицию (требование «партиция обязательна» сохраняется). `log.Fatal` из
   0.2 и `DialLeader`-код исчезают; ошибки → `return err`.
3. `GetTopicsByClusterName` — перейти на `dialerFor`; мёртвый кэш `connections` убрать
   (или починить + `defer conn.Close()` для metadata-соединения). Результат уже
   кэшируется в `topics`/`partitions` — один сетевой поход на кластер за сессию.
4. `GetPartitionsByClusterNameAndTopic` — не меняется (чистый кэш).
5. `Read` — per-call `kafka.NewReader` остаётся (offset-специфичен), но на общем `dialerFor`.
6. `Close() error` — закрыть все `kafka.Writer`; в `internal/app/di.go:210-212`
   раскомментировать `closer.Add("кафка", ...)`.

Отклонённый **Вариант B** (кэш leader-`Conn` по ключу cluster+topic+partition): ручной
re-dial при смене лидера (`NotLeaderForPartition`) и учёт FD — всё это `kafka.Writer`
уже делает сам.

### Шаг 1. Конфиг — `internal/config/config.go`
Добавить поля:
```go
McpEnabled boolean
McpAddr    string // default "127.0.0.1:18810"
McpAuthToken string
```
Экспонировать геттеры. Чтение env в `AppConfig()`: `KRW_MCP` (=1 включает),
`KRW_MCP_ADDR`, `KRW_MCP_TOKEN`. MCP выключен по умолчанию — нулевой оверхед.

### Шаг 2. Контракты — новый `internal/contracts/mcp.go`
- `MCPResult{ OK bool; Error string; Data map[string]any }`
- Запросы (tea.Msg) с reply-каналом:
  - `MCPSelectClusterRequest{Title string; Reply chan MCPResult}`
  - `MCPSelectTopicRequest{Topic string; Reply chan MCPResult}`
  - `MCPSelectPartitionRequest{Partition string; Reply chan MCPResult}`
  - `MCPSendRequest{Value string; Reply chan MCPResult}` (цель — из текущей Selection)
- Уведомления TUI: `ClusterStoreChangedEvent{}`, `MCPServerStartedEvent{Addr string}`.

### Шаг 3. Модель — `internal/windows/model.go`
1. **Безопасное чтение Selection** из MCP-хендлеров: добавить
   `muSelection sync.RWMutex` + геттеры `GetSelectedCluster/Topic/Partition() string`;
   изменение Selection — только через сеттер с блокировкой/в Update.
2. **Reply-слот**: `m.mcpReply chan contracts.MCPResult` (nil = нет активного
   MCP-ожидания; мультиплекс не нужен, т.к. MCP-слой сериализует запросы своим мьютексом).
3. **Рефакторинг без дублирования**: вынести логику send+save из
   `contracts.SendMessageToKafka` (строки 535-569) в метод
   `m.sendAndSaveMessage(cluster, topic, partition, message) tea.Cmd`; оба кейса —
   существующий и `MCPSendRequest` — используют его.
4. **Новые case в Update:**
   - `MCPSendRequest`: `m.mcpReply = msg.Reply`; вернуть `m.sendAndSaveMessage(...)`.
   - Терминалы `Successful`/`ErrorWhenSendMessageToKafka`/`ErrorWhenSaveMessageToHistory`:
     если `m.mcpReply != nil` → записать `MCPResult` (буферизованный chan(1) +
     `select{default:}` чтобы не висеть) и обнулить слот.
   - `MCPSelectClusterRequest`: запомнить reply, поставить Selection, запустить
     `m.loadTopics(cluster)`; **ответить** в `SuccessLoadedKafkaTopicsMsg` /
     `ErrorLoadedKafkaTopicsErrMsg` (эта точка гарантирует, что кэш партиций заполнен —
     его заполняет `GetTopicsByClusterName`).
   - `MCPSelectTopicRequest` / `MCPSelectPartitionRequest`: ответить сразу после
     установки Selection.
   - `ClusterStoreChangedEvent` → `uploadClusters(m)`.
   - `MCPServerStartedEvent` → `m.kafkaClusterList.GetList().NewStatusMessage("MCP: http://"+addr)`
     (статус в списке кластеров; минимальное вмешательство в вёрстку).

### Шаг 4. Новый пакет `internal/mcp/` (отдельный модуль-пакет)
```
internal/mcp/
  server.go   — жизненный цикл HTTP-сервера
  tools.go    — 8 инструментов MCP
  deps.go     — интерфейс mcp.Deps
```
**`deps.go`**
```go
type Deps struct {
    SendProgram     func(msg tea.Msg)                          // p.Send
    GetSelected     func() (cluster, topic, partition string)
    ListClusters    func() ([]ClusterInfo, error)              // без паролей
    GetCluster      func(title string) (*contracts.KafkaCluster, error)
    AddCluster      func(title, username, password, cert, url string) error
    DeleteCluster   func(title string) error
}
```
`Deps` реализует `app` через DI-контейнер + геттеры модели. MCP-пакет не импортирует
`app`/`windows`.

**`server.go`** — `New(deps)`: собрать `mcp.NewServer(...)`, зарегистрировать
инструменты, обернуть в `mcp.NewStreamableHTTPHandler(func(r){return server},
&mcp.StreamableHTTPOptions{})`; `Start()`/`Stop(ctx)`, middleware: `if token != "" &&
header != "Bearer "+token { 401 }`. Зависимость: `github.com/modelcontextprotocol/go-sdk v1.2.0`.

**`tools.go`** — инструменты (имена с префиксом `krw_`):

| Tool | Аргументы | Ответ (JSON text) |
|---|---|---|
| `krw_list_clusters` | `username?` | `{ok, clusters:[{title,url,username}]}` |
| `krw_add_cluster` | `title,username,password,trust_store_path,url` | `{ok}` / `{ok:false,error}` |
| `krw_delete_cluster` | `title` | `{ok}` / `{ok:false,error}` |
| `krw_show_selection` | — | `{ok, cluster, topic, partition}` |
| `krw_select_cluster` | `title` | `{ok, topics:[...]}` (ждёт загрузки) |
| `krw_select_topic` | `topic` | `{ok, partitions:[...]}` |
| `krw_select_partition` | `partition` (string) | `{ok}` |
| `krw_send_message` | `value` (+опц. overrides cluster/topic/partition) | `{ok}` / `{ok:false,error}` |

Правила:
- Хендлеры **select/send** идут через `Deps.SendProgram(request)` + ждут `<-Reply`.
- **list/add/delete** — напрямую через `Deps`; add/delete после успеха шлют
  `ClusterStoreChangedEvent{}`.
- **Ошибки** → `mcp.CallToolResult{IsError:true}` с JSON `{ok:false, error}`.
- Сериализация: `sync.Mutex` на весь вызов инструмента (один активный MCP-запрос).
- `krw_send_message` дополнительно сохраняет сообщение в `sent_messages_history`
  (переиспользуем существующий flow `sendAndSaveMessage`).

### Шаг 5. Сборка — `internal/app/app.go` (+ `di.go`)
1. В `Run()`: `p := tea.NewProgram(model)` сохранить в поле (или в замыкание Deps),
   чтобы MCP мог слать `p.Send`.
2. Собрать `mcp.Deps` из контейнера:
   - `ListClusters`/`GetCluster`/`AddCluster`/`DeleteCluster` → `d.LocalStorage()`.
   - `GetSelected` → `d.FullModel().GetSelected*`.
   - `SendProgram` → `p.Send`.
3. Если `AppConfig().McpEnabled`: `mcpSrv.Start()` до `p.Run()`, после `p.Run()`
   вернулся → `mcpSrv.Stop(ctx)` + закрыть через существующий `closer`. Отправить
   `contracts.MCPServerStartedEvent{Addr}` в program.
4. Порядок: запуск HTTP-сервера в goroutine → `p.Run()` блокирует → по выходу Shutdown.

### Шаг 6. Зависимость, сборка, тесты
- `go get github.com/modelcontextprotocol/go-sdk@v1.2.0`
- `gofmt -w ./... && go vet ./... && go build ./...` (task `build` из Taskfile.yml).
- Тесты:
  - `internal/mcp/tools_test.go` — юнит: хендлеры с фейковыми `Deps` (mock
    list/add/delete, send без сети).
  - `internal/mcp/server_test.go` — round-trip на `httptest.NewServer(handler)` сырым
    POST (как в примере SDK): `initialize` → `tools/list` → `tools/call krw_list_clusters`.
  - (опц.) интеграция через `tea.Program` с `WithInput`/`WithOutput`: проверить, что
    `MCPSendRequest` доходит до Send и возвращает reply.
- Ручная проверка: `KRW_MCP=1 go run ./cmd/krw`, затем
  `curl -s -X POST http://127.0.0.1:18810 -d '...initialize...'`.

### Шаг 7. Интеграция с ИИ-клиентом
```jsonc
// opencode.json
{ "mcp": { "krw": { "type": "remote", "url": "http://127.0.0.1:18810", "enabled": true } } }

// Claude Desktop claude_desktop_config.json
{ "mcpServers": { "krw": { "url": "http://127.0.0.1:18810" } } }
```
Порядок использования ИИ: `krw_show_selection` → `krw_select_cluster "sfa"` →
`krw_select_topic "orders"` → `krw_select_partition "0"` → `krw_send_message "hello"`.
Каждый шаг возвращает `ok` или текст ошибки.

---

## 4. Порядок внедрения (зависимости)

0. Блокеры (Шаг 0) — можно мержить отдельным PR, без MCP. Шаг 0.4 (кэш) — тоже
   самостоятельный, желателен до начала MCP-хендлеров.
1. Контракты + Модель (Шаги 2–3).
2. `internal/mcp` + App (Шаги 4–5) — минимальный рабочий зоопарк.
3. Инструменты, тесты, интеграция (Шаги 6–7).

Примерный размер: ~400–500 строк нового кода (без тестов), ~150–200 строк правок в
существующих файлах (из них ~100 — Шаг 0.4: рефакторинг `KafkaConnectorProvider`).

## 5. Открытые вопросы
- **«кластеры равны логину пользователя»**: показать все + фильтр `username`, или
  фильтровать по OS-логину/по токену сессии?
- Нужен ли `krw_list_topics` / `krw_list_partitions` как отдельные инструменты
  «просмотреть без выбора»? В `select_*` они уже возвращаются.
- Писать README-раздел про MCP или оставить в TODO? README в репо нет.