import {
    BatchSpecWorkspaceResolutionState,
    WorkspaceResolutionStatusResult,
    PreviewBatchSpecWorkspaceFields,
    WorkspacesAndImportingChangesetsResult,
    PreviewBatchSpecImportingChangesetFields,
} from '../../../../graphql-operations'

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
    path: '/',
    searchResultPaths: ['/first-path'],
    cachedResultFound: false,
    ignored: false,
    unsupported: false,
    ...fields,
    repository: {
        id: `repo-${id}`,
        name: `github.com/my-org/repo-${id}`,
        url: 'superfake.com',
        defaultBranch: {
            id: 'main-branch-id',
            ...fields?.repository?.defaultBranch,
        },
        ...fields?.repository,
    },
    branch: {
        id: 'main-branch-id',
        abbrevName: 'main',
        displayName: 'main',
        ...fields?.branch,
        target: {
            oid: 'asdf1234',
            ...fields?.branch?.target,
        },
        url: 'superfake.com',
    },
})

export const mockWorkspaces = (count: number): PreviewBatchSpecWorkspaceFields[] =>
    [...new Array(count).keys()].map(id => mockWorkspace(id))

const mockImportingChangeset = (id: number): PreviewBatchSpecImportingChangesetFields => ({
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

export const mockImportingChangesets = (count: number): PreviewBatchSpecImportingChangesetFields[] =>
    [...new Array(count).keys()].map(id => mockImportingChangeset(id))

export const mockWorkspacesAndImportingChangesets = (
    workspacesCount: number,
    importsCount: number
): WorkspacesAndImportingChangesetsResult => ({
    node: {
        __typename: 'BatchSpec',
        workspaceResolution: {
            workspaces: {
                nodes: mockWorkspaces(workspacesCount),
            },
        },
        importingChangesets: {
            totalCount: importsCount,
            nodes: mockImportingChangesets(importsCount).map(changeset => ({
                __typename: 'VisibleChangesetSpec',
                ...changeset,
            })),
        },
    },
})
