package ui

import (
	"a8s-tui/api"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) renderDashboardDeployment(width, height int) []string {
	lines := make([]string, 0, height)
	lines = append(lines, dashboardHeader("Deployment", m.pageLead(), width)...)
	lines = append(lines, deploymentFeatureCard(width, m.deployCursor)...)
	lines = append(lines, mainLine("", width))
	helpText := "Press enter to open the selected deployment type."
	if m.page == pageDeployment {
		helpText = "Press esc to close the selected deployment type."
	}
	lines = append(lines, mainTextLine(helpText, mainMutedStyle(colorBgMain), width))
	return fillStyled(lines, bgDark, width, height)
}

func (m model) renderDashboardDatabaseDeployForm(width, height int) []string {
	lines := make([]string, 0, height)
	lines = append(lines, dashboardHeader("Single database", "Create a single-instance database deployment.", width)...)
	lines = append(lines, databaseDeployFormCard(m, width)...)
	return fillStyled(lines, bgDark, width, height)
}

func deploymentFeatureCard(width int, activeIndex int) []string {
	cardWidth := max(width-6, 30)
	card := styleCard.Width(cardWidth)
	title := mainTitleStyle(colorBgCard)
	lines := []string{
		cardContentLine(card, "", width),
		cardContentLine(card, "  "+title.Render("Deployment"), width),
		cardContentLine(card, "", width),
	}
	for index, feature := range deploymentFeatures {
		lines = append(lines, deploymentFeatureBox(card, cardWidth, width, feature, index == activeIndex)...)
		lines = append(lines, cardContentLine(card, "", width))
	}
	return lines
}

func deploymentFeatureBox(card lipgloss.Style, cardWidth, width int, feature deploymentFeature, active bool) []string {
	rowBg := lipgloss.Color(colorBgCard)
	border := lipgloss.Color(colorBorder)
	labelStyle := mainBodyStyle(colorBgCard)
	mutedStyle := mainMutedStyle(colorBgCard)
	cursorStyle := mainBodyStyle(colorBgCard)
	cursor := "  "
	if active {
		border = lipgloss.Color(colorPrimary)
		labelStyle = mainPrimaryStyle(colorBgCard)
		cursorStyle = mainPrimaryStyle(colorBgCard)
		cursor = "> "
	}
	description := feature.description
	if !feature.ready {
		description += " Coming soon."
	}
	boxWidth := max(cardWidth-4, 24)
	labelWidth := max(boxWidth-34, 12)
	descWidth := max(boxWidth-labelWidth-8, 8)
	content := cursorStyle.Render(cursor) +
		labelStyle.Render(pad(truncatePlain(feature.label, labelWidth), labelWidth)) +
		mainBodyStyle(colorBgCard).Render("  ") +
		mutedStyle.Render(truncatePlain(description, descWidth))
	box := lipgloss.NewStyle().
		Background(rowBg).
		Foreground(lipgloss.Color(colorText)).
		Width(boxWidth).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		BorderBackground(rowBg).
		Render(content)
	rendered := strings.Split(box, "\n")
	lines := make([]string, 0, len(rendered))
	for _, line := range rendered {
		lines = append(lines, cardContentLine(card, "  "+line, width))
	}
	return lines
}

func databaseDeployFormCard(m model, width int) []string {
	cardWidth := max(width-6, 42)
	card := styleCard.Width(cardWidth)
	title := mainTitleStyle(colorBgCard)
	bodyStyle := mainBodyStyle(colorBgCard)
	mutedStyle := mainMutedStyle(colorBgCard)
	lines := []string{
		cardContentLine(card, "", width),
		cardContentLine(card, "  "+title.Render("Deploy Database")+bodyStyle.Render("  ")+mutedStyle.Render("single-instance"), width),
		cardContentLine(card, "  "+mutedStyle.Render("Use arrows or j/k to move. Left/right changes choices."), width),
		cardContentLine(card, "", width),
	}
	fields := []struct {
		index int
		label string
		value string
	}{
		{0, "Project name", m.dbForm.projectName},
		{1, "Engine", m.dbForm.engine().label},
		{2, "Database name", m.dbForm.databaseName},
		{3, "Username", m.dbForm.username},
		{4, "Password", maskValue(m.dbForm.password)},
		{5, "Version", m.dbForm.version()},
		{6, "Size", m.dbForm.size()},
	}
	for index, field := range fields {
		lines = append(lines, databaseDeployFieldBox(card, cardWidth, width, m, field.index, field.label, field.value)...)
		if index == len(fields)-1 {
			lines = append(lines, cardContentLine(card, "", width))
		}
	}
	lines = append(lines, databaseDeploySubmitBox(card, cardWidth, width, m)...)
	lines = append(lines, cardContentLine(card, "", width))
	payload := fmt.Sprintf("Payload: %s / single-instance / %s / %s", m.dbForm.engine().id, m.dbForm.version(), m.dbForm.size())
	lines = append(lines, cardContentLine(card, "  "+mutedStyle.Render(truncatePlain(payload, cardWidth-4)), width))
	if m.message != "" {
		lines = append(lines, cardContentLine(card, "  "+bodyStyle.Render("* "+truncatePlain(m.message, cardWidth-4)), width))
	}
	lines = append(lines, cardContentLine(card, "", width))
	return lines
}

