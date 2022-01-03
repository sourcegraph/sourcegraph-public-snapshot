import {
    BatchSpecWorkspaceResolutionState,
    WorkspaceResolutionStatusResult,
    PreviewBatchSpecWorkspaceFields,
    BatchSpecImportingChangesetsResult,
    PreviewBatchSpecImportingChangesetFields,
    BatchSpecWorkspacesPreviewResult,
    EditBatchChangeFields,
} from '../../../../graphql-operations'

export const mockBatchSpec = (): EditBatchChangeFields['currentSpec'] => ({
    __typename: 'BatchSpec',
    id: '1',
    originalInput: '',
    createdAt: 'yesterday',
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

export const mockWorkspace = (
    id: number,
    fields?: Partial<PreviewBatchSpecWorkspaceFields>
): PreviewBatchSpecWorkspaceFields => ({
    __typename: 'BatchSpecWorkspace',
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
        abbrevName: 'main',
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

export const mockWorkspaces = (count: number): PreviewBatchSpecWorkspaceFields[] =>
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
                    hasNextPage: true,
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
                hasNextPage: true,
                endCursor: 'end-cursor',
            },
            nodes: mockImportingChangesets(importsCount),
        },
    },
})
