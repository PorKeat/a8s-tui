package ui

import (
	"context"
	"github.com/PorKeat/a8s-tui/api"
	"github.com/PorKeat/a8s-tui/auth"
	"github.com/PorKeat/a8s-tui/config"
	"github.com/PorKeat/a8s-tui/ui/features/settings"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestModelMoveAndFilter(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.focus = focusList
	m.projects = []api.LiveProject{
		{Name: "Frontend", Kind: "monolith", Status: "DEPLOYED"},
		{Name: "Orders", Kind: "database", Status: "RUNNING", Engine: "postgres"},
	}

	next, _ := m.updateKey(keyMsg("j"))
	m = next.(model)
	if m.cursor != 1 {
		t.Fatalf("cursor = %d", m.cursor)
	}

	next, _ = m.updateKey(keyMsg("/"))
	m = next.(model)
	next, _ = m.updateKey(keyMsg("p"))
	m = next.(model)
	next, _ = m.updateKey(keyMsg("o"))
	m = next.(model)
	next, _ = m.updateKey(keyMsg("enter"))
	m = next.(model)

	visible := m.visibleProjects()
	if len(visible) != 1 || visible[0].Name != "Orders" {
		t.Fatalf("visible projects = %#v", visible)
	}
}

func TestSlashStartsProjectFilterFromAnyDashboardPage(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.page = pageDeployment
	m.navCursor = m.navigationIndexByPage(pageDeployment)
	m.projects = []api.LiveProject{
		{Name: "Frontend", Kind: "monolith", Status: "DEPLOYED"},
		{Name: "Orders", Kind: "database", Status: "RUNNING", Engine: "postgres"},
	}

	next, _ := m.updateKey(keyMsg("/"))
	m = next.(model)
	if m.page != pageProjects || !m.filtering {
		t.Fatalf("expected project filtering, page=%d filtering=%v", m.page, m.filtering)
	}

	next, _ = m.updateKey(keyMsg("o"))
	m = next.(model)
	next, _ = m.updateKey(keyMsg("r"))
	m = next.(model)
	next, _ = m.updateKey(keyMsg("enter"))
	m = next.(model)

	visible := m.visibleProjects()
	if len(visible) != 1 || visible[0].Name != "Orders" {
		t.Fatalf("visible projects = %#v", visible)
	}
}

func TestTabCyclesFocus(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	start := m.focus
	next, _ := m.updateKey(specialKeyMsg(tea.KeyTab))
	m = next.(model)
	if m.focus == start {
		t.Fatal("expected focus to change")
	}
}

func TestLauncherArrowSelection(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	items := m.launcherItems()
	if len(items) != 2 || items[0].action != "login" || items[1].action != "quit" {
		t.Fatalf("launcher items = %#v", items)
	}
	next, _ := m.updateKey(specialKeyMsg(tea.KeyDown))
	m = next.(model)
	if m.launcherCursor != 1 {
		t.Fatalf("launcher cursor = %d", m.launcherCursor)
	}

	next, _ = m.updateKey(specialKeyMsg(tea.KeyUp))
	m = next.(model)
	if m.launcherCursor != 0 {
		t.Fatalf("launcher cursor = %d", m.launcherCursor)
	}
}

func TestProjectsLoadStartsOutsideProjectWorkspace(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	next, _ := m.Update(projectsResultMsg{
		tokens:   auth.TokenSet{AccessToken: "access"},
		projects: []api.LiveProject{{ID: "project-1", Name: "web", Kind: "monolith", Status: "DEPLOYED"}},
	})
	m = next.(model)
	if m.state != stateReady || m.page != pageProjects || m.focus != focusSidebar || m.navCursor != 0 {
		t.Fatalf("expected ready state outside project workspace: %#v", m)
	}

	next, cmd := m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd != nil || m.focus != focusList || m.page != pageProjects {
		t.Fatalf("expected enter to join project workspace: %#v cmd=%v", m, cmd)
	}
}

func TestInitialThemeDefaultsToOrange(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	if m.themeLabel() != "Orange" {
		t.Fatalf("expected Orange default theme, got %q", m.themeLabel())
	}
}

func TestLoggedInNavigationSections(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.focus = focusSidebar
	m.projects = []api.LiveProject{{ID: "project-1", Name: "web", Kind: "monolith", Status: "DEPLOYED"}}

	next, _ := m.updateKey(keyMsg("j"))
	m = next.(model)
	if m.navCursor != 1 {
		t.Fatalf("nav cursor = %d", m.navCursor)
	}

	next, _ = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if m.page != pageDeployment || m.navCursor != 1 {
		t.Fatalf("expected deployment page: %#v", m)
	}

	next, cmd := m.updateKey(keyMsg("esc"))
	m = next.(model)
	if cmd != nil || m.page != pageDeployment || m.focus != focusSidebar {
		t.Fatalf("expected esc to leave deployment workspace: %#v cmd=%v", m, cmd)
	}

	m.focus = focusSidebar
	m.navCursor = 0
	next, _ = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if m.page != pageProjects || m.focus != focusList {
		t.Fatalf("expected projects page to focus project list: %#v", m)
	}
	next, cmd = m.updateKey(keyMsg("esc"))
	m = next.(model)
	if cmd != nil || m.page != pageProjects || m.focus != focusSidebar {
		t.Fatalf("expected esc to leave project workspace: %#v cmd=%v", m, cmd)
	}
	next, _ = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if m.page != pageProjects || m.focus != focusList {
		t.Fatalf("expected enter to rejoin project workspace: %#v", m)
	}
	next, _ = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if !m.projectDetailOpen {
		t.Fatalf("expected enter to open project detail after selecting Projects: %#v", m)
	}
	m.projectDetailOpen = false

	m.focus = focusSidebar
	m.moveNavigationCursor(1)

	next, _ = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if m.page != pageDeployment {
		t.Fatalf("expected deployment page after reopening: %#v", m)
	}

	next, _ = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if !m.dbFormOpen {
		t.Fatalf("expected deployment form to open: %#v", m)
	}

	m.dbFormOpen = false
	m.deployCursor = 1
	next, _ = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if !m.dbFormOpen || m.dbForm.modeOrDefault() != "cluster" {
		t.Fatalf("expected database cluster form to open: %#v", m)
	}

	m.dbFormOpen = false
	next, _ = m.updateKey(keyMsg("i"))
	m = next.(model)
	if m.page != pageDeployment || m.navCursor != 2 || m.focus != focusSidebar {
		t.Fatalf("expected image scanner shortcut to select sidebar item: %#v", m)
	}
	next, _ = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if m.page != pageImageScanner || m.navCursor != 2 {
		t.Fatalf("expected enter to open image scanner: %#v", m)
	}
}

func TestProjectArrowSelectionDoesNotLeaveWorkspace(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.focus = focusList
	m.projects = []api.LiveProject{
		{Name: "Frontend", Kind: "monolith", Status: "DEPLOYED"},
		{Name: "Orders", Kind: "database", Status: "RUNNING"},
	}

	next, _ := m.updateKey(specialKeyMsg(tea.KeyDown))
	m = next.(model)
	if m.cursor != 1 {
		t.Fatalf("cursor = %d", m.cursor)
	}

	next, _ = m.updateKey(specialKeyMsg(tea.KeyRight))
	m = next.(model)
	if m.focus != focusList {
		t.Fatalf("focus = %d", m.focus)
	}

	next, _ = m.updateKey(specialKeyMsg(tea.KeyLeft))
	m = next.(model)
	if m.focus != focusList {
		t.Fatalf("focus = %d", m.focus)
	}
}

func TestLeftRightStayInSidebarSelectionMode(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.page = pageProjects
	m.focus = focusSidebar
	m.navCursor = m.navigationIndexByPage(pageImageScanner)

	next, cmd := m.updateKey(specialKeyMsg(tea.KeyRight))
	m = next.(model)
	if cmd != nil || m.focus != focusSidebar || m.page != pageProjects || m.navCursor != m.navigationIndexByPage(pageImageScanner) {
		t.Fatalf("expected right to keep sidebar selection mode: %#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(specialKeyMsg(tea.KeyLeft))
	m = next.(model)
	if cmd != nil || m.focus != focusSidebar || m.page != pageProjects || m.navCursor != m.navigationIndexByPage(pageImageScanner) {
		t.Fatalf("expected left to keep sidebar selection mode: %#v cmd=%v", m, cmd)
	}
}

func TestDeploymentLeftRightDoesNotLeaveWorkspace(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.page = pageDeployment
	m.focus = focusList
	m.navCursor = m.navigationIndexByPage(pageDeployment)

	next, cmd := m.updateKey(specialKeyMsg(tea.KeyLeft))
	m = next.(model)
	if cmd != nil || m.page != pageDeployment || m.focus != focusList {
		t.Fatalf("expected left to stay in deployment workspace: %#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(specialKeyMsg(tea.KeyRight))
	m = next.(model)
	if cmd != nil || m.page != pageDeployment || m.focus != focusList {
		t.Fatalf("expected right to stay in deployment workspace: %#v cmd=%v", m, cmd)
	}
}

func TestImageScannerNavigationAndScanStart(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.page = pageImageScanner
	m.focus = focusList
	m.scannerImages = []api.ImageScannerImage{
		{ID: "image-1", Name: "api", Tag: "v1"},
		{ID: "image-2", Name: "worker", Tag: "v2"},
	}
	m.scannerScans = []api.ImageScanJob{{ID: "scan-1", ImageName: "api", ImageTag: "v1", Status: "COMPLETED"}}

	next, cmd := m.updateKey(specialKeyMsg(tea.KeyDown))
	m = next.(model)
	if cmd != nil || m.scannerCursor != 1 {
		t.Fatalf("expected image cursor to move: %#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(specialKeyMsg(tea.KeyRight))
	m = next.(model)
	if cmd != nil || m.scannerMode != 1 {
		t.Fatalf("expected scanner mode history: %#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd == nil || m.scannerMode != 1 || m.scannerActiveScan.ID != "scan-1" || !m.scannerReportLoading {
		t.Fatalf("expected history scan to open: %#v cmd=%v", m, cmd)
	}

	m.scannerMode = 0
	m.scannerReportLoading = false
	next, cmd = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd == nil || !m.scannerLoading || m.scannerMode != 0 {
		t.Fatalf("expected image scan to start: %#v cmd=%v", m, cmd)
	}
}

func TestImageScanReportMessageUpdatesState(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.scannerReportLoading = true

	next, _ := m.Update(imageScanReportMsg{
		tokens: auth.TokenSet{AccessToken: "fresh"},
		scanID: "scan-1",
		report: `{"SchemaVersion":2,"ArtifactName":"api:v1","Results":[]}`,
	})
	m = next.(model)
	if m.scannerReportLoading || m.scannerReportScanID != "scan-1" || !strings.Contains(m.scannerReport, "api:v1") {
		t.Fatalf("report state = %#v", m)
	}
	if m.tokens.AccessToken != "fresh" {
		t.Fatalf("tokens = %#v", m.tokens)
	}
}

func TestObservabilityNavigationAndMessages(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.focus = focusSidebar

	m.navCursor = m.navigationIndexByPage(pageLogs)
	next, cmd := m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd == nil || m.page != pageLogs || !m.logsLoading || m.focus != focusList {
		t.Fatalf("expected logs page load: %#v cmd=%v", m, cmd)
	}

	next, _ = m.Update(logsLoadMsg{
		tokens:    auth.TokenSet{AccessToken: "access"},
		namespace: "workspace-dev",
		pods: []api.PodSummary{
			{Name: "api-0", Phase: "Running"},
			{Name: "worker-0", Phase: "Pending"},
		},
		lines: []api.LogLine{{Pod: "api-0", Level: "success", Message: "server started"}},
	})
	m = next.(model)
	if m.logsLoading || m.logsNamespace != "workspace-dev" || len(m.logsPods) != 2 || len(m.logsLines) != 1 {
		t.Fatalf("logs state = %#v", m)
	}

	next, _ = m.updateKey(keyMsg("j"))
	m = next.(model)
	if m.logsCursor != 1 {
		t.Fatalf("logsCursor = %d", m.logsCursor)
	}

	m.focus = focusSidebar
	m.navCursor = m.navigationIndexByPage(pageMonitoring)
	next, cmd = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd == nil || m.page != pageMonitoring || !m.monitoringLoading {
		t.Fatalf("expected monitoring page load: %#v cmd=%v", m, cmd)
	}
}

func TestMonitoringLoadUpdatesOverview(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	next, _ := m.Update(monitoringLoadMsg{
		tokens: auth.TokenSet{AccessToken: "fresh"},
		overview: api.MonitoringOverview{
			Namespace: "workspace-dev",
			Projects: []api.MonitoringProjectMetrics{
				{Name: "api", TotalPods: 1, RunningPods: 1},
			},
		},
	})
	m = next.(model)
	if m.monitoringLoading || m.monitoringOverview.Namespace != "workspace-dev" || m.tokens.AccessToken != "fresh" {
		t.Fatalf("monitoring state = %#v", m)
	}
}

func TestClusterInputLeavesNamespaceForAuthenticatedResolution(t *testing.T) {
	form := newDatabaseDeployForm()
	form.mode = "cluster"
	form.projectName = "orders-ha"
	form.databaseName = "orders"
	form.username = "orders-user"
	form.password = "secret"

	input, err := form.clusterInput()
	if err != nil {
		t.Fatal(err)
	}
	if input.Namespace != "" {
		t.Fatalf("cluster namespace should be resolved from workspace bootstrap, got %q", input.Namespace)
	}
}

func TestResourceMonitorPercentHelpers(t *testing.T) {
	if got := resourcePercent(4, 8, 0, 0); got != 50 {
		t.Fatalf("resourcePercent with limit = %d", got)
	}
	if got := resourcePercent(0, 0, 2, 8); got != 25 {
		t.Fatalf("resourcePercent fallback = %d", got)
	}
	if got := resourcePercent(12, 8, 0, 0); got != 100 {
		t.Fatalf("resourcePercent clamp = %d", got)
	}
	if got := networkPercent(5*1024*1024, 5*1024*1024); got != 40 {
		t.Fatalf("networkPercent = %d", got)
	}
}

func TestUserSettingsCanToggleTheme(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.page = pageUserSettings
	m.focus = focusList
	m.themeIndex = 0

	next, cmd := m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd != nil || m.themeIndex != 1 {
		t.Fatalf("expected enter to switch to next theme: %#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(keyMsg("t"))
	m = next.(model)
	if cmd != nil || m.themeIndex != 2 {
		t.Fatalf("expected t to switch to next theme: %#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(keyMsg(" "))
	m = next.(model)
	if cmd != nil || m.themeIndex != 3 {
		t.Fatalf("expected space to switch to next theme: %#v cmd=%v", m, cmd)
	}

	m.themeIndex = len(settings.ThemeLabels()) - 1
	next, cmd = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd != nil || m.themeIndex != 0 {
		t.Fatalf("expected theme cycling to wrap: %#v cmd=%v", m, cmd)
	}
}

func TestProjectEnterOpensDetailAndEscCloses(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.focus = focusList
	m.projects = []api.LiveProject{{Name: "mama", Kind: "database", Status: "DEPLOYED", Engine: "PostgreSQL"}}

	next, _ := m.updateKey(keyMsg("enter"))
	m = next.(model)
	if !m.projectDetailOpen {
		t.Fatalf("expected project detail to open: %#v", m)
	}

	next, cmd := m.updateKey(keyMsg("esc"))
	m = next.(model)
	if cmd != nil || m.projectDetailOpen {
		t.Fatalf("expected project detail to close, open=%v cmd=%v", m.projectDetailOpen, cmd)
	}
}

func TestProjectDeleteConfirmationFlow(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.page = pageProjects
	m.focus = focusList
	m.projects = []api.LiveProject{{ID: "project-1", Name: "web", Kind: "monolith", Status: "DEPLOYED"}}

	next, cmd := m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd != nil || !m.projectDetailOpen {
		t.Fatalf("expected project detail, model=%#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(keyMsg("right"))
	m = next.(model)
	if cmd != nil || m.projectDetailButton != 1 {
		t.Fatalf("expected cancel selected on detail, model=%#v cmd=%v", m, cmd)
	}
	next, cmd = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd != nil || m.projectDetailOpen {
		t.Fatalf("expected detail cancel button to close project, model=%#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd != nil || !m.projectDetailOpen || m.projectDetailButton != 0 {
		t.Fatalf("expected detail to reopen with delete selected, model=%#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd != nil || !m.deleteConfirmOpen || m.deleteProject.ID != "project-1" {
		t.Fatalf("expected delete confirmation, model=%#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(keyMsg("tab"))
	m = next.(model)
	if cmd != nil || m.deleteConfirmButton != 1 {
		t.Fatalf("expected esc button selected, model=%#v cmd=%v", m, cmd)
	}
	next, cmd = m.updateKey(keyMsg("left"))
	m = next.(model)
	if cmd != nil || m.deleteConfirmButton != 0 {
		t.Fatalf("expected delete button selected, model=%#v cmd=%v", m, cmd)
	}
	next, cmd = m.updateKey(keyMsg("right"))
	m = next.(model)
	if cmd != nil || m.deleteConfirmButton != 1 {
		t.Fatalf("expected esc button selected with right arrow, model=%#v cmd=%v", m, cmd)
	}
	next, cmd = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd != nil || m.deleteConfirmOpen || m.deleteProject.ID != "" || m.deleteConfirmText != "" {
		t.Fatalf("expected selected esc button to cancel delete, model=%#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd != nil || !m.deleteConfirmOpen || m.deleteConfirmButton != 0 {
		t.Fatalf("expected delete confirmation to reopen on delete button, model=%#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(keyMsg("wrong"))
	m = next.(model)
	if cmd != nil || !m.deleteConfirmOpen || m.state != stateReady {
		t.Fatalf("expected delete to wait for exact project name, model=%#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(keyMsg("esc"))
	m = next.(model)
	if cmd != nil || m.deleteConfirmOpen || m.deleteProject.ID != "" || m.deleteConfirmText != "" {
		t.Fatalf("expected delete cancellation, model=%#v cmd=%v", m, cmd)
	}

	next, _ = m.updateKey(keyMsg("enter"))
	m = next.(model)
	for _, r := range "web" {
		next, _ = m.updateKey(keyMsg(string(r)))
		m = next.(model)
	}
	next, cmd = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd == nil || m.deleteConfirmOpen || m.state != stateLoading {
		t.Fatalf("expected delete command, model=%#v cmd=%v", m, cmd)
	}
}

func TestMonolithProjectDetailStartsRouteCheck(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.page = pageProjects
	m.focus = focusList
	m.projectDetailOpen = true
	m.projects = []api.LiveProject{{ID: "project-1", Name: "web", Kind: "monolith", Status: "DEPLOYED", DeployURL: "https://web.example.com"}}

	next, cmd := m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd == nil || !m.routeCheckLoading || m.routeCheck.ProjectID != "project-1" {
		t.Fatalf("expected route check to start: %#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(specialKeyMsg(tea.KeyRight))
	m = next.(model)
	if cmd != nil || m.projectDetailButton != 1 {
		t.Fatalf("expected delete action to be selected: %#v cmd=%v", m, cmd)
	}
}

func TestRouteCheckMessagesUpdateProjectDetail(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.routeCheckLoading = true
	m.routeCheck = api.RouteCheckJob{JobID: "job-1", ProjectID: "project-1", Status: "RUNNING"}

	next, cmd := m.Update(routeCheckPollMsg{
		tokens: auth.TokenSet{AccessToken: "fresh"},
		job: api.RouteCheckJob{
			JobID:     "job-1",
			ProjectID: "project-1",
			Status:    "COMPLETED",
			Summary:   api.RouteCheckSummary{Passed: 3, Failed: 1},
		},
	})
	m = next.(model)
	if cmd != nil || m.routeCheckLoading || m.routeCheck.Summary.Passed != 3 || !strings.Contains(m.message, "3 passed") {
		t.Fatalf("route check state = %#v cmd=%v", m, cmd)
	}
}

func TestBackspaceOnEmptyMonolithicGitRemoteDoesNotCrash(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.monolithFormOpen = true
	m.monolithForm.repoURL = ""
	m.monolithForm.focus = 1

	next, cmd := m.updateKey(keyMsg("backspace"))
	m = next.(model)
	if cmd != nil {
		t.Fatalf("expected no command, got %v", cmd)
	}
	if m.monolithForm.repoURL != "" {
		t.Fatalf("repoURL = %q", m.monolithForm.repoURL)
	}
}

func TestNewMonolithicDeployFormDoesNotDefaultGitRemote(t *testing.T) {
	form := newMonolithicDeployForm()
	if form.repoURL != "" {
		t.Fatalf("repoURL should start empty, got %q", form.repoURL)
	}
	if form.repoFullName != "" {
		t.Fatalf("repoFullName should start empty, got %q", form.repoFullName)
	}
}

func TestPasteMonolithicGitRemoteURL(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.monolithFormOpen = true
	m.monolithForm.repoURL = ""
	m.monolithForm.repoFullName = ""
	m.monolithForm.focus = 1

	next, cmd := m.Update(tea.PasteMsg{Content: "https://github.com/team/web.git\n"})
	m = next.(model)
	if cmd != nil {
		t.Fatalf("expected no command, got %v", cmd)
	}
	if m.monolithForm.repoURL != "https://github.com/team/web.git" {
		t.Fatalf("repoURL = %q", m.monolithForm.repoURL)
	}
	if m.monolithForm.repoFullName != "team/web" {
		t.Fatalf("repoFullName = %q", m.monolithForm.repoFullName)
	}
}

func TestLogoutClearsSession(t *testing.T) {
	logoutCalled := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/realms/a8s/protocol/openid-connect/logout") {
			t.Fatalf("unexpected logout path %q", r.URL.Path)
		}
		if r.URL.Query().Get("id_token_hint") != "id-token" {
			t.Fatalf("id_token_hint = %q", r.URL.Query().Get("id_token_hint"))
		}
		logoutCalled = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	m := initialModel(config.AppConfig{
		BackendBaseURL:   "http://backend",
		KeycloakURL:      server.URL,
		KeycloakRealm:    "a8s",
		KeycloakClientID: "a8s-tui",
	}, nil)
	openURLCalled := false
	m.auth.OpenURL = func(string) error {
		openURLCalled = true
		return nil
	}
	m.state = stateReady
	m.tokens = auth.TokenSet{AccessToken: "access", IDToken: "id-token"}
	m.projects = []api.LiveProject{{Name: "Frontend", Kind: "monolith", Status: "DEPLOYED"}}
	m.filter = "front"

	next, cmd := m.updateKey(keyMsg("o"))
	m = next.(model)
	if cmd == nil {
		t.Fatal("expected logout command")
	}
	if m.state != stateLoggedOut || !m.logoutSucceeded || m.tokens.AccessToken != "" || len(m.projects) != 0 || m.filter != "" {
		t.Fatalf("session was not cleared: %#v", m)
	}

	msg := cmd()
	if result, ok := msg.(logoutResultMsg); !ok || result.err != nil {
		t.Fatalf("logout result = %#v", msg)
	}
	if openURLCalled {
		t.Fatal("logout should not open the browser")
	}
	if !logoutCalled {
		t.Fatal("expected Keycloak logout endpoint to be called")
	}
}

func TestDatabaseDeployResultOpensLogView(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	next, cmd := m.Update(databaseDeployResultMsg{
		tokens: auth.TokenSet{AccessToken: "access"},
		deployment: api.DatabaseDeploymentRecord{
			ID:          "db-1",
			ProjectName: "orders",
			Status:      "DEPLOYED",
			StatusLog:   "queued\nready",
		},
	})
	m = next.(model)
	if cmd != nil {
		t.Fatal("terminal deployment should not start polling")
	}
	if !m.deployLogOpen || m.deployLog.ID != "db-1" || m.dbFormOpen {
		t.Fatalf("deploy log state = %#v", m)
	}
	if m.tokens.AccessToken != "access" {
		t.Fatalf("tokens = %#v", m.tokens)
	}
}

func TestClusterDeployResultOpensLiveLogView(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	next, cmd := m.Update(clusterDeployResultMsg{
		tokens: auth.TokenSet{AccessToken: "access"},
		deployment: api.ClusterDeploymentRecord{
			ClusterID:         "cluster-1",
			ReleaseName:       "orders-db",
			Name:              "orders",
			Namespace:         "workspace-a",
			TargetClusterName: "primary",
			Status:            "DEPLOYING",
			Stdout:            "GitOps values created.",
		},
	})
	m = next.(model)
	if cmd == nil {
		t.Fatal("cluster deployment should start its live log stream")
	}
	if !m.deployLogOpen || m.deployLog.ID != "cluster-1" || m.clusterLogRelease != "orders-db" {
		t.Fatalf("deploy log state = %#v", m)
	}
}

func TestClusterDeploymentStreamAppendsLogsAndCompletes(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.deployLogOpen = true
	m.clusterLogNamespace = "workspace-a"
	m.clusterLogRelease = "orders-db"
	m.deployLog = api.DatabaseDeploymentRecord{
		ID:          "cluster-1",
		ReleaseName: "orders-db",
		ProjectName: "orders",
		Status:      "DEPLOYING",
		StatusLog:   "GitOps values created.",
	}

	next, cmd := m.Update(clusterDeploymentStreamMsg{
		namespace:   "workspace-a",
		releaseName: "orders-db",
		chunk: api.ClusterDeploymentStreamChunk{
			Lines:     []string{"Release pod readiness: 3/3 ready, 3/3 running.", "Deployment stream completed. All release pods are Running and Ready."},
			Completed: true,
		},
	})
	m = next.(model)
	if cmd != nil {
		t.Fatal("completed cluster deployment should stop streaming")
	}
	if m.deployLog.Status != "DEPLOYED" || !strings.Contains(m.deployLog.StatusLog, "3/3 ready") {
		t.Fatalf("deploy log = %#v", m.deployLog)
	}
}

func TestClusterDeploymentStatusStopsStreamWhenBackendIsDeployed(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.deployLogOpen = true
	m.clusterLogNamespace = "workspace-a"
	m.clusterLogRelease = "orders-db"
	m.deployLog = api.DatabaseDeploymentRecord{
		ID:          "cluster-1",
		ReleaseName: "orders-db",
		ProjectName: "orders",
		Status:      "DEPLOYING",
	}

	next, cmd := m.Update(clusterDeploymentStreamMsg{
		namespace:   "workspace-a",
		releaseName: "orders-db",
		deployment: api.ClusterDeploymentRecord{
			ClusterID:     "cluster-1",
			Status:        "DEPLOYED",
			StatusMessage: "All pods ready",
		},
	})
	m = next.(model)
	if cmd != nil {
		t.Fatal("deployed cluster should stop streaming")
	}
	if m.deployLog.Status != "DEPLOYED" || !strings.Contains(m.message, "finished") {
		t.Fatalf("deploy log state = %#v message=%q", m.deployLog, m.message)
	}
}

func TestClusterDeploymentReadyStatusWinsOverStreamError(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.deployLogOpen = true
	m.clusterLogNamespace = "workspace-a"
	m.clusterLogRelease = "orders-db"
	m.deployLog = api.DatabaseDeploymentRecord{
		ID:             "orders-db",
		ReleaseName:    "orders-db",
		ProjectName:    "orders",
		DeploymentMode: "cluster",
		Status:         "DEPLOYING",
	}

	next, cmd := m.Update(clusterDeploymentStreamMsg{
		namespace:   "workspace-a",
		releaseName: "orders-db",
		deployment: api.ClusterDeploymentRecord{
			ClusterID: "cluster-uuid",
			Status:    "READY",
		},
		err: context.DeadlineExceeded,
	})
	m = next.(model)
	if cmd != nil {
		t.Fatal("ready cluster should stop even when the stream times out")
	}
	if m.deployLog.Status != "DEPLOYED" || m.deployLog.ID != "cluster-uuid" {
		t.Fatalf("deploy log = %#v", m.deployLog)
	}
}

func TestReadyClusterStillCollectsFinalStreamLogs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/cluster/namespaces/workspace-a/clusters/cluster-1":
			_, _ = w.Write([]byte(`{"clusterId":"cluster-1","releaseName":"orders-db","status":"DEPLOYED"}`))
		case "/api/kubernetes/namespaces/workspace-a/releases/orders-db/deployment-stream":
			w.Header().Set("Content-Type", "text/event-stream")
			_, _ = w.Write([]byte("event: console-log\ndata: {\"content\":\"Release pod readiness: 3/3 ready, 3/3 running.\"}\n\n"))
			_, _ = w.Write([]byte("event: console-log\ndata: {\"content\":\"Deployment stream completed. All release pods are Running and Ready.\"}\n\n"))
		default:
			t.Fatalf("path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	m := initialModel(config.AppConfig{BackendBaseURL: server.URL}, nil)
	m.tokens = auth.TokenSet{AccessToken: "access"}
	m.deployLogOpen = true
	m.clusterLogNamespace = "workspace-a"
	m.clusterLogRelease = "orders-db"
	m.deployLog = api.DatabaseDeploymentRecord{ID: "cluster-1", ProjectName: "orders", Status: "DEPLOYING"}

	msg := m.fetchClusterDeploymentStreamCmd(0)()
	result, ok := msg.(clusterDeploymentStreamMsg)
	if !ok {
		t.Fatalf("message = %#v", msg)
	}
	if result.deployment.Status != "DEPLOYED" || !result.chunk.Completed || len(result.chunk.Lines) != 2 {
		t.Fatalf("stream result = %#v", result)
	}
}

func TestDeploymentLogsFollowNewLines(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.deployLogOpen = true
	m.deployLogFollow = true
	m.deployLog = api.DatabaseDeploymentRecord{
		ID:        "db-1",
		Status:    "DEPLOYING",
		StatusLog: "queued\ncreating",
	}
	m.followLatestDeploymentLog()

	next, _ := m.Update(databaseDeploymentPollMsg{
		deployment: api.DatabaseDeploymentRecord{
			ID:        "db-1",
			Status:    "DEPLOYING",
			StatusLog: "queued\ncreating\nready",
		},
	})
	m = next.(model)
	if !m.deployLogFollow || m.deployLogOffset != 2 {
		t.Fatalf("expected logs to follow latest line: follow=%v offset=%d", m.deployLogFollow, m.deployLogOffset)
	}
}

func TestDeploymentLogManualScrollPausesAndResumesFollow(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.deployLogOpen = true
	m.deployLogFollow = true
	m.deployLog = api.DatabaseDeploymentRecord{StatusLog: "one\ntwo\nthree"}
	m.followLatestDeploymentLog()

	next, _ := m.updateDeploymentLog(specialKeyMsg(tea.KeyUp))
	m = next.(model)
	if m.deployLogFollow || m.deployLogOffset != 1 {
		t.Fatalf("expected up to pause following: follow=%v offset=%d", m.deployLogFollow, m.deployLogOffset)
	}

	next, _ = m.updateDeploymentLog(specialKeyMsg(tea.KeyDown))
	m = next.(model)
	if !m.deployLogFollow || m.deployLogOffset != 2 {
		t.Fatalf("expected returning to latest line to resume following: follow=%v offset=%d", m.deployLogFollow, m.deployLogOffset)
	}
}

func TestDeploymentLogSeverityColors(t *testing.T) {
	if deploymentLogColor("success") != fgGreen {
		t.Fatal("success logs should be green")
	}
	if deploymentLogColor("warn") != fgWarn {
		t.Fatal("warning logs should be yellow")
	}
	if deploymentLogColor("error") != fgError {
		t.Fatal("error logs should be red")
	}
	if deploymentLogColor("info") != fgText {
		t.Fatal("informational logs should use the normal text color")
	}
}

func TestDeploymentStatusDisplayShowsLoadingForActiveDeployment(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	active := deploymentStatusDisplay(m, "DEPLOYING", fgWarn)
	if !strings.Contains(active, "DEPLOYING...") {
		t.Fatalf("active status = %q", active)
	}
	complete := deploymentStatusDisplay(m, "DEPLOYED", fgGreen)
	if strings.Contains(complete, "...") {
		t.Fatalf("terminal status should not animate: %q", complete)
	}
}

func TestSaveClusterCertificateUsesChosenPathAndAvoidsOverwrite(t *testing.T) {
	directory := t.TempDir()
	certificate := api.ClusterCertificate{
		Filename: "postgresql-ca.crt",
		Content:  []byte("certificate"),
	}
	chosenPath := filepath.Join(directory, "chosen-name.crt")
	first, err := saveClusterCertificate(chosenPath, "mama deployment", certificate, false)
	if err != nil {
		t.Fatal(err)
	}
	if first != chosenPath {
		t.Fatalf("path = %q", first)
	}
	if _, err := saveClusterCertificate(chosenPath, "mama deployment", certificate, false); err == nil {
		t.Fatal("expected existing certificate path to be rejected")
	}
	info, err := os.Stat(first)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("certificate permissions = %o", info.Mode().Perm())
	}
}

func TestCertificateSaveDialogFallbackSupportsPasteAndSave(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.deployLogOpen = true
	m.clusterLogNamespace = "workspace-a"
	m.deployLog = api.DatabaseDeploymentRecord{ID: "cluster-1", DeploymentMode: "cluster", Status: "DEPLOYED"}

	next, cmd := m.updateDeploymentLog(keyMsg("c"))
	m = next.(model)
	if cmd == nil || m.certificatePathOpen {
		t.Fatalf("expected native save dialog command: %#v cmd=%v", m, cmd)
	}

	next, cmd = m.Update(certificatePathChoiceMsg{err: errNativeSaveDialogUnavailable})
	m = next.(model)
	if cmd != nil || !m.certificatePathOpen || m.certificatePath == "" {
		t.Fatalf("expected certificate path fallback: %#v cmd=%v", m, cmd)
	}

	m.certificatePath = ""
	next, _ = m.updatePaste(tea.PasteMsg{Content: "/tmp/custom-ca.crt\n"})
	m = next.(model)
	if m.certificatePath != "/tmp/custom-ca.crt" {
		t.Fatalf("certificate path = %q", m.certificatePath)
	}

	next, cmd = m.updateCertificatePath(keyMsg("enter"))
	m = next.(model)
	if cmd == nil || m.certificatePathOpen {
		t.Fatalf("expected save command: %#v cmd=%v", m, cmd)
	}
}

func TestCertificateSaveDialogChoiceStartsDownload(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	next, cmd := m.Update(certificatePathChoiceMsg{path: "/tmp/mama-ca.crt"})
	m = next.(model)
	if cmd == nil || !strings.Contains(m.message, "Downloading") {
		t.Fatalf("expected certificate download command: %#v cmd=%v", m, cmd)
	}
}

func TestClusterCertificateAvailableOnlyAfterSuccessfulClusterDeployment(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.clusterLogNamespace = "workspace-a"
	m.deployLog = api.DatabaseDeploymentRecord{ID: "cluster-1", DeploymentMode: "cluster", Status: "DEPLOYING"}
	if m.canDownloadClusterCertificate() {
		t.Fatal("certificate should not be available while deployment is running")
	}
	m.deployLog.Status = "DEPLOYED"
	if !m.canDownloadClusterCertificate() {
		t.Fatal("certificate should be available after a successful cluster deployment")
	}
}

func keyMsg(text string) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Text: text, Code: []rune(text)[0]})
}

func specialKeyMsg(code rune) tea.KeyPressMsg {
	return tea.KeyPressMsg(tea.Key{Code: code})
}
func TestDatabaseDeployFormInputValidation(t *testing.T) {
	form := newDatabaseDeployForm()
	if _, err := form.input(); err == nil {
		t.Fatal("expected required field error")
	}
	form.projectName = "orders"
	form.databaseName = "ordersdb"
	form.username = "orders_user"
	form.password = "secret"
	input, err := form.input()
	if err != nil {
		t.Fatal(err)
	}
	if input.Engine != "postgresql" || input.Version != "18" || input.SizeProfile != "small" {
		t.Fatalf("input = %#v", input)
	}
}
