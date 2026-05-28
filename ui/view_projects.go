package ui

import (
	"fmt"
	"strings"

	"github.com/ITProfessional-Gen01/a8s-cli/api"
	"github.com/ITProfessional-Gen01/a8s-cli/ui/components"
	projectfeature "github.com/ITProfessional-Gen01/a8s-cli/ui/features/projects"

	"charm.land/lipgloss/v2"
)

func projectIcon(project api.LiveProject) string {
	switch strings.ToLower(project.Kind) {
	case "database", "dbcluster":
		return fgOrange + nfDatabase + reset
	case "microservices":
		return fgGreen + nfMicroservice + reset
	case "monolith":
		return fgBlue + nfProject + reset
	default:
		return fgMuted + nfFile + reset
	}
}

func (m model) renderDashboardProjects(width, height int) []string {
	lines := make([]string, 0, height)
	counts := projectfeature.KindCounts(m.projects)
	lines = append(lines, dashboardHeader(
		"Project workspace",
		"Real database deployments, monolith apps, and microservice workspaces in your workspace.",
		width,
	)...)
	lines = append(lines, metricLine(width, []string{
		fmt.Sprintf("%d projects", len(m.projects)),
		fmt.Sprintf("%d single database", counts["database"]),
		fmt.Sprintf("%d cluster database", counts["dbcluster"]),
		fmt.Sprintf("%d monolith", counts["monolith"]),
		fmt.Sprintf("%d microservice", counts["microservices"]),
	}, "Create Project"))
	lines = append(lines, mainLine("", width))
	visible := m.visibleProjects()
	if len(visible) == 0 {
		lines = append(lines, emptyProjectCard(width)...)
		return fillStyled(lines, bgDark, width, height)
	}
	lines = append(lines, mainTextLine("PROJECTS", mainMutedStyle(colorBgMain), width))
	rowCount := max(height-len(lines)-1, 1)
	start := 0
	if m.cursor >= rowCount {
		start = m.cursor - rowCount + 1
	}
	for index := start; index < len(visible) && len(lines) < height; index++ {
		lines = append(lines, dashboardProjectRow(visible[index], width, index == m.cursor))
	}
	return fillStyled(lines, bgDark, width, height)
}

func (m model) renderDashboardProjectDetail(width, height int) []string {
	project, ok := m.selectedProject()
	if !ok {
		return m.renderDashboardProjects(width, height)
	}
	lines := make([]string, 0, height)
	lines = append(lines, dashboardHeader("Overview", "Press esc or b to go back to Projects.", width)...)
	lines = append(lines, projectOverviewCard(project, width)...)
	lines = append(lines, mainLine("", width))
	lines = append(lines, projectDetailColumns(project, width)...)
	return fillStyled(lines, bgDark, width, height)
}

func emptyProjectCard(width int) []string {
	cardWidth := max(width-6, 30)
	card := styleCard.Width(cardWidth)
	title := mainTitleStyle(colorBgCard)
	body := mainBodyStyle(colorBgCard)
	icon := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgCard)).
		Foreground(lipgloss.Color(colorPrimary)).
		Render(nfDeploy)
	description := truncatePlain("Your live database deployments, monolith apps, and microservice workspaces appear here as soon as they are created.", cardWidth-4)
	lines := []string{
		cardContentLine(card, "", width),
		cardContentLine(card, "  "+icon, width),
		cardContentLine(card, "", width),
		cardContentLine(card, "  "+title.Render("No projects yet"), width),
		cardContentLine(card, "  "+body.Render(description), width),
		cardContentLine(card, "", width),
	}
	return lines
}

