import type { MockedResponse } from '@apollo/client/testing'

import { getDocumentNode } from '@sourcegraph/http-client'

import {
    SavedSearchVisibility,
    SavedSearchesOrderBy,
    type ChangeSavedSearchVisibilityResult,
    type ChangeSavedSearchVisibilityVariables,
    type CreateSavedSearchResult,
    type CreateSavedSearchVariables,
    type DeleteSavedSearchResult,
    type DeleteSavedSearchVariables,
    type SavedSearchFields,
    type SavedSearchResult,
    type SavedSearchVariables,
    type SavedSearchesResult,
    type SavedSearchesVariables,
    type TransferSavedSearchOwnershipResult,
    type TransferSavedSearchOwnershipVariables,
    type UpdateSavedSearchResult,
    type UpdateSavedSearchVariables,
} from '../graphql-operations'
import { viewerAffiliatedNamespacesMock } from '../namespaces/graphql.mocks'

import {
    changeSavedSearchVisibilityMutation,
    createSavedSearchMutation,
    deleteSavedSearchMutation,
    savedSearchQuery,
    savedSearchesQuery,
    transferSavedSearchOwnershipMutation,
    updateSavedSearchMutation,
} from './graphql'

export const MOCK_SAVED_SEARCH_FIELDS: SavedSearchFields = {
    __typename: 'SavedSearch',
    id: '1',
    description: 'My description',
    query: 'my repo:query',
    draft: false,
    owner: {
        __typename: 'User',
        id: 'a',
        namespaceName: 'alice',
    },
    visibility: SavedSearchVisibility.SECRET,
    createdBy: {
        __typename: 'User',
        id: 'a',
        username: 'alice',
        url: '',
    },
    createdAt: '2024-04-12T15:00:00Z',
    updatedBy: {
        __typename: 'User',
        id: 'a',
        username: 'alice',
        url: '',
    },
    updatedAt: '2020-04-21T10:10:10Z',
    url: '',
    viewerCanAdminister: true,
}

export const savedSearchesMock: MockedResponse<SavedSearchesResult, SavedSearchesVariables> = {
    request: {
        query: getDocumentNode(savedSearchesQuery),
        variables: {
            query: '',
            viewerIsAffiliated: true,
            includeDrafts: true,
            owner: null,
            after: null,
            before: null,
            first: 20,
            last: null,
            orderBy: SavedSearchesOrderBy.SAVED_SEARCH_UPDATED_AT,
        },
    },
    result: {
        data: {
            savedSearches: {
                nodes: [
                    { ...MOCK_SAVED_SEARCH_FIELDS, id: '1' },
                    {
                        ...MOCK_SAVED_SEARCH_FIELDS,
                        id: '2',
                        description: 'Another',
                        query: 'foo type:diff repo:bar',
                    },
                    {
                        ...MOCK_SAVED_SEARCH_FIELDS,
                        id: '3',
                        description: 'Yet another with a longer description that is very long',
                        query: 'foo type:diff repo:bar and a long:query repo:bar and a long:query',
                        draft: true,
                        visibility: SavedSearchVisibility.PUBLIC,
                    },
                    {
                        ...MOCK_SAVED_SEARCH_FIELDS,
                        id: '4',
                        description: '444',
                        draft: true,
                    },
                    {
                        ...MOCK_SAVED_SEARCH_FIELDS,
                        id: '5',
                        description: '555',
                        visibility: SavedSearchVisibility.PUBLIC,
                        viewerCanAdminister: false,
                    },
                    {
                        ...MOCK_SAVED_SEARCH_FIELDS,
                        id: '6',
                        description: '666',
                        visibility: SavedSearchVisibility.PUBLIC,
                        viewerCanAdminister: false,
                    },
                    {
                        ...MOCK_SAVED_SEARCH_FIELDS,
                        id: '7',
                        description: '777',
                        viewerCanAdminister: false,
                    },
                    {
                        ...MOCK_SAVED_SEARCH_FIELDS,
                        id: '8',
                        description: '888',
                    },
                    {
                        ...MOCK_SAVED_SEARCH_FIELDS,
                        id: '9',
                        description: '999',
                    },
                    {
                        ...MOCK_SAVED_SEARCH_FIELDS,
                        id: '10',
                        description: '101010',
                    },
                    {
                        ...MOCK_SAVED_SEARCH_FIELDS,
                        id: '11',
                        description: '111111',
                    },
                    {
                        ...MOCK_SAVED_SEARCH_FIELDS,
                        id: '12',
                        description: '121212',
                    },
                ],
                totalCount: 2,
                pageInfo: {
                    hasNextPage: true,
                    hasPreviousPage: false,
                    endCursor: '',
                    startCursor: '',
                },
            },
        },
    },
}

