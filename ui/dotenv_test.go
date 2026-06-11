package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PorKeat/a8s-tui/api"
	"github.com/PorKeat/a8s-tui/config"
)

func TestParseDotEnvMatchesDeploymentRules(t *testing.T) {
	envVars, err := parseDotEnv(`
# comment
export API_URL="http://api:8080"
PASSWORD='first'
EMPTY=
PASSWORD=latest
`)
	if err != nil {
		t.Fatal(err)
	}
	if len(envVars) != 3 {
		t.Fatalf("env var count = %d, want 3: %#v", len(envVars), envVars)
	}
	if envVars[0] != (api.MicroserviceEnvInput{Name: "API_URL", Value: "http://api:8080"}) {
		t.Fatalf("API_URL = %#v", envVars[0])
	}
	if envVars[1] != (api.MicroserviceEnvInput{Name: "PASSWORD", Value: "latest", Secret: true}) {
		t.Fatalf("PASSWORD = %#v", envVars[1])
	}
	if envVars[2].Name != "EMPTY" || envVars[2].Value != "" {
		t.Fatalf("EMPTY = %#v", envVars[2])
	}
}

func TestParseDotEnvRejectsInvalidName(t *testing.T) {
	_, err := parseDotEnv("NOT-VALID=value")
	if err == nil || !strings.Contains(err.Error(), "invalid environment variable name") {
		t.Fatalf("err = %v", err)
	}
}

func TestMicroserviceEnvironmentFileImportTargetsSelectedService(t *testing.T) {
	path := filepath.Join(t.TempDir(), "orders.env")
	if err := os.WriteFile(path, []byte("API_URL=http://api\nPASSWORD=secret\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	m := initialModelForEnvironmentImport()
	m.monolithForm.envTarget = 1
	next, cmd := m.Update(environmentFileChoiceMsg{path: path})
	m = next.(model)

	if cmd != nil {
		t.Fatalf("unexpected command: %v", cmd)
	}
	if len(m.monolithForm.detectedServices[0].Env) != 0 {
		t.Fatalf("unexpected env on first service: %#v", m.monolithForm.detectedServices[0].Env)
	}
	envVars := m.monolithForm.detectedServices[1].Env
	if len(envVars) != 2 || envVars[1].Name != "PASSWORD" || !envVars[1].Secret {
		t.Fatalf("imported env = %#v", envVars)
	}
	if !strings.Contains(m.message, "Imported 2 environment variable(s) into orders") {
		t.Fatalf("message = %q", m.message)
	}
	input, err := m.monolithForm.microserviceInput()
	if err != nil {
		t.Fatal(err)
	}
	if len(input.Services[1].Env) != 2 {
		t.Fatalf("deployment env = %#v", input.Services[1].Env)
	}
}

func TestMicroserviceEnvironmentFieldOpensBrowserAndDoesNotAcceptText(t *testing.T) {
	m := initialModelForEnvironmentImport()
	m.monolithForm.envFilePath = ""

	next, _ := m.updateMonolithicDeployForm(keyMsg("j"))
	m = next.(model)
	if m.monolithForm.focus != 7 || m.monolithForm.envFilePath != "" {
		t.Fatalf("j should navigate from browser field: path=%q focus=%d", m.monolithForm.envFilePath, m.monolithForm.focus)
	}
	m.monolithForm.focus = 6
	next, cmd := m.updateMonolithicDeployForm(keyMsg("enter"))
	m = next.(model)

	if cmd == nil || !strings.Contains(m.message, "Opening environment file browser") {
		t.Fatalf("expected browser command: message=%q cmd=%v", m.message, cmd)
	}
}

func TestMicroserviceEnvironmentBrowserCancelKeepsExistingEnv(t *testing.T) {
	m := initialModelForEnvironmentImport()
	m.monolithForm.detectedServices[0].Env = []api.MicroserviceEnvInput{{Name: "KEEP", Value: "value"}}

	next, cmd := m.Update(environmentFileChoiceMsg{err: errNativeSaveDialogCancelled})
	m = next.(model)

	if cmd != nil || len(m.monolithForm.detectedServices[0].Env) != 1 || m.message != "Environment import cancelled" {
		t.Fatalf("cancel changed form: %#v cmd=%v", m.monolithForm, cmd)
	}
}

func initialModelForEnvironmentImport() model {
	m := initialModel(config.AppConfig{BackendBaseURL: "http://backend"}, nil)
	m.state = stateReady
	m.monolithFormOpen = true
	m.monolithForm = newMicroservicesDeployForm()
	m.monolithForm.projectName = "commerce"
	m.monolithForm.focus = 6
	m.monolithForm.detectedServices = []api.CreateMicroserviceServiceInput{
		{Name: "api", RepoURL: "https://github.com/team/commerce"},
		{Name: "orders", RepoURL: "https://github.com/team/commerce"},
	}
	return m
}
