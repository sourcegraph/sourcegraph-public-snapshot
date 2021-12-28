import { Page } from 'puppeteer'

import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { percySnapshot } from '@sourcegraph/shared/src/testing/driver'
import { readEnvironmentBoolean } from '@sourcegraph/shared/src/testing/utils'

import { WebGraphQlOperations } from '../graphql-operations'

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

const ColorSchemeToMonacoEditorClassName: Record<ColorScheme, string> = {
    dark: 'vs-dark',
    light: 'vs',
}
/**
 * Percy couldn't capture <img /> since they have `src` values with testing domain name.
 * We need to call this function before asking Percy to take snapshots,
 * <img /> with base64 data would be visible on Percy snapshot
 */
export const convertImgSourceHttpToBase64 = async (page: Page): Promise<void> => {
    await page.evaluate(() => {
        // Skip images with data-skip-percy
        // See https://github.com/sourcegraph/sourcegraph/issues/28949
        const imgs = document.querySelectorAll<HTMLImageElement>('img:not([data-skip-percy])')

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

    try {
        // Check Monaco editor is styled correctly
        await page.waitForFunction(
            (expectedClassName: string) =>
                document.querySelector('#monaco-query-input .monaco-editor') &&
                document.querySelector('#monaco-query-input .monaco-editor')?.classList.contains(expectedClassName),
            { timeout: 1000 },
            ColorSchemeToMonacoEditorClassName[scheme]
        )

        // Wait a tiny bit for Monaco syntax highlighting to be applied
        await page.waitForTimeout(500)
    } catch {
        // noop, page doesn't use monaco editor
    }
}

export interface PercySnapshotConfig {
    waitForCodeHighlighting: boolean
}

/**
 * Takes a Percy snapshot in 2 variants: dark/light
 */
export const percySnapshotWithVariants = async (
    page: Page,
    name: string,
    config?: PercySnapshotConfig
): Promise<void> => {
    const percyEnabled = readEnvironmentBoolean({ variable: 'PERCY_ON', defaultValue: false })

    if (!percyEnabled) {
        return
    }

    try {
        // Wait for Monaco editor to finish rendering before taking screenshots
        await page.waitForSelector('#monaco-query-input .monaco-editor', { visible: true, timeout: 1000 })
    } catch {
        // noop, page doesn't use monaco editor
    }

    // Theme-light
    await setColorScheme(page, 'light', config?.waitForCodeHighlighting)
    await convertImgSourceHttpToBase64(page)
    await percySnapshot(page, `${name} - light theme`)

    // Theme-dark
    await setColorScheme(page, 'dark', config?.waitForCodeHighlighting)
    await convertImgSourceHttpToBase64(page)
    await percySnapshot(page, `${name} - dark theme`)

    // Reset to light theme
    await setColorScheme(page, 'light', config?.waitForCodeHighlighting)
}
