package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	ti "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/typing-systems/typing/cmd/connections"
	"github.com/typing-systems/typing/cmd/utility"
	"golang.org/x/term"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	cpm        float64
	wpm        float64
	accuracy   float64
	wrong      = utility.ForegroundColour("#b91c1c")
	primary    = utility.ForegroundColour("#525252")
	menuColors = []string{"#e879f9", "#d946ef", "#c026d3", "#a21caf", "#86198f", "#701a75"}
)

type model struct {
	sentence     string
	userSentence string
	myLobby      string
	myLane       string

	strokes int
	cursor  int

	completed   bool
	chosen      bool
	isConnected bool

	input ti.Model
	time  time.Time

	options        []string
	correctStrokes float64
	lanes          []int

	c    connections.ConnectionsClient
	conn *grpc.ClientConn
}

func initModel() model {
	// conn, err := grpc.Dial(":9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	log.Fatalf("did not connect: %s", err)
	// }

	// connection := connections.NewConnectionsClient(conn)

	// reply, err := connection.Connected(context.Background(), &connections.Empty{})
	// if err != nil {
	// 	log.Fatalf("Connected failed: %s", err)
	// }

	input := ti.New()

	input.Focus()
	input.Prompt = ""
	input.SetCursorMode(2)

	model := model{
		options:      []string{"race others", "race yourself", "leaderboards", "stats", "options", "something"},
		input:        input,
		userSentence: "",
		completed:    false,
		lanes:        []int{1, 2, 3, 4},
		isConnected:  false,
	}

	return model
}

//////// MAIN MENU FUNCTIONS ////////

// This handles the view when a choice has not been made, ie the first screen you see.
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
	}

	return &m, nil
}

//////// OTHERS FUNCTIONS ////////
func ViewOthers(m model) string {
	physicalWidth, physicalHeight, _ := term.GetSize(int(os.Stdout.Fd()))

	topHalf := lg.NewStyle().
		Width(physicalWidth).
		Height(physicalHeight / 2).
		Background(lg.Color("#525252")).
		PaddingTop((physicalHeight / 4) - lg.Height(m.sentence))

	lane1 := strings.Repeat("=", m.lanes[0])
	lane2 := strings.Repeat("=", m.lanes[1])
	lane3 := strings.Repeat("=", m.lanes[2])
	lane4 := strings.Repeat("=", m.lanes[3])

	bottomHalf := topHalf.Copy().UnsetBackground()

	display := ""
	for i, char := range m.userSentence {
		if char == rune(m.sentence[i]) {
			display += string(char)
		} else if string(char) == " " {
			display += wrong.Render("_")
		} else {
			display += wrong.Render(string(m.sentence[i]))
		}
	}

	remaining := m.sentence[len(m.userSentence):]

	display += primary.Render(remaining)

	return lg.JoinVertical(lg.Center, topHalf.Render(lg.JoinVertical(lg.Left, lane1, lane2, lane3, lane4)), bottomHalf.Render(display))
}

func UpdateOthers(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.time.IsZero() {
			m.time = time.Now()
		}

		if m.completed {
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyCtrlQ:
				return &m, tea.Quit

			case tea.KeyCtrlB:
				m.chosen = false

			case tea.KeyBackspace:
				if len(m.userSentence) > 0 {
					m.userSentence = m.userSentence[:len(m.userSentence)-1]
					return &m, nil
				}
			}
		} else {
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyCtrlQ:
				return &m, tea.Quit

			case tea.KeyCtrlB:
				m.chosen = false

			case tea.KeyBackspace:
				if len(m.userSentence) > 0 {
					m.userSentence = m.userSentence[:len(m.userSentence)-1]
					return &m, nil
				}

			case tea.KeySpace:
				if len(m.userSentence) < len(m.sentence) {
					m.userSentence += " "
					return &m, nil
				}

				return &m, nil
			}

			if msg.Type != tea.KeyRunes {
				return &m, nil
			}
		}

		m.strokes++

		if len(m.userSentence) < len(m.sentence) {
			m.userSentence += msg.String()

			if msg.Runes[0] == rune(m.sentence[len(m.userSentence)-1]) {
				m.correctStrokes++

				reply, err := m.c.Positions(context.Background(), &connections.MyPosition{ID: m.myLobby, Lane: m.myLane})
				if err != nil {
					log.Fatal("Error calling Positions", err)
				}
				m.lanes[0], _ = strconv.Atoi(reply.Lane1)
				m.lanes[1], _ = strconv.Atoi(reply.Lane2)
				m.lanes[2], _ = strconv.Atoi(reply.Lane3)
				m.lanes[3], _ = strconv.Atoi(reply.Lane4)
			}
		}
	}

	return &m, nil
}

