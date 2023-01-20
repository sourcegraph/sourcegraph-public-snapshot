import { render, RenderResult, within, fireEvent } from '@testing-library/react'
import * as H from 'history'
import { unstable_HistoryRouter as HistoryRouter } from 'react-router-dom-v5-compat'

import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import '@sourcegraph/shared/dev/mockReactVisibilitySensor'

import { ReferencesPanel } from './ReferencesPanel'
import { buildReferencePanelMocks, defaultProps } from './ReferencesPanel.mocks'

describe('ReferencesPanel', () => {
    async function renderReferencesPanel() {
        const { url, requestMocks } = buildReferencePanelMocks()

        const externalHistory = H.createMemoryHistory()
        externalHistory.push(url)

        const result: RenderResult = render(
            <HistoryRouter history={externalHistory}>
                <MockedTestProvider mocks={requestMocks}>
                    <ReferencesPanel {...defaultProps} />
                </MockedTestProvider>
            </HistoryRouter>
        )
        await waitForNextApolloResponse()
        await waitForNextApolloResponse()

        return { result, externalHistory }
    }

    it('renders definitions correctly', async () => {
        const { result } = await renderReferencesPanel()

        expect(result.getByText('Definitions')).toBeVisible()

        // See MOCK_DEFINITIONS and highlightedLinesDiffGo/highlightedLinesGoDiffGo to see how these expectations here came to be
        const definitions = ['line 16']

        const definitionsList = result.getByTestId('definitions')
        for (const line of definitions) {
            expect(within(definitionsList).getByText(line)).toBeVisible()
        }
    })

    it('renders references correctly', async () => {
        const { result } = await renderReferencesPanel()

        expect(result.getByText('References')).toBeVisible()

        // See MOCK_REFERENCES and highlightedLinesDiffGo/highlightedLinesGoDiffGo to see how these expectations here came to be
        const references = ['line 46', 'line 16', 'line 52']

        const referencesList = result.getByTestId('references')
        for (const line of references) {
            expect(within(referencesList).getByText(line)).toBeVisible()
        }
    })

    it('renders a code view when clicking on a location', async () => {
        const { result, externalHistory } = await renderReferencesPanel()

        const definitionsList = result.getByTestId('definitions')
        const referencesList = result.getByTestId('references')

        const referenceButton = within(referencesList).getByTestId('reference-item-diff/diff.go-0')
        const fullReferenceURL =
            '/github.com/sourcegraph/go-diff@9d1f353a285b3094bc33bdae277a19aedabe8b71/-/blob/diff/diff.go?L16:2-16:10'
        expect(referenceButton).toHaveAttribute('data-href', fullReferenceURL)
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
        const codeView = within(rightPane).getByRole('table')
        expect(codeView).toHaveTextContent('package diff import')

        // Assert the current URL points at the reference panel
        expect(externalHistory.createHref(externalHistory.location)).toBe(
            '/github.com/sourcegraph/go-diff/-/blob/diff/diff.go?L16:2&subtree=true#tab=references'
        )
        // Click on reference the second time promotes the active location to the URL (and main blob view)
        fireEvent.click(referenceButton)
        expect(externalHistory.createHref(externalHistory.location)).toBe(fullReferenceURL)
    })
})
