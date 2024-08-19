package cloudrunv2service


type CloudRunV2ServiceTemplateVolumesNfs struct {
	// Path that is exported by the NFS server.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#path CloudRunV2Service#path}
	Path *string `field:"required" json:"path" yaml:"path"`
	// Hostname or IP address of the NFS server.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#server CloudRunV2Service#server}
	Server *string `field:"required" json:"server" yaml:"server"`
	// If true, mount the NFS volume as read only.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/cloud_run_v2_service#read_only CloudRunV2Service#read_only}
	ReadOnly interface{} `field:"optional" json:"readOnly" yaml:"readOnly"`
}

