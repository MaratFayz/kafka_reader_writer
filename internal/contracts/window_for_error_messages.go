package contracts

import tea "charm.land/bubbletea/v2"

type CloseWindowForErrorMessages struct{}

type ErrorMessage struct {
	Text string
}

func CreateErrorMessageEvent(text string) tea.Cmd {
	return func() tea.Msg {
		return ErrorMessage{Text: text}
	}
}
