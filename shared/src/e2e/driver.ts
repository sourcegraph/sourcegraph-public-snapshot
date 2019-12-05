import { percySnapshot as realPercySnapshot } from '@percy/puppeteer'
import * as jsonc from '@sqs/jsonc-parser'
import * as jsoncEdit from '@sqs/jsonc-parser/lib/edit'
import * as os from 'os'
import puppeteer, { PageEventObj, Page, Serializable, LaunchOptions, PageFnOptions } from 'puppeteer'
import { Key } from 'ts-key-enum'
import * as util from 'util'
import { dataOrThrowErrors, gql, GraphQLResult } from '../graphql/graphql'
import { IMutation, IQuery, ExternalServiceKind } from '../graphql/schema'
import { readEnvBoolean, retry } from './e2e-test-utils'
import * as path from 'path'
import { escapeRegExp } from 'lodash'
import { readFile } from 'mz/fs'

/**
 * Returns a Promise for the next emission of the given event on the given Puppeteer page.
 */
export const oncePageEvent = <E extends keyof PageEventObj>(page: Page, eventName: E): Promise<PageEventObj[E]> =>
    new Promise(resolve => page.once(eventName, resolve))

export const percySnapshot = readEnvBoolean({ variable: 'PERCY_ON', defaultValue: false })
    ? realPercySnapshot
    : () => Promise.resolve()

export const BROWSER_EXTENSION_DEV_ID = 'bmfbcejdknlknpncfpeloejonjoledha'

/**
 * Specifies how to select the content of the element. No
 * single method works in all cases:
 *
 * - Meta+A doesn't work in input boxes https://github.com/GoogleChrome/puppeteer/issues/1313
 * - selectall doesn't work in the Monaco editor
 */
type SelectTextMethod = 'selectall' | 'keyboard'

/**
 * Specifies how to enter text. Typing is preferred in cases where it's important to test
 * the process of manually typing out the text to enter. Pasting is preferred in cases
 * where typing would be too slow or we explicitly want to test paste behavior.
 */
type EnterTextMethod = 'type' | 'paste'

interface FindElementOptions {
    /**
     * Filter candidate elements to those with the specified CSS selector
     */
    selector?: string

    /**
     * Log the XPath quer(y|ies) used to find the element.
     */
    log?: boolean

    /**
     * Specifies how exact the search criterion is.
     */
    fuzziness?: 'exact' | 'prefix' | 'space-prefix' | 'contains'
}

function findElementRegexpStrings(
    text: string,
    { fuzziness = 'space-prefix' }: Pick<FindElementOptions, 'fuzziness'>
): string[] {
    //  Escape regexp special chars. Copied from
    //  https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Regular_Expressions
    const escapedText = escapeRegExp(text)
    const regexps = [`^${escapedText}$`]
    if (fuzziness === 'exact') {
        return regexps
    }
    regexps.push(`^${escapedText}\\b`)
    if (fuzziness === 'prefix') {
        return regexps
    }
    regexps.push(`^\\s+${escapedText}$`) // Still prefer exact
    regexps.push(`^\\s+${escapedText}\\b`)
    if (fuzziness === 'space-prefix') {
        return regexps
    }
    regexps.push(escapedText)
    return regexps
}

function findElementMatchingRegexps(tag: string, regexps: string[]): HTMLElement | null {
    for (const regexpString of regexps) {
        const regexp = new RegExp(regexpString)
        for (const el of document.querySelectorAll<HTMLElement>(tag)) {
            if (el.innerText && el.innerText.match(regexp)) {
                return el
            }
        }
    }
    return null
}

function getDebugExpressionFromRegexp(tag: string, regexp: string): string {
    return `Array.from(document.querySelectorAll(${JSON.stringify(
        tag
    )})).filter(e => e.innerText && e.innerText.match(/${regexp}/))`
}

export class Driver {
    constructor(
        public browser: puppeteer.Browser,
        public page: puppeteer.Page,
        public sourcegraphBaseUrl: string,
        public keepBrowser?: boolean
    ) {}