func databaseDeployFieldBox(card lipgloss.Style, cardWidth, width int, m model, index int, label, value string) []string {
	active := m.dbForm.focus == index
	if value == "" {
		value = "..."
	}
	rowBg := lipgloss.Color(colorBgCard)
	border := lipgloss.Color(colorBorder)
	labelStyle := mainMutedStyle(colorBgCard)
	valueStyle := mainBodyStyle(colorBgCard)
	prefixStyle := mainBodyStyle(colorBgCard)
	prefix := "  "
	if active {
		border = lipgloss.Color(colorPrimary)
		labelStyle = mainPrimaryStyle(colorBgCard)
		valueStyle = mainTitleStyle(colorBgCard)
		prefixStyle = mainPrimaryStyle(colorBgCard)
		prefix = "> "
	}
	boxWidth := max(cardWidth-4, 24)
	labelWidth := 16
	valueWidth := max(boxWidth-labelWidth-8, 8)
	content := prefixStyle.Render(prefix) +
		labelStyle.Render(pad(truncatePlain(label, labelWidth), labelWidth)) +
		mainBodyStyle(colorBgCard).Render("  ") +
		valueStyle.Render(truncatePlain(value, valueWidth))
	box := lipgloss.NewStyle().
		Background(rowBg).
		Foreground(lipgloss.Color(colorText)).
		Width(boxWidth).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		BorderBackground(rowBg).
		Render(content)
	rendered := strings.Split(box, "\n")
	lines := make([]string, 0, len(rendered))
	for _, line := range rendered {
		lines = append(lines, cardContentLine(card, "  "+line, width))
	}
	return lines
}

func databaseDeploySubmitBox(card lipgloss.Style, cardWidth, width int, m model) []string {
	active := m.dbForm.focus == 7
	rowBg := lipgloss.Color(colorBgCard)
	border := lipgloss.Color(colorBorder)
	labelStyle := mainBodyStyle(colorBgCard)
	if active {
		border = lipgloss.Color(colorPrimary)
		labelStyle = mainPrimaryStyle(colorBgCard)
	}
	label := labelStyle.Render("Deploy")
	box := lipgloss.NewStyle().
		Background(rowBg).
		Foreground(lipgloss.Color(colorText)).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		BorderBackground(rowBg).
		Render(label)
	rendered := strings.Split(box, "\n")
	lines := make([]string, 0, len(rendered))
	for _, line := range rendered {
		lines = append(lines, cardContentLine(card, "  "+line, width))
	}
	return lines
}

func (m model) databaseDeployView(width, height int) tea.View {
	contentWidth := min(max(width-12, 72), 104)
	leftMargin := max((width-contentWidth)/2, 0)
	lines := make([]string, 0, height)
	lines = append(lines, "")
	lines = append(lines, spaces(leftMargin)+bold+fgLogo+"Deploy Database"+reset+fgMuted+"  single-instance"+reset)
	lines = append(lines, spaces(leftMargin)+fgMuted+"Use arrows or j/k to move. Left/right changes choices. Enter submits on Deploy."+reset)
	lines = append(lines, "")
	lines = append(lines, spaces(leftMargin)+m.databaseFormLine(0, "Project name", m.dbForm.projectName, contentWidth))
	lines = append(lines, spaces(leftMargin)+m.databaseFormLine(1, "Engine", m.dbForm.engine().label, contentWidth))
	lines = append(lines, spaces(leftMargin)+m.databaseFormLine(2, "Database name", m.dbForm.databaseName, contentWidth))
	lines = append(lines, spaces(leftMargin)+m.databaseFormLine(3, "Username", m.dbForm.username, contentWidth))
	lines = append(lines, spaces(leftMargin)+m.databaseFormLine(4, "Password", maskValue(m.dbForm.password), contentWidth))
	lines = append(lines, spaces(leftMargin)+m.databaseFormLine(5, "Version", m.dbForm.version(), contentWidth))
	lines = append(lines, spaces(leftMargin)+m.databaseFormLine(6, "Size", m.dbForm.size(), contentWidth))
	lines = append(lines, "")
	lines = append(lines, spaces(leftMargin)+m.databaseSubmitLine(contentWidth))
	lines = append(lines, "")
	lines = append(lines, spaces(leftMargin)+fgMuted+"Payload: "+reset+fgText+m.dbForm.engine().id+" / single-instance / "+m.dbForm.version()+" / "+m.dbForm.size()+reset)
	if m.message != "" {
		lines = append(lines, spaces(leftMargin)+fgWarn+"* "+fgAccent+m.message+reset)
	}

	for len(lines) < height-1 {
		lines = append(lines, "")
	}
	var b strings.Builder
	for _, line := range lines[:height-1] {
		b.WriteString(backgroundLine(line, width, bgDark))
		b.WriteString("\n")
	}
	b.WriteString(m.databaseFormStatusline(width))
	view := tea.NewView(b.String())
	view.AltScreen = true
	view.WindowTitle = "A8S TUI - Deploy Database"
	return view
}

