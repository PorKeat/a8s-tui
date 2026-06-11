package deploy

import (
	"slices"
	"testing"

	"github.com/PorKeat/a8s-tui/api"
)

func TestApplyMicroserviceRelationshipUsesTargetName(t *testing.T) {
	services := []api.CreateMicroserviceServiceInput{
		{Name: "orders-service", Env: []api.MicroserviceEnvInput{{Name: "EUREKA_URL", Value: "http://localhost:8761/eureka/"}}},
		{Name: "eureka-server", ServiceType: "registry"},
	}

	updated, err := ApplyMicroserviceRelationship(services, 0, 1, "service_registry", "")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(updated[0].DependsOn, []string{"eureka-server"}) {
		t.Fatalf("dependsOn = %#v", updated[0].DependsOn)
	}
	for _, name := range []string{"EUREKA_URL", "EUREKA_SERVER_URL", "EUREKA_CLIENT_SERVICEURL_DEFAULTZONE"} {
		if !hasRelationship(updated[0].Relationships, name, "eureka-server") {
			t.Fatalf("missing %s relationship: %#v", name, updated[0].Relationships)
		}
	}
}

func TestDependsOnRelationshipDoesNotGenerateEnvironmentVariable(t *testing.T) {
	services := []api.CreateMicroserviceServiceInput{{Name: "api"}, {Name: "database"}}

	updated, err := ApplyMicroserviceRelationship(services, 0, 1, "depends_on", "")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(updated[0].DependsOn, []string{"database"}) {
		t.Fatalf("dependsOn = %#v", updated[0].DependsOn)
	}
	if len(updated[0].Relationships) != 0 {
		t.Fatalf("relationships = %#v", updated[0].Relationships)
	}
}

func TestApplyRelationshipUpdatesExistingTargetWithoutDuplicates(t *testing.T) {
	services := []api.CreateMicroserviceServiceInput{
		{
			Name:          "api",
			DependsOn:     []string{"config-server"},
			Relationships: []api.MicroserviceRelationInput{{Name: "OLD_CONFIG_URL", Value: "config-server"}},
		},
		{Name: "config-server"},
	}

	updated, err := ApplyMicroserviceRelationship(services, 0, 1, "config_source", "")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(updated[0].DependsOn, []string{"config-server"}) {
		t.Fatalf("dependsOn = %#v", updated[0].DependsOn)
	}
	if hasRelationship(updated[0].Relationships, "OLD_CONFIG_URL", "config-server") {
		t.Fatalf("old relationship was not replaced: %#v", updated[0].Relationships)
	}
	if !hasRelationship(updated[0].Relationships, "CONFIG_SERVER_URL", "config-server") {
		t.Fatalf("config relationship missing: %#v", updated[0].Relationships)
	}
}

func TestRemoveMicroserviceRelationship(t *testing.T) {
	services := []api.CreateMicroserviceServiceInput{
		{
			Name:      "api",
			DependsOn: []string{"config-server", "eureka-server"},
			Relationships: []api.MicroserviceRelationInput{
				{Name: "CONFIG_SERVER_URL", Value: "config-server"},
				{Name: "EUREKA_URL", Value: "eureka-server"},
			},
		},
	}

	updated, err := RemoveMicroserviceRelationship(services, 0, "config-server")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(updated[0].DependsOn, []string{"eureka-server"}) {
		t.Fatalf("dependsOn = %#v", updated[0].DependsOn)
	}
	if len(updated[0].Relationships) != 1 || updated[0].Relationships[0].Name != "EUREKA_URL" {
		t.Fatalf("relationships = %#v", updated[0].Relationships)
	}
}

func TestSuggestRelationshipKindMatchesFrontendRoles(t *testing.T) {
	tests := []struct {
		source api.CreateMicroserviceServiceInput
		target api.CreateMicroserviceServiceInput
		want   string
	}{
		{source: api.CreateMicroserviceServiceInput{Name: "web", Framework: "Next.js"}, target: api.CreateMicroserviceServiceInput{Name: "auth-service"}, want: "auth_provider"},
		{source: api.CreateMicroserviceServiceInput{Name: "orders"}, target: api.CreateMicroserviceServiceInput{Name: "eureka-server"}, want: "service_registry"},
		{source: api.CreateMicroserviceServiceInput{Name: "orders"}, target: api.CreateMicroserviceServiceInput{Name: "config-server"}, want: "config_source"},
		{source: api.CreateMicroserviceServiceInput{Name: "orders"}, target: api.CreateMicroserviceServiceInput{Name: "inventory"}, want: "http_upstream"},
	}
	for _, test := range tests {
		if got := SuggestRelationshipKind(test.source, test.target); got != test.want {
			t.Fatalf("SuggestRelationshipKind(%q, %q) = %q, want %q", test.source.Name, test.target.Name, got, test.want)
		}
	}
}

func hasRelationship(values []api.MicroserviceRelationInput, name, value string) bool {
	for _, relationship := range values {
		if relationship.Name == name && relationship.Value == value {
			return true
		}
	}
	return false
}
