package ui

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/PorKeat/a8s-tui/api"
	"github.com/PorKeat/a8s-tui/auth"
	"github.com/PorKeat/a8s-tui/config"
	"github.com/PorKeat/a8s-tui/ui/features/deploy"
	"github.com/PorKeat/a8s-tui/ui/features/settings"
	charmterm "github.com/charmbracelet/x/term"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type appState int

const (
	stateConfigError appState = iota
	stateLoggedOut
	stateLoggingIn
	stateLoading
	stateReady
	stateError
)

type focusArea int

const (
	focusSidebar focusArea = iota
	focusList
	focusDetail
)

type appPage int

const (
	pageProjects appPage = iota
	pageDeployment
	pageImageScanner
	pageLogs
	pageMonitoring
	pageUserSettings
)

type model struct {
	config              config.AppConfig
	configErr           error
	auth                auth.AuthClient
	projectsAPI         api.ProjectClient
	spinner             spinner.Model
	tokens              auth.TokenSet
	state               appState
	width               int
	height              int
	cursor              int
	launcherCursor      int
	navCursor           int
	deployCursor        int
	page                appPage
	focus               focusArea
	filter              string
	filtering           bool
	projects            []api.LiveProject
	projectDetailOpen   bool
	projectDetailButton int
	deleteConfirmOpen   bool
	deleteProject       api.LiveProject
	deleteConfirmText   string
	deleteConfirmButton int
	dbForm              databaseDeployForm
	dbFormOpen          bool
	monolithForm        monolithicDeployForm
	monolithFormOpen    bool
	deployLogOpen       bool
	deployLog           api.DatabaseDeploymentRecord
	deployLogOffset     int
	themeIndex          int
	logoutSucceeded     bool
	message             string
	lastRefreshed       time.Time
	userName            string
}

type loginResultMsg struct {
	tokens auth.TokenSet
	err    error
}

type projectsResultMsg struct {
	tokens   auth.TokenSet
	projects []api.LiveProject
	userName string
	err      error
}

type logoutResultMsg struct {
	err error
}

type databaseDeployResultMsg struct {
	tokens     auth.TokenSet
	deployment api.DatabaseDeploymentRecord
	err        error
}

type clusterDeployResultMsg struct {
	tokens     auth.TokenSet
	deployment api.ClusterDeploymentRecord
	err        error
}

type databaseDeploymentPollMsg struct {
	tokens     auth.TokenSet
	deployment api.DatabaseDeploymentRecord
	err        error
}

type monolithicDeployResultMsg struct {
	tokens     auth.TokenSet
	deployment api.MonolithicDeploymentRecord
	err        error
}

type projectDeleteResultMsg struct {
	tokens  auth.TokenSet
	project api.LiveProject
	err     error
}

type databaseDeployForm struct {
	focus        int
	mode         string
	projectName  string
	engineIndex  int
	databaseName string
	username     string
	password     string
	versionIndex int
	sizeIndex    int
}

type monolithicDeployForm struct {
	focus        int
	projectName  string
	repoURL      string
	repoFullName string
	branch       string
	appPort      string
	framework    string
	directory    string
}

