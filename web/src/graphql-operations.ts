import {
    CampaignState,
    BackgroundProcessState,
    OrganizationInvitationResponseType,
    EventSource,
    RepositoryPermission,
    SymbolKind,
    DiagnosticSeverity,
    ExternalServiceKind,
    GitRefType,
    GitObjectType,
    GitRefOrder,
    LSIFUploadState,
    LSIFIndexState,
    DiffHunkLineType,
    ChangesetState,
    ChangesetReviewState,
    ChangesetCheckState,
    RepositoryOrderBy,
    UserActivePeriod,
    SearchVersion,
    SearchPatternType,
    AlertType,
    UserEvent,
} from '../../shared/src/graphql/schema'

export type Maybe<T> = T | null
export type Exact<T extends { [key: string]: any }> = { [K in keyof T]: T[K] }
export interface WebGraphQlOperations {
    /** web/src/auth.ts*/
    CurrentAuthState: (variables: CurrentAuthStateVariables) => CurrentAuthStateResult

    /** web/src/enterprise/campaigns/detail/AddChangesetForm.tsx*/
    RepositoryID: (variables: RepositoryIDVariables) => RepositoryIDResult

    /** web/src/enterprise/campaigns/detail/AddChangesetForm.tsx*/
    CreateChangeset: (variables: CreateChangesetVariables) => CreateChangesetResult

    /** web/src/enterprise/campaigns/detail/AddChangesetForm.tsx*/
    AddChangeSetToCampaign: (variables: AddChangeSetToCampaignVariables) => AddChangeSetToCampaignResult

    /** web/src/enterprise/campaigns/detail/backend.ts*/
    UpdateCampaign: (variables: UpdateCampaignVariables) => UpdateCampaignResult

    /** web/src/enterprise/campaigns/detail/backend.ts*/
    CreateCampaign: (variables: CreateCampaignVariables) => CreateCampaignResult

    /** web/src/enterprise/campaigns/detail/backend.ts*/
    RetryCampaignChangesets: (variables: RetryCampaignChangesetsVariables) => RetryCampaignChangesetsResult

    /** web/src/enterprise/campaigns/detail/backend.ts*/
    PublishCampaignChangesets: (variables: PublishCampaignChangesetsVariables) => PublishCampaignChangesetsResult

    /** web/src/enterprise/campaigns/detail/backend.ts*/
    CloseCampaign: (variables: CloseCampaignVariables) => CloseCampaignResult

    /** web/src/enterprise/campaigns/detail/backend.ts*/
    DeleteCampaign: (variables: DeleteCampaignVariables) => DeleteCampaignResult

    /** web/src/enterprise/campaigns/detail/backend.ts*/
    CampaignByID: (variables: CampaignByIDVariables) => CampaignByIDResult

    /** web/src/enterprise/campaigns/detail/backend.ts*/
    PatchSetByID: (variables: PatchSetByIDVariables) => PatchSetByIDResult

    /** web/src/enterprise/campaigns/detail/backend.ts*/
    CampaignChangesets: (variables: CampaignChangesetsVariables) => CampaignChangesetsResult

    /** web/src/enterprise/campaigns/detail/backend.ts*/
    CampaignPatches: (variables: CampaignPatchesVariables) => CampaignPatchesResult

    /** web/src/enterprise/campaigns/detail/backend.ts*/
    PatchSetPatches: (variables: PatchSetPatchesVariables) => PatchSetPatchesResult

    /** web/src/enterprise/campaigns/detail/backend.ts*/
    PublishChangeset: (variables: PublishChangesetVariables) => PublishChangesetResult

    /** web/src/enterprise/campaigns/detail/backend.ts*/
    SyncChangeset: (variables: SyncChangesetVariables) => SyncChangesetResult

    /** web/src/enterprise/campaigns/detail/backend.ts*/
    ExternalChangesetFileDiffs: (variables: ExternalChangesetFileDiffsVariables) => ExternalChangesetFileDiffsResult

    /** web/src/enterprise/campaigns/detail/backend.ts*/
    PatchFileDiffs: (variables: PatchFileDiffsVariables) => PatchFileDiffsResult

    /** web/src/enterprise/campaigns/global/list/backend.ts*/
    Campaigns: (variables: CampaignsVariables) => CampaignsResult

    /** web/src/enterprise/campaigns/global/list/backend.ts*/
    CampaignsCount: (variables: CampaignsCountVariables) => CampaignsCountResult

    /** web/src/enterprise/codeintel/backend.tsx*/
    LsifUploads: (variables: LsifUploadsVariables) => LsifUploadsResult

    /** web/src/enterprise/codeintel/backend.tsx*/
    LsifUploadsWithRepo: (variables: LsifUploadsWithRepoVariables) => LsifUploadsWithRepoResult

    /** web/src/enterprise/codeintel/backend.tsx*/
    LsifUpload: (variables: LsifUploadVariables) => LsifUploadResult

    /** web/src/enterprise/codeintel/backend.tsx*/
    DeleteLsifUpload: (variables: DeleteLsifUploadVariables) => DeleteLsifUploadResult

    /** web/src/enterprise/codeintel/backend.tsx*/
    LsifIndexes: (variables: LsifIndexesVariables) => LsifIndexesResult

    /** web/src/enterprise/codeintel/backend.tsx*/
    LsifIndexesWithRepo: (variables: LsifIndexesWithRepoVariables) => LsifIndexesWithRepoResult

    /** web/src/enterprise/codeintel/backend.tsx*/
    LsifIndex: (variables: LsifIndexVariables) => LsifIndexResult

    /** web/src/enterprise/codeintel/backend.tsx*/
    DeleteLsifIndex: (variables: DeleteLsifIndexVariables) => DeleteLsifIndexResult

    /** web/src/enterprise/dotcom/productPlans/ProductPlanFormControl.tsx*/
    ProductPlans: (variables: ProductPlansVariables) => ProductPlansResult

    /** web/src/enterprise/extensions/explore/ExtensionsExploreSection.tsx*/
    ExploreExtensions: (variables: ExploreExtensionsVariables) => ExploreExtensionsResult

    /** web/src/enterprise/extensions/extension/RegistryExtensionManagePage.tsx*/
    UpdateRegistryExtension: (variables: UpdateRegistryExtensionVariables) => UpdateRegistryExtensionResult

    /** web/src/enterprise/extensions/extension/RegistryExtensionNewReleasePage.tsx*/
    PublishRegistryExtension: (variables: PublishRegistryExtensionVariables) => PublishRegistryExtensionResult

    /** web/src/enterprise/extensions/registry/RegistryNewExtensionPage.tsx*/
    CreateRegistryExtension: (variables: CreateRegistryExtensionVariables) => CreateRegistryExtensionResult

    /** web/src/enterprise/extensions/registry/backend.ts*/
    DeleteRegistryExtension: (variables: DeleteRegistryExtensionVariables) => DeleteRegistryExtensionResult

    /** web/src/enterprise/extensions/registry/backend.ts*/
    ViewerRegistryPublishers: (variables: ViewerRegistryPublishersVariables) => ViewerRegistryPublishersResult

    /** web/src/enterprise/namespaces/backend.ts*/
    ViewerNamespaces: (variables: ViewerNamespacesVariables) => ViewerNamespacesResult

    /** web/src/enterprise/repo/settings/backend.tsx*/
    LsifUploadsForRepo: (variables: LsifUploadsForRepoVariables) => LsifUploadsForRepoResult

    /** web/src/enterprise/repo/settings/backend.tsx*/
    LsifUploadForRepo: (variables: LsifUploadForRepoVariables) => LsifUploadForRepoResult

    /** web/src/enterprise/repo/settings/backend.tsx*/
    DeleteLsifUploadForRepo: (variables: DeleteLsifUploadForRepoVariables) => DeleteLsifUploadForRepoResult

    /** web/src/enterprise/repo/settings/backend.tsx*/
    LsifIndexesForRepo: (variables: LsifIndexesForRepoVariables) => LsifIndexesForRepoResult

    /** web/src/enterprise/repo/settings/backend.tsx*/
    LsifIndexForRepo: (variables: LsifIndexForRepoVariables) => LsifIndexForRepoResult

    /** web/src/enterprise/repo/settings/backend.tsx*/
    DeleteLsifIndexForRepo: (variables: DeleteLsifIndexForRepoVariables) => DeleteLsifIndexForRepoResult

    /** web/src/enterprise/search/stats/backend.ts*/
    SearchResultsStats: (variables: SearchResultsStatsVariables) => SearchResultsStatsResult

    /** web/src/enterprise/site-admin/SiteAdminAuthenticationProvidersPage.tsx*/
    AuthProviders: (variables: AuthProvidersVariables) => AuthProvidersResult

    /** web/src/enterprise/site-admin/SiteAdminExternalAccountsPage.tsx*/
    ExternalAccounts: (variables: ExternalAccountsVariables) => ExternalAccountsResult

    /** web/src/enterprise/site-admin/SiteAdminRegistryExtensionsPage.tsx*/
    SiteAdminRegistryExtensions: (variables: SiteAdminRegistryExtensionsVariables) => SiteAdminRegistryExtensionsResult

    /** web/src/enterprise/site-admin/backend.ts*/
    SiteAdminLsifUpload: (variables: SiteAdminLsifUploadVariables) => SiteAdminLsifUploadResult

    /** web/src/enterprise/site-admin/dotcom/customers/SiteAdminCustomerBillingLink.tsx*/
    SetCustomerBilling: (variables: SetCustomerBillingVariables) => SetCustomerBillingResult

    /** web/src/enterprise/site-admin/dotcom/customers/SiteAdminCustomersPage.tsx*/
    Customers: (variables: CustomersVariables) => CustomersResult

    /** web/src/enterprise/site-admin/dotcom/productSubscriptions/SiteAdminCreateProductSubscriptionPage.tsx*/
    CreateProductSubscription: (variables: CreateProductSubscriptionVariables) => CreateProductSubscriptionResult

    /** web/src/enterprise/site-admin/dotcom/productSubscriptions/SiteAdminCreateProductSubscriptionPage.tsx*/
    ProductSubscriptionAccounts: (variables: ProductSubscriptionAccountsVariables) => ProductSubscriptionAccountsResult

    /** web/src/enterprise/site-admin/dotcom/productSubscriptions/SiteAdminGenerateProductLicenseForSubscriptionForm.tsx*/
    GenerateProductLicenseForSubscription: (
        variables: GenerateProductLicenseForSubscriptionVariables
    ) => GenerateProductLicenseForSubscriptionResult

    /** web/src/enterprise/site-admin/dotcom/productSubscriptions/SiteAdminProductLicensesPage.tsx*/
    DotComProductLicenses: (variables: DotComProductLicensesVariables) => DotComProductLicensesResult

    /** web/src/enterprise/site-admin/dotcom/productSubscriptions/SiteAdminProductSubscriptionBillingLink.tsx*/
    SetProductSubscriptionBilling: (
        variables: SetProductSubscriptionBillingVariables
    ) => SetProductSubscriptionBillingResult

    /** web/src/enterprise/site-admin/dotcom/productSubscriptions/SiteAdminProductSubscriptionPage.tsx*/
    DotComProductSubscription: (variables: DotComProductSubscriptionVariables) => DotComProductSubscriptionResult

    /** web/src/enterprise/site-admin/dotcom/productSubscriptions/SiteAdminProductSubscriptionPage.tsx*/
    ProductLicenses: (variables: ProductLicensesVariables) => ProductLicensesResult

    /** web/src/enterprise/site-admin/dotcom/productSubscriptions/SiteAdminProductSubscriptionPage.tsx*/
    ArchiveProductSubscription: (variables: ArchiveProductSubscriptionVariables) => ArchiveProductSubscriptionResult

    /** web/src/enterprise/site-admin/dotcom/productSubscriptions/SiteAdminProductSubscriptionsPage.tsx*/
    ProductSubscriptionsDotCom: (variables: ProductSubscriptionsDotComVariables) => ProductSubscriptionsDotComResult

    /** web/src/enterprise/site-admin/productSubscription/ProductSubscriptionStatus.tsx*/
    ProductLicenseInfo: (variables: ProductLicenseInfoVariables) => ProductLicenseInfoResult

    /** web/src/enterprise/user/productSubscriptions/NewProductSubscriptionPaymentSection.tsx*/
    PreviewProductSubscriptionInvoice: (
        variables: PreviewProductSubscriptionInvoiceVariables
    ) => PreviewProductSubscriptionInvoiceResult

    /** web/src/enterprise/user/productSubscriptions/UserSubscriptionsEditProductSubscriptionPage.tsx*/
    ProductSubscriptionOnEditPage: (
        variables: ProductSubscriptionOnEditPageVariables
    ) => ProductSubscriptionOnEditPageResult

    /** web/src/enterprise/user/productSubscriptions/UserSubscriptionsEditProductSubscriptionPage.tsx*/
    UpdatePaidProductSubscription: (
        variables: UpdatePaidProductSubscriptionVariables
    ) => UpdatePaidProductSubscriptionResult

    /** web/src/enterprise/user/productSubscriptions/UserSubscriptionsNewProductSubscriptionPage.tsx*/
    CreatePaidProductSubscription: (
        variables: CreatePaidProductSubscriptionVariables
    ) => CreatePaidProductSubscriptionResult

    /** web/src/enterprise/user/productSubscriptions/UserSubscriptionsProductSubscriptionPage.tsx*/
    ProductSubscription: (variables: ProductSubscriptionVariables) => ProductSubscriptionResult

    /** web/src/enterprise/user/productSubscriptions/UserSubscriptionsProductSubscriptionsPage.tsx*/
    ProductSubscriptions: (variables: ProductSubscriptionsVariables) => ProductSubscriptionsResult

    /** web/src/enterprise/user/settings/ExternalAccountNode.tsx*/
    DeleteExternalAccount: (variables: DeleteExternalAccountVariables) => DeleteExternalAccountResult

    /** web/src/enterprise/user/settings/UserSettingsExternalAccountsPage.tsx*/
    UserExternalAccounts: (variables: UserExternalAccountsVariables) => UserExternalAccountsResult

    /** web/src/extensions/ExtensionsList.tsx*/
    RegistryExtensions: (variables: RegistryExtensionsVariables) => RegistryExtensionsResult

    /** web/src/extensions/extension/ExtensionArea.tsx*/
    RegistryExtension: (variables: RegistryExtensionVariables) => RegistryExtensionResult

    /** web/src/marketing/backend.tsx*/
    SubmitSurvey: (variables: SubmitSurveyVariables) => SubmitSurveyResult

