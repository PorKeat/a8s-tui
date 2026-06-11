package ui

import (
	"fmt"
	"math"
	"strings"

	"github.com/PorKeat/a8s-tui/api"

	"charm.land/lipgloss/v2"
)

func (m model) modernLogsList(width, height int) []string {
	lines := []string{
		modernTitleLine("Logs", width),
		modernLine(lipgloss.NewStyle().Background(lipgloss.Color(colorBgMain)).Foreground(lipgloss.Color(colorPrimary)).Bold(true).Render("● recent pod output"), width, colorBgMain),
		modernLine("", width, colorBgMain),
		modernRule(width),
	}
	if m.logsLoading {
		lines = append(lines, modernMutedLine("Loading Kubernetes pods and recent logs...", width))
		return modernCropLines(lines, width, height)
	}
	namespace := firstNonEmpty(m.resolvedLogsNamespace(), "namespace pending")
	lines = append(lines,
		modernLine("", width, colorBgMain),
		modernFieldLine("Namespace", namespace, width),
		modernFieldLine("Pods", fmt.Sprintf("%d", len(m.logsPods)), width),
		modernLine("", width, colorBgMain),
		modernTableHeader("POD", "PHASE", width),
		modernRule(width),
	)
	if len(m.logsPods) == 0 {
		lines = append(lines, modernMutedLine("No pods found yet. Deploy a project, then refresh.", width))
		return modernCropLines(lines, width, height)
	}
	rowLimit := max(height-len(lines)-4, 1)
	start := 0
	if m.logsCursor >= rowLimit {
		start = m.logsCursor - rowLimit + 1
	}
	for index := start; index < len(m.logsPods) && len(lines) < height-4; index++ {
		lines = append(lines, modernPodRow(m.logsPods[index], width, index == m.logsCursor))
	}
	lines = append(lines,
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
		modernMutedLine("enter reloads selected pod logs", width),
	)
	return modernCropLines(lines, width, height)
}

func (m model) modernLogsDetail(width, height int) []string {
	pod, ok := m.selectedLogPod()
	title := "Workspace Logs"
	if ok {
		title = pod.Name
	}
	lines := []string{
		modernTitleLine(title, width),
		modernLine(lipgloss.NewStyle().Background(lipgloss.Color(colorBgMain)).Foreground(lipgloss.Color(podPhaseColor(pod.Phase))).Bold(true).Render("● "+strings.ToLower(firstNonEmpty(pod.Phase, "waiting"))), width, colorBgMain),
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
	}
	if !ok {
		lines = append(lines,
			modernMutedLine("Open Logs to load pods for the workspace namespace.", width),
			modernLine("", width, colorBgMain),
			modernMutedLine("Press r to refresh once you have deployed workloads.", width),
		)
		return modernCropLines(lines, width, height)
	}
	lines = append(lines,
		modernFieldLine("Ready", fmt.Sprintf("%d/%d containers", pod.ReadyContainers, pod.TotalContainers), width),
		modernFieldLine("Restarts", fmt.Sprintf("%d", pod.RestartCount), width),
		modernFieldLine("Node", firstNonEmpty(pod.NodeName, "n/a"), width),
		modernFieldLine("IP", firstNonEmpty(pod.PodIP, "n/a"), width),
		modernFieldLine("Age", modernDuration(pod.AgeSeconds), width),
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
		modernMutedLine("Recent logs", width),
	)
	if len(m.logsLines) == 0 {
		lines = append(lines, modernMutedLine("No recent output returned for this pod.", width))
		return modernCropLines(lines, width, height)
	}
	available := max(height-len(lines), 1)
	start := max(len(m.logsLines)-available, 0)
	for _, line := range m.logsLines[start:] {
		lines = append(lines, modernLogLine(line, width))
	}
	return modernCropLines(lines, width, height)
}

