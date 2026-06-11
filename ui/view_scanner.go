package ui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/PorKeat/a8s-tui/api"

	"charm.land/lipgloss/v2"
)

func (m model) modernImageScannerList(width, height int) []string {
	lines := []string{
		modernTableHeader("SOURCE", "DATA", width),
		modernRule(width),
	}
	sources := []struct {
		label string
		state string
	}{
		{"Harbor", fmt.Sprintf("%d images", len(m.scannerImages))},
		{"External", "pull & scan"},
		{"Git", "build & scan"},
		{"History", fmt.Sprintf("%d scans", len(m.scannerScans))},
	}
	for index, source := range sources {
		lines = append(lines, modernImageScannerSourceRow(source.label, source.state, width, index == m.scannerMode)...)
	}
	lines = append(lines,
		modernLine("", width, colorBgMain),
		modernMutedLine("up/down selects source · enter opens", width),
		modernMutedLine("esc returns here from the right pane", width),
	)
	return modernCropLines(lines, width, height)
}

func (m model) modernImageScannerDetail(width, height int) []string {
	if m.scannerLoading && m.scannerActiveScan.ID == "" {
		return modernCropLines([]string{
			modernTitleLine("Starting scan", width),
			modernLine(mainPrimaryStyle(colorBgMain).Render(m.spinner.View()+" Sending "+scannerSourceLabel(m.scannerMode)+" request"), width, colorBgMain),
			modernLine("", width, colorBgMain),
			modernRule(width),
			modernLine("", width, colorBgMain),
			modernMutedLine("A8S is validating the source and starting the Jenkins Trivy job.", width),
			modernMutedLine("The status and findings will appear here automatically.", width),
		}, width, height)
	}
	if m.scannerActiveScan.ID != "" {
		return m.modernImageScanResult(m.scannerActiveScan, width, height)
	}
	if m.scannerMode == 1 {
		return m.modernExternalScannerWorkspace(width, height)
	}
	if m.scannerMode == 2 {
		return m.modernGitScannerWorkspace(width, height)
	}
	if m.scannerMode == 3 {
		return m.modernScannerHistoryWorkspace(width, height)
	}
	return m.modernHarborScannerWorkspace(width, height)
}

func (m model) modernHarborScannerWorkspace(width, height int) []string {
	stats := imageScannerStats(m.scannerImages)
	lines := []string{
		modernTitleLine("Harbor images", width),
		modernLine(lipgloss.NewStyle().Background(lipgloss.Color(colorBgMain)).Foreground(lipgloss.Color(colorPrimary)).Bold(true).Render("● harbor image"), width, colorBgMain),
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
		modernScannerMetricRow(stats, width),
		modernLine("", width, colorBgMain),
		modernHeading("Deployed images", width),
	}
	if len(m.scannerImages) == 0 {
		lines = append(lines, modernMutedLine("No deployed images found. Press r to refresh.", width))
	} else {
		start, end := scannerWindow(len(m.scannerImages), m.scannerCursor, 6)
		for index := start; index < end; index++ {
			lines = append(lines, modernImageScannerImageRow(m.scannerImages[index], width, index == m.scannerCursor))
		}
	}
	lines = append(lines,
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
	)
	if image, ok := m.selectedScannerImage(); ok {
		lines = append(lines,
			modernHeading("Selected image", width),
			modernFieldLine("Repository", image.Repository, width),
			modernFieldLine("Image", imageScannerImageLabel(image), width),
			modernFieldLine("Runtime", firstNonEmpty(image.Distro, "linux")+" · "+firstNonEmpty(image.Architecture, "amd64"), width),
			modernFieldLine("Size", image.SizeLabel, width),
			modernFieldLine("Last scan", imageLastScanLabel(image), width),
			modernLine("", width, colorBgMain),
			modernMutedLine("Press enter to scan this image.", width),
		)
	}
	return modernCropLines(lines, width, height)
}

