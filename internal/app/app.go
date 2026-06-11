package app

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

type App struct {
	diContainer *diContainer
}

func New() *App {
	a := &App{
		diContainer: newDIContainer(),
	}

	return a
}

func (a *App) Run() error {
	p := tea.NewProgram(a.diContainer.FullModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Atas, there's been an error: %v", err)
		os.Exit(1)
	}

	return nil
}
