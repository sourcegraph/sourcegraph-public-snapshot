package record

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type RecordConfig struct {
	// Experimental.
	Connection interface{} `field:"optional" json:"connection" yaml:"connection"`
	// Experimental.
	Count interface{} `field:"optional" json:"count" yaml:"count"`
	// Experimental.
	DependsOn *[]cdktf.ITerraformDependable `field:"optional" json:"dependsOn" yaml:"dependsOn"`
	// Experimental.
	ForEach cdktf.ITerraformIterator `field:"optional" json:"forEach" yaml:"forEach"`
	// Experimental.
	Lifecycle *cdktf.TerraformResourceLifecycle `field:"optional" json:"lifecycle" yaml:"lifecycle"`
	// Experimental.
	Provider cdktf.TerraformProvider `field:"optional" json:"provider" yaml:"provider"`
	// Experimental.
	Provisioners *[]interface{} `field:"optional" json:"provisioners" yaml:"provisioners"`
	// The name of the record. **Modifying this attribute will force creation of a new resource.**.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs/resources/record#name Record#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// The type of the record.
	//
	// Available values: `A`, `AAAA`, `CAA`, `CNAME`, `TXT`, `SRV`, `LOC`, `MX`, `NS`, `SPF`, `CERT`, `DNSKEY`, `DS`, `NAPTR`, `SMIMEA`, `SSHFP`, `TLSA`, `URI`, `PTR`, `HTTPS`, `SVCB`. **Modifying this attribute will force creation of a new resource.**
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs/resources/record#type Record#type}
	Type *string `field:"required" json:"type" yaml:"type"`
	// The zone identifier to target for the resource. **Modifying this attribute will force creation of a new resource.**.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs/resources/record#zone_id Record#zone_id}
	ZoneId *string `field:"required" json:"zoneId" yaml:"zoneId"`
	// Allow creation of this record in Terraform to overwrite an existing record, if any.
	//
	// This does not affect the ability to update the record in Terraform and does not prevent other resources within Terraform or manual changes outside Terraform from overwriting this record. **This configuration is not recommended for most environments**. Defaults to `false`.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs/resources/record#allow_overwrite Record#allow_overwrite}
	AllowOverwrite interface{} `field:"optional" json:"allowOverwrite" yaml:"allowOverwrite"`
	// Comments or notes about the DNS record. This field has no effect on DNS responses.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs/resources/record#comment Record#comment}
	Comment *string `field:"optional" json:"comment" yaml:"comment"`
	// data block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs/resources/record#data Record#data}
	Data *RecordData `field:"optional" json:"data" yaml:"data"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs/resources/record#id Record#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// The priority of the record.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs/resources/record#priority Record#priority}
	Priority *float64 `field:"optional" json:"priority" yaml:"priority"`
	// Whether the record gets Cloudflare's origin protection.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs/resources/record#proxied Record#proxied}
	Proxied interface{} `field:"optional" json:"proxied" yaml:"proxied"`
	// Custom tags for the DNS record.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs/resources/record#tags Record#tags}
	Tags *[]*string `field:"optional" json:"tags" yaml:"tags"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs/resources/record#timeouts Record#timeouts}
	Timeouts *RecordTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
	// The TTL of the record.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs/resources/record#ttl Record#ttl}
	Ttl *float64 `field:"optional" json:"ttl" yaml:"ttl"`
	// The value of the record. Conflicts with `data`.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/cloudflare/cloudflare/4.12.0/docs/resources/record#value Record#value}
	Value *string `field:"optional" json:"value" yaml:"value"`
}

