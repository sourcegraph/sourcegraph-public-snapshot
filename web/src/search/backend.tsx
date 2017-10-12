import partition from 'lodash/partition'
import 'rxjs/add/observable/defer'
import 'rxjs/add/observable/from'
import 'rxjs/add/observable/merge'
import 'rxjs/add/operator/concat'
import 'rxjs/add/operator/filter'
import 'rxjs/add/operator/ignoreElements'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/mergeMap'
import 'rxjs/add/operator/publishReplay'
import 'rxjs/add/operator/toArray'
import { Observable } from 'rxjs/Observable'
import { queryGraphQL } from '../backend/graphql'
import { Filter, FilterType, RepoFilter, RepoGroupFilter, SearchOptions } from './index'

/** Map from search profile name to repo "URIs" */
const searchProfileRepos = new Map<string, string[]>()

function fetchSearchProfiles(): Observable<GQL.ISearchProfile> {
    return queryGraphQL(`
        query SearchProfiles {
            root {
                searchProfiles {
                    name
                    repositories {
                        uri
                    }
                }
            }
        }
    `)
        .mergeMap(({ data, errors }) => {
            if (!data || !data.root || !data.root.searchProfiles) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            // Save in the cache
            for (const profile of data.root.searchProfiles) {
                searchProfileRepos.set(profile.name, profile.repositories.map(repo => repo.uri))
            }
            return data!.root.searchProfiles
        })
}

export function searchText(params: SearchOptions): Observable<GQL.ISearchResults> {
    // Subscribing to this Observable will execute the fetch lazily and only once
    const searchProfilesFetch = (fetchSearchProfiles().ignoreElements() as Observable<never>).publishReplay().refCount()
    // Get all the repositories that should be searched over
    return Observable.merge(
        // From repo filters
        Observable.from(params.filters)
            .filter((filter: Filter): filter is RepoFilter => filter.type === FilterType.Repo || filter.type === FilterType.UnknownRepo)
            .map(filter => filter.value),
        // From search profiles
        Observable.from(params.filters)
            .filter((filter: Filter): filter is RepoGroupFilter => filter.type === FilterType.RepoGroup)
            .map(filter => filter.value)
            // Try to expand the search profile from the cache
            .mergeMap(name => searchProfileRepos.get(name) ||
                    // If not found, subscribe to the fetch and try again
                    searchProfilesFetch
                        .concat(Observable.defer(() =>
                            // If still not found, ignore
                            Observable.from(searchProfileRepos.get(name) || []),
                        )),
            ),
    )
        .map(repo => ({ repo }))
        .toArray()
        .map(repositories => {
            const filePatterns = params.filters.filter(f => f.type === FilterType.File).map(f => f.value)
            const [excludePatterns, includePatterns] = partition(filePatterns, pattern => pattern[0] === '!')
            const includePattern = includePatterns.length > 0 ? '{' + includePatterns.join(',') + '}' : ''
            const excludePattern = excludePatterns.length > 0 ? '{' + excludePatterns.map(p => p.substr(1)).join(',') + '}' : ''
            return {
                pattern: params.query,
                fileMatchLimit: 500,
                isRegExp: params.matchRegex,
                isWordMatch: params.matchWord,
                repositories,
                isCaseSensitive: params.matchCase,
                includePattern,
                excludePattern,
            }
        })
        .mergeMap(variables => queryGraphQL(`
            query SearchText(
                $pattern: String!,
                $fileMatchLimit: Int!,
                $isRegExp: Boolean!,
                $isWordMatch: Boolean!,
                $repositories: [RepositoryRevision!]!,
                $isCaseSensitive: Boolean!,
                $includePattern: String!,
                $excludePattern: String!,
            ) {
                root {
                    searchRepos(
                        repositories: $repositories,
                        query: {
                            pattern: $pattern,
                            isRegExp: $isRegExp,
                            fileMatchLimit: $fileMatchLimit,
                            isWordMatch: $isWordMatch,
                            isCaseSensitive: $isCaseSensitive,
                            includePattern: $includePattern,
                            excludePattern: $excludePattern,
                    }) {
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
                    }
                }
            }
        `, variables))
        .map(({ data, errors }) => {
            if (!data || !data.root || !data.root.searchRepos) {
                throw Object.assign(new Error((errors || []).map(e => e.message).join('\n')), { errors })
            }
            return data.root.searchRepos
        })
}

export function fetchSuggestions(query: string, filters: Filter[]): Observable<GQL.SearchResult> {
   return queryGraphQL(`
        query SearchRepos($query: String!) {
            root {
                search(query: $query, repositories: $repositories) {
                    ... on Repository {
                        __typename
                        uri
                    }
                    ... on File {
                        __typename
                        name
                    }
                    ... on SearchProfile {
                        __typename
                        name
                        repositories {
                            uri
                        }
                    }
                }
            }
        }
    `, {
        query,
        repositories: filters.filter(f => f.type === FilterType.Repo).map((f: RepoFilter) => f.value),
    })
        .mergeMap(({ data, errors }) => {
            if (!data || !data.root.search) {
                const message = errors ? errors.map(e => e.message).join('\n') : 'Incomplete response from GraphQL search endpoint'
                throw Object.assign(new Error(message), { errors })
            }
            for (const item of data.root.search) {
                // Cache SearchProfile repositories to speed up expanding them on the search results page
                if (item.__typename === 'SearchProfile') {
                    searchProfileRepos.set(item.name, item.repositories.map(repo => repo.uri))
                }
            }
            return data.root.search
        })
}