func initialModel(config config.AppConfig, configErr error) model {
	state := stateLoggedOut
	message := "Press enter to authenticate with Keycloak"
	if configErr != nil {
		state = stateConfigError
		message = configErr.Error()
	}
	return model{
		config:      config,
		configErr:   configErr,
		auth:        auth.NewAuthClient(config),
		projectsAPI: api.NewProjectClient(config.BackendBaseURL),
		spinner: spinner.New(
			spinner.WithSpinner(spinner.MiniDot),
			spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#f56618"))),
		),
		dbForm:       newDatabaseDeployForm(),
		monolithForm: newMonolithicDeployForm(),
		state:        state,
		width:        118,
		height:       36,
		page:         pageProjects,
		focus:        focusSidebar,
		themeIndex:   2,
		message:      message,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg { return tea.RequestWindowSize() },
		func() tea.Msg { return m.spinner.Tick() },
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case tea.PasteMsg:
		return m.updatePaste(msg)
	case tea.KeyPressMsg:
		return m.updateKey(msg)
	case loginResultMsg:
		if msg.err != nil {
			m.state = stateError
			m.message = "Login failed: " + msg.err.Error()
			return m, nil
		}
		m.tokens = msg.tokens
		m.logoutSucceeded = false
		m.state = stateLoading
		m.page = pageProjects
		m.navCursor = m.navigationIndexByPage(pageProjects)
		m.focus = focusSidebar
		m.message = "Authenticated. Loading live projects..."
		return m, m.fetchProjectsCmd()
	case projectsResultMsg:
		if msg.err != nil {
			m.state = stateError
			m.message = "Projects failed: " + msg.err.Error()
			return m, nil
		}
		m.tokens = msg.tokens
		m.projects = msg.projects
		if msg.userName != "" {
			m.userName = msg.userName
		}
		m.cursor = clamp(m.cursor, 0, max(len(m.visibleProjects())-1, 0))
		m.state = stateReady
		m.lastRefreshed = time.Now()
		if len(m.projects) == 0 {
			m.message = "No live projects returned by the backend"
		} else {
			m.message = fmt.Sprintf("Loaded %d live projects", len(m.projects))
		}
	case logoutResultMsg:
		if msg.err != nil {
			m.message = "Logged out locally. Remote logout skipped: " + msg.err.Error()
		} else {
			m.message = "Signed out successfully. Press enter to sign in again."
		}
	case databaseDeployResultMsg:
		if msg.tokens.AccessToken != "" {
			m.tokens = msg.tokens
		}
		if msg.err != nil {
			m.state = stateReady
			m.dbFormOpen = true
			m.message = "Database deployment failed: " + msg.err.Error()
			return m, nil
		}
		name := firstNonEmpty(msg.deployment.ProjectName, msg.deployment.ReleaseName, "database")
		m.dbFormOpen = false
		m.dbForm = newDatabaseDeployForm()
		m.deployLogOpen = true
		m.deployLog = msg.deployment
		m.deployLogOffset = 0
		m.state = stateReady
		m.message = "Database deployment accepted: " + name
		if msg.deployment.ID == "" || api.DatabaseDeploymentTerminal(msg.deployment.Status) {
			return m, nil
		}
		return m, m.fetchDatabaseDeploymentCmd(msg.deployment.ID, 2*time.Second)
	case clusterDeployResultMsg:
		if msg.tokens.AccessToken != "" {
			m.tokens = msg.tokens
		}
		if msg.err != nil {
			m.state = stateReady
			m.dbFormOpen = true
			m.message = "Database cluster deployment failed: " + msg.err.Error()
			return m, nil
		}
		name := firstNonEmpty(msg.deployment.Name, msg.deployment.ReleaseName, "database cluster")
		m.dbFormOpen = false
		m.dbForm = newDatabaseDeployForm()
		m.deployLogOpen = true
		m.deployLog = clusterDeploymentLogRecord(msg.deployment)
		m.deployLogOffset = 0
		m.state = stateReady
		m.message = "Database cluster deployment accepted: " + name
		return m, nil
	case databaseDeploymentPollMsg:
		if !m.deployLogOpen {
			return m, nil
		}
		if msg.tokens.AccessToken != "" {
			m.tokens = msg.tokens
		}
		if m.deployLog.ID != "" && msg.deployment.ID != "" && msg.deployment.ID != m.deployLog.ID {
			return m, nil
		}
		if msg.err != nil {
			m.message = "Deployment log refresh failed: " + msg.err.Error()
			return m, nil
		}
		m.deployLog = msg.deployment
		m.deployLogOffset = clamp(m.deployLogOffset, 0, max(len(api.ParseDeploymentLogLines(m.deployLog.StatusLog))-1, 0))
		name := firstNonEmpty(m.deployLog.ProjectName, m.deployLog.ReleaseName, "database")
		if api.DatabaseDeploymentTerminal(m.deployLog.Status) {
			if api.DatabaseDeploymentFailed(m.deployLog.Status) {
				m.message = "Database deployment failed: " + name
			} else {
				m.message = "Database deployment finished: " + name
			}
			return m, nil
		}
		m.message = "Deployment is still running..."
		return m, m.fetchDatabaseDeploymentCmd(m.deployLog.ID, 2*time.Second)
	case monolithicDeployResultMsg:
		if msg.tokens.AccessToken != "" {
			m.tokens = msg.tokens
		}
		if msg.err != nil {
			m.state = stateReady
			m.monolithFormOpen = true
			m.message = "Monolithic deployment failed: " + msg.err.Error()
			return m, nil
		}
		name := firstNonEmpty(msg.deployment.Name, m.monolithForm.projectName, "monolith")
		m.monolithFormOpen = false
		m.monolithForm = newMonolithicDeployForm()
		m.state = stateLoading
		if msg.deployment.QueueItemID > 0 {
			m.message = fmt.Sprintf("Monolithic deployment queued: %s (#%d)", name, msg.deployment.QueueItemID)
		} else {
			m.message = "Monolithic deployment queued: " + name
		}
		return m, m.fetchProjectsCmd()
	case projectDeleteResultMsg:
		if msg.tokens.AccessToken != "" {
			m.tokens = msg.tokens
		}
		name := firstNonEmpty(msg.project.Name, msg.project.ProjectName, "project")
		if msg.err != nil {
			m.state = stateReady
			m.message = "Delete failed: " + msg.err.Error()
			return m, nil
		}
		m.state = stateLoading
		m.projectDetailOpen = false
		m.projectDetailButton = 0
		m.deleteConfirmOpen = false
		m.deleteProject = api.LiveProject{}
		m.deleteConfirmText = ""
		m.deleteConfirmButton = 0
		m.message = "Deleted " + name + ". Refreshing projects..."
		return m, m.fetchProjectsCmd()
	}
	return m, nil
}