    /** web/src/marketing/backend.tsx*/
    FetchSurveyResponses: (variables: FetchSurveyResponsesVariables) => FetchSurveyResponsesResult

    /** web/src/marketing/backend.tsx*/
    FetchAllUsersWithSurveyResponses: (
        variables: FetchAllUsersWithSurveyResponsesVariables
    ) => FetchAllUsersWithSurveyResponsesResult

    /** web/src/marketing/backend.tsx*/
    FetchSurveyResponseAggregates: (
        variables: FetchSurveyResponseAggregatesVariables
    ) => FetchSurveyResponseAggregatesResult

    /** web/src/marketing/backend.tsx*/
    RequestTrial: (variables: RequestTrialVariables) => RequestTrialResult

    /** web/src/nav/StatusMessagesNavItem.tsx*/
    StatusMessages: (variables: StatusMessagesVariables) => StatusMessagesResult

    /** web/src/org/area/OrgArea.tsx*/
    Organization: (variables: OrganizationVariables) => OrganizationResult

    /** web/src/org/area/OrgInvitationPage.tsx*/
    RespondToOrganizationInvitation: (
        variables: RespondToOrganizationInvitationVariables
    ) => RespondToOrganizationInvitationResult

    /** web/src/org/area/OrgMembersPage.tsx*/
    OrganizationMembers: (variables: OrganizationMembersVariables) => OrganizationMembersResult

    /** web/src/org/backend.tsx*/
    createOrganization: (variables: createOrganizationVariables) => createOrganizationResult

    /** web/src/org/backend.tsx*/
    removeUserFromOrganization: (variables: removeUserFromOrganizationVariables) => removeUserFromOrganizationResult

    /** web/src/org/backend.tsx*/
    UpdateOrganization: (variables: UpdateOrganizationVariables) => UpdateOrganizationResult

    /** web/src/org/invite/InviteForm.tsx*/
    InviteUserToOrganization: (variables: InviteUserToOrganizationVariables) => InviteUserToOrganizationResult

    /** web/src/org/invite/InviteForm.tsx*/
    AddUserToOrganization: (variables: AddUserToOrganizationVariables) => AddUserToOrganizationResult

    /** web/src/platform/context.ts*/
    ViewerSettings: (variables: ViewerSettingsVariables) => ViewerSettingsResult

    /** web/src/repo/GitReference.tsx*/
    RepositoryGitRefs: (variables: RepositoryGitRefsVariables) => RepositoryGitRefsResult

    /** web/src/repo/RepoRevisionSidebarCommits.tsx*/
    FetchCommits: (variables: FetchCommitsVariables) => FetchCommitsResult

    /** web/src/repo/RepositoriesPopover.tsx*/
    RepositoriesForPopover: (variables: RepositoriesForPopoverVariables) => RepositoriesForPopoverResult

    /** web/src/repo/RevisionsPopover.tsx*/
    RepositoryGitCommit: (variables: RepositoryGitCommitVariables) => RepositoryGitCommitResult

    /** web/src/repo/backend.ts*/
    RepositoryRedirect: (variables: RepositoryRedirectVariables) => RepositoryRedirectResult

    /** web/src/repo/backend.ts*/
    ResolveRev: (variables: ResolveRevVariables) => ResolveRevResult

    /** web/src/repo/backend.ts*/
    HighlightedFile: (variables: HighlightedFileVariables) => HighlightedFileResult

    /** web/src/repo/backend.ts*/
    FileExternalLinks: (variables: FileExternalLinksVariables) => FileExternalLinksResult

    /** web/src/repo/backend.ts*/
    TreeEntries: (variables: TreeEntriesVariables) => TreeEntriesResult

    /** web/src/repo/blob/BlobPage.tsx*/
    Blob: (variables: BlobVariables) => BlobResult

    /** web/src/repo/branches/RepositoryBranchesOverviewPage.tsx*/
    RepositoryGitBranchesOverview: (
        variables: RepositoryGitBranchesOverviewVariables
    ) => RepositoryGitBranchesOverviewResult

    /** web/src/repo/commit/RepositoryCommitPage.tsx*/
    RepositoryCommit: (variables: RepositoryCommitVariables) => RepositoryCommitResult

    /** web/src/repo/commits/RepositoryCommitsPage.tsx*/
    RepositoryGitCommits: (variables: RepositoryGitCommitsVariables) => RepositoryGitCommitsResult

    /** web/src/repo/compare/RepositoryCompareCommitsPage.tsx*/
    RepositoryComparisonCommits: (variables: RepositoryComparisonCommitsVariables) => RepositoryComparisonCommitsResult

    /** web/src/repo/compare/RepositoryCompareDiffPage.tsx*/
    RepositoryComparisonDiff: (variables: RepositoryComparisonDiffVariables) => RepositoryComparisonDiffResult

    /** web/src/repo/compare/RepositoryCompareOverviewPage.tsx*/
    RepositoryComparison: (variables: RepositoryComparisonVariables) => RepositoryComparisonResult

    /** web/src/repo/explore/RepositoriesExploreSection.tsx*/
    ExploreRepositories: (variables: ExploreRepositoriesVariables) => ExploreRepositoriesResult

    /** web/src/repo/settings/RepoSettingsIndexPage.tsx*/
    RepositoryTextSearchIndex: (variables: RepositoryTextSearchIndexVariables) => RepositoryTextSearchIndexResult

    /** web/src/repo/settings/backend.tsx*/
    Repository: (variables: RepositoryVariables) => RepositoryResult

    /** web/src/repo/stats/RepositoryStatsContributorsPage.tsx*/
    RepositoryContributors: (variables: RepositoryContributorsVariables) => RepositoryContributorsResult

    /** web/src/repo/tree/TreePage.tsx*/
    TreeCommits: (variables: TreeCommitsVariables) => TreeCommitsResult

    /** web/src/search/backend.tsx*/
    Search: (variables: SearchVariables) => SearchResult

    /** web/src/search/backend.tsx*/
    RepoGroups: (variables: RepoGroupsVariables) => RepoGroupsResult

    /** web/src/search/backend.tsx*/
    SearchSuggestions: (variables: SearchSuggestionsVariables) => SearchSuggestionsResult

    /** web/src/search/backend.tsx*/
    ReposByQuery: (variables: ReposByQueryVariables) => ReposByQueryResult

    /** web/src/search/backend.tsx*/
    savedSearches: (variables: savedSearchesVariables) => savedSearchesResult

    /** web/src/search/backend.tsx*/
    SavedSearch: (variables: SavedSearchVariables) => SavedSearchResult

    /** web/src/search/backend.tsx*/
    CreateSavedSearch: (variables: CreateSavedSearchVariables) => CreateSavedSearchResult

    /** web/src/search/backend.tsx*/
    UpdateSavedSearch: (variables: UpdateSavedSearchVariables) => UpdateSavedSearchResult

    /** web/src/search/backend.tsx*/
    DeleteSavedSearch: (variables: DeleteSavedSearchVariables) => DeleteSavedSearchResult

    /** web/src/search/backend.tsx*/
    highlightCode: (variables: highlightCodeVariables) => highlightCodeResult

    /** web/src/search/backend.tsx*/
    ManyReposWarning: (variables: ManyReposWarningVariables) => ManyReposWarningResult

    /** web/src/settings/SettingsArea.tsx*/
    SettingsCascade: (variables: SettingsCascadeVariables) => SettingsCascadeResult

    /** web/src/settings/tokens/AccessTokenNode.tsx*/
    DeleteAccessToken: (variables: DeleteAccessTokenVariables) => DeleteAccessTokenResult

    /** web/src/site-admin/SiteAdminAddExternalServicePage.tsx*/
    addExternalService: (variables: addExternalServiceVariables) => addExternalServiceResult

    /** web/src/site-admin/SiteAdminExternalServicePage.tsx*/
    UpdateExternalService: (variables: UpdateExternalServiceVariables) => UpdateExternalServiceResult

    /** web/src/site-admin/SiteAdminExternalServicePage.tsx*/
    ExternalService: (variables: ExternalServiceVariables) => ExternalServiceResult

    /** web/src/site-admin/SiteAdminExternalServicesPage.tsx*/
    DeleteExternalService: (variables: DeleteExternalServiceVariables) => DeleteExternalServiceResult

    /** web/src/site-admin/SiteAdminExternalServicesPage.tsx*/
    ExternalServices: (variables: ExternalServicesVariables) => ExternalServicesResult

    /** web/src/site-admin/SiteAdminTokensPage.tsx*/
    SiteAdminAccessTokens: (variables: SiteAdminAccessTokensVariables) => SiteAdminAccessTokensResult

    /** web/src/site-admin/backend.tsx*/
    Users: (variables: UsersVariables) => UsersResult

    /** web/src/site-admin/backend.tsx*/
    Organizations: (variables: OrganizationsVariables) => OrganizationsResult

    /** web/src/site-admin/backend.tsx*/
    Repositories: (variables: RepositoriesVariables) => RepositoriesResult

    /** web/src/site-admin/backend.tsx*/
    UpdateMirrorRepository: (variables: UpdateMirrorRepositoryVariables) => UpdateMirrorRepositoryResult

    /** web/src/site-admin/backend.tsx*/
    CheckMirrorRepositoryConnection: (
        variables: CheckMirrorRepositoryConnectionVariables
    ) => CheckMirrorRepositoryConnectionResult

    /** web/src/site-admin/backend.tsx*/
    ScheduleRepositoryPermissionsSync: (
        variables: ScheduleRepositoryPermissionsSyncVariables
    ) => ScheduleRepositoryPermissionsSyncResult

    /** web/src/site-admin/backend.tsx*/
    ScheduleUserPermissionsSync: (variables: ScheduleUserPermissionsSyncVariables) => ScheduleUserPermissionsSyncResult

    /** web/src/site-admin/backend.tsx*/
    UserUsageStatistics: (variables: UserUsageStatisticsVariables) => UserUsageStatisticsResult

    /** web/src/site-admin/backend.tsx*/
    SiteUsageStatistics: (variables: SiteUsageStatisticsVariables) => SiteUsageStatisticsResult

    /** web/src/site-admin/backend.tsx*/
    Site: (variables: SiteVariables) => SiteResult

    /** web/src/site-admin/backend.tsx*/
    AllConfig: (variables: AllConfigVariables) => AllConfigResult

    /** web/src/site-admin/backend.tsx*/
    UpdateSiteConfiguration: (variables: UpdateSiteConfigurationVariables) => UpdateSiteConfigurationResult

    /** web/src/site-admin/backend.tsx*/
    ReloadSite: (variables: ReloadSiteVariables) => ReloadSiteResult

    /** web/src/site-admin/backend.tsx*/
    SetUserIsSiteAdmin: (variables: SetUserIsSiteAdminVariables) => SetUserIsSiteAdminResult

    /** web/src/site-admin/backend.tsx*/
    RandomizeUserPassword: (variables: RandomizeUserPasswordVariables) => RandomizeUserPasswordResult

    /** web/src/site-admin/backend.tsx*/
    DeleteUser: (variables: DeleteUserVariables) => DeleteUserResult

    /** web/src/site-admin/backend.tsx*/
    CreateUser: (variables: CreateUserVariables) => CreateUserResult

    /** web/src/site-admin/backend.tsx*/
    DeleteOrganization: (variables: DeleteOrganizationVariables) => DeleteOrganizationResult

    /** web/src/site-admin/backend.tsx*/
    SiteUpdateCheck: (variables: SiteUpdateCheckVariables) => SiteUpdateCheckResult

    /** web/src/site-admin/backend.tsx*/
    SiteMonitoringStatistics: (variables: SiteMonitoringStatisticsVariables) => SiteMonitoringStatisticsResult

    /** web/src/site-admin/overview/SiteAdminOverviewPage.tsx*/
    Overview: (variables: OverviewVariables) => OverviewResult

    /** web/src/site-admin/overview/SiteAdminOverviewPage.tsx*/
    WAUs: (variables: WAUsVariables) => WAUsResult

    /** web/src/site/backend.tsx*/
    SiteFlags: (variables: SiteFlagsVariables) => SiteFlagsResult

    /** web/src/symbols/backend.tsx*/
    Symbols: (variables: SymbolsVariables) => SymbolsResult

    /** web/src/tracking/withActivation.tsx*/
    SiteAdminActivationStatus: (variables: SiteAdminActivationStatusVariables) => SiteAdminActivationStatusResult

    /** web/src/tracking/withActivation.tsx*/
    ActivationStatus: (variables: ActivationStatusVariables) => ActivationStatusResult

    /** web/src/tracking/withActivation.tsx*/
    LinksForRepositories: (variables: LinksForRepositoriesVariables) => LinksForRepositoriesResult

    /** web/src/user/UserEventLogsPage.tsx*/
    UserEventLogs: (variables: UserEventLogsVariables) => UserEventLogsResult

    /** web/src/user/area/UserArea.tsx*/
    User: (variables: UserVariables) => UserResult

    /** web/src/user/settings/accessTokens/UserSettingsCreateAccessTokenPage.tsx*/
    CreateAccessToken: (variables: CreateAccessTokenVariables) => CreateAccessTokenResult

    /** web/src/user/settings/accessTokens/UserSettingsTokensPage.tsx*/
    AccessTokens: (variables: AccessTokensVariables) => AccessTokensResult

    /** web/src/user/settings/backend.tsx*/
    updateUser: (variables: updateUserVariables) => updateUserResult

    /** web/src/user/settings/backend.tsx*/
    updatePassword: (variables: updatePasswordVariables) => updatePasswordResult

    /** web/src/user/settings/backend.tsx*/
    SetUserEmailVerified: (variables: SetUserEmailVerifiedVariables) => SetUserEmailVerifiedResult

    /** web/src/user/settings/backend.tsx*/
    logUserEvent: (variables: logUserEventVariables) => logUserEventResult

    /** web/src/user/settings/backend.tsx*/
    logEvent: (variables: logEventVariables) => logEventResult

    /** web/src/user/settings/emails/AddUserEmailForm.tsx*/
    AddUserEmail: (variables: AddUserEmailVariables) => AddUserEmailResult

    /** web/src/user/settings/emails/UserSettingsEmailsPage.tsx*/
    RemoveUserEmail: (variables: RemoveUserEmailVariables) => RemoveUserEmailResult

    /** web/src/user/settings/emails/UserSettingsEmailsPage.tsx*/
    UserEmails: (variables: UserEmailsVariables) => UserEmailsResult