func projectOverviewCard(project api.LiveProject, width int) []string {
	cardWidth := max(width-6, 42)
	card := styleCard.Width(cardWidth)
	title := mainTitleStyle(colorBgCard)
	body := mainBodyStyle(colorBgCard)
	muted := mainMutedStyle(colorBgCard)
	status := projectfeature.StatusLabel(project)
	subtitle := joinNonEmpty(project.Engine, project.DeploymentMode, "version "+project.Version)
	if subtitle == "" {
		subtitle = firstNonEmpty(project.Kind, "project")
	}
	projectName := truncatePlain(project.Name, max(cardWidth-18, 8))
	header := title.Render("  "+projectName+" ") +
		mainPrimaryStyle(colorBgCard).Render(status)
	lines := []string{
		cardContentLine(card, "", width),
		cardContentLine(card, header, width),
		cardContentLine(card, body.Render("  ")+muted.Render(truncatePlain(subtitle, cardWidth-4)), width),
		cardContentLine(card, "", width),
	}
	summary := []struct {
		label string
		value string
	}{
		{"Engine", firstNonEmpty(project.Engine, project.Kind, "n/a")},
		{"Mode", firstNonEmpty(project.DeploymentMode, project.ArchitectureType, "n/a")},
		{"Namespace", firstNonEmpty(project.Namespace, "n/a")},
		{"Updated", firstNonEmpty(shortTime(project.UpdatedAt), "n/a")},
	}
	for _, item := range summary {
		lines = append(lines, projectInfoPill(card, cardWidth, width, item.label, item.value)...)
	}
	lines = append(lines, cardContentLine(card, "", width))
	lines = append(lines, cardContentLine(card, body.Render("  ")+body.Render("Use the connection profile to connect from your database client."), width))
	lines = append(lines, cardContentLine(card, "", width))
	return lines
}

func projectInfoPill(card lipgloss.Style, cardWidth, width int, label, value string) []string {
	boxWidth := max(cardWidth-4, 24)
	labelStyle := mainMutedStyle(colorBgCard)
	valueStyle := mainTitleStyle(colorBgCard)
	content := labelStyle.Render(truncatePlain(label, 18)) +
		mainBodyStyle(colorBgCard).Render("  ") +
		valueStyle.Render(truncatePlain(value, max(boxWidth-24, 8)))
	box := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgCard)).
		Foreground(lipgloss.Color(colorText)).
		Width(boxWidth).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colorBorder)).
		BorderBackground(lipgloss.Color(colorBgCard)).
		Render(content)
	var lines []string
	for _, line := range strings.Split(box, "\n") {
		lines = append(lines, cardContentLine(card, "  "+line, width))
	}
	return lines
}

func projectDetailInfoBox(boxWidth int, label, value string) []string {
	boxWidth = max(boxWidth, 24)
	labelStyle := mainMutedStyle(colorBgCard)
	valueStyle := mainBodyStyle(colorBgCard)
	labelWidth := min(14, max(boxWidth/3, 8))
	valueWidth := max(boxWidth-labelWidth-8, 8)
	content := labelStyle.Render(pad(truncatePlain(label, labelWidth), labelWidth)) +
		mainBodyStyle(colorBgCard).Render("  ") +
		valueStyle.Render(truncatePlain(value, valueWidth))
	box := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgCard)).
		Foreground(lipgloss.Color(colorText)).
		Width(boxWidth).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colorBorder)).
		BorderBackground(lipgloss.Color(colorBgCard)).
		Render(content)
	return repairCardLines(strings.Split(box, "\n"))
}

func projectDetailColumns(project api.LiveProject, width int) []string {
	leftWidth := max((width-8)/2, 36)
	rightWidth := max(width-leftWidth-6, 30)
	left := projectConnectionCard(project, leftWidth)
	right := projectBackupCard(rightWidth)
	rowCount := max(len(left), len(right))
	lines := make([]string, 0, rowCount)
	for i := 0; i < rowCount; i++ {
		leftLine := styleMain.Render(spaces(leftWidth))
		rightLine := styleMain.Render(spaces(rightWidth))
		if i < len(left) {
			leftLine = left[i]
		}
		if i < len(right) {
			rightLine = right[i]
		}
		content := styleMain.Render("   ") + leftLine + styleMain.Render("   ") + rightLine
		lines = append(lines, mainContentLine(content, width))
	}
	return lines
}

