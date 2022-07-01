import { boolean } from '@storybook/addon-knobs'
import { Meta, DecoratorFn, Story } from '@storybook/react'
import classNames from 'classnames'
import { subDays } from 'date-fns'

import { isChromatic } from '@sourcegraph/storybook'

import { WebStory } from '../../../components/WebStory'
import { BatchChangeState, BatchSpecState, ListBatchChange } from '../../../graphql-operations'

import { BatchChangeNode } from './BatchChangeNode'
import { now } from './testData'

import styles from './BatchChangeListPage.module.scss'

const decorator: DecoratorFn = story => (
    <div className={classNames(styles.grid, styles.narrow, 'p-3 container')}>{story()}</div>
)

const config: Meta = {
    title: 'web/batches/list/BatchChangeNode',
    decorators: [decorator],
}

export default config

const nodes: Record<string, { __typename: 'BatchChange' } & ListBatchChange> = {
    OpenBatchChange: {
        __typename: 'BatchChange',
        id: 'test',
        url: '/users/alice/batch-change/test',
        name: 'Awesome batch',
        state: BatchChangeState.OPEN,
        description: `# What this does

This is my thorough explanation. And it can also get very long, in that case the UI doesn't break though, which is good. And one more line to finally be longer than the viewport.`,
        createdAt: subDays(now, 5).toISOString(),
        closedAt: null,
        changesetsStats: {
            open: 10,
            closed: 0,
            merged: 5,
        },
        namespace: {
            namespaceName: 'alice',
            url: '/users/alice',
        },
        currentSpec: {
            id: 'old-spec-1',
        },
        batchSpecs: {
            nodes: [
                {
                    __typename: 'BatchSpec',
                    id: 'test-1',
                    state: BatchSpecState.PROCESSING,
                    applyURL: null,
                },
            ],
        },
    },
    FailedDraft: {
        __typename: 'BatchChange',
        id: 'testdraft',
        url: '/users/alice/batch-change/test',
        name: 'Awesome batch',
        state: BatchChangeState.DRAFT,
        description: 'The execution of the batch spec failed.',
        createdAt: subDays(now, 5).toISOString(),
        closedAt: null,
        changesetsStats: {
            open: 10,
            closed: 0,
            merged: 5,
        },
        namespace: {
            namespaceName: 'alice',
            url: '/users/alice',
        },
        currentSpec: {
            id: 'empty-draft-2',
        },
        batchSpecs: {
            nodes: [
                {
                    __typename: 'BatchSpec',
                    id: 'test-2',
                    state: BatchSpecState.FAILED,
                    applyURL: null,
                },
            ],
        },
    },
    NoDescription: {
        __typename: 'BatchChange',
        id: 'test2',
        url: '/users/alice/batch-changes/test2',
        name: 'Awesome batch',
        state: BatchChangeState.OPEN,
        description: null,
        createdAt: subDays(now, 5).toISOString(),
        closedAt: null,
        changesetsStats: {
            open: 10,
            closed: 0,
            merged: 5,
        },
        namespace: {
            namespaceName: 'alice',
            url: '/users/alice',
        },
        currentSpec: {
            id: 'old-spec-3',
        },
        batchSpecs: {
            nodes: [
                {
                    __typename: 'BatchSpec',
                    id: 'test-3',
                    state: BatchSpecState.COMPLETED,
                    applyURL: '/fake-apply-url',
                },
            ],
        },
    },
    ClosedBatchChange: {
        __typename: 'BatchChange',
        id: 'test3',
        url: '/users/alice/batch-changes/test3',
        name: 'Awesome batch',
        state: BatchChangeState.CLOSED,
        description: `# My batch

        This is my thorough explanation.`,
        createdAt: subDays(now, 5).toISOString(),
        closedAt: subDays(now, 3).toISOString(),
        changesetsStats: {
            open: 0,
            closed: 10,
            merged: 5,
        },
        namespace: {
            namespaceName: 'alice',
            url: '/users/alice',
        },
        currentSpec: {
            id: 'test-4',
        },
        batchSpecs: {
            nodes: [
                {
                    __typename: 'BatchSpec',
                    id: 'test-4',
                    state: BatchSpecState.COMPLETED,
                    applyURL: '/fake-apply-url',
                },
            ],
        },
    },
}

export const OpenBatchChange: Story = () => (
    <WebStory>
        {props => (
            <BatchChangeNode
                {...props}
                node={nodes.OpenBatchChange}
                displayNamespace={boolean('Display namespace', true)}
                now={isChromatic() ? () => subDays(now, 5) : undefined}
                isExecutionEnabled={false}
            />
        )}
    </WebStory>
)

OpenBatchChange.storyName = 'Open batch change'

export const FailedDraft: Story = () => (
    <WebStory>
        {props => (
            <BatchChangeNode
                {...props}
                node={nodes.FailedDraft}
                displayNamespace={boolean('Display namespace', true)}
                now={isChromatic() ? () => subDays(now, 5) : undefined}
                isExecutionEnabled={false}
            />
        )}
    </WebStory>
)

FailedDraft.storyName = 'Failed draft'

export const NoDescription: Story = () => (
    <WebStory>
        {props => (
            <BatchChangeNode
                {...props}
                node={nodes.NoDescription}
                displayNamespace={boolean('Display namespace', true)}
                now={isChromatic() ? () => subDays(now, 5) : undefined}
                isExecutionEnabled={false}
            />
        )}
    </WebStory>
)

NoDescription.storyName = 'No description'

export const ClosedBatchChange: Story = () => (
    <WebStory>
        {props => (
            <BatchChangeNode
                {...props}
                node={nodes.ClosedBatchChange}
                displayNamespace={boolean('Display namespace', true)}
                now={isChromatic() ? () => subDays(now, 5) : undefined}
                isExecutionEnabled={false}
            />
        )}
    </WebStory>
)

ClosedBatchChange.storyName = 'Closed batch change'
