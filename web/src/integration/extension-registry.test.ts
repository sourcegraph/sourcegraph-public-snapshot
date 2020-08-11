import assert from 'assert'
import { createDriverForTest, Driver } from '../../../shared/src/testing/driver'
import { createWebIntegrationTestContext, WebIntegrationTestContext } from './context'
import { saveScreenshotsUponFailures } from '../../../shared/src/testing/screenshotReporter'
import { commonWebGraphQlResults } from './graphQlResults'
import { RegistryExtensionFieldsForList } from '../graphql-operations'
import { siteGQLID, siteID } from './jscontext'
import { ExtensionsResult } from '../../../shared/src/graphql-operations'

const typescriptRawManifest = JSON.stringify({
    activationEvents: ['*'],
    browserslist: [],
    categories: ['Programming languages'],
    contributes: {},
    description: 'TypeScript/JavaScript code intelligence',
    devDependencies: {},
    extensionID: 'sourcegraph/typescript',
    main: 'dist/extension.js',
    name: 'typescript',
    publisher: 'sourcegraph',
    readme: '# Code intelligence for TypeScript/JavaScript',
    scripts: {},
    tags: [],
    url:
        'https://sourcegraph.com/-/static/extension/7895-sourcegraph-typescript.js?c4ri18f15tv4--sourcegraph-typescript',
    version: '0.0.0-DEVELOPMENT',
})

const wordCountRawManifest = JSON.stringify({
    activationEvents: ['*'],
    browserslist: [],
    contributes: {},
    description: 'Counts the number of words in a file',
    devDependencies: {},
    extensionID: 'sqs/word-count',
    license: 'MIT',
    main: 'dist/word-count.js',
    name: 'word-count',
    publisher: 'sqs',
    readme: '# Word count (Sourcegraph extension))\n',
    scripts: {},
    title: 'Word count',
    url: 'https://sourcegraph.com/-/static/extension/593-sqs-word-count.js?bpf75c0smaeg--sqs-word-count',
    version: '0.0.0-DEVELOPMENT',
})

const registryExtensionNodes: RegistryExtensionFieldsForList[] = [
    {
        id: 'test-extension-1',
        publisher: null,
        extensionID: 'sourcegraph/typescript',
        extensionIDWithoutRegistry: 'sourcegraph/typescript',
        name: 'typescript',
        manifest: {
            raw: typescriptRawManifest,
            description: 'TypeScript/JavaScript code intelligence',
        },
        createdAt: '2019-01-26T03:39:17Z',
        updatedAt: '2019-01-26T03:39:17Z',
        url: '/extensions/sourcegraph/typescript',
        remoteURL: 'https://sourcegraph.com/extensions/sourcegraph/typescript',
        registryName: 'sourcegraph.com',
        isLocal: false,
        isWorkInProgress: false,
        viewerCanAdminister: false,
    },
    {
        id: 'test-extension-2',
        publisher: null,
        extensionID: 'sqs/word-count',
        extensionIDWithoutRegistry: 'sqs/word-count',
        name: 'word-count',
        manifest: {
            raw: wordCountRawManifest,
            description: 'Counts the number of words in a file',
        },
        createdAt: '2018-10-28T22:33:08Z',
        updatedAt: '2018-10-28T22:33:08Z',
        url: '/extensions/sqs/word-count',
        remoteURL: 'https://sourcegraph.com/extensions/sqs/word-count',
        registryName: 'sourcegraph.com',
        isLocal: false,
        isWorkInProgress: false,
        viewerCanAdminister: false,
    },
]

const extensionNodes: ExtensionsResult['extensionRegistry']['extensions']['nodes'] = [
    {
        extensionID: 'sourcegraph/typescript',
        id: 'test-extension-1',
        manifest: {
            raw: typescriptRawManifest,
        },
        url: '/extensions/sourcegraph/typescript',
        viewerCanAdminister: false,
    },
    {
        extensionID: 'sqs/word-count',
        id: 'test-extension-2',
        manifest: {
            raw: wordCountRawManifest,
        },
        url: '/extensions/sqs/word-count',
        viewerCanAdminister: false,
    },
]