func projectConnectionCard(project api.LiveProject, width int) []string {
	card := styleCard.Width(width)
	title := mainTitleStyle(colorBgCard)
	body := mainBodyStyle(colorBgCard)
	muted := mainMutedStyle(colorBgCard)
	hostname := projectConnectionHostname(project)
	port := projectConnectionPort(project)
	jdbcURL := projectJDBCURL(project)
	var rows []struct {
		label string
		value string
	}
	addRow := func(label, value string) {
		if strings.TrimSpace(value) == "" {
			return
		}
		rows = append(rows, struct {
			label string
			value string
		}{label: label, value: strings.TrimSpace(value)})
	}
	addRow("Hostname", hostname)
	addRow("Port", port)
	addRow("Database", firstNonEmpty(project.DatabaseName, project.Name))
	addRow("Username", project.DatabaseUsername)
	addRow("Engine", project.Engine)
	addRow("Version", project.Version)
	addRow("Namespace", project.Namespace)
	addRow("JDBC URL", jdbcURL)
	if len(rows) == 0 {
		addRow("Status", "No connection details yet.")
	}
	lines := []string{
		card.Render(""),
		card.Render(mainBodyStyle(colorBgCard).Render("  ") + title.Render("Connection Profile")),
		card.Render(""),
	}
	labelWidth := 11
	valueWidth := max(width-labelWidth-6, 8)
	for _, row := range rows {
		content := body.Render("  ") +
			muted.Render(pad(truncatePlain(row.label, labelWidth), labelWidth)) +
			body.Render("  ") +
			body.Render(truncatePlain(row.value, valueWidth))
		lines = append(lines, card.Render(content))
	}
	lines = append(lines, card.Render(""))
	return repairCardLines(lines)
}

func projectConnectionHostname(project api.LiveProject) string {
	return firstNonEmpty(project.ServiceHost)
}

func projectConnectionPort(project api.LiveProject) string {
	if project.ServicePort <= 0 {
		return ""
	}
	return fmt.Sprintf("%d", project.ServicePort)
}

func projectJDBCURL(project api.LiveProject) string {
	host := projectConnectionHostname(project)
	port := projectConnectionPort(project)
	database := firstNonEmpty(project.DatabaseName, project.Name)
	if host == "" || port == "" || database == "" {
		return ""
	}
	tlsEnabled := project.RequireSSL || project.ConnectionTLSEnabled
	switch strings.ToLower(strings.TrimSpace(project.Engine)) {
	case "postgresql", "postgres":
		sslMode := "disable"
		if tlsEnabled {
			sslMode = "require"
		}
		return fmt.Sprintf("jdbc:postgresql://%s:%s/%s?sslmode=%s", host, port, database, sslMode)
	case "mysql":
		sslMode := "DISABLED"
		if tlsEnabled {
			sslMode = "REQUIRED"
		}
		return fmt.Sprintf("jdbc:mysql://%s:%s/%s?sslMode=%s&allowPublicKeyRetrieval=true", host, port, database, sslMode)
	case "sqlserver":
		encrypt := "false"
		if tlsEnabled {
			encrypt = "true"
		}
		return fmt.Sprintf("jdbc:sqlserver://%s:%s;databaseName=%s;encrypt=%s;trustServerCertificate=true", host, port, database, encrypt)
	case "oracle":
		service := firstNonEmpty(project.ConnectionServiceName, project.ServiceName, database)
		return fmt.Sprintf("jdbc:oracle:thin:@//%s:%s/%s", host, port, service)
	default:
		return ""
	}
}

func projectBackupCard(width int) []string {
	card := styleCard.Width(width)
	title := mainTitleStyle(colorBgCard)
	body := mainBodyStyle(colorBgCard)
	muted := mainMutedStyle(colorBgCard)
	lines := []string{
		card.Render(""),
		card.Render(mainBodyStyle(colorBgCard).Render("  ") + title.Render("Backup history")),
		card.Render(""),
	}
	lines = append(lines, card.Render(body.Render("  ")+muted.Render("Status")+body.Render("  No backups yet.")))
	lines = append(lines, card.Render(""))
	return repairCardLines(lines)
}

