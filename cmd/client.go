package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
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
	cpm      float64
	wpm      float64
	accuracy float64
)

type model struct {
	sentence     string
	userSentence string
	myLobby      string
	myLane       string

	strokes int
	cursor  int

	completed bool
	chosen    bool

	input ti.Model
	time  time.Time

	options        []string
	correctStrokes float64
	lanes          []int

	c             connections.ConnectionsClient
	conn          *grpc.ClientConn
	positionsChan chan *connections.PositionInfo
}

func initModel() model {
	conn, err := grpc.Dial(":9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %s", err)
	}

	connection := connections.NewConnectionsClient(conn)

	reply, err := connection.Connected(context.Background(), &connections.Empty{})
	if err != nil {
		log.Fatalf("Connected failed: %s", err)
	}

	input := ti.New()

	input.Focus()
	input.Prompt = ""
	input.SetCursorMode(2)

	model := model{
		options:       []string{"Race others", "Race yourself"},
		input:         input,
		userSentence:  "",
		completed:     false,
		lanes:         []int{1, 2, 3, 4},
		c:             connection,
		conn:          conn,
		myLobby:       reply.LobbyID,
		myLane:        reply.Lane,
		positionsChan: make(chan *connections.PositionInfo),
	}

	return model
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

				_, err := m.c.UpdatePosition(context.Background(), &connections.MyPosition{LobbyID: m.myLobby, Lane: m.myLane})
				if err != nil {
					log.Fatal("Error calling Positions", err)
				}
			}
		}
	// stream, err := m.c.Positions(context.Background(), &connections.MyLobby{LobbyID: m.myLobby})
	// if err != nil {
	// 	log.Fatalf("error with positions: %v", err)
	// }
	// go func() {
	// 	for {
	// 		value, err := stream.Recv()
	// 		if err == io.EOF {
	// 			return
	// 		}
	// 		if err != nil {
	// 			log.Fatalf("error receiving from stream: %v", err)
	// 		}
	// 		m.lanes[0], _ = strconv.Atoi(value.Lane1)
	// 		m.lanes[1], _ = strconv.Atoi(value.Lane2)
	// 		m.lanes[2], _ = strconv.Atoi(value.Lane3)
	// 		m.lanes[3], _ = strconv.Atoi(value.Lane4)
	// 	}
	// }()
	case positionMsg:
		m.lanes = msg.String()
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
			return UpdateOthers(msg, *m)
		} else if m.cursor == 1 {
			return UpdateYourself(msg, *m)
		}
	} else if m.completed {
		return UpdateResults(msg, *m)
	}

	return UpdateChoice(msg, *m)
}

type positionMsg []string

func (m *model) listenForPositions() {
	stream, err := m.c.Positions(context.Background(), &connections.MyLobby{LobbyID: m.myLobby})
	if err != nil {
		log.Fatalf("error with positions: %v", err)
	}
	for {
		value, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatalf("error receiving from stream: %v", err)
		}
		m.positionsChan <- value
	}
}

func waitForPositions(position chan *connections.PositionInfo) tea.Cmd {
	return func() tea.Msg {
		positionArr := make(positionMsg, 0)
		positionArr[0] = (<-position).Lane1
		positionArr[1] = (<-position).Lane2
		positionArr[2] = (<-position).Lane3
		positionArr[3] = (<-position).Lane4
		return positionArr
	}
}

// Setup Functions
func (m *model) Init() tea.Cmd {
	return nil
}

// Main function
func main() {
	m := initModel()
	defer m.conn.Close()

	go m.listenForPositions()

	client := tea.NewProgram(&m, tea.WithAltScreen())
	if err := client.Start(); err != nil {
		fmt.Sprintln("Error starting client:", err)
		os.Exit(1)
	}
}
