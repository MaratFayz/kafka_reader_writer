package components

import (
	"marat/fayz/kafka_reader_writer/internal/windows"
	"strings"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type KafkaSendMessageTextAreaComponent struct {
	textarea *textarea.Model
	model    *windows.Model
}

func (k *KafkaSendMessageTextAreaComponent) Init() tea.Cmd {
	return tea.Batch(textarea.Blink, tea.RequestBackgroundColor)
}

func (k *KafkaSendMessageTextAreaComponent) Update(msg tea.Msg, m *windows.Model) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		// Update styling now that we know the background color.
		k.textarea.SetStyles(textarea.DefaultStyles(msg.IsDark()))

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			if k.textarea.Focused() {
				k.textarea.Blur()
			}
		case "ctrl+c":
			return m, tea.Quit
		default:
			if !k.textarea.Focused() {
				cmd = k.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}

		// 	// We handle errors just like any other message
		// case errMsg:
		// 	m.err = msg
		// 	return m, nil
	}

	ta, cmd := k.textarea.Update(msg)
	k.textarea = &ta
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (k *KafkaSendMessageTextAreaComponent) headerView() string {
	return "Tell me a story.\n"
}

func (k *KafkaSendMessageTextAreaComponent) View() string {
	const (
		footer = "\n(ctrl+c to quit)\n"
	)

	var c *tea.Cursor
	if !k.textarea.VirtualCursor() {
		c = k.textarea.Cursor()

		if c != nil {
			// Set the y offset of the cursor based on the position of the textarea
			// in the application.
			offset := lipgloss.Height(k.headerView())
			c.Y += offset
		}
	}

	f := strings.Join([]string{
		k.headerView(),
		k.textarea.View(),
		footer,
	}, "\n")

	// v := tea.NewView(f)
	// v.Cursor = c
	// return v
	return f
}

func CreateKafkaSendMessageTextArea(m *windows.Model) *KafkaSendMessageTextAreaComponent {
	ti := textarea.New()
	ti.Placeholder = "Once upon a time..."
	ti.SetVirtualCursor(false)
	ti.SetStyles(textarea.DefaultStyles(true)) // default to dark styles.
	ti.Focus()

	k := &KafkaSendMessageTextAreaComponent{textarea: &ti}
	return k
}
