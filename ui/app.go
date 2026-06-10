package ui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
	config               config.AppConfig
	configErr            error
	auth                 auth.AuthClient
	projectsAPI          api.ProjectClient
	scannerAPI           api.ImageScannerClient
	observabilityAPI     api.ObservabilityClient
	spinner              spinner.Model
	tokens               auth.TokenSet
	state                appState
	width                int
	height               int
	cursor               int
	launcherCursor       int
	navCursor            int
	deployCursor         int
	page                 appPage
	focus                focusArea
	filter               string
	filtering            bool
	projects             []api.LiveProject
	projectDetailOpen    bool
	projectDetailButton  int
	routeCheck           api.RouteCheckJob
	routeCheckLoading    bool
	deleteConfirmOpen    bool
	deleteProject        api.LiveProject
	deleteConfirmText    string
	deleteConfirmButton  int
	dbForm               databaseDeployForm
	dbFormOpen           bool
	monolithForm         monolithicDeployForm
	monolithFormOpen     bool
	deployLogOpen        bool
	deployLog            api.DatabaseDeploymentRecord
	deployLogOffset      int
	deployLogFollow      bool
	certificatePathOpen  bool
	certificatePath      string
	clusterLogNamespace  string
	clusterLogRelease    string
	clusterLogTarget     string
	jenkinsLogJob        string
	jenkinsLogQueue      int
	scannerImages        []api.ImageScannerImage
	scannerScans         []api.ImageScanJob
	scannerCursor        int
	scannerHistoryCursor int
	scannerActiveScan    api.ImageScanJob
	scannerReport        string
	scannerReportScanID  string
	scannerReportLoading bool
	scannerLoading       bool
	scannerMode          int
	monitoringOverview   api.MonitoringOverview
	monitoringLoading    bool
	monitoringCursor     int
	logsNamespace        string
	logsPods             []api.PodSummary
	logsLines            []api.LogLine
	logsCursor           int
	logsLoading          bool
	themeIndex           int
	logoutSucceeded      bool
	message              string
	lastRefreshed        time.Time
	userName             string
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

type clusterDeploymentStreamMsg struct {
	tokens      auth.TokenSet
	namespace   string
	releaseName string
	target      string
	deployment  api.ClusterDeploymentRecord
	chunk       api.ClusterDeploymentStreamChunk
	err         error
}

type clusterCertificateDownloadMsg struct {
	tokens auth.TokenSet
	path   string
	err    error
}

type certificatePathChoiceMsg struct {
	path string
	err  error
}

type monolithicDeployResultMsg struct {
	tokens     auth.TokenSet
	deployment api.MonolithicDeploymentRecord
	err        error
}

type jenkinsDeploymentStreamMsg struct {
	tokens  auth.TokenSet
	jobName string
	queueID int
	chunk   api.JenkinsLogStreamChunk
	err     error
}

type microserviceDeployResultMsg struct {
	tokens     auth.TokenSet
	deployment api.MonolithicDeploymentRecord
	err        error
}

type imageScannerLoadMsg struct {
	tokens auth.TokenSet
	images []api.ImageScannerImage
	scans  []api.ImageScanJob
	err    error
}

type imageScanStartMsg struct {
	tokens auth.TokenSet
	scan   api.ImageScanJob
	err    error
}

type imageScanPollMsg struct {
	tokens auth.TokenSet
	scan   api.ImageScanJob
	err    error
}

type imageScanReportMsg struct {
	tokens auth.TokenSet
	scanID string
	report string
	err    error
}

type monitoringLoadMsg struct {
	tokens   auth.TokenSet
	overview api.MonitoringOverview
	err      error
}

type logsLoadMsg struct {
	tokens    auth.TokenSet
	namespace string
	pods      []api.PodSummary
	lines     []api.LogLine
	err       error
}

type projectDeleteResultMsg struct {
	tokens  auth.TokenSet
	project api.LiveProject
	err     error
}

type routeCheckStartMsg struct {
	tokens auth.TokenSet
	job    api.RouteCheckJob
	err    error
}

