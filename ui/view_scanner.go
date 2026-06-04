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
		modernTitleLine("Image Scanner", width),
		modernLine(modernImageScannerModeChips(m.scannerMode), width, colorBgMain),
		modernLine("", width, colorBgMain),
		modernRule(width),
	}
	if m.scannerLoading {
		lines = append(lines, modernMutedLine("Loading scanner workspace...", width))
		return modernCropLines(lines, width, height)
	}
	switch m.scannerMode {
	case 1:
		if len(m.scannerScans) == 0 {
			lines = append(lines, modernMutedLine("No scan history yet", width))
		}
		for index, scan := range m.scannerScans {
			lines = append(lines, modernImageScanHistoryRow(scan, width, index == m.scannerHistoryCursor))
		}
	default:
		if len(m.scannerImages) == 0 {
			lines = append(lines, modernMutedLine("No deployed images found", width))
		}
		for index, image := range m.scannerImages {
			lines = append(lines, modernImageScannerImageRow(image, width, index == m.scannerCursor))
		}
	}
	lines = append(lines,
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
		modernMutedLine("enter scans or opens selected history", width),
		modernMutedLine("left/right switches scan and history", width),
	)
	return modernCropLines(lines, width, height)
}

func (m model) modernImageScannerDetail(width, height int) []string {
	if m.scannerMode == 1 {
		if scan, ok := m.selectedScannerHistory(); ok {
			return m.modernImageScanResult(scan, width, height)
		}
	}
	if m.scannerMode == 0 && m.scannerActiveScan.ID != "" {
		return m.modernImageScanResult(m.scannerActiveScan, width, height)
	}
	image, ok := m.selectedScannerImage()
	if !ok {
		lines := []string{
			modernTitleLine("Image Scanner", width),
			modernMutedLine("Open this section to load deployed images.", width),
			modernLine("", width, colorBgMain),
			modernRule(width),
			modernLine("", width, colorBgMain),
			modernMutedLine("Press r to refresh images and scan history.", width),
		}
		return modernCropLines(lines, width, height)
	}
	stats := imageScannerStats(m.scannerImages)
	lines := []string{
		modernTitleLine(imageScannerImageLabel(image), width),
		modernLine(lipgloss.NewStyle().Background(lipgloss.Color(colorBgMain)).Foreground(lipgloss.Color(colorPrimary)).Bold(true).Render("● harbor image"), width, colorBgMain),
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
		modernScannerMetricRow(stats, width),
		modernLine("", width, colorBgMain),
		modernFieldLine("Repository", image.Repository, width),
		modernFieldLine("Image", imageScannerImageLabel(image), width),
		modernFieldLine("Runtime", firstNonEmpty(image.Distro, "linux")+" · "+firstNonEmpty(image.Architecture, "amd64"), width),
		modernFieldLine("Size", image.SizeLabel, width),
		modernFieldLine("Last scan", imageLastScanLabel(image), width),
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
		modernMutedLine("Press enter to scan this image.", width),
	}
	return modernCropLines(lines, width, height)
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
		modernLine(lipgloss.NewStyle().Background(lipgloss.Color(colorBgMain)).Foreground(lipgloss.Color(statusColor(status))).Bold(true).Render("● "+strings.ToLower(status)), width, colorBgMain),
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
		modernScanSeveritySummary(counts, len(scan.Vulnerabilities), width),
		modernLine("", width, colorBgMain),
		modernFieldLine("Reference", firstNonEmpty(scan.FullReference, "n/a"), width),
		modernFieldLine("Scanner", firstNonEmpty(scan.ScannerName, "Trivy"), width),
		modernFieldLine("Progress", fmt.Sprintf("%d%%", progress), width),
	}
	if scan.StatusMessage != "" {
		lines = append(lines, modernFieldLine("Message", scan.StatusMessage, width))
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
	if len(scan.Vulnerabilities) == 0 {
		lines = append(lines, modernMutedLine("No vulnerabilities returned yet.", width))
	}
	return modernCropLines(lines, width, height)
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

func modernImageScannerModeChips(mode int) string {
	labels := []string{"Scan", "History"}
	parts := make([]string, 0, len(labels))
	for index, label := range labels {
		parts = append(parts, modernScannerChip(label, index == mode))
	}
	return strings.Join(parts, mainBodyStyle(colorBgMain).Render("  "))
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

func modernScannerChip(label string, selected bool) string {
	bg := colorBgPill
	fg := colorText
	prefix := "  "
	if selected {
		bg = colorPrimary
		fg = colorTitle
		prefix = "> "
	}
	return lipgloss.NewStyle().
		Background(lipgloss.Color(bg)).
		Foreground(lipgloss.Color(fg)).
		Bold(selected).
		Padding(0, 2).
		Render(prefix + label)
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
		return "#fb7185"
	case "HIGH":
		return "#f97316"
	case "MEDIUM":
		return "#facc15"
	case "LOW":
		return "#4ade80"
	default:
		return colorMuted
	}
}
