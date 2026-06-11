package ui

import (
	"context"
	"fmt"
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

func TestInitialThemeDefaultsToLight(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	if m.themeLabel() != "Light" {
		t.Fatalf("expected Light default theme, got %q", m.themeLabel())
	}
}

func TestLightThemeAppliesLightAndAccessibleSemanticColors(t *testing.T) {
	applyTheme(1)
	if colorPrimary != "#ea580c" {
		t.Fatalf("light theme accent is not orange: %s", colorPrimary)
	}
	if colorBgMain != "#f7f7fb" || colorBgSide != "#f1f2f8" || colorBgCard != "#ffffff" {
		t.Fatalf("light backgrounds not applied: main=%s side=%s card=%s", colorBgMain, colorBgSide, colorBgCard)
	}
	if colorBgDanger == "#3f2638" || colorBgDangerActive == "#6f2d4a" {
		t.Fatalf("light theme kept dark danger backgrounds: %s %s", colorBgDanger, colorBgDangerActive)
	}
	if statusColor("deployed") != colorSuccess || statusColor("failed") != colorError {
		t.Fatal("status colors are not using the active light semantic palette")
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
	if cmd != nil || m.scannerMode != 1 {
		t.Fatalf("expected source selection to move: %#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd != nil || m.focus != focusDetail {
		t.Fatalf("expected selected source to open on right: %#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(specialKeyMsg(tea.KeyEscape))
	m = next.(model)
	if cmd != nil || m.focus != focusList {
		t.Fatalf("expected esc to return to source list: %#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(specialKeyMsg(tea.KeyDown))
	m = next.(model)
	next, cmd = m.updateKey(specialKeyMsg(tea.KeyDown))
	m = next.(model)
	if cmd != nil || m.scannerMode != 3 {
		t.Fatalf("expected scanner mode history: %#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd != nil || m.focus != focusDetail || m.scannerActiveScan.ID != "" {
		t.Fatalf("expected history workspace to open before scan: %#v cmd=%v", m, cmd)
	}
	next, cmd = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd == nil || m.scannerMode != 3 || m.scannerActiveScan.ID != "scan-1" || !m.scannerReportLoading {
		t.Fatalf("expected history scan to open: %#v cmd=%v", m, cmd)
	}

	m.scannerMode = 0
	m.scannerReportLoading = false
	m.scannerActiveScan = api.ImageScanJob{}
	m.focus = focusDetail
	next, cmd = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd == nil || !m.scannerLoading || m.scannerMode != 0 {
		t.Fatalf("expected image scan to start: %#v cmd=%v", m, cmd)
	}
}

func TestImageScannerExternalAndGitSourceInputsMatchFrontendFlow(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.scannerMode = 1
	m.scannerForm.externalRegistry = "ghcr.io"
	m.scannerForm.externalName = "team/api"
	m.scannerForm.externalTag = "v2"

	input, err := m.imageScannerSourceInput(false)
	if err != nil {
		t.Fatal(err)
	}
	if input.SourceKind != "external" || input.ImageRef != "ghcr.io/team/api:v2" || input.ForceRescan {
		t.Fatalf("external input = %#v", input)
	}

	m.scannerMode = 2
	m.scannerForm.gitRepository = "https://github.com/team/orders.git"
	input, err = m.imageScannerSourceInput(false)
	if err != nil {
		t.Fatal(err)
	}
	if input.SourceKind != "git" || input.BranchOrTag != "main" || input.DockerfilePath != "Dockerfile" ||
		input.BuildContext != "." || input.TargetImageName != "orders" {
		t.Fatalf("git input = %#v", input)
	}
}

func TestImageScannerNormalizesDockerHubWebsiteRegistry(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.scannerMode = 1
	m.scannerForm.externalRegistry = "https://hub.docker.com/"
	m.scannerForm.externalName = "autooffensive/autooffensive-frontend"
	m.scannerForm.externalTag = "73cb00d2"

	input, err := m.imageScannerSourceInput(false)
	if err != nil {
		t.Fatal(err)
	}
	if input.RegistryURL != "docker.io" ||
		input.ImageRef != "docker.io/autooffensive/autooffensive-frontend:73cb00d2" {
		t.Fatalf("external input = %#v", input)
	}
}

func TestImageScannerRejectsDockerHubImageWebPageAsRegistry(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.scannerMode = 1
	m.scannerForm.externalRegistry = "https://hub.docker.com/r/autooffensive/autooffensive-frontend"
	m.scannerForm.externalName = "autooffensive/autooffensive-frontend"
	m.scannerForm.externalTag = "latest"

	if _, err := m.imageScannerSourceInput(false); err == nil || !strings.Contains(err.Error(), "use docker.io") {
		t.Fatalf("err = %v", err)
	}
}

func TestImageScannerValidationMovesFocusToInvalidField(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.scannerMode = 1
	m.scannerForm.focus = 4
	m.scannerForm.externalRegistry = "https://hub.docker.com/r/team/api"
	m.scannerForm.externalName = "team/api"
	m.scannerForm.externalTag = "latest"

	next, cmd := m.submitImageScannerForm()
	m = next.(model)
	if cmd != nil || m.scannerForm.focus != 0 || !strings.Contains(m.message, "use docker.io") {
		t.Fatalf("validation state = %#v cmd=%v", m, cmd)
	}
}

func TestImageScannerSourceFormsValidateAndAcceptPaste(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.page = pageImageScanner
	m.focus = focusDetail
	m.scannerMode = 2

	next, _ := m.updatePaste(tea.PasteMsg{Content: "https://github.com/team/platform.git"})
	m = next.(model)
	if m.scannerForm.gitRepository != "https://github.com/team/platform.git" {
		t.Fatalf("git repository = %q", m.scannerForm.gitRepository)
	}

	m.scannerForm.gitPrivate = true
	m.scannerForm.focus = 4
	next, cmd := m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd != nil || !strings.Contains(m.message, "username and token") {
		t.Fatalf("expected private credential validation: %#v cmd=%v", m, cmd)
	}
}

func TestImageScannerTextFieldsAcceptJAndK(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.page = pageImageScanner
	m.focus = focusDetail
	m.scannerMode = 1
	m.scannerForm.focus = 0

	next, _ := m.updateKey(keyMsg("j"))
	m = next.(model)
	next, _ = m.updateKey(keyMsg("k"))
	m = next.(model)
	if m.scannerForm.externalRegistry != "jk" || m.scannerForm.focus != 0 {
		t.Fatalf("external registry text field = %q focus=%d", m.scannerForm.externalRegistry, m.scannerForm.focus)
	}

	m.scannerForm.focus = 3
	next, _ = m.updateKey(keyMsg("j"))
	m = next.(model)
	if m.scannerForm.focus != 4 {
		t.Fatalf("expected j to navigate from non-text field, focus=%d", m.scannerForm.focus)
	}
}

func TestImageScannerRendersSourceFirstWorkflow(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.page = pageImageScanner
	m.focus = focusDetail
	m.scannerMode = 1

	left := strings.Join(m.modernImageScannerList(42, 24), "\n")
	for _, expected := range []string{"SOURCE", "Harbor", "External", "Git", "History", "enter opens"} {
		if !strings.Contains(left, expected) {
			t.Fatalf("scanner source navigation missing %q:\n%s", expected, left)
		}
	}
	if strings.Contains(left, "Registry URL") || strings.Contains(left, "Pull & Scan") {
		t.Fatalf("scanner source navigation should not contain source data:\n%s", left)
	}

	right := strings.Join(m.modernImageScannerDetail(72, 30), "\n")
	for _, expected := range []string{"External registry", "Registry URL", "Pull & Scan", "PREVIEW"} {
		if !strings.Contains(right, expected) {
			t.Fatalf("external scanner workspace missing %q:\n%s", expected, right)
		}
	}

	m.scannerMode = 2
	right = strings.Join(m.modernImageScannerDetail(72, 30), "\n")
	if !strings.Contains(right, "Repository URL") || !strings.Contains(right, "Build & Scan") {
		t.Fatalf("Git source workspace missing:\n%s", right)
	}
}

func TestImageScannerHarborAndHistoryDataRenderOnRight(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.page = pageImageScanner
	m.focus = focusList
	m.scannerImages = []api.ImageScannerImage{{ID: "image-1", Name: "api", Tag: "v1", Repository: "team/api"}}
	m.scannerScans = []api.ImageScanJob{{ID: "scan-1", ImageName: "api", ImageTag: "v1", Status: "COMPLETED"}}

	harbor := strings.Join(m.modernImageScannerDetail(72, 30), "\n")
	if !strings.Contains(harbor, "DEPLOYED IMAGES") || !strings.Contains(harbor, "SELECTED IMAGE") || !strings.Contains(harbor, "api:v1") {
		t.Fatalf("Harbor data should render on right:\n%s", harbor)
	}

	m.scannerMode = 3
	history := strings.Join(m.modernImageScannerDetail(72, 30), "\n")
	if !strings.Contains(history, "Scan history") || !strings.Contains(history, "SELECTED SCAN") || !strings.Contains(history, "completed") {
		t.Fatalf("history data should render on right:\n%s", history)
	}
}

func TestImageScannerSourceRowsKeepDataOnOneLine(t *testing.T) {
	for _, test := range []struct {
		label  string
		state  string
		active bool
	}{
		{label: "Harbor", state: "0 images", active: true},
		{label: "External", state: "pull & scan"},
		{label: "Git", state: "build & scan"},
		{label: "History", state: "9 scans"},
	} {
		lines := modernImageScannerSourceRow(test.label, test.state, 42, test.active)
		if len(lines) != 3 {
			t.Fatalf("%s source row wrapped to %d lines:\n%s", test.label, len(lines), strings.Join(lines, "\n"))
		}
		if !strings.Contains(lines[1], test.label) || !strings.Contains(lines[1], test.state) {
			t.Fatalf("%s source row data is not on one line:\n%s", test.label, strings.Join(lines, "\n"))
		}
	}
}

func TestImageScannerFooterAndFailedResultAreContextual(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.page = pageImageScanner
	m.focus = focusDetail
	m.scannerMode = 1

	footer := m.modernFooter(120)
	if !strings.Contains(footer, "field") || !strings.Contains(footer, "toggle") || strings.Contains(footer, "search") {
		t.Fatalf("scanner form footer = %q", footer)
	}

	m.scannerActiveScan = api.ImageScanJob{
		ID:            "scan-failed",
		Status:        "FAILED",
		FullReference: "docker.io/team/very-long-image-name:latest",
		StatusMessage: "Jenkins image scan failed before Trivy returned a report.",
	}
	rendered := strings.Join(m.modernImageScanResult(m.scannerActiveScan, 54, 30), "\n")
	if !strings.Contains(rendered, "Scan stopped before Trivy returned findings") ||
		!strings.Contains(rendered, "Jenkins image scan failed") {
		t.Fatalf("failed result UI:\n%s", rendered)
	}
}

func TestImageScannerEscReturnsToSourcesAndCanOpenAnotherSource(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.page = pageImageScanner
	m.focus = focusDetail
	m.scannerMode = 1
	m.scannerActiveScan = api.ImageScanJob{ID: "scan-failed", Status: "FAILED"}
	m.scannerReport = `{"failed":true}`
	m.scannerReportScanID = "scan-failed"

	next, cmd := m.updateKey(specialKeyMsg(tea.KeyEscape))
	m = next.(model)
	if cmd != nil || m.focus != focusList || m.scannerMode != 1 {
		t.Fatalf("esc should return to source list: %#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(specialKeyMsg(tea.KeyDown))
	m = next.(model)
	if cmd != nil || m.scannerMode != 2 || m.scannerActiveScan.ID != "" || m.scannerReport != "" {
		t.Fatalf("switching source should clear stale result: %#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(keyMsg("enter"))
	m = next.(model)
	if cmd != nil || m.focus != focusDetail || m.scannerMode != 2 {
		t.Fatalf("enter should open newly selected source: %#v cmd=%v", m, cmd)
	}
}

func TestImageScannerResultCanForceRescanOrChooseAnotherSource(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.page = pageImageScanner
	m.focus = focusDetail
	m.scannerMode = 0
	m.scannerImages = []api.ImageScannerImage{{ID: "image-1", Name: "api", Tag: "v1"}}
	m.scannerActiveScan = api.ImageScanJob{ID: "scan-1", SourceKind: "harbor", Status: "COMPLETED"}

	next, cmd := m.updateKey(keyMsg("x"))
	m = next.(model)
	if cmd == nil || !m.scannerLoading || m.message != "Starting forced rescan..." {
		t.Fatalf("expected forced rescan: %#v cmd=%v", m, cmd)
	}

	next, cmd = m.updateKey(keyMsg("n"))
	m = next.(model)
	if cmd != nil || m.scannerActiveScan.ID != "" || m.scannerMode != 0 {
		t.Fatalf("expected source reset: %#v cmd=%v", m, cmd)
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

func TestMonitoringFallbackUsesKubernetesPodsAndLiveProjects(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/kubernetes/namespaces/workspace-dev/pods" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{
			"namespace":"workspace-dev",
			"total":1,
			"pods":[{"name":"orders-db-0","phase":"Running","restartCount":2}]
		}`)
	}))
	defer server.Close()

	overview := applyMonitoringFallback(
		context.Background(),
		api.NewObservabilityClient(server.URL),
		api.NewProjectClient(server.URL),
		"token",
		api.MonitoringOverview{Namespace: "workspace-dev"},
		[]api.LiveProject{{ID: "db-1", Name: "orders-db", Kind: "database", Status: "DEPLOYED", Namespace: "workspace-dev"}},
	)
	if !overview.PodFallbackUsed || !overview.TelemetryPending ||
		overview.NamespaceMetrics.TotalPods != 1 || overview.NamespaceMetrics.RunningPods != 1 ||
		len(overview.Projects) != 1 || overview.Projects[0].RunningPods != 1 {
		t.Fatalf("fallback overview = %#v", overview)
	}
}

func TestMonitoringFallbackChecksProjectTargetCluster(t *testing.T) {
	var targets []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Query().Get("targetClusterName")
		targets = append(targets, target)
		w.Header().Set("Content-Type", "application/json")
		if target == "edge-cluster" {
			fmt.Fprint(w, `{"namespace":"workspace-dev","total":1,"pods":[{"name":"orders-db-0","phase":"Running"}]}`)
			return
		}
		fmt.Fprint(w, `{"namespace":"workspace-dev","total":0,"pods":[]}`)
	}))
	defer server.Close()

	overview := applyMonitoringFallback(
		context.Background(),
		api.NewObservabilityClient(server.URL),
		api.NewProjectClient(server.URL),
		"token",
		api.MonitoringOverview{Namespace: "workspace-dev"},
		[]api.LiveProject{{
			ID:                "db-1",
			Name:              "orders-db",
			Kind:              "database",
			Status:            "DEPLOYED",
			Namespace:         "workspace-dev",
			TargetClusterName: "edge-cluster",
		}},
	)
	if len(targets) != 1 || targets[0] != "edge-cluster" {
		t.Fatalf("target clusters checked = %#v", targets)
	}
	if !overview.PodFallbackUsed || !overview.TelemetryPending || overview.TelemetryUnavailable {
		t.Fatalf("fallback overview = %#v", overview)
	}
}

func TestMonitoringFallbackMarksUnavailableWhenNoPodsOrTelemetryExist(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"namespace":"workspace-dev","total":0,"pods":[]}`)
	}))
	defer server.Close()

	overview := applyMonitoringFallback(
		context.Background(),
		api.NewObservabilityClient(server.URL),
		api.NewProjectClient(server.URL),
		"token",
		api.MonitoringOverview{Namespace: "workspace-dev"},
		[]api.LiveProject{{ID: "db-1", Name: "orders-db", Kind: "database", Status: "DEPLOYED", Namespace: "workspace-dev"}},
	)
	if overview.TelemetryPending || !overview.TelemetryUnavailable {
		t.Fatalf("fallback overview = %#v", overview)
	}
}

func TestMonitoringFallbackUsesSingleDatabaseMetrics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/database-deployments/db-1/metrics":
			fmt.Fprint(w, `{
				"deploymentId":"db-1",
				"namespace":"workspace-dev",
				"podName":"orders-db-0",
				"podPhase":"Running",
				"readyReplicas":1,
				"replicas":1,
				"restartCount":2,
				"cpuRequest":"250m",
				"cpuLimit":"1",
				"memoryRequest":"512Mi",
				"memoryLimit":"1Gi",
				"storageRequested":"10Gi",
				"storageQuotaLimit":"20Gi"
			}`)
		default:
			t.Fatalf("unexpected path = %q", r.URL.Path)
		}
	}))
	defer server.Close()

	overview := applyMonitoringFallback(
		context.Background(),
		api.NewObservabilityClient(server.URL),
		api.NewProjectClient(server.URL),
		"token",
		api.MonitoringOverview{Namespace: "workspace-dev"},
		[]api.LiveProject{{
			ID:                    "db-project",
			Name:                  "orders-db",
			Kind:                  "database",
			Status:                "DEPLOYED",
			Namespace:             "workspace-dev",
			DatabaseDeploymentIDs: []string{"db-1"},
		}},
	)
	if overview.TelemetryPending || !overview.TelemetryUnavailable || !overview.AllocationFallbackUsed {
		t.Fatalf("fallback state = %#v", overview)
	}
	if overview.NamespaceMetrics.RunningPods != 1 ||
		overview.NamespaceMetrics.CPURequestsUsed != .25 ||
		overview.NamespaceMetrics.MemoryRequestsUsed != 512*1024*1024 {
		t.Fatalf("fallback metrics = %#v", overview.NamespaceMetrics)
	}
}

func TestMonitoringMergeKeepsFastSummaryWhenDetailsAreEmpty(t *testing.T) {
	summary := api.MonitoringOverview{
		Namespace: "workspace-dev",
		NamespaceMetrics: api.MonitoringNamespaceMetrics{
			TotalPods:   1,
			RunningPods: 1,
		},
	}
	details := api.MonitoringOverview{Namespace: "workspace-dev", GeneratedAt: "2026-06-11T00:00:00Z"}
	merged := mergeMonitoringOverview(summary, details)
	if merged.NamespaceMetrics.TotalPods != 1 || merged.GeneratedAt != details.GeneratedAt {
		t.Fatalf("merged overview = %#v", merged)
	}
}

func TestMonitoringPendingTelemetryRendersWaitingInsteadOfZero(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.monitoringOverview = api.MonitoringOverview{
		Namespace:        "workspace-dev",
		TelemetryPending: true,
		NamespaceMetrics: api.MonitoringNamespaceMetrics{TotalPods: 1, RunningPods: 1},
		Projects:         []api.MonitoringProjectMetrics{{Name: "orders-db", TotalPods: 1, RunningPods: 1}},
	}
	rendered := strings.Join(m.modernMonitoringDetail(72, 30), "\n")
	if !strings.Contains(rendered, "waiting") || strings.Contains(rendered, "CPU CORE") && strings.Contains(rendered, "CPU CORE            0%") {
		t.Fatalf("pending telemetry UI:\n%s", rendered)
	}
}

func TestMonitoringUnavailableRendersUnavailableInsteadOfWaiting(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.monitoringOverview = api.MonitoringOverview{
		Namespace:            "workspace-dev",
		TelemetryUnavailable: true,
		Projects:             []api.MonitoringProjectMetrics{{Name: "orders-db"}},
	}
	rendered := strings.Join(m.modernMonitoringDetail(72, 30), "\n")
	if !strings.Contains(rendered, "unavailable") || strings.Contains(rendered, "still catching up") {
		t.Fatalf("unavailable telemetry UI:\n%s", rendered)
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

func TestSplitPlainWidthPreservesLongConnectionURL(t *testing.T) {
	connectionURL := "jdbc:postgresql://mama.autonomous-istad.com:15432/mama?sslmode=require"
	lines := splitPlainWidth(connectionURL, 20)
	if len(lines) < 2 {
		t.Fatalf("expected wrapped URL, got %#v", lines)
	}
	if got := strings.Join(lines, ""); got != connectionURL {
		t.Fatalf("wrapped URL lost content: got %q want %q", got, connectionURL)
	}
	for _, line := range lines {
		if len([]rune(line)) > 20 {
			t.Fatalf("wrapped line is too wide: %q", line)
		}
	}
}

func TestMonolithicDeployResultOpensJenkinsLogs(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.monolithFormOpen = true
	m.monolithForm.projectName = "web"

	next, cmd := m.Update(monolithicDeployResultMsg{deployment: api.MonolithicDeploymentRecord{
		ProjectID:      "project-1",
		Name:           "web",
		Status:         "CREATED",
		DeployURL:      "https://web.autonomous-istad.com",
		QueueItemID:    42,
		JenkinsJobName: "deploy-monolith",
	}})
	m = next.(model)
	if cmd == nil || !m.deployLogOpen || m.jenkinsLogQueue != 42 || m.jenkinsLogJob != "deploy-monolith" {
		t.Fatalf("expected Jenkins log viewer, model=%#v cmd=%v", m, cmd)
	}
	if m.deployLog.DeploymentMode != "monolith" || m.deployLog.Status != "QUEUED" || m.deployLog.DeployURL != "https://web.autonomous-istad.com" {
		t.Fatalf("deployment log = %#v", m.deployLog)
	}

	next, cmd = m.Update(jenkinsDeploymentStreamMsg{
		queueID: 42,
		chunk: api.JenkinsLogStreamChunk{
			Lines:     []string{"Build complete"},
			Status:    "DEPLOYED",
			Message:   "Jenkins build completed: SUCCESS",
			Completed: true,
		},
	})
	m = next.(model)
	if cmd != nil || m.deployLog.Status != "DEPLOYED" || !strings.Contains(m.deployLog.StatusLog, "Build complete") {
		t.Fatalf("completed Jenkins log = %#v cmd=%v", m, cmd)
	}
	urlLines := applicationDeploymentURLLines(m.deployLog, 24, 0)
	if len(urlLines) < 2 || !strings.Contains(urlLines[0], "https://web") || !strings.Contains(urlLines[len(urlLines)-1], "com") {
		t.Fatalf("expected full wrapped deployment URL, got %#v", urlLines)
	}
}

func TestMicroserviceDeployResultOpensJenkinsLogs(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.monolithFormOpen = true
	m.monolithForm.mode = "microservices"
	m.monolithForm.projectName = "commerce"

	next, cmd := m.Update(microserviceDeployResultMsg{deployment: api.MonolithicDeploymentRecord{
		ProjectID:      "project-2",
		Name:           "commerce",
		Status:         "CREATED",
		DeployURL:      "https://commerce.autonomous-istad.com",
		QueueItemID:    43,
		JenkinsJobName: "deploy-microservices",
	}})
	m = next.(model)
	if cmd == nil || !m.deployLogOpen || m.jenkinsLogQueue != 43 || m.jenkinsLogJob != "deploy-microservices" {
		t.Fatalf("expected Jenkins log viewer, model=%#v cmd=%v", m, cmd)
	}
	if m.deployLog.DeploymentMode != "microservices" || m.deployLog.Status != "QUEUED" || m.deployLog.DeployURL != "https://commerce.autonomous-istad.com" {
		t.Fatalf("deployment log = %#v", m.deployLog)
	}

	next, cmd = m.Update(jenkinsDeploymentStreamMsg{
		queueID: 43,
		chunk: api.JenkinsLogStreamChunk{
			Lines:     []string{"Services deployed"},
			Status:    "DEPLOYED",
			Message:   "Jenkins build completed: SUCCESS",
			Completed: true,
		},
	})
	m = next.(model)
	if cmd != nil || m.deployLog.Status != "DEPLOYED" || !strings.Contains(m.deployLog.StatusLog, "Services deployed") {
		t.Fatalf("completed Jenkins log = %#v cmd=%v", m, cmd)
	}
	if m.message != "Microservices deployment finished: commerce" {
		t.Fatalf("message = %q", m.message)
	}
	urlLines := applicationDeploymentURLLines(m.deployLog, 30, 0)
	if len(urlLines) < 2 || !strings.Contains(urlLines[0], "https://commerce.auton") || !strings.Contains(urlLines[len(urlLines)-1], "omous-istad.com") {
		t.Fatalf("expected full wrapped deployment URL, got %#v", urlLines)
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

func TestNewApplicationDeployFormsDoNotDefaultNamesOrGitRemote(t *testing.T) {
	form := newMonolithicDeployForm()
	if form.projectName != "" || form.serviceName != "" {
		t.Fatalf("monolithic names should start empty, got project=%q service=%q", form.projectName, form.serviceName)
	}
	if form.repoURL != "" {
		t.Fatalf("repoURL should start empty, got %q", form.repoURL)
	}
	if form.repoFullName != "" {
		t.Fatalf("repoFullName should start empty, got %q", form.repoFullName)
	}

	form = newMicroservicesDeployForm()
	if form.projectName != "" || form.serviceName != "" {
		t.Fatalf("microservice names should start empty, got project=%q service=%q", form.projectName, form.serviceName)
	}
}

func TestMicroserviceSourceModeScanAndMergeFlow(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.monolithFormOpen = true
	m.monolithForm = newMicroservicesDeployForm()
	m.monolithForm.projectName = "commerce"
	m.monolithForm.focus = 0

	next, cmd := m.updateKey(specialKeyMsg(tea.KeyRight))
	m = next.(model)
	if cmd != nil || m.monolithForm.sourceMode() != "multi-repo" {
		t.Fatalf("expected multi-repo source mode, form=%#v cmd=%v", m.monolithForm, cmd)
	}

	next, cmd = m.Update(microserviceScanResultMsg{result: api.MicroserviceDetectionResult{
		Repository: api.DetectedMicroserviceRepository{FullName: "team/api", DefaultBranch: "main"},
		Services: []api.DetectedMicroserviceService{{
			Name:          "api",
			RepoURL:       "https://github.com/team/api",
			RepoFullName:  "team/api",
			Branch:        "main",
			AppPort:       8080,
			ServiceType:   "backend",
			ExposePublic:  true,
			PrimaryPublic: true,
		}},
	}})
	m = next.(model)
	if cmd != nil || len(m.monolithForm.detectedServices) != 1 || m.monolithForm.repoURL != "" {
		t.Fatalf("first scan form=%#v cmd=%v", m.monolithForm, cmd)
	}

	next, cmd = m.Update(microserviceScanResultMsg{result: api.MicroserviceDetectionResult{
		Repository: api.DetectedMicroserviceRepository{FullName: "team/web", DefaultBranch: "develop"},
		Services: []api.DetectedMicroserviceService{{
			Name:          "web",
			RepoURL:       "https://github.com/team/web",
			RepoFullName:  "team/web",
			Branch:        "develop",
			AppPort:       3000,
			ServiceType:   "frontend",
			ExposePublic:  true,
			PrimaryPublic: true,
		}},
	}})
	m = next.(model)
	if cmd != nil || len(m.monolithForm.detectedServices) != 2 || len(m.monolithForm.scannedRepositories) != 2 {
		t.Fatalf("merged scan form=%#v cmd=%v", m.monolithForm, cmd)
	}
	input, err := m.monolithForm.microserviceInput()
	if err != nil {
		t.Fatal(err)
	}
	primaryCount := 0
	for _, service := range input.Services {
		if service.PrimaryPublic {
			primaryCount++
		}
	}
	if primaryCount != 1 {
		t.Fatalf("expected one primary public service, input=%#v", input)
	}
}

func TestMicroserviceDetectedServicesStayVisibleAfterScan(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.monolithFormOpen = true
	m.monolithForm = newMicroservicesDeployForm()

	next, _ := m.Update(microserviceScanResultMsg{result: api.MicroserviceDetectionResult{
		Repository: api.DetectedMicroserviceRepository{FullName: "team/platform", DefaultBranch: "main"},
		Services: []api.DetectedMicroserviceService{
			{Name: "config-server", RepoURL: "https://github.com/team/platform", AppPort: 8089},
			{Name: "orders-service", RepoURL: "https://github.com/team/platform", AppPort: 8080},
		},
	}})
	m = next.(model)
	lines := m.renderDashboardMonolithicDeployForm(68, 20)
	rendered := strings.Join(lines, "\n")
	if !strings.Contains(rendered, "2 services detected") ||
		!strings.Contains(rendered, "Detected services  2") ||
		!strings.Contains(rendered, "config-server") {
		t.Fatalf("detected services not visible after scan:\n%s", rendered)
	}
	if m.monolithForm.focus != 3 {
		t.Fatalf("focus = %d, want scan result focus", m.monolithForm.focus)
	}
}

func TestMicroserviceScanFailureIsVisibleNearDetectedServices(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.monolithFormOpen = true
	m.monolithForm = newMicroservicesDeployForm()

	next, _ := m.Update(microserviceScanResultMsg{err: fmt.Errorf("public GitHub repository is unavailable")})
	m = next.(model)
	rendered := strings.Join(m.renderDashboardMonolithicDeployForm(80, 24), "\n")
	if !strings.Contains(rendered, "Scan failed: public GitHub repository is unavailable") {
		t.Fatalf("scan error not visible near detected services:\n%s", rendered)
	}
}

func TestMicroserviceDeployRequiresRepositoryScan(t *testing.T) {
	form := newMicroservicesDeployForm()
	form.projectName = "commerce"
	if _, err := form.microserviceInput(); err == nil || !strings.Contains(err.Error(), "Scan at least one repository") {
		t.Fatalf("err = %v", err)
	}
}

func TestMicroserviceDeployRejectsDuplicateDetectedServiceNames(t *testing.T) {
	form := newMicroservicesDeployForm()
	form.projectName = "commerce"
	form.detectedServices = []api.CreateMicroserviceServiceInput{
		{Name: "api", RepoURL: "https://github.com/team/one", RepoFullName: "team/one"},
		{Name: "API", RepoURL: "https://github.com/team/two", RepoFullName: "team/two"},
	}
	if _, err := form.microserviceInput(); err == nil || !strings.Contains(err.Error(), "duplicate name") {
		t.Fatalf("err = %v", err)
	}
}

func TestMicroserviceRelationshipEditorAddsAndRemovesDependency(t *testing.T) {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.monolithFormOpen = true
	m.monolithForm = newMicroservicesDeployForm()
	m.monolithForm.detectedServices = []api.CreateMicroserviceServiceInput{
		{Name: "orders-service", ServiceType: "backend"},
		{Name: "config-server", ServiceType: "backend"},
	}
	m.monolithForm.focus = 5

	next, cmd := m.updateMonolithicDeployForm(keyMsg("enter"))
	m = next.(model)
	if cmd != nil || !m.monolithForm.relationshipOpen {
		t.Fatalf("relationship editor not opened: form=%#v cmd=%v", m.monolithForm, cmd)
	}
	m.monolithForm.relationshipFocus = 4
	next, cmd = m.updateMonolithicDeployForm(keyMsg("enter"))
	m = next.(model)
	if cmd != nil {
		t.Fatalf("unexpected command: %v", cmd)
	}
	service := m.monolithForm.detectedServices[0]
	if len(service.DependsOn) != 1 || service.DependsOn[0] != "config-server" {
		t.Fatalf("dependsOn = %#v", service.DependsOn)
	}
	if len(service.Relationships) == 0 || service.Relationships[0].Value != "config-server" {
		t.Fatalf("relationships = %#v", service.Relationships)
	}

	m.monolithForm.relationshipFocus = 6
	next, cmd = m.updateMonolithicDeployForm(keyMsg("enter"))
	m = next.(model)
	service = m.monolithForm.detectedServices[0]
	if cmd != nil || len(service.DependsOn) != 0 || len(service.Relationships) != 0 {
		t.Fatalf("relationship not removed: service=%#v cmd=%v", service, cmd)
	}
}

func TestMicroserviceInputKeepsManagedRelationships(t *testing.T) {
	form := newMicroservicesDeployForm()
	form.projectName = "commerce"
	form.detectedServices = []api.CreateMicroserviceServiceInput{
		{
			Name:          "orders-service",
			RepoURL:       "https://github.com/team/commerce",
			DependsOn:     []string{"eureka-server"},
			Relationships: []api.MicroserviceRelationInput{{Name: "EUREKA_URL", Value: "eureka-server"}},
		},
		{Name: "eureka-server", RepoURL: "https://github.com/team/commerce"},
	}

	input, err := form.microserviceInput()
	if err != nil {
		t.Fatal(err)
	}
	if len(input.Services[0].DependsOn) != 1 || len(input.Services[0].Relationships) != 1 {
		t.Fatalf("managed relationships missing from input: %#v", input.Services[0])
	}
}

func TestMicroserviceInputRejectsUnknownDependency(t *testing.T) {
	form := newMicroservicesDeployForm()
	form.projectName = "commerce"
	form.detectedServices = []api.CreateMicroserviceServiceInput{{
		Name:      "orders-service",
		RepoURL:   "https://github.com/team/commerce",
		DependsOn: []string{"missing-service"},
	}}

	if _, err := form.microserviceInput(); err == nil || !strings.Contains(err.Error(), "unknown service") {
		t.Fatalf("err = %v", err)
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
