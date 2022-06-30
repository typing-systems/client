package main

import (
	"fmt"
	"os"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/typing-systems/typing/cmd/typing-client/utility"
	"golang.org/x/term"
)

type model struct {
	options []string
	cursor  int
	chosen  bool
	input   textinput.Model
}

func initModel() model {
	input := textinput.New()
	input.Focus()
	input.Prompt = ""

	return model{
		options: []string{"Race others", "Race yourself"},
		input:   input,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

//////// MAIN MENU FUNCTIONS ////////
// This handles the view when a choice has not been made, ie the first screen you see.
func ViewChoice(m model) string {
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

// Update function for when a choice hasn't been made
func UpdateChoice(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.options)-1 {
				m.cursor++
			}

		case "enter", " ":
			m.chosen = true
		}
	}
	return m, nil
}

//////// OTHERS FUNCTIONS ////////
// This handles the view for when a choice has been made.
func ViewOthers(m model) string {
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

	right := "CHOSEN OTHERS"

	right += "\n\nPress backspace to go back to the main menu."
	right += "\nPress q to quit."

	return lg.JoinHorizontal(lg.Center, leftHalf.Render(left), rightHalf.Render(right))
}

// Update function for when the user has chosen to play others
func UpdateOthers(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit

		case "ctrl+b":
			m.chosen = false
		}
	}
	return m, nil
}

//////// YOURSELF FUNCTIONS ////////
// This handles the view for when a choice has been made.
func ViewYourself(m model) string {
	sentence := utility.GetRandomSentence(10)

	physicalWidth, physicalHeight, _ := term.GetSize(int(os.Stdout.Fd()))

	var container = lg.NewStyle().
		Width(physicalWidth).
		Height(physicalHeight).
		PaddingTop((physicalHeight - lg.Height(sentence) - 1) / 2).
		PaddingLeft((physicalWidth - lg.Width(sentence)) / 2)

	return container.Render(lg.JoinVertical(lg.Left, sentence, m.input.View()))
}

// Update function for when the user has chosen to play themselves
func UpdateYourself(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit

		case "ctrl+b":
			m.chosen = false
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

//////// MAIN FUNCTIONS ////////
// Main view function, just serves to call the relevant views
func (m model) View() string {
	if m.chosen {
		if m.cursor == 0 {
			return ViewOthers(m)
		} else if m.cursor == 1 {
			return ViewYourself(m)
		}
	}

	return ViewChoice(m)
}

// Main update function, just serves to call the relevant update function
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.chosen {
		if m.cursor == 0 {
			return UpdateOthers(msg, m)
		} else if m.cursor == 1 {
			return UpdateYourself(msg, m)
		}
	}

	return UpdateChoice(msg, m)
}

func main() {
	client := tea.NewProgram(initModel(), tea.WithAltScreen())
	if err := client.Start(); err != nil {
		fmt.Println("Error starting client:", err)
		os.Exit(1)
	}
}
