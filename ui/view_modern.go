package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/PorKeat/a8s-tui/api"
	"github.com/PorKeat/a8s-tui/ui/features/deploy"
	"github.com/PorKeat/a8s-tui/ui/features/settings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	modernPaneHorizontalPadding = 2
	modernPaneVerticalPadding   = 1
)

func (m model) modernDashboardView(width, height int) tea.View {
	bodyHeight := max(height-2, 18)
	leftWidth := min(max(width*35/100, 36), 58)
	if width < 104 {
		leftWidth = min(max(width*40/100, 30), 46)
	}
	gap := 1
	rightWidth := max(width-leftWidth-gap, 42)

	left := m.modernLeftPane(leftWidth, bodyHeight)
	right := m.modernRightPane(rightWidth, bodyHeight)

	var b strings.Builder
	b.WriteString(modernTopBar(m, width))
	b.WriteString("\n")

	leftLines := strings.Split(left, "\n")
	rightLines := strings.Split(right, "\n")
	for i := 0; i < bodyHeight; i++ {
		var row string
		if i < len(leftLines) {
			row += leftLines[i]
		} else {
			row += backgroundLine("", leftWidth, bgDark)
		}
		row += backgroundLine("", gap, bgDark)
		if i < len(rightLines) {
			row += rightLines[i]
		} else {
			row += backgroundLine("", rightWidth, bgDark)
		}
		b.WriteString(modernFullWidthRow(row, width, bgDark))
		b.WriteString("\n")
	}
	b.WriteString(modernFooter(width))

	view := tea.NewView(b.String())
	view.AltScreen = true
	view.WindowTitle = "A8S TUI"
	return view
}

func (m model) modernLeftPane(width, height int) string {
	contentWidth := modernPaneInnerWidth(width)
	contentHeight := modernPaneInnerHeight(height)
	active := m.focus == focusSidebar || (m.page == pageProjects && m.focus == focusList && !m.projectDetailOpen)
	switch {
	case m.dbFormOpen || m.monolithFormOpen || (m.page == pageDeployment && m.focus != focusSidebar):
		return modernPane(width, height, active, m.modernDeploymentList(contentWidth, contentHeight))
	case m.page == pageProjects && (m.focus == focusList || m.projectDetailOpen || m.deleteConfirmOpen):
		return modernPane(width, height, active, m.modernProjectList(contentWidth, contentHeight))
	default:
		return modernPane(width, height, active, m.modernNavigationList(contentWidth, contentHeight))
	}
}

func (m model) modernRightPane(width, height int) string {
	contentWidth := modernPaneInnerWidth(width)
	contentHeight := modernPaneInnerHeight(height)
	active := m.projectDetailOpen || m.dbFormOpen || m.monolithFormOpen || (m.focus == focusList && m.page != pageProjects)

	switch {
	case m.dbFormOpen:
		return modernPane(width, height, active, modernCropLines(m.renderDashboardDatabaseDeployForm(contentWidth, contentHeight), contentWidth, contentHeight))
	case m.monolithFormOpen:
		return modernPane(width, height, active, modernCropLines(m.renderDashboardMonolithicDeployForm(contentWidth, contentHeight), contentWidth, contentHeight))
	case m.deleteConfirmOpen:
		return modernPane(width, height, true, modernCropLines(m.renderProjectDeleteConfirmation(contentWidth, contentHeight), contentWidth, contentHeight))
	case m.projectDetailOpen:
		return modernPane(width, height, true, m.modernProjectDetail(contentWidth, contentHeight, true))
	case m.page == pageProjects && m.focus == focusList:
		return modernPane(width, height, false, m.modernProjectDetail(contentWidth, contentHeight, false))
	case m.page == pageDeployment && m.focus != focusSidebar:
		return modernPane(width, height, active, m.modernDeploymentDetail(contentWidth, contentHeight))
	case m.page == pageUserSettings && m.focus != focusSidebar:
		return modernPane(width, height, active, m.modernUserSettingsDetail(contentWidth, contentHeight))
	default:
		return modernPane(width, height, active, m.modernSectionDetail(contentWidth, contentHeight))
	}
}

