import { select, withKnobs } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'

import { StepStateIcon } from './StepStateIcon'

const decorator: DecoratorFn = story => <div className="p-3">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/execute/workspaces/StepStateIcon',
    decorators: [decorator, withKnobs],
}

export default config

const options = {
    'Cache Found': { cachedResultFound: true },
    Skipped: { skipped: true },
    'Not Started': { startedAt: null },
    Running: { startedAt: 'start-time' },
    Success: { startedAt: 'start-time', finishedAt: 'start-time', exitCode: 0 },
    Failed: { startedAt: 'start-time', finishedAt: 'start-time', exitCode: 1 },
}

export const StepStateIconStory: Story = () => (
    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    <WebStory>{props => <StepStateIcon step={select('State', options, options.Success)} {...props} />}</WebStory>
)

StepStateIconStory.storyName = 'StepStateIcon'
