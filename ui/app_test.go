package ui

import (
	"a8s-tui/api"
	"a8s-tui/auth"
	"a8s-tui/config"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestModelMoveAndFilter(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
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

func TestLoggedInNavigationSections(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.focus = focusSidebar

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
	if cmd != nil || m.page != pageProjects {
		t.Fatalf("expected esc to close section, page=%d cmd=%v", m.page, cmd)
	}
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
	next, _ = m.updateKey(keyMsg("i"))
	m = next.(model)
	if m.page != pageImageScanner || m.navCursor != 2 {
		t.Fatalf("navigation state = %#v", m)
	}
}

func TestProjectArrowSelectionAndFocus(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
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
	if m.focus != focusSidebar {
		t.Fatalf("focus = %d", m.focus)
	}

	next, _ = m.updateKey(specialKeyMsg(tea.KeyLeft))
	m = next.(model)
	if m.focus != focusList {
		t.Fatalf("focus = %d", m.focus)
	}
}

func TestProjectEnterOpensDetailAndEscCloses(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
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

func TestLogoutClearsSession(t *testing.T) {
	m := initialModel(config.AppConfig{
		BackendBaseURL:   "http://backend",
		KeycloakURL:      "https://keycloak.example.com",
		KeycloakRealm:    "a8s",
		KeycloakClientID: "a8s-tui",
	}, nil)
	m.auth.OpenURL = func(string) error { return nil }
	m.state = stateReady
	m.tokens = auth.TokenSet{AccessToken: "access", IDToken: "id-token"}
	m.projects = []api.LiveProject{{Name: "Frontend", Kind: "monolith", Status: "DEPLOYED"}}
	m.filter = "front"

	next, cmd := m.updateKey(keyMsg("o"))
	m = next.(model)
	if cmd == nil {
		t.Fatal("expected logout command")
	}
	if m.state != stateLoggedOut || m.tokens.AccessToken != "" || len(m.projects) != 0 || m.filter != "" {
		t.Fatalf("session was not cleared: %#v", m)
	}

	msg := cmd()
	if result, ok := msg.(logoutResultMsg); !ok || result.err != nil {
		t.Fatalf("logout result = %#v", msg)
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
