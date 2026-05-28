package ui

import (
	"os"

	"charm.land/lipgloss/v2"
)

var (
	colorPrimary  = "#ff7a30"
	colorBgMain   = "#0d1117"
	colorBgSide   = "#111827"
	colorBgCard   = "#161f2e"
	colorBgPill   = "#1d2a3a"
	colorBgActive = "#223247"
	colorText     = "#d8e2ec"
	colorMuted    = "#8fa3b8"
	colorTitle    = "#f8fbff"
	colorBorder   = "#2b4057"
)

const (
	reset = "\x1b[0m"
	bold  = "\x1b[1m"
	dim   = "\x1b[2m"
)

var (
	fgLogo   = "\x1b[38;2;255;122;48m"
	fgLogo2  = "\x1b[38;2;146;125;255m"
	fgText   = "\x1b[38;2;216;226;236m"
	fgMuted  = "\x1b[38;2;143;163;184m"
	fgBlue   = "\x1b[38;2;96;165;250m"
	fgOrange = "\x1b[38;2;255;122;48m"
	fgPurple = "\x1b[38;2;167;139;250m"
	fgWarm   = "\x1b[38;2;216;226;236m"
	fgWhite  = "\x1b[38;2;248;251;255m"
	fgKey    = "\x1b[38;2;251;191;36m"
	fgAccent = "\x1b[38;2;45;212;191m"
	fgWarn   = "\x1b[38;2;251;191;36m"
	fgError  = "\x1b[38;2;248;113;113m"
	fgGreen  = "\x1b[38;2;74;222;128m"
	fgBorder = "\x1b[38;2;43;64;87m"
	bgDark   = "\x1b[48;2;13;17;23m"
	bgPane   = "\x1b[48;2;13;17;23m"
	bgSide   = "\x1b[48;2;17;24;39m"
	bgCard   = "\x1b[48;2;22;31;46m"
	bgPill   = "\x1b[48;2;29;42;58m"
	bgSelect = "\x1b[48;2;34;50;71m"
	bgBar    = "\x1b[48;2;17;24;39m"
)

var (
	nfSearch       = "\uf002"
	nfFolder       = "\uf07b"
	nfDeploy       = "\uf0ee"
	nfShield       = "\uf132"
	nfFile         = "\uf15b"
	nfChart        = "\uf080"
	nfGear         = "\uf013"
	nfDatabase     = "\uf1c0"
	nfMicroservice = "\ue749"
	nfProject      = "\ue7ba"
)

func init() {
	if os.Getenv("A8S_NO_ICONS") == "true" {
		nfSearch = "O"
		nfFolder = ">"
		nfDeploy = "^"
		nfShield = "#"
		nfFile = "-"
		nfChart = "~"
		nfGear = "*"
		nfDatabase = "@"
		nfMicroservice = "&"
		nfProject = "+"
	}
}

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
