//
// Copyright 2021, Sander van Harmelen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package gitlab

import (
	"net/http"
	"time"
)

// SettingsService handles communication with the application SettingsService
// related methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/settings.html
type SettingsService struct {
	client *Client
}

// Settings represents the GitLab application settings.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/settings.html
//
// The available parameters have been modeled directly after the code, as the
// documentation seems to be inaccurate.
//
// https://gitlab.com/gitlab-org/gitlab/-/blob/v14.9.3-ee/lib/api/settings.rb
// https://gitlab.com/gitlab-org/gitlab/-/blob/v14.9.3-ee/lib/api/entities/application_setting.rb#L5
// https://gitlab.com/gitlab-org/gitlab/-/blob/v14.9.3-ee/app/helpers/application_settings_helper.rb#L192
// https://gitlab.com/gitlab-org/gitlab/-/blob/v14.9.3-ee/ee/lib/ee/api/helpers/settings_helpers.rb#L10
// https://gitlab.com/gitlab-org/gitlab/-/blob/v14.9.3-ee/ee/app/helpers/ee/application_settings_helper.rb#L20
type Settings struct {
	ID                                                    int               `json:"id"`
	AbuseNotificationEmail                                string            `json:"abuse_notification_email"`
	AdminMode                                             bool              `json:"admin_mode"`
	AfterSignOutPath                                      string            `json:"after_sign_out_path"`
	AfterSignUpText                                       string            `json:"after_sign_up_text"`
	AkismetAPIKey                                         string            `json:"akismet_api_key"`
	AkismetEnabled                                        bool              `json:"akismet_enabled"`
	AllowGroupOwnersToManageLDAP                          bool              `json:"allow_group_owners_to_manage_ldap"`
	AllowLocalRequestsFromSystemHooks                     bool              `json:"allow_local_requests_from_system_hooks"`
	AllowLocalRequestsFromWebHooksAndServices             bool              `json:"allow_local_requests_from_web_hooks_and_services"`
	ArchiveBuildsInHumanReadable                          string            `json:"archive_builds_in_human_readable"`
	AssetProxyAllowlist                                   []string          `json:"asset_proxy_allowlist"`
	AssetProxyEnabled                                     bool              `json:"asset_proxy_enabled"`
	AssetProxyURL                                         string            `json:"asset_proxy_url"`
	AssetProxySecretKey                                   string            `json:"asset_proxy_secret_key"`
	AuthorizedKeysEnabled                                 bool              `json:"authorized_keys_enabled"`
	AutoDevOpsDomain                                      string            `json:"auto_devops_domain"`
	AutoDevOpsEnabled                                     bool              `json:"auto_devops_enabled"`
	AutomaticPurchasedStorageAllocation                   bool              `json:"automatic_purchased_storage_allocation"`
	CanCreateGroup                                        bool              `json:"can_create_group"`
	CheckNamespacePlan                                    bool              `json:"check_namespace_plan"`
	CommitEmailHostname                                   string            `json:"commit_email_hostname"`
	ContainerExpirationPoliciesEnableHistoricEntries      bool              `json:"container_expiration_policies_enable_historic_entries"`
	ContainerRegistryCleanupTagsServiceMaxListSize        int               `json:"container_registry_cleanup_tags_service_max_list_size"`
	ContainerRegistryDeleteTagsServiceTimeout             int               `json:"container_registry_delete_tags_service_timeout"`
	ContainerRegistryExpirationPoliciesCaching            bool              `json:"container_registry_expiration_policies_caching"`
	ContainerRegistryExpirationPoliciesWorkerCapacity     int               `json:"container_registry_expiration_policies_worker_capacity"`
	ContainerRegistryImportCreatedBefore                  *time.Time        `json:"container_registry_import_created_before"`
	ContainerRegistryImportMaxRetries                     int               `json:"container_registry_import_max_retries"`
	ContainerRegistryImportMaxStepDuration                int               `json:"container_registry_import_max_step_duration"`
	ContainerRegistryImportMaxTagsCount                   int               `json:"container_registry_import_max_tags_count"`
	ContainerRegistryImportStartMaxRetries                int               `json:"container_registry_import_start_max_retries"`
	ContainerRegistryImportTargetPlan                     string            `json:"container_registry_import_target_plan"`
	ContainerRegistryTokenExpireDelay                     int               `json:"container_registry_token_expire_delay"`
	CreatedAt                                             *time.Time        `json:"created_at"`
	CustomHTTPCloneURLRoot                                string            `json:"custom_http_clone_url_root"`
	DNSRebindingProtectionEnabled                         bool              `json:"dns_rebinding_protection_enabled"`
	DSAKeyRestriction                                     int               `json:"dsa_key_restriction"`
	DeactivateDormantUsers                                bool              `json:"deactivate_dormant_users"`
	DefaultArtifactsExpireIn                              string            `json:"default_artifacts_expire_in"`
	DefaultBranchName                                     string            `json:"default_branch_name"`
	DefaultBranchProtection                               int               `json:"default_branch_protection"`
	DefaultCiConfigPath                                   string            `json:"default_ci_config_path"`
	DefaultGroupVisibility                                VisibilityValue   `json:"default_group_visibility"`
	DefaultProjectCreation                                int               `json:"default_project_creation"`
	DefaultProjectDeletionProtection                      bool              `json:"default_project_deletion_protection"`
	DefaultProjectVisibility                              VisibilityValue   `json:"default_project_visibility"`
	DefaultProjectsLimit                                  int               `json:"default_projects_limit"`
	DefaultSnippetVisibility                              VisibilityValue   `json:"default_snippet_visibility"`
	DelayedGroupDeletion                                  bool              `json:"delayed_group_deletion"`
	DelayedProjectDeletion                                bool              `json:"delayed_project_deletion"`
	DeleteInactiveProjects                                bool              `json:"delete_inactive_projects"`
	DeletionAdjournedPeriod                               int               `json:"deletion_adjourned_period"`
	DiffMaxFiles                                          int               `json:"diff_max_files"`
	DiffMaxLines                                          int               `json:"diff_max_lines"`
	DiffMaxPatchBytes                                     int               `json:"diff_max_patch_bytes"`
	DisableFeedToken                                      bool              `json:"disable_feed_token"`
	DisableOverridingApproversPerMergeRequest             bool              `json:"disable_overriding_approvers_per_merge_request"`
	DisabledOauthSignInSources                            []string          `json:"disabled_oauth_sign_in_sources"`
	DomainAllowlist                                       []string          `json:"domain_allowlist"`
	DomainDenylist                                        []string          `json:"domain_denylist"`
	DomainDenylistEnabled                                 bool              `json:"domain_denylist_enabled"`
	ECDSAKeyRestriction                                   int               `json:"ecdsa_key_restriction"`
	ECDSASKKeyRestriction                                 int               `json:"ecdsa_sk_key_restriction"`
	EKSAccessKeyID                                        string            `json:"eks_access_key_id"`
	EKSAccountID                                          string            `json:"eks_account_id"`
	EKSIntegrationEnabled                                 bool              `json:"eks_integration_enabled"`
	EKSSecretAccessKey                                    string            `json:"eks_secret_access_key"`
	Ed25519KeyRestriction                                 int               `json:"ed25519_key_restriction"`
	Ed25519SKKeyRestriction                               int               `json:"ed25519_sk_key_restriction"`
	ElasticsearchAWS                                      bool              `json:"elasticsearch_aws"`
	ElasticsearchAWSAccessKey                             string            `json:"elasticsearch_aws_access_key"`
	ElasticsearchAWSRegion                                string            `json:"elasticsearch_aws_region"`
	ElasticsearchAWSSecretAccessKey                       string            `json:"elasticsearch_aws_secret_access_key"`
	ElasticsearchAnalyzersKuromojiEnabled                 bool              `json:"elasticsearch_analyzers_kuromoji_enabled"`
	ElasticsearchAnalyzersKuromojiSearch                  bool              `json:"elasticsearch_analyzers_kuromoji_search"`
	ElasticsearchAnalyzersSmartCNEnabled                  bool              `json:"elasticsearch_analyzers_smartcn_enabled"`
	ElasticsearchAnalyzersSmartCNSearch                   bool              `json:"elasticsearch_analyzers_smartcn_search"`
	ElasticsearchClientRequestTimeout                     int               `json:"elasticsearch_client_request_timeout"`
	ElasticsearchIndexedFieldLengthLimit                  int               `json:"elasticsearch_indexed_field_length_limit"`
	ElasticsearchIndexedFileSizeLimitKB                   int               `json:"elasticsearch_indexed_file_size_limit_kb"`
	ElasticsearchIndexing                                 bool              `json:"elasticsearch_indexing"`
	ElasticsearchLimitIndexing                            bool              `json:"elasticsearch_limit_indexing"`
	ElasticsearchMaxBulkConcurrency                       int               `json:"elasticsearch_max_bulk_concurrency"`
	ElasticsearchMaxBulkSizeMB                            int               `json:"elasticsearch_max_bulk_size_mb"`
	ElasticsearchNamespaceIDs                             []int             `json:"elasticsearch_namespace_ids"`
	ElasticsearchPassword                                 string            `json:"elasticsearch_password"`
	ElasticsearchPauseIndexing                            bool              `json:"elasticsearch_pause_indexing"`
	ElasticsearchProjectIDs                               []int             `json:"elasticsearch_project_ids"`
	ElasticsearchReplicas                                 int               `json:"elasticsearch_replicas"`
	ElasticsearchSearch                                   bool              `json:"elasticsearch_search"`
	ElasticsearchShards                                   int               `json:"elasticsearch_shards"`
	ElasticsearchURL                                      []string          `json:"elasticsearch_url"`
	ElasticsearchUsername                                 string            `json:"elasticsearch_username"`
	EmailAdditionalText                                   string            `json:"email_additional_text"`
	EmailAuthorInBody                                     bool              `json:"email_author_in_body"`
	EmailRestrictions                                     string            `json:"email_restrictions"`
	EmailRestrictionsEnabled                              bool              `json:"email_restrictions_enabled"`
	EnabledGitAccessProtocol                              string            `json:"enabled_git_access_protocol"`
	EnforceNamespaceStorageLimit                          bool              `json:"enforce_namespace_storage_limit"`
	EnforcePATExpiration                                  bool              `json:"enforce_pat_expiration"`
	EnforceSSHKeyExpiration                               bool              `json:"enforce_ssh_key_expiration"`
	EnforceTerms                                          bool              `json:"enforce_terms"`
	ExternalAuthClientCert                                string            `json:"external_auth_client_cert"`
	ExternalAuthClientKey                                 string            `json:"external_auth_client_key"`
	ExternalAuthClientKeyPass                             string            `json:"external_auth_client_key_pass"`
	ExternalAuthorizationServiceDefaultLabel              string            `json:"external_authorization_service_default_label"`
	ExternalAuthorizationServiceEnabled                   bool              `json:"external_authorization_service_enabled"`
	ExternalAuthorizationServiceTimeout                   float64           `json:"external_authorization_service_timeout"`
	ExternalAuthorizationServiceURL                       string            `json:"external_authorization_service_url"`
	ExternalPipelineValidationServiceTimeout              int               `json:"external_pipeline_validation_service_timeout"`
	ExternalPipelineValidationServiceToken                string            `json:"external_pipeline_validation_service_token"`
	ExternalPipelineValidationServiceURL                  string            `json:"external_pipeline_validation_service_url"`
	FileTemplateProjectID                                 int               `json:"file_template_project_id"`
	FirstDayOfWeek                                        int               `json:"first_day_of_week"`
	FlocEnabled                                           bool              `json:"floc_enabled"`
	GeoNodeAllowedIPs                                     string            `json:"geo_node_allowed_ips"`
	GeoStatusTimeout                                      int               `json:"geo_status_timeout"`
	GitTwoFactorSessionExpiry                             int               `json:"git_two_factor_session_expiry"`
	GitalyTimeoutDefault                                  int               `json:"gitaly_timeout_default"`
	GitalyTimeoutFast                                     int               `json:"gitaly_timeout_fast"`
	GitalyTimeoutMedium                                   int               `json:"gitaly_timeout_medium"`
	GitpodEnabled                                         bool              `json:"gitpod_enabled"`
	GitpodURL                                             string            `json:"gitpod_url"`
	GitRateLimitUsersAllowlist                            []string          `json:"git_rate_limit_users_allowlist"`
	GrafanaEnabled                                        bool              `json:"grafana_enabled"`
	GrafanaURL                                            string            `json:"grafana_url"`
	GravatarEnabled                                       bool              `json:"gravatar_enabled"`
	GroupDownloadExportLimit                              int               `json:"group_download_export_limit"`
	GroupExportLimit                                      int               `json:"group_export_limit"`
	GroupImportLimit                                      int               `json:"group_import_limit"`
	GroupOwnersCanManageDefaultBranchProtection           bool              `json:"group_owners_can_manage_default_branch_protection"`
	GroupRunnerTokenExpirationInterval                    int               `json:"group_runner_token_expiration_interval"`
	HTMLEmailsEnabled                                     bool              `json:"html_emails_enabled"`
	HashedStorageEnabled                                  bool              `json:"hashed_storage_enabled"`
	HelpPageDocumentationBaseURL                          string            `json:"help_page_documentation_base_url"`
	HelpPageHideCommercialContent                         bool              `json:"help_page_hide_commercial_content"`
	HelpPageSupportURL                                    string            `json:"help_page_support_url"`
	HelpPageText                                          string            `json:"help_page_text"`
	HelpText                                              string            `json:"help_text"`
	HideThirdPartyOffers                                  bool              `json:"hide_third_party_offers"`
	HomePageURL                                           string            `json:"home_page_url"`
	HousekeepingBitmapsEnabled                            bool              `json:"housekeeping_bitmaps_enabled"`
	HousekeepingEnabled                                   bool              `json:"housekeeping_enabled"`
	HousekeepingFullRepackPeriod                          int               `json:"housekeeping_full_repack_period"`
	HousekeepingGcPeriod                                  int               `json:"housekeeping_gc_period"`
	HousekeepingIncrementalRepackPeriod                   int               `json:"housekeeping_incremental_repack_period"`
	ImportSources                                         []string          `json:"import_sources"`
	InactiveProjectsDeleteAfterMonths                     int               `json:"inactive_projects_delete_after_months"`
	InactiveProjectsMinSizeMB                             int               `json:"inactive_projects_min_size_mb"`
	InactiveProjectsSendWarningEmailAfterMonths           int               `json:"inactive_projects_send_warning_email_after_months"`
	InProductMarketingEmailsEnabled                       bool              `json:"in_product_marketing_emails_enabled"`
	InvisibleCaptchaEnabled                               bool              `json:"invisible_captcha_enabled"`
	IssuesCreateLimit                                     int               `json:"issues_create_limit"`
	KeepLatestArtifact                                    bool              `json:"keep_latest_artifact"`
	KrokiEnabled                                          bool              `json:"kroki_enabled"`
	KrokiFormats                                          map[string]bool   `json:"kroki_formats"`
	KrokiURL                                              string            `json:"kroki_url"`
	LocalMarkdownVersion                                  int               `json:"local_markdown_version"`
	LockMembershipsToLDAP                                 bool              `json:"lock_memberships_to_ldap"`
	LoginRecaptchaProtectionEnabled                       bool              `json:"login_recaptcha_protection_enabled"`
	MailgunEventsEnabled                                  bool              `json:"mailgun_events_enabled"`
	MailgunSigningKey                                     string            `json:"mailgun_signing_key"`
	MaintenanceMode                                       bool              `json:"maintenance_mode"`
	MaintenanceModeMessage                                string            `json:"maintenance_mode_message"`
	MaxArtifactsSize                                      int               `json:"max_artifacts_size"`
	MaxAttachmentSize                                     int               `json:"max_attachment_size"`
	MaxExportSize                                         int               `json:"max_export_size"`
	MaxImportSize                                         int               `json:"max_import_size"`
	MaxNumberOfRepositoryDownloads                        int               `json:"max_number_of_repository_downloads"`
	MaxNumberOfRepositoryDownloadsWithinTimePeriod        int               `json:"max_number_of_repository_downloads_within_time_period"`
	MaxPagesSize                                          int               `json:"max_pages_size"`
	MaxPersonalAccessTokenLifetime                        int               `json:"max_personal_access_token_lifetime"`
	MaxSSHKeyLifetime                                     int               `json:"max_ssh_key_lifetime"`
	MaxYAMLDepth                                          int               `json:"max_yaml_depth"`
	MaxYAMLSizeBytes                                      int               `json:"max_yaml_size_bytes"`
	MetricsMethodCallThreshold                            int               `json:"metrics_method_call_threshold"`
	MinimumPasswordLength                                 int               `json:"minimum_password_length"`
	MirrorAvailable                                       bool              `json:"mirror_available"`
	MirrorCapacityThreshold                               int               `json:"mirror_capacity_threshold"`
	MirrorMaxCapacity                                     int               `json:"mirror_max_capacity"`
	MirrorMaxDelay                                        int               `json:"mirror_max_delay"`
	NPMPackageRequestsForwarding                          bool              `json:"npm_package_requests_forwarding"`
	NotesCreateLimit                                      int               `json:"notes_create_limit"`
	NotifyOnUnknownSignIn                                 bool              `json:"notify_on_unknown_sign_in"`
	OutboundLocalRequestsAllowlistRaw                     string            `json:"outbound_local_requests_allowlist_raw"`
	OutboundLocalRequestsWhitelist                        []string          `json:"outbound_local_requests_whitelist"`
	PackageRegistryCleanupPoliciesWorkerCapacity          int               `json:"package_registry_cleanup_policies_worker_capacity"`
	PagesDomainVerificationEnabled                        bool              `json:"pages_domain_verification_enabled"`
	PasswordAuthenticationEnabledForGit                   bool              `json:"password_authentication_enabled_for_git"`
	PasswordAuthenticationEnabledForWeb                   bool              `json:"password_authentication_enabled_for_web"`
	PasswordNumberRequired                                bool              `json:"password_number_required"`
	PasswordSymbolRequired                                bool              `json:"password_symbol_required"`
	PasswordUppercaseRequired                             bool              `json:"password_uppercase_required"`
	PasswordLowercaseRequired                             bool              `json:"password_lowercase_required"`
	PerformanceBarAllowedGroupID                          string            `json:"performance_bar_allowed_group_id"`
	PerformanceBarAllowedGroupPath                        string            `json:"performance_bar_allowed_group_path"`
	PerformanceBarEnabled                                 bool              `json:"performance_bar_enabled"`
	PersonalAccessTokenPrefix                             string            `json:"personal_access_token_prefix"`
	PipelineLimitPerProjectUserSha                        int               `json:"pipeline_limit_per_project_user_sha"`
	PlantumlEnabled                                       bool              `json:"plantuml_enabled"`
	PlantumlURL                                           string            `json:"plantuml_url"`
	PollingIntervalMultiplier                             float64           `json:"polling_interval_multiplier,string"`
	PreventMergeRequestsAuthorApproval                    bool              `json:"prevent_merge_request_author_approval"`
	PreventMergeRequestsCommittersApproval                bool              `json:"prevent_merge_request_committers_approval"`
	ProjectDownloadExportLimit                            int               `json:"project_download_export_limit"`
	ProjectExportEnabled                                  bool              `json:"project_export_enabled"`
	ProjectExportLimit                                    int               `json:"project_export_limit"`
	ProjectImportLimit                                    int               `json:"project_import_limit"`
	ProjectRunnerTokenExpirationInterval                  int               `json:"project_runner_token_expiration_interval"`
	PrometheusMetricsEnabled                              bool              `json:"prometheus_metrics_enabled"`
	ProtectedCIVariables                                  bool              `json:"protected_ci_variables"`
	PseudonymizerEnabled                                  bool              `json:"pseudonymizer_enabled"`
	PushEventActivitiesLimit                              int               `json:"push_event_activities_limit"`
	PushEventHooksLimit                                   int               `json:"push_event_hooks_limit"`
	PyPIPackageRequestsForwarding                         bool              `json:"pypi_package_requests_forwarding"`
	RSAKeyRestriction                                     int               `json:"rsa_key_restriction"`
	RateLimitingResponseText                              string            `json:"rate_limiting_response_text"`
	RawBlobRequestLimit                                   int               `json:"raw_blob_request_limit"`
	RecaptchaEnabled                                      bool              `json:"recaptcha_enabled"`
	RecaptchaPrivateKey                                   string            `json:"recaptcha_private_key"`
	RecaptchaSiteKey                                      string            `json:"recaptcha_site_key"`
	ReceiveMaxInputSize                                   int               `json:"receive_max_input_size"`
	RepositoryChecksEnabled                               bool              `json:"repository_checks_enabled"`
	RepositorySizeLimit                                   int               `json:"repository_size_limit"`
	RepositoryStorages                                    []string          `json:"repository_storages"`
	RepositoryStoragesWeighted                            map[string]int    `json:"repository_storages_weighted"`
	RequireAdminApprovalAfterUserSignup                   bool              `json:"require_admin_approval_after_user_signup"`
	RequireTwoFactorAuthentication                        bool              `json:"require_two_factor_authentication"`
	RestrictedVisibilityLevels                            []VisibilityValue `json:"restricted_visibility_levels"`
	RunnerTokenExpirationInterval                         int               `json:"runner_token_expiration_interval"`
	SearchRateLimit                                       int               `json:"search_rate_limit"`
	SearchRateLimitUnauthenticated                        int               `json:"search_rate_limit_unauthenticated"`
	SecretDetectionRevocationTokenTypesURL                string            `json:"secret_detection_revocation_token_types_url"`
	SecretDetectionTokenRevocationEnabled                 bool              `json:"secret_detection_token_revocation_enabled"`
	SecretDetectionTokenRevocationToken                   string            `json:"secret_detection_token_revocation_token"`
	SecretDetectionTokenRevocationURL                     string            `json:"secret_detection_token_revocation_url"`
	SendUserConfirmationEmail                             bool              `json:"send_user_confirmation_email"`
	SentryClientsideDSN                                   string            `json:"sentry_clientside_dsn"`
	SentryDSN                                             string            `json:"sentry_dsn"`
	SentryEnabled                                         bool              `json:"sentry_enabled"`
	SentryEnvironment                                     string            `json:"sentry_environment"`
	SessionExpireDelay                                    int               `json:"session_expire_delay"`
	SharedRunnersEnabled                                  bool              `json:"shared_runners_enabled"`
	SharedRunnersMinutes                                  int               `json:"shared_runners_minutes"`
	SharedRunnersText                                     string            `json:"shared_runners_text"`
	SidekiqJobLimiterCompressionThresholdBytes            int               `json:"sidekiq_job_limiter_compression_threshold_bytes"`
	SidekiqJobLimiterLimitBytes                           int               `json:"sidekiq_job_limiter_limit_bytes"`
	SidekiqJobLimiterMode                                 string            `json:"sidekiq_job_limiter_mode"`
	SignInText                                            string            `json:"sign_in_text"`
	SignupEnabled                                         bool              `json:"signup_enabled"`
	SlackAppEnabled                                       bool              `json:"slack_app_enabled"`
	SlackAppID                                            string            `json:"slack_app_id"`
	SlackAppSecret                                        string            `json:"slack_app_secret"`
	SlackAppSigningSecret                                 string            `json:"slack_app_signing_secret"`
	SlackAppVerificationToken                             string            `json:"slack_app_verification_token"`
	SnippetSizeLimit                                      int               `json:"snippet_size_limit"`
	SnowplowAppID                                         string            `json:"snowplow_app_id"`
	SnowplowCollectorHostname                             string            `json:"snowplow_collector_hostname"`
	SnowplowCookieDomain                                  string            `json:"snowplow_cookie_domain"`
	SnowplowEnabled                                       bool              `json:"snowplow_enabled"`
	SourcegraphEnabled                                    bool              `json:"sourcegraph_enabled"`
	SourcegraphPublicOnly                                 bool              `json:"sourcegraph_public_only"`
	SourcegraphURL                                        string            `json:"sourcegraph_url"`
	SpamCheckAPIKey                                       string            `json:"spam_check_api_key"`
	SpamCheckEndpointEnabled                              bool              `json:"spam_check_endpoint_enabled"`
	SpamCheckEndpointURL                                  string            `json:"spam_check_endpoint_url"`
	SuggestPipelineEnabled                                bool              `json:"suggest_pipeline_enabled"`
	TerminalMaxSessionTime                                int               `json:"terminal_max_session_time"`
	Terms                                                 string            `json:"terms"`
	ThrottleAuthenticatedAPIEnabled                       bool              `json:"throttle_authenticated_api_enabled"`
	ThrottleAuthenticatedAPIPeriodInSeconds               int               `json:"throttle_authenticated_api_period_in_seconds"`
	ThrottleAuthenticatedAPIRequestsPerPeriod             int               `json:"throttle_authenticated_api_requests_per_period"`
	ThrottleAuthenticatedDeprecatedAPIEnabled             bool              `json:"throttle_authenticated_deprecated_api_enabled"`
	ThrottleAuthenticatedDeprecatedAPIPeriodInSeconds     int               `json:"throttle_authenticated_deprecated_api_period_in_seconds"`
	ThrottleAuthenticatedDeprecatedAPIRequestsPerPeriod   int               `json:"throttle_authenticated_deprecated_api_requests_per_period"`
	ThrottleAuthenticatedFilesAPIEnabled                  bool              `json:"throttle_authenticated_files_api_enabled"`
	ThrottleAuthenticatedFilesAPIPeriodInSeconds          int               `json:"throttle_authenticated_files_api_period_in_seconds"`
	ThrottleAuthenticatedFilesAPIRequestsPerPeriod        int               `json:"throttle_authenticated_files_api_requests_per_period"`
	ThrottleAuthenticatedGitLFSEnabled                    bool              `json:"throttle_authenticated_git_lfs_enabled"`
	ThrottleAuthenticatedGitLFSPeriodInSeconds            int               `json:"throttle_authenticated_git_lfs_period_in_seconds"`
	ThrottleAuthenticatedGitLFSRequestsPerPeriod          int               `json:"throttle_authenticated_git_lfs_requests_per_period"`
	ThrottleAuthenticatedPackagesAPIEnabled               bool              `json:"throttle_authenticated_packages_api_enabled"`
	ThrottleAuthenticatedPackagesAPIPeriodInSeconds       int               `json:"throttle_authenticated_packages_api_period_in_seconds"`
	ThrottleAuthenticatedPackagesAPIRequestsPerPeriod     int               `json:"throttle_authenticated_packages_api_requests_per_period"`
	ThrottleAuthenticatedWebEnabled                       bool              `json:"throttle_authenticated_web_enabled"`
	ThrottleAuthenticatedWebPeriodInSeconds               int               `json:"throttle_authenticated_web_period_in_seconds"`
	ThrottleAuthenticatedWebRequestsPerPeriod             int               `json:"throttle_authenticated_web_requests_per_period"`
	ThrottleIncidentManagementNotificationEnabled         bool              `json:"throttle_incident_management_notification_enabled"`
	ThrottleIncidentManagementNotificationPerPeriod       int               `json:"throttle_incident_management_notification_per_period"`
	ThrottleIncidentManagementNotificationPeriodInSeconds int               `json:"throttle_incident_management_notification_period_in_seconds"`
	ThrottleProtectedPathsEnabled                         bool              `json:"throttle_protected_paths_enabled"`
	ThrottleProtectedPathsPeriodInSeconds                 int               `json:"throttle_protected_paths_period_in_seconds"`
	ThrottleProtectedPathsRequestsPerPeriod               int               `json:"throttle_protected_paths_requests_per_period"`
	ThrottleUnauthenticatedAPIEnabled                     bool              `json:"throttle_unauthenticated_api_enabled"`
	ThrottleUnauthenticatedAPIPeriodInSeconds             int               `json:"throttle_unauthenticated_api_period_in_seconds"`
	ThrottleUnauthenticatedAPIRequestsPerPeriod           int               `json:"throttle_unauthenticated_api_requests_per_period"`
	ThrottleUnauthenticatedDeprecatedAPIEnabled           bool              `json:"throttle_unauthenticated_deprecated_api_enabled"`
	ThrottleUnauthenticatedDeprecatedAPIPeriodInSeconds   int               `json:"throttle_unauthenticated_deprecated_api_period_in_seconds"`
	ThrottleUnauthenticatedDeprecatedAPIRequestsPerPeriod int               `json:"throttle_unauthenticated_deprecated_api_requests_per_period"`
	ThrottleUnauthenticatedFilesAPIEnabled                bool              `json:"throttle_unauthenticated_files_api_enabled"`
	ThrottleUnauthenticatedFilesAPIPeriodInSeconds        int               `json:"throttle_unauthenticated_files_api_period_in_seconds"`
	ThrottleUnauthenticatedFilesAPIRequestsPerPeriod      int               `json:"throttle_unauthenticated_files_api_requests_per_period"`
	ThrottleUnauthenticatedGitLFSEnabled                  bool              `json:"throttle_unauthenticated_git_lfs_enabled"`
	ThrottleUnauthenticatedGitLFSPeriodInSeconds          int               `json:"throttle_unauthenticated_git_lfs_period_in_seconds"`
	ThrottleUnauthenticatedGitLFSRequestsPerPeriod        int               `json:"throttle_unauthenticated_git_lfs_requests_per_period"`
	ThrottleUnauthenticatedPackagesAPIEnabled             bool              `json:"throttle_unauthenticated_packages_api_enabled"`
	ThrottleUnauthenticatedPackagesAPIPeriodInSeconds     int               `json:"throttle_unauthenticated_packages_api_period_in_seconds"`
	ThrottleUnauthenticatedPackagesAPIRequestsPerPeriod   int               `json:"throttle_unauthenticated_packages_api_requests_per_period"`
	ThrottleUnauthenticatedWebEnabled                     bool              `json:"throttle_unauthenticated_web_enabled"`
	ThrottleUnauthenticatedWebPeriodInSeconds             int               `json:"throttle_unauthenticated_web_period_in_seconds"`
	ThrottleUnauthenticatedWebRequestsPerPeriod           int               `json:"throttle_unauthenticated_web_requests_per_period"`
	TimeTrackingLimitToHours                              bool              `json:"time_tracking_limit_to_hours"`
	TwoFactorGracePeriod                                  int               `json:"two_factor_grace_period"`
	UniqueIPsLimitEnabled                                 bool              `json:"unique_ips_limit_enabled"`
	UniqueIPsLimitPerUser                                 int               `json:"unique_ips_limit_per_user"`
	UniqueIPsLimitTimeWindow                              int               `json:"unique_ips_limit_time_window"`
	UpdatedAt                                             *time.Time        `json:"updated_at"`
	UpdatingNameDisabledForUsers                          bool              `json:"updating_name_disabled_for_users"`
	UsagePingEnabled                                      bool              `json:"usage_ping_enabled"`
	UsagePingFeaturesEnabled                              bool              `json:"usage_ping_features_enabled"`
	UserDeactivationEmailsEnabled                         bool              `json:"user_deactivation_emails_enabled"`
	UserDefaultExternal                                   bool              `json:"user_default_external"`
	UserDefaultInternalRegex                              string            `json:"user_default_internal_regex"`
	UserOauthApplications                                 bool              `json:"user_oauth_applications"`
	UserShowAddSSHKeyMessage                              bool              `json:"user_show_add_ssh_key_message"`
	UsersGetByIDLimit                                     int               `json:"users_get_by_id_limit"`
	UsersGetByIDLimitAllowlistRaw                         string            `json:"users_get_by_id_limit_allowlist_raw"`
	VersionCheckEnabled                                   bool              `json:"version_check_enabled"`
	WebIDEClientsidePreviewEnabled                        bool              `json:"web_ide_clientside_preview_enabled"`
	WhatsNewVariant                                       string            `json:"whats_new_variant"`
	WikiPageMaxContentBytes                               int               `json:"wiki_page_max_content_bytes"`

	// Deprecated: Use AbuseNotificationEmail instead.
	AdminNotificationEmail string `json:"admin_notification_email"`
	// Deprecated: Use AllowLocalRequestsFromWebHooksAndServices instead.
	AllowLocalRequestsFromHooksAndServices bool `json:"allow_local_requests_from_hooks_and_services"`
	// Deprecated: Use AssetProxyAllowlist instead.
	AssetProxyWhitelist []string `json:"asset_proxy_whitelist"`
	// Deprecated: Use ThrottleUnauthenticatedWebEnabled or ThrottleUnauthenticatedAPIEnabled instead. (Deprecated in GitLab 14.3)
	ThrottleUnauthenticatedEnabled bool `json:"throttle_unauthenticated_enabled"`
	// Deprecated: Use ThrottleUnauthenticatedWebPeriodInSeconds or ThrottleUnauthenticatedAPIPeriodInSeconds instead. (Deprecated in GitLab 14.3)
	ThrottleUnauthenticatedPeriodInSeconds int `json:"throttle_unauthenticated_period_in_seconds"`
	// Deprecated: Use ThrottleUnauthenticatedWebRequestsPerPeriod or ThrottleUnauthenticatedAPIRequestsPerPeriod instead. (Deprecated in GitLab 14.3)
	ThrottleUnauthenticatedRequestsPerPeriod int `json:"throttle_unauthenticated_requests_per_period"`
	// Deprecated: Replaced by SearchRateLimit in GitLab 14.9 (removed in 15.0).
	UserEmailLookupLimit int `json:"user_email_lookup_limit"`
}

