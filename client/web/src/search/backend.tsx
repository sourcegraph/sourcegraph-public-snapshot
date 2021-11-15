import { Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/shared/src/graphql/graphql'
import * as GQL from '@sourcegraph/shared/src/graphql/schema'
import { createAggregateError } from '@sourcegraph/shared/src/util/errors'
import { memoizeObservable } from '@sourcegraph/shared/src/util/memoizeObservable'

import { AuthenticatedUser } from '../auth'
import { queryGraphQL, requestGraphQL } from '../backend/graphql'
import {
    EventLogsDataResult,
    EventLogsDataVariables,
    CreateSavedSearchResult,
    CreateSavedSearchVariables,
    DeleteSavedSearchResult,
    DeleteSavedSearchVariables,
    UpdateSavedSearchResult,
    UpdateSavedSearchVariables,
    ListSearchContextsResult,
    ListSearchContextsVariables,
    AutoDefinedSearchContextsResult,
    AutoDefinedSearchContextsVariables,
    IsSearchContextAvailableResult,
    IsSearchContextAvailableVariables,
    Scalars,
    FetchSearchContextResult,
    FetchSearchContextVariables,
    CreateSearchContextResult,
    CreateSearchContextVariables,
    UpdateSearchContextVariables,
    UpdateSearchContextResult,
    DeleteSearchContextVariables,
    DeleteSearchContextResult,
    Maybe,
    FetchSearchContextBySpecResult,
    FetchSearchContextBySpecVariables,
} from '../graphql-operations'

const searchContextFragment = gql`
    fragment SearchContextFields on SearchContext {
        __typename
        id
        name
        namespace {
            __typename
            id
            namespaceName
        }
        spec
        description
        public
        autoDefined
        updatedAt
        viewerCanManage
        repositories {
            __typename
            repository {
                name
            }
            revisions
        }
    }
`

export function fetchAutoDefinedSearchContexts(): Observable<
    AutoDefinedSearchContextsResult['autoDefinedSearchContexts']
> {
    return requestGraphQL<AutoDefinedSearchContextsResult, AutoDefinedSearchContextsVariables>(gql`
        query AutoDefinedSearchContexts {
            autoDefinedSearchContexts {
                ...SearchContextFields
            }
        }
        ${searchContextFragment}
    `).pipe(
        map(dataOrThrowErrors),
        map(({ autoDefinedSearchContexts }) => autoDefinedSearchContexts as GQL.ISearchContext[])
    )
}

export function getUserSearchContextNamespaces(authenticatedUser: AuthenticatedUser | null): Maybe<Scalars['ID']>[] {
    return authenticatedUser
        ? [null, authenticatedUser.id, ...authenticatedUser.organizations.nodes.map(org => org.id)]
        : [null]
}

export function fetchSearchContexts({
    first,
    namespaces,
    query,
    after,
    orderBy,
    descending,
}: {
    first: number
    query?: string
    namespaces?: Maybe<Scalars['ID']>[]
    after?: string
    orderBy?: GQL.SearchContextsOrderBy
    descending?: boolean
}): Observable<ListSearchContextsResult['searchContexts']> {
    return requestGraphQL<ListSearchContextsResult, ListSearchContextsVariables>(
        gql`
            query ListSearchContexts(
                $first: Int!
                $after: String
                $query: String
                $namespaces: [ID]
                $orderBy: SearchContextsOrderBy
                $descending: Boolean
            ) {
                searchContexts(
                    first: $first
                    after: $after
                    query: $query
                    namespaces: $namespaces
                    orderBy: $orderBy
                    descending: $descending
                ) {
                    nodes {
                        ...SearchContextFields
                    }
                    pageInfo {
                        hasNextPage
                        endCursor
                    }
                    totalCount
                }
            }
            ${searchContextFragment}
        `,
        {
            first,
            after: after ?? null,
            query: query ?? null,
            namespaces: namespaces ?? [],
            orderBy: orderBy ?? GQL.SearchContextsOrderBy.SEARCH_CONTEXT_SPEC,
            descending: descending ?? false,
        }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.searchContexts)
    )
}

export const fetchSearchContext = (id: Scalars['ID']): Observable<GQL.ISearchContext> => {
    const query = gql`
        query FetchSearchContext($id: ID!) {
            node(id: $id) {
                ... on SearchContext {
                    ...SearchContextFields
                }
            }
        }
        ${searchContextFragment}
    `

    return requestGraphQL<FetchSearchContextResult, FetchSearchContextVariables>(query, {
        id,
    }).pipe(
        map(dataOrThrowErrors),
        map(data => data.node as GQL.ISearchContext)
    )
}

export const fetchSearchContextBySpec = (spec: string): Observable<GQL.ISearchContext> => {
    const query = gql`
        query FetchSearchContextBySpec($spec: String!) {
            searchContextBySpec(spec: $spec) {
                ...SearchContextFields
            }
        }
        ${searchContextFragment}
    `

    return requestGraphQL<FetchSearchContextBySpecResult, FetchSearchContextBySpecVariables>(query, {
        spec,
    }).pipe(
        map(dataOrThrowErrors),
        map(data => data.searchContextBySpec as GQL.ISearchContext)
    )
}