    /** web/src/user/settings/profile/UserSettingsProfilePage.tsx*/
    UserForProfilePage: (variables: UserForProfilePageVariables) => UserForProfilePageResult
}
/** All built-in and custom scalars, mapped to their actual values */
export interface Scalars {
    ID: string
    String: string
    Boolean: boolean
    Int: number
    Float: number
    DateTime: string
    JSONCString: string
    JSONValue: any
    GitObjectID: string
}

export interface AddExternalServiceInput {
    kind: ExternalServiceKind
    displayName: Scalars['String']
    config: Scalars['String']
}

export { AlertType }

export { BackgroundProcessState }

export { CampaignState }

export { ChangesetCheckState }

export { ChangesetReviewState }

export { ChangesetState }

export interface ConfigurationEdit {
    keyPath: KeyPathSegment[]
    value?: Maybe<Scalars['JSONValue']>
    valueIsJSONCEncodedString?: Maybe<Scalars['Boolean']>
}

export interface CreateCampaignInput {
    namespace: Scalars['ID']
    name: Scalars['String']
    description?: Maybe<Scalars['String']>
    branch?: Maybe<Scalars['String']>
    patchSet?: Maybe<Scalars['ID']>
}

export interface CreateChangesetInput {
    repository: Scalars['ID']
    externalID: Scalars['String']
}

export { DiagnosticSeverity }

export { DiffHunkLineType }

export { EventSource }

export { ExternalServiceKind }

export { GitObjectType }

export { GitRefOrder }

export { GitRefType }

export interface KeyPathSegment {
    property?: Maybe<Scalars['String']>
    index?: Maybe<Scalars['Int']>
}

export { LSIFIndexState }

export { LSIFUploadState }

export interface MarkdownOptions {
    alwaysNil?: Maybe<Scalars['String']>
}

export { OrganizationInvitationResponseType }

export interface PatchInput {
    repository: Scalars['ID']
    baseRevision: Scalars['String']
    baseRef: Scalars['String']
    patch: Scalars['String']
}

export interface ProductLicenseInput {
    tags: Scalars['String'][]
    userCount: Scalars['Int']
    expiresAt: Scalars['Int']
}

export interface ProductSubscriptionInput {
    billingPlanID: Scalars['String']
    userCount: Scalars['Int']
}

export { RepositoryOrderBy }

export { RepositoryPermission }

export { SearchPatternType }

export { SearchVersion }

export interface SettingsEdit {
    keyPath: KeyPathSegment[]
    value?: Maybe<Scalars['JSONValue']>
    valueIsJSONCEncodedString?: Maybe<Scalars['Boolean']>
}

export interface SettingsMutationGroupInput {
    subject: Scalars['ID']
    lastID?: Maybe<Scalars['Int']>
}

export interface SurveySubmissionInput {
    email?: Maybe<Scalars['String']>
    score: Scalars['Int']
    reason?: Maybe<Scalars['String']>
    better?: Maybe<Scalars['String']>
}

export { SymbolKind }

export interface UpdateCampaignInput {
    id: Scalars['ID']
    name?: Maybe<Scalars['String']>
    branch?: Maybe<Scalars['String']>
    description?: Maybe<Scalars['String']>
    patchSet?: Maybe<Scalars['ID']>
}

export interface UpdateExternalServiceInput {
    id: Scalars['ID']
    displayName?: Maybe<Scalars['String']>
    config?: Maybe<Scalars['String']>
}

export { UserActivePeriod }

export { UserEvent }

export interface UserPermission {
    bindID: Scalars['String']
    permission?: Maybe<RepositoryPermission>
}

export type CurrentAuthStateVariables = Exact<{ [key: string]: never }>

export interface CurrentAuthStateResult {
    currentUser?: Maybe<{
        __typename: 'User'
        id: string
        databaseID: number
        username: string
        avatarURL?: Maybe<string>
        email: string
        displayName?: Maybe<string>
        siteAdmin: boolean
        tags: string[]
        url: string
        settingsURL?: Maybe<string>
        viewerCanAdminister: boolean
        organizations: {
            nodes: { id: string; name: string; displayName?: Maybe<string>; url: string; settingsURL?: Maybe<string> }[]
        }
        session: { canSignOut: boolean }
    }>
}

export interface FileDiffHunkRangeFields {
    startLine: number
    lines: number
}

export interface DiffStatFields {
    added: number
    changed: number
    deleted: number
}

export interface FileDiffFields {
    __typename: 'FileDiff'
    oldPath?: Maybe<string>
    newPath?: Maybe<string>
    internalID: string
    oldFile?: Maybe<
        | { __typename: 'GitBlob'; binary: boolean; byteSize: number }
        | { __typename: 'VirtualFile'; binary: boolean; byteSize: number }
    >
    newFile?: Maybe<
        | { __typename: 'GitBlob'; binary: boolean; byteSize: number }
        | { __typename: 'VirtualFile'; binary: boolean; byteSize: number }
    >
    mostRelevantFile: { __typename: 'GitBlob'; url: string } | { __typename: 'VirtualFile'; url: string }
    hunks: {
        oldNoNewlineAt: boolean
        section?: Maybe<string>
        oldRange: { startLine: number; lines: number }
        newRange: { startLine: number; lines: number }
        highlight: { aborted: boolean; lines: { kind: DiffHunkLineType; html: string }[] }
    }[]
    stat: { added: number; changed: number; deleted: number }
}

export type RepositoryIDVariables = Exact<{
    repoName: Scalars['String']
}>

export interface RepositoryIDResult {
    repository?: Maybe<{ id: string }>
}

export type CreateChangesetVariables = Exact<{
    repositoryID: Scalars['ID']
    externalID: Scalars['String']
}>

export interface CreateChangesetResult {
    createChangesets: { id: string }[]
}

export type AddChangeSetToCampaignVariables = Exact<{
    campaignID: Scalars['ID']
    changesets: Scalars['ID'][]
}>

export interface AddChangeSetToCampaignResult {
    addChangesetsToCampaign: { id: string }
}

export interface CampaignFields {
    __typename: 'Campaign'
    id: string
    name: string
    description?: Maybe<string>
    branch?: Maybe<string>
    createdAt: string
    updatedAt: string
    closedAt?: Maybe<string>
    viewerCanAdminister: boolean
    hasUnpublishedPatches: boolean
    author: { username: string; avatarURL?: Maybe<string> }
    status: { completedCount: number; pendingCount: number; state: BackgroundProcessState; errors: string[] }
    changesets: { totalCount: number }
    patches: { totalCount: number }
    changesetCountsOverTime: {
        date: string
        merged: number
        closed: number
        openApproved: number
        openChangesRequested: number
        openPending: number
        total: number
    }[]
    diffStat: DiffStatFields
}

export interface PatchSetFields {
    __typename: 'PatchSet'
    id: string
    diffStat: DiffStatFields
    patches: { totalCount: number }
}

export type UpdateCampaignVariables = Exact<{
    update: UpdateCampaignInput
}>

export interface UpdateCampaignResult {
    updateCampaign: CampaignFields
}

export type CreateCampaignVariables = Exact<{
    input: CreateCampaignInput
}>

export interface CreateCampaignResult {
    createCampaign: { id: string; url: string }
}

export type RetryCampaignChangesetsVariables = Exact<{
    campaign: Scalars['ID']
}>

export interface RetryCampaignChangesetsResult {
    retryCampaignChangesets: CampaignFields
}

export type PublishCampaignChangesetsVariables = Exact<{
    campaign: Scalars['ID']
}>

export interface PublishCampaignChangesetsResult {
    publishCampaignChangesets: CampaignFields
}

export type CloseCampaignVariables = Exact<{
    campaign: Scalars['ID']
    closeChangesets: Scalars['Boolean']
}>

export interface CloseCampaignResult {
    closeCampaign: { id: string }
}

export type DeleteCampaignVariables = Exact<{
    campaign: Scalars['ID']
    closeChangesets: Scalars['Boolean']
}>

export interface DeleteCampaignResult {
    deleteCampaign?: Maybe<{ alwaysNil?: Maybe<string> }>
}

export type CampaignByIDVariables = Exact<{
    campaign: Scalars['ID']
}>

