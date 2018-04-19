import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { startWith } from 'rxjs/operators/startWith'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { gql, mutateGraphQL, queryGraphQL } from '../backend/graphql'
import * as GQL from '../backend/graphqlschema'
import { createAggregateError } from '../util/errors'
import { resetAllMemoizationCaches } from '../util/memoize'

/**
 * Fetches all users.
 */
export function fetchAllUsers(args: { first?: number; query?: string }): Observable<GQL.IUserConnection> {
    return queryGraphQL(
        gql`
            query Users($first: Int, $query: String) {
                site {
                    users(first: $first, query: $query) {
                        nodes {
                            id
                            externalID
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
                            orgs {
                                name
                            }
                            tags {
                                name
                            }
                        }
                        totalCount
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.site || !data.site.users) {
                throw createAggregateError(errors)
            }
            return data.site.users
        })
    )
}

/**
 * Fetches all orgs.
 */
export function fetchAllOrgs(args: { first?: number; query?: string }): Observable<GQL.IOrgConnection> {
    return queryGraphQL(
        gql`
            query Orgs($first: Int, $query: String) {
                site {
                    orgs(first: $first, query: $query) {
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
                            tags {
                                name
                            }
                        }
                        totalCount
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.site || !data.site.orgs) {
                throw createAggregateError(errors)
            }
            return data.site.orgs
        })
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
export function fetchAllRepositories(args: RepositoryArgs): Observable<GQL.IRepositoryConnection> {
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
                site {
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
                            uri
                            enabled
                            createdAt
                            viewerCanAdminister
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
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.site || !data.site.repositories || !data.site.repositories.nodes) {
                throw createAggregateError(errors)
            }
            return data.site.repositories
        })
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
        map(({ data, errors }) => {
            if (!data || !data.addRepository || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            resetAllMemoizationCaches() // in case we memoized that this repository doesn't exist
            return data.addRepository
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
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            resetAllMemoizationCaches()
        })
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
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            resetAllMemoizationCaches()
        })
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
        map(({ data, errors }) => {
            if (!data || !data.updateMirrorRepository || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            resetAllMemoizationCaches()
        })
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
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            resetAllMemoizationCaches()
        })
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
        map(({ data, errors }) => {
            if (!data || !data.checkMirrorRepositoryConnection || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return data.checkMirrorRepositoryConnection
        })
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
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            resetAllMemoizationCaches()
        })
    )
}

/**
 * Fetches usage analytics for all users.
 *
 * @return Observable that emits the list of users and their usage data
 */
export function fetchUserAndSiteAnalytics(): Observable<{ users: GQL.IUser[]; siteActivity: GQL.ISiteActivity }> {
    return queryGraphQL(gql`
        query Users {
            site {
                users {
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
                }
                # TODO(Dan): fetch site activity data here, and expose on SiteAdminAnalyticsPage
            }
        }
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.site || !data.site.users) {
                throw createAggregateError(errors)
            }
            return { users: data.site.users.nodes, siteActivity: data.site.activity }
        })
    )
}

/**
 * Fetches the site and its configuration.
 *
 * @return Observable that emits the site
 */
export function fetchSite(opt: { telemetrySamples?: boolean }): Observable<GQL.ISite> {
    return queryGraphQL(gql`
        query Site {
            site {
                id
                configuration {
                    effectiveContents
                    pendingContents
                    extraValidationErrors
                    canUpdate
                    source
                }
                ${opt && opt.telemetrySamples ? 'telemetrySamples' : ''}
            }
        }
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.site) {
                throw createAggregateError(errors)
            }
            return data.site
        })
    )
}

const settingsFragment = gql`
    fragment SettingsFields on Settings {
        id
        configuration {
            contents
        }
    }
`

/**
 * Fetches global site settings.
 *
 * @return Observable that emits the settings or `null` if it doesn't exist
 */
export function fetchSiteSettings(): Observable<GQL.ISettings | null> {
    return queryGraphQL(
        gql`
            query CurrentSiteSettings() {
                currentSiteSettings {
                    ...SettingsFields
                }
            }
            ${settingsFragment}
        `
    ).pipe(
        map(({ data, errors }) => {
            if (!data) {
                throw createAggregateError(errors)
            }
            return data.currentSiteSettings
        })
    )
}

