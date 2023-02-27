import { DecoratorFn, Meta, Story } from '@storybook/react'

import { Text } from '../..'
import { BrandedStory } from '../../../stories/BrandedStory'

import { Link } from './Link'

const decorator: DecoratorFn = story => (
    <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/Link',
    component: Link,

    decorators: [decorator],

    parameters: {
        component: Link,
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
    },
}

export default config

export const Simple: Story = () => (
    <Text>
        Text can contain links, which <Link to="/">trigger a navigation to a different page</Link>.
    </Text>
)
