package ui

import (
	"github.com/PorKeat/a8s-tui/api"
	"github.com/PorKeat/a8s-tui/auth"
	"github.com/PorKeat/a8s-tui/config"
	"github.com/PorKeat/a8s-tui/ui/features/settings"
	"net/http"
	"net/http/httptest"
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
