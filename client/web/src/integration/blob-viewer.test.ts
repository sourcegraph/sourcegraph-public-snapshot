import assert from 'assert'

import type * as sourcegraph from 'sourcegraph'

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
import { commonWebGraphQlResults } from './graphQlResults'
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
                            file: {
                                content: '// Log to console\nconsole.log("Hello world")\n// Third line',
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
                            file: {
                                content: '// Log to console\nconsole.log("Hello world")',
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

        interface MockExtension {
            id: string
            extensionID: string
            extensionManifest: ExtensionManifest
            /**
             * A function whose body is a Sourcegraph extension.
             *
             * Bundle must import 'sourcegraph' (e.g. `const sourcegraph = require('sourcegraph')`)
             * */
            bundle: () => void
        }

        /**
         * This test is meant to prevent regression: https://github.com/sourcegraph/sourcegraph/pull/15099
         *
         * TODO(philipp-spiess): This test no longer works after enabling the migrated git blame
         * extension. We can remove it once we remove the extension support completely.
         */
        it.skip('adds and clears line decoration attachments properly', async () => {
            testContext.overrideJsContext({ enableLegacyExtensions: true })
            const mockExtensions: MockExtension[] = [
                {
                    id: 'test',
                    extensionID: 'test/fixed-line',
                    extensionManifest: {
                        url: new URL(
                            '/-/static/extension/0001-test-fixed-line.js?hash--test-fixed-line',
                            driver.sourcegraphBaseUrl
                        ).href,
                        activationEvents: ['*'],
                    },
                    bundle: function extensionBundle(): void {
                        // eslint-disable-next-line @typescript-eslint/no-require-imports, @typescript-eslint/no-var-requires
                        const sourcegraph = require('sourcegraph') as typeof import('sourcegraph')

                        const fixedLineDecorationType = sourcegraph.app.createDecorationType()

                        function decorateLineTwo(editor: sourcegraph.CodeEditor) {
                            // Always decorate line 2
                            editor.setDecorations(fixedLineDecorationType, [
                                {
                                    range: new sourcegraph.Range(
                                        new sourcegraph.Position(1, 0),
                                        new sourcegraph.Position(1, 0)
                                    ),
                                    after: {
                                        contentText: 'fixed line content',
                                        backgroundColor: 'red',
                                    },
                                },
                            ])
                        }

                        function activate(context: sourcegraph.ExtensionContext): void {
                            // check initial viewer in case it was already emitted before extension was activated
                            const initialViewer = sourcegraph.app.activeWindow?.activeViewComponent
                            if (initialViewer?.type === 'CodeEditor') {
                                decorateLineTwo(initialViewer)
                            }

                            // subscribe to viewer changes
                            let previousViewerSubscription: sourcegraph.Unsubscribable | undefined
                            context.subscriptions.add(
                                sourcegraph.app.activeWindowChanges.subscribe(activeWindow => {
                                    const viewerSubscription = activeWindow?.activeViewComponentChanges.subscribe(
                                        viewer => {
                                            if (viewer?.type === 'CodeEditor') {
                                                // Always decorate line 2
                                                decorateLineTwo(viewer)
                                            }
                                        }
                                    )
                                    if (previousViewerSubscription) {
                                        previousViewerSubscription.unsubscribe()
                                    }
                                    previousViewerSubscription = viewerSubscription
                                })
                            )
                        }

                        exports.activate = activate
                    },
                },
                {
                    id: 'selected-line',
                    extensionID: 'test/selected-line',
                    extensionManifest: {
                        url: new URL(
                            '/-/static/extension/0001-test-selected-line.js?hash--test-selected-line',
                            driver.sourcegraphBaseUrl
                        ).href,
                        activationEvents: ['*'],
                    },
                    bundle: function extensionBundle(): void {
                        // eslint-disable-next-line @typescript-eslint/no-require-imports, @typescript-eslint/no-var-requires
                        const sourcegraph = require('sourcegraph') as typeof import('sourcegraph')

                        const selectedLineDecorationType = sourcegraph.app.createDecorationType()

                        function activate(context: sourcegraph.ExtensionContext): void {
                            let previousViewerSubscription: sourcegraph.Unsubscribable | undefined
                            let previousSelectionsSubscription: sourcegraph.Unsubscribable | undefined

                            context.subscriptions.add(
                                sourcegraph.app.activeWindowChanges.subscribe(activeWindow => {
                                    const viewerSubscription = activeWindow?.activeViewComponentChanges.subscribe(
                                        viewer => {
                                            let selectionsSubscription: sourcegraph.Unsubscribable | undefined
                                            if (viewer?.type === 'CodeEditor') {
                                                selectionsSubscription = viewer.selectionsChanges.subscribe(
                                                    selections => {
                                                        viewer.setDecorations(
                                                            selectedLineDecorationType,
                                                            selections.map(selection => ({
                                                                range: new sourcegraph.Range(
                                                                    selection.start,
                                                                    selection.end
                                                                ),
                                                                after: {
                                                                    contentText: `selected line content for line ${selection.start.line}`,
                                                                    backgroundColor: 'green',
                                                                },
                                                            }))
                                                        )
                                                    }
                                                )
                                            }

                                            if (previousSelectionsSubscription) {
                                                previousSelectionsSubscription.unsubscribe()
                                            }
                                            previousSelectionsSubscription = selectionsSubscription
                                        }
                                    )

                                    if (previousViewerSubscription) {
                                        previousViewerSubscription.unsubscribe()
                                    }
                                    previousViewerSubscription = viewerSubscription
                                })
                            )
                        }

                        exports.activate = activate
                    },
                },
            ]

            const userSettings: Settings = {
                extensions: mockExtensions.reduce((extensionsSettings: Record<string, boolean>, mockExtension) => {
                    extensionsSettings[mockExtension.extensionID] = true
                    return extensionsSettings
                }, {}),
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
                            file: {
                                content: '// Log to console\nconsole.log("Hello world")\n// Third line',
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
                Extensions: () => ({
                    extensionRegistry: {
                        __typename: 'ExtensionRegistry',
                        extensions: {
                            nodes: mockExtensions.map(mockExtension => ({
                                ...mockExtension,
                                manifest: { jsonFields: mockExtension.extensionManifest },
                            })),
                        },
                    },
                }),
            })

            // Serve mock extension bundles
            for (const mockExtension of mockExtensions) {
                testContext.server
                    .get(new URL(mockExtension.extensionManifest.url, driver.sourcegraphBaseUrl).href)
                    .intercept((request, response) => {
                        // Create an immediately-invoked function expression for the extensionBundle function
                        const extensionBundleString = `(${mockExtension.bundle.toString()})()`
                        response.type('application/javascript; charset=utf-8').send(extensionBundleString)
                    })
            }
            const timeout = 10000
            await driver.page.goto(`${driver.sourcegraphBaseUrl}/github.com/sourcegraph/test/-/blob/test.ts`)

            // Wait for some line decoration attachment portal
            await driver.page.waitForSelector('[data-line-decoration-attachment-portal]', { timeout })
            assert(
                !(await driver.page.$('#line-decoration-attachment-1')),
                'Expected line 1 to not have a decoration attachment portal'
            )
            assert(
                await driver.page.$('#line-decoration-attachment-2'),
                'Expected line 2 to have a decoration attachment portal before selecting a line'
            )
            // Count child nodes of existing portals
            assert.strictEqual(
                await driver.page.evaluate(
                    () => document.querySelector('#line-decoration-attachment-2')?.childElementCount
                ),
                1,
                'Expected line 2 to have 1 decoration'
            )

            // Select line 1. Line 1
            await driver.page.click('[data-line="1"]')
            await driver.page.waitForSelector('#line-decoration-attachment-1', { timeout })
            assert(
                await driver.page.$('#line-decoration-attachment-1'),
                'Expected line 1 to have a decoration attachment portal'
            )
            assert(
                await driver.page.$('#line-decoration-attachment-2'),
                'Expected line 2 to have a decoration attachment portal'
            )
            assert.strictEqual(
                await driver.page.evaluate(
                    () => document.querySelector('#line-decoration-attachment-2')?.childElementCount
                ),
                1,
                'Expected line 2 to have 1 decoration'
            )

            // Select line 2. Assert that everything is normal
            await driver.page.click('[data-line="2"]')
            await driver.page.waitFor(() => !document.querySelector('#line-decoration-attachment-1'), { timeout })
            assert(
                !(await driver.page.$('#line-decoration-attachment-1')),
                'Expected line 1 to not have a decoration attachment portal after selecting line 2'
            )
            assert(
                await driver.page.$('#line-decoration-attachment-2'),
                'Expected line 2 to have a decoration attachment portal'
            )

            // Count child nodes of existing portals
            assert.strictEqual(
                await driver.page.evaluate(
                    () => document.querySelector('#line-decoration-attachment-2')?.childElementCount
                ),
                2,
                'Expected line 2 to have 2 decorations'
            )

            // Select line 1 again. before fix, line 2 will still have 2 decorations
            await driver.page.click('[data-line="1"]')
            await driver.page.waitForSelector('#line-decoration-attachment-1', { timeout })
            assert(
                await driver.page.$('#line-decoration-attachment-1'),
                'Expected line 1 to have a decoration attachment portal after it is reselected'
            )
            assert(
                await driver.page.$('#line-decoration-attachment-2'),
                'Expected line 2 to have a decoration attachment portal after selecting line 1 again'
            )

            // Count child nodes of existing portals
            assert.strictEqual(
                await driver.page.evaluate(
                    () => document.querySelector('#line-decoration-attachment-2')?.childElementCount
                ),
                1,
                'Expected line 2 to have 1 decoration'
            )
        })

        it('sends the latest document to extensions', async () => {
            testContext.overrideJsContext({ enableLegacyExtensions: true })
            // This test is meant to prevent regression of
            // "extensions receive wrong text documents": https://github.com/sourcegraph/sourcegraph/issues/14965

            /**
             * How can we verify that extensions receive the latest document?
             * The test extension has to cause some change detectable to the web application, and
             * this change must be dependent on the text document. This extension should be simple to
             * avoid bugs in the extension itself.
             *
             * Simplest possible extension that satisfies these requirements:
             * add attachment to lines that contain a certain word.
             */

            const wordFinder: MockExtension = {
                id: 'word-finder',
                extensionID: 'test/word-finder',
                extensionManifest: {
                    url: new URL(
                        '/-/static/extension/0001-test-word-finder.js?hash--test-word-finder',
                        driver.sourcegraphBaseUrl
                    ).href,
                    activationEvents: ['*'],
                },
                bundle: function extensionBundle(): void {
                    // eslint-disable-next-line @typescript-eslint/no-require-imports, @typescript-eslint/no-var-requires
                    const sourcegraph = require('sourcegraph') as typeof import('sourcegraph')

                    const fixedLineDecorationType = sourcegraph.app.createDecorationType()

                    // Match all occurrences of 'word', decorate lines
                    function decorateWordLines(editor: sourcegraph.CodeEditor) {
                        const text = editor.document.text ?? ''

                        const wordPattern = /word/g
                        const matchIndices: number[] = []

                        for (const match of text.matchAll(wordPattern)) {
                            if (match.index) {
                                matchIndices.push(match.index)
                            }
                        }

                        editor.setDecorations(
                            fixedLineDecorationType,
                            matchIndices.map(index => {
                                const line = editor.document.positionAt(index).line

                                return {
                                    range: new sourcegraph.Range(
                                        new sourcegraph.Position(line, 0),
                                        new sourcegraph.Position(line, 0)
                                    ),
                                    after: {
                                        contentText: 'found word',
                                        backgroundColor: 'red',
                                    },
                                }
                            })
                        )
                    }

                    function activate(context: sourcegraph.ExtensionContext): void {
                        // check initial viewer in case it was already emitted before extension was activated
                        const initialViewer = sourcegraph.app.activeWindow?.activeViewComponent
                        if (initialViewer?.type === 'CodeEditor') {
                            decorateWordLines(initialViewer)
                        }

                        // subscribe to viewer changes
                        let previousViewerSubscription: sourcegraph.Unsubscribable | undefined
                        context.subscriptions.add(
                            sourcegraph.app.activeWindowChanges.subscribe(activeWindow => {
                                const viewerSubscription = activeWindow?.activeViewComponentChanges.subscribe(
                                    viewer => {
                                        if (viewer?.type === 'CodeEditor') {
                                            decorateWordLines(viewer)
                                        }
                                    }
                                )
                                if (previousViewerSubscription) {
                                    previousViewerSubscription.unsubscribe()
                                }
                                previousViewerSubscription = viewerSubscription
                            })
                        )
                    }

                    exports.activate = activate
                },
            }

            const userSettings: Settings = {
                extensions: {
                    'test/word-finder': true,
                },
            }

            testContext.overrideGraphQL({
                ...commonBlobGraphQlResults,
                TreeEntries: () =>
                    createTreeEntriesResult(repositorySourcegraphUrl, ['README.md', 'test.ts', 'fake.ts']),
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
                Blob: ({ filePath }) => {
                    const file =
                        filePath === 'fake.ts'
                            ? {
                                  content: '// First word line\n// Second line\n// Third word line',
                                  richHTML: '',
                                  highlight: {
                                      aborted: false,
                                      html:
                                          // Note: whitespace in this string is significant.
                                          '<table><tbody><tr><td class="line" data-line="1"></td><td class="code"><div><span style="color: gray">&sol;&sol; First word line\n' +
                                          '</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div><span style="color: gray">&sol;&sol; Second line</span></td></tr>\n' +
                                          '<tr><td class="line" data-line="3"></td><td class="code"><div><span style="color: gray">&sol;&sol; Third word line</span></td></tr></tbody></table>',
                                      lsif: '',
                                  },
                              }
                            : {
                                  content: '// First line\n// Second word line\n// Third line',
                                  richHTML: '',
                                  highlight: {
                                      aborted: false,
                                      html:
                                          // Note: whitespace in this string is significant.
                                          '<table><tbody><tr><td class="line" data-line="1"></td><td class="code"><div><span style="color: gray">&sol;&sol; First line\n' +
                                          '</span></div></td></tr><tr><td class="line" data-line="2"></td><td class="code"><div><span style="color: gray">&sol;&sol; Second word line</span></td></tr>\n' +
                                          '<tr><td class="line" data-line="3"></td><td class="code"><div><span style="color: gray">&sol;&sol; Third line</span></td></tr></tbody></table>',
                                      lsif: '',
                                  },
                              }

                    return {
                        repository: {
                            commit: {
                                file,
                            },
                        },
                    }
                },
                Extensions: () => ({
                    extensionRegistry: {
                        __typename: 'ExtensionRegistry',
                        extensions: {
                            nodes: [
                                {
                                    ...wordFinder,
                                    manifest: {
                                        jsonFields: wordFinder.extensionManifest,
                                    },
                                },
                            ],
                        },
                    },
                }),
            })

            // Serve the word-finder extension bundle
            testContext.server
                .get(new URL(wordFinder.extensionManifest.url, driver.sourcegraphBaseUrl).href)
                .intercept((request, response) => {
                    // Create an immediately-invoked function expression for the extensionBundle function
                    const extensionBundleString = `(${wordFinder.bundle.toString()})()`
                    response.type('application/javascript; charset=utf-8').send(extensionBundleString)
                })

            const timeout = 5000
            await driver.page.goto(`${driver.sourcegraphBaseUrl}/github.com/sourcegraph/test/-/blob/test.ts`)
            await driver.page.waitForSelector('[data-testid="repo-blob"]')

            // File 1 (test.ts). Only line two contains 'word'
            try {
                await driver.page.waitForSelector('#line-decoration-attachment-2', { timeout })
            } catch {
                // Rethrow with contextual error message
                throw new Error('Timeout waiting for #line-decoration-attachment-2 (test.ts, first time)')
            }
            // await driver.page.waitFor(1000)
            assert(
                !(await driver.page.$('#line-decoration-attachment-1')),
                'Expected line 1 to not have a decoration attachment on test.ts'
            )
            assert(
                await driver.page.$('#line-decoration-attachment-2'),
                'Expected line 2 to have a decoration attachment on test.ts'
            )
            assert(
                !(await driver.page.$('#line-decoration-attachment-3')),
                'Expected line 3 to not have a decoration attachment on test.ts'
            )

            // Go to File 2 (fake.ts). Lines one and three contain 'word'
            await driver.findElementWithText('fake.ts', {
                selector: '.test-tree-file-link',
                action: 'click',
            })
            try {
                await driver.page.waitForSelector('#line-decoration-attachment-1', { timeout })
            } catch {
                throw new Error('Timeout waiting for #line-decoration-attachment-1 (fake.ts)')
            }
            assert(
                await driver.page.$('#line-decoration-attachment-1'),
                'Expected line 1 to have a decoration attachment on fake.ts'
            )
            assert(
                !(await driver.page.$('#line-decoration-attachment-2')),
                'Expected line 2 to not have a decoration attachment on fake.ts'
            )
            assert(
                await driver.page.$('#line-decoration-attachment-3'),
                'Expected line 3 to have a decoration attachment on fake.ts'
            )

            // Come back to File 1 (test.ts)
            await driver.findElementWithText('test.ts', {
                selector: '.test-tree-file-link',
                action: 'click',
            })
            try {
                await driver.page.waitForSelector('#line-decoration-attachment-2', { timeout })
            } catch {
                throw new Error('Timeout waiting for #line-decoration-attachment-2 (test.ts, second time)')
            }
            assert(
                !(await driver.page.$('#line-decoration-attachment-1')),
                'Expected line 1 to not have a decoration attachment on test.ts (second visit)'
            )
            assert(
                await driver.page.$('#line-decoration-attachment-2'),
                'Expected line 2 to have a decoration attachment on test.ts (second visit)'
            )
            assert(
                !(await driver.page.$('#line-decoration-attachment-3')),
                'Expected line 3 to not have a decoration attachment on test.ts (second visit)'
            )
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
                            file: {
                                content: `// file path: ${filePath}\nconsole.log("Hello world")`,
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

                        function activate(context: sourcegraph.ExtensionContext): void {
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