func (m model) updateKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.deleteConfirmOpen {
		return m.updateDeleteConfirmation(msg)
	}
	if m.deployLogOpen {
		return m.updateDeploymentLog(msg)
	}
	if m.dbFormOpen {
		return m.updateDatabaseDeployForm(msg)
	}
	if m.monolithFormOpen {
		return m.updateMonolithicDeployForm(msg)
	}
	if m.projectDetailOpen {
		return m.updateProjectDetail(msg)
	}

	key := msg.String()
	code := msg.Key().Code
	isEnter := key == "enter" || code == tea.KeyEnter || code == tea.KeyReturn
	if m.filtering {
		switch {
		case key == "ctrl+c":
			return m, tea.Quit
		case key == "esc" || code == tea.KeyEscape || isEnter:
			m.filtering = false
			m.message = "Filter applied"
		case key == "backspace" || key == "ctrl+h" || code == tea.KeyBackspace:
			if len(m.filter) > 0 {
				m.filter = m.filter[:len(m.filter)-1]
			}
		default:
			if key == "/" {
				return m, nil
			}
			if len(key) == 1 && key >= " " && key <= "~" {
				m.filter += key
			}
		}
		m.cursor = clamp(m.cursor, 0, max(len(m.visibleProjects())-1, 0))
		return m, nil
	}

	switch {
	case key == "ctrl+c" || key == "q":
		return m, tea.Quit
	case key == "esc" || code == tea.KeyEscape:
		if m.state == stateReady && m.focus != focusSidebar {
			m.focus = focusSidebar
			m.filtering = false
			m.message = "Left " + m.pageTitle() + " workspace"
			return m, nil
		}
		return m, tea.Quit
	case m.state == stateReady && m.page == pageUserSettings && (key == "t" || key == " " || key == "space" || (isEnter && m.focus != focusSidebar)):
		m.toggleTheme()
	case isEnter || key == "l":
		if m.state == stateReady && isEnter && m.focus == focusSidebar {
			return m.activateNavigationItem()
		}
		if m.state == stateReady && isEnter && m.page == pageProjects {
			return m.openProjectDetail()
		}
		if m.state == stateReady && isEnter && m.page == pageDeployment {
			return m.activateDeploymentFeature()
		}
		if m.state != stateReady && isEnter {
			return m.activateLauncherItem()
		}
		return m.startLoginIfAvailable()
	case key == "p":
		if m.state == stateReady {
			m.selectNavigationShortcut(pageProjects)
		}
	case key == "d":
		if m.state == stateReady {
			m.selectNavigationShortcut(pageDeployment)
		}
	case key == "i":
		if m.state == stateReady {
			m.selectNavigationShortcut(pageImageScanner)
		}
	case key == "g":
		if m.state == stateReady {
			m.selectNavigationShortcut(pageLogs)
		}
	case key == "m":
		if m.state == stateReady {
			m.selectNavigationShortcut(pageMonitoring)
		}
	case key == "u" || key == "s":
		if m.state == stateReady {
			m.selectNavigationShortcut(pageUserSettings)
		}
	case key == "r":
		if m.tokens.AccessToken != "" {
			m.state = stateLoading
			m.message = "Refreshing live projects..."
			return m, m.fetchProjectsCmd()
		}
	case key == "o":
		if m.tokens.AccessToken != "" || m.tokens.IDToken != "" || m.state == stateReady {
			return m.logout()
		}
	case key == "tab" || code == tea.KeyTab:
		m.focus = focusArea((int(m.focus) + 1) % m.focusAreaCount())
	case key == "/" || code == '/':
		if m.state == stateReady {
			m.page = pageProjects
			m.projectDetailOpen = false
			m.projectDetailButton = 0
			m.deleteConfirmOpen = false
			m.deleteProject = api.LiveProject{}
			m.deleteConfirmText = ""
			m.deleteConfirmButton = 0
			m.dbFormOpen = false
			m.navCursor = m.navigationIndexByPage(pageProjects)
			m.focus = focusList
			m.filtering = true
			m.cursor = clamp(m.cursor, 0, max(len(m.visibleProjects())-1, 0))
			m.message = "Filtering projects"
		}
	case key == "backspace" || key == "ctrl+h" || code == tea.KeyBackspace:
		if m.page == pageProjects && m.filter != "" {
			m.filter = ""
			m.cursor = 0
			m.message = "Filter cleared"
		}
	case key == "up" || key == "k" || code == tea.KeyUp:
		if m.state == stateReady && m.focus == focusSidebar {
			m.moveNavigationCursor(-1)
		} else if m.state == stateReady && m.page == pageDeployment {
			m.moveDeploymentCursor(-1)
		} else if m.state == stateReady && m.page == pageProjects {
			m.moveCursor(-1)
		} else if m.state == stateLoading {
			m.message = "Please wait, loading projects..."
		} else {
			m.moveLauncherCursor(-1)
		}
	case key == "down" || key == "j" || code == tea.KeyDown:
		if m.state == stateReady && m.focus == focusSidebar {
			m.moveNavigationCursor(1)
		} else if m.state == stateReady && m.page == pageDeployment {
			m.moveDeploymentCursor(1)
		} else if m.state == stateReady && m.page == pageProjects {
			m.moveCursor(1)
		} else if m.state == stateLoading {
			m.message = "Please wait, loading projects..."
		} else {
			m.moveLauncherCursor(1)
		}
	case key == "left" || code == tea.KeyLeft:
		if m.state == stateReady {
			if m.focus == focusSidebar {
				item := m.selectedNavigationItem()
				m.message = "Press enter to open " + item.label
				return m, nil
			}
			m.message = "Use esc to leave this workspace"
		}
	case key == "right" || code == tea.KeyRight:
		if m.state == stateReady {
			if m.focus == focusSidebar {
				item := m.selectedNavigationItem()
				m.message = "Press enter to open " + item.label
				return m, nil
			}
			m.message = "Use esc to leave this workspace"
		}
	}
	return m, nil
}

func (m model) updatePaste(msg tea.PasteMsg) (tea.Model, tea.Cmd) {
	text := sanitizePastedFieldText(msg.Content)
	if text == "" {
		return m, nil
	}
	switch {
	case m.deleteConfirmOpen:
		m.deleteConfirmText += text
		m.message = "Type project name and press enter to delete"
	case m.dbFormOpen:
		m.appendDatabaseFormText(text)
		m.message = "Pasted into field"
	case m.monolithFormOpen:
		m.appendMonolithicFormText(text)
		m.message = "Pasted into field"
	case m.filtering:
		m.filter += text
		m.cursor = clamp(m.cursor, 0, max(len(m.visibleProjects())-1, 0))
		m.message = "Filter updated"
	}
	return m, nil
}

func newDatabaseDeployForm() databaseDeployForm {
	return databaseDeployForm{
		mode:         "single-instance",
		engineIndex:  0,
		versionIndex: 0,
		sizeIndex:    0,
	}
}

func newMonolithicDeployForm() monolithicDeployForm {
	local := deploy.DetectLocalProject()
	return monolithicDeployForm{
		projectName: local.Name,
		branch:      local.Branch,
		appPort:     fmt.Sprintf("%d", local.AppPort),
		framework:   local.Framework,
		directory:   local.Directory,
	}
}

