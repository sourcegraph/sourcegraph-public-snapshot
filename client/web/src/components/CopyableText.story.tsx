import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { CopyableText } from './CopyableText'
import { WebStory } from './WebStory'

const decorator: Decorator = story => (
    <div className="container mt-3" style={{ width: 800 }}>
        {story()}
    </div>
)

const config: Meta = {
    title: 'web/CopyableText',
    decorators: [decorator],
}

export default config

export const WithoutSecret: StoryFn = () => <WebStory>{() => <CopyableText text="text that can be copied" />}</WebStory>

WithoutSecret.storyName = 'Without secret'

export const WithSecret: StoryFn = () => <WebStory>{() => <CopyableText secret={true} text="secret text" />}</WebStory>

WithSecret.storyName = 'With secret'
