import { SearchGraphQlOperations } from '@sourcegraph/search'
import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { sharedGraphQlResults } from '@sourcegraph/shared/src/testing/integration/graphQlResults'

import { VSCodeGraphQlOperations } from '../src/graphql-operations'

/**
 * Predefined results for GraphQL requests that are made on almost every user flow.
 */
export const commonVSCodeGraphQlResults: Partial<
    VSCodeGraphQlOperations & SharedGraphQlOperations & SearchGraphQlOperations
> = {
    ...sharedGraphQlResults,
    LogEvents: () => ({
        __typename: 'Mutation',
        logEvents: null,
    }),
    CurrentAuthState: () => ({
        __typename: 'Query',
        currentUser: null,
    }),
    ListSearchContexts: () => ({
        searchContexts: {
            nodes: [],
            totalCount: 0,
            pageInfo: { hasNextPage: false, endCursor: null },
        },
    }),
    AutoDefinedSearchContexts: () => ({
        autoDefinedSearchContexts: [
            {
                __typename: 'SearchContext',
                id: '1',
                spec: 'global',
                name: 'global',
                namespace: null,
                autoDefined: true,
                public: true,
                description: 'All repositories on Sourcegraph',
                updatedAt: '2021-03-15T19:39:11Z',
                repositories: [],
                query: '',
                viewerCanManage: false,
            },
            {
                __typename: 'SearchContext',
                id: '2',
                spec: '@test',
                name: 'test',
                namespace: {
                    __typename: 'User',
                    id: 'u1',
                    namespaceName: 'test',
                },
                autoDefined: true,
                public: true,
                description: 'Your repositories on Sourcegraph',
                updatedAt: '2021-03-15T19:39:11Z',
                repositories: [],
                query: '',
                viewerCanManage: false,
            },
        ],
    }),
}
