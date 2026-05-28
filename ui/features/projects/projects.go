package projects

import (
	"strings"

	"github.com/PorKeat/a8s-tui/api"
)

func KindCounts(items []api.LiveProject) map[string]int {
	counts := map[string]int{}
	for _, project := range items {
		counts[project.Kind]++
	}
	return counts
}

func StatusLabel(project api.LiveProject) string {
	status := strings.TrimSpace(project.Status)
	if status == "" {
		return "Unknown"
	}
	return status
}