type routeCheckPollMsg struct {
	tokens auth.TokenSet
	job    api.RouteCheckJob
	err    error
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
	mode         string
	projectName  string
	serviceName  string
	repoURL      string
	repoFullName string
	branch       string
	appPort      string
	framework    string
	serviceType  string
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
		config:           config,
		configErr:        configErr,
		auth:             auth.NewAuthClient(config),
		projectsAPI:      api.NewProjectClient(config.BackendBaseURL),
		scannerAPI:       api.NewImageScannerClient(config.BackendBaseURL),
		observabilityAPI: api.NewObservabilityClient(config.BackendBaseURL),
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
		m.deployLogFollow = true
		m.followLatestDeploymentLog()
		m.clusterLogNamespace = ""
		m.clusterLogRelease = ""
		m.clusterLogTarget = ""
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
		m.deployLogFollow = true
		m.followLatestDeploymentLog()
		m.clusterLogNamespace = msg.deployment.Namespace
		m.clusterLogRelease = msg.deployment.ReleaseName
		m.clusterLogTarget = msg.deployment.TargetClusterName
		m.state = stateReady
		m.message = "Database cluster deployment accepted: " + name
		if m.clusterLogNamespace == "" || m.clusterLogRelease == "" {
			return m, nil
		}
		return m, m.fetchClusterDeploymentStreamCmd(0)
	case clusterDeploymentStreamMsg:
		if !m.deployLogOpen || m.clusterLogNamespace == "" || m.clusterLogRelease == "" {
			return m, nil
		}
		if msg.namespace != m.clusterLogNamespace || msg.releaseName != m.clusterLogRelease {
			return m, nil
		}
		if msg.tokens.AccessToken != "" {
			m.tokens = msg.tokens
		}
		m.deployLog.StatusLog = appendUniqueDeploymentLogLines(m.deployLog.StatusLog, msg.chunk.Lines)
		if msg.deployment.ClusterID != "" {
			m.deployLog.ID = msg.deployment.ClusterID
		}
		if msg.deployment.Status != "" {
			m.deployLog.Status = msg.deployment.Status
			m.deployLog.StatusMessage = msg.deployment.StatusMessage
			m.deployLog.ServiceHost = firstNonEmpty(msg.deployment.ServiceHost, m.deployLog.ServiceHost)
			if msg.deployment.ServicePort > 0 {
				m.deployLog.ServicePort = msg.deployment.ServicePort
			}
		}
		m.syncDeploymentLogOffset()
		if msg.chunk.Completed || api.DatabaseDeploymentTerminal(m.deployLog.Status) {
			if msg.chunk.Completed && !api.DatabaseDeploymentFailed(m.deployLog.Status) {
				m.deployLog.Status = "DEPLOYED"
			}
			name := firstNonEmpty(m.deployLog.ProjectName, m.deployLog.ReleaseName, "cluster")
			if api.DatabaseDeploymentFailed(m.deployLog.Status) {
				m.message = "Database cluster deployment failed: " + name
				return m, nil
			}
			m.deployLog.Status = "DEPLOYED"
			m.deployLog.StatusMessage = "All release pods are Running and Ready."
			m.message = "Database cluster deployment finished: " + name
			return m, nil
		}
		if msg.err != nil {
			m.message = "Cluster deployment stream paused: " + msg.err.Error()
			return m, m.fetchClusterDeploymentStreamCmd(2 * time.Second)
		}
		m.deployLog.Status = "DEPLOYING"
		m.deployLog.StatusMessage = "Waiting for Kubernetes release pods..."
		m.message = "Database cluster deployment is still running..."
		return m, m.fetchClusterDeploymentStreamCmd(500 * time.Millisecond)
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
		m.syncDeploymentLogOffset()
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
	case clusterCertificateDownloadMsg:
		m.certificatePath = ""
		if msg.tokens.AccessToken != "" {
			m.tokens = msg.tokens
		}
		if msg.err != nil {
			m.message = "Certificate download failed: " + msg.err.Error()
			return m, nil
		}
		m.message = "SSL certificate saved: " + msg.path
		return m, nil
	case certificatePathChoiceMsg:
		switch {
		case errors.Is(msg.err, errNativeSaveDialogCancelled):
			m.message = "Certificate download cancelled"
			return m, nil
		case errors.Is(msg.err, errNativeSaveDialogUnavailable):
			m.certificatePathOpen = true
			m.certificatePath = m.defaultClusterCertificatePath()
			m.message = "Native save dialog unavailable. Enter a certificate path."
			return m, nil
		case msg.err != nil:
			m.message = "Could not open save dialog: " + msg.err.Error()
			return m, nil
		}
		m.message = "Downloading SSL certificate..."
		return m, m.downloadClusterCertificateCmd(msg.path, true)
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
		m.deployLogOpen = true
		m.deployLog = monolithicDeploymentLogRecord(msg.deployment)
		m.deployLogFollow = true
		m.followLatestDeploymentLog()
		m.jenkinsLogJob = firstNonEmpty(msg.deployment.JenkinsJobName, "deploy-pipeline")
		m.jenkinsLogQueue = msg.deployment.QueueItemID
		m.state = stateReady
		if msg.deployment.QueueItemID > 0 {
			m.message = fmt.Sprintf("Monolithic deployment queued: %s (#%d)", name, msg.deployment.QueueItemID)
			return m, m.fetchJenkinsDeploymentStreamCmd(0)
		} else {
			m.message = "Monolithic deployment accepted without Jenkins queue log context: " + name
		}
		return m, nil
	case jenkinsDeploymentStreamMsg:
		if !m.deployLogOpen || m.jenkinsLogQueue <= 0 || msg.queueID != m.jenkinsLogQueue {
			return m, nil
		}
		if msg.tokens.AccessToken != "" {
			m.tokens = msg.tokens
		}
		m.deployLog.StatusLog = appendUniqueDeploymentLogLines(m.deployLog.StatusLog, msg.chunk.Lines)
		if msg.chunk.Status != "" {
			m.deployLog.Status = msg.chunk.Status
		}
		if msg.chunk.Message != "" {
			m.deployLog.StatusMessage = msg.chunk.Message
		}
		m.syncDeploymentLogOffset()
		name := firstNonEmpty(m.deployLog.ProjectName, "monolith")
		if msg.chunk.Completed {
			if api.DatabaseDeploymentFailed(m.deployLog.Status) {
				m.message = "Monolithic deployment failed: " + name
			} else {
				m.message = "Monolithic deployment finished: " + name
			}
			return m, nil
		}
		if msg.err != nil {
			m.message = "Jenkins log stream paused: " + msg.err.Error()
			return m, m.fetchJenkinsDeploymentStreamCmd(2 * time.Second)
		}
		m.message = "Monolithic deployment is still running..."
		return m, m.fetchJenkinsDeploymentStreamCmd(500 * time.Millisecond)
	case microserviceDeployResultMsg:
		if msg.tokens.AccessToken != "" {
			m.tokens = msg.tokens
		}
		if msg.err != nil {
			m.state = stateReady
			m.monolithFormOpen = true
			m.message = "Microservices deployment failed: " + msg.err.Error()
			return m, nil
		}
		name := firstNonEmpty(msg.deployment.Name, m.monolithForm.projectName, "microservices")
		m.monolithFormOpen = false
		m.monolithForm = newMonolithicDeployForm()
		m.state = stateLoading
		if msg.deployment.QueueItemID > 0 {
			m.message = fmt.Sprintf("Microservices deployment queued: %s (#%d)", name, msg.deployment.QueueItemID)
		} else {
			m.message = "Microservices deployment queued: " + name
		}
		return m, m.fetchProjectsCmd()
	case imageScannerLoadMsg:
		if msg.tokens.AccessToken != "" {
			m.tokens = msg.tokens
		}
		m.scannerLoading = false
		if msg.err != nil {
			m.message = "Image scanner failed: " + msg.err.Error()
			return m, nil
		}
		m.scannerImages = msg.images
		m.scannerScans = msg.scans
		m.scannerCursor = clamp(m.scannerCursor, 0, max(len(m.scannerImages)-1, 0))
		m.scannerHistoryCursor = clamp(m.scannerHistoryCursor, 0, max(len(m.scannerScans)-1, 0))
		m.message = fmt.Sprintf("Loaded %d images and %d scans", len(m.scannerImages), len(m.scannerScans))
	case imageScanStartMsg:
		if msg.tokens.AccessToken != "" {
			m.tokens = msg.tokens
		}
		m.scannerLoading = false
		if msg.err != nil {
			m.message = "Image scan failed: " + msg.err.Error()
			return m, nil
		}
		m.scannerActiveScan = msg.scan
		m.scannerReport = ""
		m.scannerReportScanID = ""
		m.scannerReportLoading = false
		m.scannerMode = 0
		m.message = imageScanStatusMessage(msg.scan)
		if api.ImageScanTerminal(msg.scan.Status) {
			if !api.ImageScanFailed(msg.scan.Status) && msg.scan.ID != "" {
				m.scannerReportLoading = true
				return m, tea.Batch(m.loadImageScannerCmd(), m.loadImageScanReportCmd(msg.scan.ID))
			}
			return m, m.loadImageScannerCmd()
		}
		return m, m.pollImageScanCmd(msg.scan.ID, 2*time.Second)
	case imageScanPollMsg:
		if msg.tokens.AccessToken != "" {
			m.tokens = msg.tokens
		}
		if msg.err != nil {
			m.scannerLoading = false
			m.message = "Image scan refresh failed: " + msg.err.Error()
			return m, nil
		}
		m.scannerActiveScan = msg.scan
		m.message = imageScanStatusMessage(msg.scan)
		if api.ImageScanTerminal(msg.scan.Status) {
			m.scannerLoading = false
			if !api.ImageScanFailed(msg.scan.Status) && msg.scan.ID != "" {
				m.scannerReportLoading = true
				return m, tea.Batch(m.loadImageScannerCmd(), m.loadImageScanReportCmd(msg.scan.ID))
			}
			return m, m.loadImageScannerCmd()
		}
		return m, m.pollImageScanCmd(msg.scan.ID, 2*time.Second)
	case imageScanReportMsg:
		if msg.tokens.AccessToken != "" {
			m.tokens = msg.tokens
		}
		m.scannerReportLoading = false
		if msg.err != nil {
			m.message = "Scan report unavailable: " + msg.err.Error()
			return m, nil
		}
		m.scannerReportScanID = msg.scanID
		m.scannerReport = msg.report
		m.message = "Scan report loaded"
	case monitoringLoadMsg:
		if msg.tokens.AccessToken != "" {
			m.tokens = msg.tokens
		}
		m.monitoringLoading = false
		if msg.err != nil {
			m.message = "Monitoring failed: " + msg.err.Error()
			return m, nil
		}
		m.monitoringOverview = msg.overview
		m.monitoringCursor = clamp(m.monitoringCursor, 0, max(len(m.monitoringOverview.Projects)-1, 0))
		m.lastRefreshed = time.Now()
		m.message = fmt.Sprintf("Monitoring loaded for %s", firstNonEmpty(msg.overview.Namespace, "workspace"))
	case logsLoadMsg:
		if msg.tokens.AccessToken != "" {
			m.tokens = msg.tokens
		}
		m.logsLoading = false
		if msg.err != nil {
			m.message = "Logs failed: " + msg.err.Error()
			return m, nil
		}
		m.logsNamespace = msg.namespace
		m.logsPods = msg.pods
		m.logsLines = msg.lines
		m.logsCursor = clamp(m.logsCursor, 0, max(len(m.logsPods)-1, 0))
		m.lastRefreshed = time.Now()
		m.message = fmt.Sprintf("Loaded %d pods and %d log lines", len(m.logsPods), len(m.logsLines))
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
	case routeCheckStartMsg:
		if msg.tokens.AccessToken != "" {
			m.tokens = msg.tokens
		}
		if msg.err != nil {
			m.routeCheckLoading = false
			m.message = "Route check failed to start: " + msg.err.Error()
			return m, nil
		}
		m.routeCheck = msg.job
		m.routeCheckLoading = !api.RouteCheckTerminal(msg.job.Status)
		m.message = routeCheckStatusMessage(msg.job)
		if m.routeCheckLoading {
			return m, m.pollRouteCheckCmd(msg.job.ProjectID, msg.job.JobID, 2*time.Second)
		}
	case routeCheckPollMsg:
		if msg.tokens.AccessToken != "" {
			m.tokens = msg.tokens
		}
		if msg.err != nil {
			m.routeCheckLoading = false
			m.message = "Route check refresh failed: " + msg.err.Error()
			return m, nil
		}
		if m.routeCheck.JobID != "" && msg.job.JobID != m.routeCheck.JobID {
			return m, nil
		}
		m.routeCheck = msg.job
		m.routeCheckLoading = !api.RouteCheckTerminal(msg.job.Status)
		m.message = routeCheckStatusMessage(msg.job)
		if m.routeCheckLoading {
			return m, m.pollRouteCheckCmd(msg.job.ProjectID, msg.job.JobID, 2*time.Second)
		}
	}
	return m, nil
}

