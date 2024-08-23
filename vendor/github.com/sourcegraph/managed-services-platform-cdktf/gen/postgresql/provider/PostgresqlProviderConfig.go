package provider


type PostgresqlProviderConfig struct {
	// Alias name.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#alias PostgresqlProvider#alias}
	Alias *string `field:"optional" json:"alias" yaml:"alias"`
	// Use rds_iam instead of password authentication (see: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.IAMDBAuth.html).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#aws_rds_iam_auth PostgresqlProvider#aws_rds_iam_auth}
	AwsRdsIamAuth interface{} `field:"optional" json:"awsRdsIamAuth" yaml:"awsRdsIamAuth"`
	// AWS profile to use for IAM auth.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#aws_rds_iam_profile PostgresqlProvider#aws_rds_iam_profile}
	AwsRdsIamProfile *string `field:"optional" json:"awsRdsIamProfile" yaml:"awsRdsIamProfile"`
	// AWS region to use for IAM auth.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#aws_rds_iam_region PostgresqlProvider#aws_rds_iam_region}
	AwsRdsIamRegion *string `field:"optional" json:"awsRdsIamRegion" yaml:"awsRdsIamRegion"`
	// Use MS Azure identity OAuth token (see: https://learn.microsoft.com/en-us/azure/postgresql/flexible-server/how-to-configure-sign-in-azure-ad-authentication).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#azure_identity_auth PostgresqlProvider#azure_identity_auth}
	AzureIdentityAuth interface{} `field:"optional" json:"azureIdentityAuth" yaml:"azureIdentityAuth"`
	// MS Azure tenant ID (see: https://registry.terraform.io/providers/hashicorp/azurerm/latest/docs/data-sources/client_config.html).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#azure_tenant_id PostgresqlProvider#azure_tenant_id}
	AzureTenantId *string `field:"optional" json:"azureTenantId" yaml:"azureTenantId"`
	// clientcert block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#clientcert PostgresqlProvider#clientcert}
	Clientcert *PostgresqlProviderClientcert `field:"optional" json:"clientcert" yaml:"clientcert"`
	// Maximum wait for connection, in seconds. Zero or not specified means wait indefinitely.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#connect_timeout PostgresqlProvider#connect_timeout}
	ConnectTimeout *float64 `field:"optional" json:"connectTimeout" yaml:"connectTimeout"`
	// The name of the database to connect to in order to conenct to (defaults to `postgres`).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#database PostgresqlProvider#database}
	Database *string `field:"optional" json:"database" yaml:"database"`
	// Database username associated to the connected user (for user name maps).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#database_username PostgresqlProvider#database_username}
	DatabaseUsername *string `field:"optional" json:"databaseUsername" yaml:"databaseUsername"`
	// Specify the expected version of PostgreSQL.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#expected_version PostgresqlProvider#expected_version}
	ExpectedVersion *string `field:"optional" json:"expectedVersion" yaml:"expectedVersion"`
	// Service account to impersonate when using GCP IAM authentication.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#gcp_iam_impersonate_service_account PostgresqlProvider#gcp_iam_impersonate_service_account}
	GcpIamImpersonateServiceAccount *string `field:"optional" json:"gcpIamImpersonateServiceAccount" yaml:"gcpIamImpersonateServiceAccount"`
	// Name of PostgreSQL server address to connect to.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#host PostgresqlProvider#host}
	Host *string `field:"optional" json:"host" yaml:"host"`
	// Maximum number of connections to establish to the database. Zero means unlimited.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#max_connections PostgresqlProvider#max_connections}
	MaxConnections *float64 `field:"optional" json:"maxConnections" yaml:"maxConnections"`
	// Password to be used if the PostgreSQL server demands password authentication.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#password PostgresqlProvider#password}
	Password *string `field:"optional" json:"password" yaml:"password"`
	// The PostgreSQL port number to connect to at the server host, or socket file name extension for Unix-domain connections.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#port PostgresqlProvider#port}
	Port *float64 `field:"optional" json:"port" yaml:"port"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#scheme PostgresqlProvider#scheme}.
	Scheme *string `field:"optional" json:"scheme" yaml:"scheme"`
	// This option determines whether or with what priority a secure SSL TCP/IP connection will be negotiated with the PostgreSQL server.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#sslmode PostgresqlProvider#sslmode}
	Sslmode *string `field:"optional" json:"sslmode" yaml:"sslmode"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#ssl_mode PostgresqlProvider#ssl_mode}.
	SslMode *string `field:"optional" json:"sslMode" yaml:"sslMode"`
	// The SSL server root certificate file path. The file must contain PEM encoded data.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#sslrootcert PostgresqlProvider#sslrootcert}
	Sslrootcert *string `field:"optional" json:"sslrootcert" yaml:"sslrootcert"`
	// Specify if the user to connect as is a Postgres superuser or not.If not, some feature might be disabled (e.g.: Refreshing state password from Postgres).
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#superuser PostgresqlProvider#superuser}
	Superuser interface{} `field:"optional" json:"superuser" yaml:"superuser"`
	// PostgreSQL user name to connect as.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/sourcegraph/postgresql/1.23.0-sg.2/docs#username PostgresqlProvider#username}
	Username *string `field:"optional" json:"username" yaml:"username"`
}

