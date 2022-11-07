import { useState } from 'react'

import { render, RenderResult, within, fireEvent } from '@testing-library/react'
import * as H from 'history'
import { EMPTY, NEVER, noop, of, Subscription } from 'rxjs'

import { logger } from '@sourcegraph/common'
import { dataOrThrowErrors, useQuery } from '@sourcegraph/http-client'
import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { pretendProxySubscribable, pretendRemote } from '@sourcegraph/shared/src/api/util'
import { ViewerId } from '@sourcegraph/shared/src/api/viewerTypes'
import { Controller } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import '@sourcegraph/shared/dev/mockReactVisibilitySensor'

import { ConnectionQueryArguments } from '../components/FilteredConnection'
import { asGraphQLResult } from '../components/FilteredConnection/utils'
import { UsePreciseCodeIntelForPositionResult, UsePreciseCodeIntelForPositionVariables } from '../graphql-operations'

import { buildPreciseLocation } from './location'
import { ReferencesPanelProps, ReferencesPanelWithMemoryRouter } from './ReferencesPanel'
import { buildReferencePanelMocks, highlightedLinesDiffGo, highlightedLinesGoDiffGo } from './ReferencesPanel.mocks'
import { USE_PRECISE_CODE_INTEL_FOR_POSITION_QUERY } from './ReferencesPanelQueries'
import { UseCodeIntelParameters, UseCodeIntelResult } from './useCodeIntel'

const NOOP_SETTINGS_CASCADE = {
    subjects: null,
    final: null,
}

const NOOP_PLATFORM_CONTEXT: Pick<PlatformContext, 'urlToFile' | 'requestGraphQL' | 'settings'> = {
    requestGraphQL: () => EMPTY,
    urlToFile: () => '',
    settings: of(NOOP_SETTINGS_CASCADE),
}

const NOOP_EXTENSIONS_CONTROLLER: Controller = {
    executeCommand: () => Promise.resolve(),
    registerCommand: () => new Subscription(),
    extHostAPI: Promise.resolve(
        pretendRemote<FlatExtensionHostAPI>({
            getContributions: () => pretendProxySubscribable(NEVER),
            registerContributions: () => pretendProxySubscribable(EMPTY).subscribe(noop as never),
            haveInitialExtensionsLoaded: () => pretendProxySubscribable(of(true)),
            addTextDocumentIfNotExists: () => {},
            addViewerIfNotExists: (): ViewerId => ({ viewerId: 'MOCK_VIEWER_ID' }),
            setEditorSelections: () => {},
            removeViewer: () => {},
        })
    ),
    commandErrors: EMPTY,
    unsubscribe: noop,
}

const defaultProps: Omit<ReferencesPanelProps, 'externalHistory' | 'externalLocation'> = {
    extensionsController: NOOP_EXTENSIONS_CONTROLLER,
    telemetryService: NOOP_TELEMETRY_SERVICE,
    settingsCascade: {
        subjects: null,
        final: null,
    },
    platformContext: NOOP_PLATFORM_CONTEXT,
    isLightTheme: false,
    fetchHighlightedFileLineRanges: args => {
        if (args.filePath === 'cmd/go-diff/go-diff.go') {
            return of(highlightedLinesGoDiffGo)
        }
        if (args.filePath === 'diff/diff.go') {
            return of(highlightedLinesDiffGo)
        }
        logger.error('attempt to fetch highlighted lines for file without mocks', args.filePath)
        return of([])
    },
    useCodeIntel: ({ variables }: UseCodeIntelParameters): UseCodeIntelResult => {
        const [result, setResult] = useState<UseCodeIntelResult>({
            data: {
                implementations: { endCursor: '', nodes: [] },
                references: { endCursor: '', nodes: [] },
                definitions: { endCursor: '', nodes: [] },
            },
            loading: true,
            referencesHasNextPage: false,
            fetchMoreReferences: () => {},
            fetchMoreReferencesLoading: false,
            implementationsHasNextPage: false,
            fetchMoreImplementationsLoading: false,
            fetchMoreImplementations: () => {},
        })
        useQuery<
            UsePreciseCodeIntelForPositionResult,
            UsePreciseCodeIntelForPositionVariables & ConnectionQueryArguments
        >(USE_PRECISE_CODE_INTEL_FOR_POSITION_QUERY, {
            variables,
            notifyOnNetworkStatusChange: false,
            fetchPolicy: 'no-cache',
            skip: !result.loading,
            onCompleted: result => {
                const data = dataOrThrowErrors(asGraphQLResult({ data: result, errors: [] }))
                if (!data || !data.repository?.commit?.blob?.lsif) {
                    return
                }
                const lsif = data.repository.commit.blob.lsif
                setResult(prevResult => ({
                    ...prevResult,
                    loading: false,
                    data: {
                        implementations: {
                            endCursor: lsif.implementations.pageInfo.endCursor,
                            nodes: lsif.implementations.nodes.map(buildPreciseLocation),
                        },
                        references: {
                            endCursor: lsif.references.pageInfo.endCursor,
                            nodes: lsif.references.nodes.map(buildPreciseLocation),
                        },
                        definitions: {
                            endCursor: lsif.definitions.pageInfo.endCursor,
                            nodes: lsif.definitions.nodes.map(buildPreciseLocation),
                        },
                    },
                }))
            },
        })
        return result
    },
}

describe('ReferencesPanel', () => {
    async function renderReferencesPanel() {
        const { url, requestMocks } = buildReferencePanelMocks()

        const fakeExternalHistory = H.createMemoryHistory()
        fakeExternalHistory.push(url)

        const result: RenderResult = render(
            <MockedTestProvider mocks={requestMocks}>
                <ReferencesPanelWithMemoryRouter
                    {...defaultProps}
                    externalHistory={fakeExternalHistory}
                    externalLocation={fakeExternalHistory.location}
                />
            </MockedTestProvider>
        )
        await waitForNextApolloResponse()
        await waitForNextApolloResponse()
        return { result, externalHistory: fakeExternalHistory }
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
