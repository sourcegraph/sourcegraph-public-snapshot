import { subSeconds } from 'date-fns'

import { BatchSpecExecutionState } from '@sourcegraph/shared/src/graphql/schema'

import { BatchSpecExecutionsFields } from '../../../graphql-operations'

const COMMON_NODE_FIELDS = {
    __typename: 'BatchSpecExecution',
    createdAt: subSeconds(new Date(), 30).toISOString(),
    finishedAt: new Date().toISOString(),
    inputSpec: 'name: super-cool-spec',
    name: 'super-cool-spec',
    namespace: {
        url: '/users/courier-new',
        namespaceName: 'courier-new',
    },
    initiator: {
        username: 'courier-new',
    },
} as const

export const successNode = (id: string): BatchSpecExecutionsFields => ({
    ...COMMON_NODE_FIELDS,
    id,
    state: BatchSpecExecutionState.COMPLETED,
})

export const NODES: BatchSpecExecutionsFields[] = [
    { ...COMMON_NODE_FIELDS, id: 'id1', state: BatchSpecExecutionState.QUEUED },
    { ...COMMON_NODE_FIELDS, id: 'id2', state: BatchSpecExecutionState.PROCESSING },
    successNode('id3'),
    { ...COMMON_NODE_FIELDS, id: 'id4', state: BatchSpecExecutionState.ERRORED },
    { ...COMMON_NODE_FIELDS, id: 'id5', state: BatchSpecExecutionState.FAILED },
    { ...COMMON_NODE_FIELDS, id: 'id6', state: BatchSpecExecutionState.CANCELING },
    { ...COMMON_NODE_FIELDS, id: 'id7', state: BatchSpecExecutionState.CANCELED },
]
