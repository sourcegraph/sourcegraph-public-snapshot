import { screen } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'

import type { LazyQueryInputProps } from '@sourcegraph/branded'
import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { Input } from '@sourcegraph/wildcard'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { EditPage } from './EditPage'
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

describe('EditPage', () => {
    test('found', async () => {
        renderWithBrandedContext(
            <MockedTestProvider mocks={[savedSearchMock]}>
                <EditPage telemetryRecorder={mockTelemetryRecorder} isSourcegraphDotCom={false} />
            </MockedTestProvider>,
            { route: '/saved-searches/1/edit', path: '/saved-searches/:id/edit' }
        )
        await waitForNextApolloResponse()
        expect(screen.getByText('Editing: My description - saved search', { exact: false })).toBeInTheDocument()
        expect(screen.getByLabelText('Description')).toHaveValue('My description')
        expect(screen.getByLabelText('Query')).toHaveValue('my repo:query')
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
                <EditPage telemetryRecorder={mockTelemetryRecorder} isSourcegraphDotCom={false} />
            </MockedTestProvider>,
            { route: '/saved-searches/1/edit', path: '/saved-searches/:id/edit' }
        )
        await waitForNextApolloResponse()
        expect(screen.getByText('Saved search not found.')).toBeInTheDocument()
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
                <EditPage telemetryRecorder={mockTelemetryRecorder} isSourcegraphDotCom={false} />
            </MockedTestProvider>,
            { route: '/saved-searches/1/edit', path: '/saved-searches/:id/edit' }
        )
        await waitForNextApolloResponse()
        expect(screen.getByText('You do not have permission to edit this saved search.')).toBeInTheDocument()
    })
})
