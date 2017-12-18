import * as assert from 'assert'
import { Chromeless } from 'chromeless'

const chromeLauncher = require('chrome-launcher')

describe('e2e test suite', () => {
    const baseURL = process.env.SOURCEGRAPH_BASE_URL || 'http://localhost:3080'

    let headlessChrome: any
    let chrome: Chromeless<any>
    before(() => {
        if (!process.env.SKIP_LAUNCH_CHROME) {
            // We manually launch chrome w/ chrome launcher because chromeless doesn't offer a way to
            // launch in headless mode.
            return chromeLauncher
                .launch({
                    startingUrl: baseURL,
                    port: 9222,
                    chromeFlags: ['--headless', '--disable-gpu'],
                })
                .then((c: any) => (headlessChrome = c))
        }
    })
    beforeEach(() => {
        chrome = new Chromeless({ waitTimeout: 30000, launchChrome: false })
        return chrome.setExtraHTTPHeaders({ 'X-Oidc-Override': '2qzNBYQmUigCFdVVjDGyFfp' })
    })
    afterEach(() => chrome.end())
    after(() => {
        if (headlessChrome) {
            return headlessChrome.kill()
        }
    })

    const retry = async (expression: () => Promise<any>, numRetries = 3): Promise<any> => {
        while (true) {
            try {
                await expression()
                break
            } catch (e) {
                numRetries -= 1
                if (numRetries <= 0) {
                    throw e
                }
                await new Promise<any>(resolve => setTimeout(resolve, 5000))
            }
        }
    }

    const assertWindowLocation = async (location: string, isAbsolute = false): Promise<any> => {
        const url = isAbsolute ? location : baseURL + location
        await retry(async () => {
            assert.equal(await chrome.evaluate(() => window.location.href), url)
        })
    }

    const assertWindowLocationPrefix = async (locationPrefix: string, isAbsolute = false): Promise<any> => {
        const prefix = isAbsolute ? locationPrefix : baseURL + locationPrefix
        await retry(async () => {
            const loc = await chrome.evaluate<string>(() => window.location.href)
            assert.ok(loc.startsWith(prefix), `expected window.location to start with ${prefix}, but got ${loc}`)
        })
    }

    const assertStickyHighlightedToken = async (label: string): Promise<void> => {
        await chrome.wait('.selection-highlight-sticky') // make sure matched token is highlighted
        await retry(async () =>
            assert.equal(
                await chrome.evaluate<string>(() => document.querySelector('.selection-highlight-sticky')!.textContent),
                label
            )
        )
    }

    const assertAllHighlightedTokens = async (label: string): Promise<void> => {
        const highlightedTokens: string[] = JSON.parse(
            await chrome.evaluate<string>(() =>
                JSON.stringify(Array.from(document.querySelectorAll('.selection-highlight')).map(el => el.textContent))
            )
        )
        assert.ok(
            highlightedTokens.every(txt => txt === label),
            `unexpected tokens highlighted (expected '${label}'): ${highlightedTokens}`
        )
    }

    const assertNonemptyLocalRefs = async (): Promise<void> => {
        // verify active group is 'local'
        await chrome.wait('.references-widget__title-bar-group--active')
        assert.equal(
            await chrome.evaluate(
                () => document.querySelector('.references-widget__title-bar-group--active')!.textContent
            ),
            'This repository'
        )

        await chrome.wait('.references-widget__badge')
        await retry(
            async () =>
                assert.ok(
                    parseInt(
                        await chrome.evaluate<string>(
                            () => document.querySelector('.references-widget__badge')!.textContent
                        ),
                        10
                    ) > 0, // assert some (local) refs fetched
                    'expected some local references, got none'
                ),
            10 // additional retries since refs fetching can take a while
        )
    }

    const assertNonemptyExternalRefs = async (): Promise<void> => {
        // verify active group is 'external'
        await chrome.wait('.references-widget__title-bar-group--active')
        assert.equal(
            await chrome.evaluate(
                () => document.querySelector('.references-widget__title-bar-group--active')!.textContent
            ),
            'Other repositories'
        )

        await chrome.wait('.references-widget__badge')
        await retry(
            async () =>
                assert.ok(
                    parseInt(
                        await chrome.evaluate<string>(
                            () => document.querySelectorAll('.references-widget__badge')[1].textContent // get the external refs count
                        ),
                        10
                    ) > 0, // assert some external refs fetched
                    'expected some external references, got none'
                ),
            10 // additional retries since refs fetching can take a while
        )
    }

    describe('Theme switcher', () => {
        it('changes the theme when toggle is clicked', async () => {
            await chrome.goto(baseURL + '/github.com/gorilla/mux/-/blob/mux.go')
            await chrome.wait('.theme')
            const currentThemes = await chrome.evaluate<string[]>(() =>
                Array.from(document.querySelector('.theme')!.classList).filter(c => c.startsWith('theme-'))
            )
            assert.equal(currentThemes.length, 1, 'Expected 1 theme')
            const expectedTheme = currentThemes[0] === 'theme-dark' ? 'theme-light' : 'theme-dark'
            await chrome.click('.theme-switcher')
            assert.deepEqual(
                await chrome.evaluate<string>(() =>
                    Array.from(document.querySelector('.theme')!.classList).filter(c => c.startsWith('theme-'))
                ),
                [expectedTheme]
            )
        })
    })

    describe('Repository component', () => {
        const blobTableSelector = '.repository__viewer > code > table > tbody'
        const clickToken = async (line: number, spanOffset: number): Promise<void> => {
            await chrome.click(`${blobTableSelector} > tr:nth-child(${line}) > td.code > span:nth-child(${spanOffset})`)
        }

        const getTooltipDoc = async (): Promise<string> => {
            await chrome.wait('.tooltip__doc')
            return await chrome.evaluate<string>(() => document.querySelector('.tooltip__doc')!.textContent)
        }

        const tooltipActionsSelector = '.sg-tooltip > .tooltip__actions'
        const clickTooltipJ2D = async (): Promise<void> =>
            await chrome.click(`${tooltipActionsSelector} > a:nth-child(1)`)
        const clickTooltipFindRefs = async (): Promise<void> =>
            await chrome.click(`${tooltipActionsSelector} > a:nth-child(2)`)

        describe('file tree', () => {
            it('does navigation on file click', async () => {
                await chrome.goto(
                    baseURL + '/github.com/sourcegraph/godockerize@05bac79edd17c0f55127871fa9c6f4d91bebf07c'
                )
                await chrome.click(`[data-tree-path="godockerize.go"]`)
                await assertWindowLocation(
                    '/github.com/sourcegraph/godockerize@05bac79edd17c0f55127871fa9c6f4d91bebf07c/-/blob/godockerize.go'
                )
            })

            it('expands directory on row click (no navigation)', async () => {
                await chrome.goto(baseURL + '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d')
                await chrome.click('.tree__row-icon')
                await chrome.wait('.tree__row--selected [data-tree-path="websocket"]')
                await chrome.wait('.tree__row--expanded [data-tree-path="websocket"]')
                await assertWindowLocation('/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d')
            })

            it('does navigation on directory row click', async () => {
                await chrome.goto(baseURL + '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d')
                await chrome.click('.tree__row-label')
                await chrome.wait('.tree__row--selected [data-tree-path="websocket"]')
                await chrome.wait('.tree__row--expanded [data-tree-path="websocket"]')
                await assertWindowLocation(
                    '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d/-/tree/websocket'
                )
            })

            it('selects the current file', async () => {
                await chrome.goto(
                    baseURL +
                        '/github.com/sourcegraph/godockerize@05bac79edd17c0f55127871fa9c6f4d91bebf07c/-/blob/godockerize.go'
                )
                await chrome.wait('.tree__row--selected [data-tree-path="godockerize.go"]')
            })

            it('shows partial tree when opening directory', async () => {
                await chrome.goto(
                    baseURL +
                        '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d/-/tree/websocket'
                )
                await chrome.wait('.tree__row')
                assert.equal(await chrome.evaluate(() => document.querySelectorAll('.tree__row').length), 1)
            })

            it('responds to keyboard shortcuts', async () => {
                const assertNumRowsExpanded = async (expectedCount: number) => {
                    assert.equal(
                        await chrome.evaluate(() => document.querySelectorAll('.tree__row--expanded').length),
                        expectedCount
                    )
                }

                await chrome.goto(
                    baseURL +
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/.travis.yml'
                )
                await chrome.wait('.tree__row') // wait for tree to render

                await chrome.press(38) // arrow up to 'diff' directory
                await chrome.wait('.tree__row--selected [data-tree-path="diff"]')
                await chrome.press(39) // arrow right (expand 'diff' directory)
                await chrome.wait('.tree__row--selected [data-tree-path="diff"]')
                await chrome.wait('.tree__row--expanded [data-tree-path="diff"]')
                await chrome.press(39) // arrow right (move to nested 'diff/testdata' directory)
                await chrome.wait('.tree__row--selected [data-tree-path="diff/testdata"]')
                await assertNumRowsExpanded(1) // only `diff` directory is expanded, though `diff/testdata` is expanded

                await chrome.press(39) // arrow right (expand 'diff/testdata' directory)
                await chrome.wait('.tree__row--selected [data-tree-path="diff/testdata"]')
                await chrome.wait('.tree__row--expanded [data-tree-path="diff/testdata"]')
                await assertNumRowsExpanded(2) // `diff` and `diff/testdata` directories expanded

                // select some file nested under `diff/testdata`
                await chrome.press(40) // arrow down
                await chrome.press(40) // arrow down
                await chrome.press(40) // arrow down
                await chrome.press(40) // arrow down
                await chrome.wait('.tree__row--selected [data-tree-path="diff/testdata/empty_orig.diff"]')

                await chrome.press(37) // arrow left (navigate immediately up to parent directory `diff/testdata`)
                await chrome.wait('.tree__row--selected [data-tree-path="diff/testdata"]')
                await assertNumRowsExpanded(2) // `diff` and `diff/testdata` directories expanded

                await chrome.press(37) // arrow left
                await chrome.wait('.tree__row--selected [data-tree-path="diff/testdata"]') // `diff/testdata` still selected
                await assertNumRowsExpanded(1) // only `diff` directory expanded
            })
        })

        describe('directory page', () => {
            it('shows a row for each entry in the directory', async () => {
                await chrome.goto(baseURL + '/github.com/gorilla/securecookie@e59506cc896acb7f7bf732d4fdf5e25f7ccd8983')
                await chrome.wait('.dir-page-entry__row')
                await retry(async () =>
                    assert.equal(
                        await chrome.evaluate(() => document.querySelectorAll('.dir-page-entry__row').length),
                        8
                    )
                )
            })

            it('shows commit information on a row', async () => {
                await chrome.goto(baseURL + '/github.com/gorilla/securecookie@e59506cc896acb7f7bf732d4fdf5e25f7ccd8983')
                await chrome.wait('.dir-page-entry__commit-message-cell')
                await retry(async () =>
                    assert.equal(
                        await chrome.evaluate(
                            () => document.querySelector('.dir-page-entry__commit-message-cell')!.textContent
                        ),
                        'Add fuzz testing corpus.'
                    )
                )
                await chrome.wait('.dir-page-entry__committer-cell')
                await retry(async () =>
                    assert.equal(
                        await chrome.evaluate(
                            () => document.querySelector('.dir-page-entry__committer-cell')!.textContent
                        ),
                        'Kamil Kisiel'
                    )
                )
                await chrome.wait('.dir-page-entry__commit-hash-cell')
                await retry(async () =>
                    assert.equal(
                        await chrome.evaluate(
                            () => document.querySelector('.dir-page-entry__commit-hash-cell')!.textContent
                        ),
                        'c13558c'
                    )
                )
            })

            it('navigates when clicking on a row', async () => {
                await chrome.goto(baseURL + '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d')
                // click on directory
                await chrome.click('.dir-page__name-link')
                await assertWindowLocation(
                    '/github.com/sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d/-/tree/websocket'
                )
                // click on commit ID
                await chrome.click('.dir-page__commit-hash-link')
                await assertWindowLocation(
                    '/github.com/sourcegraph/jsonrpc2@6e06d561ec88594028846aa2a4af17a9aa0c87c4/-/tree/websocket'
                )
                // click on "up directory"
                await chrome.click('.dir-page__return-arrow-cell a')
                await assertWindowLocation('/github.com/sourcegraph/jsonrpc2@6e06d561ec88594028846aa2a4af17a9aa0c87c4')
            })
        })

        describe('rev resolution', () => {
            it('shows clone in progress interstitial page', async () => {
                await chrome.goto(baseURL + '/github.com/sourcegraphtest/AlwaysCloningTest')
                await chrome.wait('.hero-page__subtitle')
                await retry(async () =>
                    assert.equal(
                        await chrome.evaluate(() => document.querySelector('.hero-page__subtitle')!.textContent),
                        'Cloning in progress'
                    )
                )
            })

            it('resolves default branch when unspecified', async () => {
                await chrome.goto(baseURL + '/github.com/sourcegraph/go-diff/-/blob/diff/diff.go')
                await chrome.wait('.rev-switcher__input')
                await retry(async () =>
                    assert.equal(
                        await chrome.evaluate(
                            () => (document.querySelector('.rev-switcher__input') as HTMLInputElement).value
                        ),
                        'master'
                    )
                )
                // Verify file contents are loaded.
                await chrome.wait(blobTableSelector)
            })

            it('updates rev with switcher', async () => {
                await chrome.goto(baseURL + '/github.com/sourcegraph/checkup/-/blob/s3.go')
                await chrome.click('.rev-switcher__input')
                await chrome.click('.rev-switcher__rev[title="v0.1.0"]')
                await assertWindowLocation('/github.com/sourcegraph/checkup@v0.1.0/-/blob/s3.go')
            })
        })

        describe('tooltips', () => {
            it('gets displayed and updates URL when clicking on a token', async () => {
                await chrome.goto(
                    baseURL +
                        '/github.com/sourcegraph/godockerize@05bac79edd17c0f55127871fa9c6f4d91bebf07c/-/blob/godockerize.go'
                )
                await clickToken(23, 2)
                await getTooltipDoc() // verify there is a tooltip
                await assertWindowLocation(
                    '/github.com/sourcegraph/godockerize@05bac79edd17c0f55127871fa9c6f4d91bebf07c/-/blob/godockerize.go#L23:3'
                )
            })

            it('gets displayed when navigating to a URL with a token position', async () => {
                await chrome.goto(
                    baseURL +
                        '/github.com/sourcegraph/godockerize@05bac79edd17c0f55127871fa9c6f4d91bebf07c/-/blob/godockerize.go#L23:3'
                )
                await retry(async () =>
                    assert.equal(await getTooltipDoc(), `The name of the program. Defaults to path.Base(os.Args[0]) \n`)
                )
            })

            describe('jump to definition', () => {
                it('noops when on the definition', async () => {
                    await chrome.goto(
                        baseURL +
                            '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6'
                    )
                    await clickTooltipJ2D()
                    await assertWindowLocation(
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6'
                    )
                })

                it('does navigation (same repo, same file)', async () => {
                    await chrome.goto(
                        baseURL +
                            '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go'
                    )
                    await clickToken(25, 5)
                    await clickTooltipJ2D()
                    return await assertWindowLocation(
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6'
                    )
                })

                it('does navigation (same repo, different file)', async () => {
                    await chrome.goto(
                        baseURL +
                            '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/print.go#L13:31'
                    )
                    await clickTooltipJ2D()
                    await assertWindowLocation(
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/diff.pb.go#L38:6'
                    )
                    // Verify file tree is highlighting the new path.
                    return await chrome.wait('.tree__row--selected [data-tree-path="diff/diff.pb.go"]')
                })

                it('does navigation (external repo)', async () => {
                    await chrome.goto(
                        baseURL +
                            '/github.com/sourcegraph/vcsstore@267289226b15e5b03adedc9746317455be96e44c/-/blob/server/diff.go#L27:30'
                    )
                    await clickTooltipJ2D()
                    await assertWindowLocation(
                        '/github.com/sourcegraph/go-vcs@aa7c38442c17a3387b8a21f566788d8555afedd0/-/blob/vcs/repository.go#L103:6'
                    )
                })
            })

            describe('find references', () => {
                it('opens widget and fetches local references', async () => {
                    await chrome.goto(
                        baseURL +
                            '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6'
                    )
                    await clickTooltipFindRefs()
                    await assertWindowLocation(
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L29:6$references'
                    )

                    await assertNonemptyLocalRefs()

                    // verify the appropriate # of references are fetched
                    await chrome.wait('.references-widget__badge')
                    await retry(async () =>
                        assert.equal(
                            await chrome.evaluate(
                                () => document.querySelector('.references-widget__badge')!.textContent
                            ),
                            '5'
                        )
                    )

                    // verify all the matches highlight a `MultiFileDiffReader` token
                    await assertAllHighlightedTokens('MultiFileDiffReader')
                })

                // Testing external references on localhost is unreliable, since different dev environments will
                // not guarantee what repo(s) have been indexed. It's possible a developer has an environment with only
                // 1 repo, in which case there would never be external references. So we *only* run this test against
                // non-localhost servers.
                const test = baseURL === 'http://localhost:3080' ? it.skip : it
                test('opens widget and fetches external references', async () => {
                    await chrome.goto(
                        baseURL +
                            '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L32:16$references:external'
                    )

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
                await chrome.goto(baseURL + '/-/godoc/refs?def=Compare&pkg=bytes&repo=')
                await assertWindowLocationPrefix('/github.com/golang/go/-/blob/src/bytes/bytes_decl.go')
                await assertStickyHighlightedToken('Compare')
                await assertNonemptyLocalRefs()
                await assertAllHighlightedTokens('Compare')
            })

            it('resolves standard library function (from stdlib repo)', async () => {
                // https://godoc.org/github.com/golang/go/src/bytes#Compare
                await chrome.goto(
                    baseURL +
                        '/-/godoc/refs?def=Compare&pkg=github.com%2Fgolang%2Fgo%2Fsrc%2Fbytes&repo=github.com%2Fgolang%2Fgo'
                )
                await assertWindowLocationPrefix('/github.com/golang/go/-/blob/src/bytes/bytes_decl.go')
                await assertStickyHighlightedToken('Compare')
                await assertNonemptyLocalRefs()
                await assertAllHighlightedTokens('Compare')
            })

            it('resolves external package function (from gorilla/mux)', async () => {
                // https://godoc.org/github.com/gorilla/mux#Router
                await chrome.goto(
                    baseURL + '/-/godoc/refs?def=Router&pkg=github.com%2Fgorilla%2Fmux&repo=github.com%2Fgorilla%2Fmux'
                )
                await assertWindowLocationPrefix('/github.com/gorilla/mux/-/blob/mux.go')
                await assertStickyHighlightedToken('Router')
                await assertNonemptyLocalRefs()
                await assertAllHighlightedTokens('Router')
            })
        })

        describe('external code host links', () => {
            it('on line blame', async () => {
                await chrome.goto(
                    baseURL +
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L19'
                )
                await chrome.wait('.repository__viewer > code > table > tbody > tr:nth-child(19)')
                await chrome.evaluate(() => {
                    const row = document.querySelector('.repository__viewer > code > table > tbody > tr:nth-child(19)')!
                    const blame = row.querySelector('.blame')!
                    const rect = blame.getBoundingClientRect() as any
                    const ev = new MouseEvent('click', {
                        view: window,
                        bubbles: true,
                        cancelable: true,
                        clientX: rect.x + rect.width + 1,
                        clientY: rect.y + 1,
                    })
                    blame.dispatchEvent(ev)
                })
                await assertWindowLocation(
                    'https://github.com/sourcegraph/go-diff/commit/f93a4e38b36b4003edbfdff66d3b92c6cd977c1c',
                    true
                )
            })

            it('on repo navbar ("View on GitHub")', async () => {
                await chrome.goto(
                    baseURL +
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/parse.go#L19'
                )
                await chrome.wait('.repo-nav__action[title="View on GitHub"]')
                await retry(async () =>
                    assert.equal(
                        await chrome.evaluate(
                            () =>
                                (document.querySelector(
                                    '.repo-nav__action[title="View on GitHub"]'
                                ) as HTMLAnchorElement).href
                        ),
                        'https://github.com/sourcegraph/go-diff/blob/3f415a150aec0685cb81b73cc201e762e075006d/diff/parse.go#L19'
                    )
                )
            })
        })
    })

    describe('Search component', () => {
        it('renders results for sourcegraph/go-diff (no search group)', async () => {
            await chrome.goto(
                baseURL + '/search?q=diff+repo:sourcegraph/go-diff%403f415a150aec0685cb81b73cc201e762e075006d'
            )
            await chrome.wait('.search-results__header-stats')
            await retry(async () => {
                const label = await chrome.evaluate<string>(
                    () => document.querySelector('.search-results__header-stats')!.textContent
                )
                assert.equal(label.startsWith('361 results in'), true, 'incorrect number of search results')
            })
            // navigate to result on click
            await chrome.click('.file-match__item')
            await retry(async () =>
                assert.equal(
                    await chrome.evaluate(() => window.location.href),
                    baseURL +
                        '/github.com/sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d/-/blob/diff/testdata/sample_multi_file_single.diff#L1:1'
                )
            )
        })

        if (baseURL !== 'http://localhost:3080') {
            // TEMPORARY KLUDGE:
            // Currently the behavior of search is different on localhost vs. the dogfood server;
            // on localhost the repo groups are called `repogroup:sample *` while on dogfood they
            // are `repogroup:active *`.
            it.skip('renders results for sourcegraph/go-diff (w/ search group)', async () => {
                await chrome.goto(baseURL + '/search')

                // Update the input value
                await chrome.wait('input')
                await chrome.type('diff repo:sourcegraph/go-diff@3f415a150aec0685cb81b73cc201e762e075006d')

                // Update the select value
                await chrome.wait('select')
                await chrome.evaluate(() => {
                    const select = document.querySelector('select')!
                    select.value = '-file:(test|spec)'
                    select.dispatchEvent(new Event('change', { bubbles: true }))
                })

                // Submit the search
                await chrome.click('button')

                await chrome.wait('.search-results__header-stats')
                await retry(async () => {
                    const label = await chrome.evaluate<string>(
                        () => document.querySelector('.search-results__header-stats')!.textContent
                    )
                    assert.equal(label.startsWith('361 results in'), true, 'incorrect number of search results')
                })
            })
        } else {
            it('renders results for sourcegraph/go-diff (w/ search group)', async () => {
                await chrome.goto(baseURL + '/search')

                // Update the input value
                await chrome.wait('input')
                await chrome.type('test repo:sourcegraph/jsonrpc2@c6c7b9aa99fb76ee5460ccd3912ba35d419d493d')

                // Update the select value
                await chrome.wait('select')
                await chrome.evaluate(() => {
                    const select = document.querySelector('select')!
                    select.value = 'repogroup:sample file:(test|spec)'
                    select.dispatchEvent(new Event('change', { bubbles: true }))
                })

                // Submit the search
                await chrome.click('button')

                await chrome.wait('.search-results__header-stats')
                await retry(async () => {
                    const label = await chrome.evaluate<string>(
                        () => document.querySelector('.search-results__header-stats')!.textContent
                    )
                    assert.equal(label.startsWith('65 results in'), true, 'incorrect number of search results')
                })
            })
        }
    })
})
