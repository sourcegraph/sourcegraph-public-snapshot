import * as assert from 'assert'
import { Chromeless } from 'chromeless'
// tslint:disable-next-line
import * as _ from 'lodash'

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
        const highlightedTokens = JSON.parse(
            await chrome.evaluate<string>(() =>
                JSON.stringify(Array.from(document.querySelectorAll('.selection-highlight')).map(el => el.textContent))
            )
        )
        assert.ok(
            _.every(highlightedTokens, txt => txt === label),
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
        await retry(async () =>
            assert.ok(
                parseInt(
                    await chrome.evaluate<string>(
                        () => document.querySelector('.references-widget__badge')!.textContent
                    ),
                    10
                ) > 0, // assert some (local) refs fetched
                'expected some local references, got none'
            )
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
        await retry(async () =>
            assert.ok(
                parseInt(
                    await chrome.evaluate<string>(
                        () => document.querySelectorAll('.references-widget__badge')[1].textContent // get the external refs count
                    ),
                    10
                ) > 0, // assert some external refs fetched
                'expected some external references, got none'
            )
        )
    }

    describe('Repository component', () => {
        const blobTableSelector = '.repository__viewer > div > table > tbody'
        const clickToken = async (line: number, spanOffset: number): Promise<void> => {
            await chrome.click(`${blobTableSelector} > tr:nth-child(${line}) > td.code > span:nth-child(${spanOffset})`)
        }

        const getTooltipDoc = async (): Promise<string> => {
            await chrome.wait('.tooltip__doc')
            return await chrome.evaluate<string>(() => document.querySelector('.tooltip__doc')!.textContent)
        }

        const tooltipActionsSelector = '.sg-tooltip > .tooltip__actions'
        const clickTooltipJ2D = async (): Promise<void> => {
            await chrome.click(`${tooltipActionsSelector} > a:nth-child(1)`)
        }
        const clickTooltipFindRefs = async (): Promise<void> => {
            await chrome.click(`${tooltipActionsSelector} > a:nth-child(2)`)
        }

        describe('file tree', () => {
            it('does navigation on file click', async () => {
                await chrome.goto(baseURL + '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7')
                await chrome.click(`[data-tree-path="mux.go"]`)
                await assertWindowLocation(
                    '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go'
                )
            })

            it('does navigation on directory expander click', async () => {
                await chrome.goto(baseURL + '/github.com/gorilla/securecookie@e59506cc896acb7f7bf732d4fdf5e25f7ccd8983')
                await chrome.click('.tree__row-icon')
                await chrome.wait('.tree__row--selected [data-tree-path="fuzz"]')
                await chrome.wait('.tree__row--expanded [data-tree-path="fuzz"]')
                await assertWindowLocation(
                    '/github.com/gorilla/securecookie@e59506cc896acb7f7bf732d4fdf5e25f7ccd8983/-/tree/fuzz'
                )
            })

            it('expands directory on row click (no navigation)', async () => {
                await chrome.goto(baseURL + '/github.com/gorilla/securecookie@e59506cc896acb7f7bf732d4fdf5e25f7ccd8983')
                await chrome.click('.tree__row-label')
                await chrome.wait('.tree__row--selected [data-tree-path="fuzz"]')
                await chrome.wait('.tree__row--expanded [data-tree-path="fuzz"]')
                await assertWindowLocation('/github.com/gorilla/securecookie@e59506cc896acb7f7bf732d4fdf5e25f7ccd8983')
            })

            it('selects the current file', async () => {
                await chrome.goto(
                    baseURL + '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go'
                )
                await chrome.wait('.tree__row--selected [data-tree-path="mux.go"]')
            })

            it('selects and expands the current directory', async () => {
                await chrome.goto(
                    baseURL + '/github.com/gorilla/securecookie@e59506cc896acb7f7bf732d4fdf5e25f7ccd8983/-/tree/fuzz'
                )
                await chrome.wait('.tree__row--selected [data-tree-path="fuzz"]')
                await chrome.wait('.tree__row--expanded [data-tree-path="fuzz"]')
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
                        '/github.com/gorilla/securecookie@e59506cc896acb7f7bf732d4fdf5e25f7ccd8983/-/blob/.travis.yml'
                )
                await chrome.wait('.tree__row') // wait for tree to render

                await chrome.press(38) // arrow up to 'fuzz' directory
                await chrome.wait('.tree__row--selected [data-tree-path="fuzz"]')
                await chrome.press(39) // arrow right (expand 'fuzz' directory)
                await chrome.wait('.tree__row--selected [data-tree-path="fuzz"]')
                await chrome.wait('.tree__row--expanded [data-tree-path="fuzz"]')
                await chrome.press(39) // arrow right (move to nested 'fuzz/corpus' directory)
                await chrome.wait('.tree__row--selected [data-tree-path="fuzz/corpus"]')
                await assertNumRowsExpanded(1) // only `fuzz` directory is expanded, though `fuzz/corpus` is expanded

                await chrome.press(39) // arrow right (expand 'fuzz/corpus' directory)
                await chrome.wait('.tree__row--selected [data-tree-path="fuzz/corpus"]')
                await chrome.wait('.tree__row--expanded [data-tree-path="fuzz/corpus"]')
                await assertNumRowsExpanded(2) // `fuzz` and `fuzz/corpus` directories expanded

                // select some file nested under `fuzz / corpus`
                await chrome.press(40) // arrow down
                await chrome.press(40) // arrow down
                await chrome.press(40) // arrow down
                await chrome.press(40) // arrow down
                await chrome.wait('.tree__row--selected [data-tree-path="fuzz/corpus/1.sc"]')

                await chrome.press(37) // arrow left (navigate immediately up to parent directory `fuzz / corpus`)
                await chrome.wait('.tree__row--selected [data-tree-path="fuzz/corpus"]')
                await assertNumRowsExpanded(2) // `fuzz` and `fuzz / corpus` directories expanded

                await chrome.press(37) // arrow left
                await chrome.wait('.tree__row--selected [data-tree-path="fuzz/corpus"]') // `fuzz / corpus` still selected
                await assertNumRowsExpanded(1) // only `fuzz` directory expanded
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
                await chrome.goto(baseURL + '/github.com/gorilla/securecookie@e59506cc896acb7f7bf732d4fdf5e25f7ccd8983')
                // click on directory
                await chrome.click('.dir-page__name-link')
                await assertWindowLocation(
                    '/github.com/gorilla/securecookie@e59506cc896acb7f7bf732d4fdf5e25f7ccd8983/-/tree/fuzz'
                )
                // click on commit ID
                await chrome.click('.dir-page__commit-hash-link')
                await assertWindowLocation(
                    '/github.com/gorilla/securecookie@c13558c2b1c44da35e0eb043053609a5ba3a1f19/-/tree/fuzz'
                )
                // click on "up directory"
                await chrome.click('.dir-page__return-arrow-cell a')
                await assertWindowLocation('/github.com/gorilla/securecookie@c13558c2b1c44da35e0eb043053609a5ba3a1f19')
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
                await chrome.goto(baseURL + '/github.com/gorilla/mux/-/blob/mux.go')
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
                await chrome.goto(baseURL + '/github.com/gorilla/mux/-/blob/mux.go')
                await chrome.click('.rev-switcher__input')
                await chrome.click('.rev-switcher__rev[title="v1.1"]')
                await assertWindowLocation('/github.com/gorilla/mux@v1.1/-/blob/mux.go')
            })
        })

        describe('tooltips', () => {
            it('gets displayed and updates URL when clicking on a token', async () => {
                await chrome.goto(
                    baseURL + '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go'
                )
                await clickToken(21, 3)
                await getTooltipDoc() // verify there is a tooltip
                await assertWindowLocation(
                    '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L21:6'
                )
            })

            it('gets displayed when navigating to a URL with a token position', async () => {
                await chrome.goto(
                    baseURL + '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L21:6'
                )
                await retry(async () =>
                    assert.equal(await getTooltipDoc(), `NewRouter returns a new router instance. \n`)
                )
            })

            describe('jump to definition', () => {
                it('noops when on the definition', async () => {
                    await chrome.goto(
                        baseURL + '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L21:6'
                    )
                    await clickTooltipJ2D()
                    await assertWindowLocation(
                        '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L21:6'
                    )
                })

                it('does navigation (same repo, same file)', async () => {
                    await chrome.goto(
                        baseURL + '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go'
                    )
                    await clickToken(21, 8)
                    await clickTooltipJ2D()
                    await assertWindowLocation(
                        '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L43:6'
                    )
                })

                it('does navigation (same repo, different file)', async () => {
                    await chrome.goto(
                        baseURL +
                            '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L22:47'
                    )
                    await clickTooltipJ2D()
                    await assertWindowLocation(
                        '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/route.go#L17:6'
                    )
                    // Verify file tree is highlighting the new path.
                    await chrome.wait('.tree__row--selected [data-tree-path="route.go"]')
                })

                it('does navigation (external repo)', async () => {
                    await chrome.goto(
                        baseURL +
                            '/github.com/gorilla/sessions@a3acf13e802c358d65f249324d14ed24aac11370/-/blob/sessions.go#L134:10'
                    )
                    await clickTooltipJ2D()
                    await assertWindowLocation('/github.com/gorilla/context@HEAD/-/blob/context.go#L20:6')
                })
            })

            describe('find references', () => {
                it('opens widget and fetches local references', async () => {
                    await chrome.goto(
                        baseURL +
                            '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L21:19'
                    )
                    await clickTooltipFindRefs()
                    await assertWindowLocation(
                        '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L21:19$references'
                    )

                    await assertNonemptyLocalRefs()

                    // verify the appropriate # of references are fetched
                    await chrome.wait('.references-widget__badge')
                    await retry(async () =>
                        assert.equal(
                            await chrome.evaluate(
                                () => document.querySelector('.references-widget__badge')!.textContent
                            ),
                            '45'
                        )
                    )

                    // verify all the matches highlight a `Router` token
                    await assertAllHighlightedTokens('Router')
                })

                // Testing external references on localhost is unreliable, since different dev environments will
                // not guarantee what repo(s) have been indexed. It's possible a developer has an environment with only
                // 1 repo, in which case there would never be external references. So we *only* run this test against
                // non-localhost servers.
                const test = baseURL === 'http://localhost:3080' ? it.skip : it
                test('opens widget and fetches external references', async () => {
                    await chrome.goto(
                        baseURL +
                            '/github.com/gorilla/mux@2d5fef06b891c971b14aa6f71ca5ab6c03a36e0e/-/blob/mux.go#L43:6$references:external'
                    )

                    // verify some external refs are fetched (we cannot assert how many, but we can check that the matched results
                    // look like they're for the appropriate token)
                    await assertNonemptyExternalRefs()

                    // verify all the matches highlight a `Router` token
                    await assertAllHighlightedTokens('Router')
                })
            })
        })

        describe('godoc.org "Uses" links', () => {
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
                    baseURL + '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L43:6'
                )
                await chrome.wait('.repository__viewer > div > table > tbody > tr:nth-child(43)')
                await chrome.evaluate(() => {
                    const row = document.querySelector('.repository__viewer > div > table > tbody > tr:nth-child(43)')!
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
                    'https://github.com/gorilla/mux/commit/eac83ba2c004bb759a2875b1f1dbb032adf8bb4a',
                    true
                )
            })

            it('on repo navbar ("View on GitHub")', async () => {
                await chrome.goto(
                    baseURL + '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L43:6'
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
                        'https://github.com/gorilla/mux/blob/24fca303ac6da784b9e8269f724ddeb0b2eea5e7/mux.go#L43'
                    )
                )
            })
        })
    })

    describe('Search component', () => {
        it('renders results for gorilla/mux (no search group)', async () => {
            await chrome.goto(
                baseURL + '/search?q=router+repo:gorilla/mux%40eac83ba2c004bb759a2875b1f1dbb032adf8bb4a&sq='
            )
            await chrome.wait('.search-results2__stats')
            await retry(async () =>
                assert.equal(
                    await chrome.evaluate(() => document.querySelector('.search-results2__stats')!.textContent),
                    '126 results in'
                )
            )
            // navigate to result on click
            await chrome.click('.file-match__item')
            await retry(async () =>
                assert.equal(
                    await chrome.evaluate(() => window.location.href),
                    baseURL + '/github.com/gorilla/mux@eac83ba2c004bb759a2875b1f1dbb032adf8bb4a/-/blob/route.go#L17:46'
                )
            )
        })

        if (baseURL !== 'http://localhost:3080') {
            // TEMPORARY KLUDGE:
            // Currently the behavior of search is different on localhost vs. the dogfood server;
            // on localhost the repo groups are called `repogroup:sample * ` while on dogfood they
            // are `repogroup:active * `.
            it('renders results for gorilla/mux (w/ search group)', async () => {
                await chrome.goto(baseURL + '/search')

                // Update the input value
                await chrome.wait('input')
                await chrome.type('router repo:gorilla/mux@eac83ba2c004bb759a2875b1f1dbb032adf8bb4a')

                // Update the select value
                await chrome.wait('select')
                await chrome.evaluate(() => {
                    const select = document.querySelector('select')!
                    select.value = 'repogroup:active -file:(test|spec)'
                    select.dispatchEvent(new Event('change', { bubbles: true }))
                })

                // Submit the search
                await chrome.click('button')

                await chrome.wait('.search-results2__stats')
                await retry(async () =>
                    assert.equal(
                        await chrome.evaluate(() => document.querySelector('.search-results2__stats')!.textContent),
                        '91 results in'
                    )
                )
            })
        } else {
            it('renders results for gorilla/mux (w/ search group)', async () => {
                await chrome.goto(baseURL + '/search')

                // Update the input value
                await chrome.wait('input')
                await chrome.type('router repo:gorilla/mux@eac83ba2c004bb759a2875b1f1dbb032adf8bb4a')

                // Update the select value
                await chrome.wait('select')
                await chrome.evaluate(() => {
                    const select = document.querySelector('select')!
                    select.value = 'repogroup:sample -file:(test|spec)'
                    select.dispatchEvent(new Event('change', { bubbles: true }))
                })

                // Submit the search
                await chrome.click('button')

                await chrome.wait('.search-results2__stats')
                await retry(async () =>
                    assert.equal(
                        await chrome.evaluate(() => document.querySelector('.search-results2__stats')!.textContent),
                        '91 results in'
                    )
                )
            })
        }
    })
})
