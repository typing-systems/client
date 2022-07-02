package main

import (
	"fmt"
	"os"
	"time"

	ti "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/typing-systems/typing/cmd/typing-client/utility"
	"golang.org/x/term"
)

type model struct {
	options         []string
	cursor          int
	chosen          bool
	input           ti.Model
	sentence        string
	userSentence    string
	time            time.Time
	strokes         int
	correct_strokes float64
	completed       bool
}

func initModel() model {
	randSentence := utility.GetRandomSentence(10)
	input := ti.New()

	input.Focus()
	input.Prompt = ""
	input.SetCursorMode(2)
	input.CharLimit = len(randSentence)

	return model{
		options:      []string{"Race others", "Race yourself"},
		input:        input,
		sentence:     randSentence,
		userSentence: "",
		completed:    false,
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
			m.userSentence = ""
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
			m.sentence = utility.GetRandomSentence(10)
		}
	}
	return m, nil
}

//////// YOURSELF FUNCTIONS ////////

// This handles the view for when a choice has been made.
func ViewYourself(m model) string {
	if m.completed {
		return ViewResults(m)
	}

	physicalWidth, physicalHeight, _ := term.GetSize(int(os.Stdout.Fd()))

	var container = lg.NewStyle().
		Width(physicalWidth).
		Height(physicalHeight).
		PaddingTop((physicalHeight - lg.Height(m.sentence) - 1) / 2).
		PaddingLeft((physicalWidth - lg.Width(m.sentence)) / 2)

	var wrong = utility.ForegroundColour("#A7171A")
	var primary = utility.ForegroundColour("#525252")

	display := ""
	for i, char := range m.userSentence {
		if char == rune(m.sentence[i]) {
			display += string(char)
		} else if string(char) == " " {
			display += wrong.Render("_")
		} else {
			display += wrong.Render(string(char))
		}
	}

	remaining := m.sentence[len(m.userSentence):]

	display += primary.Render(remaining)

	return container.Render(lg.JoinVertical(lg.Left, display))
}

// Update function for when the user has chosen to play themselves
func UpdateYourself(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.time.IsZero() {
			m.time = time.Now()
		}

		switch msg.String() {
		case "ctrl+c", "ctrl+q":
			return m, tea.Quit

		case "ctrl+b":
			m.chosen = false
			m.sentence = utility.GetRandomSentence(10)

		case "backspace":
			if len(m.userSentence) > 0 {
				m.userSentence = m.userSentence[:len(m.userSentence)-1]
				return m, nil
			}

		case " ":
			if len(m.userSentence) < len(m.sentence) {
				m.userSentence += " "
				return m, nil
			}

		case "enter":
			if len(m.userSentence) == len(m.sentence) {
				m.completed = true
			}

			return m, nil
		}

		if msg.Type != tea.KeyRunes {
			return m, nil
		}

		if len(m.userSentence) < len(m.sentence) {
			m.userSentence += msg.String()
		}

	}

	return m, nil
}

func ViewResults(m model) string {
	var results = ""

	physicalWidth, physicalHeight, _ := term.GetSize(int(os.Stdout.Fd()))
	cpm, wpm, accuracy := utility.CalculateStats(m.correct_strokes, m.strokes, m.time)

	var container = lg.NewStyle().
		Width(physicalWidth).
		Height(physicalHeight).
		PaddingTop((physicalHeight - lg.Height(m.sentence) - 1) / 2).
		PaddingLeft((physicalWidth - lg.Width(m.sentence)) / 2)

	var scheme = utility.ForegroundColour("#FFFFFF")

	results += scheme.Render(fmt.Sprintf("CPM: %.2f\n", cpm))
	results += scheme.Render(fmt.Sprintf("WPM: %.2f\n", wpm))
	results += scheme.Render(fmt.Sprintf("ACCURACY: %.2f\n", accuracy))

	return container.Render(lg.JoinVertical(lg.Left, results))
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

// Main function
func main() {
	client := tea.NewProgram(initModel(), tea.WithAltScreen())
	if err := client.Start(); err != nil {
		fmt.Println("Error starting client:", err)
		os.Exit(1)
	}
}