func (m model) modernExternalScannerWorkspace(width, height int) []string {
	lines := []string{
		modernTitleLine("External registry", width),
		modernLine(mainPrimaryStyle(colorBgMain).Render("● pull and scan"), width, colorBgMain),
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
	}
	lines = append(lines, m.modernExternalScannerForm(width)...)
	lines = append(lines,
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
		modernHeading("Preview", width),
		modernFieldLine("Full image", firstNonEmpty(externalImageReference(m.scannerForm.externalRegistry, m.scannerForm.externalName, m.scannerForm.externalTag), "enter registry and image name"), width),
		modernFieldLine("Access", scannerAccessLabel(m.scannerForm.externalPrivate), width),
		modernLine("", width, colorBgMain),
		modernMutedLine("Use registry hosts such as docker.io or ghcr.io, not image web pages.", width),
	)
	return modernScannerWorkspaceLines(lines, width, height, m.scannerForm.focus)
}

func (m model) modernGitScannerWorkspace(width, height int) []string {
	lines := []string{
		modernTitleLine("Git repository", width),
		modernLine(mainPrimaryStyle(colorBgMain).Render("● build and scan"), width, colorBgMain),
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
	}
	lines = append(lines, m.modernGitScannerForm(width)...)
	lines = append(lines,
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
		modernHeading("Build preview", width),
		modernFieldLine("Repository", firstNonEmpty(m.scannerForm.gitRepository, "enter repository URL"), width),
		modernFieldLine("Build", "main · Dockerfile · context .", width),
		modernFieldLine("Target image", gitRepositoryImageName(m.scannerForm.gitRepository), width),
		modernFieldLine("Access", scannerAccessLabel(m.scannerForm.gitPrivate), width),
		modernLine("", width, colorBgMain),
		modernMutedLine("Clones, builds, and scans the resulting image.", width),
	)
	return modernScannerWorkspaceLines(lines, width, height, m.scannerForm.focus)
}

func (m model) modernScannerHistoryWorkspace(width, height int) []string {
	lines := []string{
		modernTitleLine("Scan history", width),
		modernLine(mainPrimaryStyle(colorBgMain).Render(fmt.Sprintf("● %d scans", len(m.scannerScans))), width, colorBgMain),
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
	}
	if len(m.scannerScans) == 0 {
		lines = append(lines, modernMutedLine("No scan history yet.", width))
		return modernCropLines(lines, width, height)
	}
	start, end := scannerWindow(len(m.scannerScans), m.scannerHistoryCursor, 7)
	for index := start; index < end; index++ {
		lines = append(lines, modernImageScanHistoryRow(m.scannerScans[index], width, index == m.scannerHistoryCursor))
	}
	lines = append(lines,
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
		modernHeading("Selected scan", width),
	)
	if scan, ok := m.selectedScannerHistory(); ok {
		counts := api.ImageScanSeverityCounts(scan.Vulnerabilities)
		lines = append(lines,
			modernFieldLine("Image", imageScanTitle(scan), width),
			modernFieldLine("Status", firstNonEmpty(scan.Status, "PENDING"), width),
			modernFieldLine("Findings", fmt.Sprintf("%d critical · %d high · %d total", counts["CRITICAL"], counts["HIGH"], len(scan.Vulnerabilities)), width),
			modernLine("", width, colorBgMain),
			modernMutedLine("Press enter to open the full report.", width),
		)
	}
	return modernCropLines(lines, width, height)
}

func modernScannerWorkspaceLines(lines []string, width, height, focus int) []string {
	const fixedHeaderLines = 5
	if len(lines) <= height || len(lines) <= fixedHeaderLines {
		return modernCropLines(lines, width, height)
	}
	header := append([]string(nil), lines[:fixedHeaderLines]...)
	body := lines[fixedHeaderLines:]
	available := max(height-fixedHeaderLines, 1)
	focusLine := 3 + max(focus, 0)*3
	offset := clamp(focusLine-available/2, 0, max(len(body)-available, 0))
	visible := append(header, body[offset:min(offset+available, len(body))]...)
	return modernCropLines(visible, width, height)
}

