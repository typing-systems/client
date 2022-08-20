package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/typing-systems/typing/cmd/utility"
)

func DefaultSettings() {
	utility.Settings["logging"] = false
	utility.Settings["setting2"] = true
	utility.Settings["setting3"] = false
}

func ViewSettings(m model) string {
	optionStyle := lg.NewStyle().Italic(true).Width(17).Align(lg.Center).Background(lg.Color("#ecfccb"))
	optionStyleOn := lg.NewStyle().Italic(true).Width(17).Align(lg.Center).Background(lg.Color("#84cc16"))

	settings := []string{"logging", "setting2", "setting3"}

	menu := ""
	var option string
	for i, isToggled := range settings {
		if utility.Settings[isToggled] {
			option = optionStyleOn.Render(settings[i])
		} else {
			option = optionStyle.Render(settings[i])
		}
		if m.settingsCursor == i {
			menu += fmt.Sprintf("> %s\n", option)
		} else {
			menu += fmt.Sprintf("  %s\n", option)
		}
	}
	return menu
}

func UpdateSettings(msg tea.Msg, m model) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyCtrlQ:
			return &m, tea.Quit

		case tea.KeyUp:
			if m.settingsCursor > 0 {
				m.settingsCursor--
			}

		case tea.KeyDown:
			if m.settingsCursor < 2 {
				m.settingsCursor++
			}

		case tea.KeyCtrlB:
			m.chosen = false

		case tea.KeyEnter, tea.KeySpace:
			switch m.settingsCursor {
			case 0:
				// logging
				utility.Settings["logging"] = !utility.Settings["logging"]
				utility.Log("toggling logging")
			case 1:
				utility.Settings["setting2"] = !utility.Settings["setting2"]
				utility.Log("toggling setting 2")
			case 2:
				utility.Settings["setting3"] = !utility.Settings["setting3"]
				utility.Log("toggling setting 3")
			}
		}
	}

	return &m, nil
}
