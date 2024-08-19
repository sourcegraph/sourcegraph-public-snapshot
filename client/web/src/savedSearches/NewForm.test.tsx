import { screen } from '@testing-library/react'
import { describe, expect, test, vi } from 'vitest'

import { type LazyQueryInputProps } from '@sourcegraph/branded'
import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { Input } from '@sourcegraph/wildcard'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { viewerAffiliatedNamespacesMock } from '../namespaces/graphql.mocks'

import { NewForm } from './NewForm'

const mockTelemetryRecorder = {
    recordEvent: vi.fn(),
}

vi.mock('@sourcegraph/branded/src/search-ui/input/LazyQueryInput', () => ({
    LazyQueryInputFormControl: (props: Pick<LazyQueryInputProps, 'queryState' | 'ariaLabelledby'>) => (
        <Input defaultValue={props.queryState.query} aria-labelledby={props.ariaLabelledby} />
    ),
}))

describe('NewForm', () => {
    test('renders', async () => {
        renderWithBrandedContext(
            <MockedTestProvider mocks={[viewerAffiliatedNamespacesMock]}>
                <NewForm isSourcegraphDotCom={false} telemetryRecorder={mockTelemetryRecorder} />
            </MockedTestProvider>
        )
        await waitForNextApolloResponse()
        expect(screen.getByRole('button', { name: 'Create saved search' })).toBeInTheDocument()
    })
})
