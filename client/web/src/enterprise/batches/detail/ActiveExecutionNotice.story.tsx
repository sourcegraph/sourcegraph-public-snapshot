import { useMemo } from 'react'

import type { Meta, StoryFn, Decorator } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'
import { BatchSpecState } from '../../../graphql-operations'

import { ActiveExecutionNotice } from './ActiveExecutionNotice'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>
const config: Meta = {
    title: 'web/batches/details',
    decorators: [decorator],
    argTypes: {
        numberActive: {
            name: 'number of specs executing',
            control: { type: 'number' },
            min: 0,
        },
        numberComplete: {
            name: 'number of specs completed',
            control: { type: 'number' },
            min: 0,
        },
    },
    args: {
        numberActive: 1,
        numberComplete: 1,
    },
}

export default config

const PROCESSING_BATCH_SPEC = { state: BatchSpecState.PROCESSING }
const COMPLETE_BATCH_SPEC = { state: BatchSpecState.COMPLETED }

export const ActiveExecutionNoticeStory: StoryFn = args => {
    const numberActive = args.numberActive
    const numberComplete = args.numberComplete

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
