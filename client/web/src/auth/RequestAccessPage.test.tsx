import { act, fireEvent } from '@testing-library/react'
import { Route, Routes } from 'react-router-dom-v5-compat'

import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import { SiteConfiguration } from '@sourcegraph/shared/src/schema/site.schema'
import { mockAuthenticatedUser } from '@sourcegraph/shared/src/testing/searchContexts/testHelpers'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { RequestAccessPage } from './RequestAccessPage'

function renderPage({
    route = '/request-access',
    authenticatedUser = null,
    sourcegraphDotComMode = false,
    allowSignup = false,
    experimentalFeatures = {},
}: {
    route?: string
    authenticatedUser?: AuthenticatedUser | null
    sourcegraphDotComMode?: boolean
    allowSignup?: boolean
    experimentalFeatures?: SiteConfiguration['experimentalFeatures']
} = {}) {
    return renderWithBrandedContext(
        <Routes>
            <Route
                path="/request-access/*"
                element={
                    <RequestAccessPage
                        authenticatedUser={authenticatedUser}
                        context={{
                            allowSignup,
                            sourcegraphDotComMode,
                            experimentalFeatures,
                            xhrHeaders: {},
                        }}
                    />
                }
            />
            <Route path="/sign-in" element={<div>Sign in</div>} />
        </Routes>,
        { route }
    )
}

describe('RequestAccessPage', () => {
    test('renders form if all conditions are met', () => {
        const { history, getByTestId } = renderPage()

        expect(history.location.pathname).toBe('/request-access')
        expect(document.title).toBe('Request access - Sourcegraph')
        expect(getByTestId('request-access-form')).toBeInTheDocument()
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

        const companyInput = getByLabelText('Extra information') as HTMLInputElement
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
            const { history } = renderPage({ authenticatedUser: mockAuthenticatedUser })
            expect(history.location.pathname).toBe('/search')
        })

        test('if allowSignup=true', () => {
            const { history } = renderPage({ allowSignup: true })
            expect(history.location.pathname).toBe('/sign-in')
        })

        test('if sourcegraphDotComMode=true', () => {
            const { history } = renderPage({ sourcegraphDotComMode: true })
            expect(history.location.pathname).toBe('/sign-in')
        })

        test('if experimentalFeatures.accessRequests.enabled=false', () => {
            const { history } = renderPage({
                experimentalFeatures: {
                    accessRequests: {
                        enabled: false,
                    },
                },
            })
            expect(history.location.pathname).toBe('/sign-in')
        })
    })
})
