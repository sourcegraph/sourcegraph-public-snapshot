import expect from 'expect'
import { test } from 'mocha'
import { Key } from 'ts-key-enum'

import { SearchGraphQlOperations } from '@sourcegraph/search'
import { SharedGraphQlOperations, SymbolKind } from '@sourcegraph/shared/src/graphql-operations'
import { SearchEvent } from '@sourcegraph/shared/src/search/stream'
import { accessibilityAudit } from '@sourcegraph/shared/src/testing/accessibility'
import { Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { WebGraphQlOperations } from '../graphql-operations'

import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'
import { percySnapshotWithVariants } from './utils'

const EDITOR_SELECTOR = '#monaco-query-input'
const EDITOR_INPUT_SELECTOR = `${EDITOR_SELECTOR} .cm-content`
const COMPLETION_SELECTOR = `${EDITOR_SELECTOR} .cm-tooltip-autocomplete`
const COMPLETION_LABEL_SELECTOR = `${COMPLETION_SELECTOR} .cm-completionLabel`

const mockDefaultStreamEvents: SearchEvent[] = [
    {
        type: 'matches',
        data: [{ type: 'repo', repository: 'github.com/Algorilla/manta-ray' }],
    },
    { type: 'progress', data: { matchCount: 30, durationMs: 103, skipped: [] } },
    {
        type: 'filters',
        data: [
            { label: 'archived:yes', value: 'archived:yes', count: 5, kind: 'generic', limitHit: true },
            { label: 'fork:yes', value: 'fork:yes', count: 46, kind: 'generic', limitHit: true },
            // Two repo filters to trigger the repository sidebar section
            {
                label: 'github.com/Algorilla/manta-ray',
                value: 'repo:^github\\.com/Algorilla/manta-ray$',
                count: 1,
                kind: 'repo',
                limitHit: true,
            },
            {
                label: 'github.com/Algorilla/manta-ray2',
                value: 'repo:^github\\.com/Algorilla/manta-ray2$',
                count: 1,
                kind: 'repo',
                limitHit: true,
            },
        ],
    },
    { type: 'done', data: {} },
]

const commonSearchGraphQLResults: Partial<WebGraphQlOperations & SharedGraphQlOperations & SearchGraphQlOperations> = {
    ...commonWebGraphQlResults,
    IsSearchContextAvailable: () => ({
        isSearchContextAvailable: true,
    }),
    ViewerSettings: () => ({
        viewerSettings: {
            __typename: 'SettingsCascade',
            subjects: [
                {
                    __typename: 'DefaultSettings',
                    settingsURL: null,
                    viewerCanAdminister: false,
                    latestSettings: {
                        id: 0,
                        contents: JSON.stringify({
                            experimentalFeatures: {
                                editor: 'codemirror6',
                            },
                        }),
                    },
                },
            ],
            final: JSON.stringify({}),
        },
    }),
}

describe('Search (CodeMirror)', () => {
    let driver: Driver
    before(async () => {
        driver = await createDriverForTest()
    })
    after(() => driver?.close())
    let testContext: WebIntegrationTestContext
    beforeEach(async function () {
        testContext = await createWebIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })
        testContext.overrideGraphQL({
            ...commonSearchGraphQLResults,
            UserAreaUserProfile: () => ({
                user: {
                    __typename: 'User',
                    id: 'user123',
                    username: 'alice',
                    displayName: 'alice',
                    url: '/users/test',
                    settingsURL: '/users/test/settings',
                    avatarURL: '',
                    viewerCanAdminister: true,
                    builtinAuth: true,
                    tags: [],
                },
            }),
        })
    })
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    // The selector has to be a string literal because the function is evaluated
    // in a different environment
    const getSearchFieldValue = (driver: Driver): Promise<string | null | undefined> =>
        driver.page.evaluate(
            () => document.querySelector<HTMLDivElement>('#monaco-query-input .cm-content')?.textContent
        )

    describe('Filter completion', () => {
        test('Completing a negated filter should insert the filter with - prefix', async () => {
            testContext.overrideSearchStreamEvents(mockDefaultStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await driver.page.waitForSelector(EDITOR_SELECTOR)
            await driver.replaceText({
                selector: EDITOR_SELECTOR,
                newText: '-file',
                enterTextMethod: 'type',
            })
            await driver.findElementWithText('-file', {
                action: 'click',
                wait: { timeout: 5000 },
                selector: COMPLETION_LABEL_SELECTOR,
            })
            expect(await getSearchFieldValue(driver)).toStrictEqual('-file:')
            await percySnapshotWithVariants(driver.page, 'Search home page')
            await accessibilityAudit(driver.page)
        })
    })

    describe('Suggestions', () => {
        test('Typing in the search field shows relevant suggestions', async () => {
            testContext.overrideSearchStreamEvents([
                {
                    type: 'matches',
                    data: [
                        { type: 'repo', repository: 'github.com/auth0/go-jwt-middleware' },
                        {
                            type: 'symbol',
                            symbols: [
                                {
                                    name: 'OnError',
                                    containerName: 'jwtmiddleware',
                                    url: '/github.com/auth0/go-jwt-middleware/-/blob/jwtmiddleware.go#L56:1-56:14',
                                    kind: SymbolKind.FUNCTION,
                                },
                            ],
                            path: 'jwtmiddleware.go',
                            repository: 'github.com/auth0/go-jwt-middleware',
                        },
                        { type: 'path', path: 'jwtmiddleware.go', repository: 'github.com/auth0/go-jwt-middleware' },
                    ],
                },

                { type: 'done', data: {} },
            ])

            // Repo autocomplete from homepage
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            // Using id selector rather than `test-` classes as Monaco doesn't allow customizing classes
            await driver.page.waitForSelector(EDITOR_SELECTOR)
            await driver.replaceText({
                selector: EDITOR_SELECTOR,
                newText: 'repo:go-jwt-middlew',
                enterTextMethod: 'type',
            })
            await driver.findElementWithText('github.com/auth0/go-jwt-middleware', {
                action: 'click',
                wait: { timeout: 5000 },
                selector: COMPLETION_LABEL_SELECTOR,
            })
            expect(await getSearchFieldValue(driver)).toStrictEqual('repo:^github\\.com/auth0/go-jwt-middleware$ ')

            // Submit search
            await driver.page.keyboard.press(Key.Enter)

            // File autocomplete from repo search bar
            await driver.page.waitForSelector(EDITOR_SELECTOR)
            await driver.page.focus(EDITOR_SELECTOR)
            await driver.page.keyboard.type('file:jwtmi')
            await driver.findElementWithText('jwtmiddleware.go', {
                selector: COMPLETION_LABEL_SELECTOR,
                wait: { timeout: 5000 },
            })
            // This timeout seems to be necessary for Tab to select the entry
            await driver.page.waitForTimeout(100)
            // NOTE: This test assumes that that the the suggestions popover shows a
            // single entry only (since the first entry is selected by default).
            // It doesn't seem to be possible to otherwise "select" a specific
            // entry from the list (other than simulating arrow key presses and
            // somehow comparing the selected entry to the expected one).
            await driver.page.keyboard.press(Key.Tab)
            expect(await getSearchFieldValue(driver)).toStrictEqual(
                'repo:^github\\.com/auth0/go-jwt-middleware$ file:^jwtmiddleware\\.go$ '
            )

            // Symbol autocomplete in top search bar
            await driver.page.keyboard.type('On')
            await driver.findElementWithText('OnError', {
                selector: COMPLETION_LABEL_SELECTOR,
                wait: { timeout: 5000 },
            })
        })
    })

    describe('Search field value', () => {
        test('Is set from the URL query parameter when loading a search-related page', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
                RegistryExtensions: () => ({
                    extensionRegistry: {
                        __typename: 'ExtensionRegistry',
                        extensions: { error: null, nodes: [] },
                        featuredExtensions: null,
                    },
                }),
            })
            testContext.overrideSearchStreamEvents(mockDefaultStreamEvents)

            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=foo')
            await driver.page.waitForSelector(EDITOR_INPUT_SELECTOR)
            expect(await getSearchFieldValue(driver)).toStrictEqual('foo')
            // Field value is cleared when navigating to a non search-related page
            await driver.page.waitForSelector('a[href="/extensions"]')
            await driver.page.click('a[href="/extensions"]')
            // Search box is gone when in a non-search page
            expect(await getSearchFieldValue(driver)).toStrictEqual(undefined)
            // Field value is restored when the back button is pressed
            await driver.page.goBack()
            expect(await getSearchFieldValue(driver)).toStrictEqual('foo')
        })
    })
})