func (m model) updateKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.deleteConfirmOpen {
		return m.updateDeleteConfirmation(msg)
	}
	if m.certificatePathOpen {
		return m.updateCertificatePath(msg)
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
		if m.state == stateReady && isEnter && m.page == pageImageScanner {
			return m.activateImageScannerSelection()
		}
		if m.state == stateReady && isEnter && m.page == pageLogs {
			return m.activateLogsSelection()
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
			if m.page == pageImageScanner {
				m.scannerLoading = true
				m.message = "Refreshing image scanner..."
				return m, m.loadImageScannerCmd()
			}
			if m.page == pageLogs {
				m.logsLoading = true
				m.message = "Refreshing logs..."
				return m, m.loadLogsCmd()
			}
			if m.page == pageMonitoring {
				m.monitoringLoading = true
				m.message = "Refreshing monitoring..."
				return m, m.loadMonitoringCmd()
			}
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
		} else if m.state == stateReady && m.page == pageImageScanner {
			m.moveImageScannerCursor(-1)
		} else if m.state == stateReady && m.page == pageLogs {
			m.moveLogsCursor(-1)
		} else if m.state == stateReady && m.page == pageMonitoring {
			m.moveMonitoringCursor(-1)
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
		} else if m.state == stateReady && m.page == pageImageScanner {
			m.moveImageScannerCursor(1)
		} else if m.state == stateReady && m.page == pageLogs {
			m.moveLogsCursor(1)
		} else if m.state == stateReady && m.page == pageMonitoring {
			m.moveMonitoringCursor(1)
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
			if m.page == pageImageScanner {
				m.moveImageScannerMode(-1)
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
			if m.page == pageImageScanner {
				m.moveImageScannerMode(1)
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
	case m.certificatePathOpen:
		m.certificatePath = text
		m.message = "Press enter to save the SSL certificate"
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
	return newProjectDeployForm("monolithic")
}

func newMicroservicesDeployForm() monolithicDeployForm {
	return newProjectDeployForm("microservices")
}

func newProjectDeployForm(mode string) monolithicDeployForm {
	local := deploy.DetectLocalProject()
	return monolithicDeployForm{
		mode:        mode,
		projectName: local.Name,
		serviceName: local.Name,
		branch:      local.Branch,
		appPort:     fmt.Sprintf("%d", local.AppPort),
		framework:   local.Framework,
		serviceType: "backend",
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
	fieldCount := m.monolithForm.fieldCount()
	submitIndex := fieldCount - 1

	switch {
	case key == "ctrl+c":
		return m, tea.Quit
	case key == "esc" || code == tea.KeyEscape:
		m.monolithFormOpen = false
		m.message = m.monolithForm.title() + " deployment canceled"
	case key == "tab" || code == tea.KeyTab || code == tea.KeyDown || key == "j":
		m.monolithForm.focus = (m.monolithForm.focus + 1) % fieldCount
	case code == tea.KeyUp || key == "k":
		m.monolithForm.focus = (m.monolithForm.focus + fieldCount - 1) % fieldCount
	case key == "backspace" || key == "ctrl+h" || code == tea.KeyBackspace:
		m.deleteMonolithicFormRune()
	case isEnter:
		if m.monolithForm.focus == submitIndex {
			if m.monolithForm.isMicroservices() {
				return m.submitMicroserviceDeployment()
			}
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
	case 4:
		if m.monolithForm.isMicroservices() {
			m.monolithForm.serviceName += text
		}
	case 5:
		if m.monolithForm.isMicroservices() {
			m.monolithForm.framework += text
		}
	case 6:
		if m.monolithForm.isMicroservices() {
			m.monolithForm.serviceType += text
		}
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
	case 4:
		if m.monolithForm.isMicroservices() {
			m.monolithForm.serviceName = trimLastRune(m.monolithForm.serviceName)
		}
	case 5:
		if m.monolithForm.isMicroservices() {
			m.monolithForm.framework = trimLastRune(m.monolithForm.framework)
		}
	case 6:
		if m.monolithForm.isMicroservices() {
			m.monolithForm.serviceType = trimLastRune(m.monolithForm.serviceType)
		}
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
		namespace, err := projectsAPI.ResolveEffectiveClusterNamespace(ctx, tokens.AccessToken, input.Namespace)
		if api.IsUnauthorized(err) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return clusterDeployResultMsg{tokens: tokens, err: err}
			}
			namespace, err = projectsAPI.ResolveEffectiveClusterNamespace(ctx, tokens.AccessToken, input.Namespace)
		}
		if err != nil {
			return clusterDeployResultMsg{tokens: tokens, err: err}
		}
		input.Namespace = namespace
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

func appendUniqueDeploymentLogLines(current string, next []string) string {
	lines := api.ParseDeploymentLogLines(current)
	seen := make(map[string]bool, len(lines))
	for _, line := range lines {
		seen[line] = true
	}
	for _, line := range next {
		line = strings.TrimSpace(line)
		if line == "" || seen[line] {
			continue
		}
		lines = append(lines, line)
		seen[line] = true
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

func (m model) submitMicroserviceDeployment() (tea.Model, tea.Cmd) {
	input, err := m.monolithForm.microserviceInput()
	if err != nil {
		m.message = err.Error()
		return m, nil
	}
	projectsAPI := m.projectsAPI
	authClient := m.auth
	tokens := m.tokens
	m.message = "Submitting microservices deployment..."
	return m, func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		var err error
		if tokens.ExpiresSoon(time.Now()) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return microserviceDeployResultMsg{tokens: tokens, err: err}
			}
		}
		deployment, err := projectsAPI.CreateMicroserviceDeployment(ctx, tokens.AccessToken, input)
		if api.IsUnauthorized(err) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return microserviceDeployResultMsg{tokens: tokens, err: err}
			}
			deployment, err = projectsAPI.CreateMicroserviceDeployment(ctx, tokens.AccessToken, input)
		}
		return microserviceDeployResultMsg{tokens: tokens, deployment: deployment, err: err}
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
		m.deployLogFollow = false
		m.certificatePathOpen = false
		m.certificatePath = ""
		m.clusterLogNamespace = ""
		m.clusterLogRelease = ""
		m.clusterLogTarget = ""
		m.jenkinsLogJob = ""
		m.jenkinsLogQueue = 0
		m.state = stateLoading
		m.message = "Refreshing live projects..."
		return m, m.fetchProjectsCmd()
	case key == "r":
		if m.jenkinsLogQueue > 0 {
			m.message = "Refreshing Jenkins deployment logs..."
			return m, m.fetchJenkinsDeploymentStreamCmd(0)
		}
		if m.clusterLogNamespace != "" && m.clusterLogRelease != "" {
			m.message = "Refreshing cluster deployment logs..."
			return m, m.fetchClusterDeploymentStreamCmd(0)
		}
		if m.deployLog.ID != "" {
			m.message = "Refreshing deployment logs..."
			return m, m.fetchDatabaseDeploymentCmd(m.deployLog.ID, 0)
		}
	case key == "c":
		if !m.canDownloadClusterCertificate() {
			m.message = "SSL certificate is available after a database cluster deployment succeeds"
			return m, nil
		}
		m.message = "Opening save dialog..."
		return m, chooseCertificatePathCmd(m.defaultClusterCertificatePath())
	case key == "up" || key == "k" || code == tea.KeyUp:
		m.deployLogFollow = false
		m.deployLogOffset = max(m.deployLogOffset-1, 0)
	case key == "down" || key == "j" || code == tea.KeyDown:
		last := m.latestDeploymentLogOffset()
		m.deployLogOffset = min(m.deployLogOffset+1, last)
		m.deployLogFollow = m.deployLogOffset == last
	case key == "end" || key == "G" || code == tea.KeyEnd:
		m.deployLogFollow = true
		m.followLatestDeploymentLog()
	}
	return m, nil
}

func (m model) updateCertificatePath(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	code := msg.Key().Code
	isEnter := key == "enter" || code == tea.KeyEnter || code == tea.KeyReturn
	switch {
	case key == "ctrl+c":
		return m, tea.Quit
	case key == "esc" || code == tea.KeyEscape:
		m.certificatePathOpen = false
		m.certificatePath = ""
		m.message = "Certificate download cancelled"
	case isEnter:
		if strings.TrimSpace(m.certificatePath) == "" {
			m.message = "Enter a certificate file path"
			return m, nil
		}
		m.certificatePathOpen = false
		m.message = "Downloading SSL certificate..."
		return m, m.downloadClusterCertificateCmd(m.certificatePath, false)
	case key == "backspace" || key == "ctrl+h" || code == tea.KeyBackspace:
		m.certificatePath = trimLastRune(m.certificatePath)
	case key == "ctrl+u":
		m.certificatePath = ""
	default:
		if len(key) == 1 && key >= " " && key <= "~" {
			m.certificatePath += key
		}
	}
	return m, nil
}

func (m *model) syncDeploymentLogOffset() {
	if m.deployLogFollow {
		m.followLatestDeploymentLog()
		return
	}
	m.deployLogOffset = clamp(m.deployLogOffset, 0, m.latestDeploymentLogOffset())
}

func (m *model) followLatestDeploymentLog() {
	m.deployLogOffset = m.latestDeploymentLogOffset()
}

func (m model) latestDeploymentLogOffset() int {
	return max(len(api.ParseDeploymentLogLines(m.deployLog.StatusLog))-1, 0)
}

func (m model) canDownloadClusterCertificate() bool {
	return m.deployLog.DeploymentMode == "cluster" &&
		m.clusterLogNamespace != "" &&
		m.deployLog.ID != "" &&
		api.DatabaseDeploymentTerminal(m.deployLog.Status) &&
		!api.DatabaseDeploymentFailed(m.deployLog.Status)
}

func (m model) defaultClusterCertificatePath() string {
	projectName := firstNonEmpty(m.deployLog.ProjectName, m.deployLog.ReleaseName, "database")
	filename := certificateFilename(projectName, ".crt")
	directory, err := os.Getwd()
	if err != nil {
		return filename
	}
	return filepath.Join(directory, filename)
}

func (m model) updateProjectDetail(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	code := msg.Key().Code
	isEnter := key == "enter" || code == tea.KeyEnter || code == tea.KeyReturn
	switch {
	case key == "ctrl+c" || key == "q":
		return m, tea.Quit
	case key == "tab" || code == tea.KeyTab:
		m.projectDetailButton = (m.projectDetailButton + 1) % m.projectDetailActionCount()
		m.message = m.projectDetailButtonMessage()
		return m, nil
	case key == "left" || key == "up" || code == tea.KeyLeft || code == tea.KeyUp:
		m.projectDetailButton = wrapIndex(m.projectDetailButton-1, m.projectDetailActionCount())
		m.message = m.projectDetailButtonMessage()
		return m, nil
	case key == "right" || key == "down" || code == tea.KeyRight || code == tea.KeyDown:
		m.projectDetailButton = wrapIndex(m.projectDetailButton+1, m.projectDetailActionCount())
		m.message = m.projectDetailButtonMessage()
		return m, nil
	case isEnter:
		project, ok := m.selectedProject()
		if !ok {
			return m, nil
		}
		if projectSupportsRouteCheck(project) {
			switch m.projectDetailButton {
			case 0:
				if m.routeCheckLoading {
					m.message = "Route check is already running"
					return m, nil
				}
				m.routeCheckLoading = true
				m.routeCheck = api.RouteCheckJob{ProjectID: project.ID, Status: "RUNNING"}
				m.message = "Starting route check for " + project.Name + "..."
				return m, m.startRouteCheckCmd(project.ID)
			case 1:
				return m.requestProjectDelete()
			default:
				return m.closeProjectDetail()
			}
		}
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
	project, _ := m.selectedProject()
	if projectSupportsRouteCheck(project) {
		switch m.projectDetailButton {
		case 0:
			return "Route check selected"
		case 1:
			return "Delete selected"
		default:
			return "Cancel selected"
		}
	}
	if m.projectDetailButton == 0 {
		return "Delete selected"
	}
	return "Cancel selected"
}

func (m model) projectDetailActionCount() int {
	project, _ := m.selectedProject()
	if projectSupportsRouteCheck(project) {
		return 3
	}
	return 2
}

func projectSupportsRouteCheck(project api.LiveProject) bool {
	kind := strings.ToLower(strings.TrimSpace(firstNonEmpty(project.Kind, project.ArchitectureType)))
	return (kind == "monolith" || kind == "monolithic") && strings.TrimSpace(project.DeployURL) != ""
}

func routeCheckStatusMessage(job api.RouteCheckJob) string {
	switch strings.ToUpper(strings.TrimSpace(job.Status)) {
	case "COMPLETED":
		return fmt.Sprintf("Route check finished: %d passed, %d failed", job.Summary.Passed, job.Summary.Failed)
	case "FAILED":
		return "Route check failed: " + firstNonEmpty(job.ErrorMessage, "unknown error")
	default:
		return "Route check is running..."
	}
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

func (f monolithicDeployForm) microserviceInput() (api.CreateMicroserviceDeploymentInput, error) {
	projectName := strings.TrimSpace(f.projectName)
	serviceName := strings.TrimSpace(f.serviceName)
	repoURL := strings.TrimSpace(f.repoURL)
	branch := strings.TrimSpace(f.branch)
	repoFullName := strings.TrimSpace(f.repoFullName)
	if projectName == "" {
		return api.CreateMicroserviceDeploymentInput{}, fmt.Errorf("Project name is required")
	}
	if serviceName == "" {
		return api.CreateMicroserviceDeploymentInput{}, fmt.Errorf("Service name is required")
	}
	if repoURL == "" {
		return api.CreateMicroserviceDeploymentInput{}, fmt.Errorf("Git remote URL is required")
	}
	if repoFullName == "" {
		return api.CreateMicroserviceDeploymentInput{}, fmt.Errorf("Git remote must include owner and repository")
	}
	if branch == "" {
		branch = "main"
	}
	serviceType := strings.ToLower(strings.TrimSpace(f.serviceType))
	if serviceType == "" {
		serviceType = "backend"
	}
	framework := strings.TrimSpace(f.framework)
	exposePublic := serviceType == "gateway" || serviceType == "frontend" || strings.EqualFold(framework, "Next.js") || strings.EqualFold(framework, "nextjs")
	return api.CreateMicroserviceDeploymentInput{
		ProjectName: projectName,
		Branch:      branch,
		Services: []api.CreateMicroserviceServiceInput{{
			Name:          serviceName,
			RepoURL:       repoURL,
			RepoFullName:  repoFullName,
			RepoProvider:  "github",
			Branch:        branch,
			AppPort:       deploy.ParsePositiveInt(f.appPort, 3000),
			ServiceType:   serviceType,
			Framework:     framework,
			ExposePublic:  exposePublic,
			PrimaryPublic: exposePublic,
		}},
	}, nil
}

func (f monolithicDeployForm) isMicroservices() bool {
	return f.mode == "microservices"
}

func (f monolithicDeployForm) title() string {
	if f.isMicroservices() {
		return "Microservices"
	}
	return "Monolithic"
}

func (f monolithicDeployForm) fieldCount() int {
	if f.isMicroservices() {
		return 8
	}
	return 5
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
	if item.page == pageImageScanner {
		m.scannerLoading = true
		m.message = "Loading image scanner..."
		return m, m.loadImageScannerCmd()
	}
	if item.page == pageLogs {
		m.logsLoading = true
		m.message = "Loading workspace logs..."
		return m, m.loadLogsCmd()
	}
	if item.page == pageMonitoring {
		m.monitoringLoading = true
		m.message = "Loading monitoring overview..."
		return m, m.loadMonitoringCmd()
	}
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
	if feature.Label == "Microservices" {
		return m.openMicroservicesDeployForm(), nil
	}
	m.message = feature.Label + " deployment is coming soon"
	return m, nil
}

func (m model) activateImageScannerSelection() (tea.Model, tea.Cmd) {
	if m.scannerLoading {
		m.message = "Image scanner is still loading"
		return m, nil
	}
	switch m.scannerMode {
	case 1:
		scan, ok := m.selectedScannerHistory()
		if !ok {
			m.message = "No scan history selected"
			return m, nil
		}
		m.scannerActiveScan = scan
		m.scannerMode = 1
		m.message = "Opened scan " + firstNonEmpty(scan.ImageName, scan.ID)
		m.scannerReport = ""
		m.scannerReportScanID = ""
		if api.ImageScanTerminal(scan.Status) && !api.ImageScanFailed(scan.Status) && scan.ID != "" {
			m.scannerReportLoading = true
			m.message = "Loading scan report..."
			return m, m.loadImageScanReportCmd(scan.ID)
		}
		return m, nil
	default:
		image, ok := m.selectedScannerImage()
		if !ok {
			m.message = "No deployed image selected"
			return m, nil
		}
		m.scannerLoading = true
		m.scannerReport = ""
		m.scannerReportScanID = ""
		m.scannerReportLoading = false
		m.scannerMode = 0
		m.message = "Starting image scan for " + imageScannerImageLabel(image)
		return m, m.startImageScanCmd(image)
	}
}

func (m model) activateLogsSelection() (tea.Model, tea.Cmd) {
	if m.logsLoading {
		m.message = "Logs are still loading"
		return m, nil
	}
	pod, ok := m.selectedLogPod()
	if !ok {
		m.message = "No pod selected"
		return m, nil
	}
	m.logsLoading = true
	m.message = "Loading logs for " + pod.Name
	return m, m.loadLogsCmd()
}

func (m model) openProjectDetail() (tea.Model, tea.Cmd) {
	project, ok := m.selectedProject()
	if !ok {
		m.message = "No project selected"
		return m, nil
	}
	if m.routeCheck.ProjectID != project.ID {
		m.routeCheck = api.RouteCheckJob{}
		m.routeCheckLoading = false
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

func (m model) openMicroservicesDeployForm() model {
	m.navCursor = m.navigationIndexByPage(pageDeployment)
	m.dbFormOpen = false
	m.monolithFormOpen = true
	m.monolithForm = newMicroservicesDeployForm()
	m.message = "Deploy one microservice now. Add more services later from the project workspace."
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
	m.deployLogFollow = false
	m.certificatePathOpen = false
	m.certificatePath = ""
	m.clusterLogNamespace = ""
	m.clusterLogRelease = ""
	m.clusterLogTarget = ""
	m.jenkinsLogJob = ""
	m.jenkinsLogQueue = 0
	m.scannerImages = nil
	m.scannerScans = nil
	m.scannerCursor = 0
	m.scannerHistoryCursor = 0
	m.scannerActiveScan = api.ImageScanJob{}
	m.scannerReport = ""
	m.scannerReportScanID = ""
	m.scannerReportLoading = false
	m.scannerLoading = false
	m.scannerMode = 0
	m.monitoringOverview = api.MonitoringOverview{}
	m.monitoringLoading = false
	m.monitoringCursor = 0
	m.logsNamespace = ""
	m.logsPods = nil
	m.logsLines = nil
	m.logsCursor = 0
	m.logsLoading = false
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

func (m model) fetchClusterDeploymentStreamCmd(delay time.Duration) tea.Cmd {
	projectsAPI := m.projectsAPI
	authClient := m.auth
	tokens := m.tokens
	namespace := m.clusterLogNamespace
	releaseName := m.clusterLogRelease
	target := m.clusterLogTarget
	clusterID := m.deployLog.ID
	projectName := m.deployLog.ProjectName
	return func() tea.Msg {
		if delay > 0 {
			time.Sleep(delay)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var err error
		if tokens.ExpiresSoon(time.Now()) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return clusterDeploymentStreamMsg{
					tokens: tokens, namespace: namespace, releaseName: releaseName, target: target, err: err,
				}
			}
		}

		deployment, statusErr := projectsAPI.ResolveClusterDeployment(ctx, tokens.AccessToken, namespace, clusterID, releaseName, projectName)
		if api.IsUnauthorized(statusErr) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return clusterDeploymentStreamMsg{
					tokens: tokens, namespace: namespace, releaseName: releaseName, target: target, err: err,
				}
			}
			deployment, statusErr = projectsAPI.ResolveClusterDeployment(ctx, tokens.AccessToken, namespace, clusterID, releaseName, projectName)
		}
		chunk, err := projectsAPI.FetchClusterDeploymentStreamChunk(ctx, tokens.AccessToken, namespace, releaseName, target)
		if api.IsUnauthorized(err) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err == nil {
				chunk, err = projectsAPI.FetchClusterDeploymentStreamChunk(ctx, tokens.AccessToken, namespace, releaseName, target)
			}
		}
		return clusterDeploymentStreamMsg{
			tokens: tokens, namespace: namespace, releaseName: releaseName, target: target, deployment: deployment, chunk: chunk, err: err,
		}
	}
}

func (m model) fetchJenkinsDeploymentStreamCmd(delay time.Duration) tea.Cmd {
	projectsAPI := m.projectsAPI
	authClient := m.auth
	tokens := m.tokens
	jobName := m.jenkinsLogJob
	queueID := m.jenkinsLogQueue
	return func() tea.Msg {
		if delay > 0 {
			time.Sleep(delay)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()

		var err error
		if tokens.ExpiresSoon(time.Now()) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return jenkinsDeploymentStreamMsg{tokens: tokens, jobName: jobName, queueID: queueID, err: err}
			}
		}
		chunk, err := projectsAPI.FetchJenkinsLogStreamChunk(ctx, tokens.AccessToken, jobName, queueID)
		if api.IsUnauthorized(err) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err == nil {
				chunk, err = projectsAPI.FetchJenkinsLogStreamChunk(ctx, tokens.AccessToken, jobName, queueID)
			}
		}
		return jenkinsDeploymentStreamMsg{tokens: tokens, jobName: jobName, queueID: queueID, chunk: chunk, err: err}
	}
}

func monolithicDeploymentLogRecord(deployment api.MonolithicDeploymentRecord) api.DatabaseDeploymentRecord {
	status := firstNonEmpty(deployment.Status, "QUEUED")
	if !api.DatabaseDeploymentTerminal(status) {
		status = "QUEUED"
	}
	initialLog := "Deployment request accepted."
	if deployment.QueueItemID > 0 {
		initialLog = fmt.Sprintf("Queued in Jenkins as item #%d.", deployment.QueueItemID)
	}
	return api.DatabaseDeploymentRecord{
		ID:             deployment.ProjectID,
		ProjectName:    deployment.Name,
		DeploymentMode: "monolith",
		Status:         status,
		StatusMessage:  initialLog,
		StatusLog:      initialLog,
		DeployURL:      deployment.DeployURL,
	}
}

func (m model) downloadClusterCertificateCmd(destination string, overwrite bool) tea.Cmd {
	projectsAPI := m.projectsAPI
	authClient := m.auth
	tokens := m.tokens
	namespace := m.clusterLogNamespace
	clusterID := m.deployLog.ID
	projectName := firstNonEmpty(m.deployLog.ProjectName, m.deployLog.ReleaseName, "database")
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var err error
		if tokens.ExpiresSoon(time.Now()) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return clusterCertificateDownloadMsg{tokens: tokens, err: err}
			}
		}
		certificate, err := projectsAPI.DownloadClusterCertificate(ctx, tokens.AccessToken, namespace, clusterID)
		if api.IsUnauthorized(err) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err == nil {
				certificate, err = projectsAPI.DownloadClusterCertificate(ctx, tokens.AccessToken, namespace, clusterID)
			}
		}
		if err != nil {
			return clusterCertificateDownloadMsg{tokens: tokens, err: err}
		}

		path, err := saveClusterCertificate(destination, projectName, certificate, overwrite)
		return clusterCertificateDownloadMsg{tokens: tokens, path: path, err: err}
	}
}

