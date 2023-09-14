import type { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'
import { ChangesetCheckState } from '../../../../graphql-operations'

import { ChangesetCheckStatusCell } from './ChangesetCheckStatusCell'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/ChangesetCheckStatusCell',
    decorators: [decorator],
}

export default config

const Template: Story<{ checkState: ChangesetCheckState }> = ({ checkState }) => (
    <WebStory>{props => <ChangesetCheckStatusCell {...props} checkState={checkState} />}</WebStory>
)

export const Pending = Template.bind({})
Pending.args = { checkState: ChangesetCheckState.PENDING }

export const Passed = Template.bind({})
Passed.args = { checkState: ChangesetCheckState.PASSED }

export const Failed = Template.bind({})
Failed.args = { checkState: ChangesetCheckState.FAILED }
