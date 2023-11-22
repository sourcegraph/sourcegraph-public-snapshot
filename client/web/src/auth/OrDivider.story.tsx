import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { Card } from '@sourcegraph/wildcard'

import { WebStory } from '../components/WebStory'

import { OrDivider } from './OrDivider'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/auth/OrDivider',
    decorators: [decorator],
}

export default config

export const Alone: StoryFn = () => (
    <WebStory>
        {() => (
            <Card className="border-0">
                <OrDivider />
            </Card>
        )}
    </WebStory>
)