export interface CampaignByIDResult {
    node?: Maybe<
        | ({ __typename: 'Campaign' } & CampaignFields)
        | { __typename: 'PatchSet' }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | { __typename: 'Repository' }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | { __typename: 'LSIFUpload' }
        | { __typename: 'LSIFIndex' }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type PatchSetByIDVariables = Exact<{
    patchSet: Scalars['ID']
}>

export interface PatchSetByIDResult {
    node?: Maybe<
        | { __typename: 'Campaign' }
        | ({ __typename: 'PatchSet' } & PatchSetFields)
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | { __typename: 'Repository' }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | { __typename: 'LSIFUpload' }
        | { __typename: 'LSIFIndex' }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type CampaignChangesetsVariables = Exact<{
    campaign: Scalars['ID']
    first?: Maybe<Scalars['Int']>
    state?: Maybe<ChangesetState>
    reviewState?: Maybe<ChangesetReviewState>
    checkState?: Maybe<ChangesetCheckState>
}>

export interface CampaignChangesetsResult {
    node?: Maybe<
        | {
              __typename: 'Campaign'
              changesets: {
                  totalCount: number
                  nodes: (
                      | {
                            __typename: 'ExternalChangeset'
                            id: string
                            title: string
                            body: string
                            reviewState: ChangesetReviewState
                            checkState?: Maybe<ChangesetCheckState>
                            externalID: string
                            state: ChangesetState
                            createdAt: string
                            updatedAt: string
                            nextSyncAt?: Maybe<string>
                            labels: { text: string; description?: Maybe<string>; color: string }[]
                            repository: { id: string; name: string; url: string }
                            externalURL: { url: string }
                            diff?: Maybe<{ fileDiffs: { diffStat: DiffStatFields } }>
                            diffStat?: Maybe<{ added: number; changed: number; deleted: number }>
                        }
                      | {
                            __typename: 'HiddenExternalChangeset'
                            id: string
                            state: ChangesetState
                            createdAt: string
                            updatedAt: string
                            nextSyncAt?: Maybe<string>
                        }
                  )[]
              }
          }
        | { __typename: 'PatchSet' }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | { __typename: 'Repository' }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | { __typename: 'LSIFUpload' }
        | { __typename: 'LSIFIndex' }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type CampaignPatchesVariables = Exact<{
    campaign: Scalars['ID']
    first?: Maybe<Scalars['Int']>
}>

export interface CampaignPatchesResult {
    node?: Maybe<
        | {
              __typename: 'Campaign'
              patches: {
                  totalCount: number
                  nodes: (
                      | {
                            __typename: 'Patch'
                            id: string
                            publishable: boolean
                            publicationEnqueued: boolean
                            repository: { id: string; name: string; url: string }
                            diff: { fileDiffs: { diffStat: DiffStatFields } }
                        }
                      | { __typename: 'HiddenPatch'; id: string }
                  )[]
              }
          }
        | { __typename: 'PatchSet' }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | { __typename: 'Repository' }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | { __typename: 'LSIFUpload' }
        | { __typename: 'LSIFIndex' }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type PatchSetPatchesVariables = Exact<{
    patchSet: Scalars['ID']
    first?: Maybe<Scalars['Int']>
}>

export interface PatchSetPatchesResult {
    node?: Maybe<
        | { __typename: 'Campaign' }
        | {
              __typename: 'PatchSet'
              patches: {
                  totalCount: number
                  nodes: (
                      | {
                            __typename: 'Patch'
                            publishable: boolean
                            publicationEnqueued: boolean
                            id: string
                            repository: { id: string; name: string; url: string }
                            diff: { fileDiffs: { diffStat: DiffStatFields } }
                        }
                      | { __typename: 'HiddenPatch'; id: string }
                  )[]
              }
          }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | { __typename: 'Repository' }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | { __typename: 'LSIFUpload' }
        | { __typename: 'LSIFIndex' }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type PublishChangesetVariables = Exact<{
    patch: Scalars['ID']
}>

export interface PublishChangesetResult {
    publishChangeset: { alwaysNil?: Maybe<string> }
}

export type SyncChangesetVariables = Exact<{
    changeset: Scalars['ID']
}>

export interface SyncChangesetResult {
    syncChangeset: { alwaysNil?: Maybe<string> }
}

export type ExternalChangesetFileDiffsVariables = Exact<{
    externalChangeset: Scalars['ID']
    first?: Maybe<Scalars['Int']>
    after?: Maybe<Scalars['String']>
    isLightTheme: Scalars['Boolean']
}>

export interface ExternalChangesetFileDiffsResult {
    node?: Maybe<
        | { __typename: 'Campaign' }
        | { __typename: 'PatchSet' }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | { __typename: 'Repository' }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | { __typename: 'LSIFUpload' }
        | { __typename: 'LSIFIndex' }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | {
              __typename: 'ExternalChangeset'
              diff?: Maybe<{
                  range: {
                      base: GitRefSpecFields_GitRef_ | GitRefSpecFields_GitRevSpecExpr_ | GitRefSpecFields_GitObject_
                      head: GitRefSpecFields_GitRef_ | GitRefSpecFields_GitRevSpecExpr_ | GitRefSpecFields_GitObject_
                  }
                  fileDiffs: {
                      totalCount?: Maybe<number>
                      nodes: FileDiffFields[]
                      pageInfo: { hasNextPage: boolean; endCursor?: Maybe<string> }
                      diffStat: DiffStatFields
                  }
              }>
          }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

interface GitRefSpecFields_GitRef_ {
    __typename: 'GitRef'
    target: { oid: string }
}

interface GitRefSpecFields_GitRevSpecExpr_ {
    __typename: 'GitRevSpecExpr'
    object?: Maybe<{ oid: string }>
}

interface GitRefSpecFields_GitObject_ {
    __typename: 'GitObject'
    oid: string
}

export type GitRefSpecFields = GitRefSpecFields_GitRef_ | GitRefSpecFields_GitRevSpecExpr_ | GitRefSpecFields_GitObject_

export type PatchFileDiffsVariables = Exact<{
    patch: Scalars['ID']
    first?: Maybe<Scalars['Int']>
    after?: Maybe<Scalars['String']>
    isLightTheme: Scalars['Boolean']
}>

export interface PatchFileDiffsResult {
    node?: Maybe<
        | { __typename: 'Campaign' }
        | { __typename: 'PatchSet' }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | { __typename: 'Repository' }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | { __typename: 'LSIFUpload' }
        | { __typename: 'LSIFIndex' }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | {
              __typename: 'Patch'
              diff: {
                  fileDiffs: {
                      totalCount?: Maybe<number>
                      nodes: FileDiffFields[]
                      pageInfo: { hasNextPage: boolean; endCursor?: Maybe<string> }
                      diffStat: DiffStatFields
                  }
              }
          }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type CampaignsVariables = Exact<{
    first?: Maybe<Scalars['Int']>
    state?: Maybe<CampaignState>
    hasPatchSet?: Maybe<Scalars['Boolean']>
    viewerCanAdminister?: Maybe<Scalars['Boolean']>
}>

export interface CampaignsResult {
    campaigns: {
        totalCount: number
        nodes: {
            id: string
            name: string
            description?: Maybe<string>
            url: string
            createdAt: string
            closedAt?: Maybe<string>
            changesets: { totalCount: number; nodes: ({ state: ChangesetState } | { state: ChangesetState })[] }
            patches: { totalCount: number }
        }[]
    }
}

export type CampaignsCountVariables = Exact<{ [key: string]: never }>

export interface CampaignsCountResult {
    campaigns: { totalCount: number }
}

export type LsifUploadsVariables = Exact<{
    state?: Maybe<LSIFUploadState>
    isLatestForRepo?: Maybe<Scalars['Boolean']>
    first?: Maybe<Scalars['Int']>
    after?: Maybe<Scalars['String']>
    query?: Maybe<Scalars['String']>
}>

export interface LsifUploadsResult {
    lsifUploads: {
        totalCount?: Maybe<number>
        nodes: {
            id: string
            state: LSIFUploadState
            inputCommit: string
            inputRoot: string
            inputIndexer: string
            uploadedAt: string
            startedAt?: Maybe<string>
            finishedAt?: Maybe<string>
            placeInQueue?: Maybe<number>
            projectRoot?: Maybe<{
                url: string
                path: string
                repository: { url: string; name: string }
                commit: { url: string; oid: string; abbreviatedOID: string }
            }>
        }[]
        pageInfo: { endCursor?: Maybe<string>; hasNextPage: boolean }
    }
}

export type LsifUploadsWithRepoVariables = Exact<{
    repository: Scalars['ID']
    state?: Maybe<LSIFUploadState>
    isLatestForRepo?: Maybe<Scalars['Boolean']>
    first?: Maybe<Scalars['Int']>
    after?: Maybe<Scalars['String']>
    query?: Maybe<Scalars['String']>
}>

export interface LsifUploadsWithRepoResult {
    node?: Maybe<
        | { __typename: 'Campaign' }
        | { __typename: 'PatchSet' }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | {
              __typename: 'Repository'
              lsifUploads: {
                  totalCount?: Maybe<number>
                  nodes: {
                      id: string
                      state: LSIFUploadState
                      inputCommit: string
                      inputRoot: string
                      inputIndexer: string
                      uploadedAt: string
                      startedAt?: Maybe<string>
                      finishedAt?: Maybe<string>
                      placeInQueue?: Maybe<number>
                      projectRoot?: Maybe<{
                          url: string
                          path: string
                          repository: { url: string; name: string }
                          commit: { url: string; oid: string; abbreviatedOID: string }
                      }>
                  }[]
                  pageInfo: { endCursor?: Maybe<string>; hasNextPage: boolean }
              }
          }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | { __typename: 'LSIFUpload' }
        | { __typename: 'LSIFIndex' }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type LsifUploadVariables = Exact<{
    id: Scalars['ID']
}>

export interface LsifUploadResult {
    node?: Maybe<
        | { __typename: 'Campaign' }
        | { __typename: 'PatchSet' }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | { __typename: 'Repository' }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | {
              __typename: 'LSIFUpload'
              id: string
              inputCommit: string
              inputRoot: string
              inputIndexer: string
              state: LSIFUploadState
              failure?: Maybe<string>
              uploadedAt: string
              startedAt?: Maybe<string>
              finishedAt?: Maybe<string>
              isLatestForRepo: boolean
              placeInQueue?: Maybe<number>
              projectRoot?: Maybe<{
                  url: string
                  path: string
                  repository: { url: string; name: string }
                  commit: { url: string; oid: string; abbreviatedOID: string }
              }>
          }
        | { __typename: 'LSIFIndex' }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type DeleteLsifUploadVariables = Exact<{
    id: Scalars['ID']
}>

export interface DeleteLsifUploadResult {
    deleteLSIFUpload?: Maybe<{ alwaysNil?: Maybe<string> }>
}

export type LsifIndexesVariables = Exact<{
    state?: Maybe<LSIFIndexState>
    first?: Maybe<Scalars['Int']>
    after?: Maybe<Scalars['String']>
    query?: Maybe<Scalars['String']>
}>

export interface LsifIndexesResult {
    lsifIndexes: {
        totalCount?: Maybe<number>
        nodes: {
            id: string
            state: LSIFIndexState
            inputCommit: string
            queuedAt: string
            startedAt?: Maybe<string>
            finishedAt?: Maybe<string>
            placeInQueue?: Maybe<number>
            projectRoot?: Maybe<{
                url: string
                path: string
                repository: { url: string; name: string }
                commit: { url: string; oid: string; abbreviatedOID: string }
            }>
        }[]
        pageInfo: { endCursor?: Maybe<string>; hasNextPage: boolean }
    }
}

export type LsifIndexesWithRepoVariables = Exact<{
    repository: Scalars['ID']
    state?: Maybe<LSIFIndexState>
    first?: Maybe<Scalars['Int']>
    after?: Maybe<Scalars['String']>
    query?: Maybe<Scalars['String']>
}>

export interface LsifIndexesWithRepoResult {
    node?: Maybe<
        | { __typename: 'Campaign' }
        | { __typename: 'PatchSet' }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | {
              __typename: 'Repository'
              lsifIndexes: {
                  totalCount?: Maybe<number>
                  nodes: {
                      id: string
                      state: LSIFIndexState
                      inputCommit: string
                      queuedAt: string
                      startedAt?: Maybe<string>
                      finishedAt?: Maybe<string>
                      placeInQueue?: Maybe<number>
                      projectRoot?: Maybe<{
                          url: string
                          path: string
                          repository: { url: string; name: string }
                          commit: { url: string; oid: string; abbreviatedOID: string }
                      }>
                  }[]
                  pageInfo: { endCursor?: Maybe<string>; hasNextPage: boolean }
              }
          }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | { __typename: 'LSIFUpload' }
        | { __typename: 'LSIFIndex' }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type LsifIndexVariables = Exact<{
    id: Scalars['ID']
}>

export interface LsifIndexResult {
    node?: Maybe<
        | { __typename: 'Campaign' }
        | { __typename: 'PatchSet' }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | { __typename: 'Repository' }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | { __typename: 'LSIFUpload' }
        | {
              __typename: 'LSIFIndex'
              id: string
              inputCommit: string
              state: LSIFIndexState
              failure?: Maybe<string>
              queuedAt: string
              startedAt?: Maybe<string>
              finishedAt?: Maybe<string>
              placeInQueue?: Maybe<number>
              projectRoot?: Maybe<{
                  url: string
                  path: string
                  repository: { url: string; name: string }
                  commit: { url: string; oid: string; abbreviatedOID: string }
              }>
          }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type DeleteLsifIndexVariables = Exact<{
    id: Scalars['ID']
}>

export interface DeleteLsifIndexResult {
    deleteLSIFIndex?: Maybe<{ alwaysNil?: Maybe<string> }>
}

export type ProductPlansVariables = Exact<{ [key: string]: never }>

export interface ProductPlansResult {
    dotcom: {
        productPlans: {
            productPlanID: string
            billingPlanID: string
            name: string
            pricePerUserPerYear: number
            minQuantity?: Maybe<number>
            maxQuantity?: Maybe<number>
            tiersMode: string
            planTiers: { unitAmount: number; upTo: number; flatAmount: number }[]
        }[]
    }
}

export interface ProductSubscriptionFields {
    id: string
    name: string
    createdAt: string
    isArchived: boolean
    url: string
    account?: Maybe<{
        id: string
        username: string
        displayName?: Maybe<string>
        emails: { email: string; verified: boolean }[]
    }>
    invoiceItem?: Maybe<{ userCount: number; expiresAt: string; plan: { nameWithBrand: string } }>
    activeLicense?: Maybe<{
        licenseKey: string
        info?: Maybe<{ productNameWithBrand: string; tags: string[]; userCount: number; expiresAt: string }>
    }>
}

export type ExploreExtensionsVariables = Exact<{
    first?: Maybe<Scalars['Int']>
    prioritizeExtensionIDs?: Maybe<Scalars['String'][]>
}>

export interface ExploreExtensionsResult {
    extensionRegistry: {
        extensions: {
            nodes: {
                id: string
                extensionIDWithoutRegistry: string
                url: string
                manifest?: Maybe<{ description?: Maybe<string> }>
            }[]
        }
    }
}

export type UpdateRegistryExtensionVariables = Exact<{
    extension: Scalars['ID']
    name?: Maybe<Scalars['String']>
}>

export interface UpdateRegistryExtensionResult {
    extensionRegistry: { updateExtension: { extension: { url: string } } }
}

export type PublishRegistryExtensionVariables = Exact<{
    extensionID: Scalars['String']
    manifest: Scalars['String']
    bundle: Scalars['String']
}>

export interface PublishRegistryExtensionResult {
    extensionRegistry: { publishExtension: { extension: { url: string } } }
}

export type CreateRegistryExtensionVariables = Exact<{
    publisher: Scalars['ID']
    name: Scalars['String']
}>

export interface CreateRegistryExtensionResult {
    extensionRegistry: { createExtension: { extension: { id: string; extensionID: string; url: string } } }
}

export type DeleteRegistryExtensionVariables = Exact<{
    extension: Scalars['ID']
}>

export interface DeleteRegistryExtensionResult {
    extensionRegistry: { deleteExtension: { alwaysNil?: Maybe<string> } }
}

export type ViewerRegistryPublishersVariables = Exact<{ [key: string]: never }>

export interface ViewerRegistryPublishersResult {
    extensionRegistry: {
        localExtensionIDPrefix?: Maybe<string>
        viewerPublishers: (
            | { __typename: 'User'; id: string; username: string }
            | { __typename: 'Org'; id: string; name: string }
        )[]
    }
}

export type ViewerNamespacesVariables = Exact<{ [key: string]: never }>

export interface ViewerNamespacesResult {
    currentUser?: Maybe<{
        __typename: 'User'
        id: string
        namespaceName: string
        url: string
        organizations: { nodes: { __typename: 'Org'; id: string; namespaceName: string; url: string }[] }
    }>
}

export type LsifUploadsForRepoVariables = Exact<{
    repository: Scalars['ID']
    state?: Maybe<LSIFUploadState>
    isLatestForRepo?: Maybe<Scalars['Boolean']>
    first?: Maybe<Scalars['Int']>
    after?: Maybe<Scalars['String']>
    query?: Maybe<Scalars['String']>
}>

export interface LsifUploadsForRepoResult {
    node?: Maybe<
        | { __typename: 'Campaign' }
        | { __typename: 'PatchSet' }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | {
              __typename: 'Repository'
              lsifUploads: {
                  totalCount?: Maybe<number>
                  nodes: {
                      id: string
                      state: LSIFUploadState
                      inputCommit: string
                      inputRoot: string
                      inputIndexer: string
                      uploadedAt: string
                      startedAt?: Maybe<string>
                      finishedAt?: Maybe<string>
                      placeInQueue?: Maybe<number>
                      projectRoot?: Maybe<{
                          path: string
                          url: string
                          commit: { abbreviatedOID: string; url: string }
                      }>
                  }[]
                  pageInfo: { endCursor?: Maybe<string>; hasNextPage: boolean }
              }
          }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | { __typename: 'LSIFUpload' }
        | { __typename: 'LSIFIndex' }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type LsifUploadForRepoVariables = Exact<{
    id: Scalars['ID']
}>

export interface LsifUploadForRepoResult {
    node?: Maybe<
        | { __typename: 'Campaign' }
        | { __typename: 'PatchSet' }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | { __typename: 'Repository' }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | {
              __typename: 'LSIFUpload'
              id: string
              inputCommit: string
              inputRoot: string
              inputIndexer: string
              state: LSIFUploadState
              failure?: Maybe<string>
              uploadedAt: string
              startedAt?: Maybe<string>
              finishedAt?: Maybe<string>
              isLatestForRepo: boolean
              placeInQueue?: Maybe<number>
              projectRoot?: Maybe<{
                  path: string
                  url: string
                  commit: {
                      oid: string
                      abbreviatedOID: string
                      url: string
                      repository: { name: string; url: string }
                  }
              }>
          }
        | { __typename: 'LSIFIndex' }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type DeleteLsifUploadForRepoVariables = Exact<{
    id: Scalars['ID']
}>

export interface DeleteLsifUploadForRepoResult {
    deleteLSIFUpload?: Maybe<{ alwaysNil?: Maybe<string> }>
}

export type LsifIndexesForRepoVariables = Exact<{
    repository: Scalars['ID']
    state?: Maybe<LSIFIndexState>
    first?: Maybe<Scalars['Int']>
    after?: Maybe<Scalars['String']>
    query?: Maybe<Scalars['String']>
}>

export interface LsifIndexesForRepoResult {
    node?: Maybe<
        | { __typename: 'Campaign' }
        | { __typename: 'PatchSet' }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | {
              __typename: 'Repository'
              lsifIndexes: {
                  totalCount?: Maybe<number>
                  nodes: {
                      id: string
                      state: LSIFIndexState
                      inputCommit: string
                      queuedAt: string
                      startedAt?: Maybe<string>
                      finishedAt?: Maybe<string>
                      placeInQueue?: Maybe<number>
                      projectRoot?: Maybe<{
                          path: string
                          url: string
                          commit: { abbreviatedOID: string; url: string }
                      }>
                  }[]
                  pageInfo: { endCursor?: Maybe<string>; hasNextPage: boolean }
              }
          }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | { __typename: 'LSIFUpload' }
        | { __typename: 'LSIFIndex' }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type LsifIndexForRepoVariables = Exact<{
    id: Scalars['ID']
}>

export interface LsifIndexForRepoResult {
    node?: Maybe<
        | { __typename: 'Campaign' }
        | { __typename: 'PatchSet' }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | { __typename: 'Repository' }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | { __typename: 'LSIFUpload' }
        | {
              __typename: 'LSIFIndex'
              id: string
              inputCommit: string
              state: LSIFIndexState
              failure?: Maybe<string>
              queuedAt: string
              startedAt?: Maybe<string>
              finishedAt?: Maybe<string>
              placeInQueue?: Maybe<number>
              projectRoot?: Maybe<{
                  path: string
                  url: string
                  commit: {
                      oid: string
                      abbreviatedOID: string
                      url: string
                      repository: { name: string; url: string }
                  }
              }>
          }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type DeleteLsifIndexForRepoVariables = Exact<{
    id: Scalars['ID']
}>

export interface DeleteLsifIndexForRepoResult {
    deleteLSIFIndex?: Maybe<{ alwaysNil?: Maybe<string> }>
}

export type SearchResultsStatsVariables = Exact<{
    query: Scalars['String']
}>

export interface SearchResultsStatsResult {
    search?: Maybe<{ results: { limitHit: boolean }; stats: { languages: { name: string; totalLines: number }[] } }>
}

export interface AuthProviderFields {
    serviceType: string
    serviceID: string
    clientID: string
    displayName: string
    isBuiltin: boolean
    authenticationURL?: Maybe<string>
}

export type AuthProvidersVariables = Exact<{ [key: string]: never }>

export interface AuthProvidersResult {
    site: { authProviders: { totalCount: number; nodes: AuthProviderFields[]; pageInfo: { hasNextPage: boolean } } }
}

export type ExternalAccountsVariables = Exact<{
    first?: Maybe<Scalars['Int']>
    user?: Maybe<Scalars['ID']>
    serviceType?: Maybe<Scalars['String']>
    serviceID?: Maybe<Scalars['String']>
    clientID?: Maybe<Scalars['String']>
}>

export interface ExternalAccountsResult {
    site: {
        externalAccounts: { totalCount: number; nodes: ExternalAccountFields[]; pageInfo: { hasNextPage: boolean } }
    }
}

export type SiteAdminRegistryExtensionsVariables = Exact<{
    first?: Maybe<Scalars['Int']>
    publisher?: Maybe<Scalars['ID']>
    query?: Maybe<Scalars['String']>
    local?: Maybe<Scalars['Boolean']>
    remote?: Maybe<Scalars['Boolean']>
}>

export interface SiteAdminRegistryExtensionsResult {
    extensionRegistry: {
        extensions: {
            totalCount: number
            error?: Maybe<string>
            nodes: RegistryExtensionFields[]
            pageInfo: { hasNextPage: boolean }
        }
    }
}

export type SiteAdminLsifUploadVariables = Exact<{
    id: Scalars['ID']
}>

export interface SiteAdminLsifUploadResult {
    node?: Maybe<
        | { __typename: 'Campaign' }
        | { __typename: 'PatchSet' }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | { __typename: 'Repository' }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | { __typename: 'LSIFUpload'; projectRoot?: Maybe<{ commit: { repository: { name: string; url: string } } }> }
        | { __typename: 'LSIFIndex' }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type SetCustomerBillingVariables = Exact<{
    user: Scalars['ID']
    billingCustomerID?: Maybe<Scalars['String']>
}>

export interface SetCustomerBillingResult {
    dotcom: { setUserBilling: { alwaysNil?: Maybe<string> } }
}

export interface CustomerFields {
    id: string
    username: string
    displayName?: Maybe<string>
    urlForSiteAdminBilling?: Maybe<string>
}

export type CustomersVariables = Exact<{
    first?: Maybe<Scalars['Int']>
    query?: Maybe<Scalars['String']>
}>

export interface CustomersResult {
    users: { totalCount: number; nodes: CustomerFields[]; pageInfo: { hasNextPage: boolean } }
}

export type CreateProductSubscriptionVariables = Exact<{
    accountID: Scalars['ID']
}>

export interface CreateProductSubscriptionResult {
    dotcom: { createProductSubscription: { urlForSiteAdmin?: Maybe<string> } }
}

export type ProductSubscriptionAccountsVariables = Exact<{
    first?: Maybe<Scalars['Int']>
    query?: Maybe<Scalars['String']>
}>

export interface ProductSubscriptionAccountsResult {
    users: {
        totalCount: number
        nodes: { id: string; username: string; emails: { email: string; verified: boolean; isPrimary: boolean }[] }[]
        pageInfo: { hasNextPage: boolean }
    }
}

export type GenerateProductLicenseForSubscriptionVariables = Exact<{
    productSubscriptionID: Scalars['ID']
    license: ProductLicenseInput
}>

export interface GenerateProductLicenseForSubscriptionResult {
    dotcom: { generateProductLicenseForSubscription: { id: string } }
}

export interface ProductLicenseFields {
    id: string
    licenseKey: string
    createdAt: string
    subscription: {
        id: string
        name: string
        urlForSiteAdmin?: Maybe<string>
        account?: Maybe<{ id: string; username: string; displayName?: Maybe<string> }>
        activeLicense?: Maybe<{ id: string }>
    }
    info?: Maybe<{ productNameWithBrand: string; tags: string[]; userCount: number; expiresAt: string }>
}

export type DotComProductLicensesVariables = Exact<{
    first?: Maybe<Scalars['Int']>
    licenseKeySubstring?: Maybe<Scalars['String']>
}>

export interface DotComProductLicensesResult {
    dotcom: {
        productLicenses: { totalCount: number; nodes: ProductLicenseFields[]; pageInfo: { hasNextPage: boolean } }
    }
}

export type SetProductSubscriptionBillingVariables = Exact<{
    id: Scalars['ID']
    billingSubscriptionID?: Maybe<Scalars['String']>
}>

export interface SetProductSubscriptionBillingResult {
    dotcom: { setProductSubscriptionBilling: { alwaysNil?: Maybe<string> } }
}

export interface SiteAdminProductSubscriptionFields {
    id: string
    name: string
    createdAt: string
    isArchived: boolean
    urlForSiteAdmin?: Maybe<string>
    account?: Maybe<{
        id: string
        username: string
        displayName?: Maybe<string>
        emails: { email: string; isPrimary: boolean }[]
    }>
    invoiceItem?: Maybe<{ userCount: number; expiresAt: string; plan: { nameWithBrand: string } }>
    activeLicense?: Maybe<{
        id: string
        licenseKey: string
        createdAt: string
        info?: Maybe<{ productNameWithBrand: string; tags: string[]; userCount: number; expiresAt: string }>
    }>
}

export type DotComProductSubscriptionVariables = Exact<{
    uuid: Scalars['String']
}>

export interface DotComProductSubscriptionResult {
    dotcom: {
        productSubscription: {
            id: string
            name: string
            createdAt: string
            isArchived: boolean
            url: string
            urlForSiteAdminBilling?: Maybe<string>
            account?: Maybe<{
                id: string
                username: string
                displayName?: Maybe<string>
                emails: { email: string; verified: boolean }[]
            }>
            invoiceItem?: Maybe<{
                userCount: number
                expiresAt: string
                plan: { billingPlanID: string; name: string; nameWithBrand: string; pricePerUserPerYear: number }
            }>
            events: { id: string; date: string; title: string; description?: Maybe<string>; url?: Maybe<string> }[]
            productLicenses: {
                totalCount: number
                nodes: {
                    id: string
                    licenseKey: string
                    createdAt: string
                    info?: Maybe<{ tags: string[]; userCount: number; expiresAt: string }>
                }[]
                pageInfo: { hasNextPage: boolean }
            }
        }
    }
}

export type ProductLicensesVariables = Exact<{
    first?: Maybe<Scalars['Int']>
    subscriptionUUID: Scalars['String']
}>

export interface ProductLicensesResult {
    dotcom: {
        productSubscription: {
            productLicenses: { totalCount: number; nodes: ProductLicenseFields[]; pageInfo: { hasNextPage: boolean } }
        }
    }
}

export type ArchiveProductSubscriptionVariables = Exact<{
    id: Scalars['ID']
}>

export interface ArchiveProductSubscriptionResult {
    dotcom: { archiveProductSubscription: { alwaysNil?: Maybe<string> } }
}

export type ProductSubscriptionsDotComVariables = Exact<{
    first?: Maybe<Scalars['Int']>
    account?: Maybe<Scalars['ID']>
    query?: Maybe<Scalars['String']>
}>

export interface ProductSubscriptionsDotComResult {
    dotcom: {
        productSubscriptions: {
            totalCount: number
            nodes: SiteAdminProductSubscriptionFields[]
            pageInfo: { hasNextPage: boolean }
        }
    }
}

export type ProductLicenseInfoVariables = Exact<{ [key: string]: never }>

export interface ProductLicenseInfoResult {
    site: {
        productSubscription: {
            productNameWithBrand: string
            actualUserCount: number
            actualUserCountDate: string
            noLicenseWarningUserCount?: Maybe<number>
            license?: Maybe<{ tags: string[]; userCount: number; expiresAt: string }>
        }
    }
}

export type PreviewProductSubscriptionInvoiceVariables = Exact<{
    account?: Maybe<Scalars['ID']>
    subscriptionToUpdate?: Maybe<Scalars['ID']>
    productSubscription: ProductSubscriptionInput
}>

export interface PreviewProductSubscriptionInvoiceResult {
    dotcom: {
        previewProductSubscriptionInvoice: {
            price: number
            prorationDate?: Maybe<string>
            isDowngradeRequiringManualIntervention: boolean
            beforeInvoiceItem?: Maybe<{
                userCount: number
                expiresAt: string
                plan: { billingPlanID: string; name: string; pricePerUserPerYear: number }
            }>
            afterInvoiceItem: {
                userCount: number
                expiresAt: string
                plan: { billingPlanID: string; name: string; pricePerUserPerYear: number }
            }
        }
    }
}

export type ProductSubscriptionOnEditPageVariables = Exact<{
    uuid: Scalars['String']
}>

export interface ProductSubscriptionOnEditPageResult {
    dotcom: { productSubscription: ProductSubscriptionFieldsOnEditPage }
}

export interface ProductSubscriptionFieldsOnEditPage {
    id: string
    name: string
    url: string
    invoiceItem?: Maybe<{ userCount: number; expiresAt: string; plan: { billingPlanID: string } }>
}

export type UpdatePaidProductSubscriptionVariables = Exact<{
    subscriptionID: Scalars['ID']
    update: ProductSubscriptionInput
    paymentToken?: Maybe<Scalars['String']>
}>

export interface UpdatePaidProductSubscriptionResult {
    dotcom: { updatePaidProductSubscription: { productSubscription: { url: string } } }
}

export type CreatePaidProductSubscriptionVariables = Exact<{
    accountID: Scalars['ID']
    productSubscription: ProductSubscriptionInput
    paymentToken?: Maybe<Scalars['String']>
}>

export interface CreatePaidProductSubscriptionResult {
    dotcom: { createPaidProductSubscription: { productSubscription: { id: string; name: string; url: string } } }
}

export type ProductSubscriptionVariables = Exact<{
    uuid: Scalars['String']
}>

export interface ProductSubscriptionResult {
    dotcom: { productSubscription: ProductSubscriptionFieldsOnSubscriptionPage }
}

export interface ProductSubscriptionFieldsOnSubscriptionPage {
    id: string
    name: string
    createdAt: string
    isArchived: boolean
    url: string
    urlForSiteAdmin?: Maybe<string>
    account?: Maybe<{
        id: string
        username: string
        displayName?: Maybe<string>
        emails: { email: string; verified: boolean }[]
    }>
    invoiceItem?: Maybe<{
        userCount: number
        expiresAt: string
        plan: { billingPlanID: string; name: string; nameWithBrand: string; pricePerUserPerYear: number }
    }>
    events: { id: string; date: string; title: string; description?: Maybe<string>; url?: Maybe<string> }[]
    activeLicense?: Maybe<{
        licenseKey: string
        info?: Maybe<{ productNameWithBrand: string; tags: string[]; userCount: number; expiresAt: string }>
    }>
}

export type ProductSubscriptionsVariables = Exact<{
    first?: Maybe<Scalars['Int']>
    account?: Maybe<Scalars['ID']>
}>

export interface ProductSubscriptionsResult {
    dotcom: {
        productSubscriptions: {
            totalCount: number
            nodes: ProductSubscriptionFields[]
            pageInfo: { hasNextPage: boolean }
        }
    }
}

export interface ExternalAccountFields {
    id: string
    serviceType: string
    serviceID: string
    clientID: string
    accountID: string
    createdAt: string
    updatedAt: string
    refreshURL?: Maybe<string>
    accountData?: Maybe<any>
    user: { id: string; username: string }
}

export type DeleteExternalAccountVariables = Exact<{
    externalAccount: Scalars['ID']
}>

export interface DeleteExternalAccountResult {
    deleteExternalAccount: { alwaysNil?: Maybe<string> }
}

export type UserExternalAccountsVariables = Exact<{
    user: Scalars['ID']
    first?: Maybe<Scalars['Int']>
}>

export interface UserExternalAccountsResult {
    node?: Maybe<{
        externalAccounts: { totalCount: number; nodes: ExternalAccountFields[]; pageInfo: { hasNextPage: boolean } }
    }>
}

export type RegistryExtensionsVariables = Exact<{
    query?: Maybe<Scalars['String']>
    prioritizeExtensionIDs: Scalars['String'][]
}>

export interface RegistryExtensionsResult {
    extensionRegistry: { extensions: { error?: Maybe<string>; nodes: RegistryExtensionFieldsForList[] } }
}

export interface RegistryExtensionFieldsForList {
    id: string
    extensionID: string
    extensionIDWithoutRegistry: string
    name: string
    createdAt?: Maybe<string>
    updatedAt?: Maybe<string>
    url: string
    remoteURL?: Maybe<string>
    registryName: string
    isLocal: boolean
    isWorkInProgress: boolean
    viewerCanAdminister: boolean
    publisher?: Maybe<
        | { __typename: 'User'; id: string; username: string; displayName?: Maybe<string>; url: string }
        | { __typename: 'Org'; id: string; name: string; displayName?: Maybe<string>; url: string }
    >
    manifest?: Maybe<{ raw: string; description?: Maybe<string> }>
}

export interface RegistryExtensionFields {
    id: string
    extensionID: string
    extensionIDWithoutRegistry: string
    name: string
    createdAt?: Maybe<string>
    updatedAt?: Maybe<string>
    publishedAt?: Maybe<string>
    url: string
    remoteURL?: Maybe<string>
    registryName: string
    isLocal: boolean
    isWorkInProgress: boolean
    viewerCanAdminister: boolean
    publisher?: Maybe<
        | { __typename: 'User'; id: string; username: string; displayName?: Maybe<string>; url: string }
        | { __typename: 'Org'; id: string; name: string; displayName?: Maybe<string>; url: string }
    >
    manifest?: Maybe<{ raw: string; description?: Maybe<string> }>
}

export type RegistryExtensionVariables = Exact<{
    extensionID: Scalars['String']
}>

export interface RegistryExtensionResult {
    extensionRegistry: { extension?: Maybe<RegistryExtensionFields> }
}

export type SubmitSurveyVariables = Exact<{
    input: SurveySubmissionInput
}>

export interface SubmitSurveyResult {
    submitSurvey?: Maybe<{ alwaysNil?: Maybe<string> }>
}

export type FetchSurveyResponsesVariables = Exact<{ [key: string]: never }>

export interface FetchSurveyResponsesResult {
    surveyResponses: {
        totalCount: number
        nodes: {
            email?: Maybe<string>
            score: number
            reason?: Maybe<string>
            better?: Maybe<string>
            createdAt: string
            user?: Maybe<{ id: string; username: string; emails: { email: string }[] }>
        }[]
    }
}

export type FetchAllUsersWithSurveyResponsesVariables = Exact<{
    activePeriod?: Maybe<UserActivePeriod>
    first?: Maybe<Scalars['Int']>
    query?: Maybe<Scalars['String']>
}>

export interface FetchAllUsersWithSurveyResponsesResult {
    users: {
        totalCount: number
        nodes: {
            id: string
            username: string
            emails: { email: string }[]
            surveyResponses: { score: number; reason?: Maybe<string>; better?: Maybe<string>; createdAt: string }[]
            usageStatistics: { lastActiveTime?: Maybe<string> }
        }[]
    }
}

export type FetchSurveyResponseAggregatesVariables = Exact<{ [key: string]: never }>

export interface FetchSurveyResponseAggregatesResult {
    surveyResponses: { totalCount: number; last30DaysCount: number; averageScore: number; netPromoterScore: number }
}

export type RequestTrialVariables = Exact<{
    email: Scalars['String']
}>

export interface RequestTrialResult {
    requestTrial?: Maybe<{ alwaysNil?: Maybe<string> }>
}

export type StatusMessagesVariables = Exact<{ [key: string]: never }>

export interface StatusMessagesResult {
    statusMessages: (
        | { __typename: 'CloningProgress'; message: string }
        | {
              __typename: 'ExternalServiceSyncError'
              message: string
              externalService: { id: string; displayName: string }
          }
        | { __typename: 'SyncError'; message: string }
    )[]
}

export type OrganizationVariables = Exact<{
    name: Scalars['String']
}>

export interface OrganizationResult {
    organization?: Maybe<{
        __typename: 'Org'
        id: string
        name: string
        displayName?: Maybe<string>
        url: string
        settingsURL?: Maybe<string>
        viewerIsMember: boolean
        viewerCanAdminister: boolean
        createdAt: string
        viewerPendingInvitation?: Maybe<{
            id: string
            respondURL?: Maybe<string>
            sender: { username: string; displayName?: Maybe<string>; avatarURL?: Maybe<string>; createdAt: string }
        }>
    }>
}

export type RespondToOrganizationInvitationVariables = Exact<{
    organizationInvitation: Scalars['ID']
    responseType: OrganizationInvitationResponseType
}>

export interface RespondToOrganizationInvitationResult {
    respondToOrganizationInvitation: { alwaysNil?: Maybe<string> }
}

export type OrganizationMembersVariables = Exact<{
    id: Scalars['ID']
}>

export interface OrganizationMembersResult {
    node?: Maybe<{
        viewerCanAdminister: boolean
        members: {
            totalCount: number
            nodes: { id: string; username: string; displayName?: Maybe<string>; avatarURL?: Maybe<string> }[]
        }
    }>
}

export type createOrganizationVariables = Exact<{
    name: Scalars['String']
    displayName?: Maybe<Scalars['String']>
}>

export interface createOrganizationResult {
    createOrganization: { id: string; name: string }
}

export type removeUserFromOrganizationVariables = Exact<{
    user: Scalars['ID']
    organization: Scalars['ID']
}>

export interface removeUserFromOrganizationResult {
    removeUserFromOrganization?: Maybe<{ alwaysNil?: Maybe<string> }>
}

export type UpdateOrganizationVariables = Exact<{
    id: Scalars['ID']
    displayName?: Maybe<Scalars['String']>
}>

export interface UpdateOrganizationResult {
    updateOrganization: { id: string }
}

export type InviteUserToOrganizationVariables = Exact<{
    organization: Scalars['ID']
    username: Scalars['String']
}>

export interface InviteUserToOrganizationResult {
    inviteUserToOrganization: { sentInvitationEmail: boolean; invitationURL: string }
}

export type AddUserToOrganizationVariables = Exact<{
    organization: Scalars['ID']
    username: Scalars['String']
}>

export interface AddUserToOrganizationResult {
    addUserToOrganization: { alwaysNil?: Maybe<string> }
}

export interface SettingsCascadeFields {
    final: string
    subjects: (
        | {
              __typename: 'User'
              id: string
              username: string
              displayName?: Maybe<string>
              settingsURL?: Maybe<string>
              viewerCanAdminister: boolean
              latestSettings?: Maybe<{ id: number; contents: string }>
          }
        | {
              __typename: 'Org'
              id: string
              name: string
              displayName?: Maybe<string>
              settingsURL?: Maybe<string>
              viewerCanAdminister: boolean
              latestSettings?: Maybe<{ id: number; contents: string }>
          }
        | {
              __typename: 'Site'
              id: string
              siteID: string
              settingsURL?: Maybe<string>
              viewerCanAdminister: boolean
              latestSettings?: Maybe<{ id: number; contents: string }>
          }
        | {
              __typename: 'DefaultSettings'
              settingsURL?: Maybe<string>
              viewerCanAdminister: boolean
              latestSettings?: Maybe<{ id: number; contents: string }>
          }
    )[]
}

export type ViewerSettingsVariables = Exact<{ [key: string]: never }>

export interface ViewerSettingsResult {
    viewerSettings: SettingsCascadeFields
}

export interface GitRefFields {
    id: string
    displayName: string
    name: string
    abbrevName: string
    url: string
    target: {
        commit?: Maybe<{
            author: SignatureFieldsForReferences
            committer?: Maybe<SignatureFieldsForReferences>
            behindAhead: { behind: number; ahead: number }
        }>
    }
}

export interface SignatureFieldsForReferences {
    date: string
    person: { displayName: string; user?: Maybe<{ username: string }> }
}

export type RepositoryGitRefsVariables = Exact<{
    repo: Scalars['ID']
    first?: Maybe<Scalars['Int']>
    query?: Maybe<Scalars['String']>
    type: GitRefType
    withBehindAhead: Scalars['Boolean']
}>

export interface RepositoryGitRefsResult {
    node?: Maybe<{ gitRefs: { totalCount: number; nodes: GitRefFields[]; pageInfo: { hasNextPage: boolean } } }>
}

export type FetchCommitsVariables = Exact<{
    repo: Scalars['ID']
    revision: Scalars['String']
    first?: Maybe<Scalars['Int']>
    currentPath?: Maybe<Scalars['String']>
    query?: Maybe<Scalars['String']>
}>

export interface FetchCommitsResult {
    node?: Maybe<
        | { __typename: 'Campaign' }
        | { __typename: 'PatchSet' }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | { __typename: 'Repository'; commit?: Maybe<{ ancestors: { nodes: GitCommitFields[] } }> }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | { __typename: 'LSIFUpload' }
        | { __typename: 'LSIFIndex' }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type RepositoriesForPopoverVariables = Exact<{
    first?: Maybe<Scalars['Int']>
    query?: Maybe<Scalars['String']>
}>

export interface RepositoriesForPopoverResult {
    repositories: {
        totalCount?: Maybe<number>
        nodes: { id: string; name: string }[]
        pageInfo: { hasNextPage: boolean }
    }
}

export type RepositoryGitCommitVariables = Exact<{
    repo: Scalars['ID']
    first?: Maybe<Scalars['Int']>
    revision: Scalars['String']
    query?: Maybe<Scalars['String']>
}>

export interface RepositoryGitCommitResult {
    node?: Maybe<
        | { __typename: 'Campaign' }
        | { __typename: 'PatchSet' }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | {
              __typename: 'Repository'
              commit?: Maybe<{
                  ancestors: {
                      nodes: {
                          id: string
                          oid: string
                          abbreviatedOID: string
                          subject: string
                          author: { date: string; person: { name: string; avatarURL: string } }
                      }[]
                      pageInfo: { hasNextPage: boolean }
                  }
              }>
          }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | { __typename: 'LSIFUpload' }
        | { __typename: 'LSIFIndex' }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type RepositoryRedirectVariables = Exact<{
    repoName: Scalars['String']
}>

export interface RepositoryRedirectResult {
    repositoryRedirect?: Maybe<
        | {
              __typename: 'Repository'
              id: string
              name: string
              url: string
              description: string
              viewerCanAdminister: boolean
              externalURLs: { url: string; serviceType?: Maybe<string> }[]
              defaultBranch?: Maybe<{ displayName: string }>
          }
        | { __typename: 'Redirect'; url: string }
    >
}

export type ResolveRevVariables = Exact<{
    repoName: Scalars['String']
    revision: Scalars['String']
}>

export interface ResolveRevResult {
    repositoryRedirect?: Maybe<
        | {
              __typename: 'Repository'
              mirrorInfo: { cloneInProgress: boolean; cloneProgress?: Maybe<string>; cloned: boolean }
              commit?: Maybe<{ oid: string; tree?: Maybe<{ url: string }> }>
              defaultBranch?: Maybe<{ abbrevName: string }>
          }
        | { __typename: 'Redirect'; url: string }
    >
}

export type HighlightedFileVariables = Exact<{
    repoName: Scalars['String']
    commitID: Scalars['String']
    filePath: Scalars['String']
    disableTimeout: Scalars['Boolean']
    isLightTheme: Scalars['Boolean']
}>

export interface HighlightedFileResult {
    repository?: Maybe<{
        commit?: Maybe<{
            file?: Maybe<
                | { isDirectory: boolean; richHTML: string; highlight: { aborted: boolean; html: string } }
                | { isDirectory: boolean; richHTML: string; highlight: { aborted: boolean; html: string } }
            >
        }>
    }>
}

export type FileExternalLinksVariables = Exact<{
    repoName: Scalars['String']
    revision: Scalars['String']
    filePath: Scalars['String']
}>

export interface FileExternalLinksResult {
    repository?: Maybe<{
        commit?: Maybe<{
            file?: Maybe<
                | { externalURLs: { url: string; serviceType?: Maybe<string> }[] }
                | { externalURLs: { url: string; serviceType?: Maybe<string> }[] }
            >
        }>
    }>
}

export type TreeEntriesVariables = Exact<{
    repoName: Scalars['String']
    revision: Scalars['String']
    commitID: Scalars['String']
    filePath: Scalars['String']
    first?: Maybe<Scalars['Int']>
}>

export interface TreeEntriesResult {
    repository?: Maybe<{
        commit?: Maybe<{
            tree?: Maybe<{
                isRoot: boolean
                url: string
                entries: (
                    | {
                          name: string
                          path: string
                          isDirectory: boolean
                          url: string
                          isSingleChild: boolean
                          submodule?: Maybe<{ url: string; commit: string }>
                      }
                    | {
                          name: string
                          path: string
                          isDirectory: boolean
                          url: string
                          isSingleChild: boolean
                          submodule?: Maybe<{ url: string; commit: string }>
                      }
                )[]
            }>
        }>
    }>
}

export type BlobVariables = Exact<{
    repoName: Scalars['String']
    commitID: Scalars['String']
    filePath: Scalars['String']
    isLightTheme: Scalars['Boolean']
    disableTimeout: Scalars['Boolean']
}>

export interface BlobResult {
    repository?: Maybe<{
        commit?: Maybe<{
            file?: Maybe<
                | { content: string; richHTML: string; highlight: { aborted: boolean; html: string } }
                | { content: string; richHTML: string; highlight: { aborted: boolean; html: string } }
            >
        }>
    }>
}

export type RepositoryGitBranchesOverviewVariables = Exact<{
    repo: Scalars['ID']
    first: Scalars['Int']
    withBehindAhead: Scalars['Boolean']
}>

export interface RepositoryGitBranchesOverviewResult {
    node?: Maybe<{
        defaultBranch?: Maybe<GitRefFields>
        gitRefs: { nodes: GitRefFields[]; pageInfo: { hasNextPage: boolean } }
    }>
}

export type RepositoryCommitVariables = Exact<{
    repo: Scalars['ID']
    revspec: Scalars['String']
}>

export interface RepositoryCommitResult {
    node?: Maybe<{ commit?: Maybe<{ __typename: 'GitCommit' } & GitCommitFields> }>
}

export interface GitCommitFields {
    id: string
    oid: string
    abbreviatedOID: string
    message: string
    subject: string
    body?: Maybe<string>
    url: string
    canonicalURL: string
    author: SignatureFields
    committer?: Maybe<SignatureFields>
    parents: { oid: string; abbreviatedOID: string; url: string }[]
    externalURLs: { url: string; serviceType?: Maybe<string> }[]
    tree?: Maybe<{ canonicalURL: string }>
}

export interface SignatureFields {
    date: string
    person: {
        avatarURL: string
        name: string
        email: string
        displayName: string
        user?: Maybe<{ id: string; username: string; url: string }>
    }
}

export type RepositoryGitCommitsVariables = Exact<{
    repo: Scalars['ID']
    revspec: Scalars['String']
    first?: Maybe<Scalars['Int']>
    query?: Maybe<Scalars['String']>
}>

export interface RepositoryGitCommitsResult {
    node?: Maybe<{ commit?: Maybe<{ ancestors: { nodes: GitCommitFields[]; pageInfo: { hasNextPage: boolean } } }> }>
}

export type RepositoryComparisonCommitsVariables = Exact<{
    repo: Scalars['ID']
    base?: Maybe<Scalars['String']>
    head?: Maybe<Scalars['String']>
    first?: Maybe<Scalars['Int']>
}>

export interface RepositoryComparisonCommitsResult {
    node?: Maybe<{ comparison: { commits: { nodes: GitCommitFields[]; pageInfo: { hasNextPage: boolean } } } }>
}

export type RepositoryComparisonDiffVariables = Exact<{
    repo: Scalars['ID']
    base?: Maybe<Scalars['String']>
    head?: Maybe<Scalars['String']>
    first?: Maybe<Scalars['Int']>
    after?: Maybe<Scalars['String']>
    isLightTheme: Scalars['Boolean']
}>

export interface RepositoryComparisonDiffResult {
    node?: Maybe<{
        comparison: {
            fileDiffs: {
                totalCount?: Maybe<number>
                nodes: FileDiffFields[]
                pageInfo: { endCursor?: Maybe<string>; hasNextPage: boolean }
                diffStat: DiffStatFields
            }
        }
    }>
}

export type RepositoryComparisonVariables = Exact<{
    repo: Scalars['ID']
    base?: Maybe<Scalars['String']>
    head?: Maybe<Scalars['String']>
}>

export interface RepositoryComparisonResult {
    node?: Maybe<{
        comparison: {
            range: {
                expr: string
                baseRevSpec: { object?: Maybe<{ oid: string }> }
                headRevSpec: { object?: Maybe<{ oid: string }> }
            }
        }
    }>
}

export type ExploreRepositoriesVariables = Exact<{
    first?: Maybe<Scalars['Int']>
    names?: Maybe<Scalars['String'][]>
}>

export interface ExploreRepositoriesResult {
    repositories: { nodes: { name: string; description: string; url: string }[] }
}

export type RepositoryTextSearchIndexVariables = Exact<{
    id: Scalars['ID']
}>

export interface RepositoryTextSearchIndexResult {
    node?: Maybe<{
        textSearchIndex?: Maybe<{
            status?: Maybe<{
                updatedAt: string
                contentByteSize: number
                contentFilesCount: number
                indexByteSize: number
                indexShardsCount: number
            }>
            refs: {
                indexed: boolean
                current: boolean
                ref: { displayName: string; url: string }
                indexedCommit?: Maybe<{ oid: string; abbreviatedOID: string; commit?: Maybe<{ url: string }> }>
            }[]
        }>
    }>
}

export type RepositoryVariables = Exact<{
    name: Scalars['String']
}>

export interface RepositoryResult {
    repository?: Maybe<{
        id: string
        name: string
        isPrivate: boolean
        viewerCanAdminister: boolean
        mirrorInfo: {
            remoteURL: string
            cloneInProgress: boolean
            cloneProgress?: Maybe<string>
            cloned: boolean
            updatedAt?: Maybe<string>
            updateSchedule?: Maybe<{ due: string; index: number; total: number }>
            updateQueue?: Maybe<{ updating: boolean; index: number; total: number }>
        }
        externalServices: { nodes: { id: string; kind: ExternalServiceKind; displayName: string }[] }
        permissionsInfo?: Maybe<{ syncedAt?: Maybe<string>; updatedAt: string }>
    }>
}

export type RepositoryContributorsVariables = Exact<{
    repo: Scalars['ID']
    first?: Maybe<Scalars['Int']>
    revisionRange?: Maybe<Scalars['String']>
    after?: Maybe<Scalars['String']>
    path?: Maybe<Scalars['String']>
}>

export interface RepositoryContributorsResult {
    node?: Maybe<{
        contributors: {
            totalCount: number
            nodes: {
                count: number
                person: {
                    name: string
                    displayName: string
                    email: string
                    avatarURL: string
                    user?: Maybe<{ username: string; url: string }>
                }
                commits: {
                    nodes: {
                        oid: string
                        abbreviatedOID: string
                        url: string
                        subject: string
                        author: { date: string }
                    }[]
                }
            }[]
            pageInfo: { hasNextPage: boolean }
        }
    }>
}

export type TreeCommitsVariables = Exact<{
    repo: Scalars['ID']
    revspec: Scalars['String']
    first?: Maybe<Scalars['Int']>
    filePath?: Maybe<Scalars['String']>
    after?: Maybe<Scalars['String']>
}>

export interface TreeCommitsResult {
    node?: Maybe<
        | { __typename: 'Campaign' }
        | { __typename: 'PatchSet' }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | {
              __typename: 'Repository'
              commit?: Maybe<{ ancestors: { nodes: GitCommitFields[]; pageInfo: { hasNextPage: boolean } } }>
          }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | { __typename: 'LSIFUpload' }
        | { __typename: 'LSIFIndex' }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type SearchVariables = Exact<{
    query: Scalars['String']
    version: SearchVersion
    patternType: SearchPatternType
    useCodemod: Scalars['Boolean']
    versionContext?: Maybe<Scalars['String']>
}>

export interface SearchResult {
    search?: Maybe<{
        results: {
            __typename: 'SearchResults'
            limitHit: boolean
            matchCount: number
            approximateResultCount: string
            repositoriesCount: number
            indexUnavailable: boolean
            elapsedMilliseconds: number
            missing: { name: string }[]
            cloning: { name: string }[]
            timedout: { name: string }[]
            dynamicFilters: { value: string; label: string; count: number; limitHit: boolean; kind: string }[]
            results: (
                | {
                      __typename: 'FileMatch'
                      limitHit: boolean
                      file: { path: string; url: string; commit: { oid: string } }
                      repository: { name: string; url: string }
                      revSpec?: Maybe<
                          | { __typename: 'GitRef'; displayName: string; url: string }
                          | {
                                __typename: 'GitRevSpecExpr'
                                expr: string
                                object?: Maybe<{ commit?: Maybe<{ url: string }> }>
                            }
                          | { __typename: 'GitObject'; abbreviatedOID: string; commit?: Maybe<{ url: string }> }
                      >
                      symbols: { name: string; containerName?: Maybe<string>; url: string; kind: SymbolKind }[]
                      lineMatches: { preview: string; lineNumber: number; offsetAndLengths: number[][] }[]
                  }
                | {
                      __typename: 'CommitSearchResult'
                      url: string
                      icon: string
                      label: { html: string }
                      detail: { html: string }
                      matches: {
                          url: string
                          body: { text: string; html: string }
                          highlights: { line: number; character: number; length: number }[]
                      }[]
                  }
                | {
                      __typename: 'Repository'
                      id: string
                      name: string
                      url: string
                      icon: string
                      label: { html: string }
                      detail: { html: string }
                      matches: {
                          url: string
                          body: { text: string; html: string }
                          highlights: { line: number; character: number; length: number }[]
                      }[]
                  }
                | {
                      __typename: 'CodemodResult'
                      url: string
                      icon: string
                      label: { html: string }
                      detail: { html: string }
                      matches: {
                          url: string
                          body: { text: string; html: string }
                          highlights: { line: number; character: number; length: number }[]
                      }[]
                  }
            )[]
            alert?: Maybe<{
                title: string
                description?: Maybe<string>
                proposedQueries?: Maybe<{ description?: Maybe<string>; query: string }[]>
            }>
        }
    }>
}

export type RepoGroupsVariables = Exact<{ [key: string]: never }>

export interface RepoGroupsResult {
    repoGroups: { __typename: 'RepoGroup'; name: string }[]
}

export type SearchSuggestionsVariables = Exact<{
    query: Scalars['String']
}>

export interface SearchSuggestionsResult {
    search?: Maybe<{
        suggestions: (
            | { __typename: 'Repository'; name: string }
            | {
                  __typename: 'File'
                  path: string
                  name: string
                  isDirectory: boolean
                  url: string
                  repository: { name: string }
              }
            | {
                  __typename: 'Symbol'
                  name: string
                  containerName?: Maybe<string>
                  url: string
                  kind: SymbolKind
                  location: { resource: { path: string; repository: { name: string } } }
              }
            | { __typename: 'Language' }
        )[]
    }>
}

export type ReposByQueryVariables = Exact<{
    query: Scalars['String']
}>

export interface ReposByQueryResult {
    search?: Maybe<{ results: { repositories: { name: string; url: string }[] } }>
}

export interface SavedSearchFields {
    id: string
    description: string
    notify: boolean
    notifySlack: boolean
    query: string
    slackWebhookURL?: Maybe<string>
    namespace: { id: string } | { id: string }
}

export type savedSearchesVariables = Exact<{ [key: string]: never }>

export interface savedSearchesResult {
    savedSearches: SavedSearchFields[]
}

export type SavedSearchVariables = Exact<{
    id: Scalars['ID']
}>

export interface SavedSearchResult {
    node?: Maybe<{
        id: string
        description: string
        query: string
        notify: boolean
        notifySlack: boolean
        slackWebhookURL?: Maybe<string>
        namespace: { id: string } | { id: string }
    }>
}

export type CreateSavedSearchVariables = Exact<{
    description: Scalars['String']
    query: Scalars['String']
    notifyOwner: Scalars['Boolean']
    notifySlack: Scalars['Boolean']
    userID?: Maybe<Scalars['ID']>
    orgID?: Maybe<Scalars['ID']>
}>

export interface CreateSavedSearchResult {
    createSavedSearch: SavedSearchFields
}

export type UpdateSavedSearchVariables = Exact<{
    id: Scalars['ID']
    description: Scalars['String']
    query: Scalars['String']
    notifyOwner: Scalars['Boolean']
    notifySlack: Scalars['Boolean']
    userID?: Maybe<Scalars['ID']>
    orgID?: Maybe<Scalars['ID']>
}>

export interface UpdateSavedSearchResult {
    updateSavedSearch: SavedSearchFields
}

export type DeleteSavedSearchVariables = Exact<{
    id: Scalars['ID']
}>

export interface DeleteSavedSearchResult {
    deleteSavedSearch?: Maybe<{ alwaysNil?: Maybe<string> }>
}

export type highlightCodeVariables = Exact<{
    code: Scalars['String']
    fuzzyLanguage: Scalars['String']
    disableTimeout: Scalars['Boolean']
    isLightTheme: Scalars['Boolean']
}>

export interface highlightCodeResult {
    highlightCode: string
}

export type ManyReposWarningVariables = Exact<{
    first?: Maybe<Scalars['Int']>
}>

export interface ManyReposWarningResult {
    repositories: { nodes: { id: string }[] }
}

export type SettingsCascadeVariables = Exact<{
    subject: Scalars['ID']
}>

export interface SettingsCascadeResult {
    settingsSubject?: Maybe<
        | {
              settingsCascade: {
                  subjects: (
                      | { latestSettings?: Maybe<{ id: number; contents: string }> }
                      | { latestSettings?: Maybe<{ id: number; contents: string }> }
                      | { latestSettings?: Maybe<{ id: number; contents: string }> }
                      | { latestSettings?: Maybe<{ id: number; contents: string }> }
                  )[]
              }
          }
        | {
              settingsCascade: {
                  subjects: (
                      | { latestSettings?: Maybe<{ id: number; contents: string }> }
                      | { latestSettings?: Maybe<{ id: number; contents: string }> }
                      | { latestSettings?: Maybe<{ id: number; contents: string }> }
                      | { latestSettings?: Maybe<{ id: number; contents: string }> }
                  )[]
              }
          }
        | {
              settingsCascade: {
                  subjects: (
                      | { latestSettings?: Maybe<{ id: number; contents: string }> }
                      | { latestSettings?: Maybe<{ id: number; contents: string }> }
                      | { latestSettings?: Maybe<{ id: number; contents: string }> }
                      | { latestSettings?: Maybe<{ id: number; contents: string }> }
                  )[]
              }
          }
        | {
              settingsCascade: {
                  subjects: (
                      | { latestSettings?: Maybe<{ id: number; contents: string }> }
                      | { latestSettings?: Maybe<{ id: number; contents: string }> }
                      | { latestSettings?: Maybe<{ id: number; contents: string }> }
                      | { latestSettings?: Maybe<{ id: number; contents: string }> }
                  )[]
              }
          }
    >
}

export interface AccessTokenFields {
    id: string
    scopes: string[]
    note: string
    createdAt: string
    lastUsedAt?: Maybe<string>
    subject: { username: string }
    creator: { username: string }
}

export type DeleteAccessTokenVariables = Exact<{
    tokenID: Scalars['ID']
}>

export interface DeleteAccessTokenResult {
    deleteAccessToken: { alwaysNil?: Maybe<string> }
}

export type addExternalServiceVariables = Exact<{
    input: AddExternalServiceInput
}>

export interface addExternalServiceResult {
    addExternalService: { id: string; kind: ExternalServiceKind; displayName: string; warning?: Maybe<string> }
}

export interface externalServiceFields {
    id: string
    kind: ExternalServiceKind
    displayName: string
    config: string
    warning?: Maybe<string>
    webhookURL?: Maybe<string>
}

export type UpdateExternalServiceVariables = Exact<{
    input: UpdateExternalServiceInput
}>

export interface UpdateExternalServiceResult {
    updateExternalService: externalServiceFields
}

export type ExternalServiceVariables = Exact<{
    id: Scalars['ID']
}>

export interface ExternalServiceResult {
    node?: Maybe<
        | { __typename: 'Campaign' }
        | { __typename: 'PatchSet' }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | { __typename: 'Repository' }
        | { __typename: 'GitCommit' }
        | ({ __typename: 'ExternalService' } & externalServiceFields)
        | { __typename: 'GitRef' }
        | { __typename: 'LSIFUpload' }
        | { __typename: 'LSIFIndex' }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type DeleteExternalServiceVariables = Exact<{
    externalService: Scalars['ID']
}>

export interface DeleteExternalServiceResult {
    deleteExternalService: { alwaysNil?: Maybe<string> }
}

export type ExternalServicesVariables = Exact<{
    first?: Maybe<Scalars['Int']>
}>

export interface ExternalServicesResult {
    externalServices: {
        totalCount: number
        nodes: { id: string; kind: ExternalServiceKind; displayName: string; config: string }[]
        pageInfo: { hasNextPage: boolean }
    }
}

export type SiteAdminAccessTokensVariables = Exact<{
    first?: Maybe<Scalars['Int']>
}>

export interface SiteAdminAccessTokensResult {
    site: { accessTokens: { totalCount: number; nodes: AccessTokenFields[]; pageInfo: { hasNextPage: boolean } } }
}

export type UsersVariables = Exact<{
    first?: Maybe<Scalars['Int']>
    query?: Maybe<Scalars['String']>
}>

export interface UsersResult {
    users: {
        totalCount: number
        nodes: {
            id: string
            username: string
            displayName?: Maybe<string>
            createdAt: string
            siteAdmin: boolean
            emails: {
                email: string
                verified: boolean
                verificationPending: boolean
                viewerCanManuallyVerify: boolean
            }[]
            latestSettings?: Maybe<{ createdAt: string; contents: string }>
            organizations: { nodes: { name: string }[] }
        }[]
    }
}

export type OrganizationsVariables = Exact<{
    first?: Maybe<Scalars['Int']>
    query?: Maybe<Scalars['String']>
}>

export interface OrganizationsResult {
    organizations: {
        totalCount: number
        nodes: {
            id: string
            name: string
            displayName?: Maybe<string>
            createdAt: string
            latestSettings?: Maybe<{ createdAt: string; contents: string }>
            members: { totalCount: number }
        }[]
    }
}

export type RepositoriesVariables = Exact<{
    first?: Maybe<Scalars['Int']>
    query?: Maybe<Scalars['String']>
    cloned?: Maybe<Scalars['Boolean']>
    cloneInProgress?: Maybe<Scalars['Boolean']>
    notCloned?: Maybe<Scalars['Boolean']>
    indexed?: Maybe<Scalars['Boolean']>
    notIndexed?: Maybe<Scalars['Boolean']>
}>

export interface RepositoriesResult {
    repositories: {
        totalCount?: Maybe<number>
        nodes: {
            id: string
            name: string
            createdAt: string
            viewerCanAdminister: boolean
            url: string
            mirrorInfo: { cloned: boolean; cloneInProgress: boolean; updatedAt?: Maybe<string> }
        }[]
        pageInfo: { hasNextPage: boolean }
    }
}

export type UpdateMirrorRepositoryVariables = Exact<{
    repository: Scalars['ID']
}>

export interface UpdateMirrorRepositoryResult {
    updateMirrorRepository: { alwaysNil?: Maybe<string> }
}

export type CheckMirrorRepositoryConnectionVariables = Exact<{
    repository?: Maybe<Scalars['ID']>
    name?: Maybe<Scalars['String']>
}>

export interface CheckMirrorRepositoryConnectionResult {
    checkMirrorRepositoryConnection: { error?: Maybe<string> }
}

export type ScheduleRepositoryPermissionsSyncVariables = Exact<{
    repository: Scalars['ID']
}>

export interface ScheduleRepositoryPermissionsSyncResult {
    scheduleRepositoryPermissionsSync: { alwaysNil?: Maybe<string> }
}

export type ScheduleUserPermissionsSyncVariables = Exact<{
    user: Scalars['ID']
}>

export interface ScheduleUserPermissionsSyncResult {
    scheduleUserPermissionsSync: { alwaysNil?: Maybe<string> }
}

export type UserUsageStatisticsVariables = Exact<{
    activePeriod?: Maybe<UserActivePeriod>
    query?: Maybe<Scalars['String']>
    first?: Maybe<Scalars['Int']>
}>

export interface UserUsageStatisticsResult {
    users: {
        totalCount: number
        nodes: {
            id: string
            username: string
            usageStatistics: {
                searchQueries: number
                pageViews: number
                codeIntelligenceActions: number
                lastActiveTime?: Maybe<string>
                lastActiveCodeHostIntegrationTime?: Maybe<string>
            }
        }[]
    }
}

export type SiteUsageStatisticsVariables = Exact<{ [key: string]: never }>

export interface SiteUsageStatisticsResult {
    site: {
        usageStatistics: {
            daus: { userCount: number; registeredUserCount: number; anonymousUserCount: number; startTime: string }[]
            waus: { userCount: number; registeredUserCount: number; anonymousUserCount: number; startTime: string }[]
            maus: { userCount: number; registeredUserCount: number; anonymousUserCount: number; startTime: string }[]
        }
    }
}

export type SiteVariables = Exact<{ [key: string]: never }>

export interface SiteResult {
    site: {
        id: string
        canReloadSite: boolean
        configuration: { id: number; effectiveContents: string; validationMessages: string[] }
    }
}

export type AllConfigVariables = Exact<{
    first?: Maybe<Scalars['Int']>
}>

export interface AllConfigResult {
    site: {
        id: string
        configuration: { id: number; effectiveContents: string }
        latestSettings?: Maybe<{ contents: string }>
        settingsCascade: { final: string }
    }
    externalServices: {
        nodes: {
            id: string
            kind: ExternalServiceKind
            displayName: string
            config: string
            createdAt: string
            updatedAt: string
            warning?: Maybe<string>
        }[]
    }
    viewerSettings: SiteAdminSettingsCascadeFields
}

export interface SiteAdminSettingsCascadeFields {
    final: string
    subjects: (
        | { __typename: 'User'; settingsURL?: Maybe<string>; latestSettings?: Maybe<{ id: number; contents: string }> }
        | { __typename: 'Org'; settingsURL?: Maybe<string>; latestSettings?: Maybe<{ id: number; contents: string }> }
        | { __typename: 'Site'; settingsURL?: Maybe<string>; latestSettings?: Maybe<{ id: number; contents: string }> }
        | {
              __typename: 'DefaultSettings'
              settingsURL?: Maybe<string>
              latestSettings?: Maybe<{ id: number; contents: string }>
          }
    )[]
}

export type UpdateSiteConfigurationVariables = Exact<{
    lastID: Scalars['Int']
    input: Scalars['String']
}>

export interface UpdateSiteConfigurationResult {
    updateSiteConfiguration: boolean
}

export type ReloadSiteVariables = Exact<{ [key: string]: never }>

export interface ReloadSiteResult {
    reloadSite?: Maybe<{ alwaysNil?: Maybe<string> }>
}

export type SetUserIsSiteAdminVariables = Exact<{
    userID: Scalars['ID']
    siteAdmin: Scalars['Boolean']
}>

export interface SetUserIsSiteAdminResult {
    setUserIsSiteAdmin?: Maybe<{ alwaysNil?: Maybe<string> }>
}

export type RandomizeUserPasswordVariables = Exact<{
    user: Scalars['ID']
}>

export interface RandomizeUserPasswordResult {
    randomizeUserPassword: { resetPasswordURL?: Maybe<string> }
}

export type DeleteUserVariables = Exact<{
    user: Scalars['ID']
    hard?: Maybe<Scalars['Boolean']>
}>

export interface DeleteUserResult {
    deleteUser?: Maybe<{ alwaysNil?: Maybe<string> }>
}

export type CreateUserVariables = Exact<{
    username: Scalars['String']
    email?: Maybe<Scalars['String']>
}>

export interface CreateUserResult {
    createUser: { resetPasswordURL?: Maybe<string> }
}

export type DeleteOrganizationVariables = Exact<{
    organization: Scalars['ID']
}>

export interface DeleteOrganizationResult {
    deleteOrganization?: Maybe<{ alwaysNil?: Maybe<string> }>
}

export type SiteUpdateCheckVariables = Exact<{ [key: string]: never }>

export interface SiteUpdateCheckResult {
    site: {
        buildVersion: string
        productVersion: string
        updateCheck: {
            pending: boolean
            checkedAt?: Maybe<string>
            errorMessage?: Maybe<string>
            updateVersionAvailable?: Maybe<string>
        }
    }
}

export type SiteMonitoringStatisticsVariables = Exact<{
    days: Scalars['Int']
}>

export interface SiteMonitoringStatisticsResult {
    site: {
        monitoringStatistics: { alerts: { serviceName: string; name: string; timestamp: string; average: number }[] }
    }
}

export type OverviewVariables = Exact<{ [key: string]: never }>

export interface OverviewResult {
    repositories: { totalCount?: Maybe<number> }
    users: { totalCount: number }
    organizations: { totalCount: number }
    surveyResponses: { totalCount: number; averageScore: number }
}

export type WAUsVariables = Exact<{ [key: string]: never }>

export interface WAUsResult {
    site: {
        usageStatistics: {
            waus: { userCount: number; registeredUserCount: number; anonymousUserCount: number; startTime: string }[]
        }
    }
}

export type SiteFlagsVariables = Exact<{ [key: string]: never }>

export interface SiteFlagsResult {
    site: {
        needsRepositoryConfiguration: boolean
        freeUsersExceeded: boolean
        disableBuiltInSearches: boolean
        sendsEmailVerificationEmails: boolean
        productVersion: string
        alerts: { type: AlertType; message: string; isDismissibleWithKey?: Maybe<string> }[]
        authProviders: {
            nodes: {
                serviceType: string
                serviceID: string
                clientID: string
                displayName: string
                isBuiltin: boolean
                authenticationURL?: Maybe<string>
            }[]
        }
        updateCheck: {
            pending: boolean
            checkedAt?: Maybe<string>
            errorMessage?: Maybe<string>
            updateVersionAvailable?: Maybe<string>
        }
        productSubscription: { noLicenseWarningUserCount?: Maybe<number>; license?: Maybe<{ expiresAt: string }> }
    }
}

export type SymbolsVariables = Exact<{
    repo: Scalars['ID']
    revision: Scalars['String']
    first?: Maybe<Scalars['Int']>
    query?: Maybe<Scalars['String']>
    includePatterns?: Maybe<Scalars['String'][]>
}>

export interface SymbolsResult {
    node?: Maybe<
        | { __typename: 'Campaign' }
        | { __typename: 'PatchSet' }
        | { __typename: 'User' }
        | { __typename: 'Org' }
        | { __typename: 'OrganizationInvitation' }
        | { __typename: 'AccessToken' }
        | { __typename: 'ExternalAccount' }
        | {
              __typename: 'Repository'
              commit?: Maybe<{
                  symbols: {
                      pageInfo: { hasNextPage: boolean }
                      nodes: {
                          name: string
                          containerName?: Maybe<string>
                          kind: SymbolKind
                          language: string
                          url: string
                          location: {
                              resource: { path: string }
                              range?: Maybe<{
                                  start: { line: number; character: number }
                                  end: { line: number; character: number }
                              }>
                          }
                      }[]
                  }
              }>
          }
        | { __typename: 'GitCommit' }
        | { __typename: 'ExternalService' }
        | { __typename: 'GitRef' }
        | { __typename: 'LSIFUpload' }
        | { __typename: 'LSIFIndex' }
        | { __typename: 'SavedSearch' }
        | { __typename: 'VersionContext' }
        | { __typename: 'RegistryExtension' }
        | { __typename: 'ProductSubscription' }
        | { __typename: 'ProductLicense' }
        | { __typename: 'ExternalChangeset' }
        | { __typename: 'ChangesetEvent' }
        | { __typename: 'Patch' }
        | { __typename: 'HiddenPatch' }
        | { __typename: 'HiddenExternalChangeset' }
    >
}

export type SiteAdminActivationStatusVariables = Exact<{ [key: string]: never }>

export interface SiteAdminActivationStatusResult {
    externalServices: { totalCount: number }
    repositories: { totalCount?: Maybe<number> }
    viewerSettings: { final: string }
    users: { totalCount: number }
    currentUser?: Maybe<{
        usageStatistics: { searchQueries: number; findReferencesActions: number; codeIntelligenceActions: number }
    }>
}

export type ActivationStatusVariables = Exact<{ [key: string]: never }>

export interface ActivationStatusResult {
    currentUser?: Maybe<{
        usageStatistics: { searchQueries: number; findReferencesActions: number; codeIntelligenceActions: number }
    }>
}

export type LinksForRepositoriesVariables = Exact<{ [key: string]: never }>

export interface LinksForRepositoriesResult {
    repositories: { nodes: { url: string; gitRefs: { totalCount: number } }[] }
}

export type UserEventLogsVariables = Exact<{
    user: Scalars['ID']
    first?: Maybe<Scalars['Int']>
}>

export interface UserEventLogsResult {
    node?: Maybe<{
        eventLogs: {
            totalCount: number
            nodes: { name: string; source: EventSource; url: string; timestamp: string }[]
            pageInfo: { hasNextPage: boolean }
        }
    }>
}

export type UserVariables = Exact<{
    username: Scalars['String']
    siteAdmin: Scalars['Boolean']
}>

export interface UserResult {
    user?: Maybe<{
        __typename: 'User'
        id: string
        username: string
        displayName?: Maybe<string>
        url: string
        settingsURL?: Maybe<string>
        avatarURL?: Maybe<string>
        viewerCanAdminister: boolean
        siteAdmin: boolean
        builtinAuth: boolean
        createdAt: string
        emails: { email: string; verified: boolean }[]
        organizations: { nodes: { id: string; displayName?: Maybe<string>; name: string }[] }
        permissionsInfo?: Maybe<{ syncedAt?: Maybe<string>; updatedAt: string }>
    }>
}

export type CreateAccessTokenVariables = Exact<{
    user: Scalars['ID']
    scopes: Scalars['String'][]
    note: Scalars['String']
}>

export interface CreateAccessTokenResult {
    createAccessToken: { id: string; token: string }
}

export type AccessTokensVariables = Exact<{
    user: Scalars['ID']
    first?: Maybe<Scalars['Int']>
}>

export interface AccessTokensResult {
    node?: Maybe<{
        accessTokens: { totalCount: number; nodes: AccessTokenFields[]; pageInfo: { hasNextPage: boolean } }
    }>
}

export type updateUserVariables = Exact<{
    user: Scalars['ID']
    username?: Maybe<Scalars['String']>
    displayName?: Maybe<Scalars['String']>
    avatarURL?: Maybe<Scalars['String']>
}>

export interface updateUserResult {
    updateUser: { alwaysNil?: Maybe<string> }
}

export type updatePasswordVariables = Exact<{
    oldPassword: Scalars['String']
    newPassword: Scalars['String']
}>

export interface updatePasswordResult {
    updatePassword?: Maybe<{ alwaysNil?: Maybe<string> }>
}

export type SetUserEmailVerifiedVariables = Exact<{
    user: Scalars['ID']
    email: Scalars['String']
    verified: Scalars['Boolean']
}>

export interface SetUserEmailVerifiedResult {
    setUserEmailVerified: { alwaysNil?: Maybe<string> }
}

export type logUserEventVariables = Exact<{
    event: UserEvent
    userCookieID: Scalars['String']
}>

export interface logUserEventResult {
    logUserEvent?: Maybe<{ alwaysNil?: Maybe<string> }>
}

export type logEventVariables = Exact<{
    event: Scalars['String']
    userCookieID: Scalars['String']
    url: Scalars['String']
    source: EventSource
    argument?: Maybe<Scalars['String']>
}>

export interface logEventResult {
    logEvent?: Maybe<{ alwaysNil?: Maybe<string> }>
}

export type AddUserEmailVariables = Exact<{
    user: Scalars['ID']
    email: Scalars['String']
}>

export interface AddUserEmailResult {
    addUserEmail: { alwaysNil?: Maybe<string> }
}

export type RemoveUserEmailVariables = Exact<{
    user: Scalars['ID']
    email: Scalars['String']
}>

export interface RemoveUserEmailResult {
    removeUserEmail: { alwaysNil?: Maybe<string> }
}

export type UserEmailsVariables = Exact<{
    user: Scalars['ID']
}>

export interface UserEmailsResult {
    node?: Maybe<{
        emails: {
            email: string
            isPrimary: boolean
            verified: boolean
            verificationPending: boolean
            viewerCanManuallyVerify: boolean
        }[]
    }>
}

export type UserForProfilePageVariables = Exact<{
    user: Scalars['ID']
}>

export interface UserForProfilePageResult {
    node?: Maybe<{
        id: string
        username: string
        displayName?: Maybe<string>
        avatarURL?: Maybe<string>
        viewerCanChangeUsername: boolean
    }>
}
