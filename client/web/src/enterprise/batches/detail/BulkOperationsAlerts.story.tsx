import { useMemo } from '@storybook/addons'
import type { Meta, Story, DecoratorFn } from '@storybook/react'

import { BulkOperationState } from '@sourcegraph/shared/src/graphql-operations'

import { WebStory } from '../../../components/WebStory'

import { BulkOperationsAlerts } from './BulkOperationsAlerts'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/details/BulkOperationsAlerts',
    decorators: [decorator],
}

export default config

export const Processing: Story = () => {
    const bulkOperations = useMemo(
        () => ({
            __typename: 'BulkOperationConnection' as const,
            totalCount: 1,
            nodes: [{ id: '132', state: BulkOperationState.PROCESSING, __typename: 'BulkOperation' as const }],
        }),
        []
    )
    return <WebStory>{props => <BulkOperationsAlerts {...props} bulkOperations={bulkOperations} />}</WebStory>
}

export const Failed: Story = () => {
    const bulkOperations = useMemo(
        () => ({
            __typename: 'BulkOperationConnection' as const,
            totalCount: 1,
            nodes: [{ id: '132', state: BulkOperationState.FAILED, __typename: 'BulkOperation' as const }],
        }),
        []
    )
    return <WebStory>{props => <BulkOperationsAlerts {...props} bulkOperations={bulkOperations} />}</WebStory>
}

export const Completed: Story = () => {
    const bulkOperations = useMemo(
        () => ({
            __typename: 'BulkOperationConnection' as const,
            totalCount: 1,
            nodes: [{ id: '132', state: BulkOperationState.COMPLETED, __typename: 'BulkOperation' as const }],
        }),
        []
    )
    return <WebStory>{props => <BulkOperationsAlerts {...props} bulkOperations={bulkOperations} />}</WebStory>
}
