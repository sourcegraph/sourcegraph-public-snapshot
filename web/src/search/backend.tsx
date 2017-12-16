import { Observable } from 'rxjs/Observable'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { map } from 'rxjs/operators/map'
import { mergeMap } from 'rxjs/operators/mergeMap'
import { queryGraphQL } from '../backend/graphql'
import { mutateConfigurationGraphQL } from '../configuration/backend'
import { currentConfiguration } from '../settings/configuration'
import { SearchOptions } from './index'

export function searchText(options: SearchOptions): Observable<GQL.ISearchResults> {
    return queryGraphQL(
        `query Search(
            $query: String!,
            $scopeQuery: String!,
        ) {
            search(query: $query, scopeQuery: $scopeQuery) {
                results {
                    limitHit
                    missing
                    cloning
                    timedout
                    results {
                        __typename
                        ... on FileMatch {
                            resource
                            limitHit
                            lineMatches {
                                preview
                                lineNumber
                                offsetAndLengths
                            }
                        }
                        ... on CommitSearchResult {
                            refs {
                                name
                                displayName
                                prefix
                                repository { uri }
                            }
                            sourceRefs {
                                name
                                displayName
                                prefix
                                repository { uri }
                            }
                            messagePreview {
                                value
                                highlights {
                                    line
                                    character
                                    length
                                }
                            }
                            diffPreview {
                                value
                                highlights {
                                    line
                                    character
                                    length
                                }
                            }
                            commit {
                                repository {
                                    uri
                                }
                                oid
                                abbreviatedOID
                                author {
                                    person {
                                        displayName
                                        avatarURL
                                    }
                                    date
                                }
                                message
                            }
                        }
                    }
                    alert {
                        title
                        description
                        proposedQueries {
                            description
                            query {
                                query
                                scopeQuery
                            }
                        }
                    }
                }
            }
        }`,
        { query: options.query, scopeQuery: options.scopeQuery || '' }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.search || !data.search.results) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.search.results
        })
    )
}

export function fetchSearchResultCount(options: SearchOptions): Observable<GQL.ISearchResults> {
    return queryGraphQL(
        `query SearchResultsCount(
            $query: String!,
            $scopeQuery: String!,
        ) {
            search(query: $query, scopeQuery: $scopeQuery) {
                results {
                    limitHit
                    missing
                    cloning
                    resultCount
                    approximateResultCount
                }
            }
        }`,
        { query: options.query, scopeQuery: options.scopeQuery || '' }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.search || !data.search.results) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.search.results
        })
    )
}

export function fetchSuggestions(options: SearchOptions): Observable<GQL.SearchSuggestion> {
    return queryGraphQL(
        `query Search(
            $query: String!,
            $scopeQuery: String!,
        ) {
            search(query: $query, scopeQuery: $scopeQuery) {
                suggestions {
                    ... on Repository {
                        __typename
                        uri
                    }
                    ... on File {
                        __typename
                        name
                        isDirectory
                        repository {
                            uri
                        }
                    }
                }
            }
        }`,
        { query: options.query, scopeQuery: options.scopeQuery || '' }
    ).pipe(
        mergeMap(({ data, errors }) => {
            if (!data || !data.search || !data.search.suggestions) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.search.suggestions
        })
    )
}

export function fetchSearchScopes(): Observable<GQL.ISearchScope[]> {
    return queryGraphQL(`
        query SearchScopes {
            searchScopes {
                name
                value
            }
        }
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.searchScopes) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.searchScopes
        })
    )
}

export function fetchRepoGroups(): Observable<GQL.IRepoGroup[]> {
    return queryGraphQL(`
        query RepoGroups {
            repoGroups {
                name
                repositories
            }
        }
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.repoGroups) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.repoGroups
        })
    )
}

const gqlSavedQuery = `
    subject {
        ... on Org { id }
        ... on User { id }
    }
    index
    description
    query {
        query
        scopeQuery
    }
`

interface ISavedQuery {
    description: string
    query?: string
    scopeQuery?: string
}

