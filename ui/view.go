package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/ITProfessional-Gen01/a8s-cli/ui/components"
	uitheme "github.com/ITProfessional-Gen01/a8s-cli/ui/theme"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func applyTheme(dark bool) {
	palette := uitheme.PaletteFor(dark)
	colorPrimary = palette.Primary
	colorBgMain = palette.BgMain
	colorBgSide = palette.BgSide
	colorBgCard = palette.BgCard
	colorBgPill = palette.BgPill
	colorBgActive = palette.BgActive
	colorText = palette.Text
	colorMuted = palette.Muted
	colorTitle = palette.Title
	colorBorder = palette.Border

	if dark {
		fgBlue = ansiFg("#7aaeff")
		fgPurple = ansiFg("#b59dff")
		fgAccent = ansiFg("#5fd7ff")
		fgWarn = ansiFg("#ffe066")
		fgError = ansiFg("#ff8787")
		fgGreen = ansiFg("#77f27f")
	} else {
		fgBlue = ansiFg("#2f6fbd")
		fgPurple = ansiFg("#765bd8")
		fgAccent = ansiFg("#0b7285")
		fgWarn = ansiFg("#a05a00")
		fgError = ansiFg("#c92a2a")
		fgGreen = ansiFg("#2f9e44")
	}

	fgLogo = ansiFg(colorPrimary)
	fgLogo2 = ansiFg("#b59dff")
	fgText = ansiFg(colorText)
	fgMuted = ansiFg(colorMuted)
	fgOrange = ansiFg(colorPrimary)
	fgWarm = ansiFg(colorText)
	fgWhite = ansiFg(colorTitle)
	fgKey = ansiFg(colorPrimary)
	fgBorder = ansiFg(colorBorder)
	bgDark = ansiBg(colorBgMain)
	bgPane = ansiBg(colorBgMain)
	bgSide = ansiBg(colorBgSide)
	bgCard = ansiBg(colorBgCard)
	bgPill = ansiBg(colorBgPill)
	bgSelect = ansiBg(colorBgActive)
	bgBar = ansiBg(colorBgSide)

	styleSide = lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgSide)).
		Foreground(lipgloss.Color(colorText))
	styleMain = lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgMain)).
		Foreground(lipgloss.Color(colorText))
	styleCard = lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgCard)).
		Foreground(lipgloss.Color(colorText))
	stylePill = lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgPill)).
		Foreground(lipgloss.Color(colorText)).
		Padding(0, 1)
	styleActive = lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgActive)).
		Foreground(lipgloss.Color(colorPrimary))
	stylePrimary = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorPrimary)).
		Bold(true)
	styleTitle = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorTitle)).
		Bold(true)
	styleMuted = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorMuted))
	styleBorder = lipgloss.NewStyle().
		Foreground(lipgloss.Color(colorBorder))
	styleSideMuted = lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgSide)).
		Foreground(lipgloss.Color(colorMuted))
	styleSideText = lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgSide)).
		Foreground(lipgloss.Color(colorText))
	styleSideBorder = lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgSide)).
		Foreground(lipgloss.Color(colorBorder))
}

func (m model) View() tea.View {
	applyTheme(m.darkMode)
	width := max(m.width, 80)
	height := max(m.height, 24)
	if m.state != stateReady {
		return m.launcherView(width, height)
	}
	if m.deployLogOpen {
		return m.databaseDeployLogView(width, height)
	}
	bodyHeight := max(height-1, 18)
	sidebarWidth := 31
	if width < 96 {
		sidebarWidth = 25
	}
	mainWidth := max(width-sidebarWidth, 48)

	var b strings.Builder
	sidebar := m.renderDashboardSidebar(sidebarWidth, bodyHeight)
	main := m.renderDashboardMain(mainWidth, bodyHeight)
	for i := 0; i < bodyHeight; i++ {
		b.WriteString(backgroundLine(sidebar[i], sidebarWidth, bgSide))
		b.WriteString(backgroundLine(main[i], mainWidth, bgDark))
		b.WriteString("\n")
	}
	if m.dbFormOpen {
		b.WriteString(m.databaseFormStatusline(width))
	} else {
		b.WriteString(m.statusline(width))
	}

	view := tea.NewView(b.String())
	view.AltScreen = true
	view.WindowTitle = "A8S TUI"
	return view
}

