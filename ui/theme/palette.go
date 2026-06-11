package theme

type Palette struct {
	Primary      string
	OnPrimary    string
	BgMain       string
	BgSide       string
	BgCard       string
	BgPill       string
	BgActive     string
	BgDanger     string
	BgDangerOpen string
	Text         string
	Muted        string
	Title        string
	Border       string
	Info         string
	Success      string
	Warning      string
	Error        string
	Track        string
}

type Theme struct {
	Label   string
	Palette Palette
}

var themes = []Theme{
	{
		Label: "Dark",
		Palette: Palette{
			Primary:      "#7c3aed",
			OnPrimary:    "#ffffff",
			BgMain:       "#171827",
			BgSide:       "#171827",
			BgCard:       "#1d1f31",
			BgPill:       "#25283b",
			BgActive:     "#4936a3",
			BgDanger:     "#3f2638",
			BgDangerOpen: "#6f2d4a",
			Text:         "#e8eaf6",
			Muted:        "#969caf",
			Title:        "#f8fafc",
			Border:       "#3c465e",
			Info:         "#7aaeff",
			Success:      "#4ade80",
			Warning:      "#facc15",
			Error:        "#fb7185",
			Track:        "#333747",
		},
	},
	{
		Label: "Light",
		Palette: Palette{
			Primary:      "#ea580c",
			OnPrimary:    "#ffffff",
			BgMain:       "#f7f7fb",
			BgSide:       "#f1f2f8",
			BgCard:       "#ffffff",
			BgPill:       "#f1ebe7",
			BgActive:     "#ffdfcc",
			BgDanger:     "#feecef",
			BgDangerOpen: "#fbd5dc",
			Text:         "#1f2333",
			Muted:        "#6b7280",
			Title:        "#111827",
			Border:       "#d6c9c0",
			Info:         "#2563a8",
			Success:      "#18864b",
			Warning:      "#9a5b00",
			Error:        "#c92a3d",
			Track:        "#d7d9e4",
		},
	},
	{
		Label: "Orange",
		Palette: Palette{
			Primary:      "#f97316",
			OnPrimary:    "#ffffff",
			BgMain:       "#191512",
			BgSide:       "#191512",
			BgCard:       "#241c16",
			BgPill:       "#332317",
			BgActive:     "#7c2d12",
			BgDanger:     "#3f2638",
			BgDangerOpen: "#6f2d4a",
			Text:         "#fff7ed",
			Muted:        "#c7a894",
			Title:        "#fffbeb",
			Border:       "#6b3f24",
			Info:         "#7aaeff",
			Success:      "#4ade80",
			Warning:      "#facc15",
			Error:        "#fb7185",
			Track:        "#42362e",
		},
	},
	{
		Label: "Green",
		Palette: Palette{
			Primary:      "#22c55e",
			OnPrimary:    "#08140d",
			BgMain:       "#101914",
			BgSide:       "#101914",
			BgCard:       "#16241b",
			BgPill:       "#1d3325",
			BgActive:     "#14532d",
			BgDanger:     "#3f2638",
			BgDangerOpen: "#6f2d4a",
			Text:         "#ecfdf5",
			Muted:        "#9ab8a6",
			Title:        "#f0fdf4",
			Border:       "#2f5d42",
			Info:         "#7aaeff",
			Success:      "#4ade80",
			Warning:      "#facc15",
			Error:        "#fb7185",
			Track:        "#294334",
		},
	},
	{
		Label: "Ocean",
		Palette: Palette{
			Primary:      "#06b6d4",
			OnPrimary:    "#07161b",
			BgMain:       "#101820",
			BgSide:       "#101820",
			BgCard:       "#152431",
			BgPill:       "#183342",
			BgActive:     "#155e75",
			BgDanger:     "#3f2638",
			BgDangerOpen: "#6f2d4a",
			Text:         "#ecfeff",
			Muted:        "#92b6c2",
			Title:        "#f0fdfa",
			Border:       "#2e5a6c",
			Info:         "#67e8f9",
			Success:      "#4ade80",
			Warning:      "#facc15",
			Error:        "#fb7185",
			Track:        "#29414d",
		},
	},
	{
		Label: "Rose",
		Palette: Palette{
			Primary:      "#f43f5e",
			OnPrimary:    "#ffffff",
			BgMain:       "#1a1118",
			BgSide:       "#1a1118",
			BgCard:       "#261722",
			BgPill:       "#3b1d2d",
			BgActive:     "#881337",
			BgDanger:     "#4b1829",
			BgDangerOpen: "#7f1d3d",
			Text:         "#fff1f2",
			Muted:        "#c9a0ad",
			Title:        "#fff7f8",
			Border:       "#6f3448",
			Info:         "#7aaeff",
			Success:      "#4ade80",
			Warning:      "#facc15",
			Error:        "#fb7185",
			Track:        "#492b3a",
		},
	},
}

func Themes() []Theme {
	out := make([]Theme, len(themes))
	copy(out, themes)
	return out
}

func ThemeLabels() []string {
	labels := make([]string, len(themes))
	for i, theme := range themes {
		labels[i] = theme.Label
	}
	return labels
}

func NormalizeThemeIndex(index int) int {
	if len(themes) == 0 {
		return 0
	}
	index %= len(themes)
	if index < 0 {
		index += len(themes)
	}
	return index
}

func PaletteForIndex(index int) Palette {
	return themes[NormalizeThemeIndex(index)].Palette
}

func ThemeLabel(index int) string {
	return themes[NormalizeThemeIndex(index)].Label
}

func PaletteFor(dark bool) Palette {
	if dark {
		return PaletteForIndex(0)
	}
	return PaletteForIndex(1)
}
