import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { Text } from '@sourcegraph/wildcard'

import { LoaderButton } from './LoaderButton'
import { WebStory } from './WebStory'

const decorator: Decorator = story => (
    <div className="container mt-3" style={{ width: 800 }}>
        {story()}
    </div>
)

const config: Meta = {
    title: 'web/LoaderButton',
    decorators: [decorator],
}

export default config

export const Inline: StoryFn = () => (
    <WebStory>
        {() => (
            <Text>
                <LoaderButton loading={true} label="loader button" variant="primary" />
            </Text>
        )}
    </WebStory>
)

export const Block: StoryFn = () => (
    <WebStory>{() => <LoaderButton loading={true} label="loader button" display="block" variant="primary" />}</WebStory>
)

export const WithLabel: StoryFn = () => (
    <WebStory>
        {() => (
            <LoaderButton
                alwaysShowLabel={true}
                loading={true}
                label="loader button"
                display="block"
                variant="primary"
            />
        )}
    </WebStory>
)
