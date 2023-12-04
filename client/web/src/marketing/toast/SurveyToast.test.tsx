import { gql } from '@apollo/client'
import { createMockClient } from '@apollo/client/testing'
import { cleanup, fireEvent, waitFor } from '@testing-library/react'
import { take } from 'rxjs/operators'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'

import type { TemporarySettings } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'
import { TemporarySettingsContext } from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsProvider'
import {
    InMemoryMockSettingsBackend,
    TemporarySettingsStorage,
} from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsStorage'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { type RenderWithBrandedContextResult, renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import type { AuthenticatedUser } from '../../auth'
import { mockVariables, submitSurveyMock } from '../page/SurveyPage.mocks'

import { SurveyToast } from '.'

export const mockAuthenticatedUser: AuthenticatedUser = {
    id: 'userID',
    username: 'username',
    emails: [{ email: 'user@me.com', isPrimary: true, verified: true }],
} as AuthenticatedUser

describe('SurveyToast', () => {
    let renderResult: RenderWithBrandedContextResult

    afterEach(() => {
        localStorage.clear()
        cleanup()
    })

    const mockClient = createMockClient(
        { temporarySettings: { contents: JSON.stringify({}) } },
        gql`
            query GetTemporarySettings {
                temporarySettings {
                    contents
                }
            }
        `
    )

    const settingsStorage = new TemporarySettingsStorage(mockClient, true)

    const renderwithTemporarySettings = (settings: TemporarySettings) => {
        settingsStorage.setSettingsBackend(new InMemoryMockSettingsBackend(settings))

        return renderWithBrandedContext(
            <MockedTestProvider mocks={[submitSurveyMock]}>
                <TemporarySettingsContext.Provider value={settingsStorage}>
                    <SurveyToast authenticatedUser={mockAuthenticatedUser} />
                </TemporarySettingsContext.Provider>
            </MockedTestProvider>
        )
    }

    const getTemporarySetting = (key: keyof TemporarySettings) =>
        new Promise(resolve => settingsStorage.get(key).pipe(take(1)).subscribe({ next: resolve }))

    describe('toast has not been dismissed by the user', () => {
        describe('before day 3', () => {
            beforeEach(() => {
                renderResult = renderwithTemporarySettings({ 'user.daysActiveCount': 1 })
            })

            it('the user is not surveyed', () => {
                expect(renderResult.container).toBeEmptyDOMElement()
            })
        })

        describe('on day 3', () => {
            const mockScore = 10

            beforeEach(() => {
                renderResult = renderwithTemporarySettings({ 'user.daysActiveCount': 3 })
            })

            it('the user is surveyed', () => {
                expect(renderResult.getByText('Tell us what you think')).toBeVisible()
            })

            it('correctly handles dismissing the toast', async () => {
                const closeIcon = renderResult.getByLabelText('Close')
                expect(closeIcon).toBeVisible()
                fireEvent.click(closeIcon)
                expect(await getTemporarySetting('npsSurvey.hasTemporarilyDismissed')).toBe(true)
            })

            it('correctly handles dismissing the toast permanently', async () => {
                const dontShowAgain = renderResult.getByLabelText("Don't show this again")
                expect(dontShowAgain).toBeVisible()
                fireEvent.click(dontShowAgain)

                const closeIcon = renderResult.getByLabelText('Close')
                expect(closeIcon).toBeVisible()
                fireEvent.click(closeIcon)

                expect(await getTemporarySetting('npsSurvey.hasPermanentlyDismissed')).toBe(true)
            })

            it('correctly proceed to use case form', () => {
                const recommendRadioGroup = renderResult.getByLabelText(
                    'How likely is it that you would recommend Sourcegraph to a friend?'
                )
                expect(recommendRadioGroup).toBeVisible()
                const score10 = renderResult.getByLabelText(mockScore)
                fireEvent.click(score10)

                const continueButton = renderResult.getByRole('button', { name: 'Continue' })
                expect(continueButton).toBeVisible()
                fireEvent.click(continueButton)
                expect(renderResult.getByLabelText('What do you use Sourcegraph for?')).toBeVisible()
            })
        })

        describe('on day 4', () => {
            beforeEach(() => {
                renderResult = renderwithTemporarySettings({ 'user.daysActiveCount': 4 })
            })

            it('the user is not surveyed', () => {
                expect(renderResult.container).toBeEmptyDOMElement()
            })
        })

        describe('on day 33', () => {
            beforeEach(() => {
                renderResult = renderwithTemporarySettings({ 'user.daysActiveCount': 33 })
            })

            it('the user is surveyed as it has been 30 days since the last notification', () => {
                expect(renderResult.getByText('Tell us what you think')).toBeVisible()
            })
        })
    })

    describe('toast has been temporarily dismissed by the user', () => {
        beforeEach(() => {
            renderResult = renderwithTemporarySettings({ 'user.daysActiveCount': 33 })
        })

        describe('on day 3', () => {
            beforeEach(() => {
                renderResult = renderwithTemporarySettings({
                    'npsSurvey.hasTemporarilyDismissed': true,
                    'user.daysActiveCount': 3,
                })
            })

            it('the user is not surveyed', () => {
                expect(renderResult.container).toBeEmptyDOMElement()
            })
        })

        describe('on day 30', () => {
            beforeEach(() => {
                renderResult = renderwithTemporarySettings({
                    'npsSurvey.hasTemporarilyDismissed': true,
                    'user.daysActiveCount': 30,
                })
            })

            it('the user is not surveyed but toast dismissal is set to false', async () => {
                expect(renderResult.container).toBeEmptyDOMElement()
                expect(await getTemporarySetting('npsSurvey.hasTemporarilyDismissed')).toBe(false)
            })
        })
    })

    describe('toast has been permanently dismissed by the user', () => {
        describe('on day 3', () => {
            beforeEach(() => {
                renderResult = renderwithTemporarySettings({
                    'npsSurvey.hasPermanentlyDismissed': true,
                    'user.daysActiveCount': 3,
                })
            })

            it('the user is not surveyed', () => {
                expect(renderResult.container).toBeEmptyDOMElement()
            })
        })

        describe('on day 33', () => {
            beforeEach(() => {
                renderResult = renderwithTemporarySettings({
                    'npsSurvey.hasPermanentlyDismissed': true,
                    'user.daysActiveCount': 3,
                })
            })

            it('the user is still not surveyed', () => {
                expect(renderResult.container).toBeEmptyDOMElement()
            })
        })
    })

    describe('user has submitted rating score', () => {
        const moveToUseCaseForm = () => {
            const mockScore = 10
            renderResult = renderwithTemporarySettings({ 'user.daysActiveCount': 3 })
            const score10 = renderResult.getByLabelText(mockScore)
            fireEvent.click(score10)

            const continueButton = renderResult.getByRole('button', { name: 'Continue' })
            fireEvent.click(continueButton)
            expect(renderResult.getByLabelText('What do you use Sourcegraph for?')).toBeVisible()
        }

        beforeEach(() => moveToUseCaseForm())

        it('Should render use case form correctly', () => {
            expect(renderResult.getByLabelText('What do you use Sourcegraph for?')).toBeVisible()
            expect(renderResult.getByLabelText('How can we make Sourcegraph better?')).toBeVisible()
        })

        it('Should show some gratitude after usecase submission', async () => {
            const reasonInput = renderResult.getByLabelText('How can we make Sourcegraph better?')
            expect(reasonInput).toBeVisible()
            fireEvent.change(reasonInput, { target: { value: mockVariables.better } })

            const otherUseCaseInput = renderResult.getByLabelText('What do you use Sourcegraph for?')
            expect(otherUseCaseInput).toBeVisible()
            fireEvent.change(otherUseCaseInput, { target: { value: mockVariables.otherUseCase } })

            const doneButton = renderResult.getByRole('button', { name: 'Done' })
            expect(doneButton).toBeVisible()
            fireEvent.click(doneButton)

            await waitFor(() => {
                expect(renderResult.getByText('Thank you for your feedback!')).toBeVisible()
            })
            expect(await getTemporarySetting('npsSurvey.hasTemporarilyDismissed')).toBe(true)
        })
    })
})
