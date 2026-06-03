package theme

type Palette struct {
	Primary  string
	BgMain   string
	BgSide   string
	BgCard   string
	BgPill   string
	BgActive string
	Text     string
	Muted    string
	Title    string
	Border   string
}

type Theme struct {
	Label   string
	Palette Palette
}

var themes = []Theme{
	{
		Label: "Dark",
		Palette: Palette{
			Primary:  "#7c3aed",
			BgMain:   "#171827",
			BgSide:   "#171827",
			BgCard:   "#1d1f31",
			BgPill:   "#25283b",
			BgActive: "#4936a3",
			Text:     "#e8eaf6",
			Muted:    "#969caf",
			Title:    "#f8fafc",
			Border:   "#3c465e",
		},
	},
	{
		Label: "Light",
		Palette: Palette{
			Primary:  "#6d28d9",
			BgMain:   "#f7f7fb",
			BgSide:   "#f7f7fb",
			BgCard:   "#ffffff",
			BgPill:   "#ececf6",
			BgActive: "#ddd6fe",
			Text:     "#1f2333",
			Muted:    "#6b7280",
			Title:    "#111827",
			Border:   "#c7c9d8",
		},
	},
	{
		Label: "Orange",
		Palette: Palette{
			Primary:  "#f97316",
			BgMain:   "#191512",
			BgSide:   "#191512",
			BgCard:   "#241c16",
			BgPill:   "#332317",
			BgActive: "#7c2d12",
			Text:     "#fff7ed",
			Muted:    "#c7a894",
			Title:    "#fffbeb",
			Border:   "#6b3f24",
		},
	},
	{
		Label: "Green",
		Palette: Palette{
			Primary:  "#22c55e",
			BgMain:   "#101914",
			BgSide:   "#101914",
			BgCard:   "#16241b",
			BgPill:   "#1d3325",
			BgActive: "#14532d",
			Text:     "#ecfdf5",
			Muted:    "#9ab8a6",
			Title:    "#f0fdf4",
			Border:   "#2f5d42",
		},
	},
	{
		Label: "Ocean",
		Palette: Palette{
			Primary:  "#06b6d4",
			BgMain:   "#101820",
			BgSide:   "#101820",
			BgCard:   "#152431",
			BgPill:   "#183342",
			BgActive: "#155e75",
			Text:     "#ecfeff",
			Muted:    "#92b6c2",
			Title:    "#f0fdfa",
			Border:   "#2e5a6c",
		},
	},
	{
		Label: "Rose",
		Palette: Palette{
			Primary:  "#f43f5e",
			BgMain:   "#1a1118",
			BgSide:   "#1a1118",
			BgCard:   "#261722",
			BgPill:   "#3b1d2d",
			BgActive: "#881337",
			Text:     "#fff1f2",
			Muted:    "#c9a0ad",
			Title:    "#fff7f8",
			Border:   "#6f3448",
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
