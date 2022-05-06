import { MATCH_ANY_PARAMETERS, MockedResponses } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'

import {
    BatchSpecWorkspaceResolutionState,
    WorkspaceResolutionStatusResult,
    BatchSpecImportingChangesetsResult,
    PreviewBatchSpecImportingChangesetFields,
    BatchSpecWorkspacesPreviewResult,
    EditBatchChangeFields,
    PreviewVisibleBatchSpecWorkspaceFields,
    BatchSpecState,
    BatchChangeState,
} from '../../../graphql-operations'
import { IMPORTING_CHANGESETS, WORKSPACES, WORKSPACE_RESOLUTION_STATUS } from '../create/backend'

export const MOCK_USER_NAMESPACE = {
    __typename: 'User',
    id: 'user1234',
    username: 'my-username',
    displayName: 'My Display Name',
    namespaceName: 'my-username',
    viewerCanAdminister: true,
    url: '/users/my-username',
} as const

export const mockBatchChange = (batchChange?: Partial<EditBatchChangeFields>): EditBatchChangeFields => ({
    __typename: 'BatchChange',
    id: 'testbc1234',
    url: '/batch-changes/my-username/my-batch-change',
    name: 'my-batch-change',
    namespace: MOCK_USER_NAMESPACE,
    description: 'This is my batch change description.',
    viewerCanAdminister: true,
    currentSpec: mockBatchSpec(),
    batchSpecs: {
        nodes: [mockBatchSpec()],
    },
    state: BatchChangeState.OPEN,
    ...batchChange,
})

export const mockBatchSpec = (
    batchSpec?: Partial<EditBatchChangeFields['currentSpec']>
): EditBatchChangeFields['currentSpec'] => ({
    __typename: 'BatchSpec',
    id: '1',
    originalInput: 'name: my-batch-change',
    createdAt: new Date().toISOString(),
    startedAt: null,
    applyURL: null,
    state: BatchSpecState.PENDING,
    ...batchSpec,
})

export const mockWorkspaceResolutionStatus = (
    status: BatchSpecWorkspaceResolutionState,
    error?: string
): WorkspaceResolutionStatusResult => ({
    node: {
        __typename: 'BatchSpec',
        workspaceResolution: {
            __typename: 'BatchSpecWorkspaceResolution',
            state: status,
            failureMessage: error || null,
        },
    },
})

export const UNSTARTED_RESOLUTION: WorkspaceResolutionStatusResult = {
    node: { __typename: 'BatchSpec', workspaceResolution: null },
}

export const mockWorkspace = (
    id: number,
    fields?: Partial<PreviewVisibleBatchSpecWorkspaceFields>
): PreviewVisibleBatchSpecWorkspaceFields => ({
    __typename: 'VisibleBatchSpecWorkspace',
    id: `id-${id}`,
    path: '/',
    searchResultPaths: ['/first-path'],
    cachedResultFound: false,
    ignored: false,
    unsupported: false,
    ...fields,
    repository: {
        __typename: 'Repository',
        id: `repo-${id}`,
        name: `github.com/my-org/repo-${id}`,
        url: 'superfake.com',
        ...fields?.repository,
    },
    branch: {
        __typename: 'GitRef',
        id: 'main-branch-id',
        displayName: 'main',
        ...fields?.branch,
        target: {
            __typename: 'GitObject',
            oid: 'asdf1234',
            ...fields?.branch?.target,
        },
        url: 'superfake.com',
    },
})

export const mockWorkspaces = (count: number): PreviewVisibleBatchSpecWorkspaceFields[] =>
    [...new Array(count).keys()].map(id => mockWorkspace(id))

const mockImportingChangeset = (
    id: number
): PreviewBatchSpecImportingChangesetFields & { __typename: 'VisibleChangesetSpec' } => ({
    __typename: 'VisibleChangesetSpec',
    id: `changeset-${id}`,
    description: {
        __typename: 'ExistingChangesetReference',
        externalID: `external-changeset-${id}`,
        baseRepository: {
            name: `repo-${id}`,
            url: 'superfake.com',
        },
    },
})

export const mockImportingChangesets = (
    count: number
): (PreviewBatchSpecImportingChangesetFields & {
    __typename: 'VisibleChangesetSpec'
})[] => [...new Array(count).keys()].map(id => mockImportingChangeset(id))

export const mockBatchSpecWorkspaces = (workspacesCount: number): BatchSpecWorkspacesPreviewResult => ({
    node: {
        __typename: 'BatchSpec',
        workspaceResolution: {
            __typename: 'BatchSpecWorkspaceResolution',
            workspaces: {
                __typename: 'BatchSpecWorkspaceConnection',
                totalCount: workspacesCount,
                pageInfo: {
                    hasNextPage: workspacesCount > 0,
                    endCursor: 'end-cursor',
                },
                nodes: mockWorkspaces(workspacesCount),
            },
        },
    },
})

export const mockBatchSpecImportingChangesets = (importsCount: number): BatchSpecImportingChangesetsResult => ({
    node: {
        __typename: 'BatchSpec',
        importingChangesets: {
            __typename: 'ChangesetSpecConnection',
            totalCount: importsCount,
            pageInfo: {
                hasNextPage: importsCount > 0,
                endCursor: 'end-cursor',
            },
            nodes: mockImportingChangesets(importsCount),
        },
    },
})

export const UNSTARTED_CONNECTION_MOCKS: MockedResponses = [
    {
        request: {
            query: getDocumentNode(WORKSPACES),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: mockBatchSpecWorkspaces(0) },
        nMatches: Number.POSITIVE_INFINITY,
    },
    {
        request: {
            query: getDocumentNode(IMPORTING_CHANGESETS),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: mockBatchSpecImportingChangesets(0) },
        nMatches: Number.POSITIVE_INFINITY,
    },
    {
        request: {
            query: getDocumentNode(WORKSPACE_RESOLUTION_STATUS),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: UNSTARTED_RESOLUTION },
        nMatches: Number.POSITIVE_INFINITY,
    },
]

export const UNSTARTED_WITH_CACHE_CONNECTION_MOCKS: MockedResponses = [
    {
        request: {
            query: getDocumentNode(WORKSPACES),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: mockBatchSpecWorkspaces(50) },
        nMatches: Number.POSITIVE_INFINITY,
    },
    {
        request: {
            query: getDocumentNode(IMPORTING_CHANGESETS),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: mockBatchSpecImportingChangesets(20) },
        nMatches: Number.POSITIVE_INFINITY,
    },
    {
        request: {
            query: getDocumentNode(WORKSPACE_RESOLUTION_STATUS),
            variables: MATCH_ANY_PARAMETERS,
        },
        result: { data: mockWorkspaceResolutionStatus(BatchSpecWorkspaceResolutionState.COMPLETED) },
        nMatches: Number.POSITIVE_INFINITY,
    },
]
