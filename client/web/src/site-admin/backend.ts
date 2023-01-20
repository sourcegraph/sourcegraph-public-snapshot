import { QueryResult } from '@apollo/client'
import { parse as parseJSONC } from 'jsonc-parser'
import { Observable } from 'rxjs'
import { map, mapTo, tap } from 'rxjs/operators'

import { repeatUntil, resetAllMemoizationCaches } from '@sourcegraph/common'
import { createInvalidGraphQLMutationResponseError, dataOrThrowErrors, gql, useQuery } from '@sourcegraph/http-client'
import { Settings } from '@sourcegraph/shared/src/settings/settings'

import { mutateGraphQL, queryGraphQL, requestGraphQL } from '../backend/graphql'
import {
    useShowMorePagination,
    UseShowMorePaginationResult,
} from '../components/FilteredConnection/hooks/useShowMorePagination'
import {
    AllConfigResult,
    CheckMirrorRepositoryConnectionResult,
    CreateUserResult,
    DeleteOrganizationResult,
    DeleteOrganizationVariables,
    ExternalServiceKind,
    FeatureFlagFields,
    FeatureFlagsResult,
    FeatureFlagsVariables,
    OrganizationsConnectionFields,
    OrganizationsResult,
    OrganizationsVariables,
    OutOfBandMigrationFields,
    OutOfBandMigrationsResult,
    OutOfBandMigrationsVariables,
    RandomizeUserPasswordResult,
    ReloadSiteResult,
    ReloadSiteVariables,
    RepositoriesResult,
    RepositoriesVariables,
    RepositoryOrderBy,
    Scalars,
    ScheduleRepositoryPermissionsSyncResult,
    ScheduleRepositoryPermissionsSyncVariables,
    SetUserIsSiteAdminResult,
    SetUserIsSiteAdminVariables,
    SiteAdminAccessTokenConnectionFields,
    SiteAdminAccessTokensResult,
    SiteAdminAccessTokensVariables,
    SiteAdminSettingsCascadeFields,
    SiteResult,
    SiteUpdateCheckResult,
    SiteUpdateCheckVariables,
    UpdateSiteConfigurationResult,
    UpdateSiteConfigurationVariables,
    WebhookByIdResult,
    WebhookByIdVariables,
    WebhookFields,
    WebhookLogFields,
    WebhookLogsByWebhookIDResult,
    WebhookLogsByWebhookIDVariables,
    WebhookPageHeaderResult,
    WebhookPageHeaderVariables,
    WebhooksListResult,
    WebhooksListVariables,
} from '../graphql-operations'
import { accessTokenFragment } from '../settings/tokens/AccessTokenNode'

import { WEBHOOK_LOGS_BY_ID } from './webhooks/backend'

/**
 * Fetches all organizations.
 */
export function fetchAllOrganizations(args: {
    first?: number
    query?: string
}): Observable<OrganizationsConnectionFields> {
    return requestGraphQL<OrganizationsResult, OrganizationsVariables>(
        gql`
            query Organizations($first: Int, $query: String) {
                organizations(first: $first, query: $query) {
                    ...OrganizationsConnectionFields
                }
            }

            fragment OrganizationsConnectionFields on OrgConnection {
                nodes {
                    ...OrganizationFields
                }
                totalCount
            }

            fragment OrganizationFields on Org {
                id
                name
                displayName
                createdAt
                latestSettings {
                    createdAt
                    contents
                }
                members {
                    totalCount
                }
            }
        `,
        {
            first: args.first ?? null,
            query: args.query ?? null,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.organizations)
    )
}

const mirrorRepositoryInfoFieldsFragment = gql`
    fragment MirrorRepositoryInfoFields on MirrorRepositoryInfo {
        cloned
        cloneInProgress
        updatedAt
        isCorrupted
        corruptionLogs {
            timestamp
        }
        lastError
        byteSize
        shard
    }
`

const externalRepositoryFieldsFragment = gql`
    fragment ExternalRepositoryFields on ExternalRepository {
        serviceType
        serviceID
    }
`

