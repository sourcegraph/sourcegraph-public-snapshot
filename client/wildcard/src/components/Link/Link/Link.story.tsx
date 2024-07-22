import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { Text } from '../..'
import { BrandedStory } from '../../../stories/BrandedStory'

import { Link } from './Link'

const decorator: Decorator = story => (
    <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/Link',
    component: Link,

    decorators: [decorator],

    parameters: {
        component: Link,
    },
}

export default config

export const Simple: StoryFn = () => (
    <Text>
        Text can contain links, which <Link to="/">trigger a navigation to a different page</Link>.
    </Text>
)
