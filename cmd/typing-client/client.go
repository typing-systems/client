package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type model struct {
	options []string
	cursor  int
}

func initModel() model {
	return model{
		options: []string{"Race others", "Race yourself"},
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) View() string {
	physicalWidth, physicalHeight, _ := term.GetSize(int(os.Stdout.Fd()))

	var leftHalf = lg.NewStyle().
		Width(physicalWidth / 2).
		Height(physicalHeight).
		Background(lg.Color("#344e41")).
		Align(lg.Center).
		PaddingTop((physicalHeight - 1) / 2)

	var rightHalf = lg.NewStyle().
		Width(physicalWidth / 2).
		Height(physicalHeight).
		Background(lg.Color("#000000")).
		Align(lg.Center).
		PaddingTop((physicalHeight - 4) / 2)

	left := "TYPING.SYSTEMS"

	right := ""

	for i, option := range m.options {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		right += (fmt.Sprintf("%s [%s]\n", cursor, option))
	}

	right += "\nPress q to quit."

	return lg.JoinHorizontal(lg.Center, leftHalf.Render(left), rightHalf.Render(right))
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func main() {
	client := tea.NewProgram(initModel(), tea.WithAltScreen())
	if err := client.Start(); err != nil {
		fmt.Println("Error starting client:", err)
		os.Exit(1)
	}
}
