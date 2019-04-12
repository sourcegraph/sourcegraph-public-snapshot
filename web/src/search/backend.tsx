import { Observable, of } from 'rxjs'
import { catchError, map, mergeMap, switchMap } from 'rxjs/operators'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { dataOrThrowErrors, gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike } from '../../../shared/src/util/errors'
import { memoizeObservable } from '../../../shared/src/util/memoizeObservable'
import { mutateGraphQL, queryGraphQL } from '../backend/graphql'

export function search(
    query: string,
    { extensionsController }: ExtensionsControllerProps<'services'>
): Observable<GQL.ISearchResults | ErrorLike> {
    /**
     * Emits whenever a search is executed, and whenever an extension registers a query transformer.
     */
    return extensionsController.services.queryTransformer.transformQuery(query).pipe(
        switchMap(query =>
            queryGraphQL(
                gql`
                    query Search($query: String!) {
                        search(query: $query) {
                            results {
                                __typename
                                limitHit
                                resultCount
                                approximateResultCount
                                missing {
                                    name
                                }
                                cloning {
                                    name
                                }
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
                                    ... on Repository {
                                        __typename
                                        id
                                        name
                                        url
                                        label {
                                            html
                                        }
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
                                    }
                                    ... on FileMatch {
                                        __typename
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
                                    }
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
                { query }
            ).pipe(
                map(({ data, errors }) => {
                    if (!data || !data.search || !data.search.results) {
                        throw createAggregateError(errors)
                    }
                    return data.search.results
                }),
                catchError(error => [asError(error)])
            )
        )
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

const savedQueryFragment = gql`
    fragment SavedQueryFields on SavedQuery {
        id
        subject {
            ... on Site {
                id
                viewerCanAdminister
            }
            ... on Org {
                id
                viewerCanAdminister
            }
            ... on User {
                id
                viewerCanAdminister
            }
        }
        index
        description
        notify
        notifySlack
        query
    }
`

export function fetchSavedQueries(): Observable<GQL.ISavedQuery[]> {
    return queryGraphQL(gql`
        query SavedQueries {
            savedQueries {
                ...SavedQueryFields
            }
        }
        ${savedQueryFragment}
    `).pipe(
        map(({ data, errors }) => {
            if (!data || !data.savedQueries) {
                throw createAggregateError(errors)
            }
            return data.savedQueries
        })
    )
}

export function createSavedQuery(
    subject: GQL.SettingsSubject | GQL.ISettingsSubject | { id: GQL.ID },
    settingsLastID: number | null,
    description: string,
    query: string,
    notify: boolean,
    notifySlack: boolean,
    disableSubscriptionNotifications?: boolean
): Observable<GQL.ISavedQuery> {
    return mutateGraphQL(
        gql`
            mutation CreateSavedQuery(
                $subject: ID!
                $lastID: Int
                $description: String!
                $query: String!
                $notify: Boolean
                $notifySlack: Boolean
                $disableSubscriptionNotifications: Boolean
            ) {
                settingsMutation(input: { subject: $subject, lastID: $lastID }) {
                    createSavedQuery(
                        description: $description
                        query: $query
                        notify: $notify
                        notifySlack: $notifySlack
                        disableSubscriptionNotifications: $disableSubscriptionNotifications
                    ) {
                        ...SavedQueryFields
                    }
                }
            }
            ${savedQueryFragment}
        `,
        {
            description,
            query,
            notify,
            notifySlack,
            disableSubscriptionNotifications: disableSubscriptionNotifications || false,
            subject: subject.id,
            lastID: settingsLastID,
        }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.settingsMutation || !data.settingsMutation.createSavedQuery) {
                throw createAggregateError(errors)
            }
            return data.settingsMutation.createSavedQuery
        })
    )
}

export function updateSavedQuery(
    subject: GQL.SettingsSubject | GQL.ISettingsSubject | { id: GQL.ID },
    settingsLastID: number | null,
    id: GQL.ID,
    description: string,
    query: string,
    notify: boolean,
    notifySlack: boolean
): Observable<GQL.ISavedQuery> {
    return mutateGraphQL(
        gql`
            mutation UpdateSavedQuery(
                $subject: ID!
                $lastID: Int
                $id: ID!
                $description: String
                $query: String
                $notify: Boolean
                $notifySlack: Boolean
            ) {
                settingsMutation(input: { subject: $subject, lastID: $lastID }) {
                    updateSavedQuery(
                        id: $id
                        description: $description
                        query: $query
                        notify: $notify
                        notifySlack: $notifySlack
                    ) {
                        ...SavedQueryFields
                    }
                }
            }
            ${savedQueryFragment}
        `,
        { id, description, query, notify, notifySlack, subject: subject.id, lastID: settingsLastID }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.settingsMutation || !data.settingsMutation.updateSavedQuery) {
                throw createAggregateError(errors)
            }
            return data.settingsMutation.updateSavedQuery
        })
    )
}

export function deleteSavedQuery(
    subject: GQL.SettingsSubject | GQL.ISettingsSubject | { id: GQL.ID },
    settingsLastID: number | null,
    id: GQL.ID,
    disableSubscriptionNotifications?: boolean
): Observable<void> {
    return mutateGraphQL(
        gql`
            mutation DeleteSavedQuery(
                $subject: ID!
                $lastID: Int
                $id: ID!
                $disableSubscriptionNotifications: Boolean
            ) {
                settingsMutation(input: { subject: $subject, lastID: $lastID }) {
                    deleteSavedQuery(id: $id, disableSubscriptionNotifications: $disableSubscriptionNotifications) {
                        alwaysNil
                    }
                }
            }
        `,
        {
            id,
            disableSubscriptionNotifications: disableSubscriptionNotifications || false,
            subject: subject.id,
            lastID: settingsLastID,
        }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.settingsMutation || !data.settingsMutation.deleteSavedQuery) {
                throw createAggregateError(errors)
            }
        })
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
export function displayPerformanceWarning(): Observable<boolean> {
    if (window.context.deployType !== 'docker-container') {
        return of(false)
    }
    const manyReposWarningLimit = 100
    return queryGraphQL(
        gql`
            query ManyReposWarning($first: Int) {
                repositories(enabled: true, first: $first) {
                    nodes {
                        enabled
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
