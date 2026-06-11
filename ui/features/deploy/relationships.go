package deploy

import (
	"fmt"
	"strings"

	"github.com/PorKeat/a8s-tui/api"
)

type RelationshipOption struct {
	Value       string
	Label       string
	Description string
}

var RelationshipOptions = []RelationshipOption{
	{
		Value:       "depends_on",
		Label:       "Depends on",
		Description: "Start this service after the target is available. No runtime URL variable is generated.",
	},
	{
		Value:       "http_upstream",
		Label:       "HTTP upstream",
		Description: "Generate runtime base URLs so this service can call the target over HTTP.",
	},
	{
		Value:       "service_registry",
		Label:       "Service registry",
		Description: "Generate Eureka service discovery and registration URLs.",
	},
	{
		Value:       "config_source",
		Label:       "Config source",
		Description: "Generate config server URLs used to load remote configuration.",
	},
	{
		Value:       "public_api",
		Label:       "Frontend API",
		Description: "Generate a public API URL for frontend runtimes such as Next.js.",
	},
	{
		Value:       "auth_provider",
		Label:       "Frontend auth",
		Description: "Generate a public auth URL for frontend sign-in and callback flows.",
	},
	{
		Value:       "custom",
		Label:       "Custom env var",
		Description: "Use a custom environment variable name for the generated target URL.",
	},
}

func RelationshipOptionIndex(value string) int {
	for index, option := range RelationshipOptions {
		if option.Value == value {
			return index
		}
	}
	return 0
}

func SuggestRelationshipKind(source, target api.CreateMicroserviceServiceInput) string {
	if isFrontendLike(source) {
		if isAuthLike(target) {
			return "auth_provider"
		}
		return "public_api"
	}
	if isRegistryLike(target) {
		return "service_registry"
	}
	if isConfigLike(target) {
		return "config_source"
	}
	return "http_upstream"
}

func ApplyMicroserviceRelationship(
	services []api.CreateMicroserviceServiceInput,
	sourceIndex, targetIndex int,
	kind, customName string,
) ([]api.CreateMicroserviceServiceInput, error) {
	if sourceIndex < 0 || sourceIndex >= len(services) || targetIndex < 0 || targetIndex >= len(services) {
		return services, fmt.Errorf("Select a valid source and target service")
	}
	if sourceIndex == targetIndex {
		return services, fmt.Errorf("A service cannot depend on itself")
	}

	source := services[sourceIndex]
	target := services[targetIndex]
	targetName := strings.TrimSpace(target.Name)
	if targetName == "" {
		return services, fmt.Errorf("Target service name is required")
	}

	envNames, err := relationshipEnvNames(kind, customName, source, target)
	if err != nil {
		return services, err
	}

	updated := append([]api.CreateMicroserviceServiceInput(nil), services...)
	source.DependsOn = appendUniqueFold(source.DependsOn, targetName)
	source.Relationships = removeRelationshipTarget(source.Relationships, targetName)
	for _, envName := range envNames {
		source.Relationships = append(source.Relationships, api.MicroserviceRelationInput{
			Name:  envName,
			Value: targetName,
		})
	}
	updated[sourceIndex] = source
	return updated, nil
}

func RemoveMicroserviceRelationship(
	services []api.CreateMicroserviceServiceInput,
	sourceIndex int,
	targetName string,
) ([]api.CreateMicroserviceServiceInput, error) {
	if sourceIndex < 0 || sourceIndex >= len(services) {
		return services, fmt.Errorf("Select a valid source service")
	}
	targetName = strings.TrimSpace(targetName)
	if targetName == "" {
		return services, fmt.Errorf("Select an existing relationship to remove")
	}

	updated := append([]api.CreateMicroserviceServiceInput(nil), services...)
	source := services[sourceIndex]
	dependsOn := make([]string, 0, len(source.DependsOn))
	for _, dependency := range source.DependsOn {
		if !strings.EqualFold(strings.TrimSpace(dependency), targetName) {
			dependsOn = append(dependsOn, dependency)
		}
	}
	source.DependsOn = dependsOn
	source.Relationships = removeRelationshipTarget(source.Relationships, targetName)
	updated[sourceIndex] = source
	return updated, nil
}

