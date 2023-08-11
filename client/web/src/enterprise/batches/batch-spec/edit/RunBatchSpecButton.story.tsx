import { useState } from 'react'

import type { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'
import type { ExecutionOptions } from '../BatchSpecContext'

import { RunBatchSpecButton } from './RunBatchSpecButton'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/batch-spec/edit/RunBatchSpecButton',
    decorators: [decorator],
}

export default config

export const Disabled: Story = () => {
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
                />
            )}
        </WebStory>
    )
}

export const Enabled: Story = () => {
    const [options, setOptions] = useState<ExecutionOptions>({ runWithoutCache: false })
    return (
        <WebStory>
            {props => (
                <RunBatchSpecButton
                    {...props}
                    execute={() => alert('executing!')}
                    options={options}
                    onChangeOptions={setOptions}
                />
            )}
        </WebStory>
    )
}
