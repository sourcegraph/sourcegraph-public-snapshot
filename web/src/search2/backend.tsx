import 'rxjs/add/operator/map'
import 'rxjs/add/operator/mergeMap'
import { Observable } from 'rxjs/Observable'
import { queryGraphQL } from '../backend/graphql'
import { SearchOptions } from './index'

export function searchText(options: SearchOptions): Observable<GQL.ISearchResults2> {
    return queryGraphQL(
        `query Search2(
            $query: String!,
            $scopeQuery: String!,
        ) {
            root {
                search2(query: $query, scopeQuery: $scopeQuery) {
                    results {
                        limitHit
                        missing
                        cloning
                        results {
                            resource
                            limitHit
                            lineMatches {
                                preview
                                lineNumber
                                offsetAndLengths
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
            }
        }`,
        { query: options.query, scopeQuery: options.scopeQuery || '' }
    ).map(({ data, errors }) => {
        if (!data || !data.root || !data.root.search2 || !data.root.search2.results) {
            throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
        }
        return data.root.search2.results
    })
}

export function fetchSuggestions(options: SearchOptions): Observable<GQL.SearchSuggestion2> {
    return queryGraphQL(
        `query Search2(
            $query: String!,
            $scopeQuery: String!,
        ) {
            root {
                search2(query: $query, scopeQuery: $scopeQuery) {
                    suggestions {
                        ... on Repository {
                            __typename
                            uri
                        }
                        ... on File {
                            __typename
                            name
                            repository {
                                uri
                            }
                        }
                    }
                }
            }
        }`,
        { query: options.query, scopeQuery: options.scopeQuery || '' }
    ).mergeMap(({ data, errors }) => {
        if (!data || !data.root || !data.root.search2 || !data.root.search2.suggestions) {
            throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
        }
        return data.root.search2.suggestions
    })
}

export function fetchSearchScopes(): Observable<GQL.ISearchScope2[]> {
    return queryGraphQL(`
        query SearchScopes2 {
            root {
                searchScopes2 {
                    name
                    value
                }
            }
        }
    `).map(({ data, errors }) => {
        if (!data || !data.root || !data.root.searchScopes2) {
            throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
        }
        return data.root.searchScopes2
    })
}

export function fetchRepoGroups(): Observable<GQL.IRepoGroup[]> {
    return queryGraphQL(`
        query RepoGroups {
            root {
                repoGroups {
                    name
                    repositories
                }
            }
        }
    `).map(({ data, errors }) => {
        if (!data || !data.root || !data.root.repoGroups) {
            throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
        }
        return data.root.repoGroups
    })
}
