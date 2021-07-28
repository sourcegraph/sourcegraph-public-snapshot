import assert from 'assert'

import delay from 'delay'
import expect from 'expect'
import type * as sourcegraph from 'sourcegraph'

import { ExtensionManifest } from '@sourcegraph/shared/src/extensions/extensionManifest'
import { Settings } from '@sourcegraph/shared/src/settings/settings'
import { createDriverForTest, Driver, percySnapshot } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'
import { retry } from '@sourcegraph/shared/src/testing/utils'

import { BrowserIntegrationTestContext, createBrowserIntegrationTestContext } from './context'
import { closeInstallPageTab } from './shared'

describe('GitHub', () => {
    let driver: Driver
    before(async () => {
        driver = await createDriverForTest({ loadExtension: true })
        // TODO(tj): Add retries in case this delay isn't large enough in CI
        // If RECORD=false, we may be able to skip this!
        await delay(1000)
        await closeInstallPageTab(driver.browser)
        if (driver.sourcegraphBaseUrl !== 'https://sourcegraph.com') {
            await driver.setExtensionSourcegraphUrl()
        }
    })
    after(() => driver?.close())

    let testContext: BrowserIntegrationTestContext
    beforeEach(async function () {
        testContext = await createBrowserIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })

        testContext.server.get('https://collector.githubapp.com/*').intercept((request, response) => {
            response.sendStatus(200)
        })
        testContext.server.any('https://api.github.com/_private/browser/*').intercept((request, response) => {
            response.sendStatus(200)
        })
        testContext.server.any('http://localhost:8890/*').intercept((request, response) => {
            response.sendStatus(200)
        })

        testContext.overrideGraphQL({
            ViewerConfiguration: () => ({
                viewerConfiguration: {
                    subjects: [],
                    merged: { contents: '', messages: [] },
                },
            }),
            ResolveRev: () => ({
                repository: {
                    mirrorInfo: {
                        cloned: true,
                    },
                    commit: {
                        oid: '1'.repeat(40),
                    },
                },
            }),
            ResolveRepo: ({ rawRepoName }) => ({
                repository: {
                    name: rawRepoName,
                },
            }),
            BlobContent: () => ({
                repository: {
                    commit: {
                        file: {
                            content:
                                'package jsonrpc2\n\n// CallOption is an option that can be provided to (*Conn).Call to\n// configure custom behavior. See Meta.\ntype CallOption interface {\n\tapply(r *Request) error\n}\n\ntype callOptionFunc func(r *Request) error\n\nfunc (c callOptionFunc) apply(r *Request) error { return c(r) }\n\n// Meta returns a call option which attaches the given meta object to\n// the JSON-RPC 2.0 request (this is a Sourcegraph extension to JSON\n// RPC 2.0 for carrying metadata).\nfunc Meta(meta interface{}) CallOption {\n\treturn callOptionFunc(func(r *Request) error {\n\t\treturn r.SetMeta(meta)\n\t})\n}\n\n// PickID returns a call option which sets the ID on a request. Care must be\n// taken to ensure there are no conflicts with any previously picked ID, nor\n// with the default sequence ID.\nfunc PickID(id ID) CallOption {\n\treturn callOptionFunc(func(r *Request) error {\n\t\tr.ID = id\n\t\treturn nil\n\t})\n}\n',
                        },
                    },
                },
            }),
        })
    })
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    it('adds "View on Sourcegraph" buttons to files', async () => {
        const repoName = 'github.com/sourcegraph/jsonrpc2'

        await driver.page.goto(
            'https://github.com/sourcegraph/jsonrpc2/blob/4fb7cd90793ee6ab445f466b900e6bffb9b63d78/call_opt.go'
        )

        await driver.page.waitForSelector('.code-view-toolbar .open-on-sourcegraph', { timeout: 10000 })
        expect(await driver.page.$$('.code-view-toolbar .open-on-sourcegraph')).toHaveLength(1)

        await retry(async () => {
            assert.strictEqual(
                await driver.page.evaluate(
                    () => document.querySelector<HTMLAnchorElement>('.code-view-toolbar .open-on-sourcegraph')?.href
                ),
                `${driver.sourcegraphBaseUrl}/${repoName}@4fb7cd90793ee6ab445f466b900e6bffb9b63d78/-/blob/call_opt.go?utm_source=${driver.browserType}-extension`
            )
        })
    })

    it('shows hover tooltips when hovering a token', async () => {
        const userSettings: Settings = {
            extensions: {
                'test/test': true,
            },
        }
        const extensionManifest: ExtensionManifest = {
            url: new URL('/-/static/extension/0001-test-test.js?hash--test-test', driver.sourcegraphBaseUrl).href,
            activationEvents: ['*'],
        }

        testContext.overrideGraphQL({
            ViewerConfiguration: () => ({
                viewerConfiguration: {
                    subjects: [
                        {
                            __typename: 'User',
                            displayName: 'Test User',
                            id: 'TestUserSettingsID',
                            latestSettings: {
                                id: 123,
                                contents: JSON.stringify(userSettings),
                            },
                            username: 'test',
                            viewerCanAdminister: true,
                            settingsURL: '/users/test/settings',
                        },
                    ],
                    merged: { contents: JSON.stringify(userSettings), messages: [] },
                },
            }),
            Extensions: () => ({
                extensionRegistry: {
                    extensions: {
                        nodes: [
                            {
                                id: 'TestExtensionID',
                                extensionID: 'test/test',
                                manifest: {
                                    raw: JSON.stringify(extensionManifest),
                                },
                                url: '/extensions/test/test',
                                viewerCanAdminister: false,
                            },
                        ],
                    },
                },
            }),
        })

        // Serve a mock extension bundle with a simple hover provider
        testContext.server
            .get(new URL(extensionManifest.url, driver.sourcegraphBaseUrl).href)
            .intercept((request, response) => {
                function extensionBundle(): void {
                    // eslint-disable-next-line @typescript-eslint/no-require-imports, @typescript-eslint/no-var-requires
                    const sourcegraph = require('sourcegraph') as typeof import('sourcegraph')

                    function activate(context: sourcegraph.ExtensionContext): void {
                        context.subscriptions.add(
                            sourcegraph.languages.registerHoverProvider(['*'], {
                                // Ensure that the correct token info reaches the extension
                                provideHover: (document, position) => {
                                    const range = document.getWordRangeAtPosition(position)
                                    const token = document.getText(range)
                                    if (!token) {
                                        return null
                                    }
                                    return {
                                        contents: {
                                            value: `User is hovering over ${token}`,
                                            kind: sourcegraph.MarkupKind.Markdown,
                                        },
                                        range,
                                    }
                                },
                            })
                        )
                    }

                    exports.activate = activate
                }
                // Create an immediately-invoked function expression for the extensionBundle function
                const extensionBundleString = `(${extensionBundle.toString()})()`
                response.type('application/javascript; charset=utf-8').send(extensionBundleString)
            })

        await driver.page.setBypassCSP(true)

        await driver.page.goto(
            'https://github.com/sourcegraph/jsonrpc2/blob/4fb7cd90793ee6ab445f466b900e6bffb9b63d78/call_opt.go'
        )
        await driver.page.waitForSelector('.code-view-toolbar .open-on-sourcegraph')

        // Pause to give codeintellify time to register listeners for
        // tokenization (only necessary in CI, not sure why).
        await driver.page.waitFor(1000)

        const lineSelector = '.js-file-line-container tr'

        // Trigger tokenization of the line.
        const lineNumber = 16
        const line = await driver.page.waitForSelector(`${lineSelector}:nth-child(${lineNumber})`, {
            timeout: 10000,
        })
        const [token] = await line.$x('//span[text()="CallOption"]')
        await token.hover()
        await driver.findElementWithText('User is hovering over CallOption', {
            selector: '.hover-overlay__content > p',
            fuzziness: 'contains',
            wait: {
                timeout: 6000,
            },
        })

        await percySnapshot(driver.page, 'Hover tooltip on GitHub')
    })
})
