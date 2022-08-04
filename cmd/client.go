package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	ti "github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/typing-systems/typing/cmd/connections"
	"github.com/typing-systems/typing/cmd/utility"
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

	c           connections.ConnectionsClient
	conn        *grpc.ClientConn
	lanesStream connections.Connections_PositionsClient
	data        chan dataMsg
}

func initModel() model {
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
		data:         make(chan dataMsg),
	}

	return model
}

// Main view function, just serves to call the relevant views
func (m *model) View() string {
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

				lanesStream, err := connection.Positions(context.Background(), &connections.MyLobby{LobbyID: reply.LobbyID})
				if err != nil {
					log.Fatal("Error calling Positions", err)
				}
				m.lanesStream = lanesStream

				m.myLobby = reply.LobbyID
				m.myLane = reply.Lane

				m.c = connection
				m.conn = conn
				m.isConnected = true
			}
			return ViewOthers(*m)
		} else if m.cursor == 1 {
			return ViewYourself(*m)
		}
	} else if m.completed {
		return ViewResults(*m)
	}

	return ViewChoice(*m)
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

// Setup Functions
func (m *model) Init() tea.Cmd {
	return tea.Batch(
		listenForLanes(m),
		waitForLanes(m.data),
	)
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
