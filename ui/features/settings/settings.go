package settings

import "github.com/PorKeat/a8s-tui/ui/theme"

func ThemeLabel(index int) string {
	return theme.ThemeLabel(index)
}

func ThemeLabels() []string {
	return theme.ThemeLabels()
}

func NormalizeThemeIndex(index int) int {
	return theme.NormalizeThemeIndex(index)
}
