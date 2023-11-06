import { gql } from '@apollo/client'
import type { MockedProviderProps, MockedResponse } from '@apollo/client/testing'
import { describe, expect, test } from '@jest/globals'
import { fireEvent, screen, waitFor } from '@testing-library/react'
import { Route, Routes } from 'react-router-dom'

import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import type { AuthenticatedUser } from '../auth'
import { GetCodyPage } from '../get-cody/GetCodyPage'
import { SUBMIT_CODY_SURVEY } from '../marketing/toast/CodySurveyToast'

import { PostSignUpPage } from './PostSignUpPage'

type MutationMocks = MockedProviderProps['mocks']

function renderPage(
    {
        hasVerifiedEmail,
        completedPostSignup,
    }: {
        hasVerifiedEmail: boolean
        completedPostSignup: boolean
    },
    {
        route,
        mocks,
    }: {
        route?: string
        mocks?: MutationMocks
    } = {}
) {
    const mockUser = {
        id: 'userID',
        username: 'username',
        hasVerifiedEmail,
        completedPostSignup,
    } as AuthenticatedUser

    return renderWithBrandedContext(
        <MockedTestProvider mocks={mocks || []}>
            <Routes>
                <Route path="/post-sign-up" element={<PostSignUpPage authenticatedUser={mockUser} />} />
                <Route
                    path="/get-cody"
                    element={<GetCodyPage authenticatedUser={mockUser} context={{ authProviders: [] }} />}
                />
            </Routes>
        </MockedTestProvider>,
        { route: route || '/post-sign-up' }
    )
}

describe('PostSignUpPage', () => {
    test('renders post signup page - with email verification', () => {
        const { asFragment } = renderPage({ completedPostSignup: false, hasVerifiedEmail: false })
        expect(document.title).toBe('Post signup - Sourcegraph')
        expect(asFragment()).toMatchSnapshot()

        // Renders email verification modal.
        expect(screen.getByText('Verify your email address')).toBeVisible()

        // Render cody survey when next button is clicked
        const nextButton = screen.getByRole('button', { name: 'Next' })
        fireEvent.click(nextButton)

        expect(screen.queryByText('Verify your email address')).not.toBeInTheDocument()
        expect(screen.getByText('How will you be using Cody, our AI assistant?')).toBeVisible()
    })

    test('renders post signup page - with cody survey', () => {
        const submitCodySurveyMock: MockedResponse = {
            request: {
                query: gql(SUBMIT_CODY_SURVEY),
                variables: {
                    isForWork: false,
                    isForPersonal: true,
                },
            },
            result: {
                data: {
                    submitCodySurvey: {
                        alwaysNil: null,
                        __typename: 'EmptyResponse',
                    },
                },
            },
        }

        const { asFragment, locationRef } = renderPage(
            { completedPostSignup: false, hasVerifiedEmail: true },
            { mocks: [submitCodySurveyMock] }
        )
        expect(document.title).toBe('Post signup - Sourcegraph')
        expect(screen.getByText('How will you be using Cody, our AI assistant?')).toBeVisible()
        expect(asFragment()).toMatchSnapshot()

        // Finish the survey
        const personalCheckbox = screen.getByLabelText('for personal stuff')
        fireEvent.click(personalCheckbox)
        const submitButton = screen.getByRole('button', { name: 'Get started' })
        expect(submitButton).toBeEnabled()
        fireEvent.click(submitButton)

        // Survey submission is asynchronous so wait for redirect
        waitFor(() => expect(locationRef.current?.pathname).toBe('/get-cody'))
    })

    test('redirects to customized page after survey submission', () => {
        const submitCodySurveyMock: MockedResponse = {
            request: {
                query: gql(SUBMIT_CODY_SURVEY),
                variables: {
                    isForWork: true,
                    isForPersonal: false,
                },
            },
            result: {
                data: {
                    submitCodySurvey: {
                        alwaysNil: null,
                        __typename: 'EmptyResponse',
                    },
                },
            },
        }

        const { locationRef } = renderPage(
            { completedPostSignup: false, hasVerifiedEmail: true },
            { route: '/post-sign-up?returnTo=/foo?bar=baz', mocks: [submitCodySurveyMock] }
        )

        const workCheckbox = screen.getByLabelText('for work')
        fireEvent.click(workCheckbox)
        const submitButton = screen.getByRole('button', { name: 'Get started' })
        fireEvent.click(submitButton)

        waitFor(() => {
            expect(locationRef.current?.pathname).toBe('/foo')
            expect(locationRef.current?.search).toBe('?bar=baz')
        })
    })

    test('renders redirect when user has completed post signup flow', () => {
        const { asFragment, locationRef } = renderPage({ completedPostSignup: true, hasVerifiedEmail: true })
        expect(document.title).toBe('Sourcegraph')
        expect(asFragment()).toMatchSnapshot()
        expect(locationRef.current?.pathname).toBe('/search')
    })

    test('renders customized redirect when user has completed post signup flow', () => {
        const { asFragment, locationRef } = renderPage(
            { completedPostSignup: true, hasVerifiedEmail: true },
            { route: '/post-sign-up?returnTo=/foo' }
        )
        expect(document.title).toBe('Sourcegraph')
        expect(asFragment()).toMatchSnapshot()
        expect(locationRef.current?.pathname).toBe('/foo')
    })
})