//////// YOURSELF FUNCTIONS ////////

// This handles the view for when a choice has been made.
func ViewYourself(m model) string {
	physicalWidth, physicalHeight, _ := term.GetSize(int(os.Stdout.Fd()))

	var container = lg.NewStyle().
		Width(physicalWidth).
		Height(physicalHeight).
		PaddingTop((physicalHeight - lg.Height(m.sentence)) / 2).
		PaddingLeft((physicalWidth - lg.Width(m.sentence)) / 2)

	display := ""
	for i, char := range m.userSentence {
		if char == rune(m.sentence[i]) {
			display += string(char)
		} else if string(char) == " " {
			display += wrong.Render("_")
		} else {
			display += wrong.Render(string(m.sentence[i]))
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

		if m.completed {
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyCtrlQ:
				return &m, tea.Quit

			case tea.KeyCtrlB:
				m.chosen = false

			case tea.KeyBackspace:
				if len(m.userSentence) > 0 {
					m.userSentence = m.userSentence[:len(m.userSentence)-1]
					return &m, nil
				}
			}
		} else {
			switch msg.Type {
			case tea.KeyCtrlC, tea.KeyCtrlQ:
				return &m, tea.Quit

			case tea.KeyCtrlB:
				m.chosen = false

			case tea.KeyBackspace:
				if len(m.userSentence) > 0 {
					m.userSentence = m.userSentence[:len(m.userSentence)-1]
					return &m, nil
				}

			case tea.KeySpace:
				if len(m.userSentence) < len(m.sentence) {
					m.userSentence += " "
					return &m, nil
				}

			case tea.KeyEnter:
				if len(m.userSentence) == len(m.sentence) {
					m.completed = true
					m.chosen = false
					cpm, wpm, accuracy = utility.CalculateStats(m.correctStrokes, m.strokes, m.time)
				}

				return &m, nil
			}

			if msg.Type != tea.KeyRunes {
				return &m, nil
			}
		}

		m.strokes++

		if len(m.userSentence) < len(m.sentence) {
			m.userSentence += msg.String()

			if msg.Runes[0] == rune(m.sentence[len(m.userSentence)-1]) {
				m.correctStrokes++
			}
		}

		if string(msg.Runes[0]) == m.sentence[len(m.userSentence)-1:] {
			m.completed = true
			m.chosen = false
			cpm, wpm, accuracy = utility.CalculateStats(m.correctStrokes, m.strokes, m.time)
		}
	}

	return &m, nil
}

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

//////// MAIN FUNCTIONS ////////

// Main view function, just serves to call the relevant views
func (m model) View() string {
	if m.chosen {
		if m.cursor == 0 {
			return ViewOthers(m)
		} else if m.cursor == 1 {
			return ViewYourself(m)
		}
	} else if m.completed {
		return ViewResults(m)
	}

	return ViewChoice(m)
}

// Main update function, just serves to call the relevant update function
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.chosen {
		if m.cursor == 0 {
			if !m.isConnected {
				conn, err := grpc.Dial(":9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					log.Fatalf("did not connect: %s", err)
				}

				connection := connections.NewConnectionsClient(conn)

				reply, err := connection.Connected(context.Background(), &connections.Empty{})
				if err != nil {
					log.Fatalf("Connected failed: %s", err)
				}
				m.myLobby = reply.ID
				m.myLane = reply.Lane

				m.c = connection
				m.conn = conn
				m.isConnected = true
			}
			return UpdateOthers(msg, *m)
		} else if m.cursor == 1 {
			return UpdateYourself(msg, *m)
		}
	} else if m.completed {
		return UpdateResults(msg, *m)
	}

	return UpdateChoice(msg, *m)
}

// Setup Functions
func (m *model) Init() tea.Cmd {
	return nil
}

// Main function
func main() {
	model := initModel()
	defer model.conn.Close()

	client := tea.NewProgram(&model, tea.WithAltScreen())
	if err := client.Start(); err != nil {
		fmt.Sprintln("Error starting client:", err)
		os.Exit(1)
	}
}
