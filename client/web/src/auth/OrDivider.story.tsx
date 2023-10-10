import type { DecoratorFn, Meta, Story } from '@storybook/react'

import { Card } from '@sourcegraph/wildcard'

import { WebStory } from '../components/WebStory'

import { OrDivider } from './OrDivider'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/auth/OrDivider',
    decorators: [decorator],
}

export default config

export const Alone: Story = () => (
    <WebStory>
        {() => (
            <Card className="border-0">
                <OrDivider />
            </Card>
        )}
    </WebStory>
)