func (s Settings) String() string {
	return Stringify(s)
}

// GetSettings gets the current application settings.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/settings.html#get-current-application-settings
func (s *SettingsService) GetSettings(options ...RequestOptionFunc) (*Settings, *Response, error) {
	req, err := s.client.NewRequest(http.MethodGet, "application/settings", nil, options)
	if err != nil {
		return nil, nil, err
	}

	as := new(Settings)
	resp, err := s.client.Do(req, as)
	if err != nil {
		return nil, resp, err
	}

	return as, resp, nil
}

// UpdateSettingsOptions represents the available UpdateSettings() options.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/settings.html#change-application-settings
type UpdateSettingsOptions struct {
	AbuseNotificationEmail                                *string            `url:"abuse_notification_email,omitempty" json:"abuse_notification_email,omitempty"`
	AdminMode                                             *bool              `url:"admin_mode,omitempty" json:"admin_mode,omitempty"`
	AdminNotificationEmail                                *string            `url:"admin_notification_email,omitempty" json:"admin_notification_email,omitempty"`
	AfterSignOutPath                                      *string            `url:"after_sign_out_path,omitempty" json:"after_sign_out_path,omitempty"`
	AfterSignUpText                                       *string            `url:"after_sign_up_text,omitempty" json:"after_sign_up_text,omitempty"`
	AkismetAPIKey                                         *string            `url:"akismet_api_key,omitempty" json:"akismet_api_key,omitempty"`
	AkismetEnabled                                        *bool              `url:"akismet_enabled,omitempty" json:"akismet_enabled,omitempty"`
	AllowGroupOwnersToManageLDAP                          *bool              `url:"allow_group_owners_to_manage_ldap,omitempty" json:"allow_group_owners_to_manage_ldap,omitempty"`
	AllowLocalRequestsFromHooksAndServices                *bool              `url:"allow_local_requests_from_hooks_and_services,omitempty" json:"allow_local_requests_from_hooks_and_services,omitempty"`
	AllowLocalRequestsFromSystemHooks                     *bool              `url:"allow_local_requests_from_system_hooks,omitempty" json:"allow_local_requests_from_system_hooks,omitempty"`
	AllowLocalRequestsFromWebHooksAndServices             *bool              `url:"allow_local_requests_from_web_hooks_and_services,omitempty" json:"allow_local_requests_from_web_hooks_and_services,omitempty"`
	ArchiveBuildsInHumanReadable                          *string            `url:"archive_builds_in_human_readable,omitempty" json:"archive_builds_in_human_readable,omitempty"`
	AssetProxyAllowlist                                   *[]string          `url:"asset_proxy_allowlist,omitempty" json:"asset_proxy_allowlist,omitempty"`
	AssetProxyEnabled                                     *bool              `url:"asset_proxy_enabled,omitempty" json:"asset_proxy_enabled,omitempty"`
	AssetProxySecretKey                                   *string            `url:"asset_proxy_secret_key,omitempty" json:"asset_proxy_secret_key,omitempty"`
	AssetProxyURL                                         *string            `url:"asset_proxy_url,omitempty" json:"asset_proxy_url,omitempty"`
	AssetProxyWhitelist                                   *[]string          `url:"asset_proxy_whitelist,omitempty" json:"asset_proxy_whitelist,omitempty"`
	AuthorizedKeysEnabled                                 *bool              `url:"authorized_keys_enabled,omitempty" json:"authorized_keys_enabled,omitempty"`
	AutoDevOpsDomain                                      *string            `url:"auto_devops_domain,omitempty" json:"auto_devops_domain,omitempty"`
	AutoDevOpsEnabled                                     *bool              `url:"auto_devops_enabled,omitempty" json:"auto_devops_enabled,omitempty"`
	AutomaticPurchasedStorageAllocation                   *bool              `url:"automatic_purchased_storage_allocation,omitempty" json:"automatic_purchased_storage_allocation,omitempty"`
	CanCreateGroup                                        *bool              `url:"can_create_group,omitempty" json:"can_create_group,omitempty"`
	CheckNamespacePlan                                    *bool              `url:"check_namespace_plan,omitempty" json:"check_namespace_plan,omitempty"`
	CommitEmailHostname                                   *string            `url:"commit_email_hostname,omitempty" json:"commit_email_hostname,omitempty"`
	ContainerExpirationPoliciesEnableHistoricEntries      *bool              `url:"container_expiration_policies_enable_historic_entries,omitempty" json:"container_expiration_policies_enable_historic_entries,omitempty"`
	ContainerRegistryCleanupTagsServiceMaxListSize        *int               `url:"container_registry_cleanup_tags_service_max_list_size,omitempty" json:"container_registry_cleanup_tags_service_max_list_size,omitempty"`
	ContainerRegistryDeleteTagsServiceTimeout             *int               `url:"container_registry_delete_tags_service_timeout,omitempty" json:"container_registry_delete_tags_service_timeout,omitempty"`
	ContainerRegistryExpirationPoliciesCaching            *bool              `url:"container_registry_expiration_policies_caching,omitempty" json:"container_registry_expiration_policies_caching,omitempty"`
	ContainerRegistryExpirationPoliciesWorkerCapacity     *int               `url:"container_registry_expiration_policies_worker_capacity,omitempty" json:"container_registry_expiration_policies_worker_capacity,omitempty"`
	ContainerRegistryImportCreatedBefore                  *time.Time         `url:"container_registry_import_created_before,omitempty" json:"container_registry_import_created_before,omitempty"`
	ContainerRegistryImportMaxRetries                     *int               `url:"container_registry_import_max_retries,omitempty" json:"container_registry_import_max_retries,omitempty"`
	ContainerRegistryImportMaxStepDuration                *int               `url:"container_registry_import_max_step_duration,omitempty" json:"container_registry_import_max_step_duration,omitempty"`
	ContainerRegistryImportMaxTagsCount                   *int               `url:"container_registry_import_max_tags_count,omitempty" json:"container_registry_import_max_tags_count,omitempty"`
	ContainerRegistryImportStartMaxRetries                *int               `url:"container_registry_import_start_max_retries,omitempty" json:"container_registry_import_start_max_retries,omitempty"`
	ContainerRegistryImportTargetPlan                     *string            `url:"container_registry_import_target_plan,omitempty" json:"container_registry_import_target_plan,omitempty"`
	ContainerRegistryTokenExpireDelay                     *int               `url:"container_registry_token_expire_delay,omitempty" json:"container_registry_token_expire_delay,omitempty"`
	CustomHTTPCloneURLRoot                                *string            `url:"custom_http_clone_url_root,omitempty" json:"custom_http_clone_url_root,omitempty"`
	DNSRebindingProtectionEnabled                         *bool              `url:"dns_rebinding_protection_enabled,omitempty" json:"dns_rebinding_protection_enabled,omitempty"`
	DSAKeyRestriction                                     *int               `url:"dsa_key_restriction,omitempty" json:"dsa_key_restriction,omitempty"`
	DeactivateDormantUsers                                *bool              `url:"deactivate_dormant_users,omitempty" json:"deactivate_dormant_users,omitempty"`
	DefaultArtifactsExpireIn                              *string            `url:"default_artifacts_expire_in,omitempty" json:"default_artifacts_expire_in,omitempty"`
	DefaultBranchName                                     *string            `url:"default_branch_name,omitempty" json:"default_branch_name,omitempty"`
	DefaultBranchProtection                               *int               `url:"default_branch_protection,omitempty" json:"default_branch_protection,omitempty"`
	DefaultCiConfigPath                                   *string            `url:"default_ci_config_path,omitempty" json:"default_ci_config_path,omitempty"`
	DefaultGroupVisibility                                *VisibilityValue   `url:"default_group_visibility,omitempty" json:"default_group_visibility,omitempty"`
	DefaultProjectCreation                                *int               `url:"default_project_creation,omitempty" json:"default_project_creation,omitempty"`
	DefaultProjectDeletionProtection                      *bool              `url:"default_project_deletion_protection,omitempty" json:"default_project_deletion_protection,omitempty"`
	DefaultProjectVisibility                              *VisibilityValue   `url:"default_project_visibility,omitempty" json:"default_project_visibility,omitempty"`
	DefaultProjectsLimit                                  *int               `url:"default_projects_limit,omitempty" json:"default_projects_limit,omitempty"`
	DefaultSnippetVisibility                              *VisibilityValue   `url:"default_snippet_visibility,omitempty" json:"default_snippet_visibility,omitempty"`
	DelayedGroupDeletion                                  *bool              `url:"delayed_group_deletion,omitempty" json:"delayed_group_deletion,omitempty"`
	DelayedProjectDeletion                                *bool              `url:"delayed_project_deletion,omitempty" json:"delayed_project_deletion,omitempty"`
	DeleteInactiveProjects                                *bool              `url:"delete_inactive_projects,omitempty" json:"delete_inactive_projects,omitempty"`
	DeletionAdjournedPeriod                               *int               `url:"deletion_adjourned_period,omitempty" json:"deletion_adjourned_period,omitempty"`
	DiffMaxFiles                                          *int               `url:"diff_max_files,omitempty" json:"diff_max_files,omitempty"`
	DiffMaxLines                                          *int               `url:"diff_max_lines,omitempty" json:"diff_max_lines,omitempty"`
	DiffMaxPatchBytes                                     *int               `url:"diff_max_patch_bytes,omitempty" json:"diff_max_patch_bytes,omitempty"`
	DisableFeedToken                                      *bool              `url:"disable_feed_token,omitempty" json:"disable_feed_token,omitempty"`
	DisableOverridingApproversPerMergeRequest             *bool              `url:"disable_overriding_approvers_per_merge_request,omitempty" json:"disable_overriding_approvers_per_merge_request,omitempty"`
	DisabledOauthSignInSources                            *[]string          `url:"disabled_oauth_sign_in_sources,omitempty" json:"disabled_oauth_sign_in_sources,omitempty"`
	DomainAllowlist                                       *[]string          `url:"domain_allowlist,omitempty" json:"domain_allowlist,omitempty"`
	DomainDenylist                                        *[]string          `url:"domain_denylist,omitempty" json:"domain_denylist,omitempty"`
	DomainDenylistEnabled                                 *bool              `url:"domain_denylist_enabled,omitempty" json:"domain_denylist_enabled,omitempty"`
	ECDSAKeyRestriction                                   *int               `url:"ecdsa_key_restriction,omitempty" json:"ecdsa_key_restriction,omitempty"`
	ECDSASKKeyRestriction                                 *int               `url:"ecdsa_sk_key_restriction,omitempty" json:"ecdsa_sk_key_restriction,omitempty"`
	EKSAccessKeyID                                        *string            `url:"eks_access_key_id,omitempty" json:"eks_access_key_id,omitempty"`
	EKSAccountID                                          *string            `url:"eks_account_id,omitempty" json:"eks_account_id,omitempty"`
	EKSIntegrationEnabled                                 *bool              `url:"eks_integration_enabled,omitempty" json:"eks_integration_enabled,omitempty"`
	EKSSecretAccessKey                                    *string            `url:"eks_secret_access_key,omitempty" json:"eks_secret_access_key,omitempty"`
	Ed25519KeyRestriction                                 *int               `url:"ed25519_key_restriction,omitempty" json:"ed25519_key_restriction,omitempty"`
	Ed25519SKKeyRestriction                               *int               `url:"ed25519_sk_key_restriction,omitempty" json:"ed25519_sk_key_restriction,omitempty"`
	ElasticsearchAWS                                      *bool              `url:"elasticsearch_aws,omitempty" json:"elasticsearch_aws,omitempty"`
	ElasticsearchAWSAccessKey                             *string            `url:"elasticsearch_aws_access_key,omitempty" json:"elasticsearch_aws_access_key,omitempty"`
	ElasticsearchAWSRegion                                *string            `url:"elasticsearch_aws_region,omitempty" json:"elasticsearch_aws_region,omitempty"`
	ElasticsearchAWSSecretAccessKey                       *string            `url:"elasticsearch_aws_secret_access_key,omitempty" json:"elasticsearch_aws_secret_access_key,omitempty"`
	ElasticsearchAnalyzersKuromojiEnabled                 *bool              `url:"elasticsearch_analyzers_kuromoji_enabled,omitempty" json:"elasticsearch_analyzers_kuromoji_enabled,omitempty"`
	ElasticsearchAnalyzersKuromojiSearch                  *int               `url:"elasticsearch_analyzers_kuromoji_search,omitempty" json:"elasticsearch_analyzers_kuromoji_search,omitempty"`
	ElasticsearchAnalyzersSmartCNEnabled                  *bool              `url:"elasticsearch_analyzers_smartcn_enabled,omitempty" json:"elasticsearch_analyzers_smartcn_enabled,omitempty"`
	ElasticsearchAnalyzersSmartCNSearch                   *int               `url:"elasticsearch_analyzers_smartcn_search,omitempty" json:"elasticsearch_analyzers_smartcn_search,omitempty"`
	ElasticsearchClientRequestTimeout                     *int               `url:"elasticsearch_client_request_timeout,omitempty" json:"elasticsearch_client_request_timeout,omitempty"`
	ElasticsearchIndexedFieldLengthLimit                  *int               `url:"elasticsearch_indexed_field_length_limit,omitempty" json:"elasticsearch_indexed_field_length_limit,omitempty"`
	ElasticsearchIndexedFileSizeLimitKB                   *int               `url:"elasticsearch_indexed_file_size_limit_kb,omitempty" json:"elasticsearch_indexed_file_size_limit_kb,omitempty"`
	ElasticsearchIndexing                                 *bool              `url:"elasticsearch_indexing,omitempty" json:"elasticsearch_indexing,omitempty"`
	ElasticsearchLimitIndexing                            *bool              `url:"elasticsearch_limit_indexing,omitempty" json:"elasticsearch_limit_indexing,omitempty"`
	ElasticsearchMaxBulkConcurrency                       *int               `url:"elasticsearch_max_bulk_concurrency,omitempty" json:"elasticsearch_max_bulk_concurrency,omitempty"`
	ElasticsearchMaxBulkSizeMB                            *int               `url:"elasticsearch_max_bulk_size_mb,omitempty" json:"elasticsearch_max_bulk_size_mb,omitempty"`
	ElasticsearchNamespaceIDs                             *[]int             `url:"elasticsearch_namespace_ids,omitempty" json:"elasticsearch_namespace_ids,omitempty"`
	ElasticsearchPassword                                 *string            `url:"elasticsearch_password,omitempty" json:"elasticsearch_password,omitempty"`
	ElasticsearchPauseIndexing                            *bool              `url:"elasticsearch_pause_indexing,omitempty" json:"elasticsearch_pause_indexing,omitempty"`
	ElasticsearchProjectIDs                               *[]int             `url:"elasticsearch_project_ids,omitempty" json:"elasticsearch_project_ids,omitempty"`
	ElasticsearchReplicas                                 *int               `url:"elasticsearch_replicas,omitempty" json:"elasticsearch_replicas,omitempty"`
	ElasticsearchSearch                                   *bool              `url:"elasticsearch_search,omitempty" json:"elasticsearch_search,omitempty"`
	ElasticsearchShards                                   *int               `url:"elasticsearch_shards,omitempty" json:"elasticsearch_shards,omitempty"`
	ElasticsearchURL                                      *string            `url:"elasticsearch_url,omitempty" json:"elasticsearch_url,omitempty"`
	ElasticsearchUsername                                 *string            `url:"elasticsearch_username,omitempty" json:"elasticsearch_username,omitempty"`
	EmailAdditionalText                                   *string            `url:"email_additional_text,omitempty" json:"email_additional_text,omitempty"`
	EmailAuthorInBody                                     *bool              `url:"email_author_in_body,omitempty" json:"email_author_in_body,omitempty"`
	EmailRestrictions                                     *string            `url:"email_restrictions,omitempty" json:"email_restrictions,omitempty"`
	EmailRestrictionsEnabled                              *bool              `url:"email_restrictions_enabled,omitempty" json:"email_restrictions_enabled,omitempty"`
	EnabledGitAccessProtocol                              *string            `url:"enabled_git_access_protocol,omitempty" json:"enabled_git_access_protocol,omitempty"`
	EnforceNamespaceStorageLimit                          *bool              `url:"enforce_namespace_storage_limit,omitempty" json:"enforce_namespace_storage_limit,omitempty"`
	EnforcePATExpiration                                  *bool              `url:"enforce_pat_expiration,omitempty" json:"enforce_pat_expiration,omitempty"`
	EnforceSSHKeyExpiration                               *bool              `url:"enforce_ssh_key_expiration,omitempty" json:"enforce_ssh_key_expiration,omitempty"`
	EnforceTerms                                          *bool              `url:"enforce_terms,omitempty" json:"enforce_terms,omitempty"`
	ExternalAuthClientCert                                *string            `url:"external_auth_client_cert,omitempty" json:"external_auth_client_cert,omitempty"`
	ExternalAuthClientKey                                 *string            `url:"external_auth_client_key,omitempty" json:"external_auth_client_key,omitempty"`
	ExternalAuthClientKeyPass                             *string            `url:"external_auth_client_key_pass,omitempty" json:"external_auth_client_key_pass,omitempty"`
	ExternalAuthorizationServiceDefaultLabel              *string            `url:"external_authorization_service_default_label,omitempty" json:"external_authorization_service_default_label,omitempty"`
	ExternalAuthorizationServiceEnabled                   *bool              `url:"external_authorization_service_enabled,omitempty" json:"external_authorization_service_enabled,omitempty"`
	ExternalAuthorizationServiceTimeout                   *float64           `url:"external_authorization_service_timeout,omitempty" json:"external_authorization_service_timeout,omitempty"`
	ExternalAuthorizationServiceURL                       *string            `url:"external_authorization_service_url,omitempty" json:"external_authorization_service_url,omitempty"`
	ExternalPipelineValidationServiceTimeout              *int               `url:"external_pipeline_validation_service_timeout,omitempty" json:"external_pipeline_validation_service_timeout,omitempty"`
	ExternalPipelineValidationServiceToken                *string            `url:"external_pipeline_validation_service_token,omitempty" json:"external_pipeline_validation_service_token,omitempty"`
	ExternalPipelineValidationServiceURL                  *string            `url:"external_pipeline_validation_service_url,omitempty" json:"external_pipeline_validation_service_url,omitempty"`
	FileTemplateProjectID                                 *int               `url:"file_template_project_id,omitempty" json:"file_template_project_id,omitempty"`
	FirstDayOfWeek                                        *int               `url:"first_day_of_week,omitempty" json:"first_day_of_week,omitempty"`
	FlocEnabled                                           *bool              `url:"floc_enabled,omitempty" json:"floc_enabled,omitempty"`
	GeoNodeAllowedIPs                                     *string            `url:"geo_node_allowed_ips,omitempty" json:"geo_node_allowed_ips,omitempty"`
	GeoStatusTimeout                                      *int               `url:"geo_status_timeout,omitempty" json:"geo_status_timeout,omitempty"`
	GitTwoFactorSessionExpiry                             *int               `url:"git_two_factor_session_expiry,omitempty" json:"git_two_factor_session_expiry,omitempty"`
	GitalyTimeoutDefault                                  *int               `url:"gitaly_timeout_default,omitempty" json:"gitaly_timeout_default,omitempty"`
	GitalyTimeoutFast                                     *int               `url:"gitaly_timeout_fast,omitempty" json:"gitaly_timeout_fast,omitempty"`
	GitalyTimeoutMedium                                   *int               `url:"gitaly_timeout_medium,omitempty" json:"gitaly_timeout_medium,omitempty"`
	GitpodEnabled                                         *bool              `url:"gitpod_enabled,omitempty" json:"gitpod_enabled,omitempty"`
	GitpodURL                                             *string            `url:"gitpod_url,omitempty" json:"gitpod_url,omitempty"`
	GitRateLimitUsersAllowlist                            *[]string          `url:"git_rate_limit_users_allowlist,omitempty" json:"git_rate_limit_users_allowlist,omitempty"`
	GrafanaEnabled                                        *bool              `url:"grafana_enabled,omitempty" json:"grafana_enabled,omitempty"`
	GrafanaURL                                            *string            `url:"grafana_url,omitempty" json:"grafana_url,omitempty"`
	GravatarEnabled                                       *bool              `url:"gravatar_enabled,omitempty" json:"gravatar_enabled,omitempty"`
	GroupDownloadExportLimit                              *int               `url:"group_download_export_limit,omitempty" json:"group_download_export_limit,omitempty"`
	GroupExportLimit                                      *int               `url:"group_export_limit,omitempty" json:"group_export_limit,omitempty"`
	GroupImportLimit                                      *int               `url:"group_import_limit,omitempty" json:"group_import_limit,omitempty"`
	GroupOwnersCanManageDefaultBranchProtection           *bool              `url:"group_owners_can_manage_default_branch_protection,omitempty" json:"group_owners_can_manage_default_branch_protection,omitempty"`
	GroupRunnerTokenExpirationInterval                    *int               `url:"group_runner_token_expiration_interval,omitempty" json:"group_runner_token_expiration_interval,omitempty"`
	HTMLEmailsEnabled                                     *bool              `url:"html_emails_enabled,omitempty" json:"html_emails_enabled,omitempty"`
	HashedStorageEnabled                                  *bool              `url:"hashed_storage_enabled,omitempty" json:"hashed_storage_enabled,omitempty"`
	HelpPageDocumentationBaseURL                          *string            `url:"help_page_documentation_base_url,omitempty" json:"help_page_documentation_base_url,omitempty"`
	HelpPageHideCommercialContent                         *bool              `url:"help_page_hide_commercial_content,omitempty" json:"help_page_hide_commercial_content,omitempty"`
	HelpPageSupportURL                                    *string            `url:"help_page_support_url,omitempty" json:"help_page_support_url,omitempty"`
	HelpPageText                                          *string            `url:"help_page_text,omitempty" json:"help_page_text,omitempty"`
	HelpText                                              *string            `url:"help_text,omitempty" json:"help_text,omitempty"`
	HideThirdPartyOffers                                  *bool              `url:"hide_third_party_offers,omitempty" json:"hide_third_party_offers,omitempty"`
	HomePageURL                                           *string            `url:"home_page_url,omitempty" json:"home_page_url,omitempty"`
	HousekeepingBitmapsEnabled                            *bool              `url:"housekeeping_bitmaps_enabled,omitempty" json:"housekeeping_bitmaps_enabled,omitempty"`
	HousekeepingEnabled                                   *bool              `url:"housekeeping_enabled,omitempty" json:"housekeeping_enabled,omitempty"`
	HousekeepingFullRepackPeriod                          *int               `url:"housekeeping_full_repack_period,omitempty" json:"housekeeping_full_repack_period,omitempty"`
	HousekeepingGcPeriod                                  *int               `url:"housekeeping_gc_period,omitempty" json:"housekeeping_gc_period,omitempty"`
	HousekeepingIncrementalRepackPeriod                   *int               `url:"housekeeping_incremental_repack_period,omitempty" json:"housekeeping_incremental_repack_period,omitempty"`
	ImportSources                                         *[]string          `url:"import_sources,omitempty" json:"import_sources,omitempty"`
	InactiveProjectsDeleteAfterMonths                     *int               `url:"inactive_projects_delete_after_months,omitempty" json:"inactive_projects_delete_after_months,omitempty"`
	InactiveProjectsMinSizeMB                             *int               `url:"inactive_projects_min_size_mb,omitempty" json:"inactive_projects_min_size_mb,omitempty"`
	InactiveProjectsSendWarningEmailAfterMonths           *int               `url:"inactive_projects_send_warning_email_after_months,omitempty" json:"inactive_projects_send_warning_email_after_months,omitempty"`
	InProductMarketingEmailsEnabled                       *bool              `url:"in_product_marketing_emails_enabled,omitempty" json:"in_product_marketing_emails_enabled,omitempty"`
	InvisibleCaptchaEnabled                               *bool              `url:"invisible_captcha_enabled,omitempty" json:"invisible_captcha_enabled,omitempty"`
	IssuesCreateLimit                                     *int               `url:"issues_create_limit,omitempty" json:"issues_create_limit,omitempty"`
	KeepLatestArtifact                                    *bool              `url:"keep_latest_artifact,omitempty" json:"keep_latest_artifact,omitempty"`
	KrokiEnabled                                          *bool              `url:"kroki_enabled,omitempty" json:"kroki_enabled,omitempty"`
	KrokiFormats                                          *map[string]bool   `url:"kroki_formats,omitempty" json:"kroki_formats,omitempty"`
	KrokiURL                                              *string            `url:"kroki_url,omitempty" json:"kroki_url,omitempty"`
	LocalMarkdownVersion                                  *int               `url:"local_markdown_version,omitempty" json:"local_markdown_version,omitempty"`
	LockMembershipsToLDAP                                 *bool              `url:"lock_memberships_to_ldap,omitempty" json:"lock_memberships_to_ldap,omitempty"`
	LoginRecaptchaProtectionEnabled                       *bool              `url:"login_recaptcha_protection_enabled,omitempty" json:"login_recaptcha_protection_enabled,omitempty"`
	MailgunEventsEnabled                                  *bool              `url:"mailgun_events_enabled,omitempty" json:"mailgun_events_enabled,omitempty"`
	MailgunSigningKey                                     *string            `url:"mailgun_signing_key,omitempty" json:"mailgun_signing_key,omitempty"`
	MaintenanceMode                                       *bool              `url:"maintenance_mode,omitempty" json:"maintenance_mode,omitempty"`
	MaintenanceModeMessage                                *string            `url:"maintenance_mode_message,omitempty" json:"maintenance_mode_message,omitempty"`
	MaxArtifactsSize                                      *int               `url:"max_artifacts_size,omitempty" json:"max_artifacts_size,omitempty"`
	MaxAttachmentSize                                     *int               `url:"max_attachment_size,omitempty" json:"max_attachment_size,omitempty"`
	MaxExportSize                                         *int               `url:"max_export_size,omitempty" json:"max_export_size,omitempty"`
	MaxImportSize                                         *int               `url:"max_import_size,omitempty" json:"max_import_size,omitempty"`
	MaxNumberOfRepositoryDownloads                        *int               `url:"max_number_of_repository_downloads,omitempty" json:"max_number_of_repository_downloads,omitempty"`
	MaxNumberOfRepositoryDownloadsWithinTimePeriod        *int               `url:"max_number_of_repository_downloads_within_time_period,omitempty" json:"max_number_of_repository_downloads_within_time_period,omitempty"`
	MaxPagesSize                                          *int               `url:"max_pages_size,omitempty" json:"max_pages_size,omitempty"`
	MaxPersonalAccessTokenLifetime                        *int               `url:"max_personal_access_token_lifetime,omitempty" json:"max_personal_access_token_lifetime,omitempty"`
	MaxSSHKeyLifetime                                     *int               `url:"max_ssh_key_lifetime,omitempty" json:"max_ssh_key_lifetime,omitempty"`
	MaxYAMLDepth                                          *int               `url:"max_yaml_depth,omitempty" json:"max_yaml_depth,omitempty"`
	MaxYAMLSizeBytes                                      *int               `url:"max_yaml_size_bytes,omitempty" json:"max_yaml_size_bytes,omitempty"`
	MetricsMethodCallThreshold                            *int               `url:"metrics_method_call_threshold,omitempty" json:"metrics_method_call_threshold,omitempty"`
	MinimumPasswordLength                                 *int               `url:"minimum_password_length,omitempty" json:"minimum_password_length,omitempty"`
	MirrorAvailable                                       *bool              `url:"mirror_available,omitempty" json:"mirror_available,omitempty"`
	MirrorCapacityThreshold                               *int               `url:"mirror_capacity_threshold,omitempty" json:"mirror_capacity_threshold,omitempty"`
	MirrorMaxCapacity                                     *int               `url:"mirror_max_capacity,omitempty" json:"mirror_max_capacity,omitempty"`
	MirrorMaxDelay                                        *int               `url:"mirror_max_delay,omitempty" json:"mirror_max_delay,omitempty"`
	NPMPackageRequestsForwarding                          *bool              `url:"npm_package_requests_forwarding,omitempty" json:"npm_package_requests_forwarding,omitempty"`
	NotesCreateLimit                                      *int               `url:"notes_create_limit,omitempty" json:"notes_create_limit,omitempty"`
	NotifyOnUnknownSignIn                                 *bool              `url:"notify_on_unknown_sign_in,omitempty" json:"notify_on_unknown_sign_in,omitempty"`
	OutboundLocalRequestsAllowlistRaw                     *string            `url:"outbound_local_requests_allowlist_raw,omitempty" json:"outbound_local_requests_allowlist_raw,omitempty"`
	OutboundLocalRequestsWhitelist                        *[]string          `url:"outbound_local_requests_whitelist,omitempty" json:"outbound_local_requests_whitelist,omitempty"`
	PackageRegistryCleanupPoliciesWorkerCapacity          *int               `url:"package_registry_cleanup_policies_worker_capacity,omitempty" json:"package_registry_cleanup_policies_worker_capacity,omitempty"`
	PagesDomainVerificationEnabled                        *bool              `url:"pages_domain_verification_enabled,omitempty" json:"pages_domain_verification_enabled,omitempty"`
	PasswordAuthenticationEnabledForGit                   *bool              `url:"password_authentication_enabled_for_git,omitempty" json:"password_authentication_enabled_for_git,omitempty"`
	PasswordAuthenticationEnabledForWeb                   *bool              `url:"password_authentication_enabled_for_web,omitempty" json:"password_authentication_enabled_for_web,omitempty"`
	PasswordNumberRequired                                *bool              `url:"password_number_required,omitempty" json:"password_number_required,omitempty"`
	PasswordSymbolRequired                                *bool              `url:"password_symbol_required,omitempty" json:"password_symbol_required,omitempty"`
	PasswordUppercaseRequired                             *bool              `url:"password_uppercase_required,omitempty" json:"password_uppercase_required,omitempty"`
	PasswordLowercaseRequired                             *bool              `url:"password_lowercase_required,omitempty" json:"password_lowercase_required,omitempty"`
	PerformanceBarAllowedGroupID                          *string            `url:"performance_bar_allowed_group_id,omitempty" json:"performance_bar_allowed_group_id,omitempty"`
	PerformanceBarAllowedGroupPath                        *string            `url:"performance_bar_allowed_group_path,omitempty" json:"performance_bar_allowed_group_path,omitempty"`
	PerformanceBarEnabled                                 *bool              `url:"performance_bar_enabled,omitempty" json:"performance_bar_enabled,omitempty"`
	PersonalAccessTokenPrefix                             *string            `url:"personal_access_token_prefix,omitempty" json:"personal_access_token_prefix,omitempty"`
	PlantumlEnabled                                       *bool              `url:"plantuml_enabled,omitempty" json:"plantuml_enabled,omitempty"`
	PlantumlURL                                           *string            `url:"plantuml_url,omitempty" json:"plantuml_url,omitempty"`
	PipelineLimitPerProjectUserSha                        *int               `url:"pipeline_limit_per_project_user_sha,omitempty" json:"pipeline_limit_per_project_user_sha,omitempty"`
	PollingIntervalMultiplier                             *float64           `url:"polling_interval_multiplier,omitempty" json:"polling_interval_multiplier,omitempty"`
	PreventMergeRequestsAuthorApproval                    *bool              `url:"prevent_merge_requests_author_approval,omitempty" json:"prevent_merge_requests_author_approval,omitempty"`
	PreventMergeRequestsCommittersApproval                *bool              `url:"prevent_merge_requests_committers_approval,omitempty" json:"prevent_merge_requests_committers_approval,omitempty"`
	ProjectDownloadExportLimit                            *int               `url:"project_download_export_limit,omitempty" json:"project_download_export_limit,omitempty"`
	ProjectExportEnabled                                  *bool              `url:"project_export_enabled,omitempty" json:"project_export_enabled,omitempty"`
	ProjectExportLimit                                    *int               `url:"project_export_limit,omitempty" json:"project_export_limit,omitempty"`
	ProjectImportLimit                                    *int               `url:"project_import_limit,omitempty" json:"project_import_limit,omitempty"`
	ProjectRunnerTokenExpirationInterval                  *int               `url:"project_runner_token_expiration_interval,omitempty" json:"project_runner_token_expiration_interval,omitempty"`
	PrometheusMetricsEnabled                              *bool              `url:"prometheus_metrics_enabled,omitempty" json:"prometheus_metrics_enabled,omitempty"`
	ProtectedCIVariables                                  *bool              `url:"protected_ci_variables,omitempty" json:"protected_ci_variables,omitempty"`
	PseudonymizerEnabled                                  *bool              `url:"pseudonymizer_enabled,omitempty" json:"pseudonymizer_enabled,omitempty"`
	PushEventActivitiesLimit                              *int               `url:"push_event_activities_limit,omitempty" json:"push_event_activities_limit,omitempty"`
	PushEventHooksLimit                                   *int               `url:"push_event_hooks_limit,omitempty" json:"push_event_hooks_limit,omitempty"`
	PyPIPackageRequestsForwarding                         *bool              `url:"pypi_package_requests_forwarding,omitempty" json:"pypi_package_requests_forwarding,omitempty"`
	RSAKeyRestriction                                     *int               `url:"rsa_key_restriction,omitempty" json:"rsa_key_restriction,omitempty"`
	RateLimitingResponseText                              *string            `url:"rate_limiting_response_text,omitempty" json:"rate_limiting_response_text,omitempty"`
	RawBlobRequestLimit                                   *int               `url:"raw_blob_request_limit,omitempty" json:"raw_blob_request_limit,omitempty"`
	RecaptchaEnabled                                      *bool              `url:"recaptcha_enabled,omitempty" json:"recaptcha_enabled,omitempty"`
	RecaptchaPrivateKey                                   *string            `url:"recaptcha_private_key,omitempty" json:"recaptcha_private_key,omitempty"`
	RecaptchaSiteKey                                      *string            `url:"recaptcha_site_key,omitempty" json:"recaptcha_site_key,omitempty"`
	ReceiveMaxInputSize                                   *int               `url:"receive_max_input_size,omitempty" json:"receive_max_input_size,omitempty"`
	RepositoryChecksEnabled                               *bool              `url:"repository_checks_enabled,omitempty" json:"repository_checks_enabled,omitempty"`
	RepositorySizeLimit                                   *int               `url:"repository_size_limit,omitempty" json:"repository_size_limit,omitempty"`
	RepositoryStorages                                    *[]string          `url:"repository_storages,omitempty" json:"repository_storages,omitempty"`
	RepositoryStoragesWeighted                            *map[string]int    `url:"repository_storages_weighted,omitempty" json:"repository_storages_weighted,omitempty"`
	RequireAdminApprovalAfterUserSignup                   *bool              `url:"require_admin_approval_after_user_signup,omitempty" json:"require_admin_approval_after_user_signup,omitempty"`
	RequireTwoFactorAuthentication                        *bool              `url:"require_two_factor_authentication,omitempty" json:"require_two_factor_authentication,omitempty"`
	RestrictedVisibilityLevels                            *[]VisibilityValue `url:"restricted_visibility_levels,omitempty" json:"restricted_visibility_levels,omitempty"`
	RunnerTokenExpirationInterval                         *int               `url:"runner_token_expiration_interval,omitempty" json:"runner_token_expiration_interval,omitempty"`
	SearchRateLimit                                       *int               `url:"search_rate_limit,omitempty" json:"search_rate_limit,omitempty"`
	SearchRateLimitUnauthenticated                        *int               `url:"search_rate_limit_unauthenticated,omitempty" json:"search_rate_limit_unauthenticated,omitempty"`
	SecretDetectionRevocationTokenTypesURL                *string            `url:"secret_detection_revocation_token_types_url,omitempty" json:"secret_detection_revocation_token_types_url,omitempty"`
	SecretDetectionTokenRevocationEnabled                 *bool              `url:"secret_detection_token_revocation_enabled,omitempty" json:"secret_detection_token_revocation_enabled,omitempty"`
	SecretDetectionTokenRevocationToken                   *string            `url:"secret_detection_token_revocation_token,omitempty" json:"secret_detection_token_revocation_token,omitempty"`
	SecretDetectionTokenRevocationURL                     *string            `url:"secret_detection_token_revocation_url,omitempty" json:"secret_detection_token_revocation_url,omitempty"`
	SendUserConfirmationEmail                             *bool              `url:"send_user_confirmation_email,omitempty" json:"send_user_confirmation_email,omitempty"`
	SentryClientsideDSN                                   *string            `url:"sentry_clientside_dsn,omitempty" json:"sentry_clientside_dsn,omitempty"`
	SentryDSN                                             *string            `url:"sentry_dsn,omitempty" json:"sentry_dsn,omitempty"`
	SentryEnabled                                         *string            `url:"sentry_enabled,omitempty" json:"sentry_enabled,omitempty"`
	SentryEnvironment                                     *string            `url:"sentry_environment,omitempty" json:"sentry_environment,omitempty"`
	SessionExpireDelay                                    *int               `url:"session_expire_delay,omitempty" json:"session_expire_delay,omitempty"`
	SharedRunnersEnabled                                  *bool              `url:"shared_runners_enabled,omitempty" json:"shared_runners_enabled,omitempty"`
	SharedRunnersMinutes                                  *int               `url:"shared_runners_minutes,omitempty" json:"shared_runners_minutes,omitempty"`
	SharedRunnersText                                     *string            `url:"shared_runners_text,omitempty" json:"shared_runners_text,omitempty"`
	SidekiqJobLimiterCompressionThresholdBytes            *int               `url:"sidekiq_job_limiter_compression_threshold_bytes,omitempty" json:"sidekiq_job_limiter_compression_threshold_bytes,omitempty"`
	SidekiqJobLimiterLimitBytes                           *int               `url:"sidekiq_job_limiter_limit_bytes,omitempty" json:"sidekiq_job_limiter_limit_bytes,omitempty"`
	SidekiqJobLimiterMode                                 *string            `url:"sidekiq_job_limiter_mode,omitempty" json:"sidekiq_job_limiter_mode,omitempty"`
	SignInText                                            *string            `url:"sign_in_text,omitempty" json:"sign_in_text,omitempty"`
	SignupEnabled                                         *bool              `url:"signup_enabled,omitempty" json:"signup_enabled,omitempty"`
	SlackAppEnabled                                       *bool              `url:"slack_app_enabled,omitempty" json:"slack_app_enabled,omitempty"`
	SlackAppID                                            *string            `url:"slack_app_id,omitempty" json:"slack_app_id,omitempty"`
	SlackAppSecret                                        *string            `url:"slack_app_secret,omitempty" json:"slack_app_secret,omitempty"`
	SlackAppSigningSecret                                 *string            `url:"slack_app_signing_secret,omitempty" json:"slack_app_signing_secret,omitempty"`
	SlackAppVerificationToken                             *string            `url:"slack_app_verification_token,omitempty" json:"slack_app_verification_token,omitempty"`
	SnippetSizeLimit                                      *int               `url:"snippet_size_limit,omitempty" json:"snippet_size_limit,omitempty"`
	SnowplowAppID                                         *string            `url:"snowplow_app_id,omitempty" json:"snowplow_app_id,omitempty"`
	SnowplowCollectorHostname                             *string            `url:"snowplow_collector_hostname,omitempty" json:"snowplow_collector_hostname,omitempty"`
	SnowplowCookieDomain                                  *string            `url:"snowplow_cookie_domain,omitempty" json:"snowplow_cookie_domain,omitempty"`
	SnowplowEnabled                                       *bool              `url:"snowplow_enabled,omitempty" json:"snowplow_enabled,omitempty"`
	SourcegraphEnabled                                    *bool              `url:"sourcegraph_enabled,omitempty" json:"sourcegraph_enabled,omitempty"`
	SourcegraphPublicOnly                                 *bool              `url:"sourcegraph_public_only,omitempty" json:"sourcegraph_public_only,omitempty"`
	SourcegraphURL                                        *string            `url:"sourcegraph_url,omitempty" json:"sourcegraph_url,omitempty"`
	SpamCheckAPIKey                                       *string            `url:"spam_check_api_key,omitempty" json:"spam_check_api_key,omitempty"`
	SpamCheckEndpointEnabled                              *bool              `url:"spam_check_endpoint_enabled,omitempty" json:"spam_check_endpoint_enabled,omitempty"`
	SpamCheckEndpointURL                                  *string            `url:"spam_check_endpoint_url,omitempty" json:"spam_check_endpoint_url,omitempty"`
	SuggestPipelineEnabled                                *bool              `url:"suggest_pipeline_enabled,omitempty" json:"suggest_pipeline_enabled,omitempty"`
	TerminalMaxSessionTime                                *int               `url:"terminal_max_session_time,omitempty" json:"terminal_max_session_time,omitempty"`
	Terms                                                 *string            `url:"terms,omitempty" json:"terms,omitempty"`
	ThrottleAuthenticatedAPIEnabled                       *bool              `url:"throttle_authenticated_api_enabled,omitempty" json:"throttle_authenticated_api_enabled,omitempty"`
	ThrottleAuthenticatedAPIPeriodInSeconds               *int               `url:"throttle_authenticated_api_period_in_seconds,omitempty" json:"throttle_authenticated_api_period_in_seconds,omitempty"`
	ThrottleAuthenticatedAPIRequestsPerPeriod             *int               `url:"throttle_authenticated_api_requests_per_period,omitempty" json:"throttle_authenticated_api_requests_per_period,omitempty"`
	ThrottleAuthenticatedDeprecatedAPIEnabled             *bool              `url:"throttle_authenticated_deprecated_api_enabled,omitempty" json:"throttle_authenticated_deprecated_api_enabled,omitempty"`
	ThrottleAuthenticatedDeprecatedAPIPeriodInSeconds     *int               `url:"throttle_authenticated_deprecated_api_period_in_seconds,omitempty" json:"throttle_authenticated_deprecated_api_period_in_seconds,omitempty"`
	ThrottleAuthenticatedDeprecatedAPIRequestsPerPeriod   *int               `url:"throttle_authenticated_deprecated_api_requests_per_period,omitempty" json:"throttle_authenticated_deprecated_api_requests_per_period,omitempty"`
	ThrottleAuthenticatedFilesAPIEnabled                  *bool              `url:"throttle_authenticated_files_api_enabled,omitempty" json:"throttle_authenticated_files_api_enabled,omitempty"`
	ThrottleAuthenticatedFilesAPIPeriodInSeconds          *int               `url:"throttle_authenticated_files_api_period_in_seconds,omitempty" json:"throttle_authenticated_files_api_period_in_seconds,omitempty"`
	ThrottleAuthenticatedFilesAPIRequestsPerPeriod        *int               `url:"throttle_authenticated_files_api_requests_per_period,omitempty" json:"throttle_authenticated_files_api_requests_per_period,omitempty"`
	ThrottleAuthenticatedGitLFSEnabled                    *bool              `url:"throttle_authenticated_git_lfs_enabled,omitempty" json:"throttle_authenticated_git_lfs_enabled,omitempty"`
	ThrottleAuthenticatedGitLFSPeriodInSeconds            *int               `url:"throttle_authenticated_git_lfs_period_in_seconds,omitempty" json:"throttle_authenticated_git_lfs_period_in_seconds,omitempty"`
	ThrottleAuthenticatedGitLFSRequestsPerPeriod          *int               `url:"throttle_authenticated_git_lfs_requests_per_period,omitempty" json:"throttle_authenticated_git_lfs_requests_per_period,omitempty"`
	ThrottleAuthenticatedPackagesAPIEnabled               *bool              `url:"throttle_authenticated_packages_api_enabled,omitempty" json:"throttle_authenticated_packages_api_enabled,omitempty"`
	ThrottleAuthenticatedPackagesAPIPeriodInSeconds       *int               `url:"throttle_authenticated_packages_api_period_in_seconds,omitempty" json:"throttle_authenticated_packages_api_period_in_seconds,omitempty"`
	ThrottleAuthenticatedPackagesAPIRequestsPerPeriod     *int               `url:"throttle_authenticated_packages_api_requests_per_period,omitempty" json:"throttle_authenticated_packages_api_requests_per_period,omitempty"`
	ThrottleAuthenticatedWebEnabled                       *bool              `url:"throttle_authenticated_web_enabled,omitempty" json:"throttle_authenticated_web_enabled,omitempty"`
	ThrottleAuthenticatedWebPeriodInSeconds               *int               `url:"throttle_authenticated_web_period_in_seconds,omitempty" json:"throttle_authenticated_web_period_in_seconds,omitempty"`
	ThrottleAuthenticatedWebRequestsPerPeriod             *int               `url:"throttle_authenticated_web_requests_per_period,omitempty" json:"throttle_authenticated_web_requests_per_period,omitempty"`
	ThrottleIncidentManagementNotificationEnabled         *bool              `url:"throttle_incident_management_notification_enabled,omitempty" json:"throttle_incident_management_notification_enabled,omitempty"`
	ThrottleIncidentManagementNotificationPerPeriod       *int               `url:"throttle_incident_management_notification_per_period,omitempty" json:"throttle_incident_management_notification_per_period,omitempty"`
	ThrottleIncidentManagementNotificationPeriodInSeconds *int               `url:"throttle_incident_management_notification_period_in_seconds,omitempty" json:"throttle_incident_management_notification_period_in_seconds,omitempty"`
	ThrottleProtectedPathsEnabled                         *bool              `url:"throttle_protected_paths_enabled_enabled,omitempty" json:"throttle_protected_paths_enabled,omitempty"`
	ThrottleProtectedPathsPeriodInSeconds                 *int               `url:"throttle_protected_paths_enabled_period_in_seconds,omitempty" json:"throttle_protected_paths_period_in_seconds,omitempty"`
	ThrottleProtectedPathsRequestsPerPeriod               *int               `url:"throttle_protected_paths_enabled_requests_per_period,omitempty" json:"throttle_protected_paths_per_period,omitempty"`
	ThrottleUnauthenticatedAPIEnabled                     *bool              `url:"throttle_unauthenticated_api_enabled,omitempty" json:"throttle_unauthenticated_api_enabled,omitempty"`
	ThrottleUnauthenticatedAPIPeriodInSeconds             *int               `url:"throttle_unauthenticated_api_period_in_seconds,omitempty" json:"throttle_unauthenticated_api_period_in_seconds,omitempty"`
	ThrottleUnauthenticatedAPIRequestsPerPeriod           *int               `url:"throttle_unauthenticated_api_requests_per_period,omitempty" json:"throttle_unauthenticated_api_requests_per_period,omitempty"`
	ThrottleUnauthenticatedDeprecatedAPIEnabled           *bool              `url:"throttle_unauthenticated_deprecated_api_enabled,omitempty" json:"throttle_unauthenticated_deprecated_api_enabled,omitempty"`
	ThrottleUnauthenticatedDeprecatedAPIPeriodInSeconds   *int               `url:"throttle_unauthenticated_deprecated_api_period_in_seconds,omitempty" json:"throttle_unauthenticated_deprecated_api_period_in_seconds,omitempty"`
	ThrottleUnauthenticatedDeprecatedAPIRequestsPerPeriod *int               `url:"throttle_unauthenticated_deprecated_api_requests_per_period,omitempty" json:"throttle_unauthenticated_deprecated_api_requests_per_period,omitempty"`
	ThrottleUnauthenticatedEnabled                        *bool              `url:"throttle_unauthenticated_enabled,omitempty" json:"throttle_unauthenticated_enabled,omitempty"`
	ThrottleUnauthenticatedFilesAPIEnabled                *bool              `url:"throttle_unauthenticated_files_api_enabled,omitempty" json:"throttle_unauthenticated_files_api_enabled,omitempty"`
	ThrottleUnauthenticatedFilesAPIPeriodInSeconds        *int               `url:"throttle_unauthenticated_files_api_period_in_seconds,omitempty" json:"throttle_unauthenticated_files_api_period_in_seconds,omitempty"`
	ThrottleUnauthenticatedFilesAPIRequestsPerPeriod      *int               `url:"throttle_unauthenticated_files_api_requests_per_period,omitempty" json:"throttle_unauthenticated_files_api_requests_per_period,omitempty"`
	ThrottleUnauthenticatedGitLFSEnabled                  *bool              `url:"throttle_unauthenticated_git_lfs_enabled,omitempty" json:"throttle_unauthenticated_git_lfs_enabled,omitempty"`
	ThrottleUnauthenticatedGitLFSPeriodInSeconds          *int               `url:"throttle_unauthenticated_git_lfs_period_in_seconds,omitempty" json:"throttle_unauthenticated_git_lfs_period_in_seconds,omitempty"`
	ThrottleUnauthenticatedGitLFSRequestsPerPeriod        *int               `url:"throttle_unauthenticated_git_lfs_requests_per_period,omitempty" json:"throttle_unauthenticated_git_lfs_requests_per_period,omitempty"`
	ThrottleUnauthenticatedPackagesAPIEnabled             *bool              `url:"throttle_unauthenticated_packages_api_enabled,omitempty" json:"throttle_unauthenticated_packages_api_enabled,omitempty"`
	ThrottleUnauthenticatedPackagesAPIPeriodInSeconds     *int               `url:"throttle_unauthenticated_packages_api_period_in_seconds,omitempty" json:"throttle_unauthenticated_packages_api_period_in_seconds,omitempty"`
	ThrottleUnauthenticatedPackagesAPIRequestsPerPeriod   *int               `url:"throttle_unauthenticated_packages_api_requests_per_period,omitempty" json:"throttle_unauthenticated_packages_api_requests_per_period,omitempty"`
	ThrottleUnauthenticatedPeriodInSeconds                *int               `url:"throttle_unauthenticated_period_in_seconds,omitempty" json:"throttle_unauthenticated_period_in_seconds,omitempty"`
	ThrottleUnauthenticatedRequestsPerPeriod              *int               `url:"throttle_unauthenticated_requests_per_period,omitempty" json:"throttle_unauthenticated_requests_per_period,omitempty"`
	ThrottleUnauthenticatedWebEnabled                     *bool              `url:"throttle_unauthenticated_web_enabled,omitempty" json:"throttle_unauthenticated_web_enabled,omitempty"`
	ThrottleUnauthenticatedWebPeriodInSeconds             *int               `url:"throttle_unauthenticated_web_period_in_seconds,omitempty" json:"throttle_unauthenticated_web_period_in_seconds,omitempty"`
	ThrottleUnauthenticatedWebRequestsPerPeriod           *int               `url:"throttle_unauthenticated_web_requests_per_period,omitempty" json:"throttle_unauthenticated_web_requests_per_period,omitempty"`
	TimeTrackingLimitToHours                              *bool              `url:"time_tracking_limit_to_hours,omitempty" json:"time_tracking_limit_to_hours,omitempty"`
	TwoFactorGracePeriod                                  *int               `url:"two_factor_grace_period,omitempty" json:"two_factor_grace_period,omitempty"`
	UniqueIPsLimitEnabled                                 *bool              `url:"unique_ips_limit_enabled,omitempty" json:"unique_ips_limit_enabled,omitempty"`
	UniqueIPsLimitPerUser                                 *int               `url:"unique_ips_limit_per_user,omitempty" json:"unique_ips_limit_per_user,omitempty"`
	UniqueIPsLimitTimeWindow                              *int               `url:"unique_ips_limit_time_window,omitempty" json:"unique_ips_limit_time_window,omitempty"`
	UpdatingNameDisabledForUsers                          *bool              `url:"updating_name_disabled_for_users,omitempty" json:"updating_name_disabled_for_users,omitempty"`
	UsagePingEnabled                                      *bool              `url:"usage_ping_enabled,omitempty" json:"usage_ping_enabled,omitempty"`
	UsagePingFeaturesEnabled                              *bool              `url:"usage_ping_features_enabled,omitempty" json:"usage_ping_features_enabled,omitempty"`
	UserDeactivationEmailsEnabled                         *bool              `url:"user_deactivation_emails_enabled,omitempty" json:"user_deactivation_emails_enabled,omitempty"`
	UserDefaultExternal                                   *bool              `url:"user_default_external,omitempty" json:"user_default_external,omitempty"`
	UserDefaultInternalRegex                              *string            `url:"user_default_internal_regex,omitempty" json:"user_default_internal_regex,omitempty"`
	UserEmailLookupLimit                                  *int               `url:"user_email_lookup_limit,omitempty" json:"user_email_lookup_limit,omitempty"`
	UserOauthApplications                                 *bool              `url:"user_oauth_applications,omitempty" json:"user_oauth_applications,omitempty"`
	UserShowAddSSHKeyMessage                              *bool              `url:"user_show_add_ssh_key_message,omitempty" json:"user_show_add_ssh_key_message,omitempty"`
	UsersGetByIDLimit                                     *int               `url:"users_get_by_id_limit,omitempty" json:"users_get_by_id_limit,omitempty"`
	UsersGetByIDLimitAllowlistRaw                         *string            `url:"users_get_by_id_limit_allowlist_raw,omitempty" json:"users_get_by_id_limit_allowlist_raw,omitempty"`
	VersionCheckEnabled                                   *bool              `url:"version_check_enabled,omitempty" json:"version_check_enabled,omitempty"`
	WebIDEClientsidePreviewEnabled                        *bool              `url:"web_ide_clientside_preview_enabled,omitempty" json:"web_ide_clientside_preview_enabled,omitempty"`
	WhatsNewVariant                                       *string            `url:"whats_new_variant,omitempty" json:"whats_new_variant,omitempty"`
	WikiPageMaxContentBytes                               *int               `url:"wiki_page_max_content_bytes,omitempty" json:"wiki_page_max_content_bytes,omitempty"`
}

// UpdateSettings updates the application settings.
//
// GitLab API docs:
// https://docs.gitlab.com/ee/api/settings.html#change-application-settings
func (s *SettingsService) UpdateSettings(opt *UpdateSettingsOptions, options ...RequestOptionFunc) (*Settings, *Response, error) {
	req, err := s.client.NewRequest(http.MethodPut, "application/settings", opt, options)
	if err != nil {
		return nil, nil, err
	}

	as := new(Settings)
	resp, err := s.client.Do(req, as)
	if err != nil {
		return nil, resp, err
	}

	return as, resp, nil
}