var unsafeFilenameCharacters = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func certificateFilename(projectName string, extension string) string {
	baseName := strings.Trim(unsafeFilenameCharacters.ReplaceAllString(strings.TrimSpace(projectName), "-"), "-.")
	if baseName == "" {
		baseName = "database"
	}
	return baseName + "-ca" + extension
}

func saveClusterCertificate(destination string, projectName string, certificate api.ClusterCertificate, overwrite bool) (string, error) {
	extension := strings.ToLower(filepath.Ext(certificate.Filename))
	if extension != ".crt" && extension != ".pem" {
		extension = ".crt"
	}
	path := strings.TrimSpace(destination)
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	path = filepath.Clean(path)
	if info, err := os.Stat(path); err == nil && info.IsDir() {
		path = filepath.Join(path, certificateFilename(projectName, extension))
	} else if strings.HasSuffix(destination, string(os.PathSeparator)) {
		path = filepath.Join(path, certificateFilename(projectName, extension))
	}
	if filepath.Ext(path) == "" {
		path += extension
	}
	flags := os.O_WRONLY | os.O_CREATE | os.O_EXCL
	if overwrite {
		flags = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	file, err := os.OpenFile(path, flags, 0o600)
	if os.IsExist(err) {
		return "", fmt.Errorf("certificate file already exists: %s", path)
	}
	if err != nil {
		return "", fmt.Errorf("create certificate file: %w", err)
	}
	if _, err := file.Write(certificate.Content); err != nil {
		_ = file.Close()
		return "", fmt.Errorf("write certificate file: %w", err)
	}
	if err := file.Close(); err != nil {
		return "", fmt.Errorf("close certificate file: %w", err)
	}
	absolutePath, err := filepath.Abs(path)
	if err != nil {
		return path, nil
	}
	return absolutePath, nil
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

func (m *model) moveImageScannerCursor(delta int) {
	switch m.scannerMode {
	case 1:
		m.scannerHistoryCursor = clamp(m.scannerHistoryCursor+delta, 0, max(len(m.scannerScans)-1, 0))
	default:
		m.scannerCursor = clamp(m.scannerCursor+delta, 0, max(len(m.scannerImages)-1, 0))
	}
}

func (m *model) moveImageScannerMode(delta int) {
	m.scannerMode = wrapIndex(m.scannerMode+delta, 2)
	switch m.scannerMode {
	case 0:
		m.message = "Scan images selected"
	case 1:
		m.message = "Scan history selected"
	}
}

func (m *model) moveLogsCursor(delta int) {
	m.logsCursor = clamp(m.logsCursor+delta, 0, max(len(m.logsPods)-1, 0))
}

func (m *model) moveMonitoringCursor(delta int) {
	m.monitoringCursor = clamp(m.monitoringCursor+delta, 0, max(len(m.monitoringOverview.Projects)-1, 0))
}

func (m model) selectedDeploymentFeature() deploy.Feature {
	if len(deploy.Features) == 0 {
		return deploy.Feature{}
	}
	return deploy.Features[clamp(m.deployCursor, 0, len(deploy.Features)-1)]
}

func (m model) selectedScannerImage() (api.ImageScannerImage, bool) {
	if len(m.scannerImages) == 0 {
		return api.ImageScannerImage{}, false
	}
	return m.scannerImages[clamp(m.scannerCursor, 0, len(m.scannerImages)-1)], true
}

func (m model) selectedScannerHistory() (api.ImageScanJob, bool) {
	if len(m.scannerScans) == 0 {
		return api.ImageScanJob{}, false
	}
	return m.scannerScans[clamp(m.scannerHistoryCursor, 0, len(m.scannerScans)-1)], true
}

func (m model) selectedLogPod() (api.PodSummary, bool) {
	if len(m.logsPods) == 0 {
		return api.PodSummary{}, false
	}
	return m.logsPods[clamp(m.logsCursor, 0, len(m.logsPods)-1)], true
}

func (m model) selectedMonitoringProject() (api.MonitoringProjectMetrics, bool) {
	if len(m.monitoringOverview.Projects) == 0 {
		return api.MonitoringProjectMetrics{}, false
	}
	return m.monitoringOverview.Projects[clamp(m.monitoringCursor, 0, len(m.monitoringOverview.Projects)-1)], true
}

func (m model) resolvedLogsNamespace() string {
	if strings.TrimSpace(m.logsNamespace) != "" {
		return strings.TrimSpace(m.logsNamespace)
	}
	if strings.TrimSpace(m.monitoringOverview.Namespace) != "" {
		return strings.TrimSpace(m.monitoringOverview.Namespace)
	}
	for _, project := range m.projects {
		if strings.TrimSpace(project.Namespace) != "" {
			return strings.TrimSpace(project.Namespace)
		}
	}
	return ""
}

func podExists(pods []api.PodSummary, name string) bool {
	for _, pod := range pods {
		if pod.Name == name {
			return true
		}
	}
	return false
}

func firstRunningPodName(pods []api.PodSummary) string {
	for _, pod := range pods {
		if strings.EqualFold(pod.Phase, "running") {
			return pod.Name
		}
	}
	if len(pods) == 0 {
		return ""
	}
	return pods[0].Name
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

func (m model) loadImageScannerCmd() tea.Cmd {
	scannerAPI := m.scannerAPI
	authClient := m.auth
	tokens := m.tokens
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()

		var err error
		if tokens.ExpiresSoon(time.Now()) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return imageScannerLoadMsg{tokens: tokens, err: err}
			}
		}

		images, err := scannerAPI.ListImages(ctx, tokens.AccessToken)
		if api.IsUnauthorized(err) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return imageScannerLoadMsg{tokens: tokens, err: err}
			}
			images, err = scannerAPI.ListImages(ctx, tokens.AccessToken)
		}
		if err != nil {
			return imageScannerLoadMsg{tokens: tokens, err: err}
		}
		scans, err := scannerAPI.ListScans(ctx, tokens.AccessToken)
		return imageScannerLoadMsg{tokens: tokens, images: images, scans: scans, err: err}
	}
}

