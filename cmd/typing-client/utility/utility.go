package utility

import (
	"bufio"
	"crypto/rand"
	"log"
	"math/big"
	"os"
	"strings"

	lg "github.com/charmbracelet/lipgloss"
)

func getRandomWord() string {
	file, err := os.Open("language/words_en")
	var pick string

	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := file.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	scanner := bufio.NewScanner(file)

	lineNum := 1

	for scanner.Scan() {
		line := scanner.Text()
		roll := getRandomValue(lineNum)

		if roll == 0 {
			pick = line
		}

		lineNum += 1
	}

	return pick
}

func getRandomValue(line int) int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(line)))
	if err != nil {
		log.Fatal(err)
	}

	return n.Int64()
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