    public async ensureLoggedIn({
        username,
        password,
        email,
    }: {
        username: string
        password: string
        email?: string
    }): Promise<void> {
        await this.page.goto(this.sourcegraphBaseUrl)
        await this.page.evaluate(() => {
            localStorage.setItem('has-dismissed-browser-ext-toast', 'true')
            localStorage.setItem('has-dismissed-integrations-toast', 'true')
            localStorage.setItem('has-dismissed-survey-toast', 'true')
        })
        const url = new URL(this.page.url())
        if (url.pathname === '/site-admin/init') {
            await this.page.waitForSelector('.e2e-signup-form')
            if (email) {
                await this.page.type('input[name=email]', email)
            }
            await this.page.type('input[name=username]', username)
            await this.page.type('input[name=password]', password)
            await this.page.click('button[type=submit]')
            await this.page.waitForNavigation({ timeout: 3 * 1000 })
        } else if (url.pathname === '/sign-in') {
            await this.page.waitForSelector('.e2e-signin-form')
            await this.page.type('input', username)
            await this.page.type('input[name=password]', password)
            await this.page.click('button[type=submit]')
            await this.page.waitForNavigation({ timeout: 3 * 1000 })
        }
    }

    /**
     * Navigates to the Sourcegraph browser extension options page and sets the sourcegraph URL.
     */
    public async setExtensionSourcegraphUrl(): Promise<void> {
        await this.page.goto(`chrome-extension://${BROWSER_EXTENSION_DEV_ID}/options.html`)
        await this.page.waitForSelector('.e2e-sourcegraph-url')
        await this.replaceText({ selector: '.e2e-sourcegraph-url', newText: this.sourcegraphBaseUrl })
        await this.page.keyboard.press(Key.Enter)
        await this.page.waitForFunction(
            () => {
                const element = document.querySelector('.e2e-connection-status')
                return element?.textContent?.includes('Connected')
            },
            { timeout: 5000 }
        )
    }

    public async close(): Promise<void> {
        if (!this.keepBrowser) {
            await this.browser.close()
        }
    }

    public async newPage(): Promise<void> {
        this.page = await this.browser.newPage()
    }

    public async selectAll(method: SelectTextMethod = 'selectall'): Promise<void> {
        switch (method) {
            case 'selectall': {
                await this.page.evaluate(() => document.execCommand('selectall', false))
                break
            }
            case 'keyboard': {
                const modifier = os.platform() === 'darwin' ? Key.Meta : Key.Control
                await this.page.keyboard.down(modifier)
                await this.page.keyboard.press('a')
                await this.page.keyboard.up(modifier)
                break
            }
        }
    }

    public async enterText(method: EnterTextMethod = 'type', text: string): Promise<void> {
        // Pasting does not work on macOS. See:  https://github.com/GoogleChrome/puppeteer/issues/1313
        method = os.platform() === 'darwin' ? 'type' : method
        switch (method) {
            case 'type':
                await this.page.keyboard.type(text)
                break
            case 'paste':
                await this.paste(text)
                break
        }
    }

    public async replaceText({
        selector,
        newText,
        selectMethod = 'selectall',
        enterTextMethod = 'type',
    }: {
        selector: string
        newText: string
        selectMethod?: SelectTextMethod
        enterTextMethod?: EnterTextMethod
    }): Promise<void> {
        // The Monaco editor sometimes detaches nodes from the DOM, causing
        // `click()` to fail unpredictably.
        await retry(async () => {
            await this.page.waitForSelector(selector)
            await this.page.click(selector)
        })
        await this.selectAll(selectMethod)
        await this.page.keyboard.press(Key.Backspace)
        await this.enterText(enterTextMethod, newText)
    }

    public async acceptNextDialog(): Promise<void> {
        const dialog = await oncePageEvent(this.page, 'dialog')
        await dialog.accept()
    }

