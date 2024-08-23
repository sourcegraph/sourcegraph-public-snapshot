package dataopsgenieteam


type DataOpsgenieTeamMember struct {
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/opsgenie/opsgenie/0.6.35/docs/data-sources/team#id DataOpsgenieTeam#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/opsgenie/opsgenie/0.6.35/docs/data-sources/team#role DataOpsgenieTeam#role}.
	Role *string `field:"optional" json:"role" yaml:"role"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/opsgenie/opsgenie/0.6.35/docs/data-sources/team#username DataOpsgenieTeam#username}.
	Username *string `field:"optional" json:"username" yaml:"username"`
}

