package main

import (
	"fmt"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
	"github.com/typing-systems/typing/cmd/settings"
	"github.com/typing-systems/typing/cmd/utility"
)

func ViewSettings(m model) string {
	optionStyle := lg.NewStyle().Italic(true).Width(17).Align(lg.Center).Background(lg.Color("#ecfccb"))
	optionStyleOn := lg.NewStyle().Italic(true).Width(17).Align(lg.Center).Background(lg.Color("#84cc16"))

	v := reflect.ValueOf(settings.Values)
	menu := ""
	var option string
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Bool() {
			option = optionStyleOn.Render(v.Type().Field(i).Name)
		} else {
			option = optionStyle.Render(v.Type().Field(i).Name)
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
				settings.Values.Logging = !settings.Values.Logging
				utility.Log("toggling logging")
			case 1:
				settings.Values.Setting2 = !settings.Values.Setting2
				utility.Log("toggling setting 2")
			case 2:
				settings.Values.Setting3 = !settings.Values.Setting3
				utility.Log("toggling setting 3")
			}
		}
	}

	return &m, nil
}