func (m model) launcherView(width, height int) tea.View {
	lines := make([]string, 0, height)
	topPad := max((height-26)/2, 1)
	for i := 0; i < topPad; i++ {
		lines = append(lines, "")
	}
	for _, line := range lazyLogoLines() {
		lines = append(lines, centerLine(fgLogo+bold+line+reset, width))
	}
	lines = append(lines, "")
	lines = append(lines, "")
	if m.state == stateLoggedOut && m.logoutSucceeded {
		lines = append(lines, centerLine(fgGreen+bold+"Signed out successfully"+reset, width))
		lines = append(lines, centerLine(fgMuted+"Local session cleared. Next login will ask you to sign in again."+reset, width))
		lines = append(lines, "")
	} else {
		lines = append(lines, "")
		lines = append(lines, "")
	}

	menuWidth := min(max(width/2, 50), 62)
	for index, item := range m.launcherItems() {
		for _, line := range launcherBlock(item.icon, item.label, item.key, menuWidth, index == m.launcherCursor) {
			lines = append(lines, centerLine(line, width))
		}
	}

	lines = append(lines, "")
	messageColor := fgAccent
	if m.state == stateConfigError || m.state == stateError {
		messageColor = fgError
	}
	statusPrefix := fgWarn + "* " + reset
	if m.state == stateLoggingIn || m.state == stateLoading {
		statusPrefix = m.spinner.View() + " "
	}
	lines = append(lines, centerLine(statusPrefix+messageColor+m.message+reset, width))

	for len(lines) < height-1 {
		lines = append(lines, "")
	}

	var b strings.Builder
	for _, line := range lines[:height-1] {
		b.WriteString(backgroundLine(line, width, bgDark))
		b.WriteString("\n")
	}
	b.WriteString(m.launcherStatusline(width))

	view := tea.NewView(b.String())
	view.AltScreen = true
	view.WindowTitle = "A8S TUI"
	return view
}

func (m model) launcherItems() []launcherItem {
	if m.state == stateLoggingIn {
		return []launcherItem{
			{icon: "..", label: "Waiting for browser login", key: "open", action: "wait"},
			{icon: "x", label: "Quit", key: "q", action: "quit"},
		}
	}
	if m.state == stateLoading {
		return []launcherItem{
			{icon: "..", label: "Loading live projects", key: "wait", action: "wait"},
			{icon: "x", label: "Quit", key: "q", action: "quit"},
		}
	}
	if m.state == stateConfigError {
		return []launcherItem{
			{icon: "x", label: "Quit", key: "q", action: "quit"},
		}
	}
	return []launcherItem{
		{icon: "->", label: "Login with Keycloak", key: "l", action: "login"},
		{icon: "x", label: "Quit", key: "q", action: "quit"},
	}
}

func (m model) navigationItems() []navigationItem {
	return []navigationItem{
		{group: "Workspace", label: "Projects", key: "p", page: pageProjects},
		{group: "Workspace", label: "Deployment", key: "d", page: pageDeployment},
		{group: "Security", label: "Image Scanner", key: "i", page: pageImageScanner},
		{group: "Observability", label: "Logs", key: "g", page: pageLogs},
		{group: "Observability", label: "Monitoring", key: "m", page: pageMonitoring},
		{group: "Settings", label: "User", key: "u", page: pageUserSettings},
	}
}

func (item navigationItem) matchesPage(page appPage) bool {
	return item.action == "" && item.page == page
}

func (m model) navigationGroups() []string {
	var groups []string
	seen := map[string]bool{}
	for _, item := range m.navigationItems() {
		if !seen[item.group] {
			groups = append(groups, item.group)
			seen[item.group] = true
		}
	}
	return groups
}

func (m model) navigationItemsForGroup(group string) []navigationItem {
	var items []navigationItem
	for _, item := range m.navigationItems() {
		if item.group == group {
			items = append(items, item)
		}
	}
	return items
}

func (m model) navigationIndexByPage(page appPage) int {
	for index, item := range m.navigationItems() {
		if item.matchesPage(page) {
			return index
		}
	}
	return 0
}

func (m model) navigationIndexByAction(action string) int {
	for index, item := range m.navigationItems() {
		if item.action == action {
			return index
		}
	}
	return 0
}

func (m model) navigationIndexByGroup(group string) int {
	for index, item := range m.navigationItems() {
		if item.group == group {
			return index
		}
	}
	return 0
}