func (m model) updateDatabaseDeployForm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	code := msg.Key().Code
	isEnter := key == "enter" || code == tea.KeyEnter || code == tea.KeyReturn
	fieldCount := 8

	switch {
	case key == "ctrl+c":
		return m, tea.Quit
	case key == "esc":
		m.dbFormOpen = false
		m.message = "Database deployment canceled"
	case key == "tab" || code == tea.KeyTab || code == tea.KeyDown:
		m.dbForm.focus = (m.dbForm.focus + 1) % fieldCount
	case code == tea.KeyUp:
		m.dbForm.focus = (m.dbForm.focus + fieldCount - 1) % fieldCount
	case key == "j":
		m.dbForm.focus = (m.dbForm.focus + 1) % fieldCount
	case key == "k":
		m.dbForm.focus = (m.dbForm.focus + fieldCount - 1) % fieldCount
	case code == tea.KeyLeft:
		m.adjustDatabaseChoice(-1)
	case code == tea.KeyRight:
		m.adjustDatabaseChoice(1)
	case key == "backspace" || key == "ctrl+h" || code == tea.KeyBackspace:
		m.deleteDatabaseFormRune()
	case isEnter:
		if m.dbForm.focus == fieldCount-1 {
			return m.submitDatabaseDeployment()
		}
		m.dbForm.focus = (m.dbForm.focus + 1) % fieldCount
	default:
		if len(key) == 1 && key >= " " && key <= "~" {
			m.appendDatabaseFormText(key)
		}
	}
	return m, nil
}

func (m model) updateMonolithicDeployForm(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	code := msg.Key().Code
	isEnter := key == "enter" || code == tea.KeyEnter || code == tea.KeyReturn
	fieldCount := 5

	switch {
	case key == "ctrl+c":
		return m, tea.Quit
	case key == "esc" || code == tea.KeyEscape:
		m.monolithFormOpen = false
		m.message = "Monolithic deployment canceled"
	case key == "tab" || code == tea.KeyTab || code == tea.KeyDown || key == "j":
		m.monolithForm.focus = (m.monolithForm.focus + 1) % fieldCount
	case code == tea.KeyUp || key == "k":
		m.monolithForm.focus = (m.monolithForm.focus + fieldCount - 1) % fieldCount
	case key == "backspace" || key == "ctrl+h" || code == tea.KeyBackspace:
		m.deleteMonolithicFormRune()
	case isEnter:
		if m.monolithForm.focus == fieldCount-1 {
			return m.submitMonolithicDeployment()
		}
		m.monolithForm.focus = (m.monolithForm.focus + 1) % fieldCount
	default:
		if len(key) == 1 && key >= " " && key <= "~" {
			m.appendMonolithicFormText(key)
		}
	}
	return m, nil
}

func (m *model) adjustDatabaseChoice(delta int) {
	switch m.dbForm.focus {
	case 1:
		m.dbForm.engineIndex = wrapIndex(m.dbForm.engineIndex+delta, len(deploy.EngineOptions))
		m.dbForm.versionIndex = 0
	case 5:
		m.dbForm.versionIndex = wrapIndex(m.dbForm.versionIndex+delta, len(m.dbForm.engine().Versions))
	case 6:
		m.dbForm.sizeIndex = wrapIndex(m.dbForm.sizeIndex+delta, len(deploy.SizeOptions))
	}
}

func (m *model) appendDatabaseFormText(text string) {
	switch m.dbForm.focus {
	case 0:
		m.dbForm.projectName += text
	case 2:
		m.dbForm.databaseName += text
	case 3:
		m.dbForm.username += text
	case 4:
		m.dbForm.password += text
	}
}

func (m *model) deleteDatabaseFormRune() {
	switch m.dbForm.focus {
	case 0:
		m.dbForm.projectName = trimLastRune(m.dbForm.projectName)
	case 2:
		m.dbForm.databaseName = trimLastRune(m.dbForm.databaseName)
	case 3:
		m.dbForm.username = trimLastRune(m.dbForm.username)
	case 4:
		m.dbForm.password = trimLastRune(m.dbForm.password)
	}
}

func (m *model) appendMonolithicFormText(text string) {
	switch m.monolithForm.focus {
	case 0:
		m.monolithForm.projectName += text
	case 1:
		m.monolithForm.repoURL += text
		m.monolithForm.repoFullName = deploy.RepoFullNameFromURL(m.monolithForm.repoURL)
	case 2:
		m.monolithForm.branch += text
	case 3:
		m.monolithForm.appPort += text
	}
}

func (m *model) deleteMonolithicFormRune() {
	switch m.monolithForm.focus {
	case 0:
		m.monolithForm.projectName = trimLastRune(m.monolithForm.projectName)
	case 1:
		m.monolithForm.repoURL = trimLastRune(m.monolithForm.repoURL)
		m.monolithForm.repoFullName = deploy.RepoFullNameFromURL(m.monolithForm.repoURL)
	case 2:
		m.monolithForm.branch = trimLastRune(m.monolithForm.branch)
	case 3:
		m.monolithForm.appPort = trimLastRune(m.monolithForm.appPort)
	}
}

func (m model) submitDatabaseDeployment() (tea.Model, tea.Cmd) {
	if m.dbForm.modeOrDefault() == "cluster" {
		return m.submitClusterDeployment()
	}
	input, err := m.dbForm.input()
	if err != nil {
		m.message = err.Error()
		return m, nil
	}
	projectsAPI := m.projectsAPI
	authClient := m.auth
	tokens := m.tokens
	m.message = "Submitting database deployment..."
	return m, func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		var err error
		if tokens.ExpiresSoon(time.Now()) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return databaseDeployResultMsg{tokens: tokens, err: err}
			}
		}
		deployment, err := projectsAPI.CreateDatabaseDeployment(ctx, tokens.AccessToken, input)
		if api.IsUnauthorized(err) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return databaseDeployResultMsg{tokens: tokens, err: err}
			}
			deployment, err = projectsAPI.CreateDatabaseDeployment(ctx, tokens.AccessToken, input)
		}
		return databaseDeployResultMsg{tokens: tokens, deployment: deployment, err: err}
	}
}

