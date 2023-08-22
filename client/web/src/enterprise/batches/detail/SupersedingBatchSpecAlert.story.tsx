import type { Meta, Story, DecoratorFn } from '@storybook/react'
import { subDays } from 'date-fns'

import { WebStory } from '../../../components/WebStory'

import { SupersedingBatchSpecAlert } from './SupersedingBatchSpecAlert'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/details/SupersedingBatchSpecAlert',
    decorators: [decorator],
}

export default config

export const NonePublished: Story = () => (
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
