import { within, fireEvent } from '@testing-library/react'
import { createPath } from 'react-router-dom'
import { describe, expect, it, vi } from 'vitest'

import { EMPTY_SETTINGS_CASCADE, SettingsProvider } from '@sourcegraph/shared/src/settings/settings'
import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'

import '@sourcegraph/shared/src/testing/mockReactVisibilitySensor'

import { Code } from '@sourcegraph/wildcard'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import type { BlobProps } from '../repo/blob/CodeMirrorBlob'

import { ReferencesPanel } from './ReferencesPanel'
import { buildReferencePanelMocks, defaultProps } from './ReferencesPanel.mocks'

// CodeMirror editor relies on contenteditable property which is not supported by `jsdom`: https://github.com/jsdom/jsdom/issues/1670.
// We need to mock `CodeMirrorBlob to avoid errors.
// More details on CodeMirror react components testing: https://gearheart.io/articles/codemirror-unit-testing-codemirror-react-components/.
function mockCodeMirrorBlob(props: BlobProps) {
    return <Code data-testid="codeMirrorBlobMock">{props.blobInfo.content}</Code>
}
vi.mock('../repo/blob/CodeMirrorBlob', () => ({ CodeMirrorBlob: mockCodeMirrorBlob }))

describe('ReferencesPanel', () => {
    async function renderReferencesPanel() {
        const { url, requestMocks } = buildReferencePanelMocks()

        const result = renderWithBrandedContext(
            <MockedTestProvider mocks={requestMocks}>
                <SettingsProvider settingsCascade={EMPTY_SETTINGS_CASCADE}>
                    <ReferencesPanel {...defaultProps} />
                </SettingsProvider>
            </MockedTestProvider>,
            { route: url }
        )

        await waitForNextApolloResponse()
        await waitForNextApolloResponse()

        return result
    }

    it('renders definitions correctly', async () => {
        const result = await renderReferencesPanel()

        expect(result.getByText('Definitions')).toBeVisible()

        // See MOCK_DEFINITIONS and highlightedLinesDiffGo/highlightedLinesGoDiffGo to see how these expectations here came to be
        const definitions = ['line 16']

        const definitionsList = result.getByTestId('definitions')
        for (const line of definitions) {
            expect(within(definitionsList).getByText(line)).toBeVisible()
        }
    })

    it('renders references correctly', async () => {
        const result = await renderReferencesPanel()

        expect(result.getByText('References')).toBeVisible()

        // See MOCK_REFERENCES and highlightedLinesDiffGo/highlightedLinesGoDiffGo to see how these expectations here came to be
        const references = ['line 46', 'line 16', 'line 52']

        const referencesList = result.getByTestId('references')
        for (const line of references) {
            expect(within(referencesList).getByText(line)).toBeVisible()
        }
    })

    it('renders a code view when clicking on a location', async () => {
        const { locationRef, ...result } = await renderReferencesPanel()

        const definitionsList = result.getByTestId('definitions')
        const referencesList = result.getByTestId('references')

        const referenceButton = within(referencesList).getByTestId('reference-item-diff/diff.go-0')
        const fullReferenceURL =
            '/github.com/sourcegraph/go-diff@9d1f353a285b3094bc33bdae277a19aedabe8b71/-/blob/diff/diff.go?L16:2-16:10'
        expect(referenceButton).toHaveAttribute('href', fullReferenceURL)
        expect(referenceButton).not.toHaveClass('locationActive')

        // Click on reference
        fireEvent.click(referenceButton)

        // Reference is marked as active
        expect(referenceButton).toHaveClass('locationActive')
        // But we have the same in Definitions too, and that should be marked active too
        const definitionButton = within(definitionsList).getByTestId('reference-item-diff/diff.go-0')
        expect(definitionButton).toHaveClass('locationActive')

        // Expect "Loading" message
        const loadingText = result.getByText('Loading ...')
        expect(loadingText).toBeVisible()
        expect(within(loadingText).getByText('diff/diff.go')).toBeVisible()

        // Wait for response
        await waitForNextApolloResponse()

        const closeCodeViewButton = result.getByTestId('close-code-view')
        expect(closeCodeViewButton).toBeVisible()

        const rightPane = result.getByTestId('right-pane')

        // This tests that the header of the "peek" code view on the right exists
        const fileLink = within(rightPane).getByRole('link', { name: 'diff/diff.go' })
        expect(fileLink).toBeVisible()

        // Assert the code view is rendered, by doing a partial match against its content
        const codeView = within(rightPane).getByTestId('codeMirrorBlobMock')
        expect(codeView).toHaveTextContent('package diff import')

        // Assert the current URL points at the reference panel
        expect(createPath(locationRef.current!)).toBe(
            '/github.com/sourcegraph/go-diff@9d1f353a285b3094bc33bdae277a19aedabe8b71/-/blob/diff/diff.go?L16:2#tab=references'
        )
        // Click on reference the second time promotes the active location to the URL (and main blob view)
        fireEvent.click(referenceButton)
        expect(createPath(locationRef.current!)).toBe(fullReferenceURL)
    })
})
