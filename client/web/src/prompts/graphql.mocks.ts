import type { MockedResponse } from '@apollo/client/testing'

import { getDocumentNode } from '@sourcegraph/http-client'

import {
    PromptVisibility,
    PromptsOrderBy,
    type ChangePromptVisibilityResult,
    type ChangePromptVisibilityVariables,
    type CreatePromptResult,
    type CreatePromptVariables,
    type DeletePromptResult,
    type DeletePromptVariables,
    type PromptFields,
    type PromptResult,
    type PromptVariables,
    type PromptsResult,
    type PromptsVariables,
    type TransferPromptOwnershipResult,
    type TransferPromptOwnershipVariables,
    type UpdatePromptResult,
    type UpdatePromptVariables,
} from '../graphql-operations'
import { viewerAffiliatedNamespacesMock } from '../namespaces/graphql.mocks'

import {
    changePromptVisibilityMutation,
    createPromptMutation,
    deletePromptMutation,
    promptQuery,
    promptsQuery,
    transferPromptOwnershipMutation,
    updatePromptMutation,
} from './graphql'

export const MOCK_PROMPT_FIELDS: PromptFields = {
    __typename: 'Prompt',
    id: '1',
    name: 'my-prompt',
    description: 'My description',
    definition: { text: 'My template text' },
    draft: false,
    owner: {
        __typename: 'User',
        id: 'a',
        namespaceName: 'alice',
    },
    visibility: PromptVisibility.SECRET,
    nameWithOwner: 'alice/my-prompt',
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
    updatedAt: '2024-04-15T17:00:00Z',
    url: '',
    viewerCanAdminister: true,
}