func dashboardProjectRow(project api.LiveProject, width int, active bool) string {
	rowBg := colorBgCard
	nameStyle := mainBodyStyle(rowBg)
	mutedStyle := mainMutedStyle(rowBg)
	prefix := "   "
	if active {
		rowBg = colorBgActive
		nameStyle = mainTitleStyle(rowBg)
		mutedStyle = mainMutedStyle(rowBg)
		prefix = " > "
	}
	nameWidth := max(width-48, 12)
	name := truncatePlain(project.Name, nameWidth)
	kind := truncatePlain(project.Kind, 12)
	status := truncatePlain(project.Status, 14)
	rowStyle := lipgloss.NewStyle().Background(lipgloss.Color(rowBg)).Foreground(lipgloss.Color(colorText))
	content := rowStyle.Render(prefix) +
		dashboardProjectIcon(project, rowBg) +
		rowStyle.Render("  ") +
		nameStyle.Render(pad(name, nameWidth)) +
		rowStyle.Render(" ") +
		mutedStyle.Render(pad(kind, 12)) +
		rowStyle.Render(" ") +
		statusPill(status, rowBg)
	box := lipgloss.NewStyle().
		Background(lipgloss.Color(rowBg)).
		Foreground(lipgloss.Color(colorText)).
		Width(max(width-6, 20)).
		Padding(0, 1).
		Render(content)
	line := styleMain.Render("   ") + strings.ReplaceAll(box, reset, reset+ansiBg(rowBg))
	return mainContentLine(line, width)
}

func statusPill(status, bg string) string {
	color := colorMuted
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "READY", "RUNNING", "DEPLOYED", "HEALTHY", "SUCCESS", "SUCCEEDED":
		color = "#77f27f"
	case "PENDING", "STARTING", "PROVISIONING", "DEPLOYING":
		color = "#ffe066"
	case "FAILED", "ERROR", "UNHEALTHY":
		color = "#ff8787"
	}
	return lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(color)).
		Bold(true).
		Render(truncatePlain(status, 14))
}

func dashboardProjectIcon(project api.LiveProject, bg string) string {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(projectIconColor(project))).
		Render(projectIconGlyph(project))
}

func projectIconGlyph(project api.LiveProject) string {
	switch strings.ToLower(project.Kind) {
	case "database", "dbcluster":
		return nfDatabase
	case "microservices":
		return nfMicroservice
	case "monolith":
		return nfProject
	default:
		return nfFile
	}
}

func projectIconColor(project api.LiveProject) string {
	switch strings.ToLower(project.Kind) {
	case "database", "dbcluster":
		return colorPrimary
	case "microservices":
		return "#77f27f"
	case "monolith":
		return "#7aaeff"
	default:
		return colorMuted
	}
}

func (m model) renderSelectedProjectOutline(width int) []string {
	project, ok := m.selectedProject()
	if !ok {
		return []string{bgPane + pad("  "+fgMuted+"No project selected"+reset, width) + reset}
	}
	rows := []string{
		project.Name,
		joinNonEmpty(project.Kind, project.Status),
		firstNonEmpty(project.Namespace, project.Engine, project.Branch),
	}
	var lines []string
	for _, row := range rows {
		if strings.TrimSpace(row) == "" {
			continue
		}
		lines = append(lines, bgPane+pad("  "+fgText+truncatePlain(row, width-4)+reset, width)+reset)
	}
	return lines
}

