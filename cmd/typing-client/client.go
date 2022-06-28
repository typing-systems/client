package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

type model struct {
	choices []string // items on the to-do list
	cursor  int      // which to-do list item our cursor is pointing at
}

func initialModel() model {
	return model{
		// Our shopping list is a grocery list
		choices: []string{"Race others", "Race yourself"},
	}
}

func (m model) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "left", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "right", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m model) View() string {
	physicalWidth, physicalHeight, _ := term.GetSize(int(os.Stdout.Fd()))

	// The header
	var header = lg.NewStyle().
		Height(3).
		Bold(true).
		Italic(true).
		Foreground(lg.Color("#588157")).
		BorderStyle(lg.ThickBorder()).
		BorderForeground(lg.Color("#344e41")).
		Padding(1, 3)

	var container = lg.NewStyle().
		Width(physicalWidth).
		Height(physicalHeight).
		Align(lg.Center).
		PaddingTop((physicalHeight - 10) / 2)
		/* minus 10 because thats the number of lines printed (lines 84, 85, 97, 101)*2
		there are definitely better ways to do this but trial and error got me here baby */

	s := header.Render("TYPING.SYSTEMS")
	s += "\n\n"

	// Iterate over our choices
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Render the row
		s += fmt.Sprintf(" %s [%s] ", cursor, choice)
	}

	// The footer
	s += "\nPress q to quit."

	// Send the UI for rendering
	return container.Render(s)
}

func main() {
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
