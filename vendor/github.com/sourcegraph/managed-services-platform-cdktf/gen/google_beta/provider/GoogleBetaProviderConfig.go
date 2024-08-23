package provider


type GoogleBetaProviderConfig struct {
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#access_approval_custom_endpoint GoogleBetaProvider#access_approval_custom_endpoint}.
	AccessApprovalCustomEndpoint *string `field:"optional" json:"accessApprovalCustomEndpoint" yaml:"accessApprovalCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#access_context_manager_custom_endpoint GoogleBetaProvider#access_context_manager_custom_endpoint}.
	AccessContextManagerCustomEndpoint *string `field:"optional" json:"accessContextManagerCustomEndpoint" yaml:"accessContextManagerCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#access_token GoogleBetaProvider#access_token}.
	AccessToken *string `field:"optional" json:"accessToken" yaml:"accessToken"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#active_directory_custom_endpoint GoogleBetaProvider#active_directory_custom_endpoint}.
	ActiveDirectoryCustomEndpoint *string `field:"optional" json:"activeDirectoryCustomEndpoint" yaml:"activeDirectoryCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#add_terraform_attribution_label GoogleBetaProvider#add_terraform_attribution_label}.
	AddTerraformAttributionLabel interface{} `field:"optional" json:"addTerraformAttributionLabel" yaml:"addTerraformAttributionLabel"`
	// Alias name.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#alias GoogleBetaProvider#alias}
	Alias *string `field:"optional" json:"alias" yaml:"alias"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#alloydb_custom_endpoint GoogleBetaProvider#alloydb_custom_endpoint}.
	AlloydbCustomEndpoint *string `field:"optional" json:"alloydbCustomEndpoint" yaml:"alloydbCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#api_gateway_custom_endpoint GoogleBetaProvider#api_gateway_custom_endpoint}.
	ApiGatewayCustomEndpoint *string `field:"optional" json:"apiGatewayCustomEndpoint" yaml:"apiGatewayCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#apigee_custom_endpoint GoogleBetaProvider#apigee_custom_endpoint}.
	ApigeeCustomEndpoint *string `field:"optional" json:"apigeeCustomEndpoint" yaml:"apigeeCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#apikeys_custom_endpoint GoogleBetaProvider#apikeys_custom_endpoint}.
	ApikeysCustomEndpoint *string `field:"optional" json:"apikeysCustomEndpoint" yaml:"apikeysCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#app_engine_custom_endpoint GoogleBetaProvider#app_engine_custom_endpoint}.
	AppEngineCustomEndpoint *string `field:"optional" json:"appEngineCustomEndpoint" yaml:"appEngineCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#apphub_custom_endpoint GoogleBetaProvider#apphub_custom_endpoint}.
	ApphubCustomEndpoint *string `field:"optional" json:"apphubCustomEndpoint" yaml:"apphubCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#artifact_registry_custom_endpoint GoogleBetaProvider#artifact_registry_custom_endpoint}.
	ArtifactRegistryCustomEndpoint *string `field:"optional" json:"artifactRegistryCustomEndpoint" yaml:"artifactRegistryCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#assured_workloads_custom_endpoint GoogleBetaProvider#assured_workloads_custom_endpoint}.
	AssuredWorkloadsCustomEndpoint *string `field:"optional" json:"assuredWorkloadsCustomEndpoint" yaml:"assuredWorkloadsCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#backup_dr_custom_endpoint GoogleBetaProvider#backup_dr_custom_endpoint}.
	BackupDrCustomEndpoint *string `field:"optional" json:"backupDrCustomEndpoint" yaml:"backupDrCustomEndpoint"`
	// batching block.
	//
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#batching GoogleBetaProvider#batching}
	Batching *GoogleBetaProviderBatching `field:"optional" json:"batching" yaml:"batching"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#beyondcorp_custom_endpoint GoogleBetaProvider#beyondcorp_custom_endpoint}.
	BeyondcorpCustomEndpoint *string `field:"optional" json:"beyondcorpCustomEndpoint" yaml:"beyondcorpCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#biglake_custom_endpoint GoogleBetaProvider#biglake_custom_endpoint}.
	BiglakeCustomEndpoint *string `field:"optional" json:"biglakeCustomEndpoint" yaml:"biglakeCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#bigquery_analytics_hub_custom_endpoint GoogleBetaProvider#bigquery_analytics_hub_custom_endpoint}.
	BigqueryAnalyticsHubCustomEndpoint *string `field:"optional" json:"bigqueryAnalyticsHubCustomEndpoint" yaml:"bigqueryAnalyticsHubCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#bigquery_connection_custom_endpoint GoogleBetaProvider#bigquery_connection_custom_endpoint}.
	BigqueryConnectionCustomEndpoint *string `field:"optional" json:"bigqueryConnectionCustomEndpoint" yaml:"bigqueryConnectionCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#big_query_custom_endpoint GoogleBetaProvider#big_query_custom_endpoint}.
	BigQueryCustomEndpoint *string `field:"optional" json:"bigQueryCustomEndpoint" yaml:"bigQueryCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#bigquery_datapolicy_custom_endpoint GoogleBetaProvider#bigquery_datapolicy_custom_endpoint}.
	BigqueryDatapolicyCustomEndpoint *string `field:"optional" json:"bigqueryDatapolicyCustomEndpoint" yaml:"bigqueryDatapolicyCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#bigquery_data_transfer_custom_endpoint GoogleBetaProvider#bigquery_data_transfer_custom_endpoint}.
	BigqueryDataTransferCustomEndpoint *string `field:"optional" json:"bigqueryDataTransferCustomEndpoint" yaml:"bigqueryDataTransferCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#bigquery_reservation_custom_endpoint GoogleBetaProvider#bigquery_reservation_custom_endpoint}.
	BigqueryReservationCustomEndpoint *string `field:"optional" json:"bigqueryReservationCustomEndpoint" yaml:"bigqueryReservationCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#bigtable_custom_endpoint GoogleBetaProvider#bigtable_custom_endpoint}.
	BigtableCustomEndpoint *string `field:"optional" json:"bigtableCustomEndpoint" yaml:"bigtableCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#billing_custom_endpoint GoogleBetaProvider#billing_custom_endpoint}.
	BillingCustomEndpoint *string `field:"optional" json:"billingCustomEndpoint" yaml:"billingCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#billing_project GoogleBetaProvider#billing_project}.
	BillingProject *string `field:"optional" json:"billingProject" yaml:"billingProject"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#binary_authorization_custom_endpoint GoogleBetaProvider#binary_authorization_custom_endpoint}.
	BinaryAuthorizationCustomEndpoint *string `field:"optional" json:"binaryAuthorizationCustomEndpoint" yaml:"binaryAuthorizationCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#blockchain_node_engine_custom_endpoint GoogleBetaProvider#blockchain_node_engine_custom_endpoint}.
	BlockchainNodeEngineCustomEndpoint *string `field:"optional" json:"blockchainNodeEngineCustomEndpoint" yaml:"blockchainNodeEngineCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#certificate_manager_custom_endpoint GoogleBetaProvider#certificate_manager_custom_endpoint}.
	CertificateManagerCustomEndpoint *string `field:"optional" json:"certificateManagerCustomEndpoint" yaml:"certificateManagerCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#cloud_asset_custom_endpoint GoogleBetaProvider#cloud_asset_custom_endpoint}.
	CloudAssetCustomEndpoint *string `field:"optional" json:"cloudAssetCustomEndpoint" yaml:"cloudAssetCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#cloud_billing_custom_endpoint GoogleBetaProvider#cloud_billing_custom_endpoint}.
	CloudBillingCustomEndpoint *string `field:"optional" json:"cloudBillingCustomEndpoint" yaml:"cloudBillingCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#cloud_build_custom_endpoint GoogleBetaProvider#cloud_build_custom_endpoint}.
	CloudBuildCustomEndpoint *string `field:"optional" json:"cloudBuildCustomEndpoint" yaml:"cloudBuildCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#cloudbuildv2_custom_endpoint GoogleBetaProvider#cloudbuildv2_custom_endpoint}.
	Cloudbuildv2CustomEndpoint *string `field:"optional" json:"cloudbuildv2CustomEndpoint" yaml:"cloudbuildv2CustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#cloud_build_worker_pool_custom_endpoint GoogleBetaProvider#cloud_build_worker_pool_custom_endpoint}.
	CloudBuildWorkerPoolCustomEndpoint *string `field:"optional" json:"cloudBuildWorkerPoolCustomEndpoint" yaml:"cloudBuildWorkerPoolCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#clouddeploy_custom_endpoint GoogleBetaProvider#clouddeploy_custom_endpoint}.
	ClouddeployCustomEndpoint *string `field:"optional" json:"clouddeployCustomEndpoint" yaml:"clouddeployCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#clouddomains_custom_endpoint GoogleBetaProvider#clouddomains_custom_endpoint}.
	ClouddomainsCustomEndpoint *string `field:"optional" json:"clouddomainsCustomEndpoint" yaml:"clouddomainsCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#cloudfunctions2_custom_endpoint GoogleBetaProvider#cloudfunctions2_custom_endpoint}.
	Cloudfunctions2CustomEndpoint *string `field:"optional" json:"cloudfunctions2CustomEndpoint" yaml:"cloudfunctions2CustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#cloud_functions_custom_endpoint GoogleBetaProvider#cloud_functions_custom_endpoint}.
	CloudFunctionsCustomEndpoint *string `field:"optional" json:"cloudFunctionsCustomEndpoint" yaml:"cloudFunctionsCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#cloud_identity_custom_endpoint GoogleBetaProvider#cloud_identity_custom_endpoint}.
	CloudIdentityCustomEndpoint *string `field:"optional" json:"cloudIdentityCustomEndpoint" yaml:"cloudIdentityCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#cloud_ids_custom_endpoint GoogleBetaProvider#cloud_ids_custom_endpoint}.
	CloudIdsCustomEndpoint *string `field:"optional" json:"cloudIdsCustomEndpoint" yaml:"cloudIdsCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#cloud_quotas_custom_endpoint GoogleBetaProvider#cloud_quotas_custom_endpoint}.
	CloudQuotasCustomEndpoint *string `field:"optional" json:"cloudQuotasCustomEndpoint" yaml:"cloudQuotasCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#cloud_resource_manager_custom_endpoint GoogleBetaProvider#cloud_resource_manager_custom_endpoint}.
	CloudResourceManagerCustomEndpoint *string `field:"optional" json:"cloudResourceManagerCustomEndpoint" yaml:"cloudResourceManagerCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#cloud_run_custom_endpoint GoogleBetaProvider#cloud_run_custom_endpoint}.
	CloudRunCustomEndpoint *string `field:"optional" json:"cloudRunCustomEndpoint" yaml:"cloudRunCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#cloud_run_v2_custom_endpoint GoogleBetaProvider#cloud_run_v2_custom_endpoint}.
	CloudRunV2CustomEndpoint *string `field:"optional" json:"cloudRunV2CustomEndpoint" yaml:"cloudRunV2CustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#cloud_scheduler_custom_endpoint GoogleBetaProvider#cloud_scheduler_custom_endpoint}.
	CloudSchedulerCustomEndpoint *string `field:"optional" json:"cloudSchedulerCustomEndpoint" yaml:"cloudSchedulerCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#cloud_tasks_custom_endpoint GoogleBetaProvider#cloud_tasks_custom_endpoint}.
	CloudTasksCustomEndpoint *string `field:"optional" json:"cloudTasksCustomEndpoint" yaml:"cloudTasksCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#composer_custom_endpoint GoogleBetaProvider#composer_custom_endpoint}.
	ComposerCustomEndpoint *string `field:"optional" json:"composerCustomEndpoint" yaml:"composerCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#compute_custom_endpoint GoogleBetaProvider#compute_custom_endpoint}.
	ComputeCustomEndpoint *string `field:"optional" json:"computeCustomEndpoint" yaml:"computeCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#container_analysis_custom_endpoint GoogleBetaProvider#container_analysis_custom_endpoint}.
	ContainerAnalysisCustomEndpoint *string `field:"optional" json:"containerAnalysisCustomEndpoint" yaml:"containerAnalysisCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#container_attached_custom_endpoint GoogleBetaProvider#container_attached_custom_endpoint}.
	ContainerAttachedCustomEndpoint *string `field:"optional" json:"containerAttachedCustomEndpoint" yaml:"containerAttachedCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#container_aws_custom_endpoint GoogleBetaProvider#container_aws_custom_endpoint}.
	ContainerAwsCustomEndpoint *string `field:"optional" json:"containerAwsCustomEndpoint" yaml:"containerAwsCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#container_azure_custom_endpoint GoogleBetaProvider#container_azure_custom_endpoint}.
	ContainerAzureCustomEndpoint *string `field:"optional" json:"containerAzureCustomEndpoint" yaml:"containerAzureCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#container_custom_endpoint GoogleBetaProvider#container_custom_endpoint}.
	ContainerCustomEndpoint *string `field:"optional" json:"containerCustomEndpoint" yaml:"containerCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#core_billing_custom_endpoint GoogleBetaProvider#core_billing_custom_endpoint}.
	CoreBillingCustomEndpoint *string `field:"optional" json:"coreBillingCustomEndpoint" yaml:"coreBillingCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#credentials GoogleBetaProvider#credentials}.
	Credentials *string `field:"optional" json:"credentials" yaml:"credentials"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#database_migration_service_custom_endpoint GoogleBetaProvider#database_migration_service_custom_endpoint}.
	DatabaseMigrationServiceCustomEndpoint *string `field:"optional" json:"databaseMigrationServiceCustomEndpoint" yaml:"databaseMigrationServiceCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#data_catalog_custom_endpoint GoogleBetaProvider#data_catalog_custom_endpoint}.
	DataCatalogCustomEndpoint *string `field:"optional" json:"dataCatalogCustomEndpoint" yaml:"dataCatalogCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#dataflow_custom_endpoint GoogleBetaProvider#dataflow_custom_endpoint}.
	DataflowCustomEndpoint *string `field:"optional" json:"dataflowCustomEndpoint" yaml:"dataflowCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#dataform_custom_endpoint GoogleBetaProvider#dataform_custom_endpoint}.
	DataformCustomEndpoint *string `field:"optional" json:"dataformCustomEndpoint" yaml:"dataformCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#data_fusion_custom_endpoint GoogleBetaProvider#data_fusion_custom_endpoint}.
	DataFusionCustomEndpoint *string `field:"optional" json:"dataFusionCustomEndpoint" yaml:"dataFusionCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#data_loss_prevention_custom_endpoint GoogleBetaProvider#data_loss_prevention_custom_endpoint}.
	DataLossPreventionCustomEndpoint *string `field:"optional" json:"dataLossPreventionCustomEndpoint" yaml:"dataLossPreventionCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#data_pipeline_custom_endpoint GoogleBetaProvider#data_pipeline_custom_endpoint}.
	DataPipelineCustomEndpoint *string `field:"optional" json:"dataPipelineCustomEndpoint" yaml:"dataPipelineCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#dataplex_custom_endpoint GoogleBetaProvider#dataplex_custom_endpoint}.
	DataplexCustomEndpoint *string `field:"optional" json:"dataplexCustomEndpoint" yaml:"dataplexCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#dataproc_custom_endpoint GoogleBetaProvider#dataproc_custom_endpoint}.
	DataprocCustomEndpoint *string `field:"optional" json:"dataprocCustomEndpoint" yaml:"dataprocCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#dataproc_metastore_custom_endpoint GoogleBetaProvider#dataproc_metastore_custom_endpoint}.
	DataprocMetastoreCustomEndpoint *string `field:"optional" json:"dataprocMetastoreCustomEndpoint" yaml:"dataprocMetastoreCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#datastore_custom_endpoint GoogleBetaProvider#datastore_custom_endpoint}.
	DatastoreCustomEndpoint *string `field:"optional" json:"datastoreCustomEndpoint" yaml:"datastoreCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#datastream_custom_endpoint GoogleBetaProvider#datastream_custom_endpoint}.
	DatastreamCustomEndpoint *string `field:"optional" json:"datastreamCustomEndpoint" yaml:"datastreamCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#default_labels GoogleBetaProvider#default_labels}.
	DefaultLabels *map[string]*string `field:"optional" json:"defaultLabels" yaml:"defaultLabels"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#deployment_manager_custom_endpoint GoogleBetaProvider#deployment_manager_custom_endpoint}.
	DeploymentManagerCustomEndpoint *string `field:"optional" json:"deploymentManagerCustomEndpoint" yaml:"deploymentManagerCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#dialogflow_custom_endpoint GoogleBetaProvider#dialogflow_custom_endpoint}.
	DialogflowCustomEndpoint *string `field:"optional" json:"dialogflowCustomEndpoint" yaml:"dialogflowCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#dialogflow_cx_custom_endpoint GoogleBetaProvider#dialogflow_cx_custom_endpoint}.
	DialogflowCxCustomEndpoint *string `field:"optional" json:"dialogflowCxCustomEndpoint" yaml:"dialogflowCxCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#discovery_engine_custom_endpoint GoogleBetaProvider#discovery_engine_custom_endpoint}.
	DiscoveryEngineCustomEndpoint *string `field:"optional" json:"discoveryEngineCustomEndpoint" yaml:"discoveryEngineCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#dns_custom_endpoint GoogleBetaProvider#dns_custom_endpoint}.
	DnsCustomEndpoint *string `field:"optional" json:"dnsCustomEndpoint" yaml:"dnsCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#document_ai_custom_endpoint GoogleBetaProvider#document_ai_custom_endpoint}.
	DocumentAiCustomEndpoint *string `field:"optional" json:"documentAiCustomEndpoint" yaml:"documentAiCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#document_ai_warehouse_custom_endpoint GoogleBetaProvider#document_ai_warehouse_custom_endpoint}.
	DocumentAiWarehouseCustomEndpoint *string `field:"optional" json:"documentAiWarehouseCustomEndpoint" yaml:"documentAiWarehouseCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#edgecontainer_custom_endpoint GoogleBetaProvider#edgecontainer_custom_endpoint}.
	EdgecontainerCustomEndpoint *string `field:"optional" json:"edgecontainerCustomEndpoint" yaml:"edgecontainerCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#edgenetwork_custom_endpoint GoogleBetaProvider#edgenetwork_custom_endpoint}.
	EdgenetworkCustomEndpoint *string `field:"optional" json:"edgenetworkCustomEndpoint" yaml:"edgenetworkCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#essential_contacts_custom_endpoint GoogleBetaProvider#essential_contacts_custom_endpoint}.
	EssentialContactsCustomEndpoint *string `field:"optional" json:"essentialContactsCustomEndpoint" yaml:"essentialContactsCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#eventarc_custom_endpoint GoogleBetaProvider#eventarc_custom_endpoint}.
	EventarcCustomEndpoint *string `field:"optional" json:"eventarcCustomEndpoint" yaml:"eventarcCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#filestore_custom_endpoint GoogleBetaProvider#filestore_custom_endpoint}.
	FilestoreCustomEndpoint *string `field:"optional" json:"filestoreCustomEndpoint" yaml:"filestoreCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#firebase_app_check_custom_endpoint GoogleBetaProvider#firebase_app_check_custom_endpoint}.
	FirebaseAppCheckCustomEndpoint *string `field:"optional" json:"firebaseAppCheckCustomEndpoint" yaml:"firebaseAppCheckCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#firebase_custom_endpoint GoogleBetaProvider#firebase_custom_endpoint}.
	FirebaseCustomEndpoint *string `field:"optional" json:"firebaseCustomEndpoint" yaml:"firebaseCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#firebase_database_custom_endpoint GoogleBetaProvider#firebase_database_custom_endpoint}.
	FirebaseDatabaseCustomEndpoint *string `field:"optional" json:"firebaseDatabaseCustomEndpoint" yaml:"firebaseDatabaseCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#firebase_extensions_custom_endpoint GoogleBetaProvider#firebase_extensions_custom_endpoint}.
	FirebaseExtensionsCustomEndpoint *string `field:"optional" json:"firebaseExtensionsCustomEndpoint" yaml:"firebaseExtensionsCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#firebase_hosting_custom_endpoint GoogleBetaProvider#firebase_hosting_custom_endpoint}.
	FirebaseHostingCustomEndpoint *string `field:"optional" json:"firebaseHostingCustomEndpoint" yaml:"firebaseHostingCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#firebaserules_custom_endpoint GoogleBetaProvider#firebaserules_custom_endpoint}.
	FirebaserulesCustomEndpoint *string `field:"optional" json:"firebaserulesCustomEndpoint" yaml:"firebaserulesCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#firebase_storage_custom_endpoint GoogleBetaProvider#firebase_storage_custom_endpoint}.
	FirebaseStorageCustomEndpoint *string `field:"optional" json:"firebaseStorageCustomEndpoint" yaml:"firebaseStorageCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#firestore_custom_endpoint GoogleBetaProvider#firestore_custom_endpoint}.
	FirestoreCustomEndpoint *string `field:"optional" json:"firestoreCustomEndpoint" yaml:"firestoreCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#gke_backup_custom_endpoint GoogleBetaProvider#gke_backup_custom_endpoint}.
	GkeBackupCustomEndpoint *string `field:"optional" json:"gkeBackupCustomEndpoint" yaml:"gkeBackupCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#gke_hub2_custom_endpoint GoogleBetaProvider#gke_hub2_custom_endpoint}.
	GkeHub2CustomEndpoint *string `field:"optional" json:"gkeHub2CustomEndpoint" yaml:"gkeHub2CustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#gke_hub_custom_endpoint GoogleBetaProvider#gke_hub_custom_endpoint}.
	GkeHubCustomEndpoint *string `field:"optional" json:"gkeHubCustomEndpoint" yaml:"gkeHubCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#gkehub_feature_custom_endpoint GoogleBetaProvider#gkehub_feature_custom_endpoint}.
	GkehubFeatureCustomEndpoint *string `field:"optional" json:"gkehubFeatureCustomEndpoint" yaml:"gkehubFeatureCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#gkeonprem_custom_endpoint GoogleBetaProvider#gkeonprem_custom_endpoint}.
	GkeonpremCustomEndpoint *string `field:"optional" json:"gkeonpremCustomEndpoint" yaml:"gkeonpremCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#healthcare_custom_endpoint GoogleBetaProvider#healthcare_custom_endpoint}.
	HealthcareCustomEndpoint *string `field:"optional" json:"healthcareCustomEndpoint" yaml:"healthcareCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#iam2_custom_endpoint GoogleBetaProvider#iam2_custom_endpoint}.
	Iam2CustomEndpoint *string `field:"optional" json:"iam2CustomEndpoint" yaml:"iam2CustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#iam_beta_custom_endpoint GoogleBetaProvider#iam_beta_custom_endpoint}.
	IamBetaCustomEndpoint *string `field:"optional" json:"iamBetaCustomEndpoint" yaml:"iamBetaCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#iam_credentials_custom_endpoint GoogleBetaProvider#iam_credentials_custom_endpoint}.
	IamCredentialsCustomEndpoint *string `field:"optional" json:"iamCredentialsCustomEndpoint" yaml:"iamCredentialsCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#iam_custom_endpoint GoogleBetaProvider#iam_custom_endpoint}.
	IamCustomEndpoint *string `field:"optional" json:"iamCustomEndpoint" yaml:"iamCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#iam_workforce_pool_custom_endpoint GoogleBetaProvider#iam_workforce_pool_custom_endpoint}.
	IamWorkforcePoolCustomEndpoint *string `field:"optional" json:"iamWorkforcePoolCustomEndpoint" yaml:"iamWorkforcePoolCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#iap_custom_endpoint GoogleBetaProvider#iap_custom_endpoint}.
	IapCustomEndpoint *string `field:"optional" json:"iapCustomEndpoint" yaml:"iapCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#identity_platform_custom_endpoint GoogleBetaProvider#identity_platform_custom_endpoint}.
	IdentityPlatformCustomEndpoint *string `field:"optional" json:"identityPlatformCustomEndpoint" yaml:"identityPlatformCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#impersonate_service_account GoogleBetaProvider#impersonate_service_account}.
	ImpersonateServiceAccount *string `field:"optional" json:"impersonateServiceAccount" yaml:"impersonateServiceAccount"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#impersonate_service_account_delegates GoogleBetaProvider#impersonate_service_account_delegates}.
	ImpersonateServiceAccountDelegates *[]*string `field:"optional" json:"impersonateServiceAccountDelegates" yaml:"impersonateServiceAccountDelegates"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#integration_connectors_custom_endpoint GoogleBetaProvider#integration_connectors_custom_endpoint}.
	IntegrationConnectorsCustomEndpoint *string `field:"optional" json:"integrationConnectorsCustomEndpoint" yaml:"integrationConnectorsCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#integrations_custom_endpoint GoogleBetaProvider#integrations_custom_endpoint}.
	IntegrationsCustomEndpoint *string `field:"optional" json:"integrationsCustomEndpoint" yaml:"integrationsCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#kms_custom_endpoint GoogleBetaProvider#kms_custom_endpoint}.
	KmsCustomEndpoint *string `field:"optional" json:"kmsCustomEndpoint" yaml:"kmsCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#logging_custom_endpoint GoogleBetaProvider#logging_custom_endpoint}.
	LoggingCustomEndpoint *string `field:"optional" json:"loggingCustomEndpoint" yaml:"loggingCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#looker_custom_endpoint GoogleBetaProvider#looker_custom_endpoint}.
	LookerCustomEndpoint *string `field:"optional" json:"lookerCustomEndpoint" yaml:"lookerCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#memcache_custom_endpoint GoogleBetaProvider#memcache_custom_endpoint}.
	MemcacheCustomEndpoint *string `field:"optional" json:"memcacheCustomEndpoint" yaml:"memcacheCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#migration_center_custom_endpoint GoogleBetaProvider#migration_center_custom_endpoint}.
	MigrationCenterCustomEndpoint *string `field:"optional" json:"migrationCenterCustomEndpoint" yaml:"migrationCenterCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#ml_engine_custom_endpoint GoogleBetaProvider#ml_engine_custom_endpoint}.
	MlEngineCustomEndpoint *string `field:"optional" json:"mlEngineCustomEndpoint" yaml:"mlEngineCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#monitoring_custom_endpoint GoogleBetaProvider#monitoring_custom_endpoint}.
	MonitoringCustomEndpoint *string `field:"optional" json:"monitoringCustomEndpoint" yaml:"monitoringCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#netapp_custom_endpoint GoogleBetaProvider#netapp_custom_endpoint}.
	NetappCustomEndpoint *string `field:"optional" json:"netappCustomEndpoint" yaml:"netappCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#network_connectivity_custom_endpoint GoogleBetaProvider#network_connectivity_custom_endpoint}.
	NetworkConnectivityCustomEndpoint *string `field:"optional" json:"networkConnectivityCustomEndpoint" yaml:"networkConnectivityCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#network_management_custom_endpoint GoogleBetaProvider#network_management_custom_endpoint}.
	NetworkManagementCustomEndpoint *string `field:"optional" json:"networkManagementCustomEndpoint" yaml:"networkManagementCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#network_security_custom_endpoint GoogleBetaProvider#network_security_custom_endpoint}.
	NetworkSecurityCustomEndpoint *string `field:"optional" json:"networkSecurityCustomEndpoint" yaml:"networkSecurityCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#network_services_custom_endpoint GoogleBetaProvider#network_services_custom_endpoint}.
	NetworkServicesCustomEndpoint *string `field:"optional" json:"networkServicesCustomEndpoint" yaml:"networkServicesCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#notebooks_custom_endpoint GoogleBetaProvider#notebooks_custom_endpoint}.
	NotebooksCustomEndpoint *string `field:"optional" json:"notebooksCustomEndpoint" yaml:"notebooksCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#org_policy_custom_endpoint GoogleBetaProvider#org_policy_custom_endpoint}.
	OrgPolicyCustomEndpoint *string `field:"optional" json:"orgPolicyCustomEndpoint" yaml:"orgPolicyCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#os_config_custom_endpoint GoogleBetaProvider#os_config_custom_endpoint}.
	OsConfigCustomEndpoint *string `field:"optional" json:"osConfigCustomEndpoint" yaml:"osConfigCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#os_login_custom_endpoint GoogleBetaProvider#os_login_custom_endpoint}.
	OsLoginCustomEndpoint *string `field:"optional" json:"osLoginCustomEndpoint" yaml:"osLoginCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#parallelstore_custom_endpoint GoogleBetaProvider#parallelstore_custom_endpoint}.
	ParallelstoreCustomEndpoint *string `field:"optional" json:"parallelstoreCustomEndpoint" yaml:"parallelstoreCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#privateca_custom_endpoint GoogleBetaProvider#privateca_custom_endpoint}.
	PrivatecaCustomEndpoint *string `field:"optional" json:"privatecaCustomEndpoint" yaml:"privatecaCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#privileged_access_manager_custom_endpoint GoogleBetaProvider#privileged_access_manager_custom_endpoint}.
	PrivilegedAccessManagerCustomEndpoint *string `field:"optional" json:"privilegedAccessManagerCustomEndpoint" yaml:"privilegedAccessManagerCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#project GoogleBetaProvider#project}.
	Project *string `field:"optional" json:"project" yaml:"project"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#public_ca_custom_endpoint GoogleBetaProvider#public_ca_custom_endpoint}.
	PublicCaCustomEndpoint *string `field:"optional" json:"publicCaCustomEndpoint" yaml:"publicCaCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#pubsub_custom_endpoint GoogleBetaProvider#pubsub_custom_endpoint}.
	PubsubCustomEndpoint *string `field:"optional" json:"pubsubCustomEndpoint" yaml:"pubsubCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#pubsub_lite_custom_endpoint GoogleBetaProvider#pubsub_lite_custom_endpoint}.
	PubsubLiteCustomEndpoint *string `field:"optional" json:"pubsubLiteCustomEndpoint" yaml:"pubsubLiteCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#recaptcha_enterprise_custom_endpoint GoogleBetaProvider#recaptcha_enterprise_custom_endpoint}.
	RecaptchaEnterpriseCustomEndpoint *string `field:"optional" json:"recaptchaEnterpriseCustomEndpoint" yaml:"recaptchaEnterpriseCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#redis_custom_endpoint GoogleBetaProvider#redis_custom_endpoint}.
	RedisCustomEndpoint *string `field:"optional" json:"redisCustomEndpoint" yaml:"redisCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#region GoogleBetaProvider#region}.
	Region *string `field:"optional" json:"region" yaml:"region"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#request_reason GoogleBetaProvider#request_reason}.
	RequestReason *string `field:"optional" json:"requestReason" yaml:"requestReason"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#request_timeout GoogleBetaProvider#request_timeout}.
	RequestTimeout *string `field:"optional" json:"requestTimeout" yaml:"requestTimeout"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#resource_manager_custom_endpoint GoogleBetaProvider#resource_manager_custom_endpoint}.
	ResourceManagerCustomEndpoint *string `field:"optional" json:"resourceManagerCustomEndpoint" yaml:"resourceManagerCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#resource_manager_v3_custom_endpoint GoogleBetaProvider#resource_manager_v3_custom_endpoint}.
	ResourceManagerV3CustomEndpoint *string `field:"optional" json:"resourceManagerV3CustomEndpoint" yaml:"resourceManagerV3CustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#runtimeconfig_custom_endpoint GoogleBetaProvider#runtimeconfig_custom_endpoint}.
	RuntimeconfigCustomEndpoint *string `field:"optional" json:"runtimeconfigCustomEndpoint" yaml:"runtimeconfigCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#runtime_config_custom_endpoint GoogleBetaProvider#runtime_config_custom_endpoint}.
	RuntimeConfigCustomEndpoint *string `field:"optional" json:"runtimeConfigCustomEndpoint" yaml:"runtimeConfigCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#scopes GoogleBetaProvider#scopes}.
	Scopes *[]*string `field:"optional" json:"scopes" yaml:"scopes"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#secret_manager_custom_endpoint GoogleBetaProvider#secret_manager_custom_endpoint}.
	SecretManagerCustomEndpoint *string `field:"optional" json:"secretManagerCustomEndpoint" yaml:"secretManagerCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#secure_source_manager_custom_endpoint GoogleBetaProvider#secure_source_manager_custom_endpoint}.
	SecureSourceManagerCustomEndpoint *string `field:"optional" json:"secureSourceManagerCustomEndpoint" yaml:"secureSourceManagerCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#security_center_custom_endpoint GoogleBetaProvider#security_center_custom_endpoint}.
	SecurityCenterCustomEndpoint *string `field:"optional" json:"securityCenterCustomEndpoint" yaml:"securityCenterCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#securityposture_custom_endpoint GoogleBetaProvider#securityposture_custom_endpoint}.
	SecuritypostureCustomEndpoint *string `field:"optional" json:"securitypostureCustomEndpoint" yaml:"securitypostureCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#security_scanner_custom_endpoint GoogleBetaProvider#security_scanner_custom_endpoint}.
	SecurityScannerCustomEndpoint *string `field:"optional" json:"securityScannerCustomEndpoint" yaml:"securityScannerCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#service_directory_custom_endpoint GoogleBetaProvider#service_directory_custom_endpoint}.
	ServiceDirectoryCustomEndpoint *string `field:"optional" json:"serviceDirectoryCustomEndpoint" yaml:"serviceDirectoryCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#service_management_custom_endpoint GoogleBetaProvider#service_management_custom_endpoint}.
	ServiceManagementCustomEndpoint *string `field:"optional" json:"serviceManagementCustomEndpoint" yaml:"serviceManagementCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#service_networking_custom_endpoint GoogleBetaProvider#service_networking_custom_endpoint}.
	ServiceNetworkingCustomEndpoint *string `field:"optional" json:"serviceNetworkingCustomEndpoint" yaml:"serviceNetworkingCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#service_usage_custom_endpoint GoogleBetaProvider#service_usage_custom_endpoint}.
	ServiceUsageCustomEndpoint *string `field:"optional" json:"serviceUsageCustomEndpoint" yaml:"serviceUsageCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#source_repo_custom_endpoint GoogleBetaProvider#source_repo_custom_endpoint}.
	SourceRepoCustomEndpoint *string `field:"optional" json:"sourceRepoCustomEndpoint" yaml:"sourceRepoCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#spanner_custom_endpoint GoogleBetaProvider#spanner_custom_endpoint}.
	SpannerCustomEndpoint *string `field:"optional" json:"spannerCustomEndpoint" yaml:"spannerCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#sql_custom_endpoint GoogleBetaProvider#sql_custom_endpoint}.
	SqlCustomEndpoint *string `field:"optional" json:"sqlCustomEndpoint" yaml:"sqlCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#storage_custom_endpoint GoogleBetaProvider#storage_custom_endpoint}.
	StorageCustomEndpoint *string `field:"optional" json:"storageCustomEndpoint" yaml:"storageCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#storage_insights_custom_endpoint GoogleBetaProvider#storage_insights_custom_endpoint}.
	StorageInsightsCustomEndpoint *string `field:"optional" json:"storageInsightsCustomEndpoint" yaml:"storageInsightsCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#storage_transfer_custom_endpoint GoogleBetaProvider#storage_transfer_custom_endpoint}.
	StorageTransferCustomEndpoint *string `field:"optional" json:"storageTransferCustomEndpoint" yaml:"storageTransferCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#tags_custom_endpoint GoogleBetaProvider#tags_custom_endpoint}.
	TagsCustomEndpoint *string `field:"optional" json:"tagsCustomEndpoint" yaml:"tagsCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#tags_location_custom_endpoint GoogleBetaProvider#tags_location_custom_endpoint}.
	TagsLocationCustomEndpoint *string `field:"optional" json:"tagsLocationCustomEndpoint" yaml:"tagsLocationCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#terraform_attribution_label_addition_strategy GoogleBetaProvider#terraform_attribution_label_addition_strategy}.
	TerraformAttributionLabelAdditionStrategy *string `field:"optional" json:"terraformAttributionLabelAdditionStrategy" yaml:"terraformAttributionLabelAdditionStrategy"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#tpu_custom_endpoint GoogleBetaProvider#tpu_custom_endpoint}.
	TpuCustomEndpoint *string `field:"optional" json:"tpuCustomEndpoint" yaml:"tpuCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#tpu_v2_custom_endpoint GoogleBetaProvider#tpu_v2_custom_endpoint}.
	TpuV2CustomEndpoint *string `field:"optional" json:"tpuV2CustomEndpoint" yaml:"tpuV2CustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#universe_domain GoogleBetaProvider#universe_domain}.
	UniverseDomain *string `field:"optional" json:"universeDomain" yaml:"universeDomain"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#user_project_override GoogleBetaProvider#user_project_override}.
	UserProjectOverride interface{} `field:"optional" json:"userProjectOverride" yaml:"userProjectOverride"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#vertex_ai_custom_endpoint GoogleBetaProvider#vertex_ai_custom_endpoint}.
	VertexAiCustomEndpoint *string `field:"optional" json:"vertexAiCustomEndpoint" yaml:"vertexAiCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#vmwareengine_custom_endpoint GoogleBetaProvider#vmwareengine_custom_endpoint}.
	VmwareengineCustomEndpoint *string `field:"optional" json:"vmwareengineCustomEndpoint" yaml:"vmwareengineCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#vpc_access_custom_endpoint GoogleBetaProvider#vpc_access_custom_endpoint}.
	VpcAccessCustomEndpoint *string `field:"optional" json:"vpcAccessCustomEndpoint" yaml:"vpcAccessCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#workbench_custom_endpoint GoogleBetaProvider#workbench_custom_endpoint}.
	WorkbenchCustomEndpoint *string `field:"optional" json:"workbenchCustomEndpoint" yaml:"workbenchCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#workflows_custom_endpoint GoogleBetaProvider#workflows_custom_endpoint}.
	WorkflowsCustomEndpoint *string `field:"optional" json:"workflowsCustomEndpoint" yaml:"workflowsCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#workstations_custom_endpoint GoogleBetaProvider#workstations_custom_endpoint}.
	WorkstationsCustomEndpoint *string `field:"optional" json:"workstationsCustomEndpoint" yaml:"workstationsCustomEndpoint"`
	// Docs at Terraform Registry: {@link https://registry.terraform.io/providers/hashicorp/google-beta/5.29.0/docs#zone GoogleBetaProvider#zone}.
	Zone *string `field:"optional" json:"zone" yaml:"zone"`
}