func navigationLine(item navigationItem, width int, active, current bool) string {
	prefix := "  "
	labelColor := fgText
	keyColor := fgMuted
	if current {
		prefix = "* "
		labelColor = bold + fgLogo
	}
	if active {
		prefix = "> "
		keyColor = fgKey
		if !current {
			labelColor = bold + fgText
		}
	}
	left := prefix + labelColor + truncatePlain(item.label, max(width-9, 4)) + reset
	right := keyColor + item.key + "  " + reset
	return bgPane + pad(left+spaces(width-visibleLen(left)-visibleLen(right))+right, width) + reset
}

func outlineLine(item navigationItem, width int, active, current bool) string {
	prefix := "  "
	labelColor := fgText
	icon := itemIcon(item)
	if current {
		prefix = "* "
		labelColor = bold + fgLogo
	}
	if active {
		prefix = "> "
		if !current {
			labelColor = bold + fgText
		}
	}
	left := prefix + icon + " " + labelColor + truncatePlain(item.label, max(width-14, 4)) + reset
	right := fgMuted + item.key + " " + reset
	return bgPane + pad(left+spaces(width-visibleLen(left)-visibleLen(right))+right, width) + reset
}

func featureListLine(item navigationItem, width int, active, current bool) string {
	prefix := "   "
	labelColor := fgText
	icon := itemIcon(item)
	if current {
		prefix = "* "
		labelColor = bold + fgLogo
	}
	if active {
		prefix = "> "
		if !current {
			labelColor = bold + fgText
		}
	}
	left := prefix + icon + " " + labelColor + truncatePlain(item.label, max(width-20, 4)) + reset
	right := fgMuted + " [" + item.key + "] " + reset
	return bgPane + pad(left+spaces(width-visibleLen(left)-visibleLen(right))+right, width) + reset
}

func tabBox(group string, width int, active bool) []string {
	color := fgMuted
	borderColor := fgBorder
	if active {
		color = bold + fgLogo
		borderColor = fgLogo
	}
	innerWidth := max(width-4, 8)
	labelWidth := max(innerWidth-4, 4)
	label := truncatePlain(group, labelWidth)
	top := " " + borderColor + "╭" + strings.Repeat("─", innerWidth) + "╮" + reset + " "
	middleText := groupIcon(group) + " " + color + label + reset
	middle := " " + borderColor + "│" + reset + " " + middleText + spaces(max(innerWidth-visibleLen(middleText)-1, 0)) + borderColor + "│" + reset + " "
	bottom := " " + borderColor + "╰" + strings.Repeat("─", innerWidth) + "╯" + reset + " "
	return []string{
		bgPane + pad(top, width) + reset,
		bgPane + pad(middle, width) + reset,
		bgPane + pad(bottom, width) + reset,
	}
}

func compactTabLine(group string, width int, active bool) string {
	color := fgMuted
	borderColor := fgBorder
	if active {
		color = bold + fgLogo
		borderColor = fgLogo
	}
	innerWidth := max(width-4, 8)
	labelWidth := max(innerWidth-4, 4)
	label := truncatePlain(group, labelWidth)
	text := groupIcon(group) + " " + color + label + reset
	line := " " + borderColor + "│" + reset + " " + text + spaces(max(innerWidth-visibleLen(text)-1, 0)) + borderColor + "│" + reset + " "
	return bgPane + pad(line, width) + reset
}

func (m model) pageTitle() string {
	switch m.page {
	case pageDeployment:
		return "Deployment"
	case pageImageScanner:
		return "Image Scanner"
	case pageLogs:
		return "Logs"
	case pageMonitoring:
		return "Monitoring"
	case pageUserSettings:
		return "User Settings"
	default:
		return "Projects"
	}
}

func (m model) pageMessage() string {
	switch m.page {
	case pageDeployment:
		return "Deployment is selected"
	case pageImageScanner:
		return "Image Scanner is selected"
	case pageLogs:
		return "Logs are selected"
	case pageMonitoring:
		return "Monitoring is selected"
	case pageUserSettings:
		return "User settings are selected"
	default:
		if len(m.projects) == 0 {
			return "No live projects returned by the backend"
		}
		return fmt.Sprintf("Loaded %d live projects", len(m.projects))
	}
}

func (m model) pageBodyLines() []string {
	switch m.page {
	case pageDeployment:
		return []string{
			"Database deployment",
			"Create a single-instance database deployment from the terminal.",
			"Press enter to open the deployment form.",
		}
	case pageImageScanner:
		return []string{
			"Security tools will live here.",
			"Image Scanner is ready as its own section in the sidebar.",
		}
	case pageLogs:
		return []string{
			"Operational logs will live here.",
			"Deployment logs still open automatically after a deployment starts.",
		}
	case pageMonitoring:
		return []string{
			"Monitoring dashboards will live here.",
			"Use this section for runtime health, metrics, and alerts.",
		}
	case pageUserSettings:
		return []string{
			"Appearance",
			"Theme: " + m.themeLabel(),
			"Press t, space, or enter to switch light and dark mode.",
		}
	default:
		return nil
	}
}