func (m model) modernExternalScannerForm(width int) []string {
	form := m.scannerForm
	lines := []string{
		modernHeading("External image", width),
		modernMutedLine("Enter a registry host and image path. Credentials stay in memory.", width),
		modernLine("", width, colorBgMain),
	}
	lines = append(lines, modernScannerFormField("Registry URL", form.externalRegistry, "docker.io (not hub.docker.com image page)", width, form.focus == 0, false)...)
	lines = append(lines, modernScannerFormField("Image name", form.externalName, "library/nginx", width, form.focus == 1, false)...)
	lines = append(lines, modernScannerFormField("Image tag", form.externalTag, "latest", width, form.focus == 2, false)...)
	lines = append(lines, modernScannerFormChoice("Private registry", form.externalPrivate, width, form.focus == 3)...)
	if form.externalPrivate {
		lines = append(lines, modernScannerFormField("Username", form.externalUsername, "registry user", width, form.focus == 4, false)...)
		lines = append(lines, modernScannerFormField("Password", form.externalPassword, "required", width, form.focus == 5, true)...)
	}
	lines = append(lines, modernLine("", width, colorBgMain))
	lines = append(lines, modernScannerFormAction("Pull & Scan", width, form.focus == m.imageScannerFormFieldCount()-1)...)
	return lines
}

func (m model) modernGitScannerForm(width int) []string {
	form := m.scannerForm
	lines := []string{
		modernHeading("Git build", width),
		modernMutedLine("Enter a repository URL. A8S builds main/Dockerfile, then scans it.", width),
		modernLine("", width, colorBgMain),
	}
	lines = append(lines, modernScannerFormField("Repository URL", form.gitRepository, "https://github.com/user/repository.git", width, form.focus == 0, false)...)
	lines = append(lines, modernScannerFormChoice("Private repository", form.gitPrivate, width, form.focus == 1)...)
	if form.gitPrivate {
		lines = append(lines, modernScannerFormField("Username", form.gitUsername, "Git user", width, form.focus == 2, false)...)
		lines = append(lines, modernScannerFormField("Token", form.gitPassword, "required", width, form.focus == 3, true)...)
	}
	lines = append(lines, modernLine("", width, colorBgMain))
	lines = append(lines, modernScannerFormAction("Build & Scan", width, form.focus == m.imageScannerFormFieldCount()-1)...)
	return lines
}

func modernScannerFormField(label, value, placeholder string, width int, active, secret bool) []string {
	rowBg := colorBgMain
	border := colorBorder
	labelStyle := mainMutedStyle(rowBg)
	valueStyle := mainBodyStyle(rowBg)
	prefixStyle := mainBodyStyle(rowBg)
	prefix := "  "
	if active {
		border = colorPrimary
		labelStyle = mainPrimaryStyle(rowBg)
		valueStyle = mainTitleStyle(rowBg)
		prefixStyle = mainPrimaryStyle(rowBg)
		prefix = "> "
	}
	display := strings.TrimSpace(value)
	if display == "" {
		display = placeholder
		valueStyle = mainMutedStyle(rowBg)
	} else if secret {
		display = strings.Repeat("•", len([]rune(display)))
	}
	boxWidth := max(width-4, 24)
	labelWidth := min(18, max(boxWidth/3, 12))
	valueWidth := max(boxWidth-labelWidth-8, 8)
	content := prefixStyle.Render(prefix) +
		labelStyle.Render(pad(truncatePlain(label, labelWidth), labelWidth)) +
		mainBodyStyle(rowBg).Render("  ") +
		valueStyle.Render(truncatePlain(display, valueWidth))
	box := lipgloss.NewStyle().
		Background(lipgloss.Color(rowBg)).
		Foreground(lipgloss.Color(colorText)).
		Width(boxWidth).
		Padding(0, 1).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(border)).
		BorderBackground(lipgloss.Color(rowBg)).
		Render(content)
	rendered := strings.Split(box, "\n")
	lines := make([]string, 0, len(rendered))
	for _, line := range rendered {
		lines = append(lines, modernLine("  "+line, width, colorBgMain))
	}
	return lines
}

