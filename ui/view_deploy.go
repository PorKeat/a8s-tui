package ui

import (
	"fmt"
	"strings"

	"github.com/PorKeat/a8s-tui/api"
	"github.com/PorKeat/a8s-tui/ui/features/deploy"

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
	title := "Single database"
	lead := "Create a single-instance database deployment."
	if m.dbForm.modeOrDefault() == "cluster" {
		title = "Database cluster"
		lead = "Create a highly available database cluster."
	}
	header := dashboardHeader(title, lead, width)
	lines = append(lines, header...)
	cardLines := databaseDeployFormCard(m, width)
	viewportHeight := max(height-len(lines), 1)
	cardLines = scrollDatabaseFormLines(cardLines, m.dbForm.focus, viewportHeight)
	lines = append(lines, cardLines...)
	return fillStyled(lines, bgDark, width, height)
}

func (m model) renderDashboardMonolithicDeployForm(width, height int) []string {
	lines := make([]string, 0, height)
	title := "Monolithic"
	lead := "Deploy the Git project from this terminal directory."
	if m.monolithForm.isMicroservices() {
		title = "Microservices"
		lead = "Deploy one service into a microservice workspace."
	}
	header := dashboardHeader(title, lead, width)
	lines = append(lines, header...)
	cardLines := monolithicDeployFormCard(m, width)
	viewportHeight := max(height-len(lines), 1)
	cardLines = scrollMonolithicFormLines(cardLines, m.monolithForm.focus, viewportHeight)
	lines = append(lines, cardLines...)
	return fillStyled(lines, bgDark, width, height)
}

func scrollDatabaseFormLines(lines []string, focus int, viewportHeight int) []string {
	if len(lines) <= viewportHeight {
		return lines
	}
	focusLine := databaseFormFocusLine(focus)
	maxOffset := max(len(lines)-viewportHeight, 0)
	offset := clamp(focusLine-viewportHeight/2, 0, maxOffset)
	return lines[offset:min(offset+viewportHeight, len(lines))]
}

func scrollMonolithicFormLines(lines []string, focus int, viewportHeight int) []string {
	if len(lines) <= viewportHeight {
		return lines
	}
	focusLine := monolithicFormFocusLine(focus)
	maxOffset := max(len(lines)-viewportHeight, 0)
	offset := clamp(focusLine-viewportHeight/2, 0, maxOffset)
	return lines[offset:min(offset+viewportHeight, len(lines))]
}

func databaseFormFocusLine(focus int) int {
	if focus < 0 {
		return 0
	}
	if focus <= 6 {
		return 5 + focus*3
	}
	return 27
}

func monolithicFormFocusLine(focus int) int {
	if focus < 0 {
		return 0
	}
	if focus <= 6 {
		return 7 + focus*3
	}
	return 29
}

func deploymentFeatureCard(width int, activeIndex int) []string {
	cardWidth := max(width-6, 30)
	card := styleCard.Width(cardWidth)
	title := mainTitleStyle(colorBgCard)
	mutedStyle := mainMutedStyle(colorBgCard)
	readyCount := 0
	for _, feature := range deploy.Features {
		if feature.Ready {
			readyCount++
		}
	}
	summary := fmt.Sprintf("%d ready / %d soon", readyCount, max(len(deploy.Features)-readyCount, 0))
	lines := []string{
		cardContentLine(card, "", width),
		cardContentLine(card, "  "+title.Render("Choose deployment type")+mutedStyle.Render("  "+summary), width),
		cardContentLine(card, "  "+mutedStyle.Render("Move with arrows or j/k. Enter opens selection."), width),
		cardContentLine(card, "", width),
	}
	for index, feature := range deploy.Features {
		lines = append(lines, deploymentFeatureBox(card, cardWidth, width, feature, index == activeIndex)...)
	}
	lines = append(lines, cardContentLine(card, "", width))
	return lines
}