/**
 * Updates global site settings.
 *
 * @return Observable that emits the newly updated settings
 */
export function updateSiteSettings(lastKnownSettingsID: number | null, contents: string): Observable<GQL.ISettings> {
    return mutateGraphQL(
        gql`
            mutation UpdateSiteSettings($lastKnownSettingsID: Int, $contents: String!) {
                updateSiteSettings(lastKnownSettingsID: $lastKnownSettingsID, contents: $contents) {
                    ...SettingsFields
                }
            }
            ${settingsFragment}
        `,
        { lastKnownSettingsID, contents }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0) || !data.updateSiteSettings) {
                throw createAggregateError(errors)
            }
            return data.updateSiteSettings
        })
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
        map(({ data, errors }) => {
            if (!data || errors) {
                throw createAggregateError(errors)
            }
            return data.updateSiteConfiguration as boolean
        })
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
        map(({ data, errors }) => {
            if (!data || !data.reloadSite) {
                throw createAggregateError(errors)
            }
            return data.reloadSite as any
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
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
        })
    )
}

export function randomizeUserPasswordBySiteAdmin(
    user: GQL.ID
): Observable<GQL.IRandomizeUserPasswordBySiteAdminResult> {
    return mutateGraphQL(
        gql`
            mutation RandomizeUserPasswordBySiteAdmin($user: ID!) {
                randomizeUserPasswordBySiteAdmin(user: $user) {
                    resetPasswordURL
                }
            }
        `,
        { user }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0) || !data.randomizeUserPasswordBySiteAdmin) {
                throw createAggregateError(errors)
            }
            return data.randomizeUserPasswordBySiteAdmin
        })
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
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0) || !data.deleteUser) {
                throw createAggregateError(errors)
            }
        })
    )
}

export function createUserBySiteAdmin(username: string, email: string): Observable<GQL.ICreateUserBySiteAdminResult> {
    return mutateGraphQL(
        gql`
            mutation CreateUserBySiteAdmin($username: String!, $email: String!) {
                createUserBySiteAdmin(username: $username, email: $email) {
                    resetPasswordURL
                }
            }
        `,
        { username, email }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0) || !data.createUserBySiteAdmin) {
                throw createAggregateError(errors)
            }
            return data.createUserBySiteAdmin
        })
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
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0) || !data.deleteOrganization) {
                throw createAggregateError(errors)
            }
        })
    )
}

/**
 * Fetches all threads.
 */
export function fetchAllThreads(args: { first?: number }): Observable<GQL.IThreadConnection> {
    return queryGraphQL(
        gql`
            query SiteThreads($first: Int) {
                site {
                    threads(first: $first) {
                        nodes {
                            id
                            repo {
                                canonicalRemoteID
                                org {
                                    name
                                }
                            }
                            repoRevisionPath
                            branch
                            repoRevisionPath
                            repoRevision
                            title
                            createdAt
                            archivedAt
                            author {
                                id
                                username
                                displayName
                            }
                            comments {
                                id
                                title
                                contents
                                createdAt
                                author {
                                    id
                                    username
                                    displayName
                                }
                            }
                        }
                        totalCount
                    }
                }
            }
        `,
        args
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.site || !data.site.threads) {
                throw createAggregateError(errors)
            }
            return data.site.threads
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
        map(({ data, errors }) => {
            if (!data || !data.site || !data.site.updateCheck) {
                throw createAggregateError(errors)
            }
            return {
                buildVersion: data.site.buildVersion,
                productVersion: data.site.productVersion,
                updateCheck: data.site.updateCheck,
            }
        })
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
                    canEnable
                    canDisable
                    canRestart
                    canUpdate
                    healthy
                }
            }
        }
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.site || !data.site.langServers) {
                throw createAggregateError(errors)
            }
            return data.site.langServers
        })
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
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return
        })
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
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return
        })
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
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return
        })
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
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0)) {
                throw createAggregateError(errors)
            }
            return
        })
    )
}
