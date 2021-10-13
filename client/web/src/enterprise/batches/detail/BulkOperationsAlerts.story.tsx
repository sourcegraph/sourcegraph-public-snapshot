import { useMemo } from '@storybook/addons'
import { storiesOf } from '@storybook/react'
import React from 'react'

import { BulkOperationState } from '@sourcegraph/shared/src/graphql-operations'

import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { BulkOperationsAlerts } from './BulkOperationsAlerts'

const { add } = storiesOf('web/batches/details/BulkOperationsAlerts', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('Processing', () => {
    const bulkOperations = useMemo(
        () => ({
            __typename: 'BulkOperationConnection' as const,
            totalCount: 1,
            nodes: [{ id: '132', state: BulkOperationState.PROCESSING, __typename: 'BulkOperation' as const }],
        }),
        []
    )
    return (
        <EnterpriseWebStory>
            {props => <BulkOperationsAlerts {...props} bulkOperations={bulkOperations} />}
        </EnterpriseWebStory>
    )
})
add('Failed', () => {
    const bulkOperations = useMemo(
        () => ({
            __typename: 'BulkOperationConnection' as const,
            totalCount: 1,
            nodes: [{ id: '132', state: BulkOperationState.FAILED, __typename: 'BulkOperation' as const }],
        }),
        []
    )
    return (
        <EnterpriseWebStory>
            {props => <BulkOperationsAlerts {...props} bulkOperations={bulkOperations} />}
        </EnterpriseWebStory>
    )
})
add('Completed', () => {
    const bulkOperations = useMemo(
        () => ({
            __typename: 'BulkOperationConnection' as const,
            totalCount: 1,
            nodes: [{ id: '132', state: BulkOperationState.COMPLETED, __typename: 'BulkOperation' as const }],
        }),
        []
    )
    return (
        <EnterpriseWebStory>
            {props => <BulkOperationsAlerts {...props} bulkOperations={bulkOperations} />}
        </EnterpriseWebStory>
    )
})
