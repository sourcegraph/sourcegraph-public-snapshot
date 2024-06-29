package schema

type SelfUpdateDefinition struct {
	SelfUpdate ComponentUpdateInformation   `yaml:"self_update"`
	Components []ComponentUpdateInformation `yaml:"components"`
}

type ComponentUpdateInformation struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	DisplayName string `yaml:"display_name"`
	UpdateUrl   string `yaml:"self_update_url"`
}
