package main

import (
	"fmt"
	"os"
	"strings"

	ti "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/typing-systems/typing/cmd/typing-client/utility"
	"golang.org/x/term"
)

type model struct {
	options  []string
	cursor   int
	chosen   bool
	input    ti.Model
	sentence string
	index    int
	wrongMap map[int]bool
}

func initModel() model {
	randSentence := utility.GetRandomSentence(10)
	input := ti.New()

	input.Focus()
	input.Prompt = ""
	input.SetCursorMode(2)
	input.CharLimit = len(randSentence)

	return model{
		options:  []string{"Race others", "Race yourself"},
		input:    input,
		sentence: randSentence,
		wrongMap: make(map[int]bool),
	}
}

//////// MAIN MENU FUNCTIONS ////////
// This handles the view when a choice has not been made, ie the first screen you see.

func ViewChoice(m model) string {
	physicalWidth, physicalHeight, _ := term.GetSize(int(os.Stdout.Fd()))

	var leftHalf = utility.HalfGen(1, physicalWidth, physicalHeight, "#344e41")
	var rightHalf = leftHalf.Copy().Background(lg.Color("#000000")).PaddingTop((physicalHeight - 4) / 2)

	left := "TYPING.SYSTEMS"
	right := ""

	for i, option := range m.options {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		right += (fmt.Sprintf("%s [%s]\n", cursor, option))
	}

	right += "\nPress ctrl+q to quit."

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
			m.index = -1
		}
	}

	return m, nil
}

//////// OTHERS FUNCTIONS ////////
// This handles the view for when a choice has been made.

func ViewOthers(m model) string {
	physicalWidth, physicalHeight, _ := term.GetSize(int(os.Stdout.Fd()))

	var leftHalf = utility.HalfGen(1, physicalWidth, physicalHeight, "#344e41")
	var rightHalf = leftHalf.Copy().Background(lg.Color("#000000")).PaddingTop((physicalHeight - 4) / 2)

	left := "TYPING.SYSTEMS"
	right := "CHOSEN OTHERS"

	right += "\n\nPress backspace to go back to the main menu."
	right += "\nPress ctrl+q to quit."

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
	physicalWidth, physicalHeight, _ := term.GetSize(int(os.Stdout.Fd()))

	var container = lg.NewStyle().
		Width(physicalWidth).
		Height(physicalHeight).
		PaddingTop((physicalHeight - lg.Height(m.sentence) - 1) / 2).
		PaddingLeft((physicalWidth - lg.Width(m.sentence)) / 2)

	var wrong = utility.ForegroundColour("#A7171A")
	var primary = utility.ForegroundColour("#525252")

	sentence := ""

	for i := 0; i < len(m.sentence); i++ {
		if m.wrongMap[i] {
			if m.sentence[i:i+1] == " " {
				sentence += strings.Replace(m.sentence[i:i+1], " ", wrong.Render("_"), 1)
			} else {
				sentence += wrong.Render(m.sentence[i : i+1])
			}
		} else if i <= m.index {
			sentence += m.sentence[i : i+1]
		} else {
			sentence += primary.Render(m.sentence[i : i+1])
		}
	}

	return container.Render(lg.JoinVertical(lg.Left, sentence))
}

// Update function for when the user has chosen to play themselves

func UpdateYourself(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit

		case "ctrl+b":
			m.chosen = false

		case "backspace":
			if m.index != -1 {
				if m.wrongMap[m.index] {
					m.wrongMap[m.index] = false
				}
				m.index--
			}
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
	}

	m.input, cmd = m.input.Update(msg)

	if m.index != len(m.sentence)-1 {
		m.index++
	}

	if m.index > -1 {
		if m.input.View()[m.index:m.index+1] != m.sentence[m.index:m.index+1] {
			m.wrongMap[m.index] = true
		} else {
			m.wrongMap[m.index] = false
		}
	}

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

// Setup Functions

func (m model) Init() tea.Cmd {
	return nil
}

func main() {
	client := tea.NewProgram(initModel(), tea.WithAltScreen())
	if err := client.Start(); err != nil {
		fmt.Println("Error starting client:", err)
		os.Exit(1)
	}
}
