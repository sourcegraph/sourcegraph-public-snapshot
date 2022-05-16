import { Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Typography } from '../..'

import { FeedbackText } from '.'

const config: Meta = {
    title: 'wildcard/FeedbackText',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],
    parameters: {
        component: FeedbackText,
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
    },
}

export default config

export const FeedbackTextExample: Story = () => (
    <>
        <Typography.H1>FeedbackText</Typography.H1>
        <p>This is an example of a feedback with a header</p>
        <FeedbackText headerText="This is a header text" />
        <p>This is an example of a feedback with a footer</p>
        <FeedbackText footerText="This is a footer text" />
    </>
)
