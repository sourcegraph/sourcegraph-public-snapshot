import { Observable } from 'rxjs/Observable'
import { map } from 'rxjs/operators/map'
import { queryGraphQL } from '../backend/graphql'

/**
 * Fetches all users.
 *
 * @return Observable that emits the list of users
 */
export function fetchAllUsers(): Observable<GQL.IUser[]> {
    return queryGraphQL(`query Users {
        users {
            nodes {
                id
                username
                displayName
                email
                createdAt
                latestSettings {
                    createdAt
                    configuration { contents }
                }
                orgs { name }
                tags { name }
            }
        }
    }`).pipe(
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
    return queryGraphQL(`query Orgs {
        orgs {
            nodes {
                id
                name
                displayName
                createdAt
                latestSettings {
                    createdAt
                    configuration { contents }
                }
                members { user { username } }
                tags { name }
            }
        }
    }`).pipe(
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
    return queryGraphQL(`query Repositories {
        repositories {
            nodes {
                id
                uri
                createdAt
            }
        }
    }`).pipe(
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
    return queryGraphQL(`query Users {
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
         }`).pipe(
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
    return queryGraphQL(`query SiteConfiguration {
        site {
            id
            configuration
            latestSettings {
                configuration {
                    contents
                }
            }
        }
    }`).pipe(
        map(({ data, errors }) => {
            if (!data || !data.site) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.site
        })
    )
}
