import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'

import { StepStateIcon } from './StepStateIcon'

const decorator: Decorator = story => <div className="p-3">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/execute/workspaces/StepStateIcon',
    decorators: [decorator],
}

export default config

const options = [
    { label: 'Cache Found', value: { cachedResultFound: true } },
    { label: 'Failed', value: { startedAt: 'start-time', finishedAt: 'start-time', exitCode: 1 } },
    { label: 'Not Started', value: { startedAt: null } },
    { label: 'Running', value: { startedAt: 'start-time' } },
    { label: 'Skipped', value: { skipped: true } },
    { label: 'Success', value: { startedAt: 'start-time', finishedAt: 'start-time', exitCode: 0 } },
]

export const StepStateIconStory: StoryFn = () => (
    <WebStory>
        {props => (
            <>
                {options.map(entry => (
                    <div key={entry.label} className="p-1">
                        {/* eslint-disable-next-line @typescript-eslint/ban-ts-comment */}
                        {/* @ts-ignore */}
                        {entry.label}: <StepStateIcon step={entry.value} {...props} />
                    </div>
                ))}
            </>
        )}
    </WebStory>
)

StepStateIconStory.storyName = 'StepStateIcon'
