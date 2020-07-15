import assert from 'assert'
import { commonWebGraphQlResults } from './graphQlResults'
import { Driver, createDriverForTest, percySnapshot } from '../../../shared/src/testing/driver'
import { ExtensionManifest } from '../../../shared/src/schema/extensionSchema'
import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import {
    createRepositoryRedirectResult,
    createResolveRevisionResult,
    createFileExternalLinksResult,
    createTreeEntriesResult,
    createBlobContentResult,
} from './graphQlResponseHelpers'
import { WebGraphQlOperations, ViewerSettingsResult } from '../graphql-operations'
import { SharedGraphQlOperations } from '../../../shared/src/graphql-operations'
import { Settings } from '../schema/settings.schema'
import type * as sourcegraph from 'sourcegraph'

describe('Blob viewer', () => {
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
    afterEach(() => testContext?.dispose())

    const repositoryName = 'github.com/sourcegraph/jsonrpc2'
    const repositorySourcegraphUrl = `/${repositoryName}`
    const fileName = 'test.ts'
    const files = ['README.md', fileName]

    const commonBlobGraphQlResults: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
        ...commonWebGraphQlResults,
        RepositoryRedirect: ({ repoName }) => createRepositoryRedirectResult(repoName),
        ResolveRev: () => createResolveRevisionResult(repositorySourcegraphUrl),
        FileExternalLinks: ({ filePath }) =>
            createFileExternalLinksResult(`https://${repositoryName}/blob/master/${filePath}`),
        TreeEntries: () => createTreeEntriesResult(repositorySourcegraphUrl, ['README.md', fileName]),
        Blob: ({ filePath }) => createBlobContentResult(`content for: ${filePath}\nsecond line\nthird line`),
    }
    beforeEach(() => {
        testContext.overrideGraphQL(commonBlobGraphQlResults)
    })

    describe('general layout for viewing a file', () => {
        it('populates editor content and FILES tab', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}/${repositoryName}/-/blob/${fileName}`)
            await driver.page.waitForSelector('.e2e-repo-blob')
            const blobContent = await driver.page.evaluate(
                () => document.querySelector<HTMLElement>('.e2e-repo-blob')?.textContent
            )

            // editor shows the return string content from Blob request
            assert.strictEqual(blobContent, `content for: ${fileName}`)

            // collect all files/links visible the the "Files" tab
            const allFilesInTheTree = await driver.page.evaluate(() => {
                const allFiles = document.querySelectorAll<HTMLAnchorElement>('.e2e-tree-file-link')

                return [...allFiles].map(fileAnchor => ({
                    content: fileAnchor.textContent,
                    href: fileAnchor.href,
                }))
            })

            // files from TreeEntries request
            assert.deepStrictEqual(
                allFilesInTheTree,
                files.map(name => ({
                    content: name,
                    href: `${driver.sourcegraphBaseUrl}/${repositoryName}/-/blob/${name}`,
                }))
            )
        })
    })

    // Describes the ways the blob viewer can be extended through Sourcegraph extensions.
    describe('extensibility', () => {
        const getHoverContents = async (): Promise<string[]> => {
            // Search for any child of e2e-tooltip-content: as e2e-tooltip-content has display: contents,
            // it will never be detected as visible by waitForSelector(), but its children will.
            await driver.page.waitForSelector('.e2e-tooltip-content *', { visible: true })
            return driver.page.evaluate(() =>
                [...document.querySelectorAll('.e2e-tooltip-content')].map(content => content.textContent ?? '')
            )
        }

        beforeEach(() => {
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
                ...commonBlobGraphQlResults,
                ViewerSettings: () =>
                    ({
                        viewerSettings: {
                            final: JSON.stringify(userSettings),
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
                        },
                    } as ViewerSettingsResult),
                Blob: () => ({
                    repository: {
                        commit: {
                            file: {
                                content: '// Log to console\nconsole.log("Hello world")',
                                richHTML: '',
                                highlight: {
                                    aborted: false,
                                    html:
                                        // Note: whitespace in this string is significant.
                                        '<table><tbody><tr><td class="line" data-line="1"></td><td class="code"><div><span style="color: gray">&sol;&sol; Log to console\n' +
                                        '</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div><span style="color: #859900;">console</span><span style="color: #657b83;">.</span><span style="color: #859900;" class="e2e-log-token">log</span><span style="color: #657b83;">(</span><span style="color: #839496;">&quot;</span><span style="color: #2aa198;">Hello world</span><span style="color: #839496;">&quot;</span><span style="color: #657b83;">)\n' +
                                        '</span></div></td></tr></tbody></table>',
                                },
                            },
                        },
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
                    console.log('bundle handler called')
                    function extensionBundle(): void {
                        // eslint-disable-next-line @typescript-eslint/no-require-imports, @typescript-eslint/no-var-requires
                        const sourcegraph = require('sourcegraph') as typeof import('sourcegraph')

                        function activate(context: sourcegraph.ExtensionContext): void {
                            context.subscriptions.add(
                                sourcegraph.languages.registerHoverProvider([{ language: 'typescript' }], {
                                    provideHover: () => ({
                                        contents: {
                                            kind: sourcegraph.MarkupKind.Markdown,
                                            value: 'Test hover content',
                                        },
                                    }),
                                })
                            )
                        }

                        exports.activate = activate
                    }
                    response.type('application/javascript; charset=utf-8').send(extensionBundle.toString())
                })
        })
        it.skip('shows a hover overlay from a hover provider when a token is hovered', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}/${repositoryName}/-/blob/${fileName}`)
            await driver.page.waitForSelector('.e2e-repo-blob')
            // TODO
        })

        it.only('shows a hover overlay from a hover provider and updates the URL when a token is clicked', async function () {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}/github.com/sourcegraph/test/-/blob/test.ts`)

            // Click on "log" in "console.log()" in line 2
            await driver.page.waitForSelector('.e2e-log-token', { visible: true })
            await driver.page.click('.e2e-log-token')

            await driver.assertWindowLocation('/github.com/sourcegraph/test/-/blob/test.ts#L2:9')
            assert.deepStrictEqual(await getHoverContents(), ['Test hover content'])
            await percySnapshot(driver.page, this.test!.fullTitle())
        })

        it.skip('gets displayed when navigating to a URL with a token position', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}/github.com/sourcegraph/test/-/test.ts#2:9`)
            // TODO
        })
    })
})
