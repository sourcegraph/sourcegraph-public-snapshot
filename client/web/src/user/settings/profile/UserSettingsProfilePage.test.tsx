import type { MockedResponse } from '@apollo/client/testing'
import { fireEvent, render, type RenderResult, act } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { beforeAll, beforeEach, describe, expect, it } from 'vitest'

import { getDocumentNode } from '@sourcegraph/http-client'
import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { UPDATE_USER } from './EditUserProfileForm'
import { UserSettingsProfilePage } from './UserSettingsProfilePage'

const mockUser = {
    id: 'x',
    username: 'initial-username',
    displayName: 'Initial Name',
    avatarURL: 'https://example.com/image.jpg',
    viewerCanChangeUsername: true,
    createdAt: new Date().toISOString(),
    scimControlled: false,
}

const newUserValues = {
    username: 'new-username',
    displayName: 'New Name',
    avatarURL: 'https://example.com/other-image.jpg',
}

const mocks: readonly MockedResponse[] = [
    {
        request: {
            query: getDocumentNode(UPDATE_USER),
            variables: {
                user: 'x',
                ...newUserValues,
            },
        },
        result: {
            data: {
                updateUser: {
                    id: 'x',
                    ...newUserValues,
                },
            },
        },
    },
]

describe('UserSettingsProfilePage', () => {
    let queries: RenderResult

    beforeAll(() => {
        window.context = { sourcegraphDotComMode: false } as any
    })

    beforeEach(() => {
        queries = render(
            <MockedTestProvider mocks={mocks}>
                <MemoryRouter>
                    <UserSettingsProfilePage user={mockUser} telemetryRecorder={noOpTelemetryRecorder} />
                </MemoryRouter>
            </MockedTestProvider>
        )
    })

    it('renders header correctly', () => {
        const heading = queries.getByRole('heading', { level: 2 })
        expect(heading).toBeVisible()
        expect(heading).toHaveTextContent('Profile')
    })

    it('renders username field correctly', () => {
        const usernameField = queries.getByLabelText('Username')
        expect(usernameField).toBeVisible()
        expect(usernameField).toHaveValue(mockUser.username)
    })

    it('renders display name field correctly', () => {
        const displayNameField = queries.getByLabelText('Display name')
        expect(displayNameField).toBeVisible()
        expect(displayNameField).toHaveValue(mockUser.displayName)
    })

    it('renders avatar URL field correctly', () => {
        const avatarURLField = queries.getByLabelText('Avatar URL')
        expect(avatarURLField).toBeVisible()
        expect(avatarURLField).toHaveValue(mockUser.avatarURL)
    })

    describe('modifying values', () => {
        it('updates values correctly', async () => {
            const usernameField = queries.getByLabelText('Username')
            const displayNameField = queries.getByLabelText('Display name')
            const avatarURLField = queries.getByLabelText('Avatar URL')

            fireEvent.change(usernameField, { target: { value: newUserValues.username } })
            fireEvent.change(displayNameField, { target: { value: newUserValues.displayName } })
            fireEvent.change(avatarURLField, { target: { value: newUserValues.avatarURL } })
            fireEvent.click(queries.getByText('Save'))

            // Wait next tick to skip loading state
            await act(() => new Promise(resolve => setTimeout(resolve, 0)))

            expect(queries.getByText('User profile updated.')).toBeVisible()
            expect(usernameField).toHaveValue(newUserValues.username)
            expect(displayNameField).toHaveValue(newUserValues.displayName)
            expect(avatarURLField).toHaveValue(newUserValues.avatarURL)
        })
    })
})