func relationshipEnvNames(
	kind, customName string,
	source, target api.CreateMicroserviceServiceInput,
) ([]string, error) {
	kind = strings.TrimSpace(kind)
	if kind == "custom" {
		customName = strings.TrimSpace(customName)
		if customName == "" {
			return nil, fmt.Errorf("Custom environment variable name is required")
		}
		return []string{customName}, nil
	}
	if kind == "" || kind == "depends_on" {
		return nil, nil
	}

	exact := make([]string, 0)
	for _, entry := range source.Env {
		name := normalizeEnvToken(entry.Name)
		value := strings.TrimSpace(entry.Value)
		switch {
		case kind == "service_registry" && (strings.HasPrefix(name, "EUREKA_") || strings.Contains(strings.ToLower(value), "eureka")):
			exact = appendUniqueFold(exact, name)
		case kind == "config_source" && isConfigPlaceholder(name, value):
			exact = appendUniqueFold(exact, name)
		case (kind == "http_upstream" || kind == "public_api" || kind == "auth_provider") &&
			isTargetURLPlaceholder(name, value, target):
			exact = appendUniqueFold(exact, name)
		}
	}

	aliases := map[string][]string{
		"http_upstream":    append([]string{"UPSTREAM_BASE_URL"}, targetSpecificAliases(target.Name)...),
		"service_registry": {"EUREKA_URL", "EUREKA_SERVER_URL", "EUREKA_CLIENT_SERVICEURL_DEFAULTZONE"},
		"config_source":    {"CONFIG_SERVER_URL", "SPRING_CONFIG_IMPORT"},
		"public_api":       {"NEXT_PUBLIC_API_BASE_URL"},
		"auth_provider":    {"NEXT_PUBLIC_AUTH_BASE_URL"},
	}
	for _, alias := range aliases[kind] {
		exact = appendUniqueFold(exact, alias)
	}
	if len(exact) == 0 {
		return nil, fmt.Errorf("Unknown relationship type %q", kind)
	}
	return exact, nil
}

func removeRelationshipTarget(values []api.MicroserviceRelationInput, targetName string) []api.MicroserviceRelationInput {
	filtered := make([]api.MicroserviceRelationInput, 0, len(values))
	for _, value := range values {
		if !strings.EqualFold(strings.TrimSpace(value.Value), targetName) {
			filtered = append(filtered, value)
		}
	}
	return filtered
}

func appendUniqueFold(values []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return values
	}
	for _, current := range values {
		if strings.EqualFold(strings.TrimSpace(current), value) {
			return values
		}
	}
	return append(values, value)
}

func normalizeEnvToken(value string) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	var out strings.Builder
	underscore := false
	for _, char := range value {
		if (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') {
			out.WriteRune(char)
			underscore = false
			continue
		}
		if out.Len() > 0 && !underscore {
			out.WriteByte('_')
			underscore = true
		}
	}
	return strings.Trim(out.String(), "_")
}

func targetSpecificAliases(name string) []string {
	token := normalizeEnvToken(name)
	if token == "" {
		return nil
	}
	aliases := []string{token + "_URL"}
	if !strings.HasSuffix(token, "_SERVICE") {
		aliases = append(aliases, token+"_SERVICE_URL")
	}
	return append(aliases, token+"_BASE_URL")
}

func isConfigPlaceholder(name, value string) bool {
	lowerValue := strings.ToLower(value)
	return name == "CONFIG_SERVER_URL" ||
		name == "SPRING_CONFIG_IMPORT" ||
		strings.Contains(name, "CONFIG_SERVER") ||
		strings.Contains(lowerValue, "configserver:") ||
		strings.Contains(lowerValue, "config-server")
}

func isTargetURLPlaceholder(name, value string, target api.CreateMicroserviceServiceInput) bool {
	for _, alias := range targetSpecificAliases(target.Name) {
		if name == alias {
			return true
		}
	}
	if name == "UPSTREAM_BASE_URL" {
		return true
	}
	if !(strings.HasSuffix(name, "_URL") || strings.HasSuffix(name, "_BASE_URL") || strings.HasSuffix(name, "_URI")) {
		return false
	}
	targetName := strings.ToLower(strings.TrimSpace(target.Name))
	lowerValue := strings.ToLower(value)
	return targetName != "" && (strings.Contains(lowerValue, targetName) ||
		strings.Contains(lowerValue, strings.ReplaceAll(targetName, "-", "_")) ||
		strings.Contains(lowerValue, strings.ReplaceAll(targetName, "-", "")))
}

func isFrontendLike(service api.CreateMicroserviceServiceInput) bool {
	return strings.EqualFold(strings.TrimSpace(service.ServiceType), "frontend") ||
		strings.Contains(strings.ToLower(service.Framework), "next")
}

func isAuthLike(service api.CreateMicroserviceServiceInput) bool {
	name := strings.ToLower(strings.TrimSpace(service.Name))
	return strings.EqualFold(strings.TrimSpace(service.ServiceType), "auth") ||
		strings.Contains(name, "auth") ||
		strings.Contains(name, "keycloak")
}

func isConfigLike(service api.CreateMicroserviceServiceInput) bool {
	return strings.Contains(strings.ToLower(strings.TrimSpace(service.Name)), "config")
}

func isRegistryLike(service api.CreateMicroserviceServiceInput) bool {
	name := strings.ToLower(strings.TrimSpace(service.Name))
	return strings.EqualFold(strings.TrimSpace(service.ServiceType), "registry") ||
		strings.Contains(name, "eureka") ||
		strings.Contains(name, "registry")
}
