package projects

import (
	"strings"

	"github.com/ITProfessional-Gen01/a8s-cli/api"
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
