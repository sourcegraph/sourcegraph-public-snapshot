import assert from 'assert'

import type { ExtensionContext } from '@sourcegraph/shared/src/codeintel/legacy-extensions/api'
import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { ExtensionManifest } from '@sourcegraph/shared/src/schema/extensionSchema'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { accessibilityAudit } from '@sourcegraph/shared/src/testing/accessibility'
import { Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { WebGraphQlOperations } from '../graphql-operations'

import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import {
    createResolveRepoRevisionResult,
    createFileExternalLinksResult,
    createTreeEntriesResult,
    createBlobContentResult,
} from './graphQlResponseHelpers'
import { commonWebGraphQlResults, createViewerSettingsGraphQLOverride } from './graphQlResults'
import { percySnapshotWithVariants } from './utils'

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
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    const repositoryName = 'github.com/sourcegraph/jsonrpc2'
    const repositorySourcegraphUrl = `/${repositoryName}`
    const fileName = 'test.ts'
    const files = ['README.md', fileName]

    const commonBlobGraphQlResults: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
        ...commonWebGraphQlResults,
        ...createViewerSettingsGraphQLOverride({
            user: {
                experimentalFeatures: {
                    enableCodeMirrorFileView: false,
                },
            },
        }),
        ResolveRepoRev: () => createResolveRepoRevisionResult(repositorySourcegraphUrl),
        FileExternalLinks: ({ filePath }) =>
            createFileExternalLinksResult(`https://${repositoryName}/blob/master/${filePath}`),
        TreeEntries: () => createTreeEntriesResult(repositorySourcegraphUrl, ['README.md', fileName]),
        Blob: ({ filePath }) => createBlobContentResult(`content for: ${filePath}\nsecond line\nthird line`),
        FileNames: () => ({
            repository: {
                id: 'repo-123',
                __typename: 'Repository',
                commit: {
                    id: 'c0ff33',
                    __typename: 'GitCommit',
                    fileNames: ['README.md'],
                },
            },
        }),
    }

    beforeEach(() => {
        testContext.overrideGraphQL(commonBlobGraphQlResults)
    })

    describe('general layout for viewing a file', () => {
        it('populates editor content and FILES tab', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}/${repositoryName}/-/blob/${fileName}`)
            await driver.page.waitForSelector('[data-testid="repo-blob"]')
            const blobContent = await driver.page.evaluate(
                () => document.querySelector<HTMLElement>('[data-testid="repo-blob"]')?.textContent
            )

            // editor shows the return string content from Blob request
            assert.strictEqual(blobContent, `content for: ${fileName}\nsecond line\nthird line`)

            // collect all files/links visible the the "Files" tab
            const allFilesInTheTree = await driver.page.evaluate(() => {
                const allFiles = document.querySelectorAll<HTMLAnchorElement>('.test-tree-file-link')

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

    describe('line number redirects', () => {
        beforeEach(() => {
            testContext.overrideGraphQL({
                ...commonBlobGraphQlResults,
                Blob: () => ({
                    repository: {
                        commit: {
                            blob: null,
                            file: {
                                __typename: 'VirtualFile',
                                content: '// Log to console\nconsole.log("Hello world")\n// Third line',
                                totalLines: 3,
                                richHTML: '',
                                highlight: {
                                    aborted: false,
                                    html:
                                        // Note: whitespace in this string is significant.
                                        '<table class="test-log-token"><tbody><tr><td class="line" data-line="1"/>' +
                                        '<td class="code"><span class="hl-source hl-js hl-react"><span class="hl-comment hl-line hl-double-slash hl-js">' +
                                        '<span class="hl-punctuation hl-definition hl-comment hl-js">//</span> ' +
                                        'Log to console\n</span></span></td></tr>' +
                                        '<tr><td class="line" data-line="2"/><td class="code"><span class="hl-source hl-js hl-react">' +
                                        '<span class="hl-meta hl-function-call hl-method hl-js">' +
                                        '<span class="hl-support hl-type hl-object hl-console hl-js">console</span>' +
                                        '<span class="hl-punctuation hl-accessor hl-js">.</span>' +
                                        '<span class="hl-support hl-function hl-console hl-js test-log-token">log</span>' +
                                        '<span class="hl-meta hl-group hl-js"><span class="hl-punctuation hl-section hl-group hl-begin hl-js">(</span>' +
                                        '<span class="hl-meta hl-string hl-js"><span class="hl-string hl-quoted hl-double hl-js">' +
                                        '<span class="hl-punctuation hl-definition hl-string hl-begin hl-js">&quot;</span>Hello world' +
                                        '<span class="hl-punctuation hl-definition hl-string hl-end hl-js">&quot;</span></span></span>' +
                                        '<span class="hl-punctuation hl-section hl-group hl-end hl-js">)</span></span>\n</span></span></td></tr>' +
                                        '<tr><td class="line" data-line="3"/><td class="code"><span class="hl-source hl-js hl-react">' +
                                        '<span class="hl-meta hl-function-call hl-method hl-js"></span>' +
                                        '<span class="hl-comment hl-line hl-double-slash hl-js"><span class="hl-punctuation hl-definition hl-comment hl-js">//</span> ' +
                                        'Third line\n</span></span></td></tr></tbody></table>',
                                    lsif: '',
                                },
                            },
                        },
                    },
                }),
            })
        })

        it('should redirect from line number hash to query parameter', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}/${repositoryName}/-/blob/${fileName}#2`)
            await driver.page.waitForSelector('[data-testid="repo-blob"]')
            await driver.assertWindowLocation(`/${repositoryName}/-/blob/${fileName}?L2`)
        })

        it('should redirect from line range hash to query parameter', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}/${repositoryName}/-/blob/${fileName}#1-3`)
            await driver.page.waitForSelector('[data-testid="repo-blob"]')
            await driver.assertWindowLocation(`/${repositoryName}/-/blob/${fileName}?L1-3`)
        })
    })

    // Describes the ways the blob viewer can be extended through Sourcegraph extensions.
    describe('extensibility', () => {
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
                ViewerSettings: () => ({
                    viewerSettings: {
                        __typename: 'SettingsCascade',
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
                }),
                Blob: () => ({
                    repository: {
                        commit: {
                            blob: null,
                            file: {
                                __typename: 'VirtualFile',
                                content: '// Log to console\nconsole.log("Hello world")',
                                totalLines: 2,
                                richHTML: '',
                                highlight: {
                                    aborted: false,
                                    html:
                                        // Note: whitespace in this string is significant.
                                        '<table><tbody><tr>' +
                                        '<td class="line" data-line="1"/>' +
                                        '<td class="code"><span class="hl-source hl-js hl-react"><span class="hl-comment hl-line hl-double-slash hl-js">' +
                                        '<span class="hl-punctuation hl-definition hl-comment hl-js">//</span> ' +
                                        'Log to console\n</span></span></td></tr>' +
                                        '<tr><td class="line" data-line="2"/>' +
                                        '<td class="code"><span class="hl-source hl-js hl-react"><span class="hl-meta hl-function-call hl-method hl-js">' +
                                        '<span class="hl-support hl-type hl-object hl-console hl-js test-console-token">console</span>' +
                                        '<span class="hl-punctuation hl-accessor hl-js">.</span>' +
                                        '<span class="hl-support hl-function hl-console hl-js test-log-token">log</span>' +
                                        '<span class="hl-meta hl-group hl-js"><span class="hl-punctuation hl-section hl-group hl-begin hl-js">(</span>' +
                                        '<span class="hl-meta hl-string hl-js"><span class="hl-string hl-quoted hl-double hl-js">' +
                                        '<span class="hl-punctuation hl-definition hl-string hl-begin hl-js">&quot;</span>' +
                                        'Hello world' +
                                        '<span class="hl-punctuation hl-definition hl-string hl-end hl-js">&quot;</span></span></span>' +
                                        '<span class="hl-punctuation hl-section hl-group hl-end hl-js">)</span></span>\n</span></span></td></tr></tbody></table>',
                                    lsif: '',
                                },
                            },
                        },
                    },
                }),
                Extensions: () => ({
                    extensionRegistry: {
                        __typename: 'ExtensionRegistry',
                        extensions: {
                            nodes: [
                                {
                                    id: 'test',
                                    extensionID: 'test/test',
                                    manifest: {
                                        jsonFields: extensionManifest,
                                    },
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

                        function activate(context: ExtensionContext): void {
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
                    // Create an immediately-invoked function expression for the extensionBundle function
                    const extensionBundleString = `(${extensionBundle.toString()})()`
                    response.type('application/javascript; charset=utf-8').send(extensionBundleString)
                })
        })
        it('truncates long file paths properly', async () => {
            await driver.page.goto(
                `${driver.sourcegraphBaseUrl}/${repositoryName}/-/blob/this_is_a_long_file_path/apps/rest-showcase/src/main/java/org/demo/rest/example/OrdersController.java`
            )
            await driver.page.waitForSelector('[data-testid="repo-blob"]')
            await driver.page.waitForSelector('.test-breadcrumb')
            // Uncomment this snapshot once https://github.com/sourcegraph/sourcegraph/issues/15126 is resolved
            // await percySnapshot(driver.page, this.test!.fullTitle())
        })

        it.skip('shows a hover overlay from a hover provider when a token is hovered', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}/${repositoryName}/-/blob/${fileName}`)
            await driver.page.waitForSelector('[data-testid="repo-blob"]')
            // TODO
        })

        it.skip('gets displayed when navigating to a URL with a token position', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}/github.com/sourcegraph/test/-/test.ts#2:9`)
            // TODO
        })

        // Disabled because it's flaky. See: https://github.com/sourcegraph/sourcegraph/issues/31806
        it.skip('properly displays reference panel for URIs with spaces', async () => {
            const repositoryName = 'github.com/sourcegraph/test%20repo'
            const files = ['test.ts', 'test spaces.ts']
            const commitID = '1234'
            const userSettings: Settings = {
                extensions: {
                    'test/references': true,
                },
            }
            const extensionManifest: ExtensionManifest = {
                url: new URL(
                    '/-/static/extension/0001-test-references.js?hash--test-references',
                    driver.sourcegraphBaseUrl
                ).href,
                activationEvents: ['*'],
            }
            testContext.overrideGraphQL({
                ...commonBlobGraphQlResults,
                ViewerSettings: () => ({
                    viewerSettings: {
                        __typename: 'SettingsCascade',
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
                }),
                Extensions: () => ({
                    extensionRegistry: {
                        __typename: 'ExtensionRegistry',
                        extensions: {
                            nodes: [
                                {
                                    id: 'test',
                                    extensionID: 'test/references',
                                    manifest: {
                                        jsonFields: extensionManifest,
                                    },
                                },
                            ],
                        },
                    },
                }),
                Blob: ({ filePath }) => ({
                    repository: {
                        commit: {
                            blob: null,
                            file: {
                                __typename: 'VirtualFile',
                                content: `// file path: ${filePath}\nconsole.log("Hello world")`,
                                totalLines: 2,
                                richHTML: '',
                                highlight: {
                                    aborted: false,
                                    html:
                                        '<table><tbody><tr><td class="line" data-line="1"/><td class="code">' +
                                        '<span class="hl-source hl-js hl-react"><span class="hl-comment hl-line hl-double-slash hl-js">' +
                                        '<span class="hl-punctuation hl-definition hl-comment hl-js">//</span>' +
                                        ` file path: ${filePath}\n` +
                                        '</span></span></td></tr>' +
                                        '<tr><td class="line" data-line="2"/>' +
                                        '<td class="code"><span class="hl-source hl-js hl-react">' +
                                        '<span class="hl-meta hl-function-call hl-method hl-js">' +
                                        '<span class="hl-support hl-type hl-object hl-console hl-js test-console-token">console</span>' +
                                        '<span class="hl-punctuation hl-accessor hl-js">.</span>' +
                                        '<span class="hl-support hl-function hl-console hl-js test-log-token">log</span>' +
                                        '<span class="hl-meta hl-group hl-js"><span class="hl-punctuation hl-section hl-group hl-begin hl-js">(</span>' +
                                        '<span class="hl-meta hl-string hl-js"><span class="hl-string hl-quoted hl-double hl-js">' +
                                        '<span class="hl-punctuation hl-definition hl-string hl-begin hl-js">&quot;</span>' +
                                        'Hello world' +
                                        '<span class="hl-punctuation hl-definition hl-string hl-end hl-js">&quot;</span></span></span>' +
                                        '<span class="hl-punctuation hl-section hl-group hl-end hl-js">)</span></span>' +
                                        '\n</span></span></td></tr></tbody></table>',
                                    lsif: '',
                                    lineRanges: [],
                                },
                            },
                        },
                    },
                }),
                TreeEntries: () => createTreeEntriesResult(repositorySourcegraphUrl, files),
                ResolveRepoRev: () => createResolveRepoRevisionResult(repositorySourcegraphUrl, commitID),
                HighlightedFile: ({ filePath }) => ({
                    repository: {
                        commit: {
                            file: {
                                isDirectory: false,
                                richHTML: '',
                                highlight: {
                                    aborted: false,
                                    html:
                                        '<table><tbody><tr><td class="line" data-line="1"/><td class="code">' +
                                        '<span class="hl-source hl-js hl-react"><span class="hl-comment hl-line hl-double-slash hl-js">' +
                                        '<span class="hl-punctuation hl-definition hl-comment hl-js">//</span>' +
                                        ` file path: ${filePath}\n` +
                                        '</span></span></td></tr>' +
                                        '<tr><td class="line" data-line="2"/>' +
                                        '<td class="code"><span class="hl-source hl-js hl-react">' +
                                        '<span class="hl-meta hl-function-call hl-method hl-js">' +
                                        '<span class="hl-support hl-type hl-object hl-console hl-js test-console-token">console</span>' +
                                        '<span class="hl-punctuation hl-accessor hl-js">.</span>' +
                                        '<span class="hl-support hl-function hl-console hl-js test-log-token">log</span>' +
                                        '<span class="hl-meta hl-group hl-js"><span class="hl-punctuation hl-section hl-group hl-begin hl-js">(</span>' +
                                        '<span class="hl-meta hl-string hl-js"><span class="hl-string hl-quoted hl-double hl-js">' +
                                        '<span class="hl-punctuation hl-definition hl-string hl-begin hl-js">&quot;</span>' +
                                        'Hello world' +
                                        '<span class="hl-punctuation hl-definition hl-string hl-end hl-js">&quot;</span></span></span>' +
                                        '<span class="hl-punctuation hl-section hl-group hl-end hl-js">)</span></span>' +
                                        '\n</span></span></td></tr></tbody></table>',
                                    lineRanges: [],
                                },
                            },
                        },
                    },
                }),
                FetchCommits: () => ({
                    node: { __typename: 'GitCommit' },
                }),
                // Required for definition provider,
                ResolveRawRepoName: () => ({
                    repository: {
                        mirrorInfo: {
                            cloned: true,
                        },
                        uri: repositoryName,
                    },
                }),
            })

            // Serve a mock extension bundle with a simple reference provider
            testContext.server
                .get(new URL(extensionManifest.url, driver.sourcegraphBaseUrl).href)
                .intercept((request, response) => {
                    function extensionBundle(): void {
                        // eslint-disable-next-line @typescript-eslint/no-require-imports, @typescript-eslint/no-var-requires
                        const sourcegraph = require('sourcegraph') as typeof import('sourcegraph')

                        function activate(context: ExtensionContext): void {
                            context.subscriptions.add(
                                sourcegraph.languages.registerReferenceProvider(['*'], {
                                    provideReferences: () => [
                                        new sourcegraph.Location(
                                            new URL('git://github.com/sourcegraph/test%20repo?1234#test%20spaces.ts'),
                                            new sourcegraph.Range(
                                                new sourcegraph.Position(0, 0),
                                                new sourcegraph.Position(1, 0)
                                            )
                                        ),
                                    ],
                                })
                            )

                            // We aren't testing definition providers in this test; we include a definition provider
                            // because the "Find references" action isn't displayed unless a definition is found
                            context.subscriptions.add(
                                sourcegraph.languages.registerDefinitionProvider(['*'], {
                                    provideDefinition: () =>
                                        new sourcegraph.Location(
                                            new URL('git://github.com/sourcegraph/test%20repo?1234#test%20spaces.ts'),
                                            new sourcegraph.Range(
                                                new sourcegraph.Position(0, 0),
                                                new sourcegraph.Position(1, 0)
                                            )
                                        ),
                                })
                            )
                        }

                        exports.activate = activate
                    }
                    // Create an immediately-invoked function expression for the extensionBundle function
                    const extensionBundleString = `(${extensionBundle.toString()})()`
                    response.type('application/javascript; charset=utf-8').send(extensionBundleString)
                })

            // TEMPORARY: Mock `Date.now` to prevent temporary Firefox from rendering.
            await driver.page.evaluateOnNewDocument(() => {
                // Number of ms between Unix epoch and July 1, 2020 (outside of Firefox campaign range)
                const mockMs = new Date('July 1, 2020 00:00:00 UTC').getTime()
                Date.now = () => mockMs
            })

            await driver.page.goto(`${driver.sourcegraphBaseUrl}/${repositoryName}/-/blob/test.ts`)

            // Click on "log" in "console.log()" in line 2
            await driver.page.waitForSelector('.test-log-token', { visible: true })
            await driver.page.click('.test-log-token')

            // Click 'Find references'
            await driver.page.waitForSelector('.test-tooltip-find-references', { visible: true })
            await driver.page.click('.test-tooltip-find-references')

            await driver.page.waitForSelector('.test-file-match-children-item', { visible: true })

            await percySnapshotWithVariants(driver.page, 'Blob Reference Panel', { waitForCodeHighlighting: true })
            await accessibilityAudit(driver.page)
            // Click on the first reference
            await driver.page.click('.test-file-match-children-item')

            // Assert that the first line of code has text content which contains: 'file path: test spaces.ts'
            try {
                await driver.page.waitForFunction(
                    () =>
                        document
                            .querySelector('[data-testid="repo-blob"] [data-line="1"]')
                            ?.nextElementSibling?.textContent?.includes('file path: test spaces.ts'),
                    { timeout: 5000 }
                )
            } catch {
                throw new Error('Expected to navigate to file after clicking on link in references panel')
            }
        })
    })
})
