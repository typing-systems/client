package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/typing-systems/typing/cmd/utility"
	"golang.org/x/term"
)

func ViewChoice(m model) string {
	physicalWidth, physicalHeight, _ := term.GetSize(int(os.Stdout.Fd()))

	leftHalf := utility.HalfGen(2, physicalWidth, physicalHeight, "#f5d0fe").Foreground(lg.Color("#262626"))
	rightHalf := utility.HalfGen(6, physicalWidth, physicalHeight, "#404040").UnsetAlign().PaddingLeft((physicalWidth - 54) / 4)

	left := "typing.systems"
	left = lg.JoinVertical(0, left, lg.NewStyle().Italic(true).Render("  タイピング.システム"))

	right := ""

	menuStyle := lg.NewStyle().Italic(true).Width(17).Align(lg.Center)

	for i, option := range m.options {
		if m.cursor == i {
			right += fmt.Sprintf("%s\n", menuStyle.Background(lg.Color("#f5f5f5")).Foreground(lg.Color("#262626")).MarginLeft(i*2).Render(option))
			continue
		}

		// right += (menuStyle.Background(lg.Color(menuColors[i])).Render(fmt.Sprintf("%s %s\n", cursor, option)))
		right += fmt.Sprintf("%s\n", menuStyle.Background(lg.Color(menuColors[i])).MarginLeft(i*2).Foreground(lg.Color("#f5f5f5")).Render(option))
	}

	// right += "\n      Press ctrl+q to quit."

	return lg.JoinHorizontal(lg.Center, leftHalf.Render(left), rightHalf.Render(right))
}

// Update function for when a choice hasn't been made
func UpdateChoice(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlQ:
			return &m, tea.Quit

		case tea.KeyUp:
			if m.cursor > 0 {
				m.cursor--
			}

		case tea.KeyDown:
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}

		case tea.KeyEnter, tea.KeySpace:
			m.chosen = true
			m.userSentence = ""
			m.sentence = utility.GetRandomSentence(10)
			m.correctStrokes = 0
			m.strokes = 0
			m.completed = false
		}
		// default:
		// 	var cmd tea.Cmd
		// 	m.spinner, cmd = m.spinner.Update(msg)
		// 	return &m, cmd

	}

	return &m, waitForLanes(m.data)
}
