import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { Description } from './Description'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/Description',
    decorators: [decorator],
}

export default config

export const Overview: StoryFn = () => (
    <WebStory>
        {props => (
            <Description
                {...props}
                description="This is an awesome batch change. It will do great things to your codebase."
            />
        )}
    </WebStory>
)