func (m model) startImageScanCmd(image api.ImageScannerImage) tea.Cmd {
	scannerAPI := m.scannerAPI
	authClient := m.auth
	tokens := m.tokens
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()

		var err error
		if tokens.ExpiresSoon(time.Now()) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return imageScanStartMsg{tokens: tokens, err: err}
			}
		}
		scan, err := scannerAPI.CreateScan(ctx, tokens.AccessToken, api.CreateImageScanInput{
			SourceKind:  "harbor",
			ImageID:     image.ID,
			ForceRescan: true,
		})
		if api.IsUnauthorized(err) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return imageScanStartMsg{tokens: tokens, err: err}
			}
			scan, err = scannerAPI.CreateScan(ctx, tokens.AccessToken, api.CreateImageScanInput{
				SourceKind:  "harbor",
				ImageID:     image.ID,
				ForceRescan: true,
			})
		}
		return imageScanStartMsg{tokens: tokens, scan: scan, err: err}
	}
}

func (m model) pollImageScanCmd(scanID string, delay time.Duration) tea.Cmd {
	scannerAPI := m.scannerAPI
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
				return imageScanPollMsg{tokens: tokens, err: err}
			}
		}
		scan, err := scannerAPI.GetScan(ctx, tokens.AccessToken, scanID)
		if api.IsUnauthorized(err) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return imageScanPollMsg{tokens: tokens, err: err}
			}
			scan, err = scannerAPI.GetScan(ctx, tokens.AccessToken, scanID)
		}
		return imageScanPollMsg{tokens: tokens, scan: scan, err: err}
	}
}

