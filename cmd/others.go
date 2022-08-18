package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/typing-systems/typing/cmd/connections"
	"github.com/typing-systems/typing/cmd/utility"
	"golang.org/x/term"
)

type dataMsg struct {
	Lane   string
	Points int
}

func listenForLanes(m *model) {
	utility.Log("listenForLanes called")
	for {
		utility.Log("waiting for reply")
		reply, err := m.lanesStream.Recv()
		utility.Log("received a reply")
		if err != nil {
			log.Fatalf("error receiving from stream: %v", err)
		}
		m.data <- dataMsg{Lane: reply.Lane, Points: int(reply.Points)}
	}
}

func waitForLanes(data chan dataMsg) tea.Cmd {
	return func() tea.Msg {
		utility.Log("waitForLanes called")
		d := <-data
		return dataMsg{Lane: d.Lane, Points: d.Points}
	}
}

func updatePosition(m model) {
	utility.Log("sending position to server")
	_, err := m.c.UpdatePosition(context.Background(), &connections.MyPosition{LobbyID: m.myLobby, Lane: m.myLane})
	if err != nil {
		log.Fatalf("error sending position to server: %v", err)
	}
	utility.Log("position sent to server")
}

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
	case dataMsg:
		utility.Log("dataMsg received")

		switch msg.Lane {
		case "lane1":
			m.lanes[0] = msg.Points
			return &m, waitForLanes(m.data)
		case "lane2":
			m.lanes[1] = msg.Points
			return &m, waitForLanes(m.data)
		case "lane3":
			m.lanes[2] = msg.Points
			return &m, waitForLanes(m.data)
		case "lane4":
			m.lanes[3] = msg.Points
			return &m, waitForLanes(m.data)
		}
	case tea.KeyMsg:
		utility.Log("keyMsg received")
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

				go updatePosition(m)
			}
		}
	}

	return &m, nil
}
