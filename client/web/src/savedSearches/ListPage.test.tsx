import { screen } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'

import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { viewerAffiliatedNamespacesMock } from '../namespaces/graphql.mocks'

import { MOCK_SAVED_SEARCH_FIELDS, savedSearchesMock } from './graphql.mocks'
import { ListPage } from './ListPage'

const mockTelemetryRecorder = {
    recordEvent: vi.fn(),
}

describe('ListPage', () => {
    test('lists', async () => {
        renderWithBrandedContext(
            <MockedTestProvider mocks={[savedSearchesMock, viewerAffiliatedNamespacesMock]}>
                <ListPage telemetryRecorder={mockTelemetryRecorder} />
            </MockedTestProvider>
        )
        await waitForNextApolloResponse()
        expect(screen.getByTestId('saved-searches-list-page')).toBeInTheDocument()
        expect(screen.getByText(MOCK_SAVED_SEARCH_FIELDS.description)).toBeInTheDocument()
    })
})
