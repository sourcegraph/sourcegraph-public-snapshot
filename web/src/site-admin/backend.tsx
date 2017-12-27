import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { gql, mutateGraphQL, queryGraphQL } from '../backend/graphql'

/**
 * Fetches all users.
 *
 * @return Observable that emits the list of users
 */
export function fetchAllUsers(): Observable<GQL.IUser[]> {
    return queryGraphQL(gql`
        query Users {
            users {
                nodes {
                    id
                    auth0ID
                    username
                    displayName
                    email
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
            }
        }
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.users) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.users.nodes
        })
    )
}

/**
 * Fetches all orgs.
 *
 * @return Observable that emits the list of orgs
 */
export function fetchAllOrgs(): Observable<GQL.IOrg[]> {
    return queryGraphQL(gql`
        query Orgs {
            orgs {
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
                        user {
                            username
                        }
                    }
                    tags {
                        name
                    }
                }
            }
        }
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.orgs) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.orgs.nodes
        })
    )
}

/**
 * Fetches all repositories.
 *
 * @return Observable that emits the list of repositories
 */
export function fetchAllRepositories(): Observable<GQL.IRepository[]> {
    return queryGraphQL(gql`
        query Repositories {
            repositories {
                nodes {
                    id
                    uri
                    createdAt
                }
            }
        }
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.repositories) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.repositories.nodes
        })
    )
}

/**
 * Fetches usage analytics for all users.
 *
 * @return Observable that emits the list of users and their usage data
 */
export function fetchUserAnalytics(): Observable<GQL.IUser[]> {
    return queryGraphQL(gql`
        query Users {
            users {
                nodes {
                    id
                    username
                    activity {
                        searchQueries
                        pageViews
                    }
                }
            }
        }
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.users) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.users.nodes
        })
    )
}

/**
 * Fetches the site and its configuration.
 *
 * @return Observable that emits the site
 */
export function fetchSite(): Observable<GQL.ISite> {
    return queryGraphQL(gql`
        query SiteConfiguration {
            site {
                id
                configuration {
                    effectiveContents
                    pendingContents
                    canUpdate
                    source
                }
            }
        }
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.site) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.site
        })
    )
}

/**
 * Updates the site's configuration.
 */
export function updateSiteConfiguration(input: string): Observable<void> {
    return mutateGraphQL(
        gql`
        mutation UpdateSiteConfiguration($input: String!) {
        updateSiteConfiguration(input: $input) {}
    }`,
        { input }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.updateSiteConfiguration) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.updateSiteConfiguration as any
        })
    )
}

/**
 * Reloads the site.
 */
export function reloadSite(): Observable<void> {
    return mutateGraphQL(gql`mutation ReloadSite() { reloadSite {} }`).pipe(
        map(({ data, errors }) => {
            if (!data || !data.reloadSite) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.reloadSite as any
        })
    )
}

export function updateDeploymentConfiguration(email: string, telemetryEnabled: boolean): Observable<void> {
    return queryGraphQL(
        gql`
            query UpdateDeploymentConfiguration($email: String, $enableTelemetry: Boolean) {
                updateDeploymentConfiguration(email: $email, enableTelemetry: $enableTelemetry) {
                    alwaysNil
                }
            }
        `,
        { email, enableTelemetry: telemetryEnabled }
    ).pipe(
        map(({ data, errors }) => {
            if (!data) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
        })
    )
}

export function setUserIsSiteAdmin(userID: GQLID, siteAdmin: boolean): Observable<void> {
    return mutateGraphQL(
        gql`
    mutation SetUserIsSiteAdmin($userID: ID!, $siteAdmin: Boolean!) {
        setUserIsSiteAdmin(userID: $userID, siteAdmin: $siteAdmin) { }
    }`,
        { userID, siteAdmin }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || (errors && errors.length > 0)) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
        })
    )
}