func (m model) loadImageScanReportCmd(scanID string) tea.Cmd {
	scannerAPI := m.scannerAPI
	authClient := m.auth
	tokens := m.tokens
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var err error
		if tokens.ExpiresSoon(time.Now()) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return imageScanReportMsg{tokens: tokens, scanID: scanID, err: err}
			}
		}
		report, err := scannerAPI.GetScanReport(ctx, tokens.AccessToken, scanID)
		if api.IsUnauthorized(err) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return imageScanReportMsg{tokens: tokens, scanID: scanID, err: err}
			}
			report, err = scannerAPI.GetScanReport(ctx, tokens.AccessToken, scanID)
		}
		return imageScanReportMsg{tokens: tokens, scanID: scanID, report: report, err: err}
	}
}

func (m model) loadMonitoringCmd() tea.Cmd {
	observabilityAPI := m.observabilityAPI
	authClient := m.auth
	tokens := m.tokens
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		var err error
		if tokens.ExpiresSoon(time.Now()) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return monitoringLoadMsg{tokens: tokens, err: err}
			}
		}
		overview, err := observabilityAPI.MonitoringOverview(ctx, tokens.AccessToken)
		if api.IsUnauthorized(err) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return monitoringLoadMsg{tokens: tokens, err: err}
			}
			overview, err = observabilityAPI.MonitoringOverview(ctx, tokens.AccessToken)
		}
		return monitoringLoadMsg{tokens: tokens, overview: overview, err: err}
	}
}

