import { DecoratorFn, Meta, Story } from '@storybook/react'

import { CopyableText } from './CopyableText'
import { WebStory } from './WebStory'

const decorator: DecoratorFn = story => (
    <div className="container mt-3" style={{ width: 800 }}>
        {story()}
    </div>
)

const config: Meta = {
    title: 'web/CopyableText',
    decorators: [decorator],
}

export default config

export const WithoutSecret: Story = () => <WebStory>{() => <CopyableText text="text that can be copied" />}</WebStory>

WithoutSecret.storyName = 'Without secret'

export const WithSecret: Story = () => <WebStory>{() => <CopyableText secret={true} text="secret text" />}</WebStory>

WithSecret.storyName = 'With secret'
