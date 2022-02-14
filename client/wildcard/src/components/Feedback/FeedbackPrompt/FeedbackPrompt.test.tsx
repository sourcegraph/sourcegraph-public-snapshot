import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import React from 'react'
import sinon from 'sinon'

import { Button } from '../../Button'
import { PopoverTrigger } from '../../Popover'

import { FeedbackPrompt } from '.'

interface SubmitHappinessFeedbackVariables {
    input: {
        score: number
        feedback: string
        currentPath: string
    }
}

const mockData: SubmitHappinessFeedbackVariables = {
    input: {
        score: 4,
        feedback: 'Lorem ipsum dolor sit amet',
        currentPath: '/some-route',
    },
}

describe('FeedbackPrompt', () => {
    const onSubmit = sinon.stub()

    beforeEach(() => {
        render(
            <FeedbackPrompt openByDefault={true} onSubmit={onSubmit}>
                <PopoverTrigger as={Button} aria-label="Feedback" variant="secondary" outline={true} size="sm">
                    <span>Feedback</span>
                </PopoverTrigger>
            </FeedbackPrompt>
        )
    })

    afterEach(async () => {
        await cleanup()
        onSubmit.resetHistory()
    })

    const submitFeedback = () => {
        userEvent.click(screen.getByLabelText('Very Happy'))
        fireEvent.change(screen.getByPlaceholderText('What’s going well? What could be better?'), {
            target: { value: mockData.input.feedback },
        })

        expect(screen.getByText('Send')).toBeEnabled()

        userEvent.click(screen.getByText('Send'))
    }

    test('should render correctly', () => {
        expect(document.body).toMatchSnapshot()
    })

    test('should enable/disable submit button correctly', () => {
        userEvent.click(screen.getByLabelText('Very Happy'))

        expect(screen.getByText('Send')).toBeDisabled()

        userEvent.type(screen.getByPlaceholderText('What’s going well? What could be better?'), mockData.input.feedback)

        expect(screen.getByText('Send')).toBeEnabled()
    })

    test('should render submit success correctly', async () => {
        onSubmit.resolves({ errorMessage: undefined, isHappinessFeedback: true })

        submitFeedback()

        expect(await screen.findByText(/thank you for your help/i)).toBeInTheDocument()
        sinon.assert.calledWith(onSubmit, mockData.input.feedback, mockData.input.score)

        expect(document.body).toMatchSnapshot()
    })

    test('should render submit error correctly', async () => {
        onSubmit.resolves({ errorMessage: 'Something went really wrong', isHappinessFeedback: false })

        submitFeedback()

        expect(await screen.findByText(/something went really wrong/i)).toBeInTheDocument()
        sinon.assert.calledWith(onSubmit, mockData.input.feedback, mockData.input.score)

        expect(document.body).toMatchSnapshot()
    })
})
