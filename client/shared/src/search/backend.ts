import { Observable, of } from 'rxjs'
import { map } from 'rxjs/operators'

import { createAggregateError, memoizeObservable } from '@sourcegraph/common'
import { gql, dataOrThrowErrors, isErrorGraphQLResult } from '@sourcegraph/http-client'

import { AuthenticatedUser } from '../auth'
import {
    EventLogsDataResult,
    EventLogsDataVariables,
    ListSearchContextsResult,
    ListSearchContextsVariables,
    IsSearchContextAvailableResult,
    IsSearchContextAvailableVariables,
    Scalars,
    CreateSearchContextResult,
    CreateSearchContextVariables,
    UpdateSearchContextVariables,
    UpdateSearchContextResult,
    DeleteSearchContextVariables,
    DeleteSearchContextResult,
    Maybe,
    FetchSearchContextBySpecResult,
    FetchSearchContextBySpecVariables,
    highlightCodeResult,
    highlightCodeVariables,
    SearchContextsOrderBy,
    // SearchContextFields,
    DefaultSearchContextSpecResult,
    DefaultSearchContextSpecVariables,
} from '../graphql-operations'
import { PlatformContext } from '../platform/context'
import { graphql, useFragment } from '../gql'
import { useQuery } from '@apollo/client'

const SearchContextFields = graphql(/* GraphQL */ `
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
            repository {
                name
            }
            revisions
        }
    }
`)

export const SearchContextRepositoryRevisionsFragment = graphql(/* GraphQL */ `
    fragment SearchContextRepositoryRevisionsFields on SearchContextRepositoryRevisions {
        repository {
            name
        }
        revisions
    }
`)

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

export const fetchSearchContextBySpec = (spec: string, platformContext: Pick<PlatformContext, 'requestGraphQL'>) => {
    // This is a bit annoying that the comment is required in VSCode to enable syntax highlighting.
    // Maybe we can figure out a solution that would make it obsolete.
    // But we have a `TypedDocumentNode` a  results of the `graphql` call!
    // When we pass it to Apollo Client hook it already know how to infer TData and TVariables.
    const query = graphql(/* GraphQL */ `
        query FetchSearchContextBySpec($spec: String!) {
            searchContextBySpec(spec: $spec) {
                ...SearchContextFields
            }
        }
    `)

    // No need to pass generated types here as generics `useQuery<TData, TVariables>`.
    // `data` type is already inferred as `FetchSearchContextBySpecQuery | undefined`.
    const { data } = useQuery(query, { variables: { spec: 'wow' } })

    if (data) {
        // TS fails here because initially we do not have access to type fields defined in the fragment.
        console.log(data.searchContextBySpec?.query)

        // Now we can access `res.data.searchContextBySpec.query`.
        // const fragmentFields: SearchContextFieldsFragment | null | undefined
        const fragmentFields = useFragment(SearchContextFields, data.searchContextBySpec)

        // Typescript is happy and tells use that `fragmentFields?.query` is `string | undefined`.
        console.log(fragmentFields?.query)
    }

    return data
}
