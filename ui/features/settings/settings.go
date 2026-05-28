package settings

func ThemeLabel(dark bool) string {
	if dark {
		return "Dark mode"
	}
	return "Light mode"
}
