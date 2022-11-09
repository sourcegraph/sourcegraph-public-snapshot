import { cleanup, fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import sinon from 'sinon'

import { Button } from '../../Button'
import { PopoverTrigger } from '../../Popover'

import { FeedbackPrompt } from '.'

const sampleFeedback = {
    feedback: 'Lorem ipsum dolor sit amet',
}

describe('FeedbackPrompt', () => {
    const onSubmit = sinon.stub()

    beforeEach(() => {
        render(
            <FeedbackPrompt
                openByDefault={true}
                onSubmit={onSubmit}
                productResearchEnabled={true}
                authenticatedUser={null}
            >
                <PopoverTrigger as={Button} aria-label="Feedback" variant="secondary" outline={true} size="sm">
                    <span>Feedback</span>
                </PopoverTrigger>
            </FeedbackPrompt>
        )
    })

    afterEach(() => {
        cleanup()
        onSubmit.resetHistory()
    })

    const submitFeedback = () => {
        fireEvent.change(screen.getByLabelText('Send feedback to Sourcegraph'), {
            target: { value: sampleFeedback.feedback },
        })

        expect(screen.getByText('Send')).toBeEnabled()

        userEvent.click(screen.getByText('Send'))
    }

    test('should render correctly', () => {
        expect(document.body).toMatchSnapshot()
    })

    test('should enable/disable submit button correctly', () => {
        expect(screen.getByText('Send')).toBeDisabled()

        userEvent.type(screen.getByLabelText('Send feedback to Sourcegraph'), sampleFeedback.feedback)

        expect(screen.getByText('Send')).toBeEnabled()
    })

    test('should render submit success correctly', async () => {
        onSubmit.resolves({ errorMessage: undefined, isHappinessFeedback: true })

        submitFeedback()

        expect(await screen.findByText(/thank you for your help/i)).toBeInTheDocument()
        sinon.assert.calledWith(onSubmit, sampleFeedback.feedback)

        expect(document.body).toMatchSnapshot()
    })

    test('should render submit error correctly', async () => {
        onSubmit.resolves({ errorMessage: 'Something went really wrong', isHappinessFeedback: false })

        submitFeedback()

        expect(await screen.findByText(/something went really wrong/i)).toBeInTheDocument()
        sinon.assert.calledWith(onSubmit, sampleFeedback.feedback)

        expect(document.body).toMatchSnapshot()
    })
})
