package deploy

type EngineOption struct {
	ID       string
	Label    string
	Versions []string
}

var EngineOptions = []EngineOption{
	{ID: "postgresql", Label: "PostgreSQL", Versions: []string{"18", "17", "16", "15"}},
	{ID: "mysql", Label: "MySQL", Versions: []string{"8.4", "8.0"}},
	{ID: "mongodb", Label: "MongoDB", Versions: []string{"8.0", "7.0"}},
	{ID: "redis", Label: "Redis", Versions: []string{"8.0", "7.2"}},
	{ID: "cassandra", Label: "Cassandra", Versions: []string{"5.0.5", "4.1.3"}},
}

var SizeOptions = []string{"small", "medium", "large"}
