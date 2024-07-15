import { screen } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'

import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { viewerAffiliatedNamespacesMock } from '../namespaces/graphql.mocks'

import { NewForm } from './NewForm'

const mockTelemetryRecorder = {
    recordEvent: vi.fn(),
}

describe('NewForm', () => {
    test('renders', async () => {
        renderWithBrandedContext(
            <MockedTestProvider mocks={[viewerAffiliatedNamespacesMock]}>
                <NewForm telemetryRecorder={mockTelemetryRecorder} />
            </MockedTestProvider>
        )
        await waitForNextApolloResponse()
        expect(screen.getByRole('button', { name: 'Create prompt' })).toBeInTheDocument()
    })
})