const siteAdminRepositoryFieldsFragment = gql`
    ${mirrorRepositoryInfoFieldsFragment}
    ${externalRepositoryFieldsFragment}

    fragment SiteAdminRepositoryFields on Repository {
        id
        name
        createdAt
        viewerCanAdminister
        url
        isPrivate
        mirrorInfo {
            ...MirrorRepositoryInfoFields
        }
        externalRepository {
            ...ExternalRepositoryFields
        }
    }
`

/**
 * Fetches all repositories.
 *
 * @returns Observable that emits the list of repositories
 */
function fetchAllRepositories(args: Partial<RepositoriesVariables>): Observable<RepositoriesResult['repositories']> {
    return requestGraphQL<RepositoriesResult, RepositoriesVariables>(
        gql`
            query Repositories(
                $first: Int
                $query: String
                $indexed: Boolean
                $notIndexed: Boolean
                $failedFetch: Boolean
                $corrupted: Boolean
                $cloneStatus: CloneStatus
                $orderBy: RepositoryOrderBy
                $descending: Boolean
                $externalService: ID
            ) {
                repositories(
                    first: $first
                    query: $query
                    indexed: $indexed
                    notIndexed: $notIndexed
                    failedFetch: $failedFetch
                    corrupted: $corrupted
                    cloneStatus: $cloneStatus
                    orderBy: $orderBy
                    descending: $descending
                    externalService: $externalService
                ) {
                    nodes {
                        ...SiteAdminRepositoryFields
                    }
                    totalCount(precise: true)
                    pageInfo {
                        hasNextPage
                    }
                }
            }

            ${siteAdminRepositoryFieldsFragment}
        `,
        {
            indexed: args.indexed ?? true,
            notIndexed: args.notIndexed ?? true,
            failedFetch: args.failedFetch ?? false,
            corrupted: args.corrupted ?? false,
            first: args.first ?? null,
            query: args.query ?? null,
            cloneStatus: args.cloneStatus ?? null,
            orderBy: args.orderBy ?? RepositoryOrderBy.REPOSITORY_NAME,
            descending: args.descending ?? false,
            externalService: args.externalService ?? null,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.repositories)
    )
}

export const REPO_PAGE_POLL_INTERVAL = 5000

export function fetchAllRepositoriesAndPollIfEmptyOrAnyCloning(
    args: Partial<RepositoriesVariables>
): Observable<RepositoriesResult['repositories']> {
    return fetchAllRepositories(args).pipe(
        // Poll every 5000ms if repositories are being cloned or the list is empty.
        repeatUntil(
            result =>
                result.nodes &&
                result.nodes.length > 0 &&
                result.nodes.every(nodes => !nodes.mirrorInfo.cloneInProgress && nodes.mirrorInfo.cloned),
            { delay: REPO_PAGE_POLL_INTERVAL }
        )
    )
}

export const SLOW_REQUESTS = gql`
    query SlowRequests($after: String) {
        slowRequests(after: $after) {
            nodes {
                index
                user {
                    username
                }
                start
                duration
                name
                source
                repository {
                    name
                }
                variables
                errors
                query
                filepath
            }
            totalCount
            pageInfo {
                endCursor
                hasNextPage
            }
        }
    }
`

export const OUTBOUND_REQUESTS = gql`
    query OutboundRequests($after: String) {
        outboundRequests(after: $after) {
            nodes {
                id
                startedAt
                method
                url
                requestHeaders {
                    name
                    values
                }
                requestBody
                statusCode
                responseHeaders {
                    name
                    values
                }
                durationMs
                errorMessage
                creationStackFrame
                callStack
            }
        }
    }
`
export const BACKGROUND_JOBS = gql`
    query BackgroundJobs($recentRunCount: Int) {
        backgroundJobs(recentRunCount: $recentRunCount) {
            nodes {
                name

                routines {
                    name
                    type
                    description
                    intervalMs
                    instances {
                        hostName
                        lastStartedAt
                        lastStoppedAt
                    }
                    recentRuns {
                        at
                        hostName
                        durationMs
                        errorMessage
                    }
                    stats {
                        since
                        runCount
                        errorCount
                        minDurationMs
                        avgDurationMs
                        maxDurationMs
                    }
                }
            }
        }
    }
`