func deploymentFeatureBox(card lipgloss.Style, cardWidth, width int, feature deploy.Feature, active bool) []string {
	rowBg := lipgloss.Color(colorBgCard)
	border := lipgloss.Color(colorBorder)
	labelStyle := mainBodyStyle(colorBgCard)
	mutedStyle := mainMutedStyle(colorBgCard)
	markerStyle := mainMutedStyle(colorBgCard)
	statusStyle := mainMutedStyle(colorBgCard)
	marker := " "
	if active {
		border = lipgloss.Color(colorPrimary)
		labelStyle = mainPrimaryStyle(colorBgCard)
		markerStyle = mainPrimaryStyle(colorBgCard)
		marker = "▌"
	}
	if feature.Ready {
		statusStyle = mainPrimaryStyle(colorBgCard)
	}
	description := feature.Description
	if !feature.Ready {
		description += " Coming soon."
	}
	boxWidth := max(cardWidth-6, 32)
	contentWidth := max(boxWidth-4, 24)
	statusWidth := 7
	labelWidth := 22
	if contentWidth < 62 {
		labelWidth = 18
	}
	descWidth := max(contentWidth-labelWidth-statusWidth-6, 8)
	status := deploymentStatusText(feature.Ready)
	content := markerStyle.Render(marker) +
		mainBodyStyle(colorBgCard).Render("  ") +
		labelStyle.Render(pad(truncatePlain(feature.Label, labelWidth), labelWidth)) +
		mainBodyStyle(colorBgCard).Render("  ") +
		mutedStyle.Render(pad(truncatePlain(description, descWidth), descWidth)) +
		mainBodyStyle(colorBgCard).Render("  ") +
		statusStyle.Render(pad(status, statusWidth))
	box := lipgloss.NewStyle().
		Background(rowBg).
		Foreground(lipgloss.Color(colorText)).
		Width(boxWidth).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		BorderBackground(rowBg).
		Padding(0, 1).
		Render(content)
	rendered := strings.Split(box, "\n")
	lines := make([]string, 0, len(rendered))
	for _, line := range rendered {
		lines = append(lines, cardContentLine(card, "  "+line, width))
	}
	return lines
}

func deploymentStatusText(ready bool) string {
	if ready {
		return "ready"
	}
	return "soon"
}

