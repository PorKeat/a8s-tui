package deploy

type Feature struct {
	Label       string
	Description string
	Ready       bool
}

var Features = []Feature{
	{Label: "Single database", Description: "Create a single-instance database deployment.", Ready: true},
	{Label: "Database cluster", Description: "Prepare highly available database clusters.", Ready: true},
	{Label: "Monolithic", Description: "Deploy the current project from Git.", Ready: true},
	{Label: "Microservices", Description: "Deploy a service group with independent releases.", Ready: true},
}