func modernTopBar(m model, width int) string {
	title := lipgloss.NewStyle().
		Background(lipgloss.Color(colorPrimary)).
		Foreground(lipgloss.Color(colorTitle)).
		Bold(true).
		Render(" a8s-cli ")
	count := lipgloss.NewStyle().
		Background(lipgloss.Color(colorPrimary)).
		Foreground(lipgloss.Color("#a7f3d0")).
		Bold(true).
		Render(fmt.Sprintf(" %d projects ", len(m.projects)))
	updated := lipgloss.NewStyle().
		Background(lipgloss.Color(colorPrimary)).
		Foreground(lipgloss.Color(colorMuted)).
		Bold(true).
		Render(" updated " + m.refreshAge() + " ")
	left := title + count + updated
	return left + lipgloss.NewStyle().
		Background(lipgloss.Color(colorPrimary)).
		Render(spaces(max(width-visibleLen(left), 0)))
}

func (m model) refreshAge() string {
	if m.lastRefreshed.IsZero() {
		return "not yet"
	}
	elapsed := time.Since(m.lastRefreshed)
	switch {
	case elapsed < time.Minute:
		return fmt.Sprintf("%ds ago", int(elapsed.Seconds()))
	case elapsed < time.Hour:
		return fmt.Sprintf("%dm ago", int(elapsed.Minutes()))
	default:
		return fmt.Sprintf("%dh ago", int(elapsed.Hours()))
	}
}

func modernFooter(width int) string {
	items := []struct {
		key   string
		label string
	}{
		{"k/up", "prev"},
		{"j/down", "next"},
		{"tab", "focus"},
		{"enter", "open"},
		{"/", "search"},
		{"r", "refresh"},
		{"o", "logout"},
		{"q", "quit"},
	}
	var line string
	for _, item := range items {
		line += modernKeyChip(item.key, item.label) + " "
	}
	return modernLine(line, width, colorBgMain)
}

func modernKeyChip(key, label string) string {
	keyStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgPill)).
		Foreground(lipgloss.Color(colorTitle)).
		Bold(true).
		Padding(0, 1)
	labelStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgMain)).
		Foreground(lipgloss.Color(colorMuted)).
		Padding(0, 1)
	return keyStyle.Render(key) + labelStyle.Render(label)
}

func modernPane(width, height int, active bool, lines []string) string {
	contentWidth := max(width-2, 1)
	contentHeight := max(height-2, 1)
	lines = modernPaddedPaneLines(lines, contentWidth, contentHeight)
	border := colorBorder
	if active {
		border = colorPrimary
	}
	rendered := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgMain)).
		Foreground(lipgloss.Color(colorText)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(border)).
		BorderBackground(lipgloss.Color(colorBgMain)).
		Width(contentWidth).
		Height(contentHeight).
		Render(strings.Join(lines, "\n"))
	return restoreBackground(rendered, bgDark)
}

func modernPaneInnerWidth(width int) int {
	return max(width-2-(modernPaneHorizontalPadding*2), 1)
}

func modernPaneInnerHeight(height int) int {
	return max(height-2-(modernPaneVerticalPadding*2), 1)
}

func modernPaddedPaneLines(lines []string, contentWidth, contentHeight int) []string {
	innerWidth := max(contentWidth-(modernPaneHorizontalPadding*2), 1)
	innerHeight := max(contentHeight-(modernPaneVerticalPadding*2), 1)
	lines = modernCropLines(lines, innerWidth, innerHeight)
	out := make([]string, 0, contentHeight)
	blank := modernLine("", contentWidth, colorBgMain)
	for i := 0; i < modernPaneVerticalPadding && len(out) < contentHeight; i++ {
		out = append(out, blank)
	}
	for _, line := range lines {
		if len(out) >= contentHeight-modernPaneVerticalPadding {
			break
		}
		out = append(out, modernPaddedPaneLine(line, contentWidth))
	}
	for len(out) < contentHeight {
		out = append(out, blank)
	}
	return out
}