export const OUTBOUND_REQUESTS_PAGE_POLL_INTERVAL_MS = 5000
export const BACKGROUND_JOBS_PAGE_POLL_INTERVAL_MS = 5000

export const UPDATE_MIRROR_REPOSITORY = gql`
    mutation UpdateMirrorRepository($repository: ID!) {
        updateMirrorRepository(repository: $repository) {
            alwaysNil
        }
    }
`

export const CHECK_MIRROR_REPOSITORY_CONNECTION = gql`
    mutation CheckMirrorRepositoryConnection($repository: ID, $name: String) {
        checkMirrorRepositoryConnection(repository: $repository, name: $name) {
            error
        }
    }
`

export function checkMirrorRepositoryConnection(
    args:
        | {
              repository: Scalars['ID']
          }
        | {
              name: string
          }
): Observable<CheckMirrorRepositoryConnectionResult['checkMirrorRepositoryConnection']> {
    return mutateGraphQL<CheckMirrorRepositoryConnectionResult>(CHECK_MIRROR_REPOSITORY_CONNECTION, args).pipe(
        map(dataOrThrowErrors),
        tap(() => resetAllMemoizationCaches()),
        map(data => data.checkMirrorRepositoryConnection)
    )
}

export function scheduleRepositoryPermissionsSync(args: { repository: Scalars['ID'] }): Observable<void> {
    return requestGraphQL<ScheduleRepositoryPermissionsSyncResult, ScheduleRepositoryPermissionsSyncVariables>(
        gql`
            mutation ScheduleRepositoryPermissionsSync($repository: ID!) {
                scheduleRepositoryPermissionsSync(repository: $repository) {
                    alwaysNil
                }
            }
        `,
        args
    ).pipe(
        map(dataOrThrowErrors),
        tap(() => resetAllMemoizationCaches()),
        mapTo(undefined)
    )
}

export const RECLONE_REPOSITORY_MUTATION = gql`
    mutation RecloneRepository($repo: ID!) {
        recloneRepository(repo: $repo) {
            alwaysNil
        }
    }
`

/**
 * Fetches the site and its configuration.
 *
 * @returns Observable that emits the site
 */
export function fetchSite(): Observable<SiteResult['site']> {
    return queryGraphQL<SiteResult>(gql`
        query Site {
            site {
                __typename
                id
                canReloadSite
                configuration {
                    id
                    effectiveContents
                    validationMessages
                }
            }
        }
    `).pipe(
        map(dataOrThrowErrors),
        map(data => data.site)
    )
}

/**
 * Placeholder for the type of the external service config (to avoid explicit 'any' type)
 */
interface ExternalServiceConfig {}

type SettingsSubject = Pick<SiteAdminSettingsCascadeFields['subjects'][number], 'settingsURL' | '__typename'> & {
    contents: Settings
}

/**
 * All configuration and settings in one place.
 */
interface AllConfig {
    site: AllConfigResult['site']
    externalServices: Partial<Record<ExternalServiceKind, ExternalServiceConfig>>
    settings: {
        subjects: SettingsSubject[]
        final: Settings | null
    }
}

/**
 * Fetches all the configuration and settings (requires site admin privileges).
 */
