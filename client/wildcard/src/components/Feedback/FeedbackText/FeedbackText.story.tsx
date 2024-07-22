import type { Meta, StoryFn } from '@storybook/react'

import { H1, Text } from '../..'
import { BrandedStory } from '../../../stories/BrandedStory'

import { FeedbackText } from '.'

const config: Meta = {
    title: 'wildcard/FeedbackText',

    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],
    parameters: {
        component: FeedbackText,
    },
}

export default config

export const FeedbackTextExample: StoryFn = () => (
    <>
        <H1>FeedbackText</H1>
        <Text>This is an example of a feedback with a header</Text>
        <FeedbackText headerText="This is a header text" />
        <Text>This is an example of a feedback with a footer</Text>
        <FeedbackText footerText="This is a footer text" />
    </>
)