func (m model) pageLead() string {
	switch m.page {
	case pageDeployment:
		return "Create and review deployment workflows from one focused workspace."
	case pageImageScanner:
		return "Scan images and review security findings from one focused workspace."
	case pageLogs:
		return "Inspect runtime and deployment logs without leaving the terminal."
	case pageMonitoring:
		return "Track service health, metrics, and operational signals."
	case pageUserSettings:
		return "Manage account preferences and session controls."
	default:
		return "Browse live projects returned by the backend."
	}
}

func groupDescription(group string) string {
	switch group {
	case "Workspace":
		return "Project work, deployments, and live application state."
	case "Security":
		return "Image scanning and security checks."
	case "Observability":
		return "Logs, monitoring, and runtime visibility."
	case "Settings":
		return "User preferences and account controls."
	default:
		return "Feature group."
	}
}

func groupIcon(group string) string {
	switch group {
	case "Workspace":
		return fgBlue + nfFolder + reset
	case "Security":
		return fgOrange + nfShield + reset
	case "Observability":
		return fgGreen + nfChart + reset
	case "Settings":
		return fgPurple + nfGear + reset
	default:
		return fgMuted + nfFile + reset
	}
}

func itemIcon(item navigationItem) string {
	if item.action == "deploy" {
		return fgOrange + nfDeploy + reset
	}
	switch item.page {
	case pageProjects:
		return fgBlue + nfFolder + reset
	case pageDeployment:
		return fgOrange + nfDeploy + reset
	case pageImageScanner:
		return fgOrange + nfShield + reset
	case pageLogs:
		return fgGreen + nfFile + reset
	case pageMonitoring:
		return fgGreen + nfChart + reset
	case pageUserSettings:
		return fgPurple + nfGear + reset
	default:
		return fgMuted + nfFile + reset
	}
}

func launcherBlock(icon, label, key string, width int, active bool) []string {
	prefix := "   "
	borderColor := fgBorder
	labelColor := fgText
	keyColor := fgMuted
	if active {
		prefix = fgKey + "> " + reset
		borderColor = bold + fgLogo
		labelColor = bold + fgLogo
		keyColor = fgKey
	}
	barWidth := max(width-visibleLen(prefix), 12)
	innerWidth := max(barWidth-2, 10)
	labelWidth := max(innerWidth-12, 1)
	label = truncatePlain(label, labelWidth)
	keyText := keyColor + key + reset
	iconText := fgAccent + icon + reset
	top := spaces(visibleLen(prefix)) + borderColor + "╭" + strings.Repeat("─", innerWidth) + "╮" + reset
	middleText := " " + iconText + "  " + labelColor + label + reset
	middle := prefix + borderColor + "│" + reset + middleText
	middle += spaces(max(innerWidth-visibleLen(middleText)-visibleLen(keyText)-1, 0)) + keyText + " " + borderColor + "│" + reset
	bottom := spaces(visibleLen(prefix)) + borderColor + "╰" + strings.Repeat("─", innerWidth) + "╯" + reset
	return []string{
		pad(top, width),
		pad(middle, width),
		pad(bottom, width),
	}
}

func lazyLogoLines() []string {
	return []string{
		" █████╗ ██╗   ██╗████████╗ ██████╗ ███╗   ██╗ ██████╗ ███╗   ███╗ ██████╗ ██╗   ██╗███████╗",
		"██╔══██╗██║   ██║╚══██╔══╝██╔═══██╗████╗  ██║██╔═══██╗████╗ ████║██╔═══██╗██║   ██║██╔════╝",
		"███████║██║   ██║   ██║   ██║   ██║██╔██╗ ██║██║   ██║██╔████╔██║██║   ██║██║   ██║███████╗",
		"██╔══██║██║   ██║   ██║   ██║   ██║██║╚██╗██║██║   ██║██║╚██╔╝██║██║   ██║██║   ██║╚════██║",
		"██║  ██║╚██████╔╝   ██║   ╚██████╔╝██║ ╚████║╚██████╔╝██║ ╚═╝ ██║╚██████╔╝╚██████╔╝███████║",
		"╚═╝  ╚═╝ ╚═════╝    ╚═╝    ╚═════╝ ╚═╝  ╚═══╝ ╚═════╝ ╚═╝     ╚═╝ ╚═════╝  ╚═════╝ ╚══════╝",
	}
}

