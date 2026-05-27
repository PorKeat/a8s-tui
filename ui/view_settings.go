package ui

func (m model) renderDashboardFeature(width, height int) []string {
	if m.page == pageDeployment {
		return m.renderDashboardDeployment(width, height)
	}
	nav := m.currentPageNavigationItem()
	lines := make([]string, 0, height)
	lines = append(lines, dashboardHeader(nav.label, m.pageLead(), width)...)
	lines = append(lines, featureCard(nav, width, m.pageBodyLines())...)
	return fillStyled(lines, bgDark, width, height)
}

func dashboardFeatureLine(item navigationItem, width int, active, current bool) string {
	rowBg := colorBgMain
	rowStyle := styleMain
	labelStyle := mainBodyStyle(rowBg)
	prefix := "   "
	if current {
		prefix = " * "
		labelStyle = mainPrimaryStyle(rowBg)
	}
	if active {
		rowBg = colorBgActive
		rowStyle = styleActive
		prefix = " > "
		if !current {
			labelStyle = mainTitleStyle(rowBg)
		} else {
			labelStyle = mainPrimaryStyle(rowBg)
		}
	}
	left := rowStyle.Render(prefix) +
		dashboardItemIcon(item, rowBg) +
		rowStyle.Render("  ") +
		labelStyle.Render(truncatePlain(item.label, max(width-18, 4)))
	right := mainMutedStyle(rowBg).Render(item.key + "   ")
	gap := spaces(max(width-visibleLen(left)-visibleLen(right), 0))
	return left + rowStyle.Render(gap) + rowStyle.Render(right)
}

func (m model) renderGroupFeatureList(width, height int, group string) []string {
	lines := make([]string, 0, height)
	lines = append(lines, paneTitle("main", width, m.focus == focusList))
	lines = append(lines, bgPane+pad("", width)+reset)
	lines = append(lines, bgPane+pad("   "+bold+fgLogo+group+reset, width)+reset)
	description := truncatePlain(groupDescription(group), width-4)
	lines = append(lines, bgPane+pad("   "+fgMuted+description+reset, width)+reset)
	lines = append(lines, bgPane+pad("", width)+reset)
	lines = append(lines, sectionLine("features", width))
	for _, item := range m.navigationItemsForGroup(group) {
		active := item.label == m.selectedNavigationItem().label && item.group == group
		current := item.matchesPage(m.page)
		lines = append(lines, featureListLine(item, width, active, current))
	}
	lines = append(lines, bgPane+pad("", width)+reset)
	hint := truncatePlain("Press enter to open the selected feature.", width-4)
	lines = append(lines, bgPane+pad("   "+fgMuted+hint+reset, width)+reset)
	return fillPane(lines, width, height)
}

func (m model) renderFeaturePage(width, height int) []string {
	lines := make([]string, 0, height)
	nav := m.currentPageNavigationItem()
	lines = append(lines, paneTitle("main", width, m.focus == focusList))
	lines = append(lines, bgPane+pad("", width)+reset)
	lines = append(lines, bgPane+pad("   "+bold+fgLogo+nav.group+reset+fgMuted+" / "+nav.label+reset, width)+reset)
	lead := truncatePlain(m.pageLead(), width-4)
	lines = append(lines, bgPane+pad("   "+fgMuted+lead+reset, width)+reset)
	lines = append(lines, bgPane+pad("", width)+reset)
	for _, text := range m.pageBodyLines() {
		lines = append(lines, bgPane+pad("   "+fgText+truncatePlain(text, width-5)+reset, width)+reset)
	}
	lines = append(lines, bgPane+pad("", width)+reset)
	help := truncatePlain("Use the sidebar to switch sections. Press d to create a database deployment.", width-4)
	lines = append(lines, bgPane+pad("   "+fgMuted+help+reset, width)+reset)
	return fillPane(lines, width, height)
}