func (m model) startRouteCheckCmd(projectID string) tea.Cmd {
	projectsAPI := m.projectsAPI
	authClient := m.auth
	tokens := m.tokens
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()

		var err error
		if tokens.ExpiresSoon(time.Now()) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return routeCheckStartMsg{tokens: tokens, err: err}
			}
		}
		input := api.RouteCheckInput{MaxRoutes: 50, MaxDepth: 2, TimeoutMS: 10000}
		job, err := projectsAPI.StartRouteCheck(ctx, tokens.AccessToken, projectID, input)
		if api.IsUnauthorized(err) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return routeCheckStartMsg{tokens: tokens, err: err}
			}
			job, err = projectsAPI.StartRouteCheck(ctx, tokens.AccessToken, projectID, input)
		}
		return routeCheckStartMsg{tokens: tokens, job: job, err: err}
	}
}

func (m model) pollRouteCheckCmd(projectID, jobID string, delay time.Duration) tea.Cmd {
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
				return routeCheckPollMsg{tokens: tokens, err: err}
			}
		}
		job, err := projectsAPI.GetRouteCheck(ctx, tokens.AccessToken, projectID, jobID)
		if api.IsUnauthorized(err) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return routeCheckPollMsg{tokens: tokens, err: err}
			}
			job, err = projectsAPI.GetRouteCheck(ctx, tokens.AccessToken, projectID, jobID)
		}
		return routeCheckPollMsg{tokens: tokens, job: job, err: err}
	}
}

