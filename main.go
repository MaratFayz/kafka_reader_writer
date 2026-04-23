package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"

	"charm.land/bubbles/v2/cursor"
	"charm.land/bubbles/v2/textarea"
	"charm.land/lipgloss/v2"
)

type Config struct {
	username            string
	password            string
	broker              string
	topic               string
	partition           int
	certificateFilePath string
}

type model struct {
	broker      string
	topic       string
	partition   int
	textarea    textarea.Model
	senderStyle lipgloss.Style
	messages    []string
	config      *Config
}

func main() {
	config := &Config{
		username:  "SFA",
		password:  "SFADEV123",
		broker:    "172.16.15.171:9093",
		topic:     "SFA_MORTG_DEALCALL_TEST",
		partition: 0,
		// certificateFilePath: "./bmc.crt",
		certificateFilePath: "./certificate.pem",
	}

	p := tea.NewProgram(initialModel(config))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func sendMessage(config *Config, message string) {
	mechanism := plain.Mechanism{
		Username: config.username,
		Password: config.password,
	}

	ca, err := os.Open(config.certificateFilePath)
	if err != nil {
		panic(err)
	}
	cab, err := ioutil.ReadAll(ca)
	if err != nil {
		panic(err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cab)

	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
		// ClientCAs: ,
	}

	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: mechanism,
		TLS:           tlsConfig,
	}

	conn, err := dialer.Dial("tcp", config.broker)

	// conn, err := dialer.DialLeader(context.Background(), "tcp", config.broker, config.topic, config.partition)
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}

	partitions, err := conn.ReadPartitions()
	if err != nil {
		panic(err.Error())
	}

	// Собираем имена топиков
	m := map[string]struct{}{}
	for _, p := range partitions {
		m[p.Topic] = struct{}{}
	}

	// Выводим имена топиков
	for k := range m {
		fmt.Println(k)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = conn.WriteMessages(
		kafka.Message{Topic: config.topic, Value: []byte(message)},
		// kafka.Message{Value: []byte("two!")},
		// kafka.Message{Value: []byte("three!")},
	)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}

	if err := conn.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
}

func initialModel(config *Config) model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.SetVirtualCursor(false)
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	s := ta.Styles()
	s.Focused.CursorLine = lipgloss.NewStyle()
	ta.SetStyles(s)

	ta.ShowLineNumbers = false

	return model{
		broker:      config.broker,
		topic:       config.topic,
		partition:   config.partition,
		textarea:    ta,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		messages:    []string{},
		config:      config,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyPressMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		case "enter":
			mes := m.textarea.Value()
			m.messages = append(m.messages, m.senderStyle.Render("You: ")+mes)
			m.textarea.Reset()
			sendMessage(m.config, mes)
			return m, nil

		default:
			// Send all other keypresses to the textarea.
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
			// // The "up" and "k" keys move the cursor up
			// case "up", "k":
			// 	if m.cursor > 0 {
			// 		m.cursor--
			// 	}

			// // The "down" and "j" keys move the cursor down
			// case "down", "j":
			// 	if m.cursor < len(m.choices)-1 {
			// 		m.cursor++
			// 	}

			// // The "enter" key and the space bar toggle the selected state
			// // for the item that the cursor is pointing at.
			// case "enter", "space":
			// 	_, ok := m.selected[m.cursor]
			// 	if ok {
			// 		delete(m.selected, m.cursor)
			// 	} else {
			// 		m.selected[m.cursor] = struct{}{}
			// 	}
		}

	case cursor.BlinkMsg:
		// Textarea should also process cursor blinks.
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() tea.View {
	// The header
	s := "Параметры\n\n"

	s += fmt.Sprintf("Brokers: %s\nTopic: %s\nPartition:%d\n", m.broker, m.topic, m.partition)

	for _, m := range m.messages {
		s += m
		s += "\n"
	}

	// The footer
	s += "\nPress q to quit.\n"

	v := tea.NewView(s + "\n" + m.textarea.View())
	c := m.textarea.Cursor()
	v.Cursor = c
	v.AltScreen = true
	return v
}
