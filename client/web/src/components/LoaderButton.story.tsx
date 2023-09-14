import type { DecoratorFn, Meta, Story } from '@storybook/react'

import { Text } from '@sourcegraph/wildcard'

import { LoaderButton } from './LoaderButton'
import { WebStory } from './WebStory'

const decorator: DecoratorFn = story => (
    <div className="container mt-3" style={{ width: 800 }}>
        {story()}
    </div>
)

const config: Meta = {
    title: 'web/LoaderButton',
    decorators: [decorator],
}

export default config

export const Inline: Story = () => (
    <WebStory>
        {() => (
            <Text>
                <LoaderButton loading={true} label="loader button" variant="primary" />
            </Text>
        )}
    </WebStory>
)

export const Block: Story = () => (
    <WebStory>{() => <LoaderButton loading={true} label="loader button" display="block" variant="primary" />}</WebStory>
)

export const WithLabel: Story = () => (
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
