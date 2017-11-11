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
        chrome = new Chromeless({ waitTimeout: 20000, launchChrome: false })
    })
    afterEach(() => chrome.end())
    after(() => {
        if (headlessChrome) {
            return headlessChrome.kill()
        }
    })

    const assertEventuallyEqual = async (expression: () => Promise<any>, value: any, numRetries = 3): Promise<any> => {
        while (true) {
            try {
                assert.deepEqual(await expression(), value)
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
            it('does navigation on click', async () => {
                await chrome.goto(baseURL + '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7')
                await chrome.click(`a[data-tree-path="mux.go"]`)
                await assertEventuallyEqual(
                    () => chrome.evaluate(() => window.location.href),
                    baseURL + '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go'
                )
            })

            it('selects the current file', async () => {
                await chrome.goto(
                    baseURL + '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go'
                )
                await chrome.wait('.tree__row--selected a[data-tree-path="mux.go"]')
            })
        })

        describe('rev resolution', () => {
            it('shows clone in progress interstitial page', async () => {
                await chrome.goto(baseURL + '/github.com/sourcegraphtest/AlwaysCloningTest')
                await chrome.wait('.hero-page__subtitle')
                await assertEventuallyEqual(
                    () => chrome.evaluate(() => document.querySelector('.hero-page__subtitle')!.textContent),
                    'Cloning in progress'
                )
            })

            it('resolves default branch when unspecified', async () => {
                await chrome.goto(baseURL + '/github.com/gorilla/mux/-/blob/mux.go')
                await chrome.wait('.rev-switcher__input')
                await assertEventuallyEqual(
                    () =>
                        chrome.evaluate(
                            () => (document.querySelector('.rev-switcher__input') as HTMLInputElement).value
                        ),
                    'master'
                )
                // Verify file contents are loaded.
                await chrome.wait(blobTableSelector)
            })

            it('updates rev with switcher', async () => {
                await chrome.goto(baseURL + '/github.com/gorilla/mux/-/blob/mux.go')
                await chrome.click('.rev-switcher__input')
                await chrome.click('.rev-switcher__rev[title="v1.1"]')
                await assertEventuallyEqual(
                    () => chrome.evaluate(() => window.location.href),
                    baseURL + '/github.com/gorilla/mux@v1.1/-/blob/mux.go'
                )
            })
        })

        describe('tooltips', () => {
            it('gets displayed and updates URL when clicking on a token', async () => {
                await chrome.goto(
                    baseURL + '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go'
                )
                await clickToken(21, 3)
                await getTooltipDoc() // verify there is a tooltip
                await assertEventuallyEqual(
                    () => chrome.evaluate(() => window.location.href),
                    baseURL + '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L21:6'
                )
            })

            it('gets displayed when navigating to a URL with a token position', async () => {
                await chrome.goto(
                    baseURL + '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L21:6'
                )
                await assertEventuallyEqual(getTooltipDoc, `NewRouter returns a new router instance. \n`)
            })

            describe('jump to definition', () => {
                it('noops when on the definition', async () => {
                    await chrome.goto(
                        baseURL + '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L21:6'
                    )
                    await clickTooltipJ2D()
                    await assertEventuallyEqual(
                        () => chrome.evaluate(() => window.location.href),
                        baseURL + '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L21:6'
                    )
                })

                it('does navigation (same repo, same file)', async () => {
                    await chrome.goto(
                        baseURL + '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go'
                    )
                    await clickToken(21, 8)
                    await clickTooltipJ2D()
                    await assertEventuallyEqual(
                        () => chrome.evaluate(() => window.location.href),
                        baseURL + '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L43:6'
                    )
                })

                it('does navigation (same repo, different file)', async () => {
                    await chrome.goto(
                        baseURL +
                            '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L22:47'
                    )
                    await clickTooltipJ2D()
                    await assertEventuallyEqual(
                        () => chrome.evaluate(() => window.location.href),
                        baseURL +
                            '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/route.go#L17:6'
                    )
                    // Verify file tree is highlighting the new path.
                    await chrome.wait('.tree__row--selected a[data-tree-path="route.go"]')
                })

                it('does navigation (external repo)', async () => {
                    await chrome.goto(
                        baseURL +
                            '/github.com/gorilla/sessions@a3acf13e802c358d65f249324d14ed24aac11370/-/blob/sessions.go#L134:10'
                    )
                    await clickTooltipJ2D()
                    await assertEventuallyEqual(
                        () => chrome.evaluate(() => window.location.href),
                        baseURL + '/github.com/gorilla/context@HEAD/-/blob/context.go#L20:6'
                    )
                })
            })

            describe('find references', () => {
                it('opens widget and fetches local references', async () => {
                    await chrome.goto(
                        baseURL +
                            '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L21:19'
                    )
                    await clickTooltipFindRefs()
                    await assertEventuallyEqual(
                        () => chrome.evaluate(() => window.location.href),
                        baseURL +
                            '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L21:19$references'
                    )
                    await chrome.wait('.references-widget__badge')
                    await assertEventuallyEqual(
                        () => chrome.evaluate(() => document.querySelector('.references-widget__badge')!.textContent),
                        '45'
                    )
                })
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
                await assertEventuallyEqual(
                    () => chrome.evaluate(() => window.location.href),
                    'https://github.com/gorilla/mux/commit/eac83ba2c004bb759a2875b1f1dbb032adf8bb4a'
                )
            })

            it('on repo navbar ("View on GitHub")', async () => {
                await chrome.goto(
                    baseURL + '/github.com/gorilla/mux@24fca303ac6da784b9e8269f724ddeb0b2eea5e7/-/blob/mux.go#L43:6'
                )
                await chrome.wait('.repo-nav__action[title="View on GitHub"]')
                await assertEventuallyEqual(
                    () =>
                        chrome.evaluate(
                            () =>
                                (document.querySelector(
                                    '.repo-nav__action[title="View on GitHub"]'
                                ) as HTMLAnchorElement).href
                        ),
                    'https://github.com/gorilla/mux/blob/24fca303ac6da784b9e8269f724ddeb0b2eea5e7/mux.go#L43'
                )
            })
        })
    })

    describe('Search component', () => {
        it('renders results for gorilla/mux (no search group)', async () => {
            await chrome.goto(
                baseURL + '/search?q=router+repo:gorilla/mux%40eac83ba2c004bb759a2875b1f1dbb032adf8bb4a&sq='
            )
            await chrome.wait('.search-results2__badge')
            await assertEventuallyEqual(
                () => chrome.evaluate(() => document.querySelectorAll('.search-results2__badge')[0].textContent),
                '126' // # results
            )
            await assertEventuallyEqual(
                () => chrome.evaluate(() => document.querySelectorAll('.search-results2__badge')[1].textContent),
                '7' // # files
            )
            await assertEventuallyEqual(
                () => chrome.evaluate(() => document.querySelectorAll('.search-results2__badge')[2].textContent),
                '1' // # repos
            )
            // navigate to result on click
            await chrome.click('.references-group__reference')
            await assertEventuallyEqual(
                () => chrome.evaluate(() => window.location.href),
                baseURL + '/github.com/gorilla/mux@eac83ba2c004bb759a2875b1f1dbb032adf8bb4a/-/blob/route.go#L17:46'
            )
        })

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

            await chrome.wait('.search-results2__badge')
            await assertEventuallyEqual(
                () => chrome.evaluate(() => document.querySelectorAll('.search-results2__badge')[0].textContent),
                '91' // # results
            )
            await assertEventuallyEqual(
                () => chrome.evaluate(() => document.querySelectorAll('.search-results2__badge')[1].textContent),
                '4' // # files
            )
            await assertEventuallyEqual(
                () => chrome.evaluate(() => document.querySelectorAll('.search-results2__badge')[2].textContent),
                '1' // # repos
            )
        })
    })
})
