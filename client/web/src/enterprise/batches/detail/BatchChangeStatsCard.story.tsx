import { Meta, Story, DecoratorFn } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'
import { BatchChangeState, ChangesetsStatsFields } from '../../../graphql-operations'

import { BatchChangeStatsCard } from './BatchChangeStatsCard'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/BatchChangeStatsCard',
    decorators: [decorator],
}

export default config

type MockStatsArgs = Omit<ChangesetsStatsFields, 'percentComplete' | '__typename' | 'isCompleted'>

const calculateIsCompleted = <T extends MockStatsArgs>(stats: T): boolean => {
    if (stats.total === 0) {
        return false
    }
    return stats.closed + stats.merged === stats.total - stats.deleted - stats.archived
}

const calculatePercentComplete = <T extends MockStatsArgs>(stats: T): number => {
    if (stats.total === 0) {
        return 0
    }
    return ((stats.closed + stats.merged) / (stats.total - stats.deleted - stats.archived)) * 100
}

export const Draft: Story<MockStatsArgs> = args => (
    <WebStory>
        {props => (
            <BatchChangeStatsCard
                {...props}
                batchChange={{
                    changesetsStats: {
                        __typename: 'ChangesetsStats',
                        deleted: args.deleted,
                        closed: args.closed,
                        merged: args.merged,
                        draft: args.draft,
                        open: args.open,
                        archived: args.archived,
                        total: args.total,
                        unpublished: args.unpublished,
                        isCompleted: calculateIsCompleted(args),
                        percentComplete: calculatePercentComplete(args),
                    },
                    diffStat: { added: 0, deleted: 0, __typename: 'DiffStat' },
                    state: BatchChangeState.DRAFT,
                }}
            />
        )}
    </WebStory>
)

Draft.argTypes = {
    closed: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    deleted: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    merged: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    draft: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    open: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    archived: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    unpublished: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    total: {
        control: { type: 'number' },
        defaultValue: 0,
    },
}

export const Open: Story<MockStatsArgs> = args => (
    <WebStory>
        {props => (
            <BatchChangeStatsCard
                {...props}
                batchChange={{
                    changesetsStats: {
                        __typename: 'ChangesetsStats',
                        closed: args.closed,
                        deleted: args.deleted,
                        merged: args.merged,
                        draft: args.draft,
                        open: args.open,
                        total: args.total,
                        archived: args.archived,
                        unpublished: args.unpublished,
                        isCompleted: calculateIsCompleted(args),
                        percentComplete: calculatePercentComplete(args),
                    },
                    diffStat: { added: 3000, deleted: 3000, __typename: 'DiffStat' },
                    state: BatchChangeState.OPEN,
                }}
            />
        )}
    </WebStory>
)

Open.argTypes = {
    closed: {
        control: { type: 'number' },
        defaultValue: 10,
    },
    deleted: {
        control: { type: 'number' },
        defaultValue: 10,
    },
    merged: {
        control: { type: 'number' },
        defaultValue: 10,
    },
    draft: {
        control: { type: 'number' },
        defaultValue: 5,
    },
    open: {
        control: { type: 'number' },
        defaultValue: 10,
    },
    archived: {
        control: { type: 'number' },
        defaultValue: 18,
    },
    unpublished: {
        control: { type: 'number' },
        defaultValue: 55,
    },
    total: {
        control: { type: 'number' },
        defaultValue: 118,
    },
}

export const OpenAndComplete: Story<MockStatsArgs> = args => (
    <WebStory>
        {props => (
            <BatchChangeStatsCard
                {...props}
                batchChange={{
                    changesetsStats: {
                        __typename: 'ChangesetsStats',
                        deleted: args.deleted,
                        closed: args.closed,
                        merged: args.merged,
                        draft: args.draft,
                        open: args.open,
                        archived: args.archived,
                        total: args.total,
                        unpublished: args.unpublished,
                        isCompleted: calculateIsCompleted(args),
                        percentComplete: calculatePercentComplete(args),
                    },
                    diffStat: { added: 3000, deleted: 3000, __typename: 'DiffStat' },
                    state: BatchChangeState.OPEN,
                }}
            />
        )}
    </WebStory>
)

OpenAndComplete.argTypes = {
    closed: {
        control: { type: 'number' },
        defaultValue: 10,
    },
    deleted: {
        control: { type: 'number' },
        defaultValue: 10,
    },
    merged: {
        control: { type: 'number' },
        defaultValue: 80,
    },
    draft: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    open: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    archived: {
        control: { type: 'number' },
        defaultValue: 18,
    },
    unpublished: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    total: {
        control: { type: 'number' },
        defaultValue: 118,
    },
}

OpenAndComplete.storyName = 'open and complete'

export const Closed: Story<MockStatsArgs> = args => (
    <WebStory>
        {props => (
            <BatchChangeStatsCard
                {...props}
                batchChange={{
                    changesetsStats: {
                        __typename: 'ChangesetsStats',
                        deleted: args.deleted,
                        closed: args.closed,
                        merged: args.merged,
                        draft: args.draft,
                        open: args.open,
                        archived: args.archived,
                        total: args.total,
                        unpublished: args.unpublished,
                        isCompleted: calculateIsCompleted(args),
                        percentComplete: calculatePercentComplete(args),
                    },
                    diffStat: { added: 3000, deleted: 3000, __typename: 'DiffStat' },
                    state: BatchChangeState.CLOSED,
                }}
            />
        )}
    </WebStory>
)

Closed.argTypes = {
    closed: {
        control: { type: 'number' },
        defaultValue: 10,
    },
    deleted: {
        control: { type: 'number' },
        defaultValue: 10,
    },
    merged: {
        control: { type: 'number' },
        defaultValue: 10,
    },
    draft: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    open: {
        control: { type: 'number' },
        defaultValue: 10,
    },
    archived: {
        control: { type: 'number' },
        defaultValue: 18,
    },
    unpublished: {
        control: { type: 'number' },
        defaultValue: 60,
    },
    total: {
        control: { type: 'number' },
        defaultValue: 118,
    },
}