func (m model) launcherStatusline(width int) string {
	left := bgBar + bold + fgLogo + " a8s " + reset
	right := bgBar + fgMuted + " arrows/jk move  enter select  q quit " + reset
	return left + fill(width-visibleLen(left)-visibleLen(right), bgBar+" "+reset) + right
}

func (m model) header(width int) string {
	left := bold + fgLogo + " A8S " + reset + fgMuted + m.pageTitle() + reset
	state := m.stateLabel()
	right := fgMuted + " " + state + " " + reset
	return left + strings.Repeat(" ", max(width-visibleLen(left)-visibleLen(right), 1)) + right
}

func (m model) subheader(width int) string {
	nav := m.currentPageNavigationItem()
	if m.page != pageProjects {
		left := fgMuted + strings.ToLower(nav.group) + " " + reset + fgAccent + strings.ToLower(nav.label) + reset
		right := fgMuted + "shortcut " + reset + fgKey + nav.key + reset
		return left + strings.Repeat(" ", max(width-visibleLen(left)-visibleLen(right), 1)) + right
	}
	filter := m.filter
	if m.filtering {
		filter = "/" + filter
	}
	if filter == "" {
		filter = "none"
	}
	left := fgMuted + "workspace " + reset + fgAccent + "projects" + reset
	right := fgMuted + "filter " + reset + fgKey + filter + reset
	return left + strings.Repeat(" ", max(width-visibleLen(left)-visibleLen(right), 1)) + right
}

func (m model) stateLabel() string {
	switch m.state {
	case stateConfigError:
		return fgError + "config error" + reset
	case stateLoggedOut:
		return fgWarn + "logged out" + reset
	case stateLoggingIn:
		return fgAccent + "logging in" + reset
	case stateLoading:
		return fgAccent + "loading" + reset
	case stateReady:
		return fgGreen + "ready" + reset
	case stateError:
		return fgError + "error" + reset
	default:
		return fgMuted + "unknown" + reset
	}
}

func (m model) renderMainPanel(width, height int) []string {
	selected := m.selectedNavigationItem()
	if selected.action != "" || !selected.matchesPage(m.page) {
		return m.renderGroupFeatureList(width, height, selected.group)
	}
	if m.page == pageProjects {
		return m.renderProjectList(width, height)
	}
	return m.renderFeaturePage(width, height)
}

func (m model) renderDashboardMain(width, height int) []string {
	if m.dbFormOpen {
		return m.renderDashboardDatabaseDeployForm(width, height)
	}
	if m.monolithFormOpen {
		return m.renderDashboardMonolithicDeployForm(width, height)
	}
	if m.projectDetailOpen {
		return m.renderDashboardProjectDetail(width, height)
	}
	selected := m.selectedNavigationItem()
	if selected.page == pageDeployment {
		return m.renderDashboardDeployment(width, height)
	}
	if selected.action != "" || !selected.matchesPage(m.page) {
		return m.renderDashboardGroup(width, height, selected.group)
	}
	if m.page == pageProjects {
		return m.renderDashboardProjects(width, height)
	}
	return m.renderDashboardFeature(width, height)
}

func (m model) renderDashboardGroup(width, height int, group string) []string {
	lines := make([]string, 0, height)
	lines = append(lines, dashboardHeader(group, groupDescription(group), width)...)
	lines = append(lines, mainTextLine("FEATURES", mainMutedStyle(colorBgMain), width))
	for _, item := range m.navigationItemsForGroup(group) {
		active := item.group == group && item.label == m.selectedNavigationItem().label
		current := item.matchesPage(m.page)
		lines = append(lines, dashboardFeatureLine(item, width, active, current))
	}
	lines = append(lines, mainLine("", width))
	lines = append(lines, mainTextLine("Press enter to open the selected feature.", mainMutedStyle(colorBgMain), width))
	return fillStyled(lines, bgDark, width, height)
}

func sideLine(text string, width int) string {
	color := styleSideMuted
	if text != "" && text != strings.ToUpper(text) {
		color = styleSideText
	}
	line := styleSide.Render("  ") + color.Render(truncatePlain(text, width-4))
	return line + styleSide.Render(spaces(max(width-visibleLen(line), 0)))
}

