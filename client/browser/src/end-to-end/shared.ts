import assert from 'assert'

import expect from 'expect'
import { describe, it } from 'mocha'
import type puppeteer from 'puppeteer'

import type { Driver } from '@sourcegraph/shared/src/testing/driver'
import { retry } from '@sourcegraph/shared/src/testing/utils'

/**
 * Defines e2e tests for a single-file page of a code host.
 */
export function testSingleFilePage({
    getDriver,
    url,
    sourcegraphBaseUrl,
    repoName,
    commitID,
    getLineSelector,
    goToDefinitionURL,
}: {
    /** Called to get the driver */
    getDriver: () => Driver

    /** The URL to sourcegraph/jsonrpc2 call_opt.go at commit {@link commitID} on the code host */
    url: string

    /** The base URL of the sourcegraph instance */
    sourcegraphBaseUrl: string

    /** The repo name of sourcgraph/jsonrpc2 on the Sourcegraph instance */
    repoName: string

    /** The full commit SHA of sourcegraph/jsonrpc2 call_opt.go on the code host */
    commitID: string

    /** The CSS selector for a line (with or without line number part) in the code view */
    getLineSelector?: (lineNumber: number) => string

    /** The expected URL for the "Go to Definition" button */
    goToDefinitionURL?: string
}): void {
    describe('File views', () => {
        it('adds "View on Sourcegraph" buttons to files', async () => {
            await getDriver().page.goto(url)

            // Make sure the tab is active, because it might not be active if the install page has opened.
            await closeInstallPageTab(getDriver().page.browser())

            await getDriver().page.waitForSelector(
                '[data-testid="code-view-toolbar"] [data-testid="open-on-sourcegraph"]',
                { timeout: 10000 }
            )
            expect(
                await getDriver().page.$$('[data-testid="code-view-toolbar"] [data-testid="open-on-sourcegraph"]')
            ).toHaveLength(1)

            await retry(async () => {
                assert.strictEqual(
                    await getDriver().page.evaluate(
                        () =>
                            document.querySelector<HTMLAnchorElement>(
                                '[data-testid="code-view-toolbar"] [data-testid="open-on-sourcegraph"]'
                            )?.href
                    ),
                    new URL(`${sourcegraphBaseUrl}/${repoName}@${commitID}/-/blob/call_opt.go`).href
                )
            })
        })

        if (getLineSelector) {
            it('shows hover tooltips when hovering a token', async () => {
                await getDriver().page.goto(url)
                await getDriver().page.waitForSelector(
                    '[data-testid="code-view-toolbar"] [data-testid="open-on-sourcegraph"]'
                )

                // Trigger tokenization of the line.
                const lineNumber = 16
                const line = await getDriver().page.waitForSelector(getLineSelector(lineNumber), {
                    timeout: 10000,
                })

                if (!line) {
                    throw new Error(`Found no line with number ${lineNumber}`)
                }

                // Hover line to give codeintellify time to register listeners for
                // tokenization (only necessary in CI, not sure why).
                await line.hover()

                const [token] = await line.$x('.//span[text()="CallOption"]')
                await token.hover()
                await getDriver().page.waitForSelector('.test-tooltip-go-to-definition')
                await getDriver().page.waitForSelector('.test-tooltip-content')
                await retry(async () => {
                    assert.strictEqual(
                        await getDriver().page.evaluate(
                            () => document.querySelector<HTMLAnchorElement>('.test-tooltip-go-to-definition')?.href
                        ),
                        goToDefinitionURL || `${sourcegraphBaseUrl}/${repoName}@${commitID}/-/blob/call_opt.go#L5:6`
                    )
                })
            })
        }
    })
}

/**
 * Find a tab that contains the browser extension's after-install page (url
 * ending in `/after_install.html`) and, if found, close it.
 *
 * The after-install page is opened automatically when the browser extension is
 * installed. In tests, this means that it's opened automatically every time we
 * start the browser (with the browser extension loaded).
 */
export async function closeInstallPageTab(browser: puppeteer.Browser): Promise<void> {
    const pages = await browser.pages()
    const installPage = pages.find(page => page.url().endsWith('/after_install.html'))
    if (installPage) {
        await installPage.close()
    }
}