func modernScannerFormChoice(label string, enabled bool, width int, active bool) []string {
	value := "public"
	if enabled {
		value = "private"
	}
	return modernScannerFormField(label, value, "", width, active, false)
}

func modernScannerFormAction(label string, width int, active bool) []string {
	rowBg := colorBgMain
	border := colorBorder
	style := mainBodyStyle(rowBg)
	if active {
		border = colorPrimary
		style = mainPrimaryStyle(rowBg)
	}
	box := lipgloss.NewStyle().
		Background(lipgloss.Color(rowBg)).
		Foreground(lipgloss.Color(colorText)).
		Padding(0, 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(border)).
		BorderBackground(lipgloss.Color(rowBg)).
		Render(style.Render(label))
	rendered := strings.Split(box, "\n")
	lines := make([]string, 0, len(rendered))
	for _, line := range rendered {
		lines = append(lines, modernLine("  "+line, width, colorBgMain))
	}
	return lines
}

func scannerAccessLabel(private bool) string {
	if private {
		return "private · credentials kept in memory"
	}
	return "public"
}

func modernImageScannerSourceRow(label, state string, width int, active bool) []string {
	rowBg := colorBgMain
	labelStyle := mainBodyStyle(rowBg)
	stateStyle := mainMutedStyle(rowBg)
	border := colorBorder
	prefix := "  "
	if active {
		border = colorPrimary
		labelStyle = mainTitleStyle(rowBg)
		stateStyle = mainPrimaryStyle(rowBg)
		prefix = "> "
	}
	boxWidth := max(width-4, 12)
	contentWidth := max(boxWidth-4, 8)
	state = truncatePlain(state, max(contentWidth-8, 4))
	labelWidth := max(contentWidth-visibleLen(state)-1, 4)
	name := labelStyle.Render(prefix + truncatePlain(label, labelWidth))
	meta := stateStyle.Render(truncatePlain(state, max(contentWidth-visibleLen(name)-1, 4)))
	line := name + spaces(max(contentWidth-visibleLen(name)-visibleLen(meta), 1)) + meta
	box := lipgloss.NewStyle().
		Background(lipgloss.Color(rowBg)).
		Foreground(lipgloss.Color(colorText)).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(border)).
		BorderBackground(lipgloss.Color(colorBgMain)).
		Padding(0, 1).
		Width(boxWidth).
		Render(line)
	rendered := strings.Split(box, "\n")
	lines := make([]string, 0, len(rendered))
	for _, line := range rendered {
		lines = append(lines, modernLine(line, width, colorBgMain))
	}
	return lines
}

func modernImageScannerImageRow(image api.ImageScannerImage, width int, active bool) string {
	rowBg := colorBgMain
	nameStyle := mainBodyStyle(rowBg)
	metaStyle := mainMutedStyle(rowBg)
	prefix := "  "
	if active {
		rowBg = colorBgActive
		nameStyle = mainTitleStyle(rowBg)
		metaStyle = mainPrimaryStyle(rowBg)
		prefix = "> "
	}
	vulns := fmt.Sprintf("%d vulns", image.VulnerabilityCount)
	nameWidth := max(width-visibleLen(vulns)-4, 8)
	name := nameStyle.Render(prefix + truncatePlain(imageScannerImageLabel(image), nameWidth))
	meta := metaStyle.Render(vulns)
	line := name + spaces(max(width-visibleLen(name)-visibleLen(meta), 1)) + meta
	return lipgloss.NewStyle().Background(lipgloss.Color(rowBg)).Render(pad(line, width))
}

func scannerWindow(total, cursor, limit int) (int, int) {
	if total <= 0 || limit <= 0 {
		return 0, 0
	}
	limit = min(limit, total)
	cursor = clamp(cursor, 0, total-1)
	start := clamp(cursor-limit/2, 0, total-limit)
	return start, start + limit
}