func modernPaddedPaneLine(line string, contentWidth int) string {
	left := backgroundLine("", modernPaneHorizontalPadding, bgDark)
	rightWidth := max(contentWidth-modernPaneHorizontalPadding-visibleLen(line), 0)
	right := backgroundLine("", rightWidth, bgDark)
	return left + line + right
}

func modernFullWidthRow(row string, width int, bg string) string {
	if missing := width - visibleLen(row); missing > 0 {
		return row + backgroundLine("", missing, bg)
	}
	return row
}

func modernCropLines(lines []string, width, height int) []string {
	out := make([]string, 0, height)
	for _, line := range lines {
		if len(out) >= height {
			break
		}
		out = append(out, modernLine(line, width, colorBgMain))
	}
	for len(out) < height {
		out = append(out, modernLine("", width, colorBgMain))
	}
	return out
}

func modernLine(content string, width int, bg string) string {
	bgCode := ansiBg(bg)
	content = restoreBackground(content, bgCode)
	return bgCode + strings.ReplaceAll(pad(content, width), reset, reset+bgCode) + reset
}

func modernHeading(text string, width int) string {
	return modernLine(lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgMain)).
		Foreground(lipgloss.Color(colorMuted)).
		Bold(true).
		Render(truncatePlain(strings.ToUpper(text), width)), width, colorBgMain)
}

func modernRule(width int) string {
	return modernLine(lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgMain)).
		Foreground(lipgloss.Color(colorBorder)).
		Render(strings.Repeat("─", width)), width, colorBgMain)
}

func (m model) modernNavigationList(width, height int) []string {
	lines := []string{
		modernHeading("Section", width),
		modernRule(width),
	}
	lastGroup := ""
	for index, item := range m.navigationItems() {
		if item.group != lastGroup {
			lines = append(lines, modernLine("", width, colorBgMain))
			lines = append(lines, modernHeading(item.group, width))
			lastGroup = item.group
		}
		active := index == m.navCursor
		current := item.matchesPage(m.page)
		lines = append(lines, modernNavigationRow(item, width, active, current)...)
	}
	lines = append(lines, modernLine("", width, colorBgMain))
	lines = append(lines, modernMutedLine("enter opens selected section", width))
	return modernCropLines(lines, width, height)
}

func modernNavigationRow(item navigationItem, width int, active, current bool) []string {
	rowBg := colorBgMain
	labelStyle := mainBodyStyle(rowBg)
	border := colorBorder
	prefix := "  "
	if current {
		labelStyle = mainPrimaryStyle(rowBg)
	}
	if active {
		rowBg = colorBgMain
		border = colorPrimary
		labelStyle = mainTitleStyle(rowBg)
		prefix = "> "
	}
	boxWidth := max(width-4, 12)
	labelWidth := max(boxWidth-4, 8)
	content := labelStyle.Render(prefix + truncatePlain(item.label, labelWidth))
	box := lipgloss.NewStyle().
		Background(lipgloss.Color(rowBg)).
		Foreground(lipgloss.Color(colorText)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(border)).
		BorderBackground(lipgloss.Color(colorBgMain)).
		Padding(0, 1).
		Width(boxWidth).
		Render(content)
	rendered := strings.Split(box, "\n")
	lines := make([]string, 0, len(rendered))
	for _, line := range rendered {
		lines = append(lines, modernLine(line, width, colorBgMain))
	}
	return lines
}

