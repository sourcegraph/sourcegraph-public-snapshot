package redisinstance

import (
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type RedisInstanceConfig struct {
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
	// Redis memory size in GiB.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#memory_size_gb RedisInstance#memory_size_gb}
	MemorySizeGb *float64 `field:"required" json:"memorySizeGb" yaml:"memorySizeGb"`
	// The ID of the instance or a fully qualified identifier for the instance.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#name RedisInstance#name}
	Name *string `field:"required" json:"name" yaml:"name"`
	// Only applicable to STANDARD_HA tier which protects the instance against zonal failures by provisioning it across two zones.
	//
	// If provided, it must be a different zone from the one provided in
	// [locationId].
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#alternative_location_id RedisInstance#alternative_location_id}
	AlternativeLocationId *string `field:"optional" json:"alternativeLocationId" yaml:"alternativeLocationId"`
	// Optional.
	//
	// Indicates whether OSS Redis AUTH is enabled for the
	// instance. If set to "true" AUTH is enabled on the instance.
	// Default value is "false" meaning AUTH is disabled.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#auth_enabled RedisInstance#auth_enabled}
	AuthEnabled interface{} `field:"optional" json:"authEnabled" yaml:"authEnabled"`
	// The full name of the Google Compute Engine network to which the instance is connected.
	//
	// If left unspecified, the default network
	// will be used.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#authorized_network RedisInstance#authorized_network}
	AuthorizedNetwork *string `field:"optional" json:"authorizedNetwork" yaml:"authorizedNetwork"`
	// The connection mode of the Redis instance. Default value: "DIRECT_PEERING" Possible values: ["DIRECT_PEERING", "PRIVATE_SERVICE_ACCESS"].
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#connect_mode RedisInstance#connect_mode}
	ConnectMode *string `field:"optional" json:"connectMode" yaml:"connectMode"`
	// Optional.
	//
	// The KMS key reference that you want to use to encrypt the data at rest for this Redis
	// instance. If this is provided, CMEK is enabled.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#customer_managed_key RedisInstance#customer_managed_key}
	CustomerManagedKey *string `field:"optional" json:"customerManagedKey" yaml:"customerManagedKey"`
	// An arbitrary and optional user-provided name for the instance.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#display_name RedisInstance#display_name}
	DisplayName *string `field:"optional" json:"displayName" yaml:"displayName"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#id RedisInstance#id}.
	//
	// Please be aware that the id field is automatically added to all resources in Terraform providers using a Terraform provider SDK version below 2.
	// If you experience problems setting this value it might not be settable. Please take a look at the provider documentation to ensure it should be settable.
	Id *string `field:"optional" json:"id" yaml:"id"`
	// Resource labels to represent user provided metadata.
	//
	// *Note**: This field is non-authoritative, and will only manage the labels present in your configuration.
	// Please refer to the field 'effective_labels' for all of the labels present on the resource.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#labels RedisInstance#labels}
	Labels *map[string]*string `field:"optional" json:"labels" yaml:"labels"`
	// The zone where the instance will be provisioned.
	//
	// If not provided,
	// the service will choose a zone for the instance. For STANDARD_HA tier,
	// instances will be created across two zones for protection against
	// zonal failures. If [alternativeLocationId] is also provided, it must
	// be different from [locationId].
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#location_id RedisInstance#location_id}
	LocationId *string `field:"optional" json:"locationId" yaml:"locationId"`
	// maintenance_policy block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#maintenance_policy RedisInstance#maintenance_policy}
	MaintenancePolicy *RedisInstanceMaintenancePolicy `field:"optional" json:"maintenancePolicy" yaml:"maintenancePolicy"`
	// persistence_config block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#persistence_config RedisInstance#persistence_config}
	PersistenceConfig *RedisInstancePersistenceConfig `field:"optional" json:"persistenceConfig" yaml:"persistenceConfig"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#project RedisInstance#project}.
	Project *string `field:"optional" json:"project" yaml:"project"`
	// Optional.
	//
	// Read replica mode. Can only be specified when trying to create the instance.
	// If not set, Memorystore Redis backend will default to READ_REPLICAS_DISABLED.
	// - READ_REPLICAS_DISABLED: If disabled, read endpoint will not be provided and the
	// instance cannot scale up or down the number of replicas.
	// - READ_REPLICAS_ENABLED: If enabled, read endpoint will be provided and the instance
	// can scale up and down the number of replicas. Possible values: ["READ_REPLICAS_DISABLED", "READ_REPLICAS_ENABLED"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#read_replicas_mode RedisInstance#read_replicas_mode}
	ReadReplicasMode *string `field:"optional" json:"readReplicasMode" yaml:"readReplicasMode"`
	// Redis configuration parameters, according to http://redis.io/topics/config. Please check Memorystore documentation for the list of supported parameters: https://cloud.google.com/memorystore/docs/redis/reference/rest/v1/projects.locations.instances#Instance.FIELDS.redis_configs.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#redis_configs RedisInstance#redis_configs}
	RedisConfigs *map[string]*string `field:"optional" json:"redisConfigs" yaml:"redisConfigs"`
	// The version of Redis software.
	//
	// If not provided, latest supported
	// version will be used. Please check the API documentation linked
	// at the top for the latest valid values.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#redis_version RedisInstance#redis_version}
	RedisVersion *string `field:"optional" json:"redisVersion" yaml:"redisVersion"`
	// The name of the Redis region of the instance.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#region RedisInstance#region}
	Region *string `field:"optional" json:"region" yaml:"region"`
	// Optional.
	//
	// The number of replica nodes. The valid range for the Standard Tier with
	// read replicas enabled is [1-5] and defaults to 2. If read replicas are not enabled
	// for a Standard Tier instance, the only valid value is 1 and the default is 1.
	// The valid value for basic tier is 0 and the default is also 0.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#replica_count RedisInstance#replica_count}
	ReplicaCount *float64 `field:"optional" json:"replicaCount" yaml:"replicaCount"`
	// The CIDR range of internal addresses that are reserved for this instance.
	//
	// If not provided, the service will choose an unused /29
	// block, for example, 10.0.0.0/29 or 192.168.0.0/29. Ranges must be
	// unique and non-overlapping with existing subnets in an authorized
	// network.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#reserved_ip_range RedisInstance#reserved_ip_range}
	ReservedIpRange *string `field:"optional" json:"reservedIpRange" yaml:"reservedIpRange"`
	// Optional.
	//
	// Additional IP range for node placement. Required when enabling read replicas on
	// an existing instance. For DIRECT_PEERING mode value must be a CIDR range of size /28, or
	// "auto". For PRIVATE_SERVICE_ACCESS mode value must be the name of an allocated address
	// range associated with the private service access connection, or "auto".
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#secondary_ip_range RedisInstance#secondary_ip_range}
	SecondaryIpRange *string `field:"optional" json:"secondaryIpRange" yaml:"secondaryIpRange"`
	// The service tier of the instance. Must be one of these values:.
	//
	// - BASIC: standalone instance
	// - STANDARD_HA: highly available primary/replica instances Default value: "BASIC" Possible values: ["BASIC", "STANDARD_HA"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#tier RedisInstance#tier}
	Tier *string `field:"optional" json:"tier" yaml:"tier"`
	// timeouts block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#timeouts RedisInstance#timeouts}
	Timeouts *RedisInstanceTimeouts `field:"optional" json:"timeouts" yaml:"timeouts"`
	// The TLS mode of the Redis instance, If not provided, TLS is disabled for the instance.
	//
	// - SERVER_AUTHENTICATION: Client to Server traffic encryption enabled with server authentication Default value: "DISABLED" Possible values: ["SERVER_AUTHENTICATION", "DISABLED"]
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google/5.29.0/docs/resources/redis_instance#transit_encryption_mode RedisInstance#transit_encryption_mode}
	TransitEncryptionMode *string `field:"optional" json:"transitEncryptionMode" yaml:"transitEncryptionMode"`
}

