import { Observable, Subject } from 'rxjs'
import { map, mergeMap, startWith, tap } from 'rxjs/operators'
import {
    createInvalidGraphQLMutationResponseError,
    dataOrThrowErrors,
    gql,
    mutateGraphQL,
    queryGraphQL,
} from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { resetAllMemoizationCaches } from '../util/memoize'

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
                            configuration {
                                contents
                            }
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
                            configuration {
                                contents
                            }
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

/**
 * Add a repository. See the GraphQL documentation for Mutation.addRepository.
 */
export function addRepository(
    name: string
): Observable<{
    /** The ID of the newly added repository (or the existing repository, if it already existed). */
    id: GQL.ID
}> {
    return mutateGraphQL(
        gql`
            mutation AddRepository($name: String!) {
                addRepository(name: $name) {
                    id
                }
            }
        `,
        { name }
    ).pipe(
        map(dataOrThrowErrors),
        tap(() => resetAllMemoizationCaches()), // in case we memoized that this repository doesn't exist
        map(data => data.addRepository)
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
 * Fetches usage analytics for all users.
 *
 * @return Observable that emits the list of users and their usage data
 */
export function fetchUserAnalytics(args: {
    activePeriod?: GQL.UserActivePeriod
    query?: string
    first?: number
}): Observable<GQL.IUserConnection> {
    return queryGraphQL(
        gql`
            query UserAnalytics($activePeriod: UserActivePeriod, $query: String, $first: Int) {
                users(activePeriod: $activePeriod, query: $query, first: $first) {
                    nodes {
                        id
                        username
                        activity {
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
 * Fetches site-wide usage analytics.
 *
 * @return Observable that emits the list of users and their usage data
 */
export function fetchSiteAnalytics(): Observable<GQL.ISiteActivity> {
    return queryGraphQL(gql`
        query SiteAnalytics {
            site {
                activity {
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
        map(data => data.site.activity)
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
                    effectiveContents
                    pendingContents
                    validationMessages
                    canUpdate
                    source
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
export function updateSiteConfiguration(input: string): Observable<boolean> {
    return mutateGraphQL(
        gql`
            mutation UpdateSiteConfiguration($input: String!) {
                updateSiteConfiguration(input: $input)
            }
        `,
        { input }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.updateSiteConfiguration as boolean)
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

export function deleteUser(user: GQL.ID): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation DeleteUser($user: ID!) {
                deleteUser(user: $user) {
                    alwaysNil
                }
            }
        `,
        { user }
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
    productName: string
    buildVersion: string
    productVersion: string
    updateCheck: GQL.IUpdateCheck
}> {
    return queryGraphQL(
        gql`
            query SiteUpdateCheck {
                site {
                    productName
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

/**
 * Fetches all known language server's information.
 *
 * @return Observable that emits the list of language servers.
 */
export function fetchLangServers(): Observable<GQL.ILangServer[]> {
    return queryGraphQL(gql`
        query LangServers {
            site {
                langServers {
                    language
                    displayName
                    homepageURL
                    issuesURL
                    docsURL
                    dataCenter
                    custom
                    state
                    pending
                    downloading
                    canEnable
                    canDisable
                    canRestart
                    canUpdate
                    healthy
                    experimental
                }
            }
        }
    `).pipe(
        map(dataOrThrowErrors),
        map(data => data.site.langServers)
    )
}

/**
 * Enables the language server for the given language.
 */
export function enableLangServer(language: string): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation EnableLangServer($language: String!) {
                langServers {
                    enable(language: $language) {
                        alwaysNil
                    }
                }
            }
        `,
        { language }
    ).pipe(
        map(dataOrThrowErrors),
        map(() => undefined)
    )
}

/**
 * Disables the language server for the given language.
 */
export function disableLangServer(language: string): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation DisableLangServer($language: String!) {
                langServers {
                    disable(language: $language) {
                        alwaysNil
                    }
                }
            }
        `,
        { language }
    ).pipe(
        map(dataOrThrowErrors),
        map(() => undefined)
    )
}

/**
 * Restarts the language server for the given language.
 */
export function restartLangServer(language: string): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation RestartLangServer($language: String!) {
                langServers {
                    restart(language: $language) {
                        alwaysNil
                    }
                }
            }
        `,
        { language }
    ).pipe(
        map(dataOrThrowErrors),
        map(() => undefined)
    )
}

/**
 * Updates the language server for the given language.
 */
export function updateLangServer(language: string): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation UpdateLangServer($language: String!) {
                langServers {
                    update(language: $language) {
                        alwaysNil
                    }
                }
            }
        `,
        { language }
    ).pipe(
        map(dataOrThrowErrors),
        map(() => undefined)
    )
}