func (m model) modernMonitoringList(width, height int) []string {
	metrics := m.monitoringOverview.NamespaceMetrics
	lines := []string{
		modernTitleLine("Monitoring", width),
		modernLine(lipgloss.NewStyle().Background(lipgloss.Color(colorBgMain)).Foreground(lipgloss.Color(colorPrimary)).Bold(true).Render("● namespace overview"), width, colorBgMain),
		modernLine("", width, colorBgMain),
		modernRule(width),
	}
	if m.monitoringLoading {
		lines = append(lines, modernMutedLine("Loading Prometheus metrics...", width))
		return modernCropLines(lines, width, height)
	}
	lines = append(lines,
		modernLine("", width, colorBgMain),
		modernFieldLine("Namespace", firstNonEmpty(m.monitoringOverview.Namespace, "n/a"), width),
		modernFieldLine("Pods", fmt.Sprintf("%.0f total / %.0f running", metrics.TotalPods, metrics.RunningPods), width),
		modernFieldLine("CPU", modernFormatCPU(metrics.CPUCores), width),
		modernFieldLine("Memory", modernFormatBytes(metrics.MemoryBytes), width),
		modernFieldLine("Restarts", fmt.Sprintf("%.0f last hour", metrics.RestartsLastHour), width),
		modernLine("", width, colorBgMain),
		modernTableHeader("PROJECT", "HEALTH", width),
		modernRule(width),
	)
	if m.monitoringOverview.TelemetryPending {
		lines = append(lines, modernMutedLine("Workloads found · waiting for Prometheus telemetry.", width))
	} else if m.monitoringOverview.AllocationFallbackUsed {
		lines = append(lines, modernMutedLine("Live usage unavailable · showing database readiness and allocation.", width))
	} else if m.monitoringOverview.TelemetryUnavailable {
		lines = append(lines, modernMutedLine("Telemetry unavailable · check the deployment cluster and Prometheus.", width))
	}
	if len(m.monitoringOverview.Projects) == 0 {
		lines = append(lines, modernMutedLine("No project metrics returned yet.", width))
		return modernCropLines(lines, width, height)
	}
	rowLimit := max(height-len(lines)-1, 1)
	start := 0
	if m.monitoringCursor >= rowLimit {
		start = m.monitoringCursor - rowLimit + 1
	}
	for index := start; index < len(m.monitoringOverview.Projects) && len(lines) < height; index++ {
		lines = append(lines, modernMonitoringProjectRow(m.monitoringOverview.Projects[index], width, index == m.monitoringCursor))
	}
	return modernCropLines(lines, width, height)
}

func (m model) modernMonitoringDetail(width, height int) []string {
	project, ok := m.selectedMonitoringProject()
	lines := []string{
		modernTitleLine("RESOURCE_MONITOR", width),
		modernLine(lipgloss.NewStyle().Background(lipgloss.Color(colorBgMain)).Foreground(lipgloss.Color(colorInfo)).Bold(true).Render("● namespace resources"), width, colorBgMain),
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
	}
	metrics := m.monitoringOverview.NamespaceMetrics
	telemetryStatus := ""
	if m.monitoringOverview.TelemetryPending {
		telemetryStatus = "waiting"
	} else if m.monitoringOverview.TelemetryUnavailable {
		telemetryStatus = "unavailable"
	}
	lines = append(lines, modernResourceMonitorLines(metrics, width, telemetryStatus)...)
	lines = append(lines, modernLine("", width, colorBgMain), modernRule(width), modernLine("", width, colorBgMain))
	if !ok {
		lines = append(lines, modernFieldLine("Namespace", firstNonEmpty(m.monitoringOverview.Namespace, "n/a"), width))
		lines = append(lines, modernFieldLine("Pods", fmt.Sprintf("%.0f / %.0f running", metrics.RunningPods, metrics.TotalPods), width))
		lines = append(lines, modernFieldLine("Restarts", fmt.Sprintf("%.0f last hour", metrics.RestartsLastHour), width))
		if m.monitoringOverview.TelemetryPending {
			lines = append(lines, modernMutedLine("Kubernetes found workloads; Prometheus telemetry is still catching up.", width))
		} else if m.monitoringOverview.AllocationFallbackUsed {
			lines = append(lines, modernMutedLine("Showing database readiness and allocation; live usage telemetry is unavailable.", width))
		} else if m.monitoringOverview.TelemetryUnavailable {
			lines = append(lines, modernMutedLine("No workload telemetry was found. Check the deployment cluster and Prometheus.", width))
		}
		lines = append(lines, modernMutedLine("Press r to refresh monitoring metrics.", width))
		return modernCropLines(lines, width, height)
	}
	lines = append(lines,
		modernLine(lipgloss.NewStyle().Background(lipgloss.Color(colorBgMain)).Foreground(lipgloss.Color(monitoringHealthColor(project))).Bold(true).Render("● "+project.Name+" · "+monitoringHealthLabel(project)), width, colorBgMain),
		modernLine("", width, colorBgMain),
		modernFieldLine("Kind", firstNonEmpty(project.Kind, "project"), width),
		modernFieldLine("Status", firstNonEmpty(project.Status, "unknown"), width),
		modernFieldLine("Namespace", firstNonEmpty(project.Namespace, "n/a"), width),
		modernFieldLine("Pods", fmt.Sprintf("%.0f total / %.0f running", project.TotalPods, project.RunningPods), width),
		modernFieldLine("Pending", fmt.Sprintf("%.0f", project.PendingPods), width),
		modernFieldLine("Failed", fmt.Sprintf("%.0f", project.FailedPods), width),
		modernFieldLine("CPU", modernFormatCPU(project.CPUCores), width),
		modernFieldLine("Memory", modernFormatBytes(project.MemoryBytes), width),
		modernFieldLine("Restarts", fmt.Sprintf("%.0f last hour", project.RestartsLastHour), width),
		modernLine("", width, colorBgMain),
		modernRule(width),
		modernLine("", width, colorBgMain),
		modernMutedLine("Use up/down to inspect another project.", width),
	)
	return modernCropLines(lines, width, height)
}