describe('Extension Registry', () => {
    let driver: Driver
    before(async () => {
        driver = await createDriverForTest()
    })
    after(() => driver?.close())
    let testContext: WebIntegrationTestContext
    beforeEach(async function () {
        testContext = await createWebIntegrationTestContext({
            driver,
            currentTest: this.currentTest!,
            directory: __dirname,
        })
    })
    saveScreenshotsUponFailures(() => driver.page)
    afterEach(() => testContext?.dispose())

    function overrideGraphQLExtensionRegistry({ enabled }: { enabled: boolean }): void {
        testContext.overrideGraphQL({
            ...commonWebGraphQlResults,
            ViewerSettings: () => ({
                viewerSettings: {
                    subjects: [
                        {
                            __typename: 'DefaultSettings',
                            settingsURL: null,
                            viewerCanAdminister: false,
                            latestSettings: {
                                id: 0,
                                contents: JSON.stringify({}),
                            },
                        },
                        {
                            __typename: 'Site',
                            id: siteGQLID,
                            siteID,
                            latestSettings: {
                                id: 470,
                                contents: JSON.stringify({}),
                            },
                            settingsURL: '/site-admin/global-settings',
                            viewerCanAdminister: true,
                        },
                        {
                            __typename: 'User',
                            id: 'TestGQLUserID',
                            username: 'testusername',
                            settingsURL: '/user/testusername/settings',
                            displayName: 'test',
                            viewerCanAdminister: true,
                            latestSettings: {
                                id: 310,
                                contents: JSON.stringify({ extensions: { 'sqs/word-count': enabled } }),
                            },
                        },
                    ],
                    final: JSON.stringify({}),
                },
            }),
            RegistryExtensions: () => ({
                extensionRegistry: {
                    extensions: {
                        error: null,
                        nodes: registryExtensionNodes,
                    },
                },
            }),
            Extensions: () => ({
                extensionRegistry: {
                    extensions: {
                        nodes: extensionNodes,
                    },
                },
            }),
            EditSettings: () => ({
                configurationMutation: {
                    editConfiguration: {
                        empty: null,
                    },
                },
            }),
        })
    }

    function elementExists(selector: string) {
        return driver.page.evaluate(selector => document.querySelector(selector) !== null, selector)
    }

    describe('filtering by category', () => {
        it('does not show language extensions until user clicks show more', async () => {
            overrideGraphQLExtensionRegistry({ enabled: false })
            await driver.page.goto(driver.sourcegraphBaseUrl + '/extensions')

            //  wait for initial set of extensions
            await driver.page.waitForSelector('[data-test="extension-toggle-sqs/word-count"]')
            assert.strictEqual(await elementExists('[data-test="extension-toggle-sourcegraph/typescript"]'), false)
            await driver.findElementWithText('Show more extensions', { action: 'click' })
            assert.strictEqual(await elementExists('[data-test="extension-toggle-sourcegraph/typescript"]'), true)
        })

        it('only shows extensions from the selected categories', async () => {
            overrideGraphQLExtensionRegistry({ enabled: false })
            await driver.page.goto(driver.sourcegraphBaseUrl + '/extensions')

            //  wait for initial set of extensions
            await driver.page.waitForSelector('[data-test="extension-toggle-sqs/word-count"]')
            assert.strictEqual(await elementExists('[data-test="extension-toggle-sqs/word-count"]'), true)
            assert.strictEqual(await elementExists('[data-test="extension-toggle-sourcegraph/typescript"]'), false)
            await driver.page.click('[data-test-extension-category="Programming languages"')
            assert.strictEqual(await elementExists('[data-test="extension-toggle-sqs/word-count"]'), false)
            assert.strictEqual(await elementExists('[data-test="extension-toggle-sourcegraph/typescript"]'), true)
        })
    })

    describe('searching', () => {
        it('input leads to the correct query', async () => {
            // testing that text input makes it through the RxJS pipeline
            overrideGraphQLExtensionRegistry({ enabled: false })
            await driver.page.goto(driver.sourcegraphBaseUrl + '/extensions')

            await driver.page.waitForSelector('.test-extension-registry-input')

            const { query } = await testContext.waitForGraphQLRequest(async () => {
                await driver.replaceText({
                    selector: '.test-extension-registry-input',
                    newText: 'sqs',
                    enterTextMethod: 'paste',
                })
            }, 'RegistryExtensions')

            assert.strictEqual(query, 'sqs')
        })
    })

    describe('toggling', () => {
        it('a disabled extension enables it', async () => {
            const enabled = false
            overrideGraphQLExtensionRegistry({ enabled })

            await driver.page.goto(driver.sourcegraphBaseUrl + '/extensions')

            // toggle typescript extension on
            const { edit } = await testContext.waitForGraphQLRequest(async () => {
                const toggle = await driver.page.waitForSelector("[data-test='extension-toggle-sqs/word-count']")
                await toggle.click()
            }, 'EditSettings')

            assert.deepStrictEqual(edit, {
                keyPath: [{ property: 'extensions' }, { property: 'sqs/word-count' }],
                value: !enabled,
            })
        })

        it('an enabled extension disables it ', async () => {
            const enabled = true
            overrideGraphQLExtensionRegistry({ enabled })

            await driver.page.goto(driver.sourcegraphBaseUrl + '/extensions')

            // toggle typescript extension off
            const { edit } = await testContext.waitForGraphQLRequest(async () => {
                const toggle = await driver.page.waitForSelector("[data-test='extension-toggle-sqs/word-count']")
                await toggle.click()
            }, 'EditSettings')

            assert.deepStrictEqual(edit, {
                keyPath: [{ property: 'extensions' }, { property: 'sqs/word-count' }],
                value: !enabled,
            })
        })
    })
})
