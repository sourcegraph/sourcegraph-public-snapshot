import { useMemo } from 'react'

import { number } from '@storybook/addon-knobs'
import { Meta, Story, DecoratorFn } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'
import { BatchSpecState } from '../../../graphql-operations'

import { ActiveExecutionNotice } from './ActiveExecutionNotice'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>
const config: Meta = {
    title: 'web/batches/details',
    decorators: [decorator],
}

export default config

const PROCESSING_BATCH_SPEC = { state: BatchSpecState.PROCESSING }
const COMPLETE_BATCH_SPEC = { state: BatchSpecState.COMPLETED }

export const ActiveExecutionNoticeStory: Story = () => {
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
}

ActiveExecutionNoticeStory.storyName = 'ActiveExecutionNotice'
