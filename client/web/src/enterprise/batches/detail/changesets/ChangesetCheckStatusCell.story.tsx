import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'
import { ChangesetCheckState } from '../../../../graphql-operations'

import { ChangesetCheckStatusCell } from './ChangesetCheckStatusCell'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/ChangesetCheckStatusCell',
    decorators: [decorator],
}

export default config

export const Pending: Story = () => (
    <WebStory>{props => <ChangesetCheckStatusCell {...props} checkState={ChangesetCheckState.PENDING} />}</WebStory>
)

export const Passed: Story = () => (
    <WebStory>{props => <ChangesetCheckStatusCell {...props} checkState={ChangesetCheckState.PASSED} />}</WebStory>
)

export const Failed: Story = () => (
    <WebStory>{props => <ChangesetCheckStatusCell {...props} checkState={ChangesetCheckState.FAILED} />}</WebStory>
)
