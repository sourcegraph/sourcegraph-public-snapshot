import { useState } from 'react'

import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'
import { ExecutionOptions } from '../BatchSpecContext'

import { RunBatchSpecButton } from './RunBatchSpecButton'

const { add } = storiesOf('web/batches/batch-spec/edit/RunBatchSpecButton', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('disabled', () => {
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
})

add('enabled', () => {
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
})