func (m model) modernProjectList(width, height int) []string {
	lines := []string{
		modernTableHeader("PROJECT", "STATUS", width),
		modernRule(width),
	}
	visible := m.visibleProjects()
	if len(visible) == 0 {
		lines = append(lines, modernMutedLine("No projects returned by backend", width))
		return modernCropLines(lines, width, height)
	}
	rowCount := max(height-len(lines), 1)
	start := 0
	if m.cursor >= rowCount {
		start = m.cursor - rowCount + 1
	}
	for index := start; index < len(visible) && len(lines) < height; index++ {
		lines = append(lines, modernProjectRow(visible[index], width, index == m.cursor))
	}
	return modernCropLines(lines, width, height)
}

func modernTableHeader(left, right string, width int) string {
	leftStyle := mainMutedStyle(colorBgMain).Bold(true)
	rightStyle := mainMutedStyle(colorBgMain).Bold(true)
	leftText := leftStyle.Render(left)
	rightText := rightStyle.Render(right)
	return modernLine(leftText+spaces(max(width-visibleLen(leftText)-visibleLen(rightText), 1))+rightText, width, colorBgMain)
}

func modernProjectRow(project api.LiveProject, width int, active bool) string {
	rowBg := colorBgMain
	nameStyle := mainBodyStyle(rowBg)
	statusStyle := lipgloss.NewStyle().Background(lipgloss.Color(rowBg)).Foreground(lipgloss.Color(statusColor(project.Status))).Bold(true)
	prefix := "  "
	if active {
		rowBg = colorBgActive
		nameStyle = mainTitleStyle(rowBg)
		statusStyle = lipgloss.NewStyle().Background(lipgloss.Color(rowBg)).Foreground(lipgloss.Color("#4ade80")).Bold(true)
		prefix = "> "
	}
	statusWidth := 12
	nameWidth := max(width-statusWidth-3, 8)
	name := nameStyle.Render(prefix + truncatePlain(project.Name, nameWidth))
	status := statusStyle.Render(truncatePlain(firstNonEmpty(project.Status, "unknown"), statusWidth))
	line := name + spaces(max(width-visibleLen(name)-visibleLen(status), 1)) + status
	return lipgloss.NewStyle().Background(lipgloss.Color(rowBg)).Render(pad(line, width))
}

func (m model) modernDeploymentList(width, height int) []string {
	lines := []string{
		modernTableHeader("DEPLOYMENT", "STATE", width),
		modernRule(width),
	}
	for index, feature := range deploy.Features {
		lines = append(lines, modernDeploymentRow(feature, width, index == m.deployCursor)...)
	}
	lines = append(lines, modernLine("", width, colorBgMain))
	lines = append(lines, modernMutedLine("enter opens selected deployment", width))
	return modernCropLines(lines, width, height)
}

func modernDeploymentRow(feature deploy.Feature, width int, active bool) []string {
	rowBg := colorBgMain
	labelStyle := mainBodyStyle(rowBg)
	stateStyle := mainMutedStyle(rowBg)
	border := colorBorder
	prefix := "  "
	if active {
		border = colorPrimary
		labelStyle = mainTitleStyle(rowBg)
		stateStyle = mainPrimaryStyle(rowBg)
		prefix = "> "
	}
	state := "soon"
	if feature.Ready {
		state = "ready"
	}
	boxWidth := max(width-4, 12)
	labelWidth := max(boxWidth-10, 8)
	label := labelStyle.Render(prefix + truncatePlain(feature.Label, labelWidth))
	status := stateStyle.Render(state)
	line := label + spaces(max(boxWidth-visibleLen(label)-visibleLen(status), 1)) + status
	box := lipgloss.NewStyle().
		Background(lipgloss.Color(rowBg)).
		Foreground(lipgloss.Color(colorText)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(border)).
		BorderBackground(lipgloss.Color(colorBgMain)).
		Padding(0, 1).
		Width(boxWidth).
		Render(line)
	rendered := strings.Split(box, "\n")
	lines := make([]string, 0, len(rendered))
	for _, line := range rendered {
		lines = append(lines, modernLine(line, width, colorBgMain))
	}
	return lines
}