export function fetchAllConfigAndSettings(): Observable<AllConfig> {
    return queryGraphQL<AllConfigResult>(
        gql`
            query AllConfig($first: Int) {
                site {
                    id
                    configuration {
                        id
                        effectiveContents
                    }
                    latestSettings {
                        contents
                    }
                    settingsCascade {
                        final
                    }
                }

                externalServices(first: $first) {
                    nodes {
                        id
                        kind
                        displayName
                        config
                        createdAt
                        updatedAt
                        warning
                    }
                }

                viewerSettings {
                    ...SiteAdminSettingsCascadeFields
                }
            }

            fragment SiteAdminSettingsCascadeFields on SettingsCascade {
                subjects {
                    __typename
                    latestSettings {
                        id
                        contents
                    }
                    settingsURL
                }
                final
            }
        `,
        { first: 100 } // assume no more than 100 external services added
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            const externalServices: Partial<Record<ExternalServiceKind, ExternalServiceConfig[]>> =
                data.externalServices.nodes
                    .filter(svc => svc.config)
                    .map(svc => [svc.kind, parseJSONC(svc.config) as ExternalServiceConfig] as const)
                    .reduce<Partial<{ [k in ExternalServiceKind]: ExternalServiceConfig[] }>>(
                        (externalServicesByKind, [kind, config]) => {
                            let services = externalServicesByKind[kind]
                            if (!services) {
                                services = []
                                externalServicesByKind[kind] = services
                            }
                            services.push(config)
                            return externalServicesByKind
                        },
                        {}
                    )
            const settingsSubjects = data.viewerSettings.subjects.map(settings => ({
                __typename: settings.__typename,
                settingsURL: settings.settingsURL,
                contents: settings.latestSettings ? parseJSONC(settings.latestSettings.contents) : null,
            }))
            const finalSettings = parseJSONC(data.viewerSettings.final)
            return {
                site:
                    data.site?.configuration?.effectiveContents &&
                    parseJSONC(data.site.configuration.effectiveContents),
                externalServices,
                settings: {
                    subjects: settingsSubjects,
                    final: finalSettings,
                },
            }
        })
    )
}

/**
 * Updates the site's configuration.
 *
 * @returns An observable indicating whether or not a service restart is
 * required for the update to be applied.
 */
export function updateSiteConfiguration(lastID: number, input: string): Observable<boolean> {
    return requestGraphQL<UpdateSiteConfigurationResult, UpdateSiteConfigurationVariables>(
        gql`
            mutation UpdateSiteConfiguration($lastID: Int!, $input: String!) {
                updateSiteConfiguration(lastID: $lastID, input: $input)
            }
        `,
        { lastID, input }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.updateSiteConfiguration)
    )
}

/**
 * Reloads the site.
 */
export function reloadSite(): Observable<void> {
    return requestGraphQL<ReloadSiteResult, ReloadSiteVariables>(
        gql`
            mutation ReloadSite {
                reloadSite {
                    alwaysNil
                }
            }
        `
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.reloadSite) {
                throw createInvalidGraphQLMutationResponseError('ReloadSite')
            }
        })
    )
}

export function setUserIsSiteAdmin(userID: Scalars['ID'], siteAdmin: boolean): Observable<void> {
    return requestGraphQL<SetUserIsSiteAdminResult, SetUserIsSiteAdminVariables>(
        gql`
            mutation SetUserIsSiteAdmin($userID: ID!, $siteAdmin: Boolean!) {
                setUserIsSiteAdmin(userID: $userID, siteAdmin: $siteAdmin) {
                    alwaysNil
                }
            }
        `,
        { userID, siteAdmin }
    ).pipe(
        map(dataOrThrowErrors),
        map(() => undefined)
    )
}

export function randomizeUserPassword(
    user: Scalars['ID']
): Observable<RandomizeUserPasswordResult['randomizeUserPassword']> {
    return mutateGraphQL<RandomizeUserPasswordResult>(
        gql`
            mutation RandomizeUserPassword($user: ID!) {
                randomizeUserPassword(user: $user) {
                    resetPasswordURL
                    emailSent
                }
            }
        `,
        { user }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.randomizeUserPassword)
    )
}

export function createUser(username: string, email: string | undefined): Observable<CreateUserResult['createUser']> {
    return mutateGraphQL<CreateUserResult>(
        gql`
            mutation CreateUser($username: String!, $email: String) {
                createUser(username: $username, email: $email) {
                    resetPasswordURL
                }
            }
        `,
        { username, email }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.createUser)
    )
}

