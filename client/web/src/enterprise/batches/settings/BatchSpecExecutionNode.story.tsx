import { storiesOf } from '@storybook/react'
import classNames from 'classnames'
import { addDays } from 'date-fns'
import { subMinutes } from 'date-fns/esm'
import React from 'react'

import { BatchSpecExecutionState } from '@sourcegraph/shared/src/graphql-operations'

import { BatchSpecExecutionsFields } from '../../../graphql-operations'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { BatchSpecExecutionNode } from './BatchSpecExecutionNode'
import styles from './BatchSpecExecutionsPage.module.scss'

const { add } = storiesOf('web/batches/settings/executions', module).addDecorator(story => (
    <div className={classNames(styles.executionsGrid, 'p-3 container')}>{story()}</div>
))

const COMMON_NODE_FIELDS = {
    __typename: 'BatchSpecExecution',
    createdAt: subMinutes(new Date(), 2).toISOString(),
    finishedAt: new Date().toISOString(),
    inputSpec: 'name: super-cool-spec',
    namespace: {
        url: '/users/courier-new',
        namespaceName: 'courier-new',
    },
    initiator: {
        username: 'courier-new',
    },
} as const

const NOW = () => addDays(new Date(), 1)

const NODES: BatchSpecExecutionsFields[] = [
    { ...COMMON_NODE_FIELDS, id: 'id1', state: BatchSpecExecutionState.QUEUED },
    { ...COMMON_NODE_FIELDS, id: 'id2', state: BatchSpecExecutionState.PROCESSING },
    { ...COMMON_NODE_FIELDS, id: 'id3', state: BatchSpecExecutionState.COMPLETED },
    { ...COMMON_NODE_FIELDS, id: 'id4', state: BatchSpecExecutionState.ERRORED },
    { ...COMMON_NODE_FIELDS, id: 'id5', state: BatchSpecExecutionState.FAILED },
    { ...COMMON_NODE_FIELDS, id: 'id6', state: BatchSpecExecutionState.CANCELING },
    { ...COMMON_NODE_FIELDS, id: 'id7', state: BatchSpecExecutionState.CANCELED },
]

add('BatchSpecExecutionNode', () => (
    <EnterpriseWebStory>
        {props => (
            <>
                {NODES.map(node => (
                    <BatchSpecExecutionNode {...props} key={node.id} node={node} now={NOW} />
                ))}
            </>
        )}
    </EnterpriseWebStory>
))
