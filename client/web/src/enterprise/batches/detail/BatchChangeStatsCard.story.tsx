import { Meta, Story, DecoratorFn } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'
import { BatchChangeState } from '../../../graphql-operations'

import { BatchChangeStatsCard } from './BatchChangeStatsCard'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/BatchChangeStatsCard',
    decorators: [decorator],
}

export default config

export const Draft: Story = () => (
    <WebStory>
        {props => (
            <BatchChangeStatsCard
                {...props}
                batchChange={{
                    changesetsStats: {
                        __typename: 'ChangesetsStats',
                        deleted: 0,
                        closed: 0,
                        merged: 0,
                        draft: 0,
                        open: 0,
                        archived: 0,
                        total: 0,
                        unpublished: 0,
                    },
                    diffStat: { added: 0, deleted: 0, __typename: 'DiffStat' },
                    state: BatchChangeState.DRAFT,
                }}
            />
        )}
    </WebStory>
)

export const Open: Story = () => (
    <WebStory>
        {props => (
            <BatchChangeStatsCard
                {...props}
                batchChange={{
                    changesetsStats: {
                        __typename: 'ChangesetsStats',
                        closed: 10,
                        deleted: 10,
                        merged: 10,
                        draft: 5,
                        open: 10,
                        total: 100,
                        archived: 18,
                        unpublished: 55,
                    },
                    diffStat: { added: 3000, deleted: 3000, __typename: 'DiffStat' },
                    state: BatchChangeState.OPEN,
                }}
            />
        )}
    </WebStory>
)

export const OpenAndComplete: Story = () => (
    <WebStory>
        {props => (
            <BatchChangeStatsCard
                {...props}
                batchChange={{
                    changesetsStats: {
                        __typename: 'ChangesetsStats',
                        deleted: 10,
                        closed: 10,
                        merged: 80,
                        draft: 0,
                        open: 0,
                        archived: 18,
                        total: 100,
                        unpublished: 0,
                    },
                    diffStat: { added: 3000, deleted: 3000, __typename: 'DiffStat' },
                    state: BatchChangeState.OPEN,
                }}
            />
        )}
    </WebStory>
)

OpenAndComplete.storyName = 'open and complete'

export const Closed: Story = () => (
    <WebStory>
        {props => (
            <BatchChangeStatsCard
                {...props}
                batchChange={{
                    changesetsStats: {
                        __typename: 'ChangesetsStats',
                        closed: 10,
                        deleted: 10,
                        merged: 10,
                        draft: 0,
                        open: 10,
                        archived: 18,
                        total: 100,
                        unpublished: 60,
                    },
                    diffStat: { added: 3000, deleted: 3000, __typename: 'DiffStat' },
                    state: BatchChangeState.CLOSED,
                }}
            />
        )}
    </WebStory>
)