interface ISavedQueryConfig {
    // TODO(sqs): can use SAVED_QUERY_CONFIG_SECTION in type literal
    // when https://github.com/Microsoft/TypeScript/pull/15473 ships.
    ['search.savedQueries']?: ISavedQuery[]
}

function savedQueriesEqual(a: ISavedQuery, b: ISavedQuery): boolean {
    return a.description === b.description && a.query === b.query && a.scopeQuery === b.scopeQuery
}

export function observeSavedQueries(): Observable<GQL.ISavedQuery[]> {
    return currentConfiguration.pipe(
        map((config: ISavedQueryConfig) => config['search.savedQueries']),
        distinctUntilChanged(
            (a, b) =>
                (!a && !b) || (!!a && !!b && a.length === b.length && a.every((q, i) => savedQueriesEqual(q, b[i])))
        ),
        mergeMap(fetchSavedQueries)
    )
}

function fetchSavedQueries(): Observable<GQL.ISavedQuery[]> {
    return queryGraphQL(`
                query SavedQueries {
                    savedQueries {
                        ${gqlSavedQuery}
                    }
                }`).pipe(
        map(({ data, errors }) => {
            if (!data || !data.savedQueries) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.savedQueries
        })
    )
}

export function updateDeploymentConfiguration(email: string, telemetryEnabled: boolean): Observable<void> {
    return queryGraphQL(
        `query UpdateDeploymentConfiguration($email: String, $enableTelemetry: Boolean) {
                updateDeploymentConfiguration(email: $email, enableTelemetry: $enableTelemetry) {
                    alwaysNil
                }
            }`,
        { email, enableTelemetry: telemetryEnabled }
    ).pipe(
        map(({ data, errors }) => {
            if (!data) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
        })
    )
}

export function createSavedQuery(
    subject: GQL.ConfigurationSubject | GQL.IConfigurationSubject | { id: GQLID },
    description: string,
    query: string,
    scopeQuery: string
): Observable<GQL.ISavedQuery> {
    return mutateConfigurationGraphQL(
        subject,
        `mutation CreateSavedQuery($subject: ID!, $lastID: Int, $description: String!, $query: String!, $scopeQuery: String!) {
            configurationMutation(input: {subject: $subject, lastID: $lastID}) {
                createSavedQuery(description: $description, query: $query, scopeQuery: $scopeQuery) {
                    ${gqlSavedQuery}
                }
            }
        }`,
        { description, query, scopeQuery }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.configurationMutation || !data.configurationMutation.createSavedQuery) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.configurationMutation.createSavedQuery
        })
    )
}

export function updateSavedQuery(
    subject: GQL.ConfigurationSubject | GQL.IConfigurationSubject | { id: GQLID },
    index: number,
    description: string,
    query: string,
    scopeQuery: string
): Observable<GQL.ISavedQuery> {
    return mutateConfigurationGraphQL(
        subject,
        `mutation UpdateSavedQuery($subject: ID!, $lastID: Int, $index: Int!, $description: String, $query: String, $scopeQuery: String) {
            configurationMutation(input: {subject: $subject, lastID: $lastID}) {
                updateSavedQuery(index: $index, description: $description, query: $query, scopeQuery: $scopeQuery) {
                    ${gqlSavedQuery}
                }
            }
        }`,
        { index, description, query, scopeQuery }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.configurationMutation || !data.configurationMutation.updateSavedQuery) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.configurationMutation.updateSavedQuery
        })
    )
}

export function deleteSavedQuery(
    subject: GQL.ConfigurationSubject | GQL.IConfigurationSubject | { id: GQLID },
    index: number
): Observable<void> {
    return mutateConfigurationGraphQL(
        subject,
        `mutation DeleteSavedQuery($subject: ID!, $lastID: Int, $index: Int!) {
            configurationMutation(input: {subject: $subject, lastID: $lastID}) {
                deleteSavedQuery(index: $index) {
                    alwaysNil
                }
            }
        }`,
        { index }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.configurationMutation || !data.configurationMutation.deleteSavedQuery) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
        })
    )
}