func (m model) submitClusterDeployment() (tea.Model, tea.Cmd) {
	input, err := m.dbForm.clusterInput()
	if err != nil {
		m.message = err.Error()
		return m, nil
	}
	projectsAPI := m.projectsAPI
	authClient := m.auth
	tokens := m.tokens
	m.message = "Submitting database cluster deployment..."
	return m, func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		var err error
		if tokens.ExpiresSoon(time.Now()) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return clusterDeployResultMsg{tokens: tokens, err: err}
			}
		}
		deployment, err := projectsAPI.CreateClusterDeployment(ctx, tokens.AccessToken, input)
		if api.IsUnauthorized(err) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return clusterDeployResultMsg{tokens: tokens, err: err}
			}
			deployment, err = projectsAPI.CreateClusterDeployment(ctx, tokens.AccessToken, input)
		}
		return clusterDeployResultMsg{tokens: tokens, deployment: deployment, err: err}
	}
}

func clusterDeploymentLogRecord(deployment api.ClusterDeploymentRecord) api.DatabaseDeploymentRecord {
	status := firstNonEmpty(deployment.Status, "DEPLOYING")
	statusLog := clusterDeploymentStatusLog(deployment)
	return api.DatabaseDeploymentRecord{
		ID:                   firstNonEmpty(deployment.ClusterID, deployment.ReleaseName),
		ReleaseName:          deployment.ReleaseName,
		Namespace:            deployment.Namespace,
		Engine:               deployment.Engine,
		DeploymentMode:       "cluster",
		ProjectName:          firstNonEmpty(deployment.Name, deployment.ReleaseName),
		DatabaseName:         firstNonEmpty(deployment.Name, deployment.ReleaseName),
		Version:              "",
		ServiceHost:          deployment.ServiceHost,
		ServicePort:          deployment.ServicePort,
		ConnectionTLSEnabled: deployment.TLSEnabled,
		Status:               status,
		StatusMessage:        deployment.StatusMessage,
		StatusLog:            statusLog,
	}
}

func clusterDeploymentStatusLog(deployment api.ClusterDeploymentRecord) string {
	var lines []string
	if len(deployment.Command) > 0 {
		lines = append(lines, "$ "+strings.Join(deployment.Command, " "))
	}
	if deployment.Stdout != "" {
		lines = append(lines, deployment.Stdout)
	}
	if deployment.Stderr != "" {
		lines = append(lines, deployment.Stderr)
	}
	if deployment.StatusMessage != "" {
		lines = append(lines, deployment.StatusMessage)
	}
	if len(lines) == 0 {
		lines = append(lines, "Cluster deployment request accepted.")
	}
	return strings.Join(lines, "\n")
}

func (m model) submitMonolithicDeployment() (tea.Model, tea.Cmd) {
	input, err := m.monolithForm.input()
	if err != nil {
		m.message = err.Error()
		return m, nil
	}
	projectsAPI := m.projectsAPI
	authClient := m.auth
	tokens := m.tokens
	m.message = "Submitting monolithic deployment..."
	return m, func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		var err error
		if tokens.ExpiresSoon(time.Now()) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return monolithicDeployResultMsg{tokens: tokens, err: err}
			}
		}
		deployment, err := projectsAPI.CreateMonolithicDeployment(ctx, tokens.AccessToken, input)
		if api.IsUnauthorized(err) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return monolithicDeployResultMsg{tokens: tokens, err: err}
			}
			deployment, err = projectsAPI.CreateMonolithicDeployment(ctx, tokens.AccessToken, input)
		}
		return monolithicDeployResultMsg{tokens: tokens, deployment: deployment, err: err}
	}
}

func (m model) updateDeploymentLog(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	code := msg.Key().Code
	switch {
	case key == "ctrl+c" || key == "q":
		return m, tea.Quit
	case key == "esc" || key == "b" || code == tea.KeyEscape:
		m.deployLogOpen = false
		m.deployLog = api.DatabaseDeploymentRecord{}
		m.deployLogOffset = 0
		m.state = stateLoading
		m.message = "Refreshing live projects..."
		return m, m.fetchProjectsCmd()
	case key == "r":
		if m.deployLog.ID != "" {
			m.message = "Refreshing deployment logs..."
			return m, m.fetchDatabaseDeploymentCmd(m.deployLog.ID, 0)
		}
	case key == "up" || key == "k" || code == tea.KeyUp:
		m.deployLogOffset = max(m.deployLogOffset-1, 0)
	case key == "down" || key == "j" || code == tea.KeyDown:
		m.deployLogOffset = min(m.deployLogOffset+1, max(len(api.ParseDeploymentLogLines(m.deployLog.StatusLog))-1, 0))
	}
	return m, nil
}

func (m model) updateProjectDetail(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	code := msg.Key().Code
	isEnter := key == "enter" || code == tea.KeyEnter || code == tea.KeyReturn
	switch {
	case key == "ctrl+c" || key == "q":
		return m, tea.Quit
	case key == "tab" || code == tea.KeyTab:
		m.projectDetailButton = (m.projectDetailButton + 1) % 2
		m.message = m.projectDetailButtonMessage()
		return m, nil
	case key == "left" || key == "up" || code == tea.KeyLeft || code == tea.KeyUp:
		m.projectDetailButton = 0
		m.message = m.projectDetailButtonMessage()
		return m, nil
	case key == "right" || key == "down" || code == tea.KeyRight || code == tea.KeyDown:
		m.projectDetailButton = 1
		m.message = m.projectDetailButtonMessage()
		return m, nil
	case isEnter:
		if m.projectDetailButton == 1 {
			return m.closeProjectDetail()
		}
		return m.requestProjectDelete()
	case key == "esc" || key == "b" || code == tea.KeyEscape:
		return m.closeProjectDetail()
	}
	return m, nil
}

