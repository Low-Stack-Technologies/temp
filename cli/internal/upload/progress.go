package upload

import (
	"fmt"
	"golang.org/x/term"
	"os"
	"strconv"
	"sync"
)

var (
	progressMutex sync.Mutex
	progressBars  []*ProgressReader
)

func DrawAllProgressBars() {
	width, height := getTerminalSize()

	// Save cursor position
	fmt.Print("\033[s")

	// Draw each progress bar
	progressMutex.Lock()
	for _, pr := range progressBars {
		drawProgressBar(pr, width, height)
	}
	progressMutex.Unlock()

	// Restore cursor position
	fmt.Print("\033[u")
}

func drawProgressBar(pr *ProgressReader, width int, height int) {
	// Get terminal width and height
	width, height = getTerminalSize()

	// Ensure minimum width and reasonable margins
	barWidth := width - (len(pr.Filename) + 15) // Add more space for percentage
	if barWidth < 20 {
		barWidth = 20 // Minimum bar width
	}

	// Calculate how many characters should be filled
	filled := int(float64(barWidth) * (pr.Percentage / 100))

	// Create the progress bar
	bar := "["
	for i := 0; i < barWidth; i++ {
		if i < filled {
			bar += "="
		} else {
			bar += " "
		}
	}
	bar += "]"

	// Ensure the index doesn't exceed terminal height
	linePosition := height - pr.Index - 1
	if linePosition < 0 {
		linePosition = 0
	}

	// Print progress bar and percentage
	fmt.Printf("\033[%d;0H", linePosition) // Move cursor to specific line from bottom
	fmt.Printf("\033[K")                   // Clear line
	fmt.Printf("%s: %s %.1f%%", pr.Filename, bar, pr.Percentage)
	fmt.Printf("\033[u") // Restore cursor position
}

func getTerminalSize() (width int, height int) {
	// Try getting size from environment variables (works in tmux)
	if w := os.Getenv("COLUMNS"); w != "" {
		if width, err := strconv.Atoi(w); err == nil {
			if h := os.Getenv("LINES"); h != "" {
				if height, err := strconv.Atoi(h); err == nil {
					return width, height
				}
			}
		}
	}

	// Fallback to term.GetSize
	if w, h, err := term.GetSize(int(os.Stderr.Fd())); err == nil && w > 0 && h > 0 {
		return w, h
	}

	// Last resort fallback
	return 80, 24
}
