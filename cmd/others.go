package main

import (
	"context"
	"log"
	"os"
	"strconv"
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

func listenForLanes(m *model) tea.Cmd {
	return func() tea.Msg {
		utility.Log("inside listenForLanes func")
		for {
			utility.Log("inside for listenForLanes")
			m.data <- dataMsg{Lane: "lane3", Points: 50}
			if data, ok := <-m.data; ok {
				utility.Log("inside if in listenForLanes")
				reply, err := m.lanesStream.Recv()
				if err != nil {
					log.Fatalf("error receiving from stream: %v", err)
				}
				m.data <- dataMsg{Lane: reply.Lane, Points: int(reply.Points)}
			} else {
				utility.Log("inside for in listenForLanes: data: " + data.Lane + " ok: " + strconv.FormatBool(ok))
			}
		}
	}
}

// func listenForLanes(m *model) tea.Cmd {
// 	return func() tea.Msg {
// 		utility.Log("inside listenForLanes func")
// 		for {
// 			utility.Log("inside for listenForLanes")
// 			reply, err := m.lanesStream.Recv()
// 			if err != nil {
// 				log.Fatalf("error receiving from stream: %v", err)
// 			}
// 			m.data <- dataMsg{Lane: reply.Lane, Points: int(reply.Points)}
// 		}
// 	}
// }

func waitForLanes(data chan dataMsg) tea.Cmd {
	return func() tea.Msg {
		utility.Log("waitForLanes")
		return dataMsg(<-data)
	}
}

// func listenForLanes(data chan dataMsg) tea.Cmd {
// 	fmt.Println("listenForLanes ran")
// 	return func() tea.Msg {
// 		for {
// 			time.Sleep(time.Second / 30)
// 			data <- dataMsg{Lane: "lane1", Points: 80}
// 		}
// 	}
// }

// func waitForLanes(data chan dataMsg) tea.Cmd {
// 	fmt.Println("waitForLanes ran")
// 	return func() tea.Msg {
// 		fmt.Println("dataMsg sent")
// 		return dataMsg(<-data)
// 	}
// }

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

				m.c.UpdatePosition(context.Background(), &connections.MyPosition{LobbyID: m.myLobby, Lane: m.myLane})
			}
		}
		// default:
		// 	var cmd tea.Cmd
		// 	m.spinner, cmd = m.spinner.Update(msg)
		// 	return &m, cmd
	}

	return &m, listenForLanes(&m)
}
