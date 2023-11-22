import type { Decorator, StoryFn, Meta } from '@storybook/react'
import { subDays } from 'date-fns'

import { WebStory } from '../../../components/WebStory'

import { BatchSpecInfoByline } from './BatchSpecInfoByline'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/preview/BatchSpecInfoByline',
    decorators: [decorator],
}

export default config

export const Default: StoryFn = () => (
    <WebStory>
        {() => (
            <BatchSpecInfoByline
                createdAt={subDays(new Date(), 3).toISOString()}
                creator={{ url: 'http://test.test/alice', username: 'alice' }}
            />
        )}
    </WebStory>
)
