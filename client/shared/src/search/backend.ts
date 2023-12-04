import { type Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

import { createAggregateError, memoizeObservable } from '@sourcegraph/common'
import { gql, dataOrThrowErrors, isErrorGraphQLResult } from '@sourcegraph/http-client'

import type { AuthenticatedUser } from '../auth'
import {
    type EventLogsDataResult,
    type EventLogsDataVariables,
    type ListSearchContextsResult,
    type ListSearchContextsVariables,
    type IsSearchContextAvailableResult,
    type IsSearchContextAvailableVariables,
    type Scalars,
    type FetchSearchContextResult,
    type FetchSearchContextVariables,
    type CreateSearchContextResult,
    type CreateSearchContextVariables,
    type UpdateSearchContextVariables,
    type UpdateSearchContextResult,
    type DeleteSearchContextVariables,
    type DeleteSearchContextResult,
    type Maybe,
    type FetchSearchContextBySpecResult,
    type FetchSearchContextBySpecVariables,
    type highlightCodeResult,
    type highlightCodeVariables,
    SearchContextsOrderBy,
    type SearchContextFields,
    type DefaultSearchContextSpecResult,
    type DefaultSearchContextSpecVariables,
} from '../graphql-operations'
import type { PlatformContext } from '../platform/context'

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
        viewerHasStarred
        viewerHasAsDefault
        query
        repositories {
            ...SearchContextRepositoryRevisionsFields
        }
    }

    fragment SearchContextRepositoryRevisionsFields on SearchContextRepositoryRevisions {
        repository {
            name
        }
        revisions
    }
`

const searchContextWithSkippableFieldsFragment = gql`
    fragment SearchContextMinimalFields on SearchContext {
        __typename
        id
        name
        spec
        description
        public
        query
        autoDefined
        updatedAt
        viewerCanManage
        viewerHasStarred
        viewerHasAsDefault
        namespace @skip(if: $useMinimalFields) {
            __typename
            id
            namespaceName
        }
        repositories @skip(if: $useMinimalFields) {
            __typename
            repository {
                name
            }
            revisions
        }
    }
`

export function getUserSearchContextNamespaces(
    authenticatedUser: Pick<AuthenticatedUser, 'id' | 'organizations'> | null
): Maybe<Scalars['ID']>[] {
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
    useMinimalFields,
    platformContext,
}: {
    first: number
    query?: string
    namespaces?: Maybe<Scalars['ID']>[]
    after?: string
    orderBy?: SearchContextsOrderBy
    descending?: boolean
    useMinimalFields?: boolean
    platformContext: Pick<PlatformContext, 'requestGraphQL'>
}): Observable<ListSearchContextsResult['searchContexts']> {
    return platformContext
        .requestGraphQL<ListSearchContextsResult, ListSearchContextsVariables>({
            request: gql`
                query ListSearchContexts(
                    $first: Int!
                    $after: String
                    $query: String
                    $namespaces: [ID]
                    $orderBy: SearchContextsOrderBy
                    $descending: Boolean
                    $useMinimalFields: Boolean!
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
                            ...SearchContextMinimalFields
                        }
                        pageInfo {
                            hasNextPage
                            endCursor
                        }
                        totalCount
                    }
                }
                ${searchContextWithSkippableFieldsFragment}
            `,
            variables: {
                first,
                after: after ?? null,
                query: query ?? null,
                namespaces: namespaces ?? [],
                orderBy: orderBy ?? SearchContextsOrderBy.SEARCH_CONTEXT_SPEC,
                descending: descending ?? false,
                useMinimalFields: useMinimalFields ?? false,
            },
            mightContainPrivateInfo: true,
        })
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.searchContexts)
        )
}

export const fetchSearchContext = (
    id: Scalars['ID'],
    platformContext: Pick<PlatformContext, 'requestGraphQL'>
): Observable<SearchContextFields> => {
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

    return platformContext
        .requestGraphQL<FetchSearchContextResult, FetchSearchContextVariables>({
            request: query,
            variables: {
                id,
            },
            mightContainPrivateInfo: true,
        })
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.node as SearchContextFields)
        )
}

export const fetchSearchContextBySpec = (
    spec: string,
    platformContext: Pick<PlatformContext, 'requestGraphQL'>
): Observable<SearchContextFields> => {
    const query = gql`
        query FetchSearchContextBySpec($spec: String!) {
            searchContextBySpec(spec: $spec) {
                ...SearchContextFields
            }
        }
        ${searchContextFragment}
    `

    return platformContext
        .requestGraphQL<FetchSearchContextBySpecResult, FetchSearchContextBySpecVariables>({
            request: query,
            variables: {
                spec,
            },
            mightContainPrivateInfo: true,
        })
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.searchContextBySpec as SearchContextFields)
        )
}

export function createSearchContext(
    variables: CreateSearchContextVariables,
    platformContext: Pick<PlatformContext, 'requestGraphQL'>
): Observable<SearchContextFields> {
    return platformContext
        .requestGraphQL<CreateSearchContextResult, CreateSearchContextVariables>({
            request: gql`
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
            variables,
            mightContainPrivateInfo: true,
        })
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.createSearchContext as SearchContextFields)
        )
}

