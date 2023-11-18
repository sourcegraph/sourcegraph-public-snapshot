import type { Meta, Decorator, StoryFn, StoryObj } from '@storybook/react'
import classNames from 'classnames'
import { subDays } from 'date-fns'

import { isChromatic } from '@sourcegraph/wildcard/src/stories'

import { WebStory } from '../../../components/WebStory'

import { BatchChangeNode } from './BatchChangeNode'
import { nodes, now } from './testData'

import styles from './BatchChangeListPage.module.scss'

const decorator: Decorator = story => (
    <div className={classNames(styles.grid, styles.narrow, 'p-3 container')}>{story()}</div>
)

const config: Meta = {
    title: 'web/batches/list/BatchChangeNode',
    decorators: [decorator],
    argTypes: {
        displayNamespace: {
            name: 'Display namespace',
            control: { type: 'boolean' },
        },
        node: {
            table: {
                disable: true,
            },
        },
    },
    args: {
        displayNamespace: true,
    },
}

export default config

const Template: StoryFn /* <{ node: ListBatchChange }>*/ = ({ node, ...args }) => (
    <WebStory>
        {props => (
            <BatchChangeNode
                {...props}
                node={node}
                displayNamespace={args.displayNamespace}
                now={isChromatic() ? () => subDays(now, 5) : undefined}
                isExecutionEnabled={false}
            />
        )}
    </WebStory>
)

type Story = StoryObj<typeof config>

export const OpenBatchChange: Story = Template.bind({})
OpenBatchChange.args = { node: nodes['Open batch change'] }
OpenBatchChange.storyName = 'Open batch change'

export const FailedDraft: Story = Template.bind({})
FailedDraft.args = { node: nodes['Failed draft'] }
FailedDraft.storyName = 'Failed draft'

export const NoDescription: Story = Template.bind({})
NoDescription.args = { node: nodes['No description'] }
NoDescription.storyName = 'No description'

export const ClosedBatchChange: Story = Template.bind({})
ClosedBatchChange.args = { node: nodes['Closed batch change'] }
ClosedBatchChange.storyName = 'Closed batch change'
