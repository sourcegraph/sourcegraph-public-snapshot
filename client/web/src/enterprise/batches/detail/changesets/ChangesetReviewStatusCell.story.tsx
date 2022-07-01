import { Meta, DecoratorFn, Story } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'
import { ChangesetReviewState } from '../../../../graphql-operations'

import { ChangesetReviewStatusCell } from './ChangesetReviewStatusCell'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/ChangesetReviewStatusCell',
    decorators: [decorator],
}

export default config

export const Approved: Story = () => (
    <WebStory>{props => <ChangesetReviewStatusCell {...props} reviewState={ChangesetReviewState.APPROVED} />}</WebStory>
)

export const ChangesRequested: Story = () => (
    <WebStory>
        {props => <ChangesetReviewStatusCell {...props} reviewState={ChangesetReviewState.CHANGES_REQUESTED} />}
    </WebStory>
)

ChangesRequested.storyName = 'Changes_requested'

export const Pending: Story = () => (
    <WebStory>{props => <ChangesetReviewStatusCell {...props} reviewState={ChangesetReviewState.PENDING} />}</WebStory>
)

export const Commented: Story = () => (
    <WebStory>
        {props => <ChangesetReviewStatusCell {...props} reviewState={ChangesetReviewState.COMMENTED} />}
    </WebStory>
)

export const Dismissed: Story = () => (
    <WebStory>
        {props => <ChangesetReviewStatusCell {...props} reviewState={ChangesetReviewState.DISMISSED} />}
    </WebStory>
)