export const savedSearchMock: MockedResponse<SavedSearchResult, SavedSearchVariables> = {
    request: {
        query: getDocumentNode(savedSearchQuery),
        variables: { id: '1' },
    },
    result: {
        data: {
            node: {
                ...MOCK_SAVED_SEARCH_FIELDS,
                __typename: 'SavedSearch',
                id: '1',
            },
        },
    },
}

const createSavedSearchMock: MockedResponse<CreateSavedSearchResult, CreateSavedSearchVariables> = {
    request: {
        query: getDocumentNode(createSavedSearchMutation),
        variables: {
            input: {
                owner: 'a',
                description: 'My description',
                query: 'my repo:query',
                draft: false,
                visibility: SavedSearchVisibility.PUBLIC,
            },
        },
    },
    delay: 500,
    result: {
        data: {
            createSavedSearch: {
                ...MOCK_SAVED_SEARCH_FIELDS,
                id: '1',
                draft: false,
            },
        },
    },
}

const updateSavedSearchMock: MockedResponse<UpdateSavedSearchResult, UpdateSavedSearchVariables> = {
    request: {
        query: getDocumentNode(updateSavedSearchMutation),
        variables: {
            id: '1',
            input: {
                description: 'My description',
                query: 'my repo:query',
                draft: true,
            },
        },
    },
    delay: 500,
    result: {
        data: {
            updateSavedSearch: {
                ...MOCK_SAVED_SEARCH_FIELDS,
                id: '1',
                draft: true,
            },
        },
    },
}

const deleteSavedSearchMock: MockedResponse<DeleteSavedSearchResult, DeleteSavedSearchVariables> = {
    request: {
        query: getDocumentNode(deleteSavedSearchMutation),
        variables: { id: '1' },
    },
    delay: 500,
    result: {
        data: {
            deleteSavedSearch: {
                alwaysNil: null,
            },
        },
    },
}

const transferSavedSearchOwnershipMock: MockedResponse<
    TransferSavedSearchOwnershipResult,
    TransferSavedSearchOwnershipVariables
> = {
    request: {
        query: getDocumentNode(transferSavedSearchOwnershipMutation),
        variables: {
            id: '1',
            newOwner: 'b',
        },
    },
    delay: 500,
    result: {
        data: {
            transferSavedSearchOwnership: {
                ...MOCK_SAVED_SEARCH_FIELDS,
                id: '1',
            },
        },
    },
}

const changeSavedSearchVisibilityMock: MockedResponse<
    ChangeSavedSearchVisibilityResult,
    ChangeSavedSearchVisibilityVariables
> = {
    request: {
        query: getDocumentNode(changeSavedSearchVisibilityMutation),
        variables: {
            id: '1',
            newVisibility: SavedSearchVisibility.PUBLIC,
        },
    },
    delay: 500,
    result: {
        data: {
            changeSavedSearchVisibility: {
                ...MOCK_SAVED_SEARCH_FIELDS,
                visibility: SavedSearchVisibility.PUBLIC,
            },
        },
    },
}

export const MOCK_REQUESTS = [
    savedSearchesMock,
    savedSearchMock,
    createSavedSearchMock,
    updateSavedSearchMock,
    deleteSavedSearchMock,
    transferSavedSearchOwnershipMock,
    changeSavedSearchVisibilityMock,
    viewerAffiliatedNamespacesMock,
]
