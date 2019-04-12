import { Observable, Subject } from 'rxjs'
import { delay, map, mergeMap, retryWhen, startWith, tap } from 'rxjs/operators'
import { createInvalidGraphQLMutationResponseError, dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { resetAllMemoizationCaches } from '../../../shared/src/util/memoizeObservable'
import { mutateGraphQL, queryGraphQL } from '../backend/graphql'

/**
 * Fetches all users.
 */
export function fetchAllUsers(args: { first?: number; query?: string }): Observable<GQL.IUserConnection> {
    return queryGraphQL(
        gql`
            query Users($first: Int, $query: String) {
                users(first: $first, query: $query) {
                    nodes {
                        id
                        username
                        displayName
                        emails {
                            email
                            verified
                            verificationPending
                            viewerCanManuallyVerify
                        }
                        createdAt
                        siteAdmin
                        latestSettings {
                            createdAt
                            contents
                        }
                        organizations {
                            nodes {
                                name
                            }
                        }
                    }
                    totalCount
                }
            }
        `,
        args
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.users)
    )
}

/**
 * Fetches all organizations.
 */
export function fetchAllOrganizations(args: { first?: number; query?: string }): Observable<GQL.IOrgConnection> {
    return queryGraphQL(
        gql`
            query Organizations($first: Int, $query: String) {
                organizations(first: $first, query: $query) {
                    nodes {
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
                    totalCount
                }
            }
        `,
        args
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.organizations)
    )
}

interface RepositoryArgs {
    first?: number
    query?: string
    enabled?: boolean
    disabled?: boolean
    cloned?: boolean
    cloneInProgress?: boolean
    notCloned?: boolean
    indexed?: boolean
    notIndexed?: boolean
}

/**
 * Fetches all repositories.
 *
 * @return Observable that emits the list of repositories
 */
function fetchAllRepositories(args: RepositoryArgs): Observable<GQL.IRepositoryConnection> {
    args = {
        enabled: true,
        disabled: false,
        cloned: true,
        cloneInProgress: true,
        notCloned: true,
        indexed: true,
        notIndexed: true,
        ...args,
    } // apply defaults
    return queryGraphQL(
        gql`
            query Repositories(
                $first: Int
                $query: String
                $enabled: Boolean
                $disabled: Boolean
                $cloned: Boolean
                $cloneInProgress: Boolean
                $notCloned: Boolean
                $indexed: Boolean
                $notIndexed: Boolean
            ) {
                repositories(
                    first: $first
                    query: $query
                    enabled: $enabled
                    disabled: $disabled
                    cloned: $cloned
                    cloneInProgress: $cloneInProgress
                    notCloned: $notCloned
                    indexed: $indexed
                    notIndexed: $notIndexed
                ) {
                    nodes {
                        id
                        name
                        enabled
                        createdAt
                        viewerCanAdminister
                        url
                        mirrorInfo {
                            cloned
                            cloneInProgress
                            updatedAt
                        }
                    }
                    totalCount(precise: true)
                    pageInfo {
                        hasNextPage
                    }
                }
            }
        `,
        args
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.repositories)
    )
}

/**
 * Checks if there are any repositories that require enable/disable state.
 *
 * @return Observable that emits a boolean allowEnableDisable
 */
export function fetchAllowEnableDisable(): Observable<boolean> {
    return queryGraphQL(
        gql`
            query AllowEnableDisable {
                internal {
                    allowEnableDisable
                }
            }
        `,
        {}
    ).pipe(
        map(dataOrThrowErrors),
        retryWhen(errors => errors.pipe(delay(1000))),
        map(data => data.internal.allowEnableDisable)
    )
}

export function fetchAllRepositoriesAndPollIfAnyCloning(args: RepositoryArgs): Observable<GQL.IRepositoryConnection> {
    // Poll if there are repositories that are being cloned.
    //
    // TODO(sqs): This is hacky, but I couldn't figure out a better way.
    const subject = new Subject<null>()
    return subject.pipe(
        startWith(null),
        mergeMap(() => fetchAllRepositories(args)),
        tap(result => {
            if (result.nodes && result.nodes.some(n => n.enabled && !n.mirrorInfo.cloned)) {
                setTimeout(() => subject.next(), 5000)
            }
        })
    )
}

export function setRepositoryEnabled(repository: GQL.ID, enabled: boolean): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation SetRepositoryEnabled($repository: ID!, $enabled: Boolean!) {
                setRepositoryEnabled(repository: $repository, enabled: $enabled) {
                    alwaysNil
                }
            }
        `,
        { repository, enabled }
    ).pipe(
        map(dataOrThrowErrors),
        tap(() => resetAllMemoizationCaches()),
        map(() => undefined)
    )
}

export function setAllRepositoriesEnabled(enabled: boolean): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation SetAllRepositoriesEnabled($enabled: Boolean!) {
                setAllRepositoriesEnabled(enabled: $enabled) {
                    alwaysNil
                }
            }
        `,
        { enabled }
    ).pipe(
        map(dataOrThrowErrors),
        tap(() => resetAllMemoizationCaches()),
        map(() => undefined)
    )
}

