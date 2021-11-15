import { storiesOf } from '@storybook/react'
import { subDays } from 'date-fns'
import React from 'react'

import { WebStory } from '../../../components/WebStory'

import { SupersedingBatchSpecAlert } from './SupersedingBatchSpecAlert'

const { add } = storiesOf('web/batches/details/SupersedingBatchSpecAlert', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('None published', () => (
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
))