    public async ensureHasExternalService({
        kind,
        displayName,
        config,
        ensureRepos,
    }: {
        kind: ExternalServiceKind
        displayName: string
        config: string
        ensureRepos?: string[]
    }): Promise<void> {
        await this.page.goto(this.sourcegraphBaseUrl + '/site-admin/external-services')
        await this.page.waitFor('.e2e-filtered-connection')
        await this.page.waitForSelector('.e2e-filtered-connection__loader', { hidden: true })

        // Matches buttons for deleting external services named ${displayName}.
        const deleteButtonSelector = `[data-e2e-external-service-name="${displayName}"] .e2e-delete-external-service-button`
        if (await this.page.$(deleteButtonSelector)) {
            await Promise.all([this.acceptNextDialog(), this.page.click(deleteButtonSelector)])
        }

        await (await this.page.waitForSelector('.e2e-goto-add-external-service-page', { visible: true })).click()

        await (
            await this.page.waitForSelector(`[data-e2e-external-service-card-link="${kind.toUpperCase()}"]`, {
                visible: true,
            })
        ).click()

        await this.replaceText({
            selector: '#e2e-external-service-form-display-name',
            newText: displayName,
        })

        // Type in a new external service configuration.
        await this.replaceText({
            selector: '.view-line',
            newText: config,
            selectMethod: 'keyboard',
        })
        await Promise.all([this.page.waitForNavigation(), this.page.click('.e2e-add-external-service-button')])

        if (ensureRepos) {
            // Clone the repositories
            for (const slug of ensureRepos) {
                await this.page.goto(
                    this.sourcegraphBaseUrl + `/site-admin/repositories?query=${encodeURIComponent(slug)}`
                )
                await this.page.waitForSelector(`.repository-node[data-e2e-repository='${slug}']`, { visible: true })
                // Workaround for https://github.com/sourcegraph/sourcegraph/issues/5286
                await this.page.goto(`${this.sourcegraphBaseUrl}/${slug}`)
            }
        }
    }

    public async paste(value: string): Promise<void> {
        await this.page.evaluate(
            async d => {
                await navigator.clipboard.writeText(d.value)
            },
            { value }
        )
        const modifier = os.platform() === 'darwin' ? Key.Meta : Key.Control
        await this.page.keyboard.down(modifier)
        await this.page.keyboard.press('v')
        await this.page.keyboard.up(modifier)
    }

    public async assertWindowLocation(location: string, isAbsolute = false): Promise<any> {
        const url = isAbsolute ? location : this.sourcegraphBaseUrl + location
        await retry(async () => {
            expect(await this.page.evaluate(() => window.location.href)).toEqual(url)
        })
    }

    public async assertWindowLocationPrefix(locationPrefix: string, isAbsolute = false): Promise<any> {
        const prefix = isAbsolute ? locationPrefix : this.sourcegraphBaseUrl + locationPrefix
        await retry(async () => {
            const loc: string = await this.page.evaluate(() => window.location.href)
            expect(loc.startsWith(prefix)).toBeTruthy()
        })
    }

    public async assertStickyHighlightedToken(label: string): Promise<void> {
        await this.page.waitForSelector('.selection-highlight-sticky', { visible: true }) // make sure matched token is highlighted
        await retry(async () =>
            expect(
                await this.page.evaluate(() => document.querySelector('.selection-highlight-sticky')!.textContent)
            ).toEqual(label)
        )
    }

    public async assertAllHighlightedTokens(label: string): Promise<void> {
        const highlightedTokens = await this.page.evaluate(() =>
            Array.from(document.querySelectorAll('.selection-highlight')).map(el => el.textContent || '')
        )
        expect(highlightedTokens.every(txt => txt === label)).toBeTruthy()
    }

    public async assertNonemptyLocalRefs(): Promise<void> {
        // verify active group is references
        await this.page.waitForXPath(
            "//*[contains(@class, 'panel__tabs')]//*[contains(@class, 'tab-bar__tab--active') and contains(text(), 'References')]"
        )
        // verify there are some references
        await this.page.waitForSelector('.panel__tabs-content .file-match-children__item', { visible: true })
    }

    public async assertNonemptyExternalRefs(): Promise<void> {
        // verify active group is references
        await this.page.waitForXPath(
            "//*[contains(@class, 'panel__tabs')]//*[contains(@class, 'tab-bar__tab--active') and contains(text(), 'References')]"
        )
        // verify there are some references
        await this.page.waitForSelector('.panel__tabs-content .hierarchical-locations-view__item', { visible: true })
    }

    private async makeRequest<T = void>({ url, init }: { url: string; init: RequestInit & Serializable }): Promise<T> {
        const handle = await this.page.evaluateHandle((url, init) => fetch(url, init).then(r => r.json()), url, init)
        return handle.jsonValue()
    }

