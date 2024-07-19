import { type MockedResponse } from '@apollo/client/testing'

import { getDocumentNode } from '@sourcegraph/http-client'

import { type ViewerAffiliatedNamespacesResult, type ViewerAffiliatedNamespacesVariables } from '../graphql-operations'

import { viewerAffiliatedNamespacesQuery } from './useAffiliatedNamespaces'

export const viewerAffiliatedNamespacesMock: MockedResponse<
    ViewerAffiliatedNamespacesResult,
    ViewerAffiliatedNamespacesVariables
> = {
    request: { query: getDocumentNode(viewerAffiliatedNamespacesQuery) },
    result: {
        data: {
            viewer: {
                affiliatedNamespaces: {
                    nodes: [
                        { __typename: 'User', id: 'user1', namespaceName: 'alice' },
                        {
                            __typename: 'Org',
                            id: 'org1',
                            namespaceName: 'abc',
                            displayName: 'ABC',
                        },
                        {
                            __typename: 'Org',
                            id: 'org2',
                            namespaceName: 'xyz',
                            displayName: 'XYZ',
                        },
                    ],
                },
            },
        },
    },
}