func searchBox(width int, placeholder string) []string {
	boxWidth := max(width-4, 12)
	contentWidth := max(boxWidth-4, 1)
	box := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgSide)).
		Foreground(lipgloss.Color(colorMuted)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(colorBorder)).
		BorderBackground(lipgloss.Color(colorBgSide)).
		Padding(0, 1).
		Width(boxWidth)
	rendered := strings.Split(box.Render(truncatePlain(nfSearch+"  "+placeholder, contentWidth)), "\n")
	lines := make([]string, 0, len(rendered))
	for _, line := range rendered {
		lines = append(lines, sideContentLine("  "+line, width))
	}
	return lines
}

func mainLine(text string, width int) string {
	return styleMain.Width(width).Render(truncatePlain(text, width))
}

func mainTextLine(text string, textStyle lipgloss.Style, width int) string {
	contentWidth := max(width-6, 1)
	return mainContentLine(styleMain.Render("   ")+textStyle.Render(truncatePlain(text, contentWidth)), width)
}

func dashboardHeader(title, lead string, width int) []string {
	contentWidth := max(width-6, 1)
	titleStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgMain)).
		Foreground(lipgloss.Color(colorTitle)).
		Bold(true)
	leadStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgMain)).
		Foreground(lipgloss.Color(colorMuted))
	rule := lipgloss.NewStyle().
		Background(lipgloss.Color(colorPrimary)).
		Foreground(lipgloss.Color(colorPrimary)).
		Render(strings.Repeat(" ", min(18, contentWidth)))
	return []string{
		mainLine("", width),
		mainLine("", width),
		mainContentLine(styleMain.Render("   ")+titleStyle.Render(truncatePlain(title, contentWidth)), width),
		mainContentLine(styleMain.Render("   ")+leadStyle.Render(truncatePlain(lead, contentWidth)), width),
		mainContentLine(styleMain.Render("   ")+rule, width),
		mainLine("", width),
	}
}

func mainTitleStyle(bg string) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(colorTitle)).
		Bold(true)
}

func mainBodyStyle(bg string) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(colorText))
}

func mainMutedStyle(bg string) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(colorMuted))
}

func mainPrimaryStyle(bg string) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(colorPrimary)).
		Bold(true)
}

func metricLine(width int, metrics []string, action string) string {
	line := styleMain.Render("   ")
	countStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgPill)).
		Foreground(lipgloss.Color(colorTitle)).
		Bold(true)
	labelStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgPill)).
		Foreground(lipgloss.Color(colorMuted))
	pillBase := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgPill)).
		Foreground(lipgloss.Color(colorText)).
		Padding(0, 1)
	for _, metric := range metrics {
		fields := strings.Fields(metric)
		if len(fields) == 0 {
			continue
		}
		metricText := countStyle.Render(fields[0]) +
			lipgloss.NewStyle().Background(lipgloss.Color(colorBgPill)).Render(" ") +
			labelStyle.Render(strings.Join(fields[1:], " "))
		pill := pillBase.Render(metricText)
		if visibleLen(line)+visibleLen(pill)+3 >= width-18 {
			break
		}
		line += pill + styleMain.Render("  ")
	}
	button := lipgloss.NewStyle().
		Background(lipgloss.Color(colorPrimary)).
		Foreground(lipgloss.Color(colorTitle)).
		Bold(true).
		Padding(0, 1).
		Render(action)
	if visibleLen(line)+visibleLen(button) < width {
		line += styleMain.Render(spaces(max(width-visibleLen(line)-visibleLen(button)-3, 1))) + button
	}
	return mainContentLine(line, width)
}

func repairCardLines(lines []string) []string {
	for i, line := range lines {
		lines[i] = strings.ReplaceAll(line, reset, reset+ansiBg(colorBgCard))
	}
	return lines
}

func featureCard(nav navigationItem, width int, body []string) []string {
	cardWidth := max(width-6, 30)
	card := styleCard.Width(cardWidth)
	title := mainTitleStyle(colorBgCard)
	bodyStyle := mainBodyStyle(colorBgCard)
	lines := []string{
		cardContentLine(card, "", width),
		cardContentLine(card, "  "+title.Render(nav.label), width),
		cardContentLine(card, "", width),
	}
	for _, text := range body {
		lines = append(lines, cardContentLine(card, "  "+bodyStyle.Render(truncatePlain(text, cardWidth-4)), width))
	}
	lines = append(lines, cardContentLine(card, "", width))
	return lines
}

func sideContentLine(content string, width int) string {
	return styleSide.Render(content) + styleSide.Render(spaces(max(width-visibleLen(content), 0)))
}

