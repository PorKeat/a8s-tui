package ui

import (
	uitheme "github.com/PorKeat/a8s-tui/ui/theme"

	"charm.land/lipgloss/v2"
)

var (
	colorPrimary        = "#ea580c"
	colorBgMain         = "#f7f7fb"
	colorBgSide         = "#f1f2f8"
	colorBgCard         = "#ffffff"
	colorBgPill         = "#f1ebe7"
	colorBgActive       = "#ffdfcc"
	colorText           = "#1f2333"
	colorMuted          = "#6b7280"
	colorTitle          = "#111827"
	colorBorder         = "#d6c9c0"
	colorOnPrimary      = "#ffffff"
	colorInfo           = "#2563a8"
	colorSuccess        = "#18864b"
	colorWarning        = "#9a5b00"
	colorError          = "#c92a3d"
	colorTrack          = "#d7d9e4"
	colorBgDanger       = "#feecef"
	colorBgDangerActive = "#fbd5dc"
)

const (
	reset = "\x1b[0m"
	bold  = "\x1b[1m"
	dim   = "\x1b[2m"
)

var (
	fgLogo   = "\x1b[38;2;234;88;12m"
	fgLogo2  = "\x1b[38;2;234;88;12m"
	fgText   = "\x1b[38;2;31;35;51m"
	fgMuted  = "\x1b[38;2;107;114;128m"
	fgBlue   = "\x1b[38;2;37;99;168m"
	fgOrange = "\x1b[38;2;234;88;12m"
	fgPurple = "\x1b[38;2;234;88;12m"
	fgWarm   = "\x1b[38;2;31;35;51m"
	fgWhite  = "\x1b[38;2;17;24;39m"
	fgKey    = "\x1b[38;2;234;88;12m"
	fgAccent = "\x1b[38;2;37;99;168m"
	fgWarn   = "\x1b[38;2;154;91;0m"
	fgError  = "\x1b[38;2;201;42;61m"
	fgGreen  = "\x1b[38;2;24;134;75m"
	fgBorder = "\x1b[38;2;214;201;192m"
	bgDark   = "\x1b[48;2;247;247;251m"
	bgPane   = "\x1b[48;2;247;247;251m"
	bgSide   = "\x1b[48;2;241;242;248m"
	bgCard   = "\x1b[48;2;255;255;255m"
	bgPill   = "\x1b[48;2;241;235;231m"
	bgSelect = "\x1b[48;2;255;223;204m"
	bgBar    = "\x1b[48;2;241;242;248m"
)

var (
	icons          = uitheme.IconSet()
	nfSearch       = icons.Search
	nfFolder       = icons.Folder
	nfDeploy       = icons.Deploy
	nfShield       = icons.Shield
	nfFile         = icons.File
	nfChart        = icons.Chart
	nfGear         = icons.Gear
	nfDatabase     = icons.Database
	nfMicroservice = icons.Microservice
	nfProject      = icons.Project
)

var (
	styleSide = lipgloss.NewStyle().
			Background(lipgloss.Color(colorBgSide)).
			Foreground(lipgloss.Color(colorText))
	styleMain = lipgloss.NewStyle().
			Background(lipgloss.Color(colorBgMain)).
			Foreground(lipgloss.Color(colorText))
	styleCard = lipgloss.NewStyle().
			Background(lipgloss.Color(colorBgCard)).
			Foreground(lipgloss.Color(colorText))
	stylePill = lipgloss.NewStyle().
			Background(lipgloss.Color(colorBgPill)).
			Foreground(lipgloss.Color(colorText)).
			Padding(0, 1)
	styleActive = lipgloss.NewStyle().
			Background(lipgloss.Color(colorBgActive)).
			Foreground(lipgloss.Color(colorPrimary))
	stylePrimary = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorPrimary)).
			Bold(true)
	styleTitle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorTitle)).
			Bold(true)
	styleMuted = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorMuted))
	styleBorder = lipgloss.NewStyle().
			Foreground(lipgloss.Color(colorBorder))
	styleSideMuted = lipgloss.NewStyle().
			Background(lipgloss.Color(colorBgSide)).
			Foreground(lipgloss.Color(colorMuted))
	styleSideText = lipgloss.NewStyle().
			Background(lipgloss.Color(colorBgSide)).
			Foreground(lipgloss.Color(colorText))
	styleSideBorder = lipgloss.NewStyle().
			Background(lipgloss.Color(colorBgSide)).
			Foreground(lipgloss.Color(colorBorder))
)

type launcherItem struct {
	icon   string
	label  string
	key    string
	action string
}

type navigationItem struct {
	group  string
	label  string
	key    string
	page   appPage
	action string
}
