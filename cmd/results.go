package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/typing-systems/typing/cmd/utility"
	"golang.org/x/term"
)

func ViewResults(m model) string {
	var results = ""

	physicalWidth, physicalHeight, _ := term.GetSize(int(os.Stdout.Fd()))

	var container = lg.NewStyle().
		Width(physicalWidth).
		Height(physicalHeight).
		PaddingTop((physicalHeight - lg.Height(m.sentence) - 1) / 2).
		Align(lg.Center)

	var scheme = utility.ForegroundColour("#FFFFFF")

	results += scheme.Render(fmt.Sprintf("CPM: %.2f\nWPM: %.2f\nACCURACY: %.2f\n", cpm, wpm, accuracy))

	return container.Render(lg.JoinVertical(lg.Left, results))
}

// ?

func UpdateResults(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlQ:
			return &m, tea.Quit

		case tea.KeyCtrlB:
			m.completed = false
		}
	}

	return &m, nil
}