export function deleteOrganization(organization: Scalars['ID'], hard?: boolean): Promise<void> {
    return requestGraphQL<DeleteOrganizationResult, DeleteOrganizationVariables>(
        gql`
            mutation DeleteOrganization($organization: ID!, $hard: Boolean) {
                deleteOrganization(organization: $organization, hard: $hard) {
                    alwaysNil
                }
            }
        `,
        { organization, hard: hard ?? null }
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => {
                if (!data.deleteOrganization) {
                    throw createInvalidGraphQLMutationResponseError('DeleteOrganization')
                }
            })
        )
        .toPromise()
}

export const SITE_UPDATE_CHECK = gql`
    query SiteUpdateCheck {
        site {
            buildVersion
            productVersion
            updateCheck {
                pending
                checkedAt
                errorMessage
                updateVersionAvailable
            }
        }
    }
`

export function fetchSiteUpdateCheck(): Observable<SiteUpdateCheckResult['site']> {
    return requestGraphQL<SiteUpdateCheckResult, SiteUpdateCheckVariables>(SITE_UPDATE_CHECK).pipe(
        map(dataOrThrowErrors),
        map(data => data.site)
    )
}

/**
 * Fetches all out-of-band migrations.
 */
export function fetchAllOutOfBandMigrations(): Observable<OutOfBandMigrationFields[]> {
    return requestGraphQL<OutOfBandMigrationsResult, OutOfBandMigrationsVariables>(
        gql`
            query OutOfBandMigrations {
                outOfBandMigrations {
                    ...OutOfBandMigrationFields
                }
            }

            fragment OutOfBandMigrationFields on OutOfBandMigration {
                id
                team
                component
                description
                introduced
                deprecated
                progress
                created
                lastUpdated
                nonDestructive
                applyReverse
                errors {
                    message
                    created
                }
            }
        `
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.outOfBandMigrations)
    )
}

/**
 * Fetches all feature flags.
 */
export function fetchFeatureFlags(): Observable<FeatureFlagFields[]> {
    return requestGraphQL<FeatureFlagsResult, FeatureFlagsVariables>(
        gql`
            query FeatureFlags {
                featureFlags {
                    ...FeatureFlagFields
                }
            }

            fragment FeatureFlagFields on FeatureFlag {
                __typename
                ... on FeatureFlagBoolean {
                    name
                    value
                    overrides {
                        ...OverrideFields
                    }
                }
                ... on FeatureFlagRollout {
                    name
                    rolloutBasisPoints
                    overrides {
                        ...OverrideFields
                    }
                }
            }

            fragment OverrideFields on FeatureFlagOverride {
                id
                value
                # Querying on namespace seems bugged, so we just get id and value for now.
            }
        `
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.featureFlags)
    )
}

export const REPOSITORY_STATS = gql`
    query RepositoryStats {
        repositoryStats {
            __typename
            total
            notCloned
            cloned
            cloning
            failedFetch
            corrupted
            indexed
        }
    }
`

export function queryAccessTokens(args: { first?: number }): Observable<SiteAdminAccessTokenConnectionFields> {
    return requestGraphQL<SiteAdminAccessTokensResult, SiteAdminAccessTokensVariables>(
        gql`
            query SiteAdminAccessTokens($first: Int) {
                site {
                    accessTokens(first: $first) {
                        ...SiteAdminAccessTokenConnectionFields
                    }
                }
            }
            fragment SiteAdminAccessTokenConnectionFields on AccessTokenConnection {
                nodes {
                    ...AccessTokenFields
                }
                totalCount
                pageInfo {
                    hasNextPage
                }
            }
            ${accessTokenFragment}
        `,
        { first: args.first ?? null }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.site.accessTokens)
    )
}

export const SITE_EXTERNAL_SERVICE_CONFIG = gql`
    query SiteExternalServiceConfig {
        site {
            externalServicesFromFile
            allowEditExternalServicesWithFile
        }
    }
`

