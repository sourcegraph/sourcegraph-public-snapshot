import { useMemo } from 'react'

import { number } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'
import { BatchSpecState } from '../../../graphql-operations'

import { ActiveExecutionNotice } from './ActiveExecutionNotice'

const { add } = storiesOf('web/batches/details', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const PROCESSING_BATCH_SPEC = { state: BatchSpecState.PROCESSING }
const COMPLETE_BATCH_SPEC = { state: BatchSpecState.COMPLETED }

add('ActiveExecutionNotice', () => {
    const numberActive = number('number of specs executing', 1, { min: 0 })
    const numberComplete = number('number of specs complete', 1, { min: 0 })

    const specs = useMemo(
        () => [
            ...new Array(numberActive).fill(0).map(() => PROCESSING_BATCH_SPEC),
            ...new Array(numberComplete).fill(0).map(() => COMPLETE_BATCH_SPEC),
        ],
        [numberActive, numberComplete]
    )

    return <WebStory>{() => <ActiveExecutionNotice batchSpecs={specs} batchChangeURL="lol.fake" />}</WebStory>
})
