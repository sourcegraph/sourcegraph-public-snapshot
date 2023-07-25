import { fireEvent, screen } from '@testing-library/react'
import { Route, Routes } from 'react-router-dom'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { AuthenticatedUser } from '../auth'
import { GetCodyPage } from '../get-cody/GetCodyPage'

import { PostSignUpPage } from './PostSignUpPage'

function renderPage({
    hasVerifiedEmail,
    completedPostSignup,
}: {
    hasVerifiedEmail: boolean
    completedPostSignup: boolean
}) {
    const mockUser = {
        id: 'userID',
        username: 'username',
        hasVerifiedEmail,
        completedPostSignup,
    } as AuthenticatedUser

    return renderWithBrandedContext(
        <MockedTestProvider mocks={[]}>
            <Routes>
                <Route path="/post-sign-up" element={<PostSignUpPage authenticatedUser={mockUser} />} />
                <Route
                    path="/get-cody"
                    element={<GetCodyPage authenticatedUser={mockUser} context={{ authProviders: [] }} />}
                />
            </Routes>
        </MockedTestProvider>,
        { route: '/post-sign-up' }
    )
}

describe('PostSignUpPage', () => {
    test('renders post signup page - with email verification', () => {
        const { asFragment } = renderPage({ completedPostSignup: false, hasVerifiedEmail: false })
        expect(document.title).toBe('Post signup - Sourcegraph')
        expect(asFragment()).toMatchSnapshot()

        // Renders email verification modal.
        expect(screen.getByText('Verify your email address')).toBeVisible()

        // Render cody survery when next button is clicked
        const nextButton = screen.getByRole('button', { name: 'Next' })
        fireEvent.click(nextButton)

        expect(screen.queryByText('Verify your email address')).not.toBeInTheDocument()
        expect(screen.getByText('How will you be using Cody, our AI assistant?')).toBeVisible()
    })

    test('renders post signup page - with cody survey', () => {
        const { asFragment } = renderPage({ completedPostSignup: false, hasVerifiedEmail: true })
        expect(document.title).toBe('Post signup - Sourcegraph')
        expect(screen.getByText('How will you be using Cody, our AI assistant?')).toBeVisible()
        expect(asFragment()).toMatchSnapshot()
    })

    test('renders redirect when user has completed post signup flow', () => {
        const { asFragment } = renderPage({ completedPostSignup: true, hasVerifiedEmail: true })
        expect(document.title).toBe('Sourcegraph')
        expect(asFragment()).toMatchSnapshot()
    })
})
