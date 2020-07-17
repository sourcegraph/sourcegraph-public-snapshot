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
} from './graphql/schema'

export type Maybe<T> = T | null
export type Exact<T extends { [key: string]: any }> = { [K in keyof T]: T[K] }
export interface SharedGraphQlOperations {
    /** shared/src/backend/repo.ts*/
    ResolveRawRepoName: (variables: ResolveRawRepoNameVariables) => ResolveRawRepoNameResult

    /** shared/src/extensions/helpers.ts*/
    Extensions: (variables: ExtensionsVariables) => ExtensionsResult

    /** shared/src/settings/edit.ts*/
    EditSettings: (variables: EditSettingsVariables) => EditSettingsResult

    /** shared/src/settings/edit.ts*/
    OverwriteSettings: (variables: OverwriteSettingsVariables) => OverwriteSettingsResult
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

export type ResolveRawRepoNameVariables = Exact<{
    repoName: Scalars['String']
}>

export interface ResolveRawRepoNameResult {
    repository?: Maybe<{ uri: string; mirrorInfo: { cloned: boolean } }>
}

export type ExtensionsVariables = Exact<{
    first: Scalars['Int']
    prioritizeExtensionIDs: Scalars['String'][]
}>

export interface ExtensionsResult {
    extensionRegistry: {
        extensions: {
            nodes: {
                id: string
                extensionID: string
                url: string
                viewerCanAdminister: boolean
                manifest?: Maybe<{ raw: string }>
            }[]
        }
    }
}

export type EditSettingsVariables = Exact<{
    subject: Scalars['ID']
    lastID?: Maybe<Scalars['Int']>
    edit: ConfigurationEdit
}>

export interface EditSettingsResult {
    configurationMutation?: Maybe<{ editConfiguration?: Maybe<{ empty?: Maybe<{ alwaysNil?: Maybe<string> }> }> }>
}

export type OverwriteSettingsVariables = Exact<{
    subject: Scalars['ID']
    lastID?: Maybe<Scalars['Int']>
    contents: Scalars['String']
}>

export interface OverwriteSettingsResult {
    settingsMutation?: Maybe<{ overwriteSettings?: Maybe<{ empty?: Maybe<{ alwaysNil?: Maybe<string> }> }> }>
}
