package contracts

import tea "charm.land/bubbletea/v2"

type CloseWindowForHelp struct{}

type AskForHelp struct {
}

type HelpMessage struct {
	Text string
}

func CreateAskForHelpEvent() tea.Cmd {
	return func() tea.Msg {
		return AskForHelp{}
	}
}

func CreateHelpMessageEvent(text string) tea.Cmd {
	return func() tea.Msg {
		return HelpMessage{Text: text}
	}
}
