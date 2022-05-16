import { Meta, Story } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { PopoverTrigger, Typography } from '../..'
import { Button } from '../../Button'

import { FeedbackPrompt } from '.'

import styles from './FeedbackPrompt.module.scss'

const config: Meta = {
    title: 'wildcard/FeedbackPrompt',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],
    parameters: {
        /**
         * Uncomment this once Storybook is upgraded to v6.4.* and the `play` function
         * is used to show the feedback prompt component.
         *
         * https://www.chromatic.com/docs/hoverfocus#javascript-triggered-hover-states
         */
        // chromatic: { disableSnapshot: false },
        component: FeedbackPrompt,
        design: {
            type: 'figma',
            name: 'figma',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=908%3A1',
        },
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

export const FeedbackPromptWithSuccessResponse = () => (
    <>
        <Typography.H1>This is a feedbackPrompt with success response</Typography.H1>
        <FeedbackPrompt onSubmit={handleSuccessSubmit}>
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

export const FeedbackPromptWithErrorResponse: Story = () => (
    <>
        <Typography.H1>This is a feedbackPrompt with error response</Typography.H1>
        <FeedbackPrompt onSubmit={handleErrorSubmit}>
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

export const FeedbackPromptWithInModal: Story = () => (
    <>
        <Typography.H1>This is a feedbackPrompt in modal</Typography.H1>
        <FeedbackPrompt onSubmit={handleSuccessSubmit} modal={true}>
            {({ onClick }) => (
                <Button onClick={onClick} aria-label="Feedback" variant="secondary" outline={true} size="sm">
                    <small>Feedback</small>
                </Button>
            )}
        </FeedbackPrompt>
    </>
)