func mainContentLine(content string, width int) string {
	return styleMain.Render(content) + styleMain.Render(spaces(max(width-visibleLen(content), 0)))
}

func cardContentLine(card lipgloss.Style, content string, width int) string {
	renderedCard := strings.ReplaceAll(card.Render(content), reset, reset+ansiBg(colorBgCard))
	line := styleMain.Render("   ") + renderedCard
	return mainContentLine(line, width)
}

func dashboardItemIcon(item navigationItem, bg string) string {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(itemIconColor(item))).
		Render(itemIconGlyph(item))
}

func itemIconGlyph(item navigationItem) string {
	if item.action == "deploy" {
		return nfDeploy
	}
	switch item.page {
	case pageProjects:
		return nfFolder
	case pageDeployment:
		return nfDeploy
	case pageImageScanner:
		return nfShield
	case pageLogs:
		return nfFile
	case pageMonitoring:
		return nfChart
	case pageUserSettings:
		return nfGear
	default:
		return nfFile
	}
}

func itemIconColor(item navigationItem) string {
	if item.action == "deploy" {
		return colorPrimary
	}
	switch item.page {
	case pageProjects:
		return "#7aaeff"
	case pageDeployment:
		return colorPrimary
	case pageImageScanner:
		return colorPrimary
	case pageLogs, pageMonitoring:
		return "#77f27f"
	case pageUserSettings:
		return "#b59dff"
	default:
		return colorMuted
	}
}

func fillStyled(lines []string, bg string, width, height int) []string {
	style := lipgloss.NewStyle().Background(lipgloss.Color(colorBgMain)).Width(width)
	if bg == bgSide {
		style = styleSide.Width(width)
	}
	for len(lines) < height {
		lines = append(lines, style.Render(""))
	}
	if len(lines) > height {
		return lines[:height]
	}
	return lines
}

func (m model) renderSectionRail(width, height int) []string {
	lines := make([]string, 0, height)
	groups := m.navigationGroups()
	activeGroup := m.selectedNavigationGroup()
	lines = append(lines, bgPane+pad("   "+fgMuted+"Tabs"+reset, width)+reset)
	lines = append(lines, bgPane+pad("", width)+reset)
	useBoxTabs := height >= 18
	for _, group := range groups {
		active := group == activeGroup
		if useBoxTabs {
			lines = append(lines, tabBox(group, width, active)...)
			lines = append(lines, bgPane+pad("", width)+reset)
			continue
		}
		lines = append(lines, compactTabLine(group, width, active))
		lines = append(lines, bgPane+pad("", width)+reset)
	}
	return fillPane(lines, width, height)
}

func (m model) renderOutline(width, height int) []string {
	if width == 0 {
		return nil
	}
	lines := make([]string, 0, height)
	group := m.selectedNavigationGroup()
	lines = append(lines, paneTitle("outline", width, m.focus == focusSidebar))
	lines = append(lines, bgPane+pad("", width)+reset)
	lines = append(lines, sectionLine(strings.ToLower(group), width))
	for _, item := range m.navigationItemsForGroup(group) {
		active := m.navCursor == m.navigationIndexByAction(item.action)
		if item.action == "" {
			active = m.navCursor == m.navigationIndexByPage(item.page)
		}
		current := item.matchesPage(m.page)
		lines = append(lines, outlineLine(item, width, active, current))
	}
	if m.page == pageProjects {
		lines = append(lines, bgPane+pad("", width)+reset)
		lines = append(lines, sectionLine("selected", width))
		lines = append(lines, m.renderSelectedProjectOutline(width)...)
	}
	lines = append(lines, bgPane+pad("", width)+reset)
	lines = append(lines, sectionLine("keys", width))
	keys := []string{"enter open", "tab focus", "r refresh", "o logout"}
	for _, key := range keys {
		lines = append(lines, bgPane+pad("  "+fgMuted+key+reset, width)+reset)
	}
	return fillPane(lines, width, height)
}

func (m model) emptyStateText() string {
	switch m.state {
	case stateConfigError:
		return "Fix TUI env config, then restart"
	case stateLoggedOut:
		return "Press enter to login with Keycloak"
	case stateLoggingIn:
		return "Complete login in your browser"
	case stateLoading:
		return "Loading live projects..."
	default:
		return m.message
	}
}

