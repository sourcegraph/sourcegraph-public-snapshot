import { Meta } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Button } from '../../Button'

import { FeedbackPrompt, FeedbackPromptTrigger } from '.'

const handleSubmit = (text?: string, rating?: number) => new Promise<undefined>(resolve => resolve(undefined))

const data = {
    submitHappinessFeedback: {
        alwaysNil: null,
    },
}

const config: Meta = {
    title: 'wildcard/FeedbackPrompt',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],
    parameters: {
        component: FeedbackPrompt,
        design: {
            type: 'figma',
            name: 'figma',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=908%3A1',
        },
    },
}

export default config

export const FeedbackPromptExample = () => (
    <>
        <h1>This is a feedbackPrompt</h1>
        <FeedbackPrompt open={false} onSubmit={handleSubmit} loading={false} data={data}>
            <FeedbackPromptTrigger as={Button} aria-label="Feedback" variant="secondary" outline={true} size="sm">
                <span>Feedback</span>
            </FeedbackPromptTrigger>
        </FeedbackPrompt>
    </>
)
