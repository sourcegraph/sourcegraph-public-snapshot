import type { Meta, StoryFn, Decorator } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'
import { BatchChangeState, type ChangesetsStatsFields } from '../../../graphql-operations'

import { BatchChangeStatsCard } from './BatchChangeStatsCard'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

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

export const Draft: StoryFn<MockStatsArgs> = args => {
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
    },
    deleted: {
        control: { type: 'number' },
    },
    merged: {
        control: { type: 'number' },
    },
    draft: {
        control: { type: 'number' },
    },
    open: {
        control: { type: 'number' },
    },
    archived: {
        control: { type: 'number' },
    },
    unpublished: {
        control: { type: 'number' },
    },
    failed: {
        control: { type: 'number' },
    },
    retrying: {
        control: { type: 'number' },
    },
    scheduled: {
        control: { type: 'number' },
    },
    processing: {
        control: { type: 'number' },
    },
}
Draft.args = {
    closed: 0,
    deleted: 0,
    merged: 0,
    draft: 0,
    open: 0,
    archived: 0,
    unpublished: 0,
    failed: 0,
    retrying: 0,
    scheduled: 0,
    processing: 0,
}

export const Open: StoryFn<MockStatsArgs> = args => {
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
    },
    deleted: {
        control: { type: 'number' },
    },
    merged: {
        control: { type: 'number' },
    },
    draft: {
        control: { type: 'number' },
    },
    open: {
        control: { type: 'number' },
    },
    archived: {
        control: { type: 'number' },
    },
    unpublished: {
        control: { type: 'number' },
    },
    failed: {
        control: { type: 'number' },
    },
    retrying: {
        control: { type: 'number' },
    },
    scheduled: {
        control: { type: 'number' },
    },
    processing: {
        control: { type: 'number' },
    },
}
Open.args = {
    closed: 10,
    deleted: 10,
    merged: 10,
    draft: 5,
    open: 10,
    archived: 18,
    unpublished: 55,
    failed: 0,
    retrying: 0,
    scheduled: 0,
    processing: 0,
}

export const OpenAndComplete: StoryFn<MockStatsArgs> = args => {
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
    },
    deleted: {
        control: { type: 'number' },
    },
    merged: {
        control: { type: 'number' },
    },
    draft: {
        control: { type: 'number' },
    },
    open: {
        control: { type: 'number' },
    },
    archived: {
        control: { type: 'number' },
    },
    unpublished: {
        control: { type: 'number' },
    },
    failed: {
        control: { type: 'number' },
    },
    retrying: {
        control: { type: 'number' },
    },
    scheduled: {
        control: { type: 'number' },
    },
    processing: {
        control: { type: 'number' },
    },
}
OpenAndComplete.args = {
    closed: 10,
    deleted: 10,
    merged: 80,
    draft: 0,
    open: 0,
    archived: 18,
    unpublished: 0,
    failed: 0,
    retrying: 0,
    scheduled: 0,
    processing: 0,
}

OpenAndComplete.storyName = 'open and complete'

export const Closed: StoryFn<MockStatsArgs> = args => {
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
    },
    deleted: {
        control: { type: 'number' },
    },
    merged: {
        control: { type: 'number' },
    },
    draft: {
        control: { type: 'number' },
    },
    open: {
        control: { type: 'number' },
    },
    archived: {
        control: { type: 'number' },
    },
    unpublished: {
        control: { type: 'number' },
    },
    failed: {
        control: { type: 'number' },
    },
    retrying: {
        control: { type: 'number' },
    },
    scheduled: {
        control: { type: 'number' },
    },
    processing: {
        control: { type: 'number' },
    },
}
Closed.args = {
    closed: 10,
    deleted: 10,
    merged: 10,
    draft: 0,
    open: 10,
    archived: 18,
    unpublished: 60,
    failed: 0,
    retrying: 0,
    scheduled: 0,
    processing: 0,
}