func (m model) statusline(width int) string {
	left := bgBar + bold + fgLogo + " a8s " + reset
	count := fmt.Sprintf(" %d/%d projects ", len(m.visibleProjects()), len(m.projects))
	if !m.lastRefreshed.IsZero() {
		count += "refreshed " + m.lastRefreshed.Format("15:04:05") + " "
	}
	right := bgBar + fgMuted + " arrows/jk move  tab focus  / filter  r refresh  o logout  q quit " + reset
	return left + bgBar + fgMuted + count + reset + fill(width-visibleLen(left)-len(count)-visibleLen(right), bgBar+" "+reset) + right
}

func (m model) databaseFormLine(index int, label, value string, width int) string {
	active := m.dbForm.focus == index
	labelColor := fgMuted
	valueColor := fgText
	prefix := "  "
	if active {
		labelColor = bold + fgLogo
		valueColor = bold + fgText
		prefix = "> "
	}
	if value == "" {
		value = "..."
		valueColor = fgMuted
	}
	field := prefix + labelColor + fmt.Sprintf("%-15s", label) + reset + " " + valueColor + value + reset
	return centerLine(field, width)
}

func (m model) databaseSubmitLine(width int) string {
	active := m.dbForm.focus == 7
	text := "Deploy"
	color := fgKey
	prefix := "  "
	if active {
		color = bold + fgLogo
		prefix = "> "
	}
	return centerLine(prefix+color+text+reset, width)
}

func (m model) databaseFormStatusline(width int) string {
	left := bgBar + bold + fgLogo + " database " + reset
	right := bgBar + fgMuted + " arrows/jk move  left/right choose  enter next/deploy  esc cancel " + reset
	return left + fill(width-visibleLen(left)-visibleLen(right), bgBar+" "+reset) + right
}

func maskValue(value string) string {
	return components.MaskValue(value)
}

func paneTitle(title string, width int, active bool) string {
	color := fgMuted
	bg := bgPane
	if active {
		color = bold + fgLogo
		bg = bgSelect
	}
	text := "   " + color + title + reset
	return bg + pad(text, width) + reset
}

func sectionLine(title string, width int) string {
	return bgPane + pad("   "+fgMuted+title+reset, width) + reset
}

func centeredPaneMessage(message string, width, height int) []string {
	lines := make([]string, 0, height)
	for i := 0; i < max(height/2-1, 0); i++ {
		lines = append(lines, bgPane+pad("", width)+reset)
	}
	lines = append(lines, bgPane+centerLine(fgMuted+message+reset, width)+reset)
	return lines
}

func wrapPaneText(text string, width int, color string) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	limit := max(width-4, 8)
	var lines []string
	for len(text) > limit {
		lines = append(lines, bgPane+pad("  "+color+text[:limit]+reset, width)+reset)
		text = text[limit:]
	}
	lines = append(lines, bgPane+pad("  "+color+text+reset, width)+reset)
	return lines
}

func fillPane(lines []string, width, height int) []string {
	for len(lines) < height {
		lines = append(lines, bgPane+pad("", width)+reset)
	}
	if len(lines) > height {
		return lines[:height]
	}
	return lines
}

func centerLine(text string, width int) string {
	padding := max((width-visibleLen(text))/2, 0)
	return spaces(padding) + text + spaces(max(width-padding-visibleLen(text), 0))
}

func pad(text string, width int) string {
	return components.Pad(text, width)
}

func backgroundLine(text string, width int, bg string) string {
	return bg + strings.ReplaceAll(pad(text, width), reset, reset+bg) + reset
}

func ansiFg(hex string) string {
	r, g, b := hexRGB(hex)
	return fmt.Sprintf("\x1b[38;2;%d;%d;%dm", r, g, b)
}

func ansiBg(hex string) string {
	r, g, b := hexRGB(hex)
	return fmt.Sprintf("\x1b[48;2;%d;%d;%dm", r, g, b)
}

func hexRGB(hex string) (int, int, int) {
	var r, g, b int
	_, _ = fmt.Sscanf(strings.TrimPrefix(hex, "#"), "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}

func fill(width int, text string) string {
	if width <= 0 {
		return ""
	}
	return strings.Repeat(text, width)
}

func spaces(width int) string {
	return components.Spaces(width)
}

func visibleLen(text string) int {
	return components.VisibleLen(text)
}

func joinNonEmpty(values ...string) string {
	var out []string
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, strings.TrimSpace(value))
		}
	}
	return strings.Join(out, " ")
}

func shortTime(value string) string {
	if value == "" {
		return ""
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.Format("2006-01-02 15:04")
	}
	return value
}
