package contracts

import tea "charm.land/bubbletea/v2"

type CloseOverlayWindow struct{}

type ReadMessagesTableChosenRow struct {
	Header   string
	Body     string
	Offset   string
	Datetime string
}

func CreateReadMessagesTableChosenRowEvent(header string, body string, offset string, datetime string) tea.Cmd {
	return func() tea.Msg {
		return ReadMessagesTableChosenRow{Header: header, Body: body, Offset: offset, Datetime: datetime}
	}
}
