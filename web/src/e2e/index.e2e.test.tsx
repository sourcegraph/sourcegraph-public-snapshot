import * as os from 'os'
import * as path from 'path'
import puppeteer from 'puppeteer'
import { Key } from 'ts-key-enum'
import { saveScreenshotsUponFailuresAndClosePage } from '../../../shared/src/util/screenshotReporter'
import { readEnvBoolean, readEnvString, retry } from '../util/e2e-test-utils'

jest.setTimeout(30000)

/**
 * Used in the external service configuration.
 */
export const gitHubToken = readEnvString({ variable: 'GITHUB_TOKEN' })

describe('e2e test suite', function(this: any): void {
    const baseURL = readEnvString({ variable: 'SOURCEGRAPH_BASE_URL', defaultValue: 'http://localhost:3080' })

    let browser: puppeteer.Browser
    let page: puppeteer.Page

    async function init(): Promise<void> {
        page = await browser.newPage()
        await ensureLoggedIn(page)
        await ensureHasExternalService(page)
    }

    // Start browser.
    beforeAll(async () => {
        let args: string[] | undefined
        if (process.getuid() === 0) {
            // TODO don't run as root in CI
            console.warn('Running as root, disabling sandbox')
            args = ['--no-sandbox', '--disable-setuid-sandbox']
        }
        browser = await puppeteer.launch({
            args,
            headless: readEnvBoolean({ variable: 'HEADLESS', defaultValue: false }),
        })
        await init()
    })

    // Close browser.
    afterAll(async () => {
        if (browser) {
            if (page && !page.isClosed()) {
                await page.close()
            }
            await browser.close()
        }
    })

    async function ensureLoggedIn(page: puppeteer.Page): Promise<void> {
        await page.goto(baseURL)
        const url = new URL(await page.url())
        if (url.pathname === '/site-admin/init') {
            await page.type('input[name=email]', 'test@test.com')
            await page.type('input[name=username]', 'test')
            await page.type('input[name=password]', 'test')
            await page.click('button[type=submit]')
            await page.waitForNavigation()
        } else if (url.pathname === '/sign-in') {
            await page.type('input', 'test')
            await page.type('input[name=password]', 'test')
            await page.click('button[type=submit]')
            await page.waitForNavigation()
        }
    }

    async function ensureHasExternalService(page: puppeteer.Page): Promise<void> {
        await page.goto(baseURL + '/site-admin/external-services')
        await page.waitFor('.filtered-connection')
        await page.waitFor(() => !document.querySelector('.filtered-connection__loader'))
        // Matches buttons for deleting external services named 'test-github'.
        const deleteButtons = await page.$x(
            "//*[contains(@class, 'external-service-node') and div/div[text()='test-github']]//*[contains(@class,'btn-danger')]"
        )
        if (deleteButtons.length > 0) {
            const accept = async (dialog: puppeteer.Dialog) => {
                await dialog.accept()
                page.off('dialog', accept)
            }
            page.on('dialog', accept)
            await deleteButtons[0].click()
        }
        await (await page.waitForXPath("//*[contains(text(), 'Add external service')]")).click()
        await (await page.waitForSelector('#external-service-form-display-name')).type('test-github')

        const editor = await page.waitForSelector('.view-line')
        await editor.click()
        const modifier = os.platform() === 'darwin' ? Key.Meta : Key.Control
        await page.keyboard.down(modifier)
        await page.keyboard.press('KeyA')
        await page.keyboard.up(modifier)
        await page.keyboard.press(Key.Backspace)
        await page.keyboard.type(
            JSON.stringify({
                url: 'https://github.com',
                token: gitHubToken,
                repos: [
                    'gorilla/mux',
                    'gorilla/securecookie',
                    'sourcegraphtest/AlwaysCloningTest',
                    'sourcegraph/godockerize',
                    'sourcegraph/jsonrpc2',
                    'sourcegraph/checkup',
                    'sourcegraph/go-diff',
                    'sourcegraph/go-vcs',
                ],
            })
        )
        await (await page.$x("//form/*[contains(text(), 'Add external service')]"))[0].click()
        // TODO figure out how to refresh the repositories
        // https://sourcegraph.slack.com/archives/C07KZF47K/p1551516553020000
        function delay(timeout: number): Promise<void> {
            return new Promise(resolve => {
                setTimeout(resolve, timeout)
            })
        }
        await delay(10000)
    }

    // Open page.
    beforeEach(async () => {
        page = await browser.newPage()
    })

    // Take a screenshot when a test fails.
    saveScreenshotsUponFailuresAndClosePage(
        path.resolve(__dirname, '..', '..', '..'),
        path.resolve(__dirname, '..', '..', '..', 'puppeteer'),
        () => page
    )

    const enableOrAddRepositoryIfNeeded = async (): Promise<any> => {
        // Disable any toasts, which can interfere with clicking on the enable/add button.
        try {
            // The toast should appear fast, but not instant, so wait and use a short timeout
            await page.waitForSelector('.toast__close-button', { timeout: 1000 })
            await page.click('.toast__close-button')
        } catch (e) {
            // Probably no toast was showing.
        }
        // Wait for the repository container or a repository error page to be shown.
        await Promise.race([
            // Repository is already enabled and added; nothing to do.
            page.waitForSelector('.repo-rev-container'),

            // Add or enable repository.
            (async () => {
                try {
                    await page.waitForSelector('.repository-error-page__btn:not([disabled])')
                } catch {
                    return
                }
                await page.click('.repository-error-page__btn:not([disabled])')
                await page.waitForSelector('.repo-rev-container')
            })(),

            // Repository is cloning.
            page.waitForSelector('.repository-cloning-in-progress-page'),
        ])
    }

    const assertWindowLocation = async (location: string, isAbsolute = false): Promise<any> => {
        const url = isAbsolute ? location : baseURL + location
        await retry(async () => {
            expect(await page.evaluate(() => window.location.href)).toEqual(url)
        })
    }

    const assertWindowLocationPrefix = async (locationPrefix: string, isAbsolute = false): Promise<any> => {
        const prefix = isAbsolute ? locationPrefix : baseURL + locationPrefix
        await retry(async () => {
            const loc: string = await page.evaluate(() => window.location.href)
            expect(loc.startsWith(prefix)).toBeTruthy()
        })
    }

    const assertStickyHighlightedToken = async (label: string): Promise<void> => {
        await page.waitForSelector('.selection-highlight-sticky') // make sure matched token is highlighted
        await retry(async () =>
            expect(
                await page.evaluate(() => document.querySelector('.selection-highlight-sticky')!.textContent)
            ).toEqual(label)
        )
    }

    const assertAllHighlightedTokens = async (label: string): Promise<void> => {
        const highlightedTokens = await page.evaluate(() =>
            Array.from(document.querySelectorAll('.selection-highlight')).map(el => el.textContent || '')
        )
        expect(highlightedTokens.every(txt => txt === label)).toBeTruthy()
    }

    const assertNonemptyRefs = async (): Promise<void> => {
        // verify active group is references
        await page.waitForXPath("//*[contains(@class, 'panel__tabs')]//*[contains(@class, 'tab-bar__tab--active')]")
        // verify there are some references
        await page.waitForXPath("//*[contains(@class, 'panel__tabs-content')]//*[contains(@class, 'file-match__list')]")
    }

    describe('Theme switcher', () => {
        test('changes the theme', async () => {
            await page.goto(baseURL + '/github.com/gorilla/mux/-/blob/mux.go')
            await enableOrAddRepositoryIfNeeded()
            await page.waitForSelector('.theme')
            const currentThemes = await page.evaluate(() =>
                Array.from(document.querySelector('.theme')!.classList).filter(c => c.startsWith('theme-'))
            )
            expect(currentThemes).toHaveLength(1)
            await page.click('.e2e-user-nav-item-toggle')
            await page.select('.e2e-theme-toggle', 'dark')
            expect(
                await page.evaluate(() =>
                    Array.from(document.querySelector('.theme')!.classList).filter(c => c.startsWith('theme-'))
                )
            ).toEqual(['theme-dark'])
            await page.select('.e2e-theme-toggle', 'light')
            expect(
                await page.evaluate(() =>
                    Array.from(document.querySelector('.theme')!.classList).filter(c => c.startsWith('theme-'))
                )
            ).toEqual(['theme-light'])
        })
    })

    describe('Repository component', () => {
        const blobTableSelector = '.e2e-blob > table'
        /**
         * @param line 1-indexed line number
         * @param spanOffset 1-indexed index of the span that's to be clicked
         */
        const clickToken = async (line: number, spanOffset: number): Promise<void> => {
            const selector = `${blobTableSelector} tr:nth-child(${line}) > td.code > span:nth-child(${spanOffset})`
            await page.waitForSelector(selector, { visible: true })
            await page.click(selector)
        }

        // expectedCount defaults to one because of we haven't specified, we just want to ensure it exists at all
        const getHoverContents = async (expectedCount = 1): Promise<string[]> => {
            const selector =
                expectedCount > 1 ? `.e2e-tooltip-content:nth-child(${expectedCount})` : `.e2e-tooltip-content`
            await page.waitForSelector(selector, { visible: true })
            return await page.evaluate(() =>
                // You can't reference hoverContentSelector in puppeteer's page.evaluate
                Array.from(document.querySelectorAll('.e2e-tooltip-content')).map(t => t.textContent || '')
            )
        }
        const assertHoverContentContains = async (val: string, count?: number) => {
            expect(await getHoverContents(count)).toContain(val)
        }

        const clickHoverJ2D = async (): Promise<void> => {
            const selector = '.e2e-tooltip-j2d'
            await page.waitForSelector(selector, { visible: true })
            await page.click(selector)
        }
        const clickHoverFindRefs = async (): Promise<void> => {
            const selector = '.e2e-tooltip-find-refs'
            await page.waitForSelector(selector, { visible: true })
            await page.click(selector)
        }

        describe('file tree', () => {
            test('does navigation on file click', async () => {
                await page.goto(
                    baseURL + '/github.com/sourcegraph/godockerize@05bac79edd17c0f55127871fa9c6f4d91bebf07c'
                )
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector(`[data-tree-path="godockerize.go"]`)
                await page.click(`[data-tree-path="godockerize.go"]`)
                await assertWindowLocation(
                    '/github.com/sourcegraph/godockerize@05bac79edd17c0f55127871fa9c6f4d91bebf07c/-/blob/godockerize.go'
                )
            })

            test('expands directory on row click (no navigation)', async () => {
                await page.goto(baseURL + '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d')
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.tree__row-icon')
                await page.click('.tree__row-icon')
                await page.waitForSelector('.tree__row--selected [data-tree-path="websocket"]')
                await page.waitForSelector('.tree__row--expanded [data-tree-path="websocket"]')
                await assertWindowLocation('/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d')
            })

            test('does navigation on directory row click', async () => {
                await page.goto(baseURL + '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d')
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.tree__row-label')
                await page.click('.tree__row-label')
                await page.waitForSelector('.tree__row--selected [data-tree-path="websocket"]')
                await page.waitForSelector('.tree__row--expanded [data-tree-path="websocket"]')
                await assertWindowLocation(
                    '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d/-/tree/websocket'
                )
            })

            test('selects the current file', async () => {
                await page.goto(
                    baseURL +
                        '/github.com/sourcegraph/godockerize@05bac79edd17c0f55127871fa9c6f4d91bebf07c/-/blob/godockerize.go'
                )
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.tree__row--active [data-tree-path="godockerize.go"]')
            })

            test('shows partial tree when opening directory', async () => {
                await page.goto(
                    baseURL +
                        '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d/-/tree/websocket'
                )
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.tree__row')
                expect(await page.evaluate(() => document.querySelectorAll('.tree__row').length)).toEqual(1)
            })

            test('responds to keyboard shortcuts', async () => {
                const assertNumRowsExpanded = async (expectedCount: number) => {
                    expect(await page.evaluate(() => document.querySelectorAll('.tree__row--expanded').length)).toEqual(
                        expectedCount
                    )
                }

                await page.goto(
                    baseURL +
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/.travis.yml'
                )
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.tree__row') // waitForSelector for tree to render

                await page.click('.tree')
                await page.keyboard.press('ArrowUp') // arrow up to 'diff' directory
                await page.waitForSelector('.tree__row--selected [data-tree-path="diff"]')
                await page.keyboard.press('ArrowRight') // arrow right (expand 'diff' directory)
                await page.waitForSelector('.tree__row--selected [data-tree-path="diff"]')
                await page.waitForSelector('.tree__row--expanded [data-tree-path="diff"]')
                await page.waitForSelector('.tree__row [data-tree-path="diff/testdata"]')
                await page.keyboard.press('ArrowRight') // arrow right (move to nested 'diff/testdata' directory)
                await page.waitForSelector('.tree__row--selected [data-tree-path="diff/testdata"]')
                await assertNumRowsExpanded(1) // only `diff` directory is expanded, though `diff/testdata` is expanded

                await page.keyboard.press('ArrowRight') // arrow right (expand 'diff/testdata' directory)
                await page.waitForSelector('.tree__row--selected [data-tree-path="diff/testdata"]')
                await page.waitForSelector('.tree__row--expanded [data-tree-path="diff/testdata"]')
                await assertNumRowsExpanded(2) // `diff` and `diff/testdata` directories expanded

                await page.waitForSelector('.tree__row [data-tree-path="diff/testdata/empty.diff"]')
                // select some file nested under `diff/testdata`
                await page.keyboard.press('ArrowDown') // arrow down
                await page.keyboard.press('ArrowDown') // arrow down
                await page.keyboard.press('ArrowDown') // arrow down
                await page.keyboard.press('ArrowDown') // arrow down
                await page.waitForSelector('.tree__row--selected [data-tree-path="diff/testdata/empty_orig.diff"]')

                await page.keyboard.press('ArrowLeft') // arrow left (navigate immediately up to parent directory `diff/testdata`)
                await page.waitForSelector('.tree__row--selected [data-tree-path="diff/testdata"]')
                await assertNumRowsExpanded(2) // `diff` and `diff/testdata` directories expanded

                await page.keyboard.press('ArrowLeft') // arrow left
                await page.waitForSelector('.tree__row--selected [data-tree-path="diff/testdata"]') // `diff/testdata` still selected
                await assertNumRowsExpanded(1) // only `diff` directory expanded
            })
        })

        describe('directory page', () => {
            // TODO(slimsag:discussions): temporarily disabled because the discussions feature flag removes this component.
            /*
            it('shows a row for each file in the directory', async () => {
                await page.goto(baseURL + '/github.com/gorilla/securecookie@e59506cc896acb7f7bf732d4fdf5e25f7ccd8983')
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.tree-page__entries-directories')
                await retry(async () =>
                    assert.equal(
                        await page.evaluate(
                            () => document.querySelectorAll('.tree-page__entries-directories .tree-entry').length
                        ),
                        1
                    )
                )
                await retry(async () =>
                    assert.equal(
                        await page.evaluate(
                            () => document.querySelectorAll('.tree-page__entries-files .tree-entry').length
                        ),
                        7
                    )
                )
            })
            */

            test('shows commit information on a row', async () => {
                await page.goto(baseURL + '/github.com/gorilla/securecookie@e59506cc896acb7f7bf732d4fdf5e25f7ccd8983', {
                    waitUntil: 'domcontentloaded',
                })
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.git-commit-node__message')
                await retry(async () =>
                    expect(
                        await page.evaluate(() => document.querySelectorAll('.git-commit-node__message')[2].textContent)
                    ).toContain('Add fuzz testing corpus.')
                )
                await retry(async () =>
                    expect(
                        await page.evaluate(() =>
                            document.querySelectorAll('.git-commit-node-byline')[2].textContent!.trim()
                        )
                    ).toContain('Kamil Kisiel')
                )
                await retry(async () =>
                    expect(
                        await page.evaluate(() => document.querySelectorAll('.git-commit-node__oid')[2].textContent)
                    ).toEqual('c13558c')
                )
            })

            // TODO(slimsag:discussions): temporarily disabled because the discussions feature flag removes this component.
            /*
            it('navigates when clicking on a row', async () => {
                await page.goto(baseURL + '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d')
                await enableOrAddRepositoryIfNeeded()
                // click on directory
                await page.waitForSelector('.tree-entry')
                await page.click('.tree-entry')
                await assertWindowLocation(
                    '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d/-/tree/websocket'
                )
            })
            */
        })

        describe('rev resolution', () => {
            test('shows clone in progress interstitial page', async () => {
                await page.goto(baseURL + '/github.com/sourcegraphtest/AlwaysCloningTest')
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.hero-page__subtitle')
                await retry(async () =>
                    expect(
                        await page.evaluate(() => document.querySelector('.hero-page__subtitle')!.textContent)
                    ).toEqual('Cloning in progress')
                )
            })

            test('resolves default branch when unspecified', async () => {
                await page.goto(baseURL + '/github.com/sourcegraph/go-diff/-/blob/diff/diff.go')
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.repo-header__rev')
                await retry(async () => {
                    expect(
                        await page.evaluate(() => document.querySelector('.repo-header__rev')!.textContent!.trim())
                    ).toEqual('master')
                })
                // Verify file contents are loaded.
                await page.waitForSelector(blobTableSelector)
            })

            test('updates rev with switcher', async () => {
                await page.goto(baseURL + '/github.com/sourcegraph/checkup/-/blob/s3.go')
                await enableOrAddRepositoryIfNeeded()
                // Open rev switcher
                await page.waitForSelector('.repo-header__rev')
                await page.click('.repo-header__rev')
                // Click "Tags" tab
                await page.click('.revisions-popover .tab-bar__tab:nth-child(2)')
                await page.waitForSelector('a.git-ref-node[href*="0.1.0"]')
                await page.click('a.git-ref-node[href*="0.1.0"]')
                await assertWindowLocation('/github.com/sourcegraph/checkup@v0.1.0/-/blob/s3.go')
            })
        })

        describe('hovers', () => {
            describe(`Blob`, () => {
                // Temporarely skipped because of flakiness. TODO find cause
                test.skip('gets displayed and updates URL when clicking on a token', async () => {
                    await page.goto(
                        baseURL +
                            '/github.com/sourcegraph/godockerize@05bac79edd17c0f55127871fa9c6f4d91bebf07c/-/blob/godockerize.go'
                    )
                    await enableOrAddRepositoryIfNeeded()
                    await page.waitForSelector(blobTableSelector)
                    await clickToken(23, 2)
                    await assertWindowLocation(
                        '/github.com/sourcegraph/godockerize@05bac79edd17c0f55127871fa9c6f4d91bebf07c/-/blob/godockerize.go#L23:3'
                    )
                    await getHoverContents() // verify there is a hover
                })

                // Skipped until the Go language server is capable of being run locally (without needing
                // access to sourcegraph-frontend's internal API)
                // TODO@ggilmore
                // TODO@chrismwendt
                test.skip('gets displayed when navigating to a URL with a token position', async () => {
                    await page.goto(
                        baseURL +
                            '/github.com/sourcegraph/godockerize@05bac79edd17c0f55127871fa9c6f4d91bebf07c/-/blob/godockerize.go#L23:3'
                    )
                    await enableOrAddRepositoryIfNeeded()
                    await retry(
                        async () =>
                            await assertHoverContentContains(
                                `The name of the program. Defaults to path.Base(os.Args[0]) \n`,
                                2
                            )
                    )
                })

                // Skipped until the Go language server is capable of being run locally (without needing
                // access to sourcegraph-frontend's internal API)
                // TODO@ggilmore
                // TODO@chrismwendt
                describe.skip('jump to definition', () => {
                    test('noops when on the definition', async () => {
                        await page.goto(
                            baseURL +
                                '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6'
                        )
                        await enableOrAddRepositoryIfNeeded()
                        await clickHoverJ2D()
                        await assertWindowLocation(
                            '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6'
                        )
                    })

                    test('does navigation (same repo, same file)', async () => {
                        await page.goto(
                            baseURL +
                                '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L25:10'
                        )
                        await enableOrAddRepositoryIfNeeded()
                        await clickHoverJ2D()
                        return await assertWindowLocation(
                            '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6'
                        )
                    })

                    test('does navigation (same repo, different file)', async () => {
                        await page.goto(
                            baseURL +
                                '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/print.go#L13:31'
                        )
                        await enableOrAddRepositoryIfNeeded()
                        await clickHoverJ2D()
                        await assertWindowLocation(
                            '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/diff.pb.go#L38:6'
                        )
                        // Verify file tree is highlighting the new path.
                        return await page.waitForSelector('.tree__row--active [data-tree-path="diff/diff.pb.go"]')
                    })

                    test('does navigation (external repo)', async () => {
                        await page.goto(
                            baseURL +
                                '/github.com/sourcegraph/vcsstore@267289226b15e5b03adedc9746317455be96e44c/-/blob/server/diff.go#L27:30'
                        )
                        await enableOrAddRepositoryIfNeeded()
                        await clickHoverJ2D()
                        await assertWindowLocation(
                            '/github.com/sourcegraph/go-vcs@aa7c38442c17a3387b8a21f566788d8555afedd0/-/blob/vcs/repository.go#L103:6'
                        )
                    })
                })

                describe('find references', () => {
                    // Skipped until the Go language server is capable of being run locally (without needing
                    // access to sourcegraph-frontend's internal API)
                    // TODO@ggilmore
                    // TODO@chrismwendt
                    test.skip('opens widget and fetches local references', async (): Promise<void> => {
                        jest.setTimeout(120000)

                        await page.goto(
                            baseURL +
                                '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6'
                        )
                        await enableOrAddRepositoryIfNeeded()
                        await clickHoverFindRefs()
                        await assertWindowLocation(
                            '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6&tab=references'
                        )

                        await assertNonemptyRefs()

                        // verify the appropriate # of references are fetched
                        await page.waitForSelector('.panel__tabs-content .file-match__list')
                        await retry(async () =>
                            expect(
                                await page.evaluate(
                                    () => document.querySelectorAll('.panel__tabs-content .file-match__item').length
                                )
                            ).toEqual(
                                // 4 references, two of which got merged into one because their context overlaps
                                3
                            )
                        )

                        // verify all the matches highlight a `MultiFileDiffReader` token
                        await assertAllHighlightedTokens('MultiFileDiffReader')
                    })

                    // Testing external references on localhost is unreliable, since different dev environments will
                    // not guarantee what repo(s) have been indexed. It's possible a developer has an environment with only
                    // 1 repo, in which case there would never be external references. So we *only* run this test against
                    // non-localhost servers.
                    const skipExternalReferences = baseURL === 'http://localhost:3080'
                    ;(skipExternalReferences ? test.skip : test)(
                        'opens widget and fetches external references',
                        async () => {
                            await page.goto(
                                baseURL +
                                    '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L32:16&tab=references'
                            )
                            await enableOrAddRepositoryIfNeeded()

                            // verify some external refs are fetched (we cannot assert how many, but we can check that the matched results
                            // look like they're for the appropriate token)
                            await assertNonemptyRefs()

                            // verify all the matches highlight a `Reader` token
                            await assertAllHighlightedTokens('Reader')
                        }
                    )
                })
            })
        })

        describe.skip('godoc.org "Uses" links', () => {
            test('resolves standard library function', async () => {
                // https://godoc.org/bytes#Compare
                await page.goto(baseURL + '/-/godoc/refs?def=Compare&pkg=bytes&repo=')
                await enableOrAddRepositoryIfNeeded()
                await assertWindowLocationPrefix('/github.com/golang/go/-/blob/src/bytes/bytes_decl.go')
                await assertStickyHighlightedToken('Compare')
                await assertNonemptyRefs()
                await assertAllHighlightedTokens('Compare')
            })

            test('resolves standard library function (from stdlib repo)', async () => {
                // https://godoc.org/github.com/golang/go/src/bytes#Compare
                await page.goto(
                    baseURL +
                        '/-/godoc/refs?def=Compare&pkg=github.com%2Fgolang%2Fgo%2Fsrc%2Fbytes&repo=github.com%2Fgolang%2Fgo'
                )
                await enableOrAddRepositoryIfNeeded()
                await assertWindowLocationPrefix('/github.com/golang/go/-/blob/src/bytes/bytes_decl.go')
                await assertStickyHighlightedToken('Compare')
                await assertNonemptyRefs()
                await assertAllHighlightedTokens('Compare')
            })

            test('resolves external package function (from gorilla/mux)', async () => {
                // https://godoc.org/github.com/gorilla/mux#Router
                await page.goto(
                    baseURL + '/-/godoc/refs?def=Router&pkg=github.com%2Fgorilla%2Fmux&repo=github.com%2Fgorilla%2Fmux'
                )
                await enableOrAddRepositoryIfNeeded()
                await assertWindowLocationPrefix('/github.com/gorilla/mux/-/blob/mux.go')
                await assertStickyHighlightedToken('Router')
                await assertNonemptyRefs()
                await assertAllHighlightedTokens('Router')
            })
        })

        describe('external code host links', () => {
            test('on repo navbar ("View on GitHub")', async () => {
                await page.goto(
                    baseURL +
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L19',
                    { waitUntil: 'domcontentloaded' }
                )
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.nav-link[href*="https://github"]')
                await retry(async () =>
                    expect(
                        await page.evaluate(
                            () =>
                                (document.querySelector('.nav-link[href*="https://github"]') as HTMLAnchorElement).href
                        )
                    ).toEqual(
                        'https://github.com/sourcegraph/go-diff/blob/3f415a150aec0685cb81b73cc201e762e075006d/diff/parse.go#L19'
                    )
                )
            })
        })
    })

    describe('Search component', () => {
        test.skip('renders results for sourcegraph/go-diff (no search group)', async () => {
            await page.goto(
                baseURL + '/search?q=diff+repo:sourcegraph/go-diff%403f415a150aec0685cb81b73cc201e762e075006d+type:file'
            )
            await page.waitForSelector('.search-results__stats')
            await retry(async () => {
                const label = await page.evaluate(
                    () => document.querySelector('.search-results__stats')!.textContent || ''
                )
                expect(label.includes('results')).toEqual(true)
            })
            // navigate to result on click
            await page.click('.file-match__item')
            await retry(async () => {
                expect(await page.evaluate(() => window.location.href)).toEqual(
                    baseURL +
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/testdata/sample_file_extended_empty_rename.diff#L1:1a'
                )
            })
        })

        test('accepts query for sourcegraph/jsonrpc2', async () => {
            await page.goto(baseURL + '/search')

            // Update the input value
            await page.waitForSelector('input.query-input2__input')
            await page.keyboard.type('test repo:sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d')

            // TODO: test search scopes

            // Submit the search
            await page.click('.search-button')

            await page.waitForSelector('.e2e-search-results-stats')
            await retry(async () => {
                const label = await page.evaluate(
                    () => document.querySelector('.e2e-search-results-stats')!.textContent || ''
                )
                const match = /(\d+) results?/.exec(label)
                if (!match) {
                    throw new Error(
                        `.e2e-search-results-stats textContent did not match regex '(\d+) results': '${label}'`
                    )
                }
                const numberOfResults = parseInt(match[1], 10)
                expect(numberOfResults).toBeGreaterThan(0)
            })
        })
    })
})