func modernImageScanHistoryRow(scan api.ImageScanJob, width int, active bool) string {
	rowBg := colorBgMain
	nameStyle := mainBodyStyle(rowBg)
	statusStyle := lipgloss.NewStyle().Background(lipgloss.Color(rowBg)).Foreground(lipgloss.Color(statusColor(scan.Status)))
	prefix := "  "
	if active {
		rowBg = colorBgActive
		nameStyle = mainTitleStyle(rowBg)
		statusStyle = statusStyle.Background(lipgloss.Color(rowBg)).Bold(true)
		prefix = "> "
	}
	status := strings.ToLower(firstNonEmpty(scan.Status, "pending"))
	nameWidth := max(width-visibleLen(status)-4, 8)
	name := nameStyle.Render(prefix + truncatePlain(imageScanTitle(scan), nameWidth))
	state := statusStyle.Render(status)
	line := name + spaces(max(width-visibleLen(name)-visibleLen(state), 1)) + state
	return lipgloss.NewStyle().Background(lipgloss.Color(rowBg)).Render(pad(line, width))
}

func (m model) modernImageScanResult(scan api.ImageScanJob, width, height int) []string {
	counts := api.ImageScanSeverityCounts(scan.Vulnerabilities)
	status := firstNonEmpty(scan.Status, "PENDING")
	progress := scan.Progress
	if progress == 0 && api.ImageScanTerminal(status) {
		progress = 100
	}
	lines := []string{
		modernTitleLine(imageScanTitle(scan), width),
		modernScanStatusBanner(scan, width),
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
		modernScanSeveritySummary(counts, len(scan.Vulnerabilities), width),
		modernLine("", width, colorBgMain),
		modernFieldLine("Scanner", firstNonEmpty(scan.ScannerName, "Trivy"), width),
		modernFieldLine("Progress", fmt.Sprintf("%d%%", progress), width),
	}
	lines = append(lines, modernWrappedFieldLines("Reference", firstNonEmpty(scan.FullReference, "n/a"), width)...)
	if scan.StatusMessage != "" {
		lines = append(lines, modernWrappedFieldLines("Message", scan.StatusMessage, width)...)
	}
	lines = append(lines,
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
	)
	reportLines := m.modernImageScanReportLines(scan, width, 7)
	if len(reportLines) > 0 {
		lines = append(lines, modernMutedLine("Report", width))
		lines = append(lines, reportLines...)
		lines = append(lines,
			modernLine("", width, colorBgMain),
			modernRule(width),
			modernLine("", width, colorBgMain),
		)
	}
	lines = append(lines,
		modernMutedLine("Findings", width),
	)
	for _, finding := range topImageScanFindings(scan.Vulnerabilities, 6) {
		lines = append(lines, modernImageScanFindingLine(finding, width))
	}
	if len(scan.Vulnerabilities) == 0 && api.ImageScanFailed(status) {
		lines = append(lines, modernLine(lipgloss.NewStyle().
			Background(lipgloss.Color(colorBgMain)).
			Foreground(lipgloss.Color(colorError)).
			Bold(true).
			Render("Scan stopped before Trivy returned findings."), width, colorBgMain))
	} else if len(scan.Vulnerabilities) == 0 {
		lines = append(lines, modernMutedLine("No vulnerabilities returned yet.", width))
	}
	lines = append(lines,
		modernLine("", width, colorBgMain),
		modernMutedLine("n new scan  ·  x force rescan  ·  right history", width),
	)
	return modernCropLines(lines, width, height)
}

func modernScanStatusBanner(scan api.ImageScanJob, width int) string {
	status := strings.ToUpper(firstNonEmpty(scan.Status, "PENDING"))
	message := strings.TrimSpace(scan.StatusMessage)
	if message == "" {
		switch {
		case api.ImageScanFailed(status):
			message = "Scan failed before findings were produced"
		case api.ImageScanTerminal(status):
			message = "Scan complete"
		default:
			message = "Scan is running"
		}
	}
	bg := colorBgPill
	if api.ImageScanFailed(status) {
		bg = colorBgDanger
	}
	statusStyle := lipgloss.NewStyle().Background(lipgloss.Color(bg)).Foreground(lipgloss.Color(statusColor(status))).Bold(true)
	messageStyle := lipgloss.NewStyle().Background(lipgloss.Color(bg)).Foreground(lipgloss.Color(colorText))
	content := statusStyle.Render("● "+strings.ToLower(status)) + messageStyle.Render("  "+truncatePlain(message, max(width-16, 12)))
	return lipgloss.NewStyle().Background(lipgloss.Color(bg)).Render(pad(content, width))
}

