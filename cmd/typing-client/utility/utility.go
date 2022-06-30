package utility

import (
	"bufio"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	lg "github.com/charmbracelet/lipgloss"
)

func getRandomWord() string {
	file, err := os.Open("language/words_en")
	var pick string

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	randSource := rand.NewSource(time.Now().UnixNano())
	randGenerator := rand.New(randSource)

	lineNum := 1

	for scanner.Scan() {
		line := scanner.Text()
		roll := randGenerator.Intn(lineNum)
		if roll == 0 {
			pick = line
		}

		lineNum += 1
	}

	return pick
}

func GetRandomSentence(words int) string {
	arr := make([]string, 5)

	for j := 0; j <= 10; j++ {
		arr = append(arr, getRandomWord())
	}

	return strings.Join(arr, " ")
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