export const promptsMock: MockedResponse<PromptsResult, PromptsVariables> = {
    request: {
        query: getDocumentNode(promptsQuery),
        variables: {
            query: '',
            owner: null,
            viewerIsAffiliated: true,
            includeDrafts: true,
            after: null,
            before: null,
            first: 20,
            last: null,
            orderBy: PromptsOrderBy.PROMPT_UPDATED_AT,
        },
    },
    result: {
        data: {
            prompts: {
                nodes: [
                    {
                        ...MOCK_PROMPT_FIELDS,
                        id: '1',
                        name: 'my-prompt',
                        nameWithOwner: 'alice/my-prompt',
                    },
                    {
                        ...MOCK_PROMPT_FIELDS,
                        id: '2',
                        name: 'another-prompt',
                        description: 'Another',
                        definition: { text: 'Another template text' },
                        nameWithOwner: 'alice/another-prompt',
                    },
                    {
                        ...MOCK_PROMPT_FIELDS,
                        id: '4',
                        name: 'prompt-4',
                        nameWithOwner: 'alice/prompt-4',
                        description: '444',
                        draft: true,
                    },
                    {
                        ...MOCK_PROMPT_FIELDS,
                        id: '5',
                        name: 'prompt-5',
                        nameWithOwner: 'alice/prompt-5',
                        description: '555',
                        draft: true,
                    },
                    {
                        ...MOCK_PROMPT_FIELDS,
                        id: '6',
                        name: 'prompt-6',
                        nameWithOwner: 'alice/prompt-6',
                        description: '666',
                        draft: true,
                        visibility: PromptVisibility.PUBLIC,
                    },
                    {
                        ...MOCK_PROMPT_FIELDS,
                        id: '7',
                        name: 'prompt-7',
                        nameWithOwner: 'alice/prompt-7',
                        description: '777',
                        viewerCanAdminister: false,
                    },
                    {
                        ...MOCK_PROMPT_FIELDS,
                        id: '8',
                        name: 'prompt-8',
                        nameWithOwner: 'alice/prompt-8',
                        description: '888',
                        visibility: PromptVisibility.PUBLIC,
                        viewerCanAdminister: false,
                    },
                    {
                        ...MOCK_PROMPT_FIELDS,
                        id: '9',
                        name: 'prompt-9',
                        nameWithOwner: 'alice/prompt-9',
                        description: '999',
                    },
                    {
                        ...MOCK_PROMPT_FIELDS,
                        id: '10',
                        name: 'prompt-10',
                        nameWithOwner: 'alice/prompt-10',
                        description: '101010',
                    },
                    {
                        ...MOCK_PROMPT_FIELDS,
                        id: '11',
                        name: 'prompt-11',
                        nameWithOwner: 'alice/prompt-11',
                        description: '111111',
                    },
                    {
                        ...MOCK_PROMPT_FIELDS,
                        id: '12',
                        name: 'prompt-12',
                        nameWithOwner: 'alice/prompt-12',
                        description: '121212',
                    },
                ],
                totalCount: 30,
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

export const promptMock: MockedResponse<PromptResult, PromptVariables> = {
    request: {
        query: getDocumentNode(promptQuery),
        variables: { id: '1' },
    },
    result: {
        data: {
            node: {
                ...MOCK_PROMPT_FIELDS,
                __typename: 'Prompt',
                id: '1',
                name: 'my-prompt',
                nameWithOwner: 'alice/my-prompt',
            },
        },
    },
}

const createPromptMock: MockedResponse<CreatePromptResult, CreatePromptVariables> = {
    request: {
        query: getDocumentNode(createPromptMutation),
        variables: {
            input: {
                owner: 'a',
                name: 'my-prompt',
                description: 'My description',
                definitionText: 'My template text',
                draft: false,
                visibility: PromptVisibility.PUBLIC,
            },
        },
    },
    delay: 500,
    result: {
        data: {
            createPrompt: {
                ...MOCK_PROMPT_FIELDS,
                id: '1',
                name: 'my-prompt',
                nameWithOwner: 'alice/my-prompt',
            },
        },
    },
}

const updatePromptMock: MockedResponse<UpdatePromptResult, UpdatePromptVariables> = {
    request: {
        query: getDocumentNode(updatePromptMutation),
        variables: {
            id: '1',
            input: {
                name: 'my-prompt',
                description: 'My description',
                definitionText: 'My template text',
                draft: false,
            },
        },
    },
    delay: 500,
    result: {
        data: {
            updatePrompt: {
                ...MOCK_PROMPT_FIELDS,
                id: '1',
                name: 'my-prompt',
                nameWithOwner: 'alice/my-prompt',
            },
        },
    },
}

const transferPromptOwnershipMock: MockedResponse<TransferPromptOwnershipResult, TransferPromptOwnershipVariables> = {
    request: {
        query: getDocumentNode(transferPromptOwnershipMutation),
        variables: {
            id: '1',
            newOwner: 'b',
        },
    },
    delay: 500,
    result: {
        data: {
            transferPromptOwnership: {
                ...MOCK_PROMPT_FIELDS,
                id: '1',
                name: 'my-prompt',
                nameWithOwner: 'b/my-prompt',
            },
        },
    },
}

const changePromptVisibilityMock: MockedResponse<ChangePromptVisibilityResult, ChangePromptVisibilityVariables> = {
    request: {
        query: getDocumentNode(changePromptVisibilityMutation),
        variables: {
            id: '1',
            newVisibility: PromptVisibility.PUBLIC,
        },
    },
    delay: 500,
    result: {
        data: {
            changePromptVisibility: {
                ...MOCK_PROMPT_FIELDS,
                visibility: PromptVisibility.PUBLIC,
            },
        },
    },
}

const deletePromptMock: MockedResponse<DeletePromptResult, DeletePromptVariables> = {
    request: {
        query: getDocumentNode(deletePromptMutation),
        variables: { id: '1' },
    },
    delay: 500,
    result: {
        data: {
            deletePrompt: {
                alwaysNil: null,
            },
        },
    },
}

export const MOCK_REQUESTS = [
    promptsMock,
    promptMock,
    createPromptMock,
    updatePromptMock,
    deletePromptMock,
    transferPromptOwnershipMock,
    changePromptVisibilityMock,
    viewerAffiliatedNamespacesMock,
]
