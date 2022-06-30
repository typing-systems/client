package utility

import (
	lg "github.com/charmbracelet/lipgloss"
)

func GetRandomSentence(words int) string {
	return "the quick brown fox jumps over the really lazy dog"
}

func ForegroundColour(hex string) lg.Style {
	return lg.NewStyle().Foreground(lg.Color(hex))
}

func HalfGen(j int, physicalWidth int, physicalHeight int, hex string) lg.Style {
	return lg.NewStyle().
		Width(physicalWidth / 2).
		Height(physicalHeight).
		Background(lg.Color(hex)).
		Align(lg.Center).
		PaddingTop((physicalHeight - j) / 2)
}
