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

func PaletteFor(dark bool) Palette {
	if dark {
		return Palette{
			Primary:  "#f56618",
			BgMain:   "#130e0b",
			BgSide:   "#1f1712",
			BgCard:   "#2b221b",
			BgPill:   "#231b16",
			BgActive: "#4f2f20",
			Text:     "#ccbeb2",
			Muted:    "#9f9186",
			Title:    "#f5f1eb",
			Border:   "#5b4638",
		}
	}
	return Palette{
		Primary:  "#f56618",
		BgMain:   "#fbf7f2",
		BgSide:   "#f0e6db",
		BgCard:   "#fffaf4",
		BgPill:   "#eadccd",
		BgActive: "#ffe4d4",
		Text:     "#49372d",
		Muted:    "#7f6d62",
		Title:    "#211712",
		Border:   "#b79b89",
	}
}