func modernResourceMonitorLines(metrics api.MonitoringNamespaceMetrics, width int, telemetryStatus string) []string {
	lines := []string{}
	cpuPct := resourcePercent(metrics.CPURequestsUsed, metrics.CPURequestsLimit, metrics.CPUCores, 8)
	memPct := resourcePercent(metrics.MemoryRequestsUsed, metrics.MemoryRequestsLimit, metrics.MemoryBytes, 16*1024*1024*1024)
	storagePct := resourcePercent(metrics.StorageRequestsUsed, metrics.StorageRequestsLimit, 0, 0)
	networkPct := networkPercent(metrics.NetworkReceiveBytesPerSecond, metrics.NetworkTransmitBytesPerSecond)
	telemetryAvailable := telemetryStatus == ""
	items := []resourceMonitorItem{
		{label: "CPU CORE", value: cpuPct, color: colorInfo, available: telemetryAvailable || metrics.CPURequestsUsed > 0, unavailableLabel: telemetryStatus},
		{label: "MEMORY (RAM)", value: memPct, color: colorWarning, available: telemetryAvailable || metrics.MemoryRequestsUsed > 0, unavailableLabel: telemetryStatus},
		{label: "NVME STORAGE", value: storagePct, color: colorError, available: metrics.StorageRequestsUsed > 0, unavailableLabel: firstNonEmpty(telemetryStatus, "unavailable")},
		{label: "NETWORK UP/DOWN", value: networkPct, color: colorSuccess, available: telemetryAvailable, unavailableLabel: telemetryStatus},
	}
	for index, item := range items {
		if index > 0 {
			lines = append(lines, modernLine("", width, colorBgMain))
		}
		lines = append(lines, modernResourceBar(item, width)...)
	}
	return lines
}

type resourceMonitorItem struct {
	label            string
	value            int
	color            string
	available        bool
	unavailableLabel string
}

func modernResourceBar(item resourceMonitorItem, width int) []string {
	value := clamp(item.value, 0, 100)
	valueText := fmt.Sprintf("%d%%", value)
	if !item.available {
		valueText = firstNonEmpty(item.unavailableLabel, "unavailable")
	}
	labelStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgMain)).
		Foreground(lipgloss.Color(colorTitle)).
		Bold(true)
	valueStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgMain)).
		Foreground(lipgloss.Color(item.color)).
		Bold(true)
	label := labelStyle.Render(item.label)
	pct := valueStyle.Render(valueText)
	header := label + spaces(max(width-visibleLen(label)-visibleLen(pct), 1)) + pct
	barWidth := max(width-2, 8)
	filled := int(math.Round(float64(barWidth) * float64(value) / 100))
	filled = clamp(filled, 0, barWidth)
	empty := max(barWidth-filled, 0)
	bar := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgMain)).
		Foreground(lipgloss.Color(item.color)).
		Render(strings.Repeat("█", filled)) +
		lipgloss.NewStyle().
			Background(lipgloss.Color(colorBgMain)).
			Foreground(lipgloss.Color(colorTrack)).
			Render(strings.Repeat("█", empty))
	return []string{
		modernLine(header, width, colorBgMain),
		modernLine(bar, width, colorBgMain),
	}
}

func modernPodRow(pod api.PodSummary, width int, active bool) string {
	rowBg := colorBgMain
	nameStyle := mainBodyStyle(rowBg)
	phaseStyle := lipgloss.NewStyle().Background(lipgloss.Color(rowBg)).Foreground(lipgloss.Color(podPhaseColor(pod.Phase))).Bold(true)
	prefix := "  "
	if active {
		rowBg = colorBgActive
		nameStyle = mainTitleStyle(rowBg)
		phaseStyle = phaseStyle.Background(lipgloss.Color(rowBg))
		prefix = "> "
	}
	phase := strings.ToLower(firstNonEmpty(pod.Phase, "unknown"))
	nameWidth := max(width-14, 8)
	name := nameStyle.Render(prefix + truncatePlain(pod.Name, nameWidth))
	state := phaseStyle.Render(truncatePlain(phase, 12))
	line := name + spaces(max(width-visibleLen(name)-visibleLen(state), 1)) + state
	return lipgloss.NewStyle().Background(lipgloss.Color(rowBg)).Render(pad(line, width))
}

