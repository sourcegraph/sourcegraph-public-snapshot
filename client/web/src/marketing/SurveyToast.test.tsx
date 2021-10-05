import { cleanup, within, fireEvent } from '@testing-library/react'
import React from 'react'

import { renderWithRouter, RenderWithRouterResult } from '@sourcegraph/shared/src/testing/render-with-router'

import {
    DAYS_ACTIVE_STORAGE_KEY,
    HAS_DISMISSED_TOAST_STORAGE_KEY,
    HAS_PERMANENTLY_DISMISSED_TOAST_STORAGE_KEY,
} from './constants'
import { SurveyToast } from './SurveyToast'

describe('SurveyToast', () => {
    let renderResult: RenderWithRouterResult

    afterEach(() => {
        localStorage.clear()
        cleanup()
    })

    const setDaysActive = (daysActive: string) => localStorage.setItem(DAYS_ACTIVE_STORAGE_KEY, daysActive)
    const setToastDismissed = (dismissed: string) => localStorage.setItem(HAS_DISMISSED_TOAST_STORAGE_KEY, dismissed)
    const setToastPermanentlyDismissed = (dismissed: string) =>
        localStorage.setItem(HAS_PERMANENTLY_DISMISSED_TOAST_STORAGE_KEY, dismissed)

    describe('toast has not been dismissed by the user', () => {
        beforeEach(() => {
            setToastDismissed('false')
        })

        describe('before day 3', () => {
            beforeEach(() => {
                setDaysActive('1')
                renderResult = renderWithRouter(<SurveyToast />)
            })

            it('the user is not surveyed', () => {
                expect(renderResult.container).toBeEmptyDOMElement()
            })
        })

        describe('on day 3', () => {
            const mockScore = 10

            beforeEach(() => {
                setDaysActive('3')
                renderResult = renderWithRouter(<SurveyToast />)
            })

            it('the user is surveyed', () => {
                expect(renderResult.getByText('Tell us what you think')).toBeVisible()
            })

            it('correctly handles dismissing the toast', () => {
                const closeIcon = renderResult.getByLabelText('Close')
                expect(closeIcon).toBeVisible()
                fireEvent.click(closeIcon)
                expect(localStorage.getItem(HAS_DISMISSED_TOAST_STORAGE_KEY)).toBe('true')
            })

            it('correctly handles dismissing the toast permanently', () => {
                const dontShowAgain = renderResult.getByLabelText("Don't show this again")
                expect(dontShowAgain).toBeVisible()
                fireEvent.click(dontShowAgain)

                const closeIcon = renderResult.getByLabelText('Close')
                expect(closeIcon).toBeVisible()
                fireEvent.click(closeIcon)

                expect(localStorage.getItem(HAS_DISMISSED_TOAST_STORAGE_KEY)).toBe('true')
                expect(localStorage.getItem(HAS_PERMANENTLY_DISMISSED_TOAST_STORAGE_KEY)).toBe('true')
            })

            it('correctly submits and dismisses the toast', () => {
                const recommendRadioGroup = renderResult.getByLabelText(
                    'How likely is it that you would recommend Sourcegraph to a friend?'
                )
                expect(recommendRadioGroup).toBeVisible()
                const score10 = within(recommendRadioGroup).getByLabelText(mockScore)
                fireEvent.click(score10)
                expect(renderResult.history.location.pathname).toBe(`/survey/${mockScore}`)
                expect(localStorage.getItem(HAS_DISMISSED_TOAST_STORAGE_KEY)).toBe('true')
            })

            it('correctly submits and permanently dismisses the toast', () => {
                const dontShowAgain = renderResult.getByLabelText("Don't show this again")
                expect(dontShowAgain).toBeVisible()
                fireEvent.click(dontShowAgain)

                const recommendRadioGroup = renderResult.getByLabelText(
                    'How likely is it that you would recommend Sourcegraph to a friend?'
                )
                expect(recommendRadioGroup).toBeVisible()
                const score10 = within(recommendRadioGroup).getByLabelText(mockScore)
                fireEvent.click(score10)

                expect(renderResult.history.location.pathname).toBe(`/survey/${mockScore}`)
                expect(localStorage.getItem(HAS_DISMISSED_TOAST_STORAGE_KEY)).toBe('true')
                expect(localStorage.getItem(HAS_PERMANENTLY_DISMISSED_TOAST_STORAGE_KEY)).toBe('true')
            })
        })

        describe('on day 4', () => {
            beforeEach(() => {
                setDaysActive('4')
                renderResult = renderWithRouter(<SurveyToast />)
            })

            it('the user is not surveyed', () => {
                expect(renderResult.container).toBeEmptyDOMElement()
            })
        })

        describe('on day 33', () => {
            beforeEach(() => {
                setDaysActive('33')
                renderResult = renderWithRouter(<SurveyToast />)
            })

            it('the user is surveyed as it has been 30 days since the last notification', () => {
                expect(renderResult.getByText('Tell us what you think')).toBeVisible()
            })
        })
    })

    describe('toast has been dismissed by the user', () => {
        beforeEach(() => {
            setToastDismissed('true')
        })

        describe('on day 3', () => {
            beforeEach(() => {
                setDaysActive('3')
                renderResult = renderWithRouter(<SurveyToast />)
            })

            it('the user is not surveyed', () => {
                expect(renderResult.container).toBeEmptyDOMElement()
            })
        })

        describe('on day 30', () => {
            beforeEach(() => {
                setDaysActive('30')
                renderResult = renderWithRouter(<SurveyToast />)
            })

            it('the user is not surveyed but toast dismissal is cleared', () => {
                expect(renderResult.container).toBeEmptyDOMElement()
                expect(localStorage.getItem(HAS_DISMISSED_TOAST_STORAGE_KEY)).toBe('false')
            })
        })
    })

    describe('toast has been permanently dismissed by the user', () => {
        beforeEach(() => {
            setToastPermanentlyDismissed('true')
        })

        describe('on day 3', () => {
            beforeEach(() => {
                setDaysActive('3')
                renderResult = renderWithRouter(<SurveyToast />)
            })

            it('the user is not surveyed', () => {
                expect(renderResult.container).toBeEmptyDOMElement()
            })
        })

        describe('on day 33', () => {
            beforeEach(() => {
                setDaysActive('33')
                renderResult = renderWithRouter(<SurveyToast />)
            })

            it('the user is still not surveyed', () => {
                expect(renderResult.container).toBeEmptyDOMElement()
            })
        })
    })
})
