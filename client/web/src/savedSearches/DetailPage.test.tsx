import { screen } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'

import type { LazyQueryInputProps } from '@sourcegraph/branded'
import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { Input } from '@sourcegraph/wildcard'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { DetailPage } from './DetailPage'
import { savedSearchQuery } from './graphql'
import { MOCK_SAVED_SEARCH_FIELDS, savedSearchMock } from './graphql.mocks'

const mockTelemetryRecorder = {
    recordEvent: vi.fn(),
}

vi.mock('@sourcegraph/branded/src/search-ui/input/LazyQueryInput', () => ({
    LazyQueryInputFormControl: (props: Pick<LazyQueryInputProps, 'queryState' | 'ariaLabelledby'>) => (
        <Input defaultValue={props.queryState.query} aria-labelledby={props.ariaLabelledby} />
    ),
}))

describe('DetailPage', () => {
    test('found', async () => {
        renderWithBrandedContext(
            <MockedTestProvider mocks={[savedSearchMock]}>
                <DetailPage telemetryRecorder={mockTelemetryRecorder} />
            </MockedTestProvider>,
            { route: '/saved-searches/1', path: '/saved-searches/:id' }
        )
        await waitForNextApolloResponse()
        expect(screen.getAllByText('My description - saved search', { exact: false }).at(0)).toBeInTheDocument()
        expect(screen.queryByRole('link', { name: 'Edit' })).toBeInTheDocument()
    })

    test('!viewerCanAdminister', async () => {
        renderWithBrandedContext(
            <MockedTestProvider
                mocks={[
                    {
                        request: {
                            query: getDocumentNode(savedSearchQuery),
                            variables: { id: '1' },
                        },
                        result: {
                            data: {
                                node: {
                                    ...MOCK_SAVED_SEARCH_FIELDS,
                                    viewerCanAdminister: false,
                                },
                            },
                        },
                    },
                ]}
            >
                <DetailPage telemetryRecorder={mockTelemetryRecorder} />
            </MockedTestProvider>,
            { route: '/saved-searches/1', path: '/saved-searches/:id' }
        )
        await waitForNextApolloResponse()
        expect(screen.getAllByText('My description - saved search', { exact: false }).at(0)).toBeInTheDocument()
        expect(screen.queryByRole('link', { name: 'Edit' })).not.toBeInTheDocument()
    })

    test('not found', async () => {
        renderWithBrandedContext(
            <MockedTestProvider
                mocks={[
                    {
                        request: {
                            query: getDocumentNode(savedSearchQuery),
                            variables: { id: '1' },
                        },
                        result: {
                            data: {
                                node: null,
                            },
                        },
                    },
                ]}
            >
                <DetailPage telemetryRecorder={mockTelemetryRecorder} />
            </MockedTestProvider>,
            { route: '/saved-searches/1', path: '/saved-searches/:id' }
        )
        await waitForNextApolloResponse()
        expect(screen.getByText('Saved search not found.')).toBeInTheDocument()
    })
})
