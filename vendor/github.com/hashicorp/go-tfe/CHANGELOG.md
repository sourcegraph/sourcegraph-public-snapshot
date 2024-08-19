# v1.32.1

## Dependency Update
* Updated go-slug dependency to v0.12.1

# v1.32.0

## Enhancements
* Added BETA support for adding and updating custom permissions to `TeamProjectAccesses`. A `TeamProjectAccessType` of `"custom"` can set various permissions applied at
the project level to the project itself (`TeamProjectAccessProjectPermissionsOptions`) and all of the workspaces in a project (`TeamProjectAccessWorkspacePermissionsOptions`) by @rberecka [#745](https://github.com/hashicorp/go-tfe/pull/745)
* Added BETA field `Provisional` to `ConfigurationVersions` by @brandonc [#746](https://github.com/hashicorp/go-tfe/pull/746)


# v1.31.0

## Enhancements
* Added BETA support for including `projects` relationship and `projects-count` attribute to policy_set on create by @hs26gill [#737](https://github.com/hashicorp/go-tfe/pull/737)
* Added BETA method `AddProjects` and `RemoveProjects` for attaching/detaching policy set to projects by @Netra2104 [#735](https://github.com/hashicorp/go-tfe/pull/735)

# v1.30.0

## Enhancements
* Adds `SignatureSigningMethod` and `SignatureDigestMethod` fields in `AdminSAMLSetting` struct by @karvounis-form3 [#731](https://github.com/hashicorp/go-tfe/pull/731)
* Adds `Certificate`, `PrivateKey`, `TeamManagementEnabled`, `AuthnRequestsSigned`, `WantAssertionsSigned`, `SignatureSigningMethod`, `SignatureDigestMethod` fields in `AdminSAMLSettingsUpdateOptions` struct by @karvounis-form3 [#731](https://github.com/hashicorp/go-tfe/pull/731)

# v1.29.0

## Enhancements
* Adds `RunPreApplyCompleted` run status by @uk1288 [#727](https://github.com/hashicorp/go-tfe/pull/727)
* Added BETA support for saved plan runs, by @nfagerlund [#724](https://github.com/hashicorp/go-tfe/pull/724)
    * New `SavePlan` fields in `Run` and `RunCreateOptions`
    * New `RunPlannedAndSaved` `RunStatus` value
    * New `PlannedAndSavedAt` field in `RunStatusTimestamps`
    * New `RunOperationSavePlan` constant for run list filters

# v1.28.0

## Enhancements
* Update `Workspaces` to include associated `project` resource by @glennsarti [#714](https://github.com/hashicorp/go-tfe/pull/714)
* Adds BETA method `Upload` method to `StateVersions` and support for pending state versions by @brandonc [#717](https://github.com/hashicorp/go-tfe/pull/717)
* Adds support for the query parameter `q` to search `Organization Tags` by name by @sharathrnair87 [#720](https://github.com/hashicorp/go-tfe/pull/720)
* Added ContextWithResponseHeaderHook support to `IPRanges` by @brandonc [#717](https://github.com/hashicorp/go-tfe/pull/717)

## Bug Fixes
* `ConfigurationVersions`, `PolicySetVersions`, and `RegistryModules` `Upload` methods were sending API credentials to the specified upload URL, which was unnecessary by @brandonc [#717](https://github.com/hashicorp/go-tfe/pull/717)

# v1.27.0

## Enhancements
* Adds `RunPreApplyRunning` and `RunQueuingApply` run statuses by @uk1288 [#712](https://github.com/hashicorp/go-tfe/pull/712)

## Bug Fixes
* AgentPool `Update` is not able to remove all allowed workspaces from an agent pool. That operation is now handled by a separate `UpdateAllowedWorkspaces` method using `AgentPoolAllowedWorkspacesUpdateOptions` by @hs26gill [#701](https://github.com/hashicorp/go-tfe/pull/701)

# v1.26.0

## Enhancements

* Adds BETA fields `ResourceImports` count to both `Plan` and `Apply` types as well as `AllowConfigGeneration` to the `Run` struct type. These fields are not generally available and are subject to change in a future release.

# v1.25.1

## Bug Fixes
* Workspace safe delete conflict error when workspace is locked has been restored
to the original message using the error `ErrWorkspaceLockedCannotDelete` instead of
`ErrWorkspaceLocked`

# v1.25.0

## Enhancements
* Workspace safe delete 409 conflict errors associated with resources still being managed or being processed (indicating that you should try again later) are now the named errors  `ErrWorkspaceStillProcessing` and `ErrWorkspaceNotSafeToDelete` by @brandonc [#703](https://github.com/hashicorp/go-tfe/pull/703)

# v1.24.0

## Enhancements
* Adds support for a new variable field `version-id` by @arybolovlev [#697](https://github.com/hashicorp/go-tfe/pull/697)
* Adds `ExpiredAt` field to `OrganizationToken`, `TeamToken`, and `UserToken`. This enhancement will be available in TFE release, v202305-1. @JuliannaTetreault [#672](https://github.com/hashicorp/go-tfe/pull/672)
* Adds `ContextWithResponseHeaderHook` context for use with the ClientRequest Do method that allows callers to define a callback which receives raw http Response headers.  @apparentlymart [#689](https://github.com/hashicorp/go-tfe/pull/689)


# v1.23.0

## Features
* `ApplyToProjects` and `RemoveFromProjects` to `VariableSets` endpoints now generally available.
* `ListForProject` to `VariableSets` endpoints now generally available.

## Enhancements
* Adds `OrganizationScoped` and `AllowedWorkspaces` fields for creating workspace scoped agent pools and adds `AllowedWorkspacesName` for filtering agents pools associated with a given workspace by @hs26gill [#682](https://github.com/hashicorp/go-tfe/pull/682/files)

## Bug Fixes


# v1.22.0

## Beta API Changes
* The beta `no_code` field in `RegistryModuleCreateOptions` has been changed from `bool` to `*bool` and will be removed in a future version because a new, preferred method for managing no-code registry modules has been added in this release.

## Features
* Add beta endpoints `Create`, `Read`, `Update`, and `Delete` to manage no-code provisioning for a `RegistryModule`. This allows users to enable no-code provisioning for a registry module, and to configure the provisioning settings for that module version. This also allows users to disable no-code provisioning for a module version. @dsa0x [#669](https://github.com/hashicorp/go-tfe/pull/669)

# v1.21.0

## Features
* Add beta endpoints `ApplyToProjects`  and `RemoveFromProjects` to `VariableSets`.  Applying a variable set to a project will apply that variable set to all current and future workspaces in that project.
* Add beta endpoint `ListForProject` to `VariableSets` to list all variable sets applied to a project.
* Add endpoint `RunEvents` which lists events for a specific run by @glennsarti [#680](https://github.com/hashicorp/go-tfe/pull/680)

## Bug Fixes
* `VariableSets.Read` did not honor the Include values due to a syntax error in the struct tag of `VariableSetReadOptions` by @sgap [#678](https://github.com/hashicorp/go-tfe/pull/678)

## Enhancements
* Adds `ProjectID` filter to allow filtering of workspaces of a given project in an organization by @hs26gill [#671](https://github.com/hashicorp/go-tfe/pull/671)
* Adds `Name` filter to allow filtering of projects by @hs26gill [#668](https://github.com/hashicorp/go-tfe/pull/668/files)
* Adds `ManageMembership` permission to team `OrganizationAccess` by @JarrettSpiker [#652](https://github.com/hashicorp/go-tfe/pull/652)
* Adds `RotateKey` and `TrimKey` Admin endpoints by @mpminardi [#666](https://github.com/hashicorp/go-tfe/pull/666)
* Adds `Permissions` to `User` by @jeevanragula [#674](https://github.com/hashicorp/go-tfe/pull/674)
* Adds `IsEnterprise` and `IsCloud` boolean methods to the client by @sebasslash [#675](https://github.com/hashicorp/go-tfe/pull/675)

# v1.20.0

## Enhancements
* Update team project access to include additional project roles by @joekarl [#642](https://github.com/hashicorp/go-tfe/pull/642)

# v1.19.0

## Enhancements
* Removed Beta tags from `Project` features by @hs26gill [#637](https://github.com/hashicorp/go-tfe/pull/637)
* Add `Filter` and `Sort` fields to `AdminWorkspaceListOptions` to allow filtering and sorting of workspaces by @laurenolivia [#641](https://github.com/hashicorp/go-tfe/pull/641)
* Add support for `List` and `Read` Github app installation APIs by @roleesinhaHC [#655](https://github.com/hashicorp/go-tfe/pull/655)
* Add `GHAInstallationID` field to `VCSRepoOptions` and `VCSRepo` structs by @roleesinhaHC [#655](https://github.com/hashicorp/go-tfe/pull/655)

# v1.18.0

## Enhancements
* Adds `BaseURL` and `BaseRegistryURL` methods to `Client` to expose its configuration by @brandonc [#638](https://github.com/hashicorp/go-tfe/pull/638)
* Adds `ReadWorkspaces` and `ReadProjects` permissions to `Organizations` by @JuliannaTetreault [#614](https://github.com/hashicorp/go-tfe/pull/614)

# v1.17.0

## Enhancements
* Add Beta endpoint `TeamProjectAccesses` to manage Project Access for Teams by @hs26gill [#599](https://github.com/hashicorp/go-tfe/pull/599)
* Updates api doc links from terraform.io to developer.hashicorp domain by @uk1288 [#629](https://github.com/hashicorp/go-tfe/pull/629)
* Adds `UploadTarGzip()` method to `RegistryModules` and `ConfigurationVersions` interface by @sebasslash [#623](https://github.com/hashicorp/go-tfe/pull/623)
* Adds `ManageProjects` field to `OrganizationAccess` struct by @hs26gill [#633](https://github.com/hashicorp/go-tfe/pull/633)
* Adds agent-count to `AgentPools` endpoint. @evilensky [#611](https://github.com/hashicorp/go-tfe/pull/611)
* Adds `Links` to `Workspace`, (currently contains "self" and "self-html" paths) @brandonc [#622](https://github.com/hashicorp/go-tfe/pull/622)

# v1.16.0

## Bug Fixes

* Project names were being incorrectly validated as ID's @brandonc [#608](https://github.com/hashicorp/go-tfe/pull/608)

## Enhancements
* Adds `List()` method to `GPGKeys` interface by @sebasslash [#602](https://github.com/hashicorp/go-tfe/pull/602)
* Adds `ProviderBinaryUploaded` field to `RegistryPlatforms` struct by @sebasslash [#602](https://github.com/hashicorp/go-tfe/pull/602)

# v1.15.0

## Enhancements

* Add Beta `Projects` endpoint. The API is in not yet available to all users @hs26gill [#564](https://github.com/hashicorp/go-tfe/pull/564)

# v1.14.0

## Enhancements

* Adds Beta parameter `Overridable` for OPA `policy set` update API (`PolicySetUpdateOptions`) @mrinalirao [#594](https://github.com/hashicorp/go-tfe/pull/594)
* Adds new task stage status values representing `canceled`, `errored`, `unreachable` @mrinalirao [#594](https://github.com/hashicorp/go-tfe/pull/594)

# v1.13.0

## Bug Fixes

* Fixes `AuditTrail` pagination parameters (`CurrentPage`, `PreviousPage`, `NextPage`, `TotalPages`, `TotalCount`), which were not deserialized after reading from the List endpoint by @brandonc [#586](https://github.com/hashicorp/go-tfe/pull/586)

## Enhancements

* Add OPA support to the Policy Set APIs by @mrinalirao [#575](https://github.com/hashicorp/go-tfe/pull/575)
* Add OPA support to the Policy APIs by @mrinalirao [#579](https://github.com/hashicorp/go-tfe/pull/579)
* Add support for enabling no-code provisioning in an existing or new `RegistryModule` by @miguelhrocha [#562](https://github.com/hashicorp/go-tfe/pull/562)
* Add Policy Evaluation and Policy Set Outcome APIs by @mrinalirao [#583](https://github.com/hashicorp/go-tfe/pull/583)
* Add OPA support to Task Stage APIs by @mrinalirao [#584](https://github.com/hashicorp/go-tfe/pull/584)

# v1.12.0

## Enhancements

* Add `search[wildcard-name]` to `WorkspaceListOptions` by @laurenolivia [#569](https://github.com/hashicorp/go-tfe/pull/569)
* Add `NotificationTriggerAssessmentCheckFailed` notification trigger type by @rexredinger [#549](https://github.com/hashicorp/go-tfe/pull/549)
* Add `RemoteTFEVersion()` to the `Client` interface, which exposes the `X-TFE-Version` header set by a remote TFE instance by @sebasslash [#563](https://github.com/hashicorp/go-tfe/pull/563)
* Validate the module version as a version instead of an ID [#409](https://github.com/hashicorp/go-tfe/pull/409)
* Add `AllowForceDeleteWorkspaces` setting to `Organizations` by @JarrettSpiker [#539](https://github.com/hashicorp/go-tfe/pull/539)
* Add `SafeDelete` and `SafeDeleteID` APIs to `Workspaces` by @JarrettSpiker [#539](https://github.com/hashicorp/go-tfe/pull/539)
* Add `ForceExecute()` to `Runs` to allow force executing a run by @annawinkler [#570](https://github.com/hashicorp/go-tfe/pull/570)
* Pre-plan and Pre-Apply Run Tasks are now generally available (beta comments removed) by @glennsarti [#555](https://github.com/hashicorp/go-tfe/pull/555)

# v1.11.0

## Enhancements

* Add `Query` and `Status` fields to `OrganizationMembershipListOptions` to allow filtering memberships by status or username by @sebasslash [#550](https://github.com/hashicorp/go-tfe/pull/550)
* Add `ListForWorkspace` method to `VariableSets` interface to enable fetching variable sets associated with a workspace by @tstapler [#552](https://github.com/hashicorp/go-tfe/pull/552)
* Add `NotificationTriggerAssessmentDrifted` and `NotificationTriggerAssessmentFailed` notification trigger types by @lawliet89 [#542](https://github.com/hashicorp/go-tfe/pull/542)

## Bug Fixes
* Fix marshalling of run variables in `RunCreateOptions`. The `Variables` field type in `Run` struct has changed from `[]*RunVariable` to `[]*RunVariableAttr` by @Uk1288 [#531](https://github.com/hashicorp/go-tfe/pull/531)

# v1.10.0

## Enhancements

* Add `Query` param field to `OrganizationListOptions` to allow searching based on name or email by @laurenolivia [#529](https://github.com/hashicorp/go-tfe/pull/529)
* Add optional `AssessmentsEnforced` to organizations and `AssessmentsEnabled` to workspaces for managing the workspace and organization health assessment (drift detection) setting by @rexredinger [#462](https://github.com/hashicorp/go-tfe/pull/462)

## Bug Fixes
* Fixes null value returned in variable set relationship in `VariableSetVariable` by @sebasslash [#521](https://github.com/hashicorp/go-tfe/pull/521)

# v1.9.0

## Enhancements
* `RunListOptions` is generally available, and rename field (Name -> User) by @mjyocca [#472](https://github.com/hashicorp/go-tfe/pull/472)
* [Beta] Adds optional `JsonState` field to `StateVersionCreateOptions` by @megan07 [#514](https://github.com/hashicorp/go-tfe/pull/514)

## Bug Fixes
* Fixed invalid memory address error when using `TaskResults` field by @glennsarti [#517](https://github.com/hashicorp/go-tfe/pull/517)

# v1.8.0

## Enhancements

* Adds support for reading and listing Agents by @laurenolivia [#456](https://github.com/hashicorp/go-tfe/pull/456)
* It was previously logged that we added an `Include` param field to `PolicySetListOptions` to allow policy list to include related resource data such as workspaces, policies, newest_version, or current_version by @Uk1288 [#497](https://github.com/hashicorp/go-tfe/pull/497) in 1.7.0, but this was a mistake and the field is added in v1.8.0

# v1.7.0

## Enhancements

* Adds new run creation attributes: `allow-empty-apply`, `terraform-version`, `plan-only` by @sebasslash [#482](https://github.com/hashicorp/go-tfe/pull/447)
* Adds additional Task Stage and Run Statuses for Pre-plan run tasks by @glennsarti [#469](https://github.com/hashicorp/go-tfe/pull/469)
* Adds `stage` field to the create and update methods for Workspace Run Tasks by @glennsarti [#469](https://github.com/hashicorp/go-tfe/pull/469)
* Adds `ResourcesProcessed`, `StateVersion`, `TerraformVersion`, `Modules`, `Providers`, and `Resources` fields to the State Version struct by @laurenolivia [#484](https://github.com/hashicorp/go-tfe/pull/484)
* Add `Include` param field to `PolicySetListOptions` to allow policy list to include related resource data such as workspaces, policies, newest_version, or current_version by @Uk1288 [#497](https://github.com/hashicorp/go-tfe/pull/497)
* Allow FileTriggersEnabled to be set to false when Git tags are present by @mjyocca @hashimoon  [#468] (https://github.com/hashicorp/go-tfe/pull/468)

# v1.6.0

## Enhancements
* Remove beta messaging for Run Tasks by @glennsarti [#447](https://github.com/hashicorp/go-tfe/pull/447)
* Adds `Description` field to the `RunTask` object by @glennsarti [#447](https://github.com/hashicorp/go-tfe/pull/447)
* Add `Name` field to `OAuthClient` by @barrettclark [#466](https://github.com/hashicorp/go-tfe/pull/466)
* Add support for creating both public and private `RegistryModule` with no VCS connection by @Uk1288 [#460](https://github.com/hashicorp/go-tfe/pull/460)
* Add `ConfigurationSourceAdo` configuration source option by @mjyocca [#467](https://github.com/hashicorp/go-tfe/pull/467)
* [beta] state version outputs may now include a detailed-type attribute in a future API release by @brandonc [#479](https://github.com/hashicorp/go-tfe/pull/429)

# v1.5.0

## Enhancements
* [beta] Add support for triggering Workspace runs through matching Git tags [#434](https://github.com/hashicorp/go-tfe/pull/434)
* Add `Query` param field to `AgentPoolListOptions` to allow searching based on agent pool name, by @JarrettSpiker [#417](https://github.com/hashicorp/go-tfe/pull/417)
* Add organization scope and allowed workspaces field for scope agents by @Netra2104 [#453](https://github.com/hashicorp/go-tfe/pull/453)
* Adds `Namespace` and `RegistryName` fields to `RegistryModuleID` to allow reading of Public Registry Modules by @Uk1288 [#464](https://github.com/hashicorp/go-tfe/pull/464)

## Bug fixes
* Fixed JSON mapping for Configuration Versions failing to properly set the `speculative` property [#459](https://github.com/hashicorp/go-tfe/pull/459)

# v1.4.0

## Enhancements
* Adds `RetryServerErrors` field to the `Config` object by @sebasslash [#439](https://github.com/hashicorp/go-tfe/pull/439)
* Adds support for the GPG Keys API by @sebasslash [#429](https://github.com/hashicorp/go-tfe/pull/429)
* Adds support for new `WorkspaceLimit` Admin setting for organizations [#425](https://github.com/hashicorp/go-tfe/pull/425)
* Adds support for new `ExcludeTags` workspace list filter field by @Uk1288 [#438](https://github.com/hashicorp/go-tfe/pull/438)
* [beta] Adds additional filter fields to `RunListOptions` by @mjyocca [#424](https://github.com/hashicorp/go-tfe/pull/424)
* [beta] Renames the optional StateVersion field `ExtState` to `JSONStateOutputs` and changes the purpose and type by @annawinkler [#444](https://github.com/hashicorp/go-tfe/pull/444) and @brandoncroft [#452](https://github.com/hashicorp/go-tfe/pull/452)

# v1.3.0

## Enhancements
* Adds support for Microsoft Teams notification configuration by @JarrettSpiker [#398](https://github.com/hashicorp/go-tfe/pull/389)
* Add support for Audit Trail API by @sebasslash [#407](https://github.com/hashicorp/go-tfe/pull/407)
* Adds Private Registry Provider, Provider Version, and Provider Platform APIs support by @joekarl and @annawinkler [#313](https://github.com/hashicorp/go-tfe/pull/313)
* Adds List Registry Modules endpoint by @chroju [#385](https://github.com/hashicorp/go-tfe/pull/385)
* Adds `WebhookURL` field to `VCSRepo` struct by @kgns [#413](https://github.com/hashicorp/go-tfe/pull/413)
* Adds `Category` field to `VariableUpdateOptions` struct by @jtyr [#397](https://github.com/hashicorp/go-tfe/pull/397)
* Adds `TriggerPatterns` to `Workspace` by @matejrisek [#400](https://github.com/hashicorp/go-tfe/pull/400)
* [beta] Adds `ExtState` field to `StateVersionCreateOptions` by @brandonc [#416](https://github.com/hashicorp/go-tfe/pull/416)

# v1.2.0

## Enhancements
* Adds support for reading current state version outputs to StateVersionOutputs, which can be useful for reading outputs when users don't have the necessary permissions to read the entire state by @brandonc [#370](https://github.com/hashicorp/go-tfe/pull/370)
* Adds Variable Set methods for `ApplyToWorkspaces` and `RemoveFromWorkspaces` by @byronwolfman [#375](https://github.com/hashicorp/go-tfe/pull/375)
* Adds `Names` query param field to `TeamListOptions` by @sebasslash [#393](https://github.com/hashicorp/go-tfe/pull/393)
* Adds `Emails` query param field to `OrganizationMembershipListOptions` by @sebasslash [#393](https://github.com/hashicorp/go-tfe/pull/393)
* Adds Run Tasks API support by @glennsarti [#381](https://github.com/hashicorp/go-tfe/pull/381), [#382](https://github.com/hashicorp/go-tfe/pull/382) and [#383](https://github.com/hashicorp/go-tfe/pull/383)


## Bug fixes
* Fixes ignored comment when performing apply, discard, cancel, and force-cancel run actions [#388](https://github.com/hashicorp/go-tfe/pull/388)

# v1.1.0

## Enhancements

* Add Variable Set API support by @rexredinger [#305](https://github.com/hashicorp/go-tfe/pull/305)
* Add Comments API support by @alex-ikse [#355](https://github.com/hashicorp/go-tfe/pull/355)
* Add beta support for SSOTeamID to `Team`, `TeamCreateOptions`, `TeamUpdateOptions` by @xlgmokha [#364](https://github.com/hashicorp/go-tfe/pull/364)

# v1.0.0

## Breaking Changes
* Renamed methods named Generate to Create for `AgentTokens`, `OrganizationTokens`, `TeamTokens`, `UserTokens` by @sebasslash [#327](https://github.com/hashicorp/go-tfe/pull/327)
* Methods that express an action on a relationship have been prefixed with a verb, e.g `Current()` is now `ReadCurrent()` by @sebasslash [#327](https://github.com/hashicorp/go-tfe/pull/327)
* All list option structs are now pointers @uturunku1 [#309](https://github.com/hashicorp/go-tfe/pull/309)
* All errors have been refactored into constants in `errors.go` @uturunku1 [#310](https://github.com/hashicorp/go-tfe/pull/310)
* The `ID` field in Create/Update option structs has been renamed to `Type` in accordance with the JSON:API spec by @omarismail, @uturunku1 [#190](https://github.com/hashicorp/go-tfe/pull/190), [#323](https://github.com/hashicorp/go-tfe/pull/323), [#332](https://github.com/hashicorp/go-tfe/pull/332)
* Nested URL params (consisting of an organization, module and provider name) used to identify a `RegistryModule` have been refactored into a struct `RegistryModuleID` by @sebasslash [#337](https://github.com/hashicorp/go-tfe/pull/337)


## Enhancements
* Added missing include fields for `AdminRuns`, `AgentPools`, `ConfigurationVersions`, `OAuthClients`, `Organizations`, `PolicyChecks`, `PolicySets`, `Policies` and `RunTriggers` by @uturunku1 [#339](https://github.com/hashicorp/go-tfe/pull/339)
* Cleanup documentation and improve consistency by @uturunku1 [#331](https://github.com/hashicorp/go-tfe/pull/331)
* Add more linters to our CI pipeline by @sebasslash [#326](https://github.com/hashicorp/go-tfe/pull/326)
* Resolve `TFE_HOSTNAME` as fallback for `TFE_ADDRESS` by @sebasslash [#340](https://github.com/hashicorp/go-tfe/pull/326)
* Adds a `fetching` status to `RunStatus` and adds the `Archive` method to the ConfigurationVersions interface by @mpminardi [#338](https://github.com/hashicorp/go-tfe/pull/338)
* Added a `Download` method to the `ConfigurationVersions` interface by @tylerwolf [#358](https://github.com/hashicorp/go-tfe/pull/358)
* API Coverage documentation by @laurenolivia [#334](https://github.com/hashicorp/go-tfe/pull/334)

## Bug Fixes
* Fixed invalid memory address error when `AdminSMTPSettingsUpdateOptions.Auth` field is empty and accessed by @uturunku1 [#335](https://github.com/hashicorp/go-tfe/pull/335)