    private async makeGraphQLRequest<T extends IQuery | IMutation>({
        request,
        variables,
    }: {
        request: string
        variables: {}
    }): Promise<GraphQLResult<T>> {
        const nameMatch = request.match(/^\s*(?:query|mutation)\s+(\w+)/)
        const xhrHeaders =
            (await this.page.evaluate(
                sourcegraphBaseUrl =>
                    location.href.startsWith(sourcegraphBaseUrl) && (window as any).context.xhrHeaders,
                this.sourcegraphBaseUrl
            )) || {}
        const response = await this.makeRequest<GraphQLResult<T>>({
            url: `${this.sourcegraphBaseUrl}/.api/graphql${nameMatch ? '?' + nameMatch[1] : ''}`,
            init: {
                method: 'POST',
                body: JSON.stringify({ query: request, variables }),
                headers: {
                    ...xhrHeaders,
                    Accept: 'application/json',
                    'Content-Type': 'application/json',
                },
            },
        })
        return response
    }

    public async setConfig(path: jsonc.JSONPath, f: (oldValue: jsonc.Node | undefined) => any): Promise<void> {
        const currentConfigResponse = await this.makeGraphQLRequest<IQuery>({
            request: gql`
                query Site {
                    site {
                        id
                        configuration {
                            id
                            effectiveContents
                            validationMessages
                        }
                    }
                }
            `,
            variables: {},
        })
        const { site } = dataOrThrowErrors(currentConfigResponse)
        const currentConfig = site.configuration.effectiveContents
        const newConfig = modifyJSONC(currentConfig, path, f)
        const updateConfigResponse = await this.makeGraphQLRequest<IMutation>({
            request: gql`
                mutation UpdateSiteConfiguration($lastID: Int!, $input: String!) {
                    updateSiteConfiguration(lastID: $lastID, input: $input)
                }
            `,
            variables: { lastID: site.configuration.id, input: newConfig },
        })
        dataOrThrowErrors(updateConfigResponse)
    }

    public async ensureHasCORSOrigin({ corsOriginURL }: { corsOriginURL: string }): Promise<void> {
        await this.setConfig(['corsOrigin'], oldCorsOrigin => {
            const urls = oldCorsOrigin ? oldCorsOrigin.value.split(' ') : []
            return (urls.includes(corsOriginURL) ? urls : [...urls, corsOriginURL]).join(' ')
        })
    }

    public async resetUserSettings(): Promise<void> {
        const currentSettingsResponse = await this.makeGraphQLRequest<IQuery>({
            request: gql`
                query UserSettings {
                    currentUser {
                        id
                        settingsCascade {
                            subjects {
                                latestSettings {
                                    id
                                    contents
                                }
                            }
                        }
                    }
                }
            `,
            variables: {},
        })

        const { currentUser } = dataOrThrowErrors(currentSettingsResponse)

        if (currentUser?.settingsCascade) {
            const emptySettings = '{}'
            const [{ latestSettings }] = currentUser.settingsCascade.subjects.slice(-1)

            if (latestSettings && latestSettings.contents !== emptySettings) {
                const updateConfigResponse = await this.makeGraphQLRequest<IMutation>({
                    request: gql`
                        mutation OverwriteSettings($subject: ID!, $lastID: Int, $contents: String!) {
                            settingsMutation(input: { subject: $subject, lastID: $lastID }) {
                                overwriteSettings(contents: $contents) {
                                    empty {
                                        alwaysNil
                                    }
                                }
                            }
                        }
                    `,
                    variables: {
                        contents: emptySettings,
                        subject: currentUser.id,
                        lastID: latestSettings.id,
                    },
                })
                dataOrThrowErrors(updateConfigResponse)
            }
        }
    }

