import { cleanup, within, fireEvent } from '@testing-library/react'
import React from 'react'

import { renderWithRouter, RenderWithRouterResult } from '@sourcegraph/shared/src/testing/render-with-router'

import { DAYS_ACTIVE_STORAGE_KEY, HAS_DISMISSED_TOAST_STORAGE_KEY } from './constants'
import { SurveyToast } from './SurveyToast'

describe('SurveyToast', () => {
    let renderResult: RenderWithRouterResult

    afterEach(cleanup)

    const setDaysActive = (daysActive: string) => localStorage.setItem(DAYS_ACTIVE_STORAGE_KEY, daysActive)
    const setToastDismissed = (dismissed: string) => localStorage.setItem(HAS_DISMISSED_TOAST_STORAGE_KEY, dismissed)

    describe('before day 3', () => {
        beforeEach(() => {
            setDaysActive('1')
            setToastDismissed('false')
            renderResult = renderWithRouter(<SurveyToast />)
        })

        it('the user is not surveyed', () => {
            expect(renderResult.container).toBeEmptyDOMElement()
        })
    })

    describe('on day 3', () => {
        beforeEach(() => {
            setDaysActive('3')
            setToastDismissed('false')
            renderResult = renderWithRouter(<SurveyToast />)
        })

        it('the user is surveyed', () => {
            expect(renderResult.getByText('Tell us what you think')).toBeVisible()
        })

        it('submitting the form correctly navigates to the survey page and dismisses the toast', () => {
            const mockScore = 10
            const recommendRadioGroup = renderResult.getByLabelText(
                'How likely is it that you would recommend Sourcegraph to a friend?'
            )
            expect(recommendRadioGroup).toBeVisible()
            const score10 = within(recommendRadioGroup).getByLabelText(mockScore)
            fireEvent.click(score10)
            expect(renderResult.history.location.pathname).toBe(`/survey/${mockScore}`)
            expect(localStorage.getItem(HAS_DISMISSED_TOAST_STORAGE_KEY)).toBe('true')
        })
    })

    describe('on day 4', () => {
        beforeEach(() => {
            setDaysActive('4')
            setToastDismissed('false')
            renderResult = renderWithRouter(<SurveyToast />)
        })

        it('the user is not surveyed', () => {
            expect(renderResult.container).toBeEmptyDOMElement()
        })
    })

    describe('on day 33', () => {
        beforeEach(() => {
            setDaysActive('33')
            setToastDismissed('false')
            renderResult = renderWithRouter(<SurveyToast />)
        })

        it('the user is surveyed as it has been 30 days since the last notification', () => {
            expect(renderResult.getByText('Tell us what you think')).toBeVisible()
        })
    })
})
