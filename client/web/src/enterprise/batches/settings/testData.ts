import { subSeconds } from 'date-fns'

import { BatchSpecListFields, BatchSpecState } from '../../../graphql-operations'

const COMMON_NODE_FIELDS = {
    __typename: 'BatchSpec',
    createdAt: subSeconds(new Date(), 30).toISOString(),
    finishedAt: new Date().toISOString(),
    originalInput: 'name: super-cool-spec',
    description: {
        name: 'super-cool-spec',
    },
    namespace: {
        url: '/users/courier-new',
        namespaceName: 'courier-new',
    },
    creator: {
        username: 'courier-new',
    },
} as const

export const successNode = (id: string): BatchSpecListFields => ({
    ...COMMON_NODE_FIELDS,
    id,
    state: BatchSpecState.COMPLETED,
})

export const NODES: BatchSpecListFields[] = [
    { ...COMMON_NODE_FIELDS, id: 'id1', state: BatchSpecState.QUEUED },
    { ...COMMON_NODE_FIELDS, id: 'id2', state: BatchSpecState.PROCESSING },
    successNode('id3'),
    { ...COMMON_NODE_FIELDS, id: 'id4', state: BatchSpecState.ERRORED },
    { ...COMMON_NODE_FIELDS, id: 'id5', state: BatchSpecState.FAILED },
    { ...COMMON_NODE_FIELDS, id: 'id6', state: BatchSpecState.CANCELING },
    { ...COMMON_NODE_FIELDS, id: 'id7', state: BatchSpecState.CANCELED },
]
