package time_utils

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
)

func ParseDuration(duration string) (time.Duration, error) {
	// Regular expression to match number followed by unit
	re := regexp.MustCompile(`(\d+)([ywdhms])`)
	matches := re.FindAllStringSubmatch(duration, -1)

	if len(matches) == 0 {
		return 0, fmt.Errorf("invalid duration format: %s", duration)
	}

	var total time.Duration
	for _, match := range matches {
		num, err := strconv.Atoi(match[1])
		if err != nil {
			return 0, fmt.Errorf("invalid number: %s", match[1])
		}

		switch match[2] {
		case "y":
			total += time.Duration(num) * time.Hour * 24 * 365
		case "w":
			total += time.Duration(num) * time.Hour * 24 * 7
		case "d":
			total += time.Duration(num) * time.Hour * 24
		case "h":
			total += time.Duration(num) * time.Hour
		case "m":
			total += time.Duration(num) * time.Minute
		case "s":
			total += time.Duration(num) * time.Second
		default:
			return 0, fmt.Errorf("invalid unit: %s", match[2])
		}
	}

	return total, nil
}
