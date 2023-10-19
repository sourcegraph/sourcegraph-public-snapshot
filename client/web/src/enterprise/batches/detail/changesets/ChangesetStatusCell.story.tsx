import type { Meta, Decorator, StoryFn } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'
import { ChangesetState } from '../../../../graphql-operations'

import { ChangesetStatusCell } from './ChangesetStatusCell'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>
const config: Meta = {
    title: 'web/batches/ChangesetStatusCell',
    decorators: [decorator],
}

export default config

const Template: StoryFn<{ state: ChangesetState }> = ({ state }) => (
    <WebStory>{() => <ChangesetStatusCell state={state} className="d-flex text-muted" />}</WebStory>
)

export const Unpublished = Template.bind({})
Unpublished.args = { state: ChangesetState.UNPUBLISHED }

export const Failed = Template.bind({})
Failed.args = { state: ChangesetState.FAILED }

export const Retrying = Template.bind({})
Retrying.args = { state: ChangesetState.RETRYING }

export const Scheduled = Template.bind({})
Scheduled.args = { state: ChangesetState.SCHEDULED }

export const Processing = Template.bind({})
Processing.args = { state: ChangesetState.PROCESSING }

export const Open = Template.bind({})
Open.args = { state: ChangesetState.OPEN }

export const Draft = Template.bind({})
Draft.args = { state: ChangesetState.DRAFT }

export const Closed = Template.bind({})
Closed.args = { state: ChangesetState.CLOSED }

export const Merged = Template.bind({})
Merged.args = { state: ChangesetState.MERGED }

export const Deleted = Template.bind({})
Deleted.args = { state: ChangesetState.DELETED }