func (m model) modernProjectDetail(width, height int, actions bool) []string {
	project, ok := m.selectedProject()
	if !ok {
		return modernCropLines([]string{
			modernTitleLine("Project workspace", width),
			modernMutedLine("No project selected", width),
		}, width, height)
	}
	status := firstNonEmpty(project.Status, "unknown")
	lines := []string{
		modernTitleLine(project.Name, width),
		modernLine(lipgloss.NewStyle().Background(lipgloss.Color(colorBgMain)).Foreground(lipgloss.Color(statusColor(status))).Bold(true).Render("● "+strings.ToLower(status)), width, colorBgMain),
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
	}
	fields := []struct {
		label string
		value string
	}{
		{"Kind", project.Kind},
		{"Engine", firstNonEmpty(project.Engine, "n/a")},
		{"Mode", firstNonEmpty(project.DeploymentMode, project.ArchitectureType, "n/a")},
		{"Namespace", firstNonEmpty(project.Namespace, "n/a")},
		{"Version", firstNonEmpty(project.Version, "n/a")},
		{"Repo", firstNonEmpty(project.RepoFullName, project.RepoURL, "n/a")},
		{"Branch", firstNonEmpty(project.Branch, "n/a")},
		{"Updated", firstNonEmpty(shortTime(project.UpdatedAt), "n/a")},
	}
	for _, field := range fields {
		if field.value == "" {
			continue
		}
		lines = append(lines, modernFieldLine(field.label, field.value, width))
	}
	lines = append(lines, modernLine("", width, colorBgMain))
	lines = append(lines, modernRule(width))
	lines = append(lines, modernLine("", width, colorBgMain))
	if actions && api.ProjectKindSupportsDelete(project.Kind) {
		lines = append(lines, modernLine(modernActionButtons(m.projectDetailButton), width, colorBgMain))
	} else {
		lines = append(lines, modernMutedLine("enter opens detail actions", width))
	}
	return modernCropLines(lines, width, height)
}

func (m model) modernDeploymentDetail(width, height int) []string {
	feature := m.selectedDeploymentFeature()
	lines := []string{
		modernTitleLine(feature.Label, width),
		modernLine(lipgloss.NewStyle().Background(lipgloss.Color(colorBgMain)).Foreground(lipgloss.Color("#4ade80")).Bold(true).Render("● "+deploymentStatusText(feature.Ready)), width, colorBgMain),
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
		modernFieldLine("Description", feature.Description, width),
		modernFieldLine("Shortcut", "enter", width),
		modernFieldLine("Status", deploymentStatusText(feature.Ready), width),
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
		modernMutedLine("Use the list on the left to select a deployment type.", width),
	}
	return modernCropLines(lines, width, height)
}

func (m model) modernUserSettingsDetail(width, height int) []string {
	lines := []string{
		modernTitleLine("User Settings", width),
		modernLine(lipgloss.NewStyle().Background(lipgloss.Color(colorBgMain)).Foreground(lipgloss.Color(colorPrimary)).Bold(true).Render("● appearance"), width, colorBgMain),
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
		modernFieldLine("Theme", m.themeLabel(), width),
		modernLine("", width, colorBgMain),
	}
	lines = append(lines, modernThemeChoiceLines(m.themeIndex, width)...)
	lines = append(lines,
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
		modernMutedLine("Press enter, space, or t to switch theme.", width),
	)
	return modernCropLines(lines, width, height)
}

func modernThemeChoiceLines(current int, width int) []string {
	labels := settings.ThemeLabels()
	lines := make([]string, 0, 2)
	var row string
	gap := mainBodyStyle(colorBgMain).Render("  ")
	for index, label := range labels {
		chip := modernThemeChip(label, index == settings.NormalizeThemeIndex(current))
		next := chip
		if row != "" {
			next = row + gap + chip
		}
		if row != "" && visibleLen(next) > width {
			lines = append(lines, modernLine(row, width, colorBgMain))
			row = chip
			continue
		}
		row = next
	}
	if row != "" {
		lines = append(lines, modernLine(row, width, colorBgMain))
	}
	if len(lines) == 0 {
		lines = append(lines, modernMutedLine("No themes available", width))
	}
	return lines
}

