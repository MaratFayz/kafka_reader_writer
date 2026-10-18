# План: соответствие Go-конвенциям, линтер и инструменты

Единый PR. Согласовано, ожидает подтверждения перед реализацией.

## 1. Переименование файлов и каталога (Go-конвенции именования)

`git mv` 23 файлов и каталога `readWriteTabs` → `readwritetabs`. Внутренности файлов (типы, функции) НЕ меняются.

### internal/contracts/
| Файл сейчас | Новое имя |
|---|---|
| `KafkaReadWriteTabsComponent.go` | `kafka_read_write_tabs_component.go` |
| `windowForAppendNewCluster.go` | `window_for_append_new_cluster.go` |
| `windowForDetailedKafkaMessage.go` | `window_for_detailed_kafka_message.go` |
| `messagesRow.go` | `messages_row.go` |
| `windowForErrorMessages.go` | `window_for_error_messages.go` |
| `windowForDeleteCluster.go` | `window_for_delete_cluster.go` |
| `windowForEditCluster.go` | `window_for_edit_cluster.go` |
| `windowForHelp.go` | `window_for_help.go` |

### internal/components/
| Файл сейчас | Новое имя |
|---|---|
| `kafkaPartitionList.go` | `kafka_partition_list.go` |
| `textArea.go` | `text_area.go` |
| `popupForErrorMessages.go` | `popup_for_error_messages.go` |
| `popupForDeleteCluster.go` | `popup_for_delete_cluster.go` |
| `kafkaClusterList.go` | `kafka_cluster_list.go` |
| `popupForDetailedReadKafkaMessage.go` | `popup_for_detailed_read_kafka_message.go` |
| `popupAppendNewCluster.go` | `popup_append_new_cluster.go` |
| `popupEditCluster.go` | `popup_edit_cluster.go` |
| `popupForHelp.go` | `popup_for_help.go` |
| `kafkaTopicList.go` | `kafka_topic_list.go` |

### internal/components/readWriteTabs/ (каталог + 5 файлов)
| Файл/каталог сейчас | Новое имя |
|---|---|
| каталог `readWriteTabs` | `readwritetabs` |
| `kafkaReadWriteTabs.go` | `kafka_read_write_tabs.go` |
| `kafkaSendMessageTextArea.go` | `kafka_send_message_text_area.go` |
| `spinnerWhenLoadMessagesFromKafka.go` | `spinner_when_load_messages_from_kafka.go` |
| `kafkaReadMessagesTable.go` | `kafka_read_messages_table.go` |
| `kafkaSentMessagesTable.go` | `kafka_sent_messages_table.go` |

### Обновление импорта
- `internal/app/di.go:8`: путь `internal/components/readWriteTabs` → `.../readwritetabs`, избыточный алиас `readwritetabs` убрать.

## 2. Линтер golangci-lint
- Установка: `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest` (кросс-платформенно).
- `.golangci.yml` (v2):
  - `linters.default: standard` (govet, ineffassign, staticcheck, errcheck и др.)
  - `linters.enable`: `revive`, `gofumpt`, `gocritic`, `gocyclo`, `prealloc`, `misspell`, `unconvert`, `nakedret`, `errorlint`, `durationcheck`, `nilnil`, `thelper`, `copyloopvar`, `intrange`, `gosec`
  - `formatters.enable`: `gofmt`

## 3. Taskfile — установка инструментов, компилятора и проверок
```yaml
install:
  desc: Установка инструментов и зависимостей
  cmds:
    - task: install:lint
    - task: install:sqlite
    - task: check:compiler
    - task: install:compiler

install:lint:
  desc: Установка golangci-lint (go install, кросс-платформенно)
  cmds:
    - go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest

install:sqlite:
  desc: Скачивание зависимостей, включая go-sqlite3 (cgo)
  cmds:
    - go mod download

check:compiler:
  desc: Проверка наличия C-компилятора (требуется cgo/go-sqlite3)
  cmds:
    - cc --version

install:compiler:
  desc: Установка C-компилятора под текущую платформу (пропускается, если cc уже есть)
  status:
    - cc --version
  cmds:
    - cmd: xcode-select --install
      platforms: [darwin]
    - cmd: sudo apt-get install -y build-essential
      platforms: [linux]

lint:
  desc: Запуск линтера
  cmds:
    - golangci-lint run
```

Нюансы:
- `status: cc --version` в `install:compiler` — если компилятор уже установлен, задача пропускается автоматически.
- macOS: `xcode-select --install` попросит подтверждение в GUI.
- Linux: требует `apt` (Debian/Ubuntu). При желании позже добавим `windows` (мин: `choco install mingw -y`).

## 4. Верификация после изменений
- `go build ./...`
- `go vet ./...`
- Первый `golangci-lint run` — замечания текущего кода фиксируем в сводке, правим отдельной итерацией (отложено по договорённости).

## Замечания вне скоупа (не нарушения, к сведению)
- Модуль `marat/fayz/kafka_reader_writer` содержит `_` — допустимо, но не конвенционно.
- Пакет `connection` в `kafka/connection` — слишком общий нейминг, но стилистически валиден.
- go-sqlite3 уже в go.mod (`v1.14.44`); для сборки требуется cgo + C-компилятор.
- Ни один mainstream-линтер не проверяет snake_case имён файлов — это чистая конвенция.