package contracts

import (
	tea "charm.land/bubbletea/v2"
)

type CloseWindowForHelp struct{}

type AskForHelp struct {
}

type HelpMessage struct {
	Values []KeyValue
}

type KeyValue struct {
	Key   string
	Value string
}

func CreateAskForHelpEvent() tea.Cmd {
	return func() tea.Msg {
		return AskForHelp{}
	}
}

func CreateHelpMessageEvent(values []KeyValue) tea.Cmd {
	return func() tea.Msg {
		return HelpMessage{Values: values}
	}
}
