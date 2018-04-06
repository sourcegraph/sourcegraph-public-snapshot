import { assert } from 'chai'
import { Browser, connect, launch, Page } from 'puppeteer'
import { retry } from '../util/e2e-test-utils'

describe('e2e test suite', () => {
    let authenticate: (page: Page) => Promise<void>
    let baseURL: string
    if (process.env.SOURCEGRAPH_BASE_URL) {
        baseURL = process.env.SOURCEGRAPH_BASE_URL
        // Assume that the dogfood (sourcegraph.sgdev.org) OIDC override token works.
        authenticate = page => page.setExtraHTTPHeaders({ 'X-Oidc-Override': '2qzNBYQmUigCFdVVjDGyFfp' })
    } else {
        baseURL = 'http://localhost:3080'
        const sessionAuthorization = process.env.SOURCEGRAPH_SESSION
        if (!sessionAuthorization) {
            console.log(`Fatal: You must set the SOURCEGRAPH_SESSION env var to a session token valid for a site admin on ${baseURL}.

To obtain this value, run the following in your web browser's JavaScript console in a tab where you're logged into ${baseURL} as a site admin:

  >  window.context.xhrHeaders.Authorization
  <- "session MTUyMDgzNTEyNnx..."

Then rerun this test command with that value in the SOURCEGRAPH_SESSION env var. For example:

  SOURCEGRAPH_SESSION="session MTUyMDgzNTEyNnx..." npm run test-e2e
`)
            process.exit(1)
            return // for sessionAuthorization type inference
        }
        authenticate = page => page.setExtraHTTPHeaders({ Authorization: sessionAuthorization })
    }
    const browserWSEndpoint = process.env.BROWSER_WS_ENDPOINT

    let browser: Browser
    let page: Page
    if (browserWSEndpoint) {
        before('Connect to browser', async () => {
            browser = await connect({ browserWSEndpoint })
        })
        after('Disconnect from browser', async () => {
            if (browser) {
                await browser.disconnect()
            }
        })
    } else {
        before('Start browser', async () => {
            browser = await launch()
        })
        after('Close browser', async () => {
            if (browser) {
                await browser.close()
            }
        })
    }
    beforeEach('Open page', async () => {
        page = await browser.newPage()
        await authenticate(page)
    })
    afterEach('Close page', async () => {
        await page.close()
    })

    const enableOrAddRepositoryIfNeeded = async (): Promise<any> => {
        // Disable any toasts, which can interfere with clicking on the enable/add button.
        try {
            await page.click('.toast__close-button')
        } catch (e) {
            // Probably no toast was showing.
        }
        // Wait for the repository container or a repository error page to be shown.
        await Promise.race([
            // Add or enable repository.
            page
                .waitForSelector('.repository-error-page__btn:not([disabled])')
                .then(
                    () =>
                        page
                            .click('.repository-error-page__btn:not([disabled])')
                            .then(() => page.waitForSelector('.repo-rev-container')),
                    () => void 0
                ),

            // Repository is cloning.
            page.waitForSelector('.repository-cloning-page'),

            // Repository is already enabled and added; nothing to do.
            page.waitForSelector('.repo-rev-container'),
        ])
    }

    const assertWindowLocation = async (location: string, isAbsolute = false): Promise<any> => {
        const url = isAbsolute ? location : baseURL + location
        await retry(async () => {
            assert.equal(await page.evaluate(() => window.location.href), url)
        })
    }

    const assertWindowLocationPrefix = async (locationPrefix: string, isAbsolute = false): Promise<any> => {
        const prefix = isAbsolute ? locationPrefix : baseURL + locationPrefix
        await retry(async () => {
            const loc: string = await page.evaluate(() => window.location.href)
            assert.ok(loc.startsWith(prefix), `expected window.location to start with ${prefix}, but got ${loc}`)
        })
    }

    const assertStickyHighlightedToken = async (label: string): Promise<void> => {
        await page.waitForSelector('.selection-highlight-sticky') // make sure matched token is highlighted
        await retry(async () =>
            assert.equal(
                await page.evaluate(() => document.querySelector('.selection-highlight-sticky')!.textContent),
                label
            )
        )
    }

    const assertAllHighlightedTokens = async (label: string): Promise<void> => {
        const highlightedTokens: string[] = await page.evaluate(() =>
            Array.from(document.querySelectorAll('.selection-highlight')).map(el => el.textContent)
        )
        assert.ok(
            highlightedTokens.every(txt => txt === label),
            `unexpected tokens highlighted (expected '${label}'): ${highlightedTokens}`
        )
    }

    const assertNonemptyLocalRefs = async (): Promise<void> => {
        // verify active group is 'local'
        await page.waitForSelector('.blob-references-panel .tab-bar__tab--active')
        assert.equal(
            await page.evaluate(() =>
                document.querySelector('.blob-references-panel .tab-bar__tab--active')!.textContent!.replace(/\d/g, '')
            ),
            'This repository'
        )

        await page.waitForSelector('.blob-references-panel .badge')
        await retry(async () =>
            assert.ok(
                parseInt(
                    await page.evaluate(
                        () => document.querySelector('.blob-references-panel .tab-bar__tab--active .badge')!.textContent
                    ),
                    10
                ) > 0, // assert some (local) refs fetched
                'expected some local references, got none'
            )
        )
    }

    const assertNonemptyExternalRefs = async (): Promise<void> => {
        // verify active group is 'external'
        await page.waitForSelector('.blob-references-panel .tab-bar__tab--active')
        assert.equal(
            await page.evaluate(
                () => document.querySelector('.blob-references-panel .tab-bar__tab--active')!.textContent
            ),
            'Other repositories'
        )
        await page.waitForSelector('.blob-references-panel .tab-bar__tab--active .badge')
        await retry(async () => {
            assert.ok(
                parseInt(
                    await page.evaluate(
                        () =>
                            document.querySelector('.blob-references-panel .tab-bar__tab--active .badge')!.textContent! // get the external refs count
                    ),
                    10
                ) > 0, // assert some external refs fetched
                'expected some external references, got none'
            )
        })
    }

    describe('Theme switcher', () => {
        it('changes the theme when toggle is clicked', async () => {
            await page.goto(baseURL + '/github.com/gorilla/mux/-/blob/mux.go')
            await enableOrAddRepositoryIfNeeded()
            await page.waitForSelector('.theme')
            const currentThemes: string[] = await page.evaluate(() =>
                Array.from(document.querySelector('.theme')!.classList).filter(c => c.startsWith('theme-'))
            )
            assert.equal(currentThemes.length, 1, 'Expected 1 theme')
            const expectedTheme = currentThemes[0] === 'theme-dark' ? 'theme-light' : 'theme-dark'
            await page.click('.theme-switcher')
            assert.deepEqual(
                await page.evaluate(() =>
                    Array.from(document.querySelector('.theme')!.classList).filter(c => c.startsWith('theme-'))
                ),
                [expectedTheme]
            )
        })
    })

    describe('Repository component', () => {
        const blobSelector = '.blob > table'
        const clickToken = async (line: number, spanOffset: number): Promise<void> => {
            const selector = `${blobSelector} tr:nth-child(${line}) > td.code > span:nth-child(${spanOffset})`
            await page.waitForSelector(selector)
            await page.click(selector)
        }

        const getTooltipDoc = async (): Promise<string> => {
            await page.waitForSelector('.tooltip__doc')
            return await page.evaluate(() => document.querySelector('.tooltip__doc')!.textContent)
        }

        const tooltipActionsSelector = '.sg-tooltip > .tooltip__actions'
        const clickTooltipJ2D = async (): Promise<void> => {
            const selector = `${tooltipActionsSelector} > a:nth-child(1)`
            await page.waitForSelector(selector)
            await page.click(selector)
        }
        const clickTooltipFindRefs = async (): Promise<void> => {
            const selector = `${tooltipActionsSelector} > a:nth-child(2)`
            await page.waitForSelector(selector)
            await page.click(selector)
        }

        describe('file tree', () => {
            it('does navigation on file click', async () => {
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

            it('expands directory on row click (no navigation)', async () => {
                await page.goto(baseURL + '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d')
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.tree__row-icon')
                await page.click('.tree__row-icon')
                await page.waitForSelector('.tree__row--active [data-tree-path="websocket"]')
                await page.waitForSelector('.tree__row--expanded [data-tree-path="websocket"]')
                await assertWindowLocation('/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d')
            })

            it('does navigation on directory row click', async () => {
                await page.goto(baseURL + '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d')
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.tree__row-label')
                await page.click('.tree__row-label')
                await page.waitForSelector('.tree__row--active [data-tree-path="websocket"]')
                await page.waitForSelector('.tree__row--expanded [data-tree-path="websocket"]')
                await assertWindowLocation(
                    '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d/-/tree/websocket'
                )
            })

            it('selects the current file', async () => {
                await page.goto(
                    baseURL +
                        '/github.com/sourcegraph/godockerize@05bac79edd17c0f55127871fa9c6f4d91bebf07c/-/blob/godockerize.go'
                )
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.tree__row--active [data-tree-path="godockerize.go"]')
            })

            it('shows partial tree when opening directory', async () => {
                await page.goto(
                    baseURL +
                        '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d/-/tree/websocket'
                )
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.tree__row')
                assert.equal(await page.evaluate(() => document.querySelectorAll('.tree__row').length), 1)
            })

            it('responds to keyboard shortcuts', async () => {
                const assertNumRowsExpanded = async (expectedCount: number) => {
                    assert.equal(
                        await page.evaluate(() => document.querySelectorAll('.tree__row--expanded').length),
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
                await page.waitForSelector('.tree__row--active [data-tree-path="diff"]')
                await page.keyboard.press('ArrowRight') // arrow right (expand 'diff' directory)
                await page.waitForSelector('.tree__row--active [data-tree-path="diff"]')
                await page.waitForSelector('.tree__row--expanded [data-tree-path="diff"]')
                await page.keyboard.press('ArrowRight') // arrow right (move to nested 'diff/testdata' directory)
                await page.waitForSelector('.tree__row--active [data-tree-path="diff/testdata"]')
                await assertNumRowsExpanded(1) // only `diff` directory is expanded, though `diff/testdata` is expanded

                await page.keyboard.press('ArrowRight') // arrow right (expand 'diff/testdata' directory)
                await page.waitForSelector('.tree__row--active [data-tree-path="diff/testdata"]')
                await page.waitForSelector('.tree__row--expanded [data-tree-path="diff/testdata"]')
                await assertNumRowsExpanded(2) // `diff` and `diff/testdata` directories expanded

                // select some file nested under `diff/testdata`
                await page.keyboard.press('ArrowDown') // arrow down
                await page.keyboard.press('ArrowDown') // arrow down
                await page.keyboard.press('ArrowDown') // arrow down
                await page.keyboard.press('ArrowDown') // arrow down
                await page.waitForSelector('.tree__row--active [data-tree-path="diff/testdata/empty_orig.diff"]')

                await page.keyboard.press('ArrowLeft') // arrow left (navigate immediately up to parent directory `diff/testdata`)
                await page.waitForSelector('.tree__row--active [data-tree-path="diff/testdata"]')
                await assertNumRowsExpanded(2) // `diff` and `diff/testdata` directories expanded

                await page.keyboard.press('ArrowLeft') // arrow left
                await page.waitForSelector('.tree__row--active [data-tree-path="diff/testdata"]') // `diff/testdata` still selected
                await assertNumRowsExpanded(1) // only `diff` directory expanded
            })
        })

        describe('directory page', () => {
            it('shows a row for each file in the directory', async () => {
                await page.goto(baseURL + '/github.com/gorilla/securecookie@e59506cc896acb7f7bf732d4fdf5e25f7ccd8983')
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.directory-page__entries-directories')
                await retry(async () =>
                    assert.equal(
                        await page.evaluate(
                            () =>
                                document.querySelectorAll('.directory-page__entries-directories .directory-entry')
                                    .length
                        ),
                        1
                    )
                )
                await retry(async () =>
                    assert.equal(
                        await page.evaluate(
                            () => document.querySelectorAll('.directory-page__entries-files .directory-entry').length
                        ),
                        7
                    )
                )
            })

            it('shows commit information on a row', async () => {
                await page.goto(baseURL + '/github.com/gorilla/securecookie@e59506cc896acb7f7bf732d4fdf5e25f7ccd8983')
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.directory-page__commit-message')
                await retry(async () =>
                    assert.equal(
                        await page.evaluate(
                            () => document.querySelectorAll('.directory-page__commit-message')[2].textContent
                        ),
                        'Add fuzz testing corpus.'
                    )
                )
                await retry(async () =>
                    assert.equal(
                        await page.evaluate(() =>
                            document.querySelectorAll('.directory-page__commit-author')[2].textContent!.trim()
                        ),
                        'Kamil Kisiel'
                    )
                )
                await retry(async () =>
                    assert.equal(
                        await page.evaluate(
                            () => document.querySelectorAll('.directory-page__commit-id')[2].textContent
                        ),
                        'c13558c'
                    )
                )
            })

            it('navigates when clicking on a row', async () => {
                await page.goto(baseURL + '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d')
                await enableOrAddRepositoryIfNeeded()
                // click on directory
                await page.waitForSelector('.directory-entry')
                await page.click('.directory-entry')
                await assertWindowLocation(
                    '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d/-/tree/websocket'
                )
            })
        })

        describe('rev resolution', () => {
            it('shows clone in progress interstitial page', async () => {
                await page.goto(baseURL + '/github.com/sourcegraphtest/AlwaysCloningTest')
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.hero-page__subtitle')
                await retry(async () =>
                    assert.equal(
                        await page.evaluate(() => document.querySelector('.hero-page__subtitle')!.textContent),
                        'Cloning in progress'
                    )
                )
            })

            it('resolves default branch when unspecified', async () => {
                await page.goto(baseURL + '/github.com/sourcegraph/go-diff/-/blob/diff/diff.go')
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.repo-header__rev')
                await retry(async () => {
                    assert.equal(
                        await page.evaluate(() => document.querySelector('.repo-header__rev')!.textContent!.trim()),
                        'master'
                    )
                })
                // Verify file contents are loaded.
                await page.waitForSelector(blobSelector)
            })

            it('updates rev with switcher', async () => {
                await page.goto(baseURL + '/github.com/sourcegraph/checkup/-/blob/s3.go')
                await enableOrAddRepositoryIfNeeded()
                // Open rev switcher
                await page.waitForSelector('.repo-header__rev')
                await page.click('.repo-header__rev')
                // Click "Tags" tab
                await page.click('.repo-header__rev .tab-bar__tab:nth-child(2)')
                await page.waitForSelector('.popover__node-link[href*="0.1.0"]')
                await page.click('.popover__node-link[href*="0.1.0"]')
                await assertWindowLocation('/github.com/sourcegraph/checkup@v0.1.0/-/blob/s3.go')
            })
        })

        describe('tooltips', () => {
            it('gets displayed and updates URL when clicking on a token', async () => {
                await page.goto(
                    baseURL +
                        '/github.com/sourcegraph/godockerize@05bac79edd17c0f55127871fa9c6f4d91bebf07c/-/blob/godockerize.go'
                )
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.blob')
                await clickToken(23, 2)
                await assertWindowLocation(
                    '/github.com/sourcegraph/godockerize@05bac79edd17c0f55127871fa9c6f4d91bebf07c/-/blob/godockerize.go#L23:3'
                )
                await getTooltipDoc() // verify there is a tooltip
            })

            it('gets displayed when navigating to a URL with a token position', async () => {
                await page.goto(
                    baseURL +
                        '/github.com/sourcegraph/godockerize@05bac79edd17c0f55127871fa9c6f4d91bebf07c/-/blob/godockerize.go#L23:3'
                )
                await enableOrAddRepositoryIfNeeded()
                await retry(async () =>
                    assert.equal(await getTooltipDoc(), `The name of the program. Defaults to path.Base(os.Args[0]) \n`)
                )
            })

            describe('jump to definition', () => {
                it('noops when on the definition', async () => {
                    await page.goto(
                        baseURL +
                            '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6'
                    )
                    await enableOrAddRepositoryIfNeeded()
                    await clickTooltipJ2D()
                    await assertWindowLocation(
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6'
                    )
                })

                it('does navigation (same repo, same file)', async () => {
                    await page.goto(
                        baseURL +
                            '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go'
                    )
                    await enableOrAddRepositoryIfNeeded()
                    await clickToken(25, 5)
                    await clickTooltipJ2D()
                    return await assertWindowLocation(
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6'
                    )
                })

                it('does navigation (same repo, different file)', async () => {
                    await page.goto(
                        baseURL +
                            '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/print.go#L13:31'
                    )
                    await enableOrAddRepositoryIfNeeded()
                    await clickTooltipJ2D()
                    await assertWindowLocation(
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/diff.pb.go#L38:6'
                    )
                    // Verify file tree is highlighting the new path.
                    return await page.waitForSelector('.tree__row--active [data-tree-path="diff/diff.pb.go"]')
                })

                it('does navigation (external repo)', async () => {
                    await page.goto(
                        baseURL +
                            '/github.com/sourcegraph/vcsstore@267289226b15e5b03adedc9746317455be96e44c/-/blob/server/diff.go#L27:30'
                    )
                    await enableOrAddRepositoryIfNeeded()
                    await clickTooltipJ2D()
                    await assertWindowLocation(
                        '/github.com/sourcegraph/go-vcs@aa7c38442c17a3387b8a21f566788d8555afedd0/-/blob/vcs/repository.go#L103:6'
                    )
                })
            })

            describe('find references', () => {
                it('opens widget and fetches local references', async function(): Promise<void> {
                    this.timeout(120000)

                    await page.goto(
                        baseURL +
                            '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6'
                    )
                    await enableOrAddRepositoryIfNeeded()
                    await clickTooltipFindRefs()
                    await assertWindowLocation(
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6$references'
                    )

                    await assertNonemptyLocalRefs()

                    // verify the appropriate # of references are fetched
                    await page.waitForSelector('.blob-references-panel .badge')
                    await retry(async () =>
                        assert.equal(
                            await page.evaluate(
                                () => document.querySelector('.blob-references-panel .badge')!.textContent
                            ),
                            '5'
                        )
                    )

                    // verify all the matches highlight a `MultiFileDiffReader` token
                    await assertAllHighlightedTokens('MultiFileDiffReader')
                })

                it('opens widget and fetches external references', async function(): Promise<void> {
                    // Testing external references on localhost is unreliable, since different dev environments will
                    // not guarantee what repo(s) have been indexed. It's possible a developer has an environment with only
                    // 1 repo, in which case there would never be external references. So we *only* run this test against
                    // non-localhost servers.
                    if (baseURL === 'http://localhost:3080') {
                        this.skip()
                        return
                    }

                    await page.goto(
                        baseURL +
                            '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L32:16$references:external'
                    )
                    await enableOrAddRepositoryIfNeeded()

                    // verify some external refs are fetched (we cannot assert how many, but we can check that the matched results
                    // look like they're for the appropriate token)
                    await assertNonemptyExternalRefs()

                    // verify all the matches highlight a `Reader` token
                    await assertAllHighlightedTokens('Reader')
                })
            })
        })

        describe.skip('godoc.org "Uses" links', () => {
            it('resolves standard library function', async () => {
                // https://godoc.org/bytes#Compare
                await page.goto(baseURL + '/-/godoc/refs?def=Compare&pkg=bytes&repo=')
                await enableOrAddRepositoryIfNeeded()
                await assertWindowLocationPrefix('/github.com/golang/go/-/blob/src/bytes/bytes_decl.go')
                await assertStickyHighlightedToken('Compare')
                await assertNonemptyLocalRefs()
                await assertAllHighlightedTokens('Compare')
            })

            it('resolves standard library function (from stdlib repo)', async () => {
                // https://godoc.org/github.com/golang/go/src/bytes#Compare
                await page.goto(
                    baseURL +
                        '/-/godoc/refs?def=Compare&pkg=github.com%2Fgolang%2Fgo%2Fsrc%2Fbytes&repo=github.com%2Fgolang%2Fgo'
                )
                await enableOrAddRepositoryIfNeeded()
                await assertWindowLocationPrefix('/github.com/golang/go/-/blob/src/bytes/bytes_decl.go')
                await assertStickyHighlightedToken('Compare')
                await assertNonemptyLocalRefs()
                await assertAllHighlightedTokens('Compare')
            })

            it('resolves external package function (from gorilla/mux)', async () => {
                // https://godoc.org/github.com/gorilla/mux#Router
                await page.goto(
                    baseURL + '/-/godoc/refs?def=Router&pkg=github.com%2Fgorilla%2Fmux&repo=github.com%2Fgorilla%2Fmux'
                )
                await enableOrAddRepositoryIfNeeded()
                await assertWindowLocationPrefix('/github.com/gorilla/mux/-/blob/mux.go')
                await assertStickyHighlightedToken('Router')
                await assertNonemptyLocalRefs()
                await assertAllHighlightedTokens('Router')
            })
        })

        describe('external code host links', () => {
            it('on line blame', async () => {
                await page.goto(
                    baseURL +
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L19'
                )
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.blob > table > tbody > tr:nth-child(19) .blame')
                await page.evaluate(() => {
                    const blame = document.querySelector('.blob > table > tbody > tr:nth-child(19) .blame')!
                    const rect = blame.getBoundingClientRect() as DOMRect
                    const ev = new MouseEvent('click', {
                        view: window,
                        bubbles: true,
                        cancelable: true,
                        clientX: rect.x + rect.width + 2,
                        clientY: rect.y + 2,
                    })
                    blame.dispatchEvent(ev)
                })
                await assertWindowLocation(
                    'https://github.com/sourcegraph/go-diff/commit/f93a4e38b36b4003edbfdff66d3b92c6cd977c1c',
                    true
                )
            })

            it('on repo navbar ("View on GitHub")', async () => {
                await page.goto(
                    baseURL +
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L19',
                    { waitUntil: 'domcontentloaded' }
                )
                await enableOrAddRepositoryIfNeeded()
                await page.waitForSelector('.composite-container__header-action[href*="https://github"]')
                await retry(async () =>
                    assert.equal(
                        await page.evaluate(
                            () =>
                                (document.querySelector(
                                    '.composite-container__header-action[href*="https://github"]'
                                ) as HTMLAnchorElement).href
                        ),
                        'https://github.com/sourcegraph/go-diff/blob/3f415a150aec0685cb81b73cc201e762e075006d/diff/parse.go#L19'
                    )
                )
            })
        })
    })

    describe('Search component', () => {
        it.skip('renders results for sourcegraph/go-diff (no search group)', async () => {
            await page.goto(
                baseURL + '/search?q=diff+repo:sourcegraph/go-diff%403f415a150aec0685cb81b73cc201e762e075006d+type:file'
            )
            await page.waitForSelector('.search-results__stats')
            await retry(async () => {
                const label: string = await page.evaluate(
                    () => document.querySelector('.search-results__stats')!.textContent
                )
                assert.equal(label.includes('results'), true, 'incorrect label for search results')
            })
            // navigate to result on click
            await page.click('.file-match__item')
            await retry(async () => {
                assert.equal(
                    await page.evaluate(() => window.location.href),
                    baseURL +
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/testdata/sample_file_extended_empty_rename.diff#L1:1a',
                    'Unexpected window.location.href after clicking result'
                )
            })
        })

        it('accepts query for sourcegraph/jsonrpc2', async () => {
            await page.goto(baseURL + '/search')

            // Update the input value
            await page.waitForSelector('input')
            await page.keyboard.type('test repo:sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d')

            // TODO: test search scopes

            // Submit the search
            await page.click('button')

            await page.waitForSelector('.search-results__stats')
            await retry(async () => {
                const label: string = await page.evaluate(
                    () => document.querySelector('.search-results__stats')!.textContent
                )
                const match = /(\d+) results/.exec(label)
                const numberOfResults = parseInt(match![1], 10)
                assert.isAbove(numberOfResults, 0, 'Expected >0 search results')
            })
        })
    })
})
