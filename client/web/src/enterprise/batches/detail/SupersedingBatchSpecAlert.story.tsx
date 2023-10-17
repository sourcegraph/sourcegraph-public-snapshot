import type { Meta, StoryFn, Decorator } from '@storybook/react'
import { subDays } from 'date-fns'

import { WebStory } from '../../../components/WebStory'

import { SupersedingBatchSpecAlert } from './SupersedingBatchSpecAlert'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/details/SupersedingBatchSpecAlert',
    decorators: [decorator],
}

export default config

export const NonePublished: StoryFn = () => (
    <WebStory>
        {() => (
            <SupersedingBatchSpecAlert
                spec={{
                    applyURL: '/users/alice/batch-changes/preview/123456SAMPLEID',
                    createdAt: subDays(new Date(), 1).toISOString(),
                }}
            />
        )}
    </WebStory>
)

NonePublished.storyName = 'None published'
