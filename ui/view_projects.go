package ui

import (
	"fmt"
	"strings"

	"github.com/PorKeat/a8s-tui/api"
	"github.com/PorKeat/a8s-tui/ui/components"
	projectfeature "github.com/PorKeat/a8s-tui/ui/features/projects"

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
		"Move with arrows or j/k. Enter opens the selected project. Esc leaves workspace.",
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
	if m.deleteConfirmOpen {
		return m.renderProjectDeleteConfirmation(width, height)
	}
	lines := make([]string, 0, height)
	lines = append(lines, dashboardHeader("Overview", "Review details, press enter on Delete, or esc/b to go back.", width)...)
	lines = append(lines, projectOverviewCard(project, width, m.projectDetailButton)...)
	lines = append(lines, mainLine("", width))
	lines = append(lines, projectDetailColumns(project, width)...)
	return fillStyled(lines, bgDark, width, height)
}

func (m model) renderProjectDeleteConfirmation(width, height int) []string {
	project := m.deleteProject
	if strings.TrimSpace(project.ID) == "" {
		project, _ = m.selectedProject()
	}
	projectName := firstNonEmpty(project.Name, project.ProjectName, "project")
	cardWidth := min(max(width-8, 42), 86)
	card := styleCard.Width(cardWidth)
	title := mainTitleStyle(colorBgCard)
	body := mainBodyStyle(colorBgCard)
	muted := mainMutedStyle(colorBgCard)
	danger := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgCard)).
		Foreground(lipgloss.Color("#ff8787")).
		Bold(true)
	input := m.deleteConfirmText
	emptyInput := input == ""
	inputStyle := body
	if emptyInput {
		input = "type project name"
		inputStyle = muted
	}
	inputBox := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgCard)).
		Foreground(lipgloss.Color(colorText)).
		Width(max(cardWidth-6, 24)).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colorPrimary)).
		BorderBackground(lipgloss.Color(colorBgCard)).
		Render(inputStyle.Render(truncatePlain(input, max(cardWidth-10, 16))))

	lines := make([]string, 0, height)
	lines = append(lines, dashboardHeader("Delete project", "Type the project name exactly, then press enter.", width)...)
	lines = append(lines, cardContentLine(card, "", width))
	lines = append(lines, cardContentLine(card, "  "+danger.Render(truncatePlain("Delete "+projectName, cardWidth-4)), width))
	lines = append(lines, cardContentLine(card, "", width))
	lines = append(lines, cardContentLine(card, "  "+body.Render(truncatePlain(projectDeleteWarning(project), cardWidth-4)), width))
	lines = append(lines, cardContentLine(card, "  "+muted.Render("Required text: ")+title.Render(truncatePlain(projectName, max(cardWidth-22, 8))), width))
	lines = append(lines, cardContentLine(card, "", width))
	for _, line := range strings.Split(inputBox, "\n") {
		lines = append(lines, cardContentLine(card, "  "+line, width))
	}
	lines = append(lines, cardContentLine(card, "", width))
	lines = append(lines, projectDeleteDialogActions(card, width, m.deleteConfirmButton)...)
	lines = append(lines, cardContentLine(card, "  "+muted.Render("tab/arrows select  enter confirm  esc cancel"), width))
	lines = append(lines, cardContentLine(card, "", width))
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

func projectOverviewCard(project api.LiveProject, width, selectedAction int) []string {
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
	lines = append(lines, cardContentLine(card, "  "+muted.Render(truncatePlain(projectDetailHint(project), cardWidth-4)), width))
	if api.ProjectKindSupportsDelete(project.Kind) {
		lines = append(lines, cardContentLine(card, "", width))
		lines = append(lines, projectDetailActions(card, width, selectedAction)...)
	}
	lines = append(lines, cardContentLine(card, "", width))
	return lines
}

func projectDetailHint(project api.LiveProject) string {
	switch strings.ToLower(strings.TrimSpace(project.Kind)) {
	case "database", "dbcluster":
		return "Connection details appear below when the backend provides them."
	case "monolith", "microservices":
		return "Deployment metadata appears below when the backend provides it."
	default:
		return "Project metadata appears below when the backend provides it."
	}
}

func projectDeleteWarning(project api.LiveProject) string {
	switch strings.ToLower(strings.TrimSpace(project.Kind)) {
	case "database":
		return "This deletes linked database deployments and starts backend cleanup."
	case "dbcluster":
		return "This deletes the database cluster with data cleanup enabled."
	case "monolith", "microservices":
		return "This removes the deployed application project from A8S."
	default:
		return "This removes the deployed project from A8S."
	}
}

func projectDetailActions(card lipgloss.Style, width, selected int) []string {
	deleteButton := projectDialogButton("Delete", selected == 0, true)
	cancelButton := projectDialogButton("Cancel", selected == 1, false)
	return []string{cardContentLine(card, "  "+deleteButton+mainBodyStyle(colorBgCard).Render("  ")+cancelButton, width)}
}

func projectDeleteDialogActions(card lipgloss.Style, width, selected int) []string {
	deleteButton := projectDialogButton("Delete", selected == 0, true)
	cancelButton := projectDialogButton("Cancel", selected == 1, false)
	return []string{cardContentLine(card, "  "+deleteButton+mainBodyStyle(colorBgCard).Render("  ")+cancelButton, width)}
}

func projectDialogButton(label string, selected, danger bool) string {
	bg := colorBgPill
	fg := colorText
	if danger {
		bg = "#3a2424"
		fg = "#ff8787"
	}
	if selected {
		if danger {
			bg = "#5a2b2b"
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
	addRow("Password", project.DatabasePassword)
	addRow("Engine", project.Engine)
	addRow("Version", project.Version)
	addRow("Namespace", project.Namespace)
	addRow("URL", jdbcURL)
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
		values := []string{truncatePlain(row.value, valueWidth)}
		if row.label == "URL" {
			values = splitPlainWidth(row.value, valueWidth)
		}
		for index, value := range values {
			label := row.label
			if index > 0 {
				label = ""
			}
			content := body.Render("  ") +
				muted.Render(pad(truncatePlain(label, labelWidth), labelWidth)) +
				body.Render("  ") +
				body.Render(value)
			lines = append(lines, card.Render(content))
		}
	}
	lines = append(lines, card.Render(""))
	return repairCardLines(lines)
}

func splitPlainWidth(text string, width int) []string {
	text = strings.TrimSpace(text)
	if text == "" || width <= 0 {
		return nil
	}
	runes := []rune(text)
	lines := make([]string, 0, (len(runes)+width-1)/width)
	for len(runes) > 0 {
		count := min(width, len(runes))
		lines = append(lines, string(runes[:count]))
		runes = runes[count:]
	}
	return lines
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
	case "redis":
		scheme := "redis"
		if tlsEnabled {
			scheme = "rediss"
		}
		return fmt.Sprintf("%s://%s:%s/%s", scheme, host, port, database)
	case "mongodb", "mongo":
		return fmt.Sprintf("mongodb://%s:%s/%s?authSource=%s&tls=%t", host, port, database, database, tlsEnabled)
	case "cassandra":
		suffix := ""
		if tlsEnabled {
			suffix = "?sslenabled=true&verifyservercertificate=true"
		}
		return fmt.Sprintf("jdbc:cassandra://%s:%s/%s%s", host, port, database, suffix)
	default:
		return fmt.Sprintf("%s:%s", host, port)
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
