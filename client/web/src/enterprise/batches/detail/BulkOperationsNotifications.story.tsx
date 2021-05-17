import { useMemo } from '@storybook/addons'
import { storiesOf } from '@storybook/react'
import React from 'react'

import { BulkOperationState } from '@sourcegraph/shared/src/graphql-operations'

import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { BulkOperationsNotifications } from './BulkOperationsNotifications'

const { add } = storiesOf('web/batches/details/BulkOperationsNotifications', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

add('Processing', () => {
    const bulkOperations = useMemo(
        () => ({ totalCount: 1, nodes: [{ id: '132', state: BulkOperationState.PROCESSING }] }),
        []
    )
    return (
        <EnterpriseWebStory>
            {props => <BulkOperationsNotifications {...props} bulkOperations={bulkOperations} />}
        </EnterpriseWebStory>
    )
})
add('Failed', () => {
    const bulkOperations = useMemo(
        () => ({ totalCount: 1, nodes: [{ id: '132', state: BulkOperationState.FAILED }] }),
        []
    )
    return (
        <EnterpriseWebStory>
            {props => <BulkOperationsNotifications {...props} bulkOperations={bulkOperations} />}
        </EnterpriseWebStory>
    )
})
add('Completed', () => {
    const bulkOperations = useMemo(
        () => ({ totalCount: 1, nodes: [{ id: '132', state: BulkOperationState.COMPLETED }] }),
        []
    )
    return (
        <EnterpriseWebStory>
            {props => <BulkOperationsNotifications {...props} bulkOperations={bulkOperations} />}
        </EnterpriseWebStory>
    )
})
