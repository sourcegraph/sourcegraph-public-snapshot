import type { MockedResponse } from '@apollo/client/testing'

import { getDocumentNode } from '@sourcegraph/http-client'

import {
    WorkflowsOrderBy,
    type CreateWorkflowResult,
    type CreateWorkflowVariables,
    type DeleteWorkflowResult,
    type DeleteWorkflowVariables,
    type TransferWorkflowOwnershipResult,
    type TransferWorkflowOwnershipVariables,
    type UpdateWorkflowResult,
    type UpdateWorkflowVariables,
    type WorkflowFields,
    type WorkflowResult,
    type WorkflowVariables,
    type WorkflowsResult,
    type WorkflowsVariables,
} from '../graphql-operations'

import {
    createWorkflowMutation,
    deleteWorkflowMutation,
    transferWorkflowOwnershipMutation,
    updateWorkflowMutation,
    workflowQuery,
    workflowsQuery,
} from './graphql'

const WORKFLOW_FIELDS: Pick<
    WorkflowFields,
    | '__typename'
    | 'description'
    | 'template'
    | 'draft'
    | 'owner'
    | 'createdBy'
    | 'createdAt'
    | 'updatedBy'
    | 'updatedAt'
    | 'url'
    | 'viewerCanAdminister'
> = {
    __typename: 'Workflow',
    description: 'My description',
    template: { text: 'My template text' },
    draft: false,
    owner: {
        __typename: 'User',
        id: 'a',
        namespaceName: 'alice',
    },
    createdBy: {
        __typename: 'User',
        id: 'a',
        username: 'alice',
    },
    createdAt: '2024-04-12T15:00:00Z',
    updatedBy: {
        __typename: 'User',
        id: 'a',
        username: 'alice',
    },
    updatedAt: '2024-04-15T17:00:00Z',
    url: '',
    viewerCanAdminister: true,
}

const workflowsMock: MockedResponse<WorkflowsResult, WorkflowsVariables> = {
    request: {
        query: getDocumentNode(workflowsQuery),
        variables: {
            query: null,
            owner: '1',
            viewerIsAffiliated: null,
            includeDrafts: true,
            after: null,
            before: null,
            first: 100,
            last: null,
            orderBy: WorkflowsOrderBy.WORKFLOW_UPDATED_AT,
        },
    },
    result: {
        data: {
            workflows: {
                nodes: [
                    {
                        ...WORKFLOW_FIELDS,
                        id: '1',
                        name: 'my-workflow',
                        nameWithOwner: 'alice/my-workflow',
                    },
                    {
                        ...WORKFLOW_FIELDS,
                        id: '2',
                        name: 'another-workflow',
                        description: 'Another',
                        template: { text: 'Another template text' },
                        nameWithOwner: 'alice/another-workflow',
                    },
                    {
                        ...WORKFLOW_FIELDS,
                        id: '4',
                        name: 'workflow-4',
                        nameWithOwner: 'alice/workflow-4',
                        description: '444',
                    },
                    {
                        ...WORKFLOW_FIELDS,
                        id: '5',
                        name: 'workflow-5',
                        nameWithOwner: 'alice/workflow-5',
                        description: '555',
                    },
                    {
                        ...WORKFLOW_FIELDS,
                        id: '6',
                        name: 'workflow-6',
                        nameWithOwner: 'alice/workflow-6',
                        description: '666',
                    },
                    {
                        ...WORKFLOW_FIELDS,
                        id: '7',
                        name: 'workflow-7',
                        nameWithOwner: 'alice/workflow-7',
                        description: '777',
                    },
                    {
                        ...WORKFLOW_FIELDS,
                        id: '8',
                        name: 'workflow-8',
                        nameWithOwner: 'alice/workflow-8',
                        description: '888',
                    },
                    {
                        ...WORKFLOW_FIELDS,
                        id: '9',
                        name: 'workflow-9',
                        nameWithOwner: 'alice/workflow-9',
                        description: '999',
                    },
                    {
                        ...WORKFLOW_FIELDS,
                        id: '10',
                        name: 'workflow-10',
                        nameWithOwner: 'alice/workflow-10',
                        description: '101010',
                    },
                    {
                        ...WORKFLOW_FIELDS,
                        id: '11',
                        name: 'workflow-11',
                        nameWithOwner: 'alice/workflow-11',
                        description: '111111',
                    },
                    {
                        ...WORKFLOW_FIELDS,
                        id: '12',
                        name: 'workflow-12',
                        nameWithOwner: 'alice/workflow-12',
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

const workflowMock: MockedResponse<WorkflowResult, WorkflowVariables> = {
    request: {
        query: getDocumentNode(workflowQuery),
        variables: { id: '1' },
    },
    result: {
        data: {
            node: {
                ...WORKFLOW_FIELDS,
                __typename: 'Workflow',
                id: '1',
                name: 'my-workflow',
                nameWithOwner: 'alice/my-workflow',
            },
        },
    },
}

const createWorkflowMock: MockedResponse<CreateWorkflowResult, CreateWorkflowVariables> = {
    request: {
        query: getDocumentNode(createWorkflowMutation),
        variables: {
            input: {
                owner: 'a',
                name: 'my-workflow',
                description: 'My description',
                templateText: 'My template text',
                draft: false,
            },
        },
    },
    delay: 500,
    result: {
        data: {
            createWorkflow: {
                ...WORKFLOW_FIELDS,
                id: '1',
                name: 'my-workflow',
                nameWithOwner: 'alice/my-workflow',
            },
        },
    },
}

const updateWorkflowMock: MockedResponse<UpdateWorkflowResult, UpdateWorkflowVariables> = {
    request: {
        query: getDocumentNode(updateWorkflowMutation),
        variables: {
            id: '1',
            input: {
                name: 'my-workflow',
                description: 'My description',
                templateText: 'My template text',
                draft: false,
            },
        },
    },
    delay: 500,
    result: {
        data: {
            updateWorkflow: {
                ...WORKFLOW_FIELDS,
                id: '1',
                name: 'my-workflow',
                nameWithOwner: 'alice/my-workflow',
            },
        },
    },
}

const transferWorkflowOwnershipMock: MockedResponse<
    TransferWorkflowOwnershipResult,
    TransferWorkflowOwnershipVariables
> = {
    request: {
        query: getDocumentNode(transferWorkflowOwnershipMutation),
        variables: {
            id: '1',
            newOwner: 'b',
        },
    },
    delay: 500,
    result: {
        data: {
            transferWorkflowOwnership: {
                ...WORKFLOW_FIELDS,
                id: '1',
                name: 'my-workflow',
                nameWithOwner: 'b/my-workflow',
            },
        },
    },
}

const deleteWorkflowMock: MockedResponse<DeleteWorkflowResult, DeleteWorkflowVariables> = {
    request: {
        query: getDocumentNode(deleteWorkflowMutation),
        variables: { id: '1' },
    },
    delay: 500,
    result: {
        data: {
            deleteWorkflow: {
                alwaysNil: null,
            },
        },
    },
}

export const MOCK_REQUESTS = [
    workflowsMock,
    workflowMock,
    createWorkflowMock,
    updateWorkflowMock,
    transferWorkflowOwnershipMock,
    deleteWorkflowMock,
]
