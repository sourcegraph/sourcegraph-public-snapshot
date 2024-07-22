import React from 'react'

import type { Args, Meta, StoryFn } from '@storybook/react'

import { H1, PopoverTrigger } from '../..'
import { BrandedStory } from '../../../stories/BrandedStory'
import { Button } from '../../Button'

import { FeedbackPrompt } from '.'

import styles from './FeedbackPrompt.module.scss'

const config: Meta = {
    title: 'wildcard/FeedbackPrompt',

    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],
    parameters: {
        component: FeedbackPrompt,
        design: {
            type: 'figma',
            name: 'figma',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=908%3A1',
        },
    },
    argTypes: {
        authenticatedUser: {
            control: { type: 'boolean' },
        },
        productResearchEnabled: {
            control: { type: 'boolean' },
        },
    },
    args: {
        authenticatedUser: true,
        productResearchEnabled: true,
    },
}

export default config

const handleSuccessSubmit = () =>
    Promise.resolve({
        errorMessage: undefined,
        isHappinessFeedback: true,
    })
const handleErrorSubmit = () =>
    Promise.resolve({
        errorMessage: 'Something went really wrong',
        isHappinessFeedback: false,
    })

const commonProps = (
    props: Args
): Pick<
    React.ComponentProps<typeof FeedbackPrompt>,
    'authenticatedUser' | 'openByDefault' | 'productResearchEnabled'
> => ({
    authenticatedUser: props.authenticatedUser ? { username: 'logan', email: 'logan@example.com' } : null,
    openByDefault: true, // to save storybook viewers from needing to click to see the prompt
    productResearchEnabled: props.productResearchEnabled,
})

export const FeedbackPromptWithSuccessResponse: StoryFn = args => (
    <>
        <H1>This is a feedbackPrompt with success response</H1>
        <FeedbackPrompt onSubmit={handleSuccessSubmit} {...commonProps(args)}>
            <PopoverTrigger
                className={styles.feedbackPrompt}
                as={Button}
                aria-label="Feedback"
                variant="secondary"
                outline={true}
                size="sm"
            >
                <span>Feedback</span>
            </PopoverTrigger>
        </FeedbackPrompt>
    </>
)

export const FeedbackPromptWithErrorResponse: StoryFn = args => (
    <>
        <H1>This is a feedbackPrompt with error response</H1>
        <FeedbackPrompt onSubmit={handleErrorSubmit} {...commonProps(args)}>
            <PopoverTrigger
                className={styles.feedbackPrompt}
                as={Button}
                aria-label="Feedback"
                variant="secondary"
                outline={true}
                size="sm"
            >
                <span>Feedback</span>
            </PopoverTrigger>
        </FeedbackPrompt>
    </>
)

export const FeedbackPromptWithInModal: StoryFn = args => (
    <>
        <H1>This is a feedbackPrompt in modal</H1>
        <FeedbackPrompt onSubmit={handleSuccessSubmit} modal={true} {...commonProps(args)}>
            {({ onClick }) => (
                <Button onClick={onClick} aria-label="Feedback" variant="secondary" outline={true} size="sm">
                    <small>Feedback</small>
                </Button>
            )}
        </FeedbackPrompt>
    </>
)
