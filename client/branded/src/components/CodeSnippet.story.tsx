import type { Meta, StoryFn } from '@storybook/react'

import { Container, Text, Code } from '@sourcegraph/wildcard'
import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { CodeSnippet } from './CodeSnippet'

const config: Meta = {
    title: 'branded/CodeSnippet',
    component: CodeSnippet,

    decorators: [story => <BrandedStory>{() => <div className="container mt-3 pb-3">{story()}</div>}</BrandedStory>],
}

export default config

export const Simple: StoryFn = () => (
    <Container>
        <Text>
            Highlighted code pieces should go in a panel separating it from the surrounding content. Use{' '}
            <Code>{'<CodeSnippet />'}</Code> for these uses.
        </Text>
        <CodeSnippet code="property: 1" language="yaml" />
    </Container>
)
