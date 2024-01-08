import type { EditorView } from '@codemirror/view'
import { merge } from 'lodash'
import type { Page } from 'puppeteer'

import type { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import type { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { type Driver, percySnapshot } from '@sourcegraph/shared/src/testing/driver'
import { readEnvironmentBoolean } from '@sourcegraph/shared/src/testing/utils'

import type { WebGraphQlOperations } from '../graphql-operations'

const CODE_HIGHLIGHTING_QUERIES: Partial<keyof (WebGraphQlOperations & SharedGraphQlOperations)>[] = [
    'highlightCode',
    'Blob',
    'HighlightedFile',
]

/**
 * Matches a URL against an expected query that will handle code highlighting.
 */
const isCodeHighlightQuery = (url: string): boolean =>
    CODE_HIGHLIGHTING_QUERIES.some(query => url.includes(`graphql?${query}`))

/**
 * Watches and waits for requests to complete that match expected code highlighting queries.
 * Useful to ensure styles are fully loaded after changing the webapp color scheme.
 */
const waitForCodeHighlighting = async (page: Page): Promise<void> => {
    const requestDidFire = await page
        .waitForRequest(request => isCodeHighlightQuery(request.url()), { timeout: 1000 })
        .catch(
            () =>
                // request won't always fire if data is cached
                false
        )

    if (requestDidFire) {
        await page.waitForResponse(request => isCodeHighlightQuery(request.url()), { timeout: 1000 })
    }
}

type ColorScheme = 'dark' | 'light'

/**
 * Percy couldn't capture <img /> since they have `src` values with testing domain name.
 * We need to call this function before asking Percy to take snapshots,
 * <img /> with base64 data would be visible on Percy snapshot
 */
export const convertImgSourceHttpToBase64 = async (page: Page): Promise<void> => {
    await page.evaluate(() => {
        // Skip images with data-skip-percy
        // Skip images with .cm-widgetBuffer, which CodeMirror uses when using a widget decoration
        // See https://github.com/sourcegraph/sourcegraph/issues/28949
        const imgs = document.querySelectorAll<HTMLImageElement>('img:not([data-skip-percy]):not(.cm-widgetBuffer)')

        for (const img of imgs) {
            if (img.src.startsWith('data:image')) {
                continue
            }

            const canvas = document.createElement('canvas')
            canvas.width = img.width
            canvas.height = img.height

            const context = canvas.getContext('2d')
            context?.drawImage(img, 0, 0)

            img.src = canvas.toDataURL('image/png')
        }
    })
}

/**
 * Update all theme styling on the Sourcegraph webapp to match a color scheme.
 */
export const setColorScheme = async (
    page: Page,
    scheme: ColorScheme,
    shouldWaitForCodeHighlighting?: boolean
): Promise<void> => {
    const isAlreadySet = await page.evaluate(
        (scheme: ColorScheme) => matchMedia(`(prefers-color-scheme: ${scheme})`).matches,
        scheme
    )

    if (isAlreadySet) {
        return
    }

    await Promise.all([
        page.emulateMediaFeatures([{ name: 'prefers-color-scheme', value: scheme }]),
        shouldWaitForCodeHighlighting ? waitForCodeHighlighting(page) : Promise.resolve(),
    ])
}

export interface PercySnapshotConfig {
    /**
     * How long to wait for the UI to settle before taking a screenshot.
     */
    timeout: number
    waitForCodeHighlighting: boolean
}

/**
 * Takes a Percy snapshot in 2 variants: dark/light
 */
export const percySnapshotWithVariants = async (
    page: Page,
    name: string,
    { timeout = 1000, waitForCodeHighlighting = false } = {}
): Promise<void> => {
    const percyEnabled = readEnvironmentBoolean({ variable: 'PERCY_ON', defaultValue: false })

    if (!percyEnabled) {
        return
    }

    // Theme-dark
    await setColorScheme(page, 'dark', waitForCodeHighlighting)
    // Wait for the theme class set by `useLayoutEffect` in `client/web/src/LegacyLayout.tsx`
    await page.waitForSelector('html.theme.theme-dark')
    // Wait for the UI to settle before converting images and taking the
    // screenshot.
    await page.waitForTimeout(timeout)
    await convertImgSourceHttpToBase64(page)
    await percySnapshot(page, `${name} - dark theme`)

    // Theme-light
    await setColorScheme(page, 'light', waitForCodeHighlighting)
    // Wait for the theme class set by `useLayoutEffect` in `client/web/src/LegacyLayout.tsx`
    await page.waitForSelector('html.theme.theme-light')
    // Wait for the UI to settle before converting images and taking the
    // screenshot.
    await page.waitForTimeout(timeout)
    await convertImgSourceHttpToBase64(page)
    await percySnapshot(page, `${name} - light theme`)
}

type Editor = 'monaco' | 'codemirror6' | 'v2'

export interface EditorAPI {
    name: Editor
    /**
     * Wait for editor to appear in the DOM.
     */
    waitForIt: (options?: Parameters<Page['waitForSelector']>[1]) => Promise<void>
    /**
     * Wait for suggestion with provided label appears
     */
    waitForSuggestion: (label?: string) => Promise<void>
    /**
     * Moves focus to the editor's root node.
     */
    focus: () => Promise<void>
    /**
     * Returns the current value of the editor.
     */
    getValue: () => Promise<string | null | undefined>
    /**
     * Replaces the editor's content with the provided input.
     */
    replace: (input: string, method?: 'type' | 'paste') => Promise<void>
    /**
     * Append the provided input to the editor's content
     */
    append: (input: string, method?: 'type' | 'paste') => Promise<void>
    /**
     * Triggers application of the specified suggestion.
     */
    selectSuggestion: (label: string) => Promise<void>
}

const editors: Record<Editor, (driver: Driver, rootSelector: string) => EditorAPI> = {
    monaco: (driver: Driver, rootSelector: string) => {
        const inputSelector = `${rootSelector} textarea`
        // Selector to use to wait for the editor to be complete loaded
        const readySelector = `${rootSelector} .view-lines`
        const completionSelector = `${rootSelector} .suggest-widget.visible`
        const completionLabelSelector = `${completionSelector} span`

        const api: EditorAPI = {
            name: 'monaco',
            async waitForIt(options) {
                await driver.page.waitForSelector(readySelector, options)
            },
            async focus() {
                await api.waitForIt()
                await driver.page.click(rootSelector)
            },
            getValue() {
                return driver.page.evaluate(
                    (inputSelector: string) => document.querySelector<HTMLTextAreaElement>(inputSelector)?.value,
                    inputSelector
                )
            },
            replace(newText: string, method = 'type') {
                return driver.replaceText({
                    selector: rootSelector,
                    newText,
                    enterTextMethod: method,
                    selectMethod: 'keyboard',
                })
            },
            async append(newText: string, method = 'type') {
                await api.focus()
                return driver.enterText(method, newText)
            },
            async waitForSuggestion(suggestion?: string) {
                await driver.page.waitForSelector(completionSelector)
                if (suggestion !== undefined) {
                    await driver.findElementWithText(suggestion, {
                        selector: completionLabelSelector,
                        wait: { timeout: 5000 },
                    })
                }
            },
            async selectSuggestion(suggestion: string) {
                await driver.page.waitForSelector(completionSelector)
                await driver.findElementWithText(suggestion, {
                    action: 'click',
                    selector: completionLabelSelector,
                    wait: { timeout: 5000 },
                })
            },
        }
        return api
    },
    codemirror6: (driver: Driver, rootSelector: string) => {
        // Selector to use to wait for the editor to be complete loaded
        const readySelector = `${rootSelector} .cm-line`
        const completionSelector = `${rootSelector} .cm-tooltip-autocomplete`
        const completionLabelSelector = `${completionSelector} .cm-completionLabel`

        const api: EditorAPI = {
            name: 'codemirror6',
            async waitForIt(options) {
                await driver.page.waitForSelector(readySelector, options)
            },
            async focus() {
                await api.waitForIt()
                await driver.page.click(readySelector)
            },
            getValue() {
                return driver.page.evaluate((selector: string) => {
                    // Typecast "as any" is used to avoid TypeScript complaining
                    // about window not having this property. We decided that
                    // it's fine to use this in a test context
                    // eslint-disable-next-line @typescript-eslint/no-unsafe-member-access,@typescript-eslint/no-explicit-any
                    const fromDOM = (window as any).CodeMirrorFindFromDOM as
                        | typeof EditorView['findFromDOM']
                        | undefined
                    if (!fromDOM) {
                        throw new Error(
                            'CodeMirror DOM API not exposed. Ensure the web app is built with INTEGRATION_TESTS=true.'
                        )
                    }
                    const editorElement = document.querySelector<HTMLElement>(selector)
                    if (editorElement) {
                        // Returns an EditorView
                        // See https://codemirror.net/docs/ref/#view.EditorView^findFromDOM
                        return fromDOM(editorElement)?.state.sliceDoc()
                    }
                    return undefined
                }, rootSelector)
            },
            replace(newText: string, method = 'type') {
                return driver.replaceText({
                    selector: rootSelector,
                    newText,
                    enterTextMethod: method,
                    selectMethod: 'selectall',
                })
            },
            async append(newText: string, method = 'type') {
                await api.focus()
                return driver.enterText(method, newText)
            },
            async waitForSuggestion(suggestion?: string) {
                await driver.page.waitForSelector(completionSelector)
                if (suggestion !== undefined) {
                    await driver.findElementWithText(suggestion, {
                        selector: completionLabelSelector,
                        wait: { timeout: 5000 },
                    })
                }
                // It seems CodeMirror needs some additional time before it
                // recognizes events on the suggestions element (such as
                // selecting a suggestion via the Tab key)
                await driver.page.waitForTimeout(100)
            },
            async selectSuggestion(suggestion: string) {
                await driver.page.waitForSelector(completionSelector)
                await driver.findElementWithText(suggestion, {
                    action: 'click',
                    selector: completionLabelSelector,
                    wait: { timeout: 5000 },
                })
            },
        }
        return api
    },
    v2: (driver: Driver, rootSelector: string) => {
        // Selector to use to wait for the editor to be complete loaded
        const completionSelector = `${rootSelector} [role="grid"]`
        const completionLabelSelector = `${completionSelector} .test-option-label`

        const api = {
            ...editors.codemirror6(driver, `${rootSelector} .test-query-input`),
            async waitForSuggestion(suggestion?: string) {
                await driver.page.waitForSelector(completionSelector)
                if (suggestion !== undefined) {
                    await driver.findElementWithText(suggestion, {
                        selector: completionLabelSelector,
                        wait: { timeout: 5000 },
                    })
                }
                // It seems CodeMirror needs some additional time before it
                // recognizes events on the suggestions element (such as
                // selecting a suggestion via the Tab key)
                await driver.page.waitForTimeout(100)
            },
            async selectSuggestion(suggestion: string) {
                await driver.page.waitForSelector(completionSelector)
                await driver.findElementWithText(suggestion, {
                    action: 'click',
                    selector: completionLabelSelector,
                    wait: { timeout: 5000 },
                })
            },
        }
        return api
    },
}

/**
 * Returns an object for accessing editor related information at `rootSelector`.
 * It also waits for the editor to be ready
 */
export const createEditorAPI = async (driver: Driver, rootSelector: string): Promise<EditorAPI> => {
    // We append `.test-editor` to make sure to wait for the actual editor
    // component to load and not e.g. target the placeholder input used by
    // LazyMonacoSearchQueryInput
    await driver.page.waitForSelector(`${rootSelector}.test-editor`)
    const editor = await driver.page.evaluate<(selector: string) => string | undefined>(
        selector => (document.querySelector(selector) as HTMLElement).dataset.editor,
        rootSelector
    )
    if (!editor) {
        throw new Error("Can't determine editor, data-editor=... is not set.")
    }
    if (!Object.hasOwn(editors, editor)) {
        throw new Error(`${editor} is not a supported editor`)
    }
    const api = editors[editor as Editor](driver, rootSelector)
    await api.waitForIt()
    return api
}

export type SearchQueryInput = Extract<Editor, 'codemirror6' | 'v2'>
interface SearchQueryInputAPI {
    /**
     * The name of the currently used query input implementation. Can be used to dynamically generate
     * test names.
     */
    name: SearchQueryInput
    waitForInput: (driver: Driver, selector: string) => Promise<EditorAPI>
    applySettings: (settings?: Settings) => Settings
}

const searchInputNames: SearchQueryInput[] = ['codemirror6', 'v2']

const searchInputConfigs: Record<SearchQueryInput, SearchQueryInputAPI> = {
    codemirror6: {
        name: 'codemirror6',
        waitForInput: (driver: Driver, rootSelector: string) => createEditorAPI(driver, rootSelector),
        applySettings: (settings = {}) =>
            merge(settings, { experimentalFeatures: { searchQueryInput: 'v1' } } satisfies Settings),
    },
    v2: {
        name: 'v2',
        waitForInput: (driver: Driver, rootSelector: string) => createEditorAPI(driver, rootSelector),
        applySettings: (settings = {}) =>
            merge(settings, { experimentalFeatures: { searchQueryInput: 'v2' } } satisfies Settings),
    },
}

export const getSearchQueryInputConfig = (input: SearchQueryInput): SearchQueryInputAPI => searchInputConfigs[input]

/**
 * Helper function for abstracting away testing different search query input
 * implementations. The callback function gets passed an object to interact with
 * the input and to configure the necessary settings.
 */
export const withSearchQueryInput = (callback: (config: SearchQueryInputAPI) => void): void => {
    for (const input of searchInputNames) {
        // This callback is supposed to be called multiple times
        // eslint-disable-next-line callback-return
        callback(getSearchQueryInputConfig(input))
    }
}

/**
 * This helper function removes any context:... filter in the query (via regular expression)
 * to make it easier to compare query inputs when the context doesn't amtter.
 */
export const removeContextFromQuery = (input: string): string => input.replace(/\s*context:\S*\s*/, '')

export const isElementDisabled = (driver: Driver, query: string): Promise<boolean> =>
    driver.page.evaluate((query: string) => {
        const element = document.querySelector<HTMLButtonElement>(query)

        const disabledAttribute = element!.disabled
        const ariaDisabled = element!.getAttribute('aria-disabled')

        return disabledAttribute || ariaDisabled === 'true'
    }, query)
