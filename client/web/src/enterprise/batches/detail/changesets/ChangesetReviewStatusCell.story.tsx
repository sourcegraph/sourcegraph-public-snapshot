import type { Meta, Decorator, StoryFn, StoryObj } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'
import { ChangesetReviewState } from '../../../../graphql-operations'

import { ChangesetReviewStatusCell } from './ChangesetReviewStatusCell'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta<typeof ChangesetReviewStatusCell> = {
    title: 'web/batches/ChangesetReviewStatusCell',
    decorators: [decorator],
}

export default config

const Template: StoryFn<{ reviewState: ChangesetReviewState }> = ({ reviewState }) => (
    <WebStory>{props => <ChangesetReviewStatusCell {...props} reviewState={reviewState} />}</WebStory>
)

type Story = StoryObj<typeof config>

export const Approved: Story = Template.bind({})
Approved.args = { reviewState: ChangesetReviewState.APPROVED }

export const ChangesRequested: Story = Template.bind({})
ChangesRequested.args = { reviewState: ChangesetReviewState.CHANGES_REQUESTED }
ChangesRequested.storyName = 'Changes_requested'

export const Commented: Story = Template.bind({})
Commented.args = { reviewState: ChangesetReviewState.COMMENTED }

export const Pending: Story = Template.bind({})
Pending.args = { reviewState: ChangesetReviewState.PENDING }

export const Dismissed: Story = Template.bind({})
Dismissed.args = { reviewState: ChangesetReviewState.DISMISSED }