func (m model) closeProjectDetail() (tea.Model, tea.Cmd) {
	m.projectDetailOpen = false
	m.projectDetailButton = 0
	m.message = "Back to projects"
	return m, nil
}

func (m model) projectDetailButtonMessage() string {
	if m.projectDetailButton == 0 {
		return "Delete selected"
	}
	return "Cancel selected"
}

func (m model) updateDeleteConfirmation(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	code := msg.Key().Code
	isEnter := key == "enter" || code == tea.KeyEnter || code == tea.KeyReturn
	expectedName := firstNonEmpty(m.deleteProject.Name, m.deleteProject.ProjectName)
	switch {
	case key == "ctrl+c" || key == "q":
		return m, tea.Quit
	case key == "tab" || code == tea.KeyTab:
		m.deleteConfirmButton = (m.deleteConfirmButton + 1) % 2
		m.message = m.deleteConfirmButtonMessage()
		return m, nil
	case key == "left" || key == "up" || code == tea.KeyLeft || code == tea.KeyUp:
		m.deleteConfirmButton = 0
		m.message = m.deleteConfirmButtonMessage()
		return m, nil
	case key == "right" || key == "down" || code == tea.KeyRight || code == tea.KeyDown:
		m.deleteConfirmButton = 1
		m.message = m.deleteConfirmButtonMessage()
		return m, nil
	case isEnter:
		if m.deleteConfirmButton == 1 {
			return m.cancelProjectDelete()
		}
		if strings.TrimSpace(m.deleteConfirmText) != expectedName {
			m.message = "Type the project name exactly to confirm delete"
			return m, nil
		}
		project := m.deleteProject
		m.deleteConfirmOpen = false
		m.deleteProject = api.LiveProject{}
		m.deleteConfirmText = ""
		m.deleteConfirmButton = 0
		m.state = stateLoading
		m.message = "Deleting " + firstNonEmpty(project.Name, project.ProjectName, "project") + "..."
		return m, m.deleteProjectCmd(project)
	case key == "esc" || code == tea.KeyEscape:
		return m.cancelProjectDelete()
	case key == "backspace" || key == "ctrl+h" || code == tea.KeyBackspace:
		m.deleteConfirmText = trimLastRune(m.deleteConfirmText)
	default:
		if len(key) == 1 && key >= " " && key <= "~" {
			m.deleteConfirmText += key
		}
	}
	m.message = "Type " + expectedName + " and press enter to delete"
	return m, nil
}

func (m model) cancelProjectDelete() (tea.Model, tea.Cmd) {
	m.deleteConfirmOpen = false
	m.deleteProject = api.LiveProject{}
	m.deleteConfirmText = ""
	m.deleteConfirmButton = 0
	m.message = "Delete canceled"
	return m, nil
}

func (m model) deleteConfirmButtonMessage() string {
	if m.deleteConfirmButton == 0 {
		return "Delete selected"
	}
	return "Esc selected"
}

func (f databaseDeployForm) engine() deploy.EngineOption {
	return deploy.EngineOptions[clamp(f.engineIndex, 0, len(deploy.EngineOptions)-1)]
}

func (f databaseDeployForm) version() string {
	versions := f.engine().Versions
	return versions[clamp(f.versionIndex, 0, len(versions)-1)]
}

func (f databaseDeployForm) size() string {
	return deploy.SizeOptions[clamp(f.sizeIndex, 0, len(deploy.SizeOptions)-1)]
}

func (f databaseDeployForm) modeOrDefault() string {
	if strings.TrimSpace(f.mode) == "" {
		return "single-instance"
	}
	return strings.TrimSpace(f.mode)
}

func (f databaseDeployForm) input() (api.CreateDatabaseDeploymentInput, error) {
	projectName := strings.TrimSpace(f.projectName)
	databaseName := strings.TrimSpace(f.databaseName)
	username := strings.TrimSpace(f.username)
	password := f.password
	if projectName == "" {
		return api.CreateDatabaseDeploymentInput{}, fmt.Errorf("Project name is required")
	}
	if databaseName == "" {
		return api.CreateDatabaseDeploymentInput{}, fmt.Errorf("Database name is required")
	}
	if username == "" {
		return api.CreateDatabaseDeploymentInput{}, fmt.Errorf("Username is required")
	}
	if strings.TrimSpace(password) == "" {
		return api.CreateDatabaseDeploymentInput{}, fmt.Errorf("Password is required")
	}
	return api.CreateDatabaseDeploymentInput{
		ProjectName:    projectName,
		Engine:         f.engine().ID,
		DeploymentMode: f.modeOrDefault(),
		DatabaseName:   databaseName,
		Username:       username,
		Password:       password,
		Version:        f.version(),
		SizeProfile:    f.size(),
	}, nil
}

func (f databaseDeployForm) clusterInput() (api.CreateClusterDeploymentInput, error) {
	input, err := f.input()
	if err != nil {
		return api.CreateClusterDeploymentInput{}, err
	}
	return api.CreateClusterDeploymentInput{
		Namespace:     "default",
		ProjectName:   input.ProjectName,
		Engine:        input.Engine,
		DatabaseName:  input.DatabaseName,
		Username:      input.Username,
		Password:      input.Password,
		Version:       input.Version,
		SizeProfile:   input.SizeProfile,
		TargetCluster: "k8s-cluster2",
	}, nil
}

