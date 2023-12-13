import { act, fireEvent } from '@testing-library/react'
import { Route, Routes } from 'react-router-dom'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import type { SourcegraphContext } from '../jscontext'

import { RequestAccessPage } from './RequestAccessPage'

function renderPage({
    route = '/request-access',
    sourcegraphDotComMode = false,
    allowSignup = false,
    isAuthenticatedUser = false,
    xhrHeaders = {},
    authAccessRequest,
}: {
    route?: string
    isAuthenticatedUser?: boolean
    sourcegraphDotComMode?: boolean
    allowSignup?: boolean
    authAccessRequest?: SourcegraphContext['authAccessRequest']
    xhrHeaders?: Record<string, string>
} = {}) {
    window.context = {
        sourcegraphDotComMode,
        allowSignup,
        authAccessRequest,
        isAuthenticatedUser,
        xhrHeaders,
    } as any
    return renderWithBrandedContext(
        <Routes>
            <Route path="/request-access/*" element={<RequestAccessPage telemetryRecorder={noOpTelemetryRecorder} />} />
            <Route path="/sign-in" element={<div>Sign in</div>} />
        </Routes>,
        { route }
    )
}

describe('RequestAccessPage', () => {
    const origContext = window.context
    afterEach(() => {
        window.context = origContext
    })

    test('renders form if all conditions are met', () => {
        const { locationRef, getByTestId, asFragment } = renderPage()

        expect(locationRef?.current?.pathname).toBe('/request-access')
        expect(document.title).toBe('Request access - Sourcegraph')
        expect(getByTestId('request-access-form')).toBeInTheDocument()
        expect(asFragment()).toMatchSnapshot()
    })

    test('handles inputs', () => {
        const { getByLabelText, getByText } = renderPage()
        const nameInput = getByLabelText('Name') as HTMLInputElement
        expect(nameInput.value).toBe('')
        fireEvent.change(nameInput, { target: { value: 'John Wick' } })
        expect(nameInput.value).toBe('John Wick')

        const emailInput = getByLabelText('Email Address') as HTMLInputElement
        expect(emailInput.value).toBe('')
        fireEvent.change(emailInput, { target: { value: 'john@wick.com' } })
        expect(emailInput.value).toBe('john@wick.com')

        const companyInput = getByLabelText('Notes for administrator') as HTMLInputElement
        expect(companyInput.value).toBe('')
        fireEvent.change(companyInput, { target: { value: "I think I', back" } })
        expect(companyInput.value).toBe("I think I', back")

        const button = getByText('Request access')
        expect(button).toBeInTheDocument()
        expect(button).toBeInstanceOf(HTMLButtonElement)
        act(() => {
            fireEvent.click(button)
        })
    })

    describe('redirects', () => {
        test('if authenticatedUser', () => {
            const { locationRef } = renderPage({ isAuthenticatedUser: true })
            expect(locationRef?.current?.pathname).toBe('/search')
        })

        test('if allowSignup=true', () => {
            const { locationRef } = renderPage({ allowSignup: true })
            expect(locationRef?.current?.pathname).toBe('/sign-in')
        })

        test('if sourcegraphDotComMode=true', () => {
            const { locationRef } = renderPage({ sourcegraphDotComMode: true })
            expect(locationRef?.current?.pathname).toBe('/sign-in')
        })

        test('if auth.accessRequest.enabled=false', () => {
            const { locationRef } = renderPage({
                authAccessRequest: {
                    enabled: false,
                },
            })
            expect(locationRef?.current?.pathname).toBe('/sign-in')
        })
    })
})
