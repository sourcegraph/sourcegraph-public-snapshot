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

import { Location } from './location'
import { getLineContent, ReferencesPanelWithMemoryRouter } from './ReferencesPanel'
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
        expect(referenceButton).toHaveAttribute('data-test-reference-url', fullReferenceURL)
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

        const closeCodeViewButton = result.getByTestId('close-code-view')
        expect(closeCodeViewButton).toBeVisible()

        const rightPane = result.getByTestId('right-pane')

        // This tests that the header of the "peek" code view on the right exists
        const fileLink = within(rightPane).getByRole('link', { name: 'diff/diff.go' })
        expect(fileLink).toBeVisible()

        // Assert the code view is rendered, by doing a partial match against its content
        const codeView = within(rightPane).getByRole('table')
        expect(codeView).toHaveTextContent('package diff import')
    })
})

describe('getLineContent', () => {
    const testFileLines = [
        'package main',
        '',
        'import "fmt"',
        '',
        'type Animal interface {',
        '\tSound() string',
        '}',
        '',
        'var _ Animal = Cat{}',
        '',
        'type Cat struct{}',
        '',
        'func (c Cat) Sound() string { return "i am a cat" }',
        '',
        'type Dog struct{}',
        '',
        'func (d Dog) Sound() string { return "it is i, the dog" }',
        '',
        'func animalFarm() {',
        '\tmakeSound(Cat{})',
        '\tmakeSound(Dog{})',
        '}',
        '',
        'func makeSound(a Animal) {',
        '\tfmt.Printf("animal made a sound: %s", a.Sound())',
        '',
        '\tfoo := a',
        '',
        '\tfmt.Printf("another animal: %s", foo.Sound())',
        '}',
        '',
    ]

    it('returns pre/post and token of line', () => {
        const location: Location = {
            repo: 'a',
            file: 'file',
            commitID: 'f00b4r',
            url: 'url',
            precise: true,
            content: '',
            lines: testFileLines,
            range: {
                start: { line: 23, character: 5 },
                end: { line: 23, character: 14 },
            },
        }

        const content = getLineContent(location)
        expect(content.prePostToken?.pre).toEqual('func ')
        expect(content.prePostToken?.token).toEqual('makeSound')
        expect(content.prePostToken?.post).toEqual('(a Animal) {')
    })

    it('handles token on multiple lines', () => {
        const location: Location = {
            repo: 'a',
            file: 'file',
            commitID: 'f00b4r',
            url: 'url',
            precise: true,
            content: '',
            lines: ['line0 start-of-tok', 'en-ends-here line1'],
            range: {
                start: { line: 0, character: 6 },
                end: { line: 1, character: 12 },
            },
        }

        const content = getLineContent(location)
        expect(content.prePostToken?.pre).toEqual('line0 ')
        expect(content.prePostToken?.token).toEqual('start-of-tok')
        expect(content.prePostToken?.post).toEqual('')
    })
})
