import {
    BatchSpecWorkspaceListFields,
    EditBatchChangeFields,
    BatchSpecWorkspaceState,
} from '../../../../graphql-operations'

export const mockBatchSpec = (): EditBatchChangeFields['currentSpec'] => ({
    __typename: 'BatchSpec',
    id: '1',
    originalInput: '',
    createdAt: 'yesterday',
})

export const mockWorkspace = (
    id: number,
    fields?: Partial<BatchSpecWorkspaceListFields>
): BatchSpecWorkspaceListFields => ({
    __typename: 'BatchSpecWorkspace',
    id: `spec-${id}`,
    state: BatchSpecWorkspaceState.PROCESSING,
    placeInQueue: id,
    path: '/some/path/to/workspace',
    cachedResultFound: false,
    ignored: false,
    unsupported: false,
    ...fields,
    repository: {
        __typename: 'Repository',
        name: `github.com/my-org/repo-${id}`,
        url: 'superfake.com',
        ...fields?.repository,
    },
    branch: {
        __typename: 'GitRef',
        displayName: 'main',
        ...fields?.branch,
    },
    diffStat: {
        __typename: 'DiffStat',
        added: 10,
        changed: 20,
        deleted: 5,
        ...fields?.diffStat,
    },
})

export const mockWorkspaces = (count: number): BatchSpecWorkspaceListFields[] =>
    [...new Array(count).keys()].map(id => mockWorkspace(id))
