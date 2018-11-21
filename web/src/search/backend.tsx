import { isEqual } from 'lodash'
import { Observable } from 'rxjs'
import { catchError, distinctUntilChanged, map, mergeMap, switchMap } from 'rxjs/operators'
import { SearchOptions } from '.'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { gql } from '../../../shared/src/graphql/graphql'
import * as GQL from '../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike } from '../../../shared/src/util/errors'
import { queryGraphQL } from '../backend/graphql'
import { mutateSettingsGraphQL } from '../configuration/backend'
import { viewerSettings } from '../settings/configuration'

export function search(
    options: SearchOptions,
    { extensionsController }: ExtensionsControllerProps
): Observable<GQL.ISearchResults | ErrorLike> {
    /**
     * Emits whenever a search is executed, and whenever an extension registers a query transformer.
     */
    return extensionsController.registries.queryTransformer.transformQuery(options.query).pipe(
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
                                        refs {
                                            name
                                            displayName
                                            prefix
                                            repository {
                                                name
                                            }
                                        }
                                        sourceRefs {
                                            name
                                            displayName
                                            prefix
                                            repository {
                                                name
                                            }
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
                                            id
                                            repository {
                                                name
                                                url
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
                                            url
                                            tree(path: "") {
                                                canonicalURL
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

export function fetchSearchResultStats(options: SearchOptions): Observable<GQL.ISearchResultsStats> {
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
        { query: options.query }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.search || !data.search.stats) {
                throw createAggregateError(errors)
            }
            return data.search.stats
        })
    )
}

export function fetchSuggestions(options: SearchOptions): Observable<GQL.SearchSuggestion> {
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
        { query: options.query }
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
        showOnHomepage
        notify
        notifySlack
        query
    }
`

export function observeSavedQueries(): Observable<GQL.ISavedQuery[]> {
    return viewerSettings.pipe(
        map(config => config['search.savedQueries']),
        distinctUntilChanged((a, b) => isEqual(a, b)),
        mergeMap(fetchSavedQueries)
    )
}

function fetchSavedQueries(): Observable<GQL.ISavedQuery[]> {
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
    description: string,
    query: string,
    showOnHomepage: boolean,
    notify: boolean,
    notifySlack: boolean,
    disableSubscriptionNotifications?: boolean
): Observable<GQL.ISavedQuery> {
    return mutateSettingsGraphQL(
        subject,
        gql`
            mutation CreateSavedQuery(
                $subject: ID!
                $lastID: Int
                $description: String!
                $query: String!
                $showOnHomepage: Boolean
                $notify: Boolean
                $notifySlack: Boolean
                $disableSubscriptionNotifications: Boolean
            ) {
                settingsMutation(input: { subject: $subject, lastID: $lastID }) {
                    createSavedQuery(
                        description: $description
                        query: $query
                        showOnHomepage: $showOnHomepage
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
            showOnHomepage,
            notify,
            notifySlack,
            disableSubscriptionNotifications: disableSubscriptionNotifications || false,
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
    id: GQL.ID,
    description: string,
    query: string,
    showOnHomepage: boolean,
    notify: boolean,
    notifySlack: boolean
): Observable<GQL.ISavedQuery> {
    return mutateSettingsGraphQL(
        subject,
        gql`
            mutation UpdateSavedQuery(
                $subject: ID!
                $lastID: Int
                $id: ID!
                $description: String
                $query: String
                $showOnHomepage: Boolean
                $notify: Boolean
                $notifySlack: Boolean
            ) {
                settingsMutation(input: { subject: $subject, lastID: $lastID }) {
                    updateSavedQuery(
                        id: $id
                        description: $description
                        query: $query
                        showOnHomepage: $showOnHomepage
                        notify: $notify
                        notifySlack: $notifySlack
                    ) {
                        ...SavedQueryFields
                    }
                }
            }
            ${savedQueryFragment}
        `,
        { id, description, query, showOnHomepage, notify, notifySlack }
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
    id: GQL.ID,
    disableSubscriptionNotifications?: boolean
): Observable<void> {
    return mutateSettingsGraphQL(
        subject,
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
        { id, disableSubscriptionNotifications: disableSubscriptionNotifications || false }
    ).pipe(
        map(({ data, errors }) => {
            if (!data || !data.settingsMutation || !data.settingsMutation.deleteSavedQuery) {
                throw createAggregateError(errors)
            }
        })
    )
}
