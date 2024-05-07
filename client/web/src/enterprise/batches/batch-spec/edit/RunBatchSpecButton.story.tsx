import { useState } from 'react'

import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'

import { WebStory } from '../../../../components/WebStory'
import type { ExecutionOptions } from '../BatchSpecContext'

import { RunBatchSpecButton } from './RunBatchSpecButton'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/edit/RunBatchSpecButton',
    decorators: [decorator],
}

export default config

export const Disabled: StoryFn = () => {
    const [options, setOptions] = useState<ExecutionOptions>({ runWithoutCache: false })
    return (
        <WebStory>
            {props => (
                <RunBatchSpecButton
                    {...props}
                    execute={() => alert('executing!')}
                    isExecutionDisabled="There's a problem with your batch spec."
                    options={options}
                    onChangeOptions={setOptions}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            )}
        </WebStory>
    )
}

export const Enabled: StoryFn = () => {
    const [options, setOptions] = useState<ExecutionOptions>({ runWithoutCache: false })
    return (
        <WebStory>
            {props => (
                <RunBatchSpecButton
                    {...props}
                    execute={() => alert('executing!')}
                    options={options}
                    onChangeOptions={setOptions}
                    telemetryRecorder={noOpTelemetryRecorder}
                />
            )}
        </WebStory>
    )
}