func modernMonitoringProjectRow(project api.MonitoringProjectMetrics, width int, active bool) string {
	rowBg := colorBgMain
	nameStyle := mainBodyStyle(rowBg)
	healthStyle := lipgloss.NewStyle().Background(lipgloss.Color(rowBg)).Foreground(lipgloss.Color(monitoringHealthColor(project))).Bold(true)
	prefix := "  "
	if active {
		rowBg = colorBgActive
		nameStyle = mainTitleStyle(rowBg)
		healthStyle = healthStyle.Background(lipgloss.Color(rowBg))
		prefix = "> "
	}
	label := monitoringHealthLabel(project)
	nameWidth := max(width-visibleLen(label)-4, 8)
	name := nameStyle.Render(prefix + truncatePlain(project.Name, nameWidth))
	health := healthStyle.Render(label)
	line := name + spaces(max(width-visibleLen(name)-visibleLen(health), 1)) + health
	return lipgloss.NewStyle().Background(lipgloss.Color(rowBg)).Render(pad(line, width))
}

func modernLogLine(line api.LogLine, width int) string {
	level := strings.ToUpper(firstNonEmpty(line.Level, "info"))
	levelStyle := lipgloss.NewStyle().
		Background(lipgloss.Color(colorBgMain)).
		Foreground(lipgloss.Color(logLevelColor(level))).
		Bold(true).
		Render(pad(truncatePlain(level, 7), 7))
	body := mainBodyStyle(colorBgMain).Render(" " + truncatePlain(line.Message, max(width-9, 8)))
	return modernLine(levelStyle+body, width, colorBgMain)
}

func podPhaseColor(phase string) string {
	switch strings.ToLower(strings.TrimSpace(phase)) {
	case "running", "succeeded":
		return colorSuccess
	case "pending":
		return colorWarning
	case "failed":
		return colorError
	default:
		return colorMuted
	}
}

func logLevelColor(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "error":
		return colorError
	case "warn":
		return colorWarning
	case "success":
		return colorSuccess
	default:
		return colorMuted
	}
}

func monitoringHealthLabel(project api.MonitoringProjectMetrics) string {
	if project.Name == "" {
		return "overview"
	}
	if project.FailedPods > 0 || project.PendingPods > 0 {
		return "attention"
	}
	if project.TotalPods > 0 && project.RunningPods >= project.TotalPods {
		return "healthy"
	}
	if project.TotalPods > 0 {
		return "starting"
	}
	return "no pods"
}

func monitoringHealthColor(project api.MonitoringProjectMetrics) string {
	switch monitoringHealthLabel(project) {
	case "healthy":
		return colorSuccess
	case "attention":
		return colorError
	case "starting":
		return colorWarning
	default:
		return colorPrimary
	}
}

func resourcePercent(used, limit, fallbackUsed, fallbackLimit float64) int {
	if limit > 0 {
		return percent(used, limit)
	}
	if fallbackLimit > 0 {
		return percent(fallbackUsed, fallbackLimit)
	}
	return 0
}

func networkPercent(receiveBytesPerSecond, transmitBytesPerSecond float64) int {
	total := math.Max(receiveBytesPerSecond+transmitBytesPerSecond, 0)
	if total <= 0 {
		return 0
	}
	// A quiet workspace should still show small activity; 25 MiB/s fills the bar.
	return percent(total, 25*1024*1024)
}

func percent(value, total float64) int {
	if total <= 0 || math.IsNaN(value) || math.IsNaN(total) || math.IsInf(value, 0) || math.IsInf(total, 0) {
		return 0
	}
	return clamp(int(math.Round((value/total)*100)), 0, 100)
}

func modernFormatCPU(value float64) string {
	if value <= 0 {
		return "0m"
	}
	if value < 1 {
		return fmt.Sprintf("%.0fm", value*1000)
	}
	if value >= 10 {
		return fmt.Sprintf("%.0f cores", value)
	}
	return fmt.Sprintf("%.2f cores", value)
}

func modernFormatBytes(value float64) string {
	if value <= 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return "0 B"
	}
	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	index := 0
	for value >= 1024 && index < len(units)-1 {
		value /= 1024
		index++
	}
	if index == 0 {
		return fmt.Sprintf("%.0f %s", value, units[index])
	}
	return fmt.Sprintf("%.1f %s", value, units[index])
}

func modernDuration(seconds int) string {
	switch {
	case seconds <= 0:
		return "n/a"
	case seconds < 60:
		return fmt.Sprintf("%ds", seconds)
	case seconds < 3600:
		return fmt.Sprintf("%dm", seconds/60)
	case seconds < 86400:
		return fmt.Sprintf("%dh", seconds/3600)
	default:
		return fmt.Sprintf("%dd", seconds/86400)
	}
}
