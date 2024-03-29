package utility

import (
	"bufio"
	"crypto/rand"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	lg "github.com/charmbracelet/lipgloss"
)

// Generates a random word in constant memory and O(n) time.
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

		lineNum++
	}

	return pick
}

// Generates a random value [0,n) and returns it as an int64.
func getRandomValue(line int) int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(line)))
	if err != nil {
		log.Fatal(err)
	}

	return n.Int64()
}

// Creates an array and populates it with randomly generated words, returns a string.
func GetRandomSentence(words int) string {
	arr := make([]string, 0)

	for i := 0; i < words; i++ {
		arr = append(arr, getRandomWord())
	}

	return strings.Join(arr, " ")
}

// Serves as a utility class for syntactic sugar, returns a lip gloss style.
func ForegroundColour(hex string) lg.Style {
	return lg.NewStyle().Foreground(lg.Color(hex))
}

// Serves as a utility class for syntactic sugar, returns a lip gloss style.
func HalfGen(numVertLines int, physicalWidth int, physicalHeight int, hex string) lg.Style {
	return lg.NewStyle().
		Width(physicalWidth / 2).
		Height(physicalHeight).
		Background(lg.Color(hex)).
		Align(lg.Center).
		PaddingTop((physicalHeight - numVertLines) / 2)
}

func CalculateStats(correctStrokes float64, strokes int, startTime time.Time) (float64, float64, float64) {
	var cpm = (correctStrokes / time.Since(startTime).Minutes())
	var wpm = (correctStrokes / 5) / (time.Since(startTime).Minutes())
	var accuracy = ((correctStrokes / float64(strokes)) * 100)

	return cpm, wpm, accuracy
}

func Log(text string) {
	f, err := os.OpenFile("./client.log", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	if _, err = f.WriteString(time.Now().Format("01-02-2006 15:04:05.000000		") + text + "\n"); err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		log.Fatalf("error closing file: %v", err)
	}
}