func (f monolithicDeployForm) input() (api.CreateMonolithicDeploymentInput, error) {
	projectName := strings.TrimSpace(f.projectName)
	repoURL := strings.TrimSpace(f.repoURL)
	branch := strings.TrimSpace(f.branch)
	if projectName == "" {
		return api.CreateMonolithicDeploymentInput{}, fmt.Errorf("Project name is required")
	}
	if repoURL == "" {
		return api.CreateMonolithicDeploymentInput{}, fmt.Errorf("Git remote URL is required")
	}
	if branch == "" {
		branch = "main"
	}
	return api.CreateMonolithicDeploymentInput{
		ProjectName:       projectName,
		RepoURL:           repoURL,
		RepoFullName:      strings.TrimSpace(f.repoFullName),
		Branch:            branch,
		AppPort:           deploy.ParsePositiveInt(f.appPort, 3000),
		ArchitectureType:  "monolithic",
		AutoDeployEnabled: false,
	}, nil
}

func (m model) startLoginIfAvailable() (tea.Model, tea.Cmd) {
	if m.state == stateLoggedOut || m.state == stateError {
		m.state = stateLoggingIn
		m.logoutSucceeded = false
		m.message = "Opening Keycloak login in your browser..."
		return m, m.loginCmd()
	}
	return m, nil
}

func (m model) activateLauncherItem() (tea.Model, tea.Cmd) {
	items := m.launcherItems()
	if len(items) == 0 {
		return m, nil
	}
	item := items[clamp(m.launcherCursor, 0, len(items)-1)]
	switch item.action {
	case "login":
		return m.startLoginIfAvailable()
	case "quit":
		return m, tea.Quit
	default:
		m.message = item.label + " is available after login"
		return m, nil
	}
}

func (m model) activateNavigationItem() (tea.Model, tea.Cmd) {
	items := m.navigationItems()
	if len(items) == 0 {
		return m, nil
	}
	item := items[clamp(m.navCursor, 0, len(items)-1)]
	m.setPage(item.page)
	return m, nil
}

func (m *model) selectNavigationShortcut(page appPage) {
	m.navCursor = m.navigationIndexByPage(page)
	m.focus = focusSidebar
	item := m.selectedNavigationItem()
	m.message = "Press enter to open " + item.label
}

func (m model) activateDeploymentFeature() (tea.Model, tea.Cmd) {
	feature := m.selectedDeploymentFeature()
	if !feature.Ready {
		m.message = feature.Label + " deployment is coming soon"
		return m, nil
	}
	if feature.Label == "Single database" {
		return m.openDatabaseDeployForm(), nil
	}
	if feature.Label == "Database cluster" {
		return m.openDatabaseClusterDeployForm(), nil
	}
	if feature.Label == "Monolithic" {
		return m.openMonolithicDeployForm(), nil
	}
	m.message = feature.Label + " deployment is coming soon"
	return m, nil
}

func (m model) openProjectDetail() (tea.Model, tea.Cmd) {
	if _, ok := m.selectedProject(); !ok {
		m.message = "No project selected"
		return m, nil
	}
	m.projectDetailOpen = true
	m.projectDetailButton = 0
	m.focus = focusList
	return m, nil
}

func (m model) requestProjectDelete() (tea.Model, tea.Cmd) {
	project, ok := m.selectedProject()
	if !ok {
		m.message = "No project selected"
		return m, nil
	}
	if !api.ProjectKindSupportsDelete(project.Kind) {
		m.message = "Delete is not available for " + firstNonEmpty(project.Kind, "this") + " projects"
		return m, nil
	}
	m.deleteConfirmOpen = true
	m.deleteProject = project
	m.deleteConfirmText = ""
	m.deleteConfirmButton = 0
	m.message = fmt.Sprintf("Type %s and press enter to delete.", firstNonEmpty(project.Name, project.ProjectName, "project"))
	return m, nil
}

func (m model) openDatabaseDeployForm() model {
	m.navCursor = m.navigationIndexByPage(pageDeployment)
	m.dbFormOpen = true
	m.monolithFormOpen = false
	m.dbForm = newDatabaseDeployForm()
	m.message = "Create a single-instance database deployment"
	return m
}

func (m model) openDatabaseClusterDeployForm() model {
	m.navCursor = m.navigationIndexByPage(pageDeployment)
	m.dbFormOpen = true
	m.monolithFormOpen = false
	m.dbForm = newDatabaseDeployForm()
	m.dbForm.mode = "cluster"
	m.message = "Create a highly available database cluster"
	return m
}

func (m model) openMonolithicDeployForm() model {
	m.navCursor = m.navigationIndexByPage(pageDeployment)
	m.dbFormOpen = false
	m.monolithFormOpen = true
	m.monolithForm = newMonolithicDeployForm()
	m.message = "Paste or enter a Git remote URL before deploying."
	return m
}

func (m *model) setPage(page appPage) {
	m.page = page
	m.projectDetailOpen = false
	m.projectDetailButton = 0
	m.deleteConfirmOpen = false
	m.deleteProject = api.LiveProject{}
	m.deleteConfirmText = ""
	m.deleteConfirmButton = 0
	m.dbFormOpen = false
	m.monolithFormOpen = false
	m.navCursor = m.navigationIndexByPage(page)
	if page == pageDeployment {
		m.deployCursor = 0
	}
	m.filtering = false
	m.focus = focusList
	m.message = m.pageMessage()
}

func (m *model) toggleTheme() {
	m.themeIndex = settings.NormalizeThemeIndex(m.themeIndex + 1)
	m.message = "Theme changed to " + m.themeLabel()
}

func (m model) themeLabel() string {
	return settings.ThemeLabel(m.themeIndex)
}

