package theme

import "os"

type Icons struct {
	Search       string
	Folder       string
	Deploy       string
	Shield       string
	File         string
	Chart        string
	Gear         string
	Database     string
	Microservice string
	Project      string
}

func IconSet() Icons {
	if os.Getenv("A8S_NO_ICONS") == "true" {
		return Icons{
			Search:       "O",
			Folder:       ">",
			Deploy:       "^",
			Shield:       "#",
			File:         "-",
			Chart:        "~",
			Gear:         "*",
			Database:     "@",
			Microservice: "&",
			Project:      "+",
		}
	}
	return Icons{
		Search:       "\uf002",
		Folder:       "\uf07b",
		Deploy:       "\uf0ee",
		Shield:       "\uf132",
		File:         "\uf15b",
		Chart:        "\uf080",
		Gear:         "\uf013",
		Database:     "\uf1c0",
		Microservice: "\ue749",
		Project:      "\ue7ba",
	}
}
