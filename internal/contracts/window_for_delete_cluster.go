package contracts

type ClientDisagreeDeleteCluster struct{}

type ClientAgreeDeleteCluster struct {
	Name string
}

type DeleteCluster struct {
	Name string
}

// type ErrorMessage struct {
// 	Text string
// }

// func CreateErrorMessageEvent(text string) tea.Cmd {
// 	return func() tea.Msg {
// 		return ErrorMessage{Text: text}
// 	}
// }
