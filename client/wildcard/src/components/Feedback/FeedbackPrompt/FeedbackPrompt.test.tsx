import { cleanup, fireEvent, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import sinon from 'sinon'
import { afterEach, beforeEach, describe, expect, test } from 'vitest'

import { assertAriaDisabled, assertAriaEnabled } from '@sourcegraph/testing'

import { renderWithBrandedContext } from '../../../testing'
import { Button } from '../../Button'
import { PopoverTrigger } from '../../Popover'

import { FeedbackPrompt } from '.'

const sampleFeedback = {
    feedback: 'Lorem ipsum dolor sit amet',
}

describe('FeedbackPrompt', () => {
    const onSubmit = sinon.stub()

    beforeEach(() => {
        renderWithBrandedContext(
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

        assertAriaEnabled(screen.getByText('Send'))

        userEvent.click(screen.getByText('Send'))
    }

    test('should render correctly', () => {
        expect(document.body).toMatchSnapshot()
    })

    test('should enable/disable submit button correctly', () => {
        assertAriaDisabled(screen.getByText('Send'))

        userEvent.type(screen.getByLabelText('Send feedback to Sourcegraph'), sampleFeedback.feedback)

        assertAriaEnabled(screen.getByText('Send'))
    })

    test('should render submit success correctly', async () => {
        onSubmit.resolves({ errorMessage: undefined, isHappinessFeedback: true })

        submitFeedback()

        expect(await screen.findByText(/thank you/i)).toBeInTheDocument()
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