func (m model) renderProjectList(width, height int) []string {
	lines := make([]string, 0, height)
	lines = append(lines, paneTitle("main", width, m.focus == focusList))
	visible := m.visibleProjects()
	lines = append(lines, bgPane+pad("", width)+reset)
	lines = append(lines, bgPane+pad("   "+bold+fgLogo+"Workspace"+reset+fgMuted+" / Projects"+reset, width)+reset)
	projectLead := truncatePlain("Browse live projects returned by the backend.", width-4)
	lines = append(lines, bgPane+pad("   "+fgMuted+projectLead+reset, width)+reset)
	lines = append(lines, bgPane+pad("", width)+reset)
	header := "  Name" + spaces(max(width-42, 1)) + "Kind        Status"
	lines = append(lines, bgPane+fgMuted+pad(truncatePlain(header, width), width)+reset)
	lines = append(lines, bgPane+pad("", width)+reset)
	if m.state == stateConfigError || m.state == stateLoggedOut || m.state == stateLoggingIn || m.state == stateLoading {
		lines = append(lines, centeredPaneMessage(m.emptyStateText(), width, height-4)...)
		return fillPane(lines, width, height)
	}
	if len(visible) == 0 {
		lines = append(lines, centeredPaneMessage("No projects match the current filter", width, height-4)...)
		return fillPane(lines, width, height)
	}

	rowCount := max(height-len(lines), 1)
	start := 0
	if m.cursor >= rowCount {
		start = m.cursor - rowCount + 1
	}
	for index := start; index < len(visible) && len(lines) < height; index++ {
		project := visible[index]
		active := index == m.cursor
		lines = append(lines, m.projectRow(project, active, width))
	}
	return fillPane(lines, width, height)
}

func (m model) projectRow(project api.LiveProject, active bool, width int) string {
	nameWidth := max(width-36, 12)
	name := truncatePlain(project.Name, nameWidth)
	kind := truncatePlain(project.Kind, 12)
	status := truncatePlain(project.Status, 14)
	prefix := "  "
	rowBg := bgPane
	nameColor := fgText
	if active {
		prefix = "> "
		rowBg = bgSelect
		nameColor = bold + fgLogo
	}
	line := fmt.Sprintf("%s%s %s%-*s%s  %-12s %s", prefix, projectIcon(project), nameColor, nameWidth, name, reset, kind, status)
	return rowBg + pad(line, width) + reset
}

func (m model) renderDetail(width, height int) []string {
	if width == 0 {
		return nil
	}
	lines := make([]string, 0, height)
	lines = append(lines, paneTitle("detail", width, m.focus == focusDetail))
	lines = append(lines, bgPane+pad("", width)+reset)
	project, ok := m.selectedProject()
	if !ok || m.state != stateReady {
		lines = append(lines, centeredPaneMessage("Select a project", width, height-2)...)
		return fillPane(lines, width, height)
	}

	lines = append(lines, bgPane+pad("  "+bold+fgLogo+truncatePlain(project.Name, width-4)+reset, width)+reset)
	lines = append(lines, bgPane+pad("", width)+reset)
	fields := []struct {
		label string
		value string
	}{
		{"kind", project.Kind},
		{"status", project.Status},
		{"health", project.HealthStatus},
		{"namespace", project.Namespace},
		{"engine", joinNonEmpty(project.Engine, project.Version)},
		{"repo", firstNonEmpty(project.RepoFullName, project.RepoURL)},
		{"branch", project.Branch},
		{"framework", project.Framework},
		{"deploy", project.DeployURL},
		{"cluster", project.TargetClusterName},
		{"release", project.CurrentReleaseID},
		{"updated", shortTime(project.UpdatedAt)},
		{"id", project.ID},
	}
	for _, field := range fields {
		if field.value == "" {
			continue
		}
		label := fgMuted + "  " + fmt.Sprintf("%-10s", field.label) + reset
		value := fgText + truncatePlain(field.value, width-visibleLen(label)-2) + reset
		lines = append(lines, bgPane+pad(label+value, width)+reset)
	}
	return fillPane(lines, width, height)
}

func truncatePlain(text string, width int) string {
	return components.TruncatePlain(strings.TrimSpace(text), width)
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
