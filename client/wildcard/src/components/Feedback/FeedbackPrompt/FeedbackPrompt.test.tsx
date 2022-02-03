import { ApolloError } from '@apollo/client'
import { render, fireEvent, screen } from '@testing-library/react'
import React from 'react'

import { Button } from '../../Button'

import { FeedbackPrompt, FeedbackPromptTrigger } from '.'

const handleSubmit = (text?: string, rating?: number) => new Promise<undefined>(resolve => resolve(undefined))

describe('FeedbackPrompt', () => {
    test('Renders heading and textarea correctly', () => {
        render(
            <FeedbackPrompt open={true} data={null} loading={false} onSubmit={handleSubmit}>
                <FeedbackPromptTrigger as={Button} aria-label="Feedback" variant="secondary" outline={true} size="sm">
                    Share feedback
                </FeedbackPromptTrigger>
            </FeedbackPrompt>
        )
        expect(screen.getByRole('heading', { level: 3 })).toBeVisible()
        expect(screen.getByPlaceholderText('What’s going well? What could be better?')).toBeVisible()
    })

    test('Renders correct emoji toggles', () => {
        render(
            <FeedbackPrompt open={true} data={null} loading={false} onSubmit={handleSubmit}>
                <FeedbackPromptTrigger as={Button} aria-label="Feedback" variant="secondary" outline={true} size="sm">
                    Share feedback
                </FeedbackPromptTrigger>
            </FeedbackPrompt>
        )
        expect(screen.getByLabelText('Very sad')).toBeVisible()
    })

    test('Send button is initially disabled', () => {
        render(
            <FeedbackPrompt open={true} loading={true} onSubmit={handleSubmit}>
                <FeedbackPromptTrigger as={Button} aria-label="Feedback" variant="secondary" outline={true} size="sm">
                    Share feedback
                </FeedbackPromptTrigger>
            </FeedbackPrompt>
        )
        expect(screen.getByTestId('send-feedback-btn')).toBeDisabled()
    })

    test('Send button is disabled when a happiness rating is selected and textarea is empty', () => {
        render(
            <FeedbackPrompt open={true} data={null} loading={false} onSubmit={handleSubmit}>
                <FeedbackPromptTrigger as={Button} aria-label="Feedback" variant="secondary" outline={true} size="sm">
                    Share feedback
                </FeedbackPromptTrigger>
            </FeedbackPrompt>
        )
        fireEvent.click(screen.getByLabelText('Very Happy'))
        expect(screen.getByTestId('send-feedback-btn')).toBeDisabled()
    })

    test('Send button is not disabled when a textarea is not empty and happiness rating is selected', () => {
        render(
            <FeedbackPrompt open={true} data={null} loading={false} onSubmit={handleSubmit}>
                <FeedbackPromptTrigger as={Button} aria-label="Feedback" variant="secondary" outline={true} size="sm">
                    Share feedback
                </FeedbackPromptTrigger>
            </FeedbackPrompt>
        )
        fireEvent.change(screen.getByPlaceholderText('What’s going well? What could be better?'), {
            target: { value: 'Lorem ipsum dolor sit amet' },
        })
        fireEvent.click(screen.getByLabelText('Very Happy'))
        expect(screen.getByTestId('send-feedback-btn')).toBeEnabled()
    })
})

describe('submission', () => {
    describe('Success', () => {
        test('Renders success page correctly', () => {
            const data = {
                submitHappinessFeedback: {
                    alwaysNil: null,
                },
            }
            render(
                <FeedbackPrompt open={true} data={data} loading={false} onSubmit={handleSubmit}>
                    <FeedbackPromptTrigger
                        as={Button}
                        aria-label="Feedback"
                        variant="secondary"
                        outline={true}
                        size="sm"
                    >
                        Share feedback
                    </FeedbackPromptTrigger>
                </FeedbackPrompt>
            )
            expect(screen.getByText(/want to help keep making sourcegraph better?/i)).toBeVisible()
        })
    })

    describe('Error', () => {
        const mockError = new ApolloError({
            errorMessage: 'Something went really wrong',
        })

        test('Renders error alert correctly', () => {
            render(
                <FeedbackPrompt open={true} data={null} error={mockError} loading={false} onSubmit={handleSubmit}>
                    <FeedbackPromptTrigger
                        as={Button}
                        aria-label="Feedback"
                        variant="secondary"
                        outline={true}
                        size="sm"
                    >
                        Share feedback
                    </FeedbackPromptTrigger>
                </FeedbackPrompt>
            )

            expect(screen.getByText(/error submitting feedback:/i)).toBeVisible()
        })
    })
})