func databaseDeployFormCard(m model, width int) []string {
	cardWidth := max(width-6, 42)
	card := styleCard.Width(cardWidth)
	title := mainTitleStyle(colorBgCard)
	bodyStyle := mainBodyStyle(colorBgCard)
	mutedStyle := mainMutedStyle(colorBgCard)
	lines := []string{
		cardContentLine(card, "", width),
		cardContentLine(card, "  "+title.Render(databaseDeployTitle(m))+bodyStyle.Render("  ")+mutedStyle.Render(m.dbForm.modeOrDefault()), width),
		cardContentLine(card, "  "+mutedStyle.Render("Use arrows or j/k to move. Left/right changes choices."), width),
		cardContentLine(card, "", width),
	}
	fields := []struct {
		index int
		label string
		value string
	}{
		{0, "Project name", m.dbForm.projectName},
		{1, "Engine", m.dbForm.engine().Label},
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
	payload := fmt.Sprintf("Payload: %s / %s / %s / %s", m.dbForm.engine().ID, m.dbForm.modeOrDefault(), m.dbForm.version(), m.dbForm.size())
	lines = append(lines, cardContentLine(card, "  "+mutedStyle.Render(truncatePlain(payload, cardWidth-4)), width))
	if m.message != "" {
		lines = append(lines, cardContentLine(card, "  "+bodyStyle.Render("* "+truncatePlain(m.message, cardWidth-4)), width))
	}
	lines = append(lines, cardContentLine(card, "", width))
	return lines
}

func databaseDeployTitle(m model) string {
	if m.dbForm.modeOrDefault() == "cluster" {
		return "Deploy Database Cluster"
	}
	return "Deploy Database"
}

func monolithicDeployFormCard(m model, width int) []string {
	cardWidth := max(width-6, 42)
	card := styleCard.Width(cardWidth)
	title := mainTitleStyle(colorBgCard)
	bodyStyle := mainBodyStyle(colorBgCard)
	mutedStyle := mainMutedStyle(colorBgCard)
	directory := truncatePlain(m.monolithForm.directory, cardWidth-16)
	if directory == "" {
		directory = "."
	}
	lines := []string{
		cardContentLine(card, "", width),
		cardContentLine(card, "  "+title.Render("Deploy "+m.monolithForm.title())+bodyStyle.Render("  ")+mutedStyle.Render("current directory"), width),
		cardContentLine(card, "  "+mutedStyle.Render(monolithicDeployHelpText(m.monolithForm)), width),
		cardContentLine(card, "  "+mutedStyle.Render("Directory: ")+bodyStyle.Render(directory), width),
		cardContentLine(card, "", width),
	}
	fields := []struct {
		index int
		label string
		value string
	}{
		{0, "Project name", m.monolithForm.projectName},
		{1, "Git remote", m.monolithForm.repoURL},
		{2, "Branch", m.monolithForm.branch},
		{3, "App port", m.monolithForm.appPort},
	}
	if m.monolithForm.isMicroservices() {
		fields = append(fields,
			struct {
				index int
				label string
				value string
			}{4, "Service name", m.monolithForm.serviceName},
			struct {
				index int
				label string
				value string
			}{5, "Framework", m.monolithForm.framework},
			struct {
				index int
				label string
				value string
			}{6, "Service type", m.monolithForm.serviceType},
		)
	}
	for _, field := range fields {
		lines = append(lines, monolithicDeployFieldBox(card, cardWidth, width, m, field.index, field.label, field.value)...)
	}
	lines = append(lines, cardContentLine(card, "", width))
	lines = append(lines, monolithicDeploySubmitBox(card, cardWidth, width, m)...)
	lines = append(lines, cardContentLine(card, "", width))
	repoFullName := firstNonEmpty(m.monolithForm.repoFullName, "unknown repository")
	framework := firstNonEmpty(m.monolithForm.framework, "auto")
	payload := fmt.Sprintf("Detected: %s / %s / %s", repoFullName, firstNonEmpty(m.monolithForm.branch, "main"), framework)
	lines = append(lines, cardContentLine(card, "  "+mutedStyle.Render(truncatePlain(payload, cardWidth-4)), width))
	if m.message != "" {
		lines = append(lines, cardContentLine(card, "  "+bodyStyle.Render("* "+truncatePlain(m.message, cardWidth-4)), width))
	}
	lines = append(lines, cardContentLine(card, "", width))
	return lines
}

func monolithicDeployHelpText(form monolithicDeployForm) string {
	if form.isMicroservices() {
		return "Paste a Git remote. Enter submits a one-service microservice deploy."
	}
	return "Vercel-style deploy from Git. Enter submits on Deploy."
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

func monolithicDeployFieldBox(card lipgloss.Style, cardWidth, width int, m model, index int, label, value string) []string {
	active := m.monolithForm.focus == index
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
	labelText := "Deploy"
	if m.dbForm.modeOrDefault() == "cluster" {
		labelText = "Deploy Cluster"
	}
	label := labelStyle.Render(labelText)
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

func monolithicDeploySubmitBox(card lipgloss.Style, cardWidth, width int, m model) []string {
	active := m.monolithForm.focus == m.monolithForm.fieldCount()-1
	rowBg := lipgloss.Color(colorBgCard)
	border := lipgloss.Color(colorBorder)
	labelStyle := mainBodyStyle(colorBgCard)
	if active {
		border = lipgloss.Color(colorPrimary)
		labelStyle = mainPrimaryStyle(colorBgCard)
	}
	labelText := "Deploy"
	if m.monolithForm.isMicroservices() {
		labelText = "Deploy Microservice"
	}
	label := labelStyle.Render(labelText)
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
	lines = append(lines, spaces(leftMargin)+m.databaseFormLine(1, "Engine", m.dbForm.engine().Label, contentWidth))
	lines = append(lines, spaces(leftMargin)+m.databaseFormLine(2, "Database name", m.dbForm.databaseName, contentWidth))
	lines = append(lines, spaces(leftMargin)+m.databaseFormLine(3, "Username", m.dbForm.username, contentWidth))
	lines = append(lines, spaces(leftMargin)+m.databaseFormLine(4, "Password", maskValue(m.dbForm.password), contentWidth))
	lines = append(lines, spaces(leftMargin)+m.databaseFormLine(5, "Version", m.dbForm.version(), contentWidth))
	lines = append(lines, spaces(leftMargin)+m.databaseFormLine(6, "Size", m.dbForm.size(), contentWidth))
	lines = append(lines, "")
	lines = append(lines, spaces(leftMargin)+m.databaseSubmitLine(contentWidth))
	lines = append(lines, "")
	lines = append(lines, spaces(leftMargin)+fgMuted+"Payload: "+reset+fgText+m.dbForm.engine().ID+" / single-instance / "+m.dbForm.version()+" / "+m.dbForm.size()+reset)
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
	deployKind := "Database"
	if deployment.DeploymentMode == "monolith" {
		deployKind = "Monolithic"
	}
	lines = append(lines, spaces(leftMargin)+bold+fgLogo+"Deploy "+deployKind+reset+fgMuted+"  logs"+reset)
	lines = append(lines, spaces(leftMargin)+fgMuted+"Project "+reset+fgText+truncatePlain(title, contentWidth-18)+reset)
	lines = append(lines, spaces(leftMargin)+fgMuted+"Status  "+reset+deploymentStatusDisplay(m, status, statusColor)+deployStatusSuffix(deployment, contentWidth))
	lines = append(lines, monolithicDeploymentURLLines(deployment, contentWidth, leftMargin)...)
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
	footerLines := 0
	if m.message != "" {
		footerLines++
	}
	if m.certificatePathOpen {
		footerLines += 2
	} else if m.canDownloadClusterCertificate() {
		footerLines++
	}
	logLimit := max(height-2-footerLines, len(lines))
	for index := start; index < len(logLines) && len(lines) < logLimit; index++ {
		level := api.InferLogLevel(logLines[index])
		logColor := deploymentLogColor(level)
		prefix := fgMuted + fmt.Sprintf("%03d", index+1) + reset + logColor + " | " + reset
		text := logColor + truncatePlain(logLines[index], contentWidth-visibleLen(prefix)-3) + reset
		lines = append(lines, spaces(leftMargin)+bgPane+pad(" "+prefix+text, contentWidth)+reset)
	}
	for len(lines) < logLimit {
		lines = append(lines, spaces(leftMargin)+bgPane+pad("", contentWidth)+reset)
	}
	if m.certificatePathOpen {
		lines = append(lines, spaces(leftMargin)+fgMuted+"Save certificate path"+reset)
		lines = append(lines, spaces(leftMargin)+fgGreen+"> "+reset+fgText+truncatePlain(m.certificatePath, contentWidth-3)+reset)
	} else if m.canDownloadClusterCertificate() {
		lines = append(lines, spaces(leftMargin)+fgGreen+"c "+reset+fgMuted+"download SSL certificate"+reset)
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

func monolithicDeploymentURLLines(deployment api.DatabaseDeploymentRecord, width, leftMargin int) []string {
	if deployment.DeploymentMode != "monolith" ||
		!api.DatabaseDeploymentTerminal(deployment.Status) ||
		api.DatabaseDeploymentFailed(deployment.Status) ||
		strings.TrimSpace(deployment.DeployURL) == "" {
		return nil
	}
	label := "URL     "
	valueWidth := max(width-len(label), 8)
	parts := splitPlainWidth(deployment.DeployURL, valueWidth)
	lines := make([]string, 0, len(parts))
	for index, part := range parts {
		currentLabel := label
		if index > 0 {
			currentLabel = strings.Repeat(" ", len(label))
		}
		lines = append(lines, spaces(leftMargin)+fgMuted+currentLabel+reset+fgGreen+part+reset)
	}
	return lines
}

func deploymentStatusDisplay(m model, status string, color string) string {
	if api.DatabaseDeploymentTerminal(status) {
		return color + status + reset
	}
	return m.spinner.View() + " " + color + status + "..." + reset
}

func deploymentLogColor(level string) string {
	switch level {
	case "success":
		return fgGreen
	case "warn":
		return fgWarn
	case "error":
		return fgError
	default:
		return fgText
	}
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
	certificate := ""
	if m.certificatePathOpen {
		certificate = " enter save  esc cancel "
	} else if m.canDownloadClusterCertificate() {
		certificate = " c certificate "
	}
	right := bgBar + fgMuted + " arrows/jk scroll  r refresh " + certificate + " b/esc projects  q quit " + reset
	return left + fill(width-visibleLen(left)-visibleLen(right), bgBar+" "+reset) + right
}