export function createSearchContext(variables: CreateSearchContextVariables): Observable<GQL.ISearchContext> {
    return requestGraphQL<CreateSearchContextResult, CreateSearchContextVariables>(
        gql`
            mutation CreateSearchContext(
                $searchContext: SearchContextInput!
                $repositories: [SearchContextRepositoryRevisionsInput!]!
            ) {
                createSearchContext(searchContext: $searchContext, repositories: $repositories) {
                    ...SearchContextFields
                }
            }
            ${searchContextFragment}
        `,
        variables
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.createSearchContext as GQL.ISearchContext)
    )
}

export function updateSearchContext(variables: UpdateSearchContextVariables): Observable<GQL.ISearchContext> {
    return requestGraphQL<UpdateSearchContextResult, UpdateSearchContextVariables>(
        gql`
            mutation UpdateSearchContext(
                $id: ID!
                $searchContext: SearchContextEditInput!
                $repositories: [SearchContextRepositoryRevisionsInput!]!
            ) {
                updateSearchContext(id: $id, searchContext: $searchContext, repositories: $repositories) {
                    ...SearchContextFields
                }
            }
            ${searchContextFragment}
        `,
        variables
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.updateSearchContext as GQL.ISearchContext)
    )
}

export function deleteSearchContext(id: GQL.ID): Observable<DeleteSearchContextResult> {
    return requestGraphQL<DeleteSearchContextResult, DeleteSearchContextVariables>(
        gql`
            mutation DeleteSearchContext($id: ID!) {
                deleteSearchContext(id: $id) {
                    alwaysNil
                }
            }
        `,
        { id }
    ).pipe(map(dataOrThrowErrors))
}

export function isSearchContextAvailable(
    spec: string
): Observable<IsSearchContextAvailableResult['isSearchContextAvailable']> {
    return requestGraphQL<IsSearchContextAvailableResult, IsSearchContextAvailableVariables>(
        gql`
            query IsSearchContextAvailable($spec: String!) {
                isSearchContextAvailable(spec: $spec)
            }
        `,
        { spec }
    ).pipe(map(result => result.data?.isSearchContextAvailable ?? false))
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
        namespace {
            __typename
            id
            namespaceName
        }
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

export function fetchSavedSearch(id: Scalars['ID']): Observable<GQL.ISavedSearch> {
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
                        namespace {
                            id
                        }
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
    userId: Scalars['ID'] | null,
    orgId: Scalars['ID'] | null
): Observable<void> {
    return requestGraphQL<CreateSavedSearchResult, CreateSavedSearchVariables>(
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
    id: Scalars['ID'],
    description: string,
    query: string,
    notify: boolean,
    notifySlack: boolean,
    userId: Scalars['ID'] | null,
    orgId: Scalars['ID'] | null
): Observable<void> {
    return requestGraphQL<UpdateSavedSearchResult, UpdateSavedSearchVariables>(
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

export function deleteSavedSearch(id: Scalars['ID']): Observable<void> {
    return requestGraphQL<DeleteSavedSearchResult, DeleteSavedSearchVariables>(
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
    (context: { code: string; fuzzyLanguage: string; disableTimeout: boolean }): Observable<string> =>
        queryGraphQL(
            gql`
                query highlightCode($code: String!, $fuzzyLanguage: String!, $disableTimeout: Boolean!) {
                    highlightCode(code: $code, fuzzyLanguage: $fuzzyLanguage, disableTimeout: $disableTimeout)
                }
            `,
            context
        ).pipe(
            map(({ data, errors }) => {
                if (!data || !data.highlightCode) {
                    throw createAggregateError(errors)
                }
                return data.highlightCode
            })
        ),
    context => `${context.code}:${context.fuzzyLanguage}:${String(context.disableTimeout)}`
)

export interface EventLogResult {
    totalCount: number
    nodes: { argument: string | null; timestamp: string; url: string }[]
    pageInfo: { hasNextPage: boolean }
}

function fetchEvents(userId: Scalars['ID'], first: number, eventName: string): Observable<EventLogResult | null> {
    if (!userId) {
        return of(null)
    }

    const result = requestGraphQL<EventLogsDataResult, EventLogsDataVariables>(
        gql`
            query EventLogsData($userId: ID!, $first: Int, $eventName: String!) {
                node(id: $userId) {
                    ... on User {
                        __typename
                        eventLogs(first: $first, eventName: $eventName) {
                            nodes {
                                argument
                                timestamp
                                url
                            }
                            pageInfo {
                                hasNextPage
                            }
                            totalCount
                        }
                    }
                }
            }
        `,
        { userId, first: first ?? null, eventName }
    )

    return result.pipe(
        map(dataOrThrowErrors),
        map(
            (data: EventLogsDataResult): EventLogResult => {
                if (!data.node || data.node.__typename !== 'User') {
                    throw new Error('User not found')
                }
                return data.node.eventLogs
            }
        )
    )
}

export function fetchRecentSearches(userId: Scalars['ID'], first: number): Observable<EventLogResult | null> {
    return fetchEvents(userId, first, 'SearchResultsQueried')
}

export function fetchRecentFileViews(userId: Scalars['ID'], first: number): Observable<EventLogResult | null> {
    return fetchEvents(userId, first, 'ViewBlob')
}
