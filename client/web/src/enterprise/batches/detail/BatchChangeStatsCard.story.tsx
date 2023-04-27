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

type MockStatsArgs = Omit<ChangesetsStatsFields, 'percentComplete' | '__typename' | 'isCompleted' | 'total'>

// The methods (calculateIsCompleted & calculatePercentComplete) below are implemented on the backend,
// we are duplicating the logic here to make them available for Storybook.
const calculateIsCompleted = <T extends MockStatsArgs>(stats: T, total: number): boolean => {
    if (total === 0) {
        return false
    }
    return stats.closed + stats.merged === total - stats.deleted - stats.archived
}

const calculatePercentComplete = <T extends MockStatsArgs>(stats: T, total: number): number => {
    if (total === 0) {
        return 0
    }
    return ((stats.closed + stats.merged) / (total - stats.deleted - stats.archived)) * 100
}

const calculateTotal = <T extends MockStatsArgs>(stats: T): number =>
    stats.closed + stats.deleted + stats.merged + stats.draft + stats.open + stats.archived + stats.unpublished

export const Draft: Story<MockStatsArgs> = args => {
    const total = calculateTotal(args)
    return (
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
                            failed: args.failed,
                            scheduled: args.scheduled,
                            processing: args.processing,
                            retrying: args.retrying,
                            total,
                            unpublished: args.unpublished,
                            isCompleted: calculateIsCompleted(args, total),
                            percentComplete: calculatePercentComplete(args, total),
                        },
                        diffStat: { added: 0, deleted: 0, __typename: 'DiffStat' },
                        state: BatchChangeState.DRAFT,
                    }}
                />
            )}
        </WebStory>
    )
}

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
    failed: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    retrying: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    scheduled: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    processing: {
        control: { type: 'number' },
        defaultValue: 0,
    },
}

export const Open: Story<MockStatsArgs> = args => {
    const total = calculateTotal(args)
    return (
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
                            failed: args.failed,
                            scheduled: args.scheduled,
                            processing: args.processing,
                            retrying: args.retrying,
                            total,
                            archived: args.archived,
                            unpublished: args.unpublished,
                            isCompleted: calculateIsCompleted(args, total),
                            percentComplete: calculatePercentComplete(args, total),
                        },
                        diffStat: { added: 3000, deleted: 3000, __typename: 'DiffStat' },
                        state: BatchChangeState.OPEN,
                    }}
                />
            )}
        </WebStory>
    )
}

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
    failed: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    retrying: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    scheduled: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    processing: {
        control: { type: 'number' },
        defaultValue: 0,
    },
}

export const OpenAndComplete: Story<MockStatsArgs> = args => {
    const total = calculateTotal(args)
    return (
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
                            failed: args.failed,
                            scheduled: args.scheduled,
                            processing: args.processing,
                            retrying: args.retrying,
                            total,
                            unpublished: args.unpublished,
                            isCompleted: calculateIsCompleted(args, total),
                            percentComplete: calculatePercentComplete(args, total),
                        },
                        diffStat: { added: 3000, deleted: 3000, __typename: 'DiffStat' },
                        state: BatchChangeState.OPEN,
                    }}
                />
            )}
        </WebStory>
    )
}

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
    failed: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    retrying: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    scheduled: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    processing: {
        control: { type: 'number' },
        defaultValue: 0,
    },
}

OpenAndComplete.storyName = 'open and complete'

export const Closed: Story<MockStatsArgs> = args => {
    const total = calculateTotal(args)
    return (
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
                            failed: args.failed,
                            scheduled: args.scheduled,
                            processing: args.processing,
                            retrying: args.retrying,
                            total,
                            unpublished: args.unpublished,
                            isCompleted: calculateIsCompleted(args, total),
                            percentComplete: calculatePercentComplete(args, total),
                        },
                        diffStat: { added: 3000, deleted: 3000, __typename: 'DiffStat' },
                        state: BatchChangeState.CLOSED,
                    }}
                />
            )}
        </WebStory>
    )
}

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
    failed: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    retrying: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    scheduled: {
        control: { type: 'number' },
        defaultValue: 0,
    },
    processing: {
        control: { type: 'number' },
        defaultValue: 0,
    },
}