func (m model) logout() (tea.Model, tea.Cmd) {
	authClient := m.auth
	tokens := m.tokens
	m.tokens = auth.TokenSet{}
	m.projects = nil
	m.projectDetailOpen = false
	m.projectDetailButton = 0
	m.deleteConfirmOpen = false
	m.deleteProject = api.LiveProject{}
	m.deleteConfirmText = ""
	m.deleteConfirmButton = 0
	m.dbFormOpen = false
	m.monolithFormOpen = false
	m.deployLogOpen = false
	m.deployLog = api.DatabaseDeploymentRecord{}
	m.deployLogOffset = 0
	m.dbForm = newDatabaseDeployForm()
	m.monolithForm = newMonolithicDeployForm()
	m.cursor = 0
	m.deployCursor = 0
	m.filter = ""
	m.filtering = false
	m.navCursor = 0
	m.page = pageProjects
	m.focus = focusSidebar
	m.launcherCursor = 0
	m.lastRefreshed = time.Time{}
	m.state = stateLoggedOut
	m.logoutSucceeded = true
	m.message = "Logging out locally..."
	return m, func() tea.Msg {
		return logoutResultMsg{err: authClient.Logout(tokens)}
	}
}

func (m model) fetchDatabaseDeploymentCmd(deploymentID string, delay time.Duration) tea.Cmd {
	projectsAPI := m.projectsAPI
	authClient := m.auth
	tokens := m.tokens
	return func() tea.Msg {
		if delay > 0 {
			time.Sleep(delay)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var err error
		if tokens.ExpiresSoon(time.Now()) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return databaseDeploymentPollMsg{tokens: tokens, err: err}
			}
		}

		deployment, err := projectsAPI.FetchDatabaseDeployment(ctx, tokens.AccessToken, deploymentID)
		if api.IsUnauthorized(err) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return databaseDeploymentPollMsg{tokens: tokens, err: err}
			}
			deployment, err = projectsAPI.FetchDatabaseDeployment(ctx, tokens.AccessToken, deploymentID)
		}
		return databaseDeploymentPollMsg{tokens: tokens, deployment: deployment, err: err}
	}
}

func (m model) deleteProjectCmd(project api.LiveProject) tea.Cmd {
	projectsAPI := m.projectsAPI
	authClient := m.auth
	tokens := m.tokens
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer cancel()

		var err error
		if tokens.ExpiresSoon(time.Now()) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return projectDeleteResultMsg{tokens: tokens, project: project, err: err}
			}
		}

		err = projectsAPI.DeleteLiveProject(ctx, tokens.AccessToken, project)
		if api.IsUnauthorized(err) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return projectDeleteResultMsg{tokens: tokens, project: project, err: err}
			}
			err = projectsAPI.DeleteLiveProject(ctx, tokens.AccessToken, project)
		}
		return projectDeleteResultMsg{tokens: tokens, project: project, err: err}
	}
}

func (m *model) moveCursor(delta int) {
	visible := m.visibleProjects()
	if len(visible) == 0 {
		m.cursor = 0
		return
	}
	m.cursor = clamp(m.cursor+delta, 0, len(visible)-1)
}

func (m *model) moveLauncherCursor(delta int) {
	itemCount := len(m.launcherItems())
	if itemCount == 0 {
		m.launcherCursor = 0
		return
	}
	m.launcherCursor = clamp(m.launcherCursor+delta, 0, itemCount-1)
}

func (m *model) moveNavigationCursor(delta int) {
	itemCount := len(m.navigationItems())
	if itemCount == 0 {
		m.navCursor = 0
		return
	}
	m.navCursor = clamp(m.navCursor+delta, 0, itemCount-1)
}

func (m *model) moveDeploymentCursor(delta int) {
	m.deployCursor = clamp(m.deployCursor+delta, 0, max(len(deploy.Features)-1, 0))
}

func (m model) selectedDeploymentFeature() deploy.Feature {
	if len(deploy.Features) == 0 {
		return deploy.Feature{}
	}
	return deploy.Features[clamp(m.deployCursor, 0, len(deploy.Features)-1)]
}

func (m model) focusAreaCount() int {
	return 2
}

func (m model) loginCmd() tea.Cmd {
	authClient := m.auth
	return func() tea.Msg {
		tokens, err := authClient.Login(context.Background())
		return loginResultMsg{tokens: tokens, err: err}
	}
}

func (m model) fetchProjectsCmd() tea.Cmd {
	projectsAPI := m.projectsAPI
	authClient := m.auth
	tokens := m.tokens
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var err error
		if tokens.ExpiresSoon(time.Now()) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return projectsResultMsg{tokens: tokens, err: err}
			}
		}

		projects, userName, err := projectsAPI.FetchLiveProjects(ctx, tokens.AccessToken)
		if api.IsUnauthorized(err) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return projectsResultMsg{tokens: tokens, err: err}
			}
			projects, userName, err = projectsAPI.FetchLiveProjects(ctx, tokens.AccessToken)
		}
		return projectsResultMsg{tokens: tokens, projects: projects, userName: userName, err: err}
	}
}

func (m model) visibleProjects() []api.LiveProject {
	return api.FilteredProjects(m.projects, m.filter)
}

func (m model) selectedProject() (api.LiveProject, bool) {
	visible := m.visibleProjects()
	if len(visible) == 0 {
		return api.LiveProject{}, false
	}
	return visible[clamp(m.cursor, 0, len(visible)-1)], true
}

func Run() error {
	cfg, configErr := config.LoadConfig()
	options := []tea.ProgramOption{}
	if width, height, err := charmterm.GetSize(os.Stdout.Fd()); err == nil && width > 0 && height > 0 {
		options = append(options, tea.WithWindowSize(width, height))
	}
	p := tea.NewProgram(initialModel(cfg, configErr), options...)
	_, err := p.Run()
	return err
}

func clamp(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func wrapIndex(index, length int) int {
	if length <= 0 {
		return 0
	}
	for index < 0 {
		index += length
	}
	return index % length
}

func trimLastRune(value string) string {
	runes := []rune(value)
	if len(runes) == 0 {
		return value
	}
	return string(runes[:len(runes)-1])
}

func sanitizePastedFieldText(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	value = strings.ReplaceAll(value, "\n", "")
	value = strings.ReplaceAll(value, "\t", " ")
	return strings.TrimSpace(value)
}
