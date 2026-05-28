package components

import (
	"strings"
	"unicode/utf8"
)

func MaskValue(value string) string {
	if value == "" {
		return ""
	}
	return strings.Repeat("*", max(4, utf8.RuneCountInString(value)))
}

func Pad(text string, width int) string {
	if width <= 0 {
		return ""
	}
	plainWidth := VisibleLen(text)
	if plainWidth >= width {
		return text
	}
	return text + strings.Repeat(" ", width-plainWidth)
}

func Spaces(width int) string {
	if width <= 0 {
		return ""
	}
	return strings.Repeat(" ", width)
}

func VisibleLen(text string) int {
	length := 0
	inEscape := false
	for _, r := range text {
		if r == '\x1b' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		length++
	}
	return length
}

func TruncatePlain(text string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= width {
		return text
	}
	if width <= 1 {
		return string(runes[:width])
	}
	return string(runes[:width-1]) + "."
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