const WEBHOOK_FIELDS_FRAGMENT = gql`
    fragment WebhookFields on Webhook {
        id
        uuid
        url
        name
        codeHostKind
        codeHostURN
        secret
        updatedAt
        createdAt
        createdBy {
            username
            url
        }
        updatedBy {
            username
            url
        }
    }
`

export const WEBHOOKS = gql`
    ${WEBHOOK_FIELDS_FRAGMENT}

    query WebhooksList {
        webhooks {
            nodes {
                ...WebhookFields
            }
            totalCount
            pageInfo {
                hasNextPage
            }
        }
    }
`

export const WEBHOOK_BY_ID = gql`
    ${WEBHOOK_FIELDS_FRAGMENT}

    query WebhookById($id: ID!) {
        node(id: $id) {
            __typename
            ...WebhookFields
        }
    }
`

export const DELETE_WEBHOOK = gql`
    mutation DeleteWebhook($hookID: ID!) {
        deleteWebhook(id: $hookID) {
            alwaysNil
        }
    }
`

export const WEBHOOK_PAGE_HEADER = gql`
    query WebhookPageHeader {
        webhooks {
            nodes {
                webhookLogs {
                    totalCount
                }
            }
        }

        errorsOnly: webhooks {
            nodes {
                webhookLogs(onlyErrors: true) {
                    totalCount
                }
            }
        }
    }
`

export const useWebhookPageHeader = (): { loading: boolean; totalErrors: number; totalNoEvents: number } => {
    const { data, loading } = useQuery<WebhookPageHeaderResult, WebhookPageHeaderVariables>(WEBHOOK_PAGE_HEADER, {})
    const totalNoEvents = data?.webhooks.nodes.filter(webhook => webhook.webhookLogs?.totalCount === 0).length || 0
    const totalErrors =
        data?.errorsOnly.nodes.reduce((sum, webhook) => sum + (webhook.webhookLogs?.totalCount || 0), 0) || 0
    return { loading, totalErrors, totalNoEvents }
}

export const useWebhooksConnection = (): UseShowMorePaginationResult<WebhookFields> =>
    useShowMorePagination<WebhooksListResult, WebhooksListVariables, WebhookFields>({
        query: WEBHOOKS,
        variables: {},
        getConnection: result => {
            const { webhooks } = dataOrThrowErrors(result)
            return webhooks
        },
    })

export const useWebhookQuery = (id: string): QueryResult<WebhookByIdResult, WebhookByIdVariables> =>
    useQuery<WebhookByIdResult, WebhookByIdVariables>(WEBHOOK_BY_ID, {
        variables: { id },
    })

export const useWebhookLogsConnection = (
    webhookID: string,
    first: number,
    onlyErrors: boolean
): UseShowMorePaginationResult<WebhookLogFields> =>
    useShowMorePagination<WebhookLogsByWebhookIDResult, WebhookLogsByWebhookIDVariables, WebhookLogFields>({
        query: WEBHOOK_LOGS_BY_ID,
        variables: {
            first: first ?? 20,
            after: null,
            onlyErrors,
            onlyUnmatched: false,
            webhookID,
        },
        getConnection: result => {
            const { webhookLogs } = dataOrThrowErrors(result)
            return webhookLogs
        },
        options: {
            fetchPolicy: 'cache-first',
        },
    })

export const CREATE_WEBHOOK_QUERY = gql`
    mutation CreateWebhook($name: String!, $codeHostKind: String!, $codeHostURN: String!, $secret: String) {
        createWebhook(name: $name, codeHostKind: $codeHostKind, codeHostURN: $codeHostURN, secret: $secret) {
            id
        }
    }
`

export const UPDATE_WEBHOOK_QUERY = gql`
    mutation UpdateWebhook($id: ID!, $name: String!, $codeHostKind: String!, $codeHostURN: String!, $secret: String) {
        updateWebhook(id: $id, name: $name, codeHostKind: $codeHostKind, codeHostURN: $codeHostURN, secret: $secret) {
            id
        }
    }
`
