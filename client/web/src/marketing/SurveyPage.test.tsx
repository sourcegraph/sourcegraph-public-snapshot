import { MockedProviderProps } from '@apollo/client/testing'
import { cleanup, fireEvent, within, waitFor } from '@testing-library/react'
import { createMemoryHistory } from 'history'
import { Route } from 'react-router'

import { renderWithBrandedContext, RenderWithBrandedContextResult } from '@sourcegraph/shared/src/testing'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { SurveyPage } from './SurveyPage'
import { mockVariables, submitSurveyMock } from './SurveyPage.mocks'

interface RenderSurveyPageParameters {
    mocks: MockedProviderProps['mocks']
    routerProps?: {
        matchParam?: string
        locationState?: {
            score: number
            feedback: string
        }
    }
}

describe('SurveyPage', () => {
    let renderResult: RenderWithBrandedContextResult

    afterEach(cleanup)

    const renderSurveyPage = ({ mocks, routerProps }: RenderSurveyPageParameters) => {
        const history = createMemoryHistory()
        history.push(`/survey/${routerProps?.matchParam || ''}`, routerProps?.locationState)

        return renderWithBrandedContext(
            <MockedTestProvider mocks={mocks}>
                <Route path="/survey/:score?">
                    <SurveyPage authenticatedUser={null} />
                </Route>
            </MockedTestProvider>,
            { route: '/', history }
        )
    }

    describe('Prior to submission', () => {
        beforeEach(() => {
            renderResult = renderSurveyPage({ mocks: [submitSurveyMock] })
        })

        it('renders and handles form fields correctly', async () => {
            const recommendRadioGroup = renderResult.getByLabelText(
                'How likely is it that you would recommend Sourcegraph to a friend?'
            )
            expect(recommendRadioGroup).toBeVisible()
            const score10 = within(recommendRadioGroup).getByLabelText(mockVariables.score)
            fireEvent.click(score10)

            const reasonInput = renderResult.getByLabelText(
                'What is the most important reason for the score you gave Sourcegraph?'
            )
            expect(reasonInput).toBeVisible()
            fireEvent.change(reasonInput, { target: { value: mockVariables.reason } })

            const betterProductInput = renderResult.getByLabelText(
                'What could Sourcegraph do to provide a better product?'
            )
            expect(betterProductInput).toBeVisible()
            fireEvent.change(betterProductInput, { target: { value: mockVariables.better } })

            fireEvent.click(renderResult.getByText('Submit'))

            await waitFor(() => expect(renderResult.history.location.pathname).toBe('/survey/thanks'))
        })
    })

    describe('After submission', () => {
        beforeEach(() => {
            renderResult = renderSurveyPage({
                mocks: [],
                routerProps: { matchParam: 'thanks', locationState: { score: 10, feedback: 'great' } },
            })
        })

        it('renders correct thank you message', () => {
            expect(renderResult.getByText('Thanks for the feedback!')).toBeVisible()
            expect(renderResult.getByText('Tweet feedback', { selector: 'a' })).toBeVisible()
        })
    })
})