export function updateSearchContext(
    variables: UpdateSearchContextVariables,
    platformContext: Pick<PlatformContext, 'requestGraphQL'>
): Observable<SearchContextFields> {
    return platformContext
        .requestGraphQL<UpdateSearchContextResult, UpdateSearchContextVariables>({
            request: gql`
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
            variables,
            mightContainPrivateInfo: true,
        })
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.updateSearchContext as SearchContextFields)
        )
}

export function deleteSearchContext(
    id: Scalars['ID'],
    platformContext: Pick<PlatformContext, 'requestGraphQL'>
): Observable<DeleteSearchContextResult> {
    return platformContext
        .requestGraphQL<DeleteSearchContextResult, DeleteSearchContextVariables>({
            request: gql`
                mutation DeleteSearchContext($id: ID!) {
                    deleteSearchContext(id: $id) {
                        alwaysNil
                    }
                }
            `,
            variables: { id },
            mightContainPrivateInfo: false,
        })
        .pipe(map(dataOrThrowErrors))
}

export function isSearchContextAvailable(
    spec: string,
    platformContext: Pick<PlatformContext, 'requestGraphQL'>
): Observable<IsSearchContextAvailableResult['isSearchContextAvailable']> {
    return platformContext
        .requestGraphQL<IsSearchContextAvailableResult, IsSearchContextAvailableVariables>({
            request: gql`
                query IsSearchContextAvailable($spec: String!) {
                    isSearchContextAvailable(spec: $spec)
                }
            `,
            variables: { spec },
            mightContainPrivateInfo: true,
        })
        .pipe(map(result => result.data?.isSearchContextAvailable ?? false))
}

export const highlightCode = memoizeObservable(
    ({
        platformContext,
        ...context
    }: {
        code: string
        fuzzyLanguage: string
        disableTimeout: boolean
        platformContext: Pick<PlatformContext, 'requestGraphQL'>
    }): Observable<string> =>
        platformContext
            .requestGraphQL<highlightCodeResult, highlightCodeVariables>({
                request: gql`
                    query highlightCode($code: String!, $fuzzyLanguage: String!, $disableTimeout: Boolean!) {
                        highlightCode(code: $code, fuzzyLanguage: $fuzzyLanguage, disableTimeout: $disableTimeout)
                    }
                `,
                variables: context,
                mightContainPrivateInfo: true,
            })
            .pipe(
                map(({ data, errors }) => {
                    if (!data?.highlightCode) {
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

function fetchEvents(
    userId: Scalars['ID'],
    first: number,
    eventName: string,
    platformContext: Pick<PlatformContext, 'requestGraphQL'>
): Observable<EventLogResult | null> {
    if (!userId) {
        return of(null)
    }

    const result = platformContext.requestGraphQL<EventLogsDataResult, EventLogsDataVariables>({
        request: gql`
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
        variables: { userId, first: first ?? null, eventName },
        mightContainPrivateInfo: true,
    })

    return result.pipe(
        map(dataOrThrowErrors),
        map((data: EventLogsDataResult): EventLogResult => {
            if (!data.node || data.node.__typename !== 'User') {
                throw new Error('User not found')
            }
            return data.node.eventLogs
        })
    )
}

export function fetchRecentSearches(
    userId: Scalars['ID'],
    first: number,
    platformContext: Pick<PlatformContext, 'requestGraphQL'>
): Observable<EventLogResult | null> {
    return fetchEvents(userId, first, 'SearchResultsQueried', platformContext)
}

export function fetchRecentFileViews(
    userId: Scalars['ID'],
    first: number,
    platformContext: Pick<PlatformContext, 'requestGraphQL'>
): Observable<EventLogResult | null> {
    return fetchEvents(userId, first, 'ViewBlob', platformContext)
}

export function fetchDefaultSearchContextSpec(
    platformContext: Pick<PlatformContext, 'requestGraphQL'>
): Observable<string | null> {
    return platformContext
        .requestGraphQL<DefaultSearchContextSpecResult, DefaultSearchContextSpecVariables>({
            request: gql`
                query DefaultSearchContextSpec {
                    defaultSearchContext {
                        spec
                    }
                }
            `,
            variables: {},
            mightContainPrivateInfo: true,
        })
        .pipe(map(result => (isErrorGraphQLResult(result) ? null : result.data?.defaultSearchContext?.spec ?? null)))
}
