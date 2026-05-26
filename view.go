package main

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	colorPrimary  = "#f56618"
	colorBgMain   = "#130e0b"
	colorBgSide   = "#1f1712"
	colorBgCard   = "#2b221b"
	colorBgPill   = "#231b16"
	colorBgActive = "#4f2f20"
	colorText     = "#ccbeb2"
	colorMuted    = "#9f9186"
	colorTitle    = "#f5f1eb"
	colorBorder   = "#5b4638"
)

const (
	reset = "\x1b[0m"
	bold  = "\x1b[1m"
	dim   = "\x1b[2m"
)

var (
	fgLogo   = "\x1b[38;2;245;102;24m"
	fgLogo2  = "\x1b[38;5;147m"
	fgText   = "\x1b[38;5;159m"
	fgMuted  = "\x1b[38;5;103m"
	fgBlue   = "\x1b[38;2;122;174;255m"
	fgOrange = "\x1b[38;2;245;102;24m"
	fgPurple = "\x1b[38;2;181;157;255m"
	fgWarm   = "\x1b[38;2;204;190;178m"
	fgWhite  = "\x1b[38;2;245;241;235m"
	fgKey    = "\x1b[38;5;216m"
	fgAccent = "\x1b[38;5;81m"
	fgWarn   = "\x1b[38;5;221m"
	fgError  = "\x1b[38;5;203m"
	fgGreen  = "\x1b[38;5;120m"
	fgBorder = "\x1b[38;2;69;54;44m"
	bgDark   = "\x1b[48;2;19;14;11m"
	bgPane   = "\x1b[48;2;19;14;11m"
	bgSide   = "\x1b[48;2;31;23;18m"
	bgCard   = "\x1b[48;2;43;34;27m"
	bgPill   = "\x1b[48;2;35;27;22m"
	bgSelect = "\x1b[48;2;79;47;32m"
	bgBar    = "\x1b[48;2;31;23;18m"
)

const (
	nfSearch       = "\uf002"
	nfFolder       = "\uf07b"
	nfDeploy       = "\uf0ee"
	nfShield       = "\uf132"
	nfFile         = "\uf15b"
	nfChart        = "\uf080"
	nfGear         = "\uf013"
	nfDatabase     = "\uf1c0"
	nfMicroservice = "\ue749"
	nfProject      = "\ue7ba"
)

var (
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
)

