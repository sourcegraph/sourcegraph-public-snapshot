import { screen } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { DetailPage } from './DetailPage'
import { promptQuery } from './graphql'
import { MOCK_PROMPT_FIELDS, promptMock } from './graphql.mocks'

const mockTelemetryRecorder = {
    recordEvent: vi.fn(),
}

describe('DetailPage', () => {
    test('found', async () => {
        renderWithBrandedContext(
            <MockedTestProvider mocks={[promptMock]}>
                <DetailPage telemetryRecorder={mockTelemetryRecorder} />
            </MockedTestProvider>,
            { route: '/prompts/1', path: '/prompts/:id' }
        )
        await waitForNextApolloResponse()
        expect(screen.getAllByText('Prompt alice/my-prompt', { exact: false }).at(0)).toBeInTheDocument()
        expect(screen.queryByRole('link', { name: 'Edit' })).toBeInTheDocument()
    })

    test('!viewerCanAdminister', async () => {
        renderWithBrandedContext(
            <MockedTestProvider
                mocks={[
                    {
                        request: {
                            query: getDocumentNode(promptQuery),
                            variables: { id: '1' },
                        },
                        result: {
                            data: {
                                node: {
                                    ...MOCK_PROMPT_FIELDS,
                                    viewerCanAdminister: false,
                                },
                            },
                        },
                    },
                ]}
            >
                <DetailPage telemetryRecorder={mockTelemetryRecorder} />
            </MockedTestProvider>,
            { route: '/prompts/1', path: '/prompts/:id' }
        )
        await waitForNextApolloResponse()
        expect(screen.getAllByText('Prompt alice/my-prompt', { exact: false }).at(0)).toBeInTheDocument()
        expect(screen.queryByRole('link', { name: 'Edit' })).not.toBeInTheDocument()
    })

    test('not found', async () => {
        renderWithBrandedContext(
            <MockedTestProvider
                mocks={[
                    {
                        request: {
                            query: getDocumentNode(promptQuery),
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
            { route: '/prompts/1', path: '/prompts/:id' }
        )
        await waitForNextApolloResponse()
        expect(screen.getByText('Prompt not found.')).toBeInTheDocument()
    })
})
