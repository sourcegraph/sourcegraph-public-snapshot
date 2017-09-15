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
import 'rxjs/add/operator/toPromise'
import { Observable } from 'rxjs/Observable'
import { queryGraphQL } from 'sourcegraph/backend/graphql'
import { FileFilter, FileGlobFilter, Filter, FilterType, RepoFilter, RepoGroupFilter, SearchOptions } from 'sourcegraph/search'

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
            .filter((filter: Filter): filter is RepoFilter => filter.type === FilterType.Repo)
            .map(filter => filter.repoPath),
        // From search profiles
        Observable.from(params.filters)
            .filter((filter: Filter): filter is RepoGroupFilter => filter.type === FilterType.RepoGroup)
            .map(filter => filter.name)
            // Try to expand the search profile from the cache
            .mergeMap(name => searchProfileRepos.get(name) ||
                // If not found, subscribe to the fetch and try again
                searchProfilesFetch
                    .concat(Observable.defer(() =>
                        // If still not found, ignore
                        Observable.from(searchProfileRepos.get(name) || [])
                    ))
            )
    )
        .map(repo => ({ repo }))
        .toArray()
        .map(repositories => ({
            pattern: params.query,
            fileMatchLimit: 500,
            isRegExp: params.matchRegex,
            isWordMatch: params.matchWord,
            repositories,
            isCaseSensitive: params.matchCase,
            includePattern: [
                ...params.filters.filter(f => f.type === FilterType.File).map((f: FileFilter) => f.filePath),
                ...params.filters.filter(f => f.type === FilterType.FileGlob).map((f: FileGlobFilter) => f.glob)
            ].join(','),
            excludePattern: '{.git,**/.git,.svn,**/.svn,.hg,**/.hg,CVS,**/CVS,.DS_Store,**/.DS_Store,node_modules,bower_components,vendor,dist,out,Godeps,third_party}'
        }))
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

export function fetchActiveRepos(): Promise<GQL.IActiveRepoResults | null> {
    return queryGraphQL(`
        query ActiveRepos() {
            root {
                activeRepos() {
                    active
                    inactive
                }
            }
        }
    `).toPromise().then(result => {
        if (!result.data ||
            !result.data.root ||
            !result.data.root.activeRepos) {
            return null
        }
        return result.data.root.activeRepos
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
        repositories: filters.filter(f => f.type === FilterType.Repo).map((f: RepoFilter) => f.repoPath)
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
