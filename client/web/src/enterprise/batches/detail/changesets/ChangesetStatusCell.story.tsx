import type { Meta, Decorator, StoryFn, StoryObj } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'
import { ChangesetState } from '../../../../graphql-operations'

import { ChangesetStatusCell } from './ChangesetStatusCell'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>
const config: Meta<typeof ChangesetStatusCell> = {
    title: 'web/batches/ChangesetStatusCell',
    decorators: [decorator],
}

export default config

const Template: StoryFn<{ state: ChangesetState }> = ({ state }) => (
    <WebStory>{() => <ChangesetStatusCell state={state} className="d-flex text-muted" />}</WebStory>
)

type Story = StoryObj<typeof config>

export const Unpublished: Story = Template.bind({})
Unpublished.args = { state: ChangesetState.UNPUBLISHED }

export const Failed: Story = Template.bind({})
Failed.args = { state: ChangesetState.FAILED }

export const Retrying: Story = Template.bind({})
Retrying.args = { state: ChangesetState.RETRYING }

export const Scheduled: Story = Template.bind({})
Scheduled.args = { state: ChangesetState.SCHEDULED }

export const Processing: Story = Template.bind({})
Processing.args = { state: ChangesetState.PROCESSING }

export const Open: Story = Template.bind({})
Open.args = { state: ChangesetState.OPEN }

export const Draft: Story = Template.bind({})
Draft.args = { state: ChangesetState.DRAFT }

export const Closed: Story = Template.bind({})
Closed.args = { state: ChangesetState.CLOSED }

export const Merged: Story = Template.bind({})
Merged.args = { state: ChangesetState.MERGED }

export const Deleted: Story = Template.bind({})
Deleted.args = { state: ChangesetState.DELETED }
