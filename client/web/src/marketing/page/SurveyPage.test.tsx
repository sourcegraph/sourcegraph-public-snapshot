import type { MockedProviderProps } from '@apollo/client/testing'
import { cleanup, fireEvent, within, waitFor } from '@testing-library/react'

import { TelemetryV2Props, noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { type RenderWithBrandedContextResult, renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { SurveyPage } from './SurveyPage'
import { mockVariables, submitSurveyMock } from './SurveyPage.mocks'

interface RenderSurveyPageParameters extends TelemetryV2Props {
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

    const renderSurveyPage = ({ mocks, routerProps, telemetryRecorder }: RenderSurveyPageParameters) =>
        renderWithBrandedContext(
            <MockedTestProvider mocks={mocks}>
                <SurveyPage authenticatedUser={null} telemetryRecorder={telemetryRecorder} />
            </MockedTestProvider>,
            {
                path: '/survey/:score?',
                route: {
                    pathname: `/survey/${routerProps?.matchParam || ''}`,
                    state: routerProps?.locationState,
                },
            }
        )

    describe('Prior to submission', () => {
        beforeEach(() => {
            renderResult = renderSurveyPage({ mocks: [submitSurveyMock], telemetryRecorder: noOpTelemetryRecorder })
        })

        it('renders and handles form fields correctly', async () => {
            const recommendRadioGroup = renderResult.getByLabelText(
                'How likely is it that you would recommend Sourcegraph to a friend?'
            )
            expect(recommendRadioGroup).toBeVisible()
            const score10 = within(recommendRadioGroup).getByLabelText(mockVariables.score)
            fireEvent.click(score10)

            const otherUseCaseInput = renderResult.getByLabelText('What do you use Sourcegraph for?')
            expect(otherUseCaseInput).toBeVisible()
            fireEvent.change(otherUseCaseInput, { target: { value: mockVariables.otherUseCase } })

            const reasonInput = renderResult.getByLabelText('How can we make Sourcegraph better?')
            expect(reasonInput).toBeVisible()
            fireEvent.change(reasonInput, { target: { value: mockVariables.better } })

            fireEvent.click(renderResult.getByText('Submit'))

            await waitFor(() => expect(renderResult.locationRef.current?.pathname).toBe('/survey/thanks'))
        })
    })

    describe('After submission', () => {
        beforeEach(() => {
            renderResult = renderSurveyPage({
                mocks: [],
                routerProps: { matchParam: 'thanks', locationState: { score: 10, feedback: 'great' } },
                telemetryRecorder: noOpTelemetryRecorder,
            })
        })

        it('renders correct thank you message', () => {
            expect(renderResult.getByText('Thanks for the feedback!')).toBeVisible()
            expect(renderResult.getByText('Tweet feedback', { selector: 'a' })).toBeVisible()
        })
    })
})