func modernImageScanFindingLine(finding api.ImageScanFinding, width int) string {
	severity := strings.ToUpper(firstNonEmpty(finding.Severity, "LOW"))
	sev := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgMain)).
		Foreground(lipgloss.Color(severityColor(severity))).
		Bold(true).
		Render(pad(severity, 8))
	body := mainBodyStyle(colorBgMain).Render(" " + truncatePlain(firstNonEmpty(finding.CVEID, "CVE"), 16))
	pkg := mainMutedStyle(colorBgMain).Render(" " + truncatePlain(finding.PackageName+" "+finding.PackageVersion, max(width-28, 8)))
	return modernLine(sev+body+pkg, width, colorBgMain)
}

type imageScannerSummary struct {
	total     int
	scanned   int
	unscanned int
}

func imageScannerStats(images []api.ImageScannerImage) imageScannerSummary {
	stats := imageScannerSummary{total: len(images)}
	for _, image := range images {
		if strings.TrimSpace(image.LastScannedAtISO) != "" || image.VulnerabilityCount > 0 {
			stats.scanned++
		}
	}
	stats.unscanned = max(stats.total-stats.scanned, 0)
	return stats
}

func modernScannerMetricRow(stats imageScannerSummary, width int) string {
	parts := fitStyledParts([]string{
		modernMiniMetric("Images", fmt.Sprintf("%d", stats.total)),
		modernMiniMetric("Scanned", fmt.Sprintf("%d", stats.scanned)),
		modernMiniMetric("New", fmt.Sprintf("%d", stats.unscanned)),
	}, width)
	return modernLine(strings.Join(parts, mainBodyStyle(colorBgMain).Render("  ")), width, colorBgMain)
}

func modernScanSeveritySummary(counts map[string]int, total int, width int) string {
	parts := fitStyledParts([]string{
		modernMiniMetric("Total", fmt.Sprintf("%d", total)),
		modernMiniMetric("Critical", fmt.Sprintf("%d", counts["CRITICAL"])),
		modernMiniMetric("High", fmt.Sprintf("%d", counts["HIGH"])),
		modernMiniMetric("Medium", fmt.Sprintf("%d", counts["MEDIUM"])),
		modernMiniMetric("Low", fmt.Sprintf("%d", counts["LOW"])),
	}, width)
	return modernLine(strings.Join(parts, mainBodyStyle(colorBgMain).Render("  ")), width, colorBgMain)
}

func modernMiniMetric(label, value string) string {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgPill)).
		Foreground(lipgloss.Color(colorText)).
		Padding(0, 1).
		Render(label + " " + value)
}

func fitStyledParts(parts []string, width int) []string {
	var out []string
	separatorWidth := 2
	for _, part := range parts {
		nextWidth := visibleLen(part)
		if len(out) > 0 {
			nextWidth += visibleLen(strings.Join(out, "")) + separatorWidth*len(out)
		}
		if nextWidth > width && len(out) > 0 {
			break
		}
		out = append(out, part)
	}
	return out
}

