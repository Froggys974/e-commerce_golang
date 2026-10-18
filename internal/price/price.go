package price

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

var ErrEurosFormat = errors.New("format attendu : 19,99")

func Format(cents int64) string {
	return fmt.Sprintf("%d,%02d €", cents/100, cents%100)
}

func ParseEuros(text string) (int64, error) {
	text = strings.ReplaceAll(strings.TrimSpace(text), ",", ".")
	euros, cents, hasCents := strings.Cut(text, ".")
	if !isDigits(euros) || (hasCents && (!isDigits(cents) || len(cents) > 2)) {
		return 0, ErrEurosFormat
	}
	if len(cents) == 1 {
		cents += "0"
	}

	eurosValue, err := strconv.ParseInt(euros, 10, 64)
	if err != nil || eurosValue > (math.MaxInt64-99)/100 {
		return 0, ErrEurosFormat
	}
	var centsValue int64
	if hasCents {
		centsValue, _ = strconv.ParseInt(cents, 10, 64)
	}
	return eurosValue*100 + centsValue, nil
}

func isDigits(text string) bool {
	if text == "" {
		return false
	}
	for _, character := range text {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
