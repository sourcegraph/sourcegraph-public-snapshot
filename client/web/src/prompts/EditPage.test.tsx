import { screen } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { EditPage } from './EditPage'
import { promptQuery } from './graphql'
import { MOCK_PROMPT_FIELDS, promptMock } from './graphql.mocks'

const mockTelemetryRecorder = {
    recordEvent: vi.fn(),
}

describe('EditPage', () => {
    test('found', async () => {
        renderWithBrandedContext(
            <MockedTestProvider mocks={[promptMock]}>
                <EditPage telemetryRecorder={mockTelemetryRecorder} />
            </MockedTestProvider>,
            { route: '/prompts/1/edit', path: '/prompts/:id/edit' }
        )
        await waitForNextApolloResponse()
        expect(screen.getByText('Editing: alice/my-prompt - prompt', { exact: false })).toBeInTheDocument()
        expect(screen.getByLabelText('Description (optional)')).toHaveValue('My description')
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
                <EditPage telemetryRecorder={mockTelemetryRecorder} />
            </MockedTestProvider>,
            { route: '/prompts/1/edit', path: '/prompts/:id/edit' }
        )
        await waitForNextApolloResponse()
        expect(screen.getByText('Prompt not found.')).toBeInTheDocument()
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
                <EditPage telemetryRecorder={mockTelemetryRecorder} />
            </MockedTestProvider>,
            { route: '/prompts/1/edit', path: '/prompts/:id/edit' }
        )
        await waitForNextApolloResponse()
        expect(screen.getByText('You do not have permission to edit this prompt.')).toBeInTheDocument()
    })
})
