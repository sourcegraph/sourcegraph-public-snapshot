import type { Decorator, Meta, StoryFn, StoryObj } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'
import { ChangesetCheckState } from '../../../../graphql-operations'

import { ChangesetCheckStatusCell } from './ChangesetCheckStatusCell'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta<typeof ChangesetCheckStatusCell> = {
    title: 'web/batches/ChangesetCheckStatusCell',
    decorators: [decorator],
}

export default config

const Template: StoryFn<{ checkState: ChangesetCheckState }> = ({ checkState }) => (
    <WebStory>{props => <ChangesetCheckStatusCell {...props} checkState={checkState} />}</WebStory>
)

type Story = StoryObj<typeof config>

export const Pending: Story = Template.bind({})
Pending.args = { checkState: ChangesetCheckState.PENDING }

export const Passed: Story = Template.bind({})
Passed.args = { checkState: ChangesetCheckState.PASSED }

export const Failed: Story = Template.bind({})
Failed.args = { checkState: ChangesetCheckState.FAILED }
