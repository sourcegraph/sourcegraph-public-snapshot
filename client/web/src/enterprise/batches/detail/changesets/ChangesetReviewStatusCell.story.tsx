import type { Meta, DecoratorFn, Story } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'
import { ChangesetReviewState } from '../../../../graphql-operations'

import { ChangesetReviewStatusCell } from './ChangesetReviewStatusCell'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/ChangesetReviewStatusCell',
    decorators: [decorator],
}

export default config

const Template: Story<{ reviewState: ChangesetReviewState }> = ({ reviewState }) => (
    <WebStory>{props => <ChangesetReviewStatusCell {...props} reviewState={reviewState} />}</WebStory>
)

export const Approved = Template.bind({})
Approved.args = { reviewState: ChangesetReviewState.APPROVED }

export const ChangesRequested = Template.bind({})
ChangesRequested.args = { reviewState: ChangesetReviewState.CHANGES_REQUESTED }
ChangesRequested.storyName = 'Changes_requested'

export const Commented = Template.bind({})
Commented.args = { reviewState: ChangesetReviewState.COMMENTED }

export const Pending = Template.bind({})
Pending.args = { reviewState: ChangesetReviewState.PENDING }

export const Dismissed = Template.bind({})
Dismissed.args = { reviewState: ChangesetReviewState.DISMISSED }