func modernThemeChip(label string, selected bool) string {
	bg := colorBgPill
	fg := colorText
	prefix := "  "
	if selected {
		bg = colorPrimary
		fg = colorTitle
		prefix = "> "
	}
	return lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(fg)).
		Bold(selected).
		Padding(0, 2).
		Render(prefix + label)
}

func (m model) modernSectionDetail(width, height int) []string {
	item := m.selectedNavigationItem()
	lines := []string{
		modernTitleLine(item.label, width),
		modernLine(lipgloss.NewStyle().Background(lipgloss.Color(colorBgMain)).Foreground(lipgloss.Color(colorPrimary)).Bold(true).Render("● selected"), width, colorBgMain),
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
		modernFieldLine("Section", item.group, width),
		modernFieldLine("Auth", m.stateLabelPlain(), width),
		modernFieldLine("Projects", fmt.Sprintf("%d live", len(m.projects)), width),
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
		modernMutedLine("Press enter to open "+item.label+".", width),
	}
	return modernCropLines(lines, width, height)
}

func modernTitleLine(title string, width int) string {
	return modernLine(lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgMain)).
		Foreground(lipgloss.Color(colorTitle)).
		Bold(true).
		Render(truncatePlain(title, width)), width, colorBgMain)
}

func modernMutedLine(text string, width int) string {
	return modernLine(lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgMain)).
		Foreground(lipgloss.Color(colorMuted)).
		Render(truncatePlain(text, width)), width, colorBgMain)
}

func modernFieldLine(label, value string, width int) string {
	labelWidth := min(18, max(width/3, 8))
	valueWidth := max(width-labelWidth-2, 4)
	labelStyle := mainMutedStyle(colorBgMain).Bold(true)
	valueStyle := mainTitleStyle(colorBgMain)
	content := labelStyle.Render(pad(truncatePlain(label, labelWidth), labelWidth)) +
		mainBodyStyle(colorBgMain).Render("  ") +
		valueStyle.Render(truncatePlain(value, valueWidth))
	return modernLine(content, width, colorBgMain)
}

func modernActionButtons(selected int) string {
	deleteButton := modernButton("Delete", selected == 0, true)
	cancelButton := modernButton("Cancel", selected == 1, false)
	return deleteButton + mainBodyStyle(colorBgMain).Render("  ") + cancelButton
}

func modernButton(label string, selected, danger bool) string {
	bg := colorBgPill
	fg := colorText
	if danger {
		bg = "#3f2638"
		fg = "#fb7185"
	}
	if selected {
		if danger {
			bg = "#6f2d4a"
		} else {
			bg = colorPrimary
			fg = colorTitle
		}
		label = "> " + label
	} else {
		label = "  " + label
	}
	return lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(fg)).
		Bold(selected || danger).
		Padding(0, 2).
		Render(label)
}

func statusColor(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "READY", "RUNNING", "DEPLOYED", "HEALTHY", "SUCCESS", "SUCCEEDED":
		return "#4ade80"
	case "PENDING", "STARTING", "PROVISIONING", "DEPLOYING":
		return "#facc15"
	case "FAILED", "ERROR", "UNHEALTHY":
		return "#fb7185"
	default:
		return colorMuted
	}
}

func (m model) stateLabelPlain() string {
	switch m.state {
	case stateConfigError:
		return "config error"
	case stateLoggedOut:
		return "logged out"
	case stateLoggingIn:
		return "logging in"
	case stateLoading:
		return "loading"
	case stateReady:
		return "ready"
	case stateError:
		return "error"
	default:
		return "unknown"
	}
}