export function updateMirrorRepository(args: { repository: GQL.ID }): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation UpdateMirrorRepository($repository: ID!) {
                updateMirrorRepository(repository: $repository) {
                    alwaysNil
                }
            }
        `,
        args
    ).pipe(
        map(dataOrThrowErrors),
        tap(() => resetAllMemoizationCaches()),
        map(() => undefined)
    )
}

export function updateAllMirrorRepositories(): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation UpdateAllMirrorRepositories() {
                updateAllMirrorRepositories() {
                    alwaysNil
                }
            }
        `
    ).pipe(
        map(dataOrThrowErrors),
        tap(() => resetAllMemoizationCaches()),
        map(() => undefined)
    )
}

export function checkMirrorRepositoryConnection(
    args:
        | {
              repository: GQL.ID
          }
        | {
              name: string
          }
): Observable<GQL.ICheckMirrorRepositoryConnectionResult> {
    return mutateGraphQL(
        gql`
            mutation CheckMirrorRepositoryConnection($repository: ID, $name: String) {
                checkMirrorRepositoryConnection(repository: $repository, name: $name) {
                    error
                }
            }
        `,
        args
    ).pipe(
        map(dataOrThrowErrors),
        tap(() => resetAllMemoizationCaches()),
        map(data => data.checkMirrorRepositoryConnection)
    )
}

export function deleteRepository(repository: GQL.ID): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation DeleteRepository($repository: ID!) {
                deleteRepository(repository: $repository) {
                    alwaysNil
                }
            }
        `,
        { repository }
    ).pipe(
        map(dataOrThrowErrors),
        tap(() => resetAllMemoizationCaches()),
        map(() => undefined)
    )
}

/**
 * Fetches usage statistics for all users.
 *
 * @return Observable that emits the list of users and their usage data
 */
export function fetchUserUsageStatistics(args: {
    activePeriod?: GQL.UserActivePeriod
    query?: string
    first?: number
}): Observable<GQL.IUserConnection> {
    return queryGraphQL(
        gql`
            query UserUsageStatistics($activePeriod: UserActivePeriod, $query: String, $first: Int) {
                users(activePeriod: $activePeriod, query: $query, first: $first) {
                    nodes {
                        id
                        username
                        usageStatistics {
                            searchQueries
                            pageViews
                            codeIntelligenceActions
                            lastActiveTime
                            lastActiveCodeHostIntegrationTime
                        }
                    }
                    totalCount
                }
            }
        `,
        args
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.users)
    )
}

/**
 * Fetches site-wide usage statitics.
 *
 * @return Observable that emits the list of users and their usage data
 */
export function fetchSiteUsageStatistics(): Observable<GQL.ISiteUsageStatistics> {
    return queryGraphQL(gql`
        query SiteUsageStatistics {
            site {
                usageStatistics {
                    daus {
                        userCount
                        registeredUserCount
                        anonymousUserCount
                        startTime
                    }
                    waus {
                        userCount
                        registeredUserCount
                        anonymousUserCount
                        startTime
                    }
                    maus {
                        userCount
                        registeredUserCount
                        anonymousUserCount
                        startTime
                    }
                }
            }
        }
    `).pipe(
        map(dataOrThrowErrors),
        map(data => data.site.usageStatistics)
    )
}

/**
 * Fetches the site and its configuration.
 *
 * @return Observable that emits the site
 */
export function fetchSite(): Observable<GQL.ISite> {
    return queryGraphQL(gql`
        query Site {
            site {
                id
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
 * Updates the site's configuration.
 *
 * @returns An observable indicating whether or not a service restart is
 * required for the update to be applied.
 */
export function updateSiteConfiguration(lastID: number, input: string): Observable<boolean> {
    return mutateGraphQL(
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
    return mutateGraphQL(
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

export function setUserIsSiteAdmin(userID: GQL.ID, siteAdmin: boolean): Observable<void> {
    return mutateGraphQL(
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

export function randomizeUserPassword(user: GQL.ID): Observable<GQL.IRandomizeUserPasswordResult> {
    return mutateGraphQL(
        gql`
            mutation RandomizeUserPassword($user: ID!) {
                randomizeUserPassword(user: $user) {
                    resetPasswordURL
                }
            }
        `,
        { user }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.randomizeUserPassword)
    )
}

export function deleteUser(user: GQL.ID, hard?: boolean): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation DeleteUser($user: ID!, $hard: Boolean) {
                deleteUser(user: $user, hard: $hard) {
                    alwaysNil
                }
            }
        `,
        { user, hard }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.deleteUser) {
                throw createInvalidGraphQLMutationResponseError('DeleteUser')
            }
        })
    )
}

export function createUser(username: string, email: string | undefined): Observable<GQL.ICreateUserResult> {
    return mutateGraphQL(
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

export function deleteOrganization(organization: GQL.ID): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation DeleteOrganization($organization: ID!) {
                deleteOrganization(organization: $organization) {
                    alwaysNil
                }
            }
        `,
        { organization }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data.deleteOrganization) {
                throw createInvalidGraphQLMutationResponseError('DeleteOrganization')
            }
        })
    )
}

export function fetchSiteUpdateCheck(): Observable<{
    buildVersion: string
    productVersion: string
    updateCheck: GQL.IUpdateCheck
}> {
    return queryGraphQL(
        gql`
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
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.site)
    )
}
