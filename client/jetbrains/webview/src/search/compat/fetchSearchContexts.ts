import type { Observable } from 'rxjs'
import { map } from 'rxjs/operators'

import { dataOrThrowErrors, gql } from '@sourcegraph/http-client'
import {
    type ListSearchContextsResult,
    type Maybe,
    type Scalars,
    SearchContextsOrderBy,
} from '@sourcegraph/shared/src/graphql-operations'
import type { PlatformContext } from '@sourcegraph/shared/src/platform/context'

import type { ListSearchContextsCompatResult, ListSearchContextsCompatVariables } from '../../graphql-operations'

// A version of fetchSearchContext that works with instances prior 4.3
export function fetchSearchContextsCompat({
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
        .requestGraphQL<ListSearchContextsCompatResult, ListSearchContextsCompatVariables>({
            request: gql`
                query ListSearchContextsCompat(
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
                            ...SearchContextMinimalFieldsCompat
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
            map(data => {
                const contexts = {
                    ...data.searchContexts,
                    nodes: data.searchContexts.nodes.map((context: any) => ({
                        ...context,
                        viewerHasStarred: false,
                        viewerHasAsDefault: false,
                    })),
                }

                if (!after) {
                    contexts.nodes.unshift({
                        __typename: 'SearchContext',
                        id: 'test',
                        name: 'global',
                        spec: 'global',
                        description: 'All repositories on Sourcegraph',
                        public: true,
                        query: '',
                        autoDefined: true,
                        updatedAt: '2022-10-14T14:33:24Z',
                        viewerCanManage: true,
                    })
                }
                return contexts
            })
        )
}

const searchContextWithSkippableFieldsFragment = gql`
    fragment SearchContextMinimalFieldsCompat on SearchContext {
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