func applyTheme(dark bool) {
	if dark {
		colorPrimary = "#f56618"
		colorBgMain = "#130e0b"
		colorBgSide = "#1f1712"
		colorBgCard = "#2b221b"
		colorBgPill = "#231b16"
		colorBgActive = "#4f2f20"
		colorText = "#ccbeb2"
		colorMuted = "#9f9186"
		colorTitle = "#f5f1eb"
		colorBorder = "#5b4638"
		fgBlue = ansiFg("#7aaeff")
		fgPurple = ansiFg("#b59dff")
		fgAccent = ansiFg("#5fd7ff")
		fgWarn = ansiFg("#ffe066")
		fgError = ansiFg("#ff8787")
		fgGreen = ansiFg("#77f27f")
	} else {
		colorPrimary = "#f56618"
		colorBgMain = "#fbf7f2"
		colorBgSide = "#f0e6db"
		colorBgCard = "#fffaf4"
		colorBgPill = "#eadccd"
		colorBgActive = "#ffe4d4"
		colorText = "#49372d"
		colorMuted = "#7f6d62"
		colorTitle = "#211712"
		colorBorder = "#b79b89"
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
	if m.dbFormOpen {
		return m.databaseDeployView(width, height)
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
	b.WriteString(m.statusline(width))

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
	lines = append(lines, "")

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

type launcherItem struct {
	icon   string
	label  string
	key    string
	action string
}

type navigationItem struct {
	group  string
	label  string
	key    string
	page   appPage
	action string
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
		return fgBlue + nfProject + reset
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

func projectIcon(project liveProject) string {
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

func launcherBlock(icon, label, key string, width int, active bool) []string {
	_ = icon
	_ = key
	prefix := "   "
	borderColor := fgGreen
	labelColor := fgText
	if active {
		prefix = fgKey + "> " + reset
		borderColor = bold + fgGreen
		labelColor = bold + fgLogo
	}
	barWidth := max(width-visibleLen(prefix), 12)
	innerWidth := max(barWidth-2, 10)
	label = truncatePlain(label, max(innerWidth-4, 1))
	top := spaces(visibleLen(prefix)) + borderColor + "╭" + strings.Repeat("─", innerWidth) + "╮" + reset
	middle := prefix + borderColor + "│" + reset + "  " + labelColor + label + reset
	middle += spaces(max(innerWidth-visibleLen(label)-2, 0)) + borderColor + "│" + reset
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

func (m model) renderDashboardSidebar(width, height int) []string {
	lines := make([]string, 0, height)
	lines = append(lines, sideLine("SEARCH", width))
	lines = append(lines, sideLine("", width))
	lines = append(lines, searchBox(width, "Search workspace")...)
	lines = append(lines, sideLine("", width))
	lastGroup := ""
	for index, item := range m.navigationItems() {
		if item.group != lastGroup {
			if lastGroup != "" {
				lines = append(lines, sideLine("", width))
			}
			lines = append(lines, sideLine(strings.ToUpper(item.group), width))
			if item.group == "Workspace" {
				lines = append(lines, sideLine("Cheng Devith Workspace", width))
			}
			lastGroup = item.group
		}
		active := index == m.navCursor
		current := item.matchesPage(m.page)
		lines = append(lines, sidebarNavLine(item, width, active, current)...)
	}
	return fillStyled(lines, bgSide, width, height)
}

func (m model) renderDashboardMain(width, height int) []string {
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
	lines = append(lines, mainLine("", width))
	lines = append(lines, mainLine("", width))
	lines = append(lines, mainTextLine(group, mainTitleStyle(colorBgMain), width))
	lines = append(lines, mainTextLine(groupDescription(group), mainBodyStyle(colorBgMain), width))
	lines = append(lines, mainLine("", width))
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

func (m model) renderDashboardProjects(width, height int) []string {
	lines := make([]string, 0, height)
	counts := m.kindCounts()
	lines = append(lines, mainLine("", width))
	lines = append(lines, mainLine("", width))
	lines = append(lines, mainTextLine("Project workspace", mainTitleStyle(colorBgMain), width))
	lines = append(lines, mainTextLine("Real database deployments, monolith apps, and microservice workspaces in your workspace", mainBodyStyle(colorBgMain), width))
	lines = append(lines, mainTextLine("show up here as soon as they are created.", mainBodyStyle(colorBgMain), width))
	lines = append(lines, mainLine("", width))
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

func (m model) renderDashboardFeature(width, height int) []string {
	if m.page == pageDeployment {
		return m.renderDashboardDeployment(width, height)
	}
	nav := m.currentPageNavigationItem()
	lines := make([]string, 0, height)
	lines = append(lines, mainLine("", width))
	lines = append(lines, mainLine("", width))
	lines = append(lines, mainTextLine(nav.label, mainTitleStyle(colorBgMain), width))
	lines = append(lines, mainTextLine(m.pageLead(), mainBodyStyle(colorBgMain), width))
	lines = append(lines, mainLine("", width))
	lines = append(lines, featureCard(nav, width, m.pageBodyLines())...)
	return fillStyled(lines, bgDark, width, height)
}

func (m model) renderDashboardDeployment(width, height int) []string {
	lines := make([]string, 0, height)
	lines = append(lines, mainLine("", width))
	lines = append(lines, mainLine("", width))
	lines = append(lines, mainTextLine("Deployment", mainTitleStyle(colorBgMain), width))
	lines = append(lines, mainTextLine(m.pageLead(), mainBodyStyle(colorBgMain), width))
	lines = append(lines, mainLine("", width))
	lines = append(lines, deploymentFeatureCard(width, m.deployCursor)...)
	lines = append(lines, mainLine("", width))
	helpText := "Press enter to open the selected deployment type."
	if m.page == pageDeployment {
		helpText = "Press esc to close the selected deployment type."
	}
	lines = append(lines, mainTextLine(helpText, mainMutedStyle(colorBgMain), width))
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

func sidebarNavLine(item navigationItem, width int, active, current bool) []string {
	_ = current
	rowBg := lipgloss.Color(colorBgSide)
	border := lipgloss.Color(colorBorder)
	labelStyle := styleSideText
	iconBg := colorBgSide
	cursor := "  "
	if active {
		border = lipgloss.Color(colorPrimary)
		cursor = "> "
		labelStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(colorBgSide)).
			Foreground(lipgloss.Color(colorPrimary)).
			Bold(true)
	}
	rowPad := lipgloss.NewStyle().
		Background(rowBg).
		Foreground(lipgloss.Color(colorText))
	left := rowPad.Render("  ") +
		dashboardItemIcon(item, iconBg) +
		rowPad.Render("  ") +
		labelStyle.Render(truncatePlain(item.label, max(width-12, 4)))
	box := lipgloss.NewStyle().
		Background(rowBg).
		Foreground(lipgloss.Color(colorText)).
		Width(max(width-6, 8)).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(border).
		BorderBackground(rowBg).
		Render(left)
	rendered := strings.Split(box, "\n")
	lines := make([]string, 0, len(rendered))
	cursorStyle := styleSideText
	if active {
		cursorStyle = lipgloss.NewStyle().
			Background(lipgloss.Color(colorBgSide)).
			Foreground(lipgloss.Color(colorPrimary)).
			Bold(true)
	}
	for index, line := range rendered {
		prefix := styleSide.Render("  ")
		if index == 1 {
			prefix = cursorStyle.Render(cursor)
		}
		lines = append(lines, sideContentLine(prefix+line, width))
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

func dashboardFeatureLine(item navigationItem, width int, active, current bool) string {
	rowBg := colorBgMain
	rowStyle := styleMain
	labelStyle := mainBodyStyle(rowBg)
	prefix := "   "
	if current {
		prefix = " * "
		labelStyle = mainPrimaryStyle(rowBg)
	}
	if active {
		rowBg = colorBgActive
		rowStyle = styleActive
		prefix = " > "
		if !current {
			labelStyle = mainTitleStyle(rowBg)
		} else {
			labelStyle = mainPrimaryStyle(rowBg)
		}
	}
	left := rowStyle.Render(prefix) +
		dashboardItemIcon(item, rowBg) +
		rowStyle.Render("  ") +
		labelStyle.Render(truncatePlain(item.label, max(width-18, 4)))
	right := mainMutedStyle(rowBg).Render(item.key + "   ")
	gap := spaces(max(width-visibleLen(left)-visibleLen(right), 0))
	return left + rowStyle.Render(gap) + rowStyle.Render(right)
}

func metricLine(width int, metrics []string, action string) string {
	line := styleMain.Render("   ")
	countStyle := mainTitleStyle(colorBgMain)
	labelStyle := mainBodyStyle(colorBgMain)
	for _, metric := range metrics {
		fields := strings.Fields(metric)
		if len(fields) == 0 {
			continue
		}
		metricText := countStyle.Render(fields[0]) +
			styleMain.Render(" ") +
			labelStyle.Render(strings.Join(fields[1:], " "))
		if visibleLen(line)+visibleLen(metricText)+3 >= width-18 {
			break
		}
		line += metricText + styleMain.Render("   ")
	}
	button := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgActive)).
		Foreground(lipgloss.Color(colorTitle)).
		Bold(true).
		Padding(0, 1).
		Render(action)
	if visibleLen(line)+visibleLen(button) < width {
		line += styleMain.Render(spaces(max(width-visibleLen(line)-visibleLen(button)-3, 1))) + button
	}
	return mainContentLine(line, width)
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

func dashboardProjectRow(project liveProject, width int, active bool) string {
	rowBg := colorBgMain
	rowStyle := styleMain
	nameStyle := mainBodyStyle(rowBg)
	prefix := "   "
	if active {
		rowBg = colorBgActive
		rowStyle = styleActive
		nameStyle = mainTitleStyle(rowBg)
		prefix = " > "
	}
	nameWidth := max(width-44, 12)
	name := truncatePlain(project.Name, nameWidth)
	kind := truncatePlain(project.Kind, 12)
	status := truncatePlain(project.Status, 14)
	line := rowStyle.Render(prefix) +
		dashboardProjectIcon(project, rowBg) +
		rowStyle.Render("  ") +
		nameStyle.Render(pad(name, nameWidth)) +
		rowStyle.Render("  ") +
		mainMutedStyle(rowBg).Render(pad(kind, 12)) +
		rowStyle.Render("  ") +
		mainBodyStyle(rowBg).Render(status)
	return line + rowStyle.Render(spaces(max(width-visibleLen(line), 0)))
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

func dashboardProjectIcon(project liveProject, bg string) string {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(projectIconColor(project))).
		Render(projectIconGlyph(project))
}

func itemIconGlyph(item navigationItem) string {
	if item.action == "deploy" {
		return nfDeploy
	}
	switch item.page {
	case pageProjects:
		return nfProject
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

func projectIconGlyph(project liveProject) string {
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

func projectIconColor(project liveProject) string {
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

func (m model) renderGroupFeatureList(width, height int, group string) []string {
	lines := make([]string, 0, height)
	lines = append(lines, paneTitle("main", width, m.focus == focusList))
	lines = append(lines, bgPane+pad("", width)+reset)
	lines = append(lines, bgPane+pad("   "+bold+fgLogo+group+reset, width)+reset)
	description := truncatePlain(groupDescription(group), width-4)
	lines = append(lines, bgPane+pad("   "+fgMuted+description+reset, width)+reset)
	lines = append(lines, bgPane+pad("", width)+reset)
	lines = append(lines, sectionLine("features", width))
	for _, item := range m.navigationItemsForGroup(group) {
		active := item.label == m.selectedNavigationItem().label && item.group == group
		current := item.matchesPage(m.page)
		lines = append(lines, featureListLine(item, width, active, current))
	}
	lines = append(lines, bgPane+pad("", width)+reset)
	hint := truncatePlain("Press enter to open the selected feature.", width-4)
	lines = append(lines, bgPane+pad("   "+fgMuted+hint+reset, width)+reset)
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

func (m model) renderFeaturePage(width, height int) []string {
	lines := make([]string, 0, height)
	nav := m.currentPageNavigationItem()
	lines = append(lines, paneTitle("main", width, m.focus == focusList))
	lines = append(lines, bgPane+pad("", width)+reset)
	lines = append(lines, bgPane+pad("   "+bold+fgLogo+nav.group+reset+fgMuted+" / "+nav.label+reset, width)+reset)
	lead := truncatePlain(m.pageLead(), width-4)
	lines = append(lines, bgPane+pad("   "+fgMuted+lead+reset, width)+reset)
	lines = append(lines, bgPane+pad("", width)+reset)
	for _, text := range m.pageBodyLines() {
		lines = append(lines, bgPane+pad("   "+fgText+truncatePlain(text, width-5)+reset, width)+reset)
	}
	lines = append(lines, bgPane+pad("", width)+reset)
	help := truncatePlain("Use the sidebar to switch sections. Press d to create a database deployment.", width-4)
	lines = append(lines, bgPane+pad("   "+fgMuted+help+reset, width)+reset)
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

func (m model) projectRow(project liveProject, active bool, width int) string {
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

func (m model) kindCounts() map[string]int {
	counts := map[string]int{}
	for _, project := range m.projects {
		counts[project.Kind]++
	}
	return counts
}

func (m model) statusline(width int) string {
	left := bgBar + bold + fgLogo + " a8s " + reset
	count := fmt.Sprintf(" %d/%d projects ", len(m.visibleProjects()), len(m.projects))
	if !m.lastRefreshed.IsZero() {
		count += "refreshed " + m.lastRefreshed.Format("15:04:05") + " "
	}
	right := bgBar + fgMuted + " arrows/jk move  d deploy db  / filter  r refresh  o logout  q quit " + reset
	return left + bgBar + fgMuted + count + reset + fill(width-visibleLen(left)-len(count)-visibleLen(right), bgBar+" "+reset) + right
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

func (m model) databaseDeployLogView(width, height int) tea.View {
	contentWidth := min(max(width-10, 76), 132)
	leftMargin := max((width-contentWidth)/2, 0)
	logHeight := max(height-10, 8)
	deployment := m.deployLog
	title := firstNonEmpty(deployment.ProjectName, deployment.ReleaseName, "Database deployment")
	status := firstNonEmpty(deployment.Status, "PENDING")
	statusColor := fgAccent
	if databaseDeploymentTerminal(status) && !databaseDeploymentFailed(status) {
		statusColor = fgGreen
	}
	if databaseDeploymentFailed(status) {
		statusColor = fgError
	}

	lines := make([]string, 0, height)
	lines = append(lines, "")
	lines = append(lines, spaces(leftMargin)+bold+fgLogo+"Deploy Database"+reset+fgMuted+"  logs"+reset)
	lines = append(lines, spaces(leftMargin)+fgMuted+"Project "+reset+fgText+truncatePlain(title, contentWidth-18)+reset)
	lines = append(lines, spaces(leftMargin)+fgMuted+"Status  "+reset+statusColor+status+reset+deployStatusSuffix(deployment, contentWidth))
	lines = append(lines, "")
	lines = append(lines, spaces(leftMargin)+paneTitle("view logs", contentWidth, true))

	logLines := parseDeploymentLogLines(deployment.StatusLog)
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

func deployStatusSuffix(deployment databaseDeploymentRecord, width int) string {
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

func maskValue(value string) string {
	if value == "" {
		return ""
	}
	return strings.Repeat("*", max(len([]rune(value)), 8))
}

func paneTitle(title string, width int, active bool) string {
	color := fgMuted
	if active {
		color = bold + fgLogo
	}
	text := "   " + color + title + reset
	return bgPane + pad(text, width) + reset
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
	remaining := width - visibleLen(text)
	if remaining <= 0 {
		return text
	}
	return text + spaces(remaining)
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
	if width <= 0 {
		return ""
	}
	return strings.Repeat(" ", width)
}

func visibleLen(text string) int {
	count := 0
	inEscape := false
	for _, r := range text {
		switch {
		case r == '\x1b':
			inEscape = true
		case inEscape && r == 'm':
			inEscape = false
		case !inEscape:
			count++
		}
	}
	return count
}

func truncatePlain(text string, width int) string {
	text = strings.TrimSpace(text)
	if width <= 0 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= width {
		return text
	}
	if width <= 1 {
		return string(runes[:width])
	}
	return string(runes[:width-1]) + "."
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
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
