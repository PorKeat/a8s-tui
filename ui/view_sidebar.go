package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

func (m model) selectedNavigationItem() navigationItem {
	items := m.navigationItems()
	if len(items) == 0 {
		return navigationItem{group: "Workspace", label: "Projects", key: "p", page: pageProjects}
	}
	return items[clamp(m.navCursor, 0, len(items)-1)]
}

func (m model) currentPageNavigationItem() navigationItem {
	items := m.navigationItems()
	if len(items) == 0 {
		return navigationItem{group: "Workspace", label: "Projects", key: "p", page: pageProjects}
	}
	for _, item := range items {
		if item.matchesPage(m.page) {
			return item
		}
	}
	return items[0]
}

func (m model) selectedNavigationGroup() string {
	return m.selectedNavigationItem().group
}

func (m model) renderSidebar(width, height int) []string {
	lines := make([]string, 0, height)
	lines = append(lines, paneTitle("navigation", width, m.focus == focusSidebar))
	lines = append(lines, bgPane+pad("", width)+reset)

	lastGroup := ""
	for index, item := range m.navigationItems() {
		if item.group != lastGroup {
			if lastGroup != "" {
				lines = append(lines, bgPane+pad("", width)+reset)
			}
			lines = append(lines, sectionLine(strings.ToLower(item.group), width))
			lastGroup = item.group
		}
		active := index == m.navCursor
		current := item.matchesPage(m.page)
		lines = append(lines, navigationLine(item, width, active, current))
	}

	if m.page == pageProjects {
		counts := m.kindCounts()
		lines = append(lines, bgPane+pad("", width)+reset)
		lines = append(lines, sectionLine("project counts", width))
		rows := []struct {
			label string
			count int
		}{
			{"all", len(m.projects)},
			{"monolith", counts["monolith"]},
			{"microservices", counts["microservices"]},
			{"database", counts["database"]},
			{"dbcluster", counts["dbcluster"]},
		}
		for _, row := range rows {
			left := fgText + "  " + row.label + reset
			right := fgMuted + fmt.Sprintf("%d  ", row.count) + reset
			lines = append(lines, bgPane+pad(left+spaces(width-visibleLen(left)-visibleLen(right))+right, width)+reset)
		}
	}

	lines = append(lines, bgPane+pad("", width)+reset)
	lines = append(lines, bgPane+pad("", width)+reset)
	lines = append(lines, sectionLine("message", width))
	lines = append(lines, wrapPaneText(m.message, width, fgMuted)...)
	return fillPane(lines, width, height)
}

func (m model) renderDashboardSidebar(width, height int) []string {
	lines := make([]string, 0, height)
	lines = append(lines, sidebarBrand(width)...)
	lines = append(lines, sideLine("", width))
	lines = append(lines, searchBox(width, m.sidebarSearchText())...)
	lines = append(lines, sideLine("", width))
	lastGroup := ""
	for index, item := range m.navigationItems() {
		if item.group != lastGroup {
			if lastGroup != "" {
				lines = append(lines, sideLine("", width))
			}
			lines = append(lines, sideSectionLine(strings.ToUpper(item.group), width))
			if item.group == "Workspace" {
				username := "My"
				if m.userName != "" {
					username = m.userName + "'s"
				}
				lines = append(lines, sideLine(username+" Workspace", width))
			}
			lastGroup = item.group
		}
		active := index == m.navCursor
		current := item.matchesPage(m.page)
		lines = append(lines, sidebarNavLine(item, width, active, current)...)
	}
	return fillStyled(lines, bgSide, width, height)
}

func (m model) sidebarSearchText() string {
	if m.filtering {
		if strings.TrimSpace(m.filter) == "" {
			return "/ Search projects..."
		}
		return "/ " + m.filter
	}
	if strings.TrimSpace(m.filter) != "" {
		return m.filter
	}
	return "Press / to filter"
}

func sidebarBrand(width int) []string {
	title := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgSide)).
		Foreground(lipgloss.Color(colorPrimary)).
		Bold(true).
		Render("AUTONOMOUS")
	line := styleSide.Render("  ") + title
	rule := styleSide.Render("  ") +
		lipgloss.NewStyle().
			Background(lipgloss.Color(colorPrimary)).
			Foreground(lipgloss.Color(colorPrimary)).
			Render(strings.Repeat(" ", max(width-4, 1)))
	return []string{
		sideContentLine(line, width),
		sideContentLine(rule, width),
	}
}

func sideSectionLine(text string, width int) string {
	label := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgSide)).
		Foreground(lipgloss.Color(colorMuted)).
		Bold(true).
		Render("  " + text)
	return sideContentLine(label, width)
}

func sidebarNavLine(item navigationItem, width int, active, current bool) []string {
	rowBg := lipgloss.Color(colorBgSide)
	labelStyle := styleSideText
	keyStyle := styleSideMuted
	iconBg := colorBgSide
	markerStyle := lipgloss.NewStyle().
		Background(rowBg).
		Foreground(lipgloss.Color(colorText))
	prefix := markerStyle.Render("   ")
	if active {
		rowBg = lipgloss.Color(colorBgActive)
		iconBg = colorBgActive
		markerStyle = lipgloss.NewStyle().
			Background(rowBg).
			Foreground(lipgloss.Color(colorPrimary))
		rowSpaceStyle := lipgloss.NewStyle().
			Background(rowBg).
			Foreground(lipgloss.Color(colorText))
		prefix = rowSpaceStyle.Render(" ") + markerStyle.Render("▌") + rowSpaceStyle.Render(" ")
		labelStyle = lipgloss.NewStyle().
			Background(rowBg).
			Foreground(lipgloss.Color(colorPrimary)).
			Bold(true)
		keyStyle = lipgloss.NewStyle().
			Background(rowBg).
			Foreground(lipgloss.Color(colorPrimary)).
			Bold(true)
	} else if current {
		markerStyle = lipgloss.NewStyle().
			Background(rowBg).
			Foreground(lipgloss.Color(colorText))
		prefix = markerStyle.Render(" ") +
			lipgloss.NewStyle().
				Background(rowBg).
				Foreground(lipgloss.Color(colorPrimary)).
				Render("•") +
			markerStyle.Render(" ")
		labelStyle = lipgloss.NewStyle().
			Background(rowBg).
			Foreground(lipgloss.Color(colorTitle)).
			Bold(true)
	}
	rowPad := lipgloss.NewStyle().
		Background(rowBg).
		Foreground(lipgloss.Color(colorText))
	left := prefix +
		dashboardItemIcon(item, iconBg) +
		rowPad.Render("  ") +
		labelStyle.Render(truncatePlain(item.label, max(width-15, 4)))
	right := keyStyle.Render(item.key + "  ")
	contentWidth := max(width-2, 8)
	lead := styleSide.Render(" ")
	if active {
		contentWidth = max(width, 8)
		lead = ""
	}
	gap := rowPad.Render(spaces(max(contentWidth-visibleLen(left)-visibleLen(right), 0)))
	box := lipgloss.NewStyle().
		Background(rowBg).
		Foreground(lipgloss.Color(colorText)).
		Width(contentWidth).
		Render(left + gap + right)
	return []string{sideContentLine(lead+box, width)}
}
