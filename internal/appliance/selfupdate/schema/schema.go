package schema

type SelfUpdateDefinition struct {
	Version    string                       `yaml:"version"`
	SelfUpdate ComponentUpdateInformation   `yaml:"self_update"`
	Components []ComponentUpdateInformation `yaml:"components"`
}

type ComponentUpdateInformation struct {
	Name        string `yaml:"name"`
	DisplayName string `yaml:"display_name"`
	UpdateUrl   string `yaml:"self_update_url"`
}
