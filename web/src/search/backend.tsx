import { Observable, of } from 'rxjs'
import { catchError, map, mergeMap, switchMap } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike } from '../../../shared/src/util/errors'
import { memoizeObservable } from '../../../shared/src/util/memoizeObservable'
import { mutateGraphQL, queryGraphQL } from '../backend/graphql'
import { USE_CODEMOD } from '../enterprise/codemod'

const genericSearchResultInterfaceFields = gql`
  __typename
  label {
      html
  }
  url
  icon
  detail {
      html
  }
  matches {
      url
      body {
          text
          html
      }
      highlights {
          line
          character
          length
      }
  }
`

export function search(
    query: string,
    version: string,
    patternType: GQL.SearchPatternType,
    { extensionsController }: ExtensionsControllerProps<'services'>
): Observable<GQL.ISearchResults | ErrorLike> {
    /**
     * Emits whenever a search is executed, and whenever an extension registers a query transformer.
     */
    return extensionsController.services.queryTransformer.transformQuery(query).pipe(
        switchMap(query => {
            const codemodActive = USE_CODEMOD
                ? `... on CodemodResult {
                ${genericSearchResultInterfaceFields}
            }`
                : ''
            return queryGraphQL(
                gql`
                    query Search($query: String!, $version: SearchVersion!, $patternType: SearchPatternType!) {
                        search(query: $query, version: $version, patternType: $patternType) {
                            results {
                                __typename
                                limitHit
                                matchCount
                                approximateResultCount
                                missing {
                                    name
                                }
                                cloning {
                                    name
                                }
                                repositoriesCount
                                timedout {
                                    name
                                }
                                indexUnavailable
                                dynamicFilters {
                                    value
                                    label
                                    count
                                    limitHit
                                    kind
                                }
                                results {
                                    __typename
                                    ... on Repository {
                                        id
                                        name
                                        ${genericSearchResultInterfaceFields}
                                    }
                                    ... on FileMatch {
                                        file {
                                            path
                                            url
                                            commit {
                                                oid
                                            }
                                        }
                                        repository {
                                            name
                                            url
                                        }
                                        limitHit
                                        symbols {
                                            name
                                            containerName
                                            url
                                            kind
                                        }
                                        lineMatches {
                                            preview
                                            lineNumber
                                            offsetAndLengths
                                        }
                                    }
                                    ... on CommitSearchResult {
                                        ${genericSearchResultInterfaceFields}
                                    }
                                    ${codemodActive}
                                }
                                alert {
                                    title
                                    description
                                    proposedQueries {
                                        description
                                        query
                                    }
                                }
                                elapsedMilliseconds
                            }
                        }
                    }
                `,
                { query, version, patternType }
            ).pipe(
                map(({ data, errors }) => {
                    if (!data || !data.search || !data.search.results) {
                        throw createAggregateError(errors)
                    }
                    return data.search.results
                }),
                catchError(error => [asError(error)])
            )
        })
    )
}

export function fetchSearchResultStats(query: string): Observable<GQL.ISearchResultsStats> {
    return queryGraphQL(
        gql`
            query SearchResultsStats($query: String!) {
                search(query: $query) {
                    stats {
                        approximateResultCount
                        sparkline
                    }
                }
            }
        `,
        { query }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.search || !data.search.stats) {
                throw createAggregateError(errors)
            }
            return data.search.stats
        })
    )
}

export function fetchSuggestions(query: string): Observable<GQL.SearchSuggestion> {
    return queryGraphQL(
        gql`
            query SearchSuggestions($query: String!) {
                search(query: $query) {
                    suggestions {
                        __typename
                        ... on Repository {
                            name
                        }
                        ... on File {
                            path
                            name
                            isDirectory
                            url
                            repository {
                                name
                            }
                        }
                        ... on Symbol {
                            name
                            containerName
                            url
                            kind
                            location {
                                resource {
                                    path
                                    repository {
                                        name
                                    }
                                }
                            }
                        }
                    }
                }
            }
        `,
        { query }
    ).pipe(
        mergeMap(({ data, errors }) => {
            if (!data || !data.search || !data.search.suggestions) {
                throw createAggregateError(errors)
            }
            return data.search.suggestions
        })
    )
}

export function fetchReposByQuery(query: string): Observable<{ name: string; url: string }[]> {
    return queryGraphQL(
        gql`
            query ReposByQuery($query: String!) {
                search(query: $query) {
                    results {
                        repositories {
                            name
                            url
                        }
                    }
                }
            }
        `,
        { query }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.search || !data.search.results || !data.search.results.repositories) {
                throw createAggregateError(errors)
            }
            return data.search.results.repositories
        })
    )
}

