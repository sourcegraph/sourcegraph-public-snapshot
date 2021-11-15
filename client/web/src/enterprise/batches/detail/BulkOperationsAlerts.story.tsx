import { useMemo } from '@storybook/addons'
import { storiesOf } from '@storybook/react'
import React from 'react'

import { BulkOperationState } from '@sourcegraph/shared/src/graphql-operations'

import { WebStory } from '../../../components/WebStory'

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
    return <WebStory>{props => <BulkOperationsAlerts {...props} bulkOperations={bulkOperations} />}</WebStory>
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
    return <WebStory>{props => <BulkOperationsAlerts {...props} bulkOperations={bulkOperations} />}</WebStory>
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
    return <WebStory>{props => <BulkOperationsAlerts {...props} bulkOperations={bulkOperations} />}</WebStory>
})