func (m model) loadLogsCmd() tea.Cmd {
	observabilityAPI := m.observabilityAPI
	authClient := m.auth
	tokens := m.tokens
	namespace := m.resolvedLogsNamespace()
	selectedPodName := ""
	if pod, ok := m.selectedLogPod(); ok {
		selectedPodName = pod.Name
	}
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
		defer cancel()

		var err error
		if tokens.ExpiresSoon(time.Now()) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return logsLoadMsg{tokens: tokens, namespace: namespace, err: err}
			}
		}
		if namespace == "" {
			overview, overviewErr := observabilityAPI.MonitoringOverview(ctx, tokens.AccessToken)
			if api.IsUnauthorized(overviewErr) && tokens.CanRefresh() {
				tokens, err = authClient.Refresh(ctx, tokens)
				if err != nil {
					return logsLoadMsg{tokens: tokens, err: err}
				}
				overview, overviewErr = observabilityAPI.MonitoringOverview(ctx, tokens.AccessToken)
			}
			if overviewErr == nil {
				namespace = strings.TrimSpace(overview.Namespace)
			}
		}
		if namespace == "" {
			return logsLoadMsg{tokens: tokens, err: fmt.Errorf("no workspace namespace found yet")}
		}
		pods, err := observabilityAPI.ListPods(ctx, tokens.AccessToken, namespace, "primary")
		if api.IsUnauthorized(err) && tokens.CanRefresh() {
			tokens, err = authClient.Refresh(ctx, tokens)
			if err != nil {
				return logsLoadMsg{tokens: tokens, namespace: namespace, err: err}
			}
			pods, err = observabilityAPI.ListPods(ctx, tokens.AccessToken, namespace, "primary")
		}
		if err != nil {
			return logsLoadMsg{tokens: tokens, namespace: namespace, err: err}
		}
		podName := selectedPodName
		if podName == "" || !podExists(pods.Pods, podName) {
			podName = firstRunningPodName(pods.Pods)
		}
		var lines []api.LogLine
		if podName != "" {
			logCtx, logCancel := context.WithTimeout(ctx, 8*time.Second)
			lines, err = observabilityAPI.FetchPodLogs(logCtx, tokens.AccessToken, namespace, podName, "primary", 160)
			logCancel()
			if api.IsUnauthorized(err) && tokens.CanRefresh() {
				tokens, err = authClient.Refresh(ctx, tokens)
				if err != nil {
					return logsLoadMsg{tokens: tokens, namespace: namespace, pods: pods.Pods, err: err}
				}
				logCtx, logCancel = context.WithTimeout(ctx, 8*time.Second)
				lines, err = observabilityAPI.FetchPodLogs(logCtx, tokens.AccessToken, namespace, podName, "primary", 160)
				logCancel()
			}
			if err != nil && len(lines) == 0 {
				return logsLoadMsg{tokens: tokens, namespace: namespace, pods: pods.Pods, err: err}
			}
		}
		return logsLoadMsg{tokens: tokens, namespace: namespace, pods: pods.Pods, lines: lines}
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

func imageScannerImageLabel(image api.ImageScannerImage) string {
	name := firstNonEmpty(image.Name, image.Repository, "image")
	tag := firstNonEmpty(image.Tag, "latest")
	return name + ":" + tag
}

func imageScanTitle(scan api.ImageScanJob) string {
	name := firstNonEmpty(scan.ImageName, "image")
	tag := firstNonEmpty(scan.ImageTag, "latest")
	return name + ":" + tag
}

func imageScanStatusMessage(scan api.ImageScanJob) string {
	status := strings.ToUpper(firstNonEmpty(scan.Status, "PENDING"))
	if api.ImageScanFailed(status) {
		return "Image scan failed: " + firstNonEmpty(scan.StatusMessage, imageScanTitle(scan))
	}
	if api.ImageScanTerminal(status) {
		counts := api.ImageScanSeverityCounts(scan.Vulnerabilities)
		return fmt.Sprintf("Image scan completed: %d critical, %d high", counts["CRITICAL"], counts["HIGH"])
	}
	if scan.Progress > 0 {
		return fmt.Sprintf("Image scan running: %d%%", scan.Progress)
	}
	return "Image scan running..."
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