const savedSearchFragment = gql`
    fragment SavedSearchFields on SavedSearch {
        id
        description
        notify
        notifySlack
        query
        userID
        orgID
        slackWebhookURL
    }
`

export function fetchSavedSearches(): Observable<GQL.ISavedSearch[]> {
    return queryGraphQL(gql`
        query savedSearches {
            savedSearches {
                ...SavedSearchFields
            }
        }
        ${savedSearchFragment}
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.savedSearches) {
                throw createAggregateError(errors)
            }
            return data.savedSearches
        })
    )
}

export function fetchSavedSearch(id: GQL.ID): Observable<GQL.ISavedSearch> {
    return queryGraphQL(
        gql`
            query SavedSearch($id: ID!) {
                node(id: $id) {
                    ... on SavedSearch {
                        id
                        description
                        query
                        notify
                        notifySlack
                        slackWebhookURL
                        orgID
                        userID
                    }
                }
            }
        `,
        { id }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.node as GQL.ISavedSearch)
    )
}

export function createSavedSearch(
    description: string,
    query: string,
    notify: boolean,
    notifySlack: boolean,
    userId: GQL.ID | null,
    orgId: GQL.ID | null
): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation CreateSavedSearch(
                $description: String!
                $query: String!
                $notifyOwner: Boolean!
                $notifySlack: Boolean!
                $userID: ID
                $orgID: ID
            ) {
                createSavedSearch(
                    description: $description
                    query: $query
                    notifyOwner: $notifyOwner
                    notifySlack: $notifySlack
                    userID: $userID
                    orgID: $orgID
                ) {
                    ...SavedSearchFields
                }
            }
            ${savedSearchFragment}
        `,
        {
            description,
            query,
            notifyOwner: notify,
            notifySlack,
            userID: userId,
            orgID: orgId,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(() => undefined)
    )
}

export function updateSavedSearch(
    id: GQL.ID,
    description: string,
    query: string,
    notify: boolean,
    notifySlack: boolean,
    userId: GQL.ID | null,
    orgId: GQL.ID | null
): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation UpdateSavedSearch(
                $id: ID!
                $description: String!
                $query: String!
                $notifyOwner: Boolean!
                $notifySlack: Boolean!
                $userID: ID
                $orgID: ID
            ) {
                updateSavedSearch(
                    id: $id
                    description: $description
                    query: $query
                    notifyOwner: $notifyOwner
                    notifySlack: $notifySlack
                    userID: $userID
                    orgID: $orgID
                ) {
                    ...SavedSearchFields
                }
            }
            ${savedSearchFragment}
        `,
        {
            id,
            description,
            query,
            notifyOwner: notify,
            notifySlack,
            userID: userId,
            orgID: orgId,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(() => undefined)
    )
}

export function deleteSavedSearch(id: GQL.ID): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation DeleteSavedSearch($id: ID!) {
                deleteSavedSearch(id: $id) {
                    alwaysNil
                }
            }
        `,
        { id }
    ).pipe(
        map(dataOrThrowErrors),
        map(() => undefined)
    )
}

export const highlightCode = memoizeObservable(
    (ctx: {
        code: string
        fuzzyLanguage: string
        disableTimeout: boolean
        isLightTheme: boolean
    }): Observable<string> =>
        queryGraphQL(
            gql`
                query highlightCode(
                    $code: String!
                    $fuzzyLanguage: String!
                    $disableTimeout: Boolean!
                    $isLightTheme: Boolean!
                ) {
                    highlightCode(
                        code: $code
                        fuzzyLanguage: $fuzzyLanguage
                        disableTimeout: $disableTimeout
                        isLightTheme: $isLightTheme
                    )
                }
            `,
            ctx
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.highlightCode) {
                    throw createAggregateError(errors)
                }
                return data.highlightCode
            })
        ),
    ctx => `${ctx.code}:${ctx.fuzzyLanguage}:${ctx.disableTimeout}:${ctx.isLightTheme}`
)

/**
 * Returns true if search performance and accuracy are limited because this is a
 * single-node Docker deployment that is configured with more than 100 repositories.
 */
export function shouldDisplayPerformanceWarning(deployType: DeployType): Observable<boolean> {
    if (deployType !== 'docker-container') {
        return of(false)
    }
    const manyReposWarningLimit = 100
    return queryGraphQL(
        gql`
            query ManyReposWarning($first: Int) {
                repositories(first: $first) {
                    nodes {
                        id
                    }
                }
            }
        `,
        {
            first: manyReposWarningLimit + 1,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => (data.repositories.nodes || []).length > manyReposWarningLimit)
    )
}
