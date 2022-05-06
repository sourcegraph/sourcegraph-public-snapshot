import { render, RenderResult, within, fireEvent } from '@testing-library/react'
import * as H from 'history'
import { EMPTY, NEVER, noop, of, Subscription } from 'rxjs'
import { HoverThresholdProps } from 'src/repo/RepoContainer'

import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { pretendProxySubscribable, pretendRemote } from '@sourcegraph/shared/src/api/util'
import { ViewerId } from '@sourcegraph/shared/src/api/viewerTypes'
import { Controller, ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContext, PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE, TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider, waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { ReferencesPanelWithMemoryRouter } from './ReferencesPanel'
import { buildReferencePanelMocks } from './ReferencesPanel.mocks'

const NOOP_SETTINGS_CASCADE = {
    subjects: null,
    final: null,
}

const NOOP_PLATFORM_CONTEXT: Pick<
    PlatformContext,
    'urlToFile' | 'requestGraphQL' | 'settings' | 'forceUpdateTooltip'
> = {
    requestGraphQL: () => EMPTY,
    urlToFile: () => '',
    settings: of(NOOP_SETTINGS_CASCADE),
    forceUpdateTooltip: () => {},
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

const defaultProps: SettingsCascadeProps &
    PlatformContextProps<'urlToFile' | 'requestGraphQL' | 'settings' | 'forceUpdateTooltip'> &
    TelemetryProps &
    HoverThresholdProps &
    ExtensionsControllerProps &
    ThemeProps = {
    extensionsController: NOOP_EXTENSIONS_CONTROLLER,
    telemetryService: NOOP_TELEMETRY_SERVICE,
    settingsCascade: {
        subjects: null,
        final: null,
    },
    platformContext: NOOP_PLATFORM_CONTEXT,
    isLightTheme: false,
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
        return result
    }

    it('renders definitions correctly', async () => {
        const result = await renderReferencesPanel()

        expect(result.getByText('Definitions')).toBeVisible()

        const definitions = [['', 'string']]

        const definitionsList = result.getByTestId('definitions')
        for (const line of definitions) {
            for (const surrounding of line) {
                if (surrounding === '') {
                    continue
                }
                expect(within(definitionsList).getByText(surrounding)).toBeVisible()
            }
        }
    })

    it('renders references correctly', async () => {
        const result = await renderReferencesPanel()

        expect(result.getByText('References')).toBeVisible()

        const references = [
            ['', 'string'],
            ['label = fmt.Sprintf("orig(%s) new(%s)", fdiff.', ', fdiff.NewName)'],
            ['if err := printFileHeader(&buf, "--- ", d.', ', d.OrigTime); err != nil {'],
        ]

        const referencesList = result.getByTestId('references')
        for (const line of references) {
            for (const surrounding of line) {
                if (surrounding === '') {
                    continue
                }
                expect(within(referencesList).getByText(surrounding)).toBeVisible()
            }
        }
    })

    it('renders a code view when clicking on a location', async () => {
        const result = await renderReferencesPanel()

        const definitionsList = result.getByTestId('definitions')
        const referencesList = result.getByTestId('references')

        const referenceButton = within(referencesList).getByRole('button', { name: '16: OrigName string' })
        const fullReferenceURL =
            '/github.com/sourcegraph/go-diff@9d1f353a285b3094bc33bdae277a19aedabe8b71/-/blob/diff/diff.go?L16:2-16:10'
        expect(referenceButton).toHaveAttribute('to', fullReferenceURL)
        expect(referenceButton.parentNode).not.toHaveClass('locationActive')

        // Click on reference
        fireEvent.click(referenceButton)

        // Reference is marked as active
        expect(referenceButton.parentNode).toHaveClass('locationActive')
        // But we have the same in Definitions too, and that should be marked active too
        const definitionButton = within(definitionsList).getByRole('button', { name: '16: OrigName string' })
        expect(definitionButton.parentNode).toHaveClass('locationActive')

        // Expect "Loading" message
        const loadingText = result.getByText('Loading ...')
        expect(loadingText).toBeVisible()
        expect(within(loadingText).getByText('diff/diff.go')).toBeVisible()

        // Wait for response
        await waitForNextApolloResponse()

        const closeCodeViewButton = result.getByRole('button', { name: 'Close code view' })
        expect(closeCodeViewButton).toBeVisible()

        const fileLink = result.getByRole('link', { name: 'diff/diff.go' })
        expect(fileLink).toBeVisible()

        const codeView = result.getByRole('table')
        // Assert the code view is rendered, by doing a partial match against its content
        expect(codeView).toHaveTextContent('package diff import')
    })
})