func (m model) modernImageScanReportLines(scan api.ImageScanJob, width, limit int) []string {
	if m.scannerReportLoading && (m.scannerReportScanID == "" || m.scannerReportScanID == scan.ID) {
		return []string{modernMutedLine("Loading Trivy JSON report...", width)}
	}
	if m.scannerReportScanID != scan.ID || strings.TrimSpace(m.scannerReport) == "" {
		if api.ImageScanTerminal(scan.Status) && !api.ImageScanFailed(scan.Status) {
			return []string{modernMutedLine("Report will appear after scan completion.", width)}
		}
		return nil
	}
	summary := imageScanReportSummary(m.scannerReport)
	if len(summary) == 0 {
		return []string{modernMutedLine("Report loaded, but no summary fields were detected.", width)}
	}
	if len(summary) > limit {
		summary = summary[:limit]
	}
	lines := make([]string, 0, len(summary)+1)
	for _, line := range summary {
		lines = append(lines, modernLine(mainBodyStyle(colorBgMain).Render(truncatePlain(line, width)), width, colorBgMain))
	}
	lines = append(lines, modernMutedLine(fmt.Sprintf("raw JSON: %d bytes", len(m.scannerReport)), width))
	return lines
}

func imageScanReportSummary(raw string) []string {
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return []string{"JSON report loaded"}
	}
	var lines []string
	if schema := readReportValue(payload["SchemaVersion"]); schema != "" {
		lines = append(lines, "Schema version  "+schema)
	}
	if artifact := readReportValue(payload["ArtifactName"]); artifact != "" {
		lines = append(lines, "Artifact        "+artifact)
	}
	if artifactType := readReportValue(payload["ArtifactType"]); artifactType != "" {
		lines = append(lines, "Artifact type   "+artifactType)
	}
	if createdAt := readReportValue(payload["CreatedAt"]); createdAt != "" {
		lines = append(lines, "Created         "+createdAt)
	}
	if results, ok := payload["Results"].([]any); ok {
		lines = append(lines, fmt.Sprintf("Targets         %d", len(results)))
		totalVulns := 0
		for _, item := range results {
			result, ok := item.(map[string]any)
			if !ok {
				continue
			}
			target := firstNonEmpty(readReportValue(result["Target"]), "target")
			className := readReportValue(result["Class"])
			typeName := readReportValue(result["Type"])
			vulns := 0
			if vulnerabilities, ok := result["Vulnerabilities"].([]any); ok {
				vulns = len(vulnerabilities)
			}
			totalVulns += vulns
			meta := strings.TrimSpace(strings.Join([]string{className, typeName}, " "))
			if meta != "" {
				lines = append(lines, fmt.Sprintf("%s  %s  %d vulns", target, meta, vulns))
			} else {
				lines = append(lines, fmt.Sprintf("%s  %d vulns", target, vulns))
			}
		}
		lines = append(lines, fmt.Sprintf("Total vulns     %d", totalVulns))
	}
	return lines
}

func readReportValue(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return fmt.Sprintf("%.0f", typed)
	case bool:
		if typed {
			return "true"
		}
		return "false"
	default:
		return ""
	}
}

func imageLastScanLabel(image api.ImageScannerImage) string {
	if strings.TrimSpace(image.LastScannedAtISO) != "" {
		return shortTime(image.LastScannedAtISO)
	}
	if image.VulnerabilityCount > 0 {
		return fmt.Sprintf("%d previous findings", image.VulnerabilityCount)
	}
	return "not scanned"
}

func topImageScanFindings(findings []api.ImageScanFinding, limit int) []api.ImageScanFinding {
	out := append([]api.ImageScanFinding(nil), findings...)
	sort.SliceStable(out, func(i, j int) bool {
		left := severityRank(out[i].Severity)
		right := severityRank(out[j].Severity)
		if left != right {
			return left < right
		}
		return out[i].CVSSScore > out[j].CVSSScore
	})
	if len(out) > limit {
		return out[:limit]
	}
	return out
}

func severityRank(severity string) int {
	switch strings.ToUpper(strings.TrimSpace(severity)) {
	case "CRITICAL":
		return 0
	case "HIGH":
		return 1
	case "MEDIUM":
		return 2
	case "LOW":
		return 3
	default:
		return 4
	}
}

func severityColor(severity string) string {
	switch strings.ToUpper(strings.TrimSpace(severity)) {
	case "CRITICAL":
		return colorError
	case "HIGH":
		return colorError
	case "MEDIUM":
		return colorWarning
	case "LOW":
		return colorSuccess
	default:
		return colorMuted
	}
}