    /**
     * Finds the first HTML element matching the text using the regular expressions returned in
     * {@link findElementRegexpStrings}.
     *
     * @param options specifies additional parameters for finding the element. If you want to wait
     * until the element appears, specify the `wait` field (which can contain additional inner
     * options for how long to wait).
     */
    public async findElementWithText(
        text: string,
        options: FindElementOptions & { wait?: PageFnOptions | boolean } = {}
    ): Promise<puppeteer.ElementHandle<Element>> {
        const { selector: tagName, fuzziness, wait } = options
        const tag = tagName || '*'
        const regexps = findElementRegexpStrings(text, { fuzziness })

        const notFoundErr = (underlying?: Error): Error => {
            const debuggingExpressions = regexps.map(r => getDebugExpressionFromRegexp(tag, r))
            return new Error(
                `Could not find element with text ${JSON.stringify(text)}, options: ${JSON.stringify(options)}` +
                    (underlying ? `. Underlying error was: ${JSON.stringify(underlying.message)}.` : '') +
                    ` Debug expressions: ${debuggingExpressions.join('\n')}`
            )
        }

        const handlePromise = wait
            ? this.page
                  .waitForFunction(findElementMatchingRegexps, typeof wait === 'object' ? wait : {}, tag, regexps)
                  .catch(err => {
                      throw notFoundErr(err)
                  })
            : this.page.evaluateHandle(findElementMatchingRegexps, tag, regexps)

        const el = (await handlePromise).asElement()
        if (!el) {
            throw notFoundErr()
        }
        return el
    }

    public async waitUntilURL(url: string, options: PageFnOptions = {}): Promise<void> {
        await this.page.waitForFunction(url => document.location.href === url, options, url)
    }

    public async goToURLWithInvalidTLS(url: string): Promise<void> {
        try {
            await this.page.goto(url)
        } catch (err) {
            if (!err.message.includes('net::ERR_CERT_AUTHORITY_INVALID')) {
                throw err
            }
            await this.page.waitForSelector('#details-button')
            await this.page.click('#details-button')
            await (
                await this.findElementWithText('Proceed to', {
                    selector: 'a',
                    wait: { timeout: 2000 },
                })
            ).click()
        }
        await this.page.waitForSelector('.monaco-editor', { timeout: 2000 })
    }
}

export function modifyJSONC(text: string, path: jsonc.JSONPath, f: (oldValue: jsonc.Node | undefined) => any): any {
    const old = jsonc.findNodeAtLocation(jsonc.parseTree(text), path)
    return jsonc.applyEdits(
        text,
        jsoncEdit.setProperty(text, path, f(old), {
            eol: '\n',
            insertSpaces: true,
            tabSize: 2,
        })
    )
}

interface DriverOptions extends LaunchOptions {
    /** If true, load the Sourcegraph browser extension. */
    loadExtension?: boolean

    sourcegraphBaseUrl: string

    /** If true, print browser console messages to stdout. */
    logBrowserConsole?: boolean

    /** If true, keep browser open when driver is closed */
    keepBrowser?: boolean
}

export async function createDriverForTest(options: DriverOptions): Promise<Driver> {
    const { loadExtension, sourcegraphBaseUrl, logBrowserConsole, keepBrowser } = options
    const args = ['--window-size=1280,1024']
    if (process.getuid() === 0) {
        // TODO don't run as root in CI
        console.warn('Running as root, disabling sandbox')
        args.push('--no-sandbox', '--disable-setuid-sandbox')
    }
    if (loadExtension) {
        const chromeExtensionPath = path.resolve(__dirname, '..', '..', '..', 'browser', 'build', 'chrome')
        const manifest = JSON.parse(await readFile(path.resolve(chromeExtensionPath, 'manifest.json'), 'utf-8'))
        if (!manifest.permissions.includes('<all_urls>')) {
            throw new Error(
                'Browser extension was not built with permissions for all URLs.\nThis is necessary because permissions cannot be granted by e2e tests.\nTo fix, run `EXTENSION_PERMISSIONS_ALL_URLS=true yarn run dev` inside the browser/ directory.'
            )
        }
        args.push(`--disable-extensions-except=${chromeExtensionPath}`, `--load-extension=${chromeExtensionPath}`)
    }

    const browser = await puppeteer.launch({
        ...options,
        args,
        headless: readEnvBoolean({ variable: 'HEADLESS', defaultValue: false }),
        defaultViewport: null,
    })
    const page = await browser.newPage()
    if (logBrowserConsole) {
        page.on('console', message => {
            if (message.text().includes('Download the React DevTools')) {
                return
            }
            if (message.text().includes('[HMR]') || message.text().includes('[WDS]')) {
                return
            }
            console.log('Browser console:', util.inspect(message, { colors: true, depth: 2, breakLength: Infinity }))
        })
    }
    return new Driver(browser, page, sourcegraphBaseUrl, keepBrowser)
}