func (m model) databaseDeployLogView(width, height int) tea.View {
	contentWidth := min(max(width-10, 76), 132)
	leftMargin := max((width-contentWidth)/2, 0)
	logHeight := max(height-10, 8)
	deployment := m.deployLog
	title := firstNonEmpty(deployment.ProjectName, deployment.ReleaseName, "Database deployment")
	status := firstNonEmpty(deployment.Status, "PENDING")
	statusColor := fgAccent
	if api.DatabaseDeploymentTerminal(status) && !api.DatabaseDeploymentFailed(status) {
		statusColor = fgGreen
	}
	if api.DatabaseDeploymentFailed(status) {
		statusColor = fgError
	}

	lines := make([]string, 0, height)
	lines = append(lines, "")
	lines = append(lines, spaces(leftMargin)+bold+fgLogo+"Deploy Database"+reset+fgMuted+"  logs"+reset)
	lines = append(lines, spaces(leftMargin)+fgMuted+"Project "+reset+fgText+truncatePlain(title, contentWidth-18)+reset)
	lines = append(lines, spaces(leftMargin)+fgMuted+"Status  "+reset+statusColor+status+reset+deployStatusSuffix(deployment, contentWidth))
	lines = append(lines, "")
	lines = append(lines, spaces(leftMargin)+paneTitle("view logs", contentWidth, true))

	logLines := api.ParseDeploymentLogLines(deployment.StatusLog)
	if len(logLines) == 0 {
		logLines = []string{
			"Submitting deployment request...",
			"Waiting for backend deployment logs...",
		}
	}
	start := clamp(m.deployLogOffset, 0, max(len(logLines)-1, 0))
	if len(logLines)-start < logHeight {
		start = max(len(logLines)-logHeight, 0)
	}
	for index := start; index < len(logLines) && len(lines) < height-2; index++ {
		prefix := fgMuted + fmt.Sprintf("%03d", index+1) + reset + fgAccent + " | " + reset
		text := fgText + truncatePlain(logLines[index], contentWidth-visibleLen(prefix)-3) + reset
		lines = append(lines, spaces(leftMargin)+bgPane+pad(" "+prefix+text, contentWidth)+reset)
	}
	for len(lines) < height-2 {
		lines = append(lines, spaces(leftMargin)+bgPane+pad("", contentWidth)+reset)
	}
	if m.message != "" {
		lines = append(lines, spaces(leftMargin)+fgWarn+"* "+fgAccent+truncatePlain(m.message, contentWidth-2)+reset)
	}
	for len(lines) < height-1 {
		lines = append(lines, "")
	}

	var b strings.Builder
	for _, line := range lines[:height-1] {
		b.WriteString(backgroundLine(line, width, bgDark))
		b.WriteString("\n")
	}
	b.WriteString(m.databaseDeployLogStatusline(width))
	view := tea.NewView(b.String())
	view.AltScreen = true
	view.WindowTitle = "A8S TUI - Deployment Logs"
	return view
}

func deployStatusSuffix(deployment api.DatabaseDeploymentRecord, width int) string {
	message := firstNonEmpty(deployment.StatusMessage, deployment.ID)
	if message == "" {
		return ""
	}
	return fgMuted + "  " + truncatePlain(message, max(width-18, 12)) + reset
}

func (m model) databaseDeployLogStatusline(width int) string {
	left := bgBar + bold + fgLogo + " deploy logs " + reset
	right := bgBar + fgMuted + " arrows/jk scroll  r refresh  b/esc projects  q quit " + reset
	return left + fill(width-visibleLen(left)-visibleLen(right), bgBar+" "+reset) + right
}
