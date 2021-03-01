import React from 'react'
import { render, RenderResult, fireEvent } from '@testing-library/react'
import * as sinon from 'sinon'
import { createMemoryHistory } from 'history'

import { FeedbackPrompt, HAPPINESS_FEEDBACK_OPTIONS } from './FeedbackPrompt'
import { SubmitHappinessFeedbackResult, SubmitHappinessFeedbackVariables } from '../../graphql-operations'
import { MutationResult } from '../../hooks/useMutation'
import { routes } from '../../routes'

let mockResponse: MutationResult<SubmitHappinessFeedbackResult> = { loading: false }
const mockSubmitFn = sinon.spy((parameters: SubmitHappinessFeedbackVariables) => undefined)
const history = createMemoryHistory()

jest.mock('../../hooks', () => ({
    useMutation: () => [mockSubmitFn, mockResponse],
    useRoutesMatch: () => '/some-route',
}))

describe('FeedbackPrompt', () => {
    let queries: RenderResult

    beforeAll(() => {
        window.context = { productResearchPageEnabled: true } as any
    })

    beforeEach(() => {
        queries = render(<FeedbackPrompt history={history} routes={routes} />)
    })

    test('Renders heading correctly', () => {
        expect(queries.getByText('What’s on your mind?')).toBeVisible()
    })

    test('Renders textarea correctly', () => {
        expect(queries.getByPlaceholderText('What’s going well? What could be better?')).toBeVisible()
    })

    test('Renders correct emoji toggles', () => {
        for (const option of HAPPINESS_FEEDBACK_OPTIONS) {
            expect(queries.getByLabelText(option.name)).toBeVisible()
        }
    })

    test('Send button is initially disabled', () => {
        const sendButton = queries.getByText('Send') as HTMLButtonElement
        expect(sendButton.disabled).toBe(true)
    })

    test('Send button is enabled when a happiness rating is selected', () => {
        fireEvent.click(queries.getByLabelText('Very Happy'))
        const sendButton = queries.getByText('Send') as HTMLButtonElement
        expect(sendButton.disabled).toBe(false)
    })

    describe('Submission', () => {
        beforeEach(() => {
            const textArea = queries.getByPlaceholderText('What’s going well? What could be better?')
            const radioButton = queries.getByLabelText('Very Happy')
            const sendButton = queries.getByText('Send')
            fireEvent.change(textArea, { target: { value: 'Lorem ipsum dolor sit amet' } })
            fireEvent.click(radioButton)
            fireEvent.click(sendButton)
        })

        test('Submits data correctly', () => {
            expect(mockSubmitFn.calledOnce).toBe(true)
            sinon.assert.calledWith(mockSubmitFn, {
                input: {
                    score: 4,
                    feedback: 'Lorem ipsum dolor sit amet',
                    currentPath: '/some-route',
                },
            })
        })
    })

    describe('Success', () => {
        beforeAll(() => {
            mockResponse = { loading: false, data: { submitHappinessFeedback: { alwaysNil: null } } }
        })

        test('Renders success page correctly', () => {
            expect(queries.getByText(/Want to help keep making Sourcegraph better?/)).toBeVisible()
        })
    })

    describe('Error', () => {
        const mockError = new Error('Something went really wrong')
        beforeAll(() => {
            mockResponse = { loading: false, error: mockError }
        })

        test('Renders error alert correctly', () => {
            expect(queries.getByText('Error submitting feedback:')).toBeVisible()
            expect(queries.getByText(mockError.message)).toBeVisible()
        })
    })
})
