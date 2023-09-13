import fs from 'fs'
import path from 'path'

import { subDays } from 'date-fns'
import expect from 'expect'
import { afterEach, beforeEach, describe, it } from 'mocha'

import type { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import {
    highlightFileResult,
    mixedSearchStreamEvents,
} from '@sourcegraph/shared/src/search/integration/streaming-search-mocks'
import type { SearchEvent } from '@sourcegraph/shared/src/search/stream'
import { accessibilityAudit } from '@sourcegraph/shared/src/testing/accessibility'
import { type Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import {
    type CreateNotebookBlockInput,
    type NotebookFields,
    type WebGraphQlOperations,
    NotebookBlockType,
    SymbolKind,
} from '../graphql-operations'
import type { BlockType } from '../notebooks'

import { type WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import { createResolveRepoRevisionResult } from './graphQlResponseHelpers'
import { commonWebGraphQlResults } from './graphQlResults'
import { siteGQLID, siteID } from './jscontext'
import { percySnapshotWithVariants } from './utils'

const viewerSettings: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
    ViewerSettings: () => ({
        viewerSettings: {
            __typename: 'SettingsCascade',
            subjects: [
                {
                    __typename: 'DefaultSettings',
                    id: 'TestDefaultSettingsID',
                    settingsURL: null,
                    viewerCanAdminister: false,
                    latestSettings: {
                        id: 0,
                        contents: JSON.stringify({
                            experimentalFeatures: {
                                showSearchContext: true,
                                showSearchNotebook: true,
                            },
                        }),
                    },
                },
                {
                    __typename: 'Site',
                    id: siteGQLID,
                    siteID,
                    latestSettings: {
                        id: 470,
                        contents: JSON.stringify({
                            experimentalFeatures: {
                                showSearchNotebook: true,
                            },
                        }),
                    },
                    settingsURL: '/site-admin/global-settings',
                    viewerCanAdminister: true,
                    allowSiteSettingsEdits: true,
                },
            ],
            final: JSON.stringify({}),
        },
    }),
}

const now = new Date()
const downloadPath = process.env.TEST_TMPDIR || __dirname

const notebookFixture = (id: string, title: string, blocks: NotebookFields['blocks']): NotebookFields => ({
    __typename: 'Notebook',
    id,
    title,
    createdAt: subDays(now, 5).toISOString(),
    updatedAt: subDays(now, 5).toISOString(),
    public: true,
    viewerCanManage: true,
    viewerHasStarred: true,
    namespace: { __typename: 'User', id: '1', namespaceName: 'user1' },
    stars: { totalCount: 123 },
    creator: { __typename: 'User', username: 'user1' },
    updater: { __typename: 'User', username: 'user1' },
    blocks,
})

const GQLBlockInputToResponse = (block: CreateNotebookBlockInput): NotebookFields['blocks'][number] => {
    switch (block.type) {
        case NotebookBlockType.MARKDOWN: {
            return { __typename: 'MarkdownBlock', id: block.id, markdownInput: block.markdownInput ?? '' }
        }
        case NotebookBlockType.QUERY: {
            return { __typename: 'QueryBlock', id: block.id, queryInput: block.queryInput ?? '' }
        }
        case NotebookBlockType.FILE: {
            return {
                __typename: 'FileBlock',
                id: block.id,
                fileInput: {
                    __typename: 'FileBlockInput',
                    repositoryName: block.fileInput?.repositoryName ?? '',
                    filePath: block.fileInput?.filePath ?? '',
                    revision: block.fileInput?.revision ?? '',
                    lineRange: {
                        __typename: 'FileBlockLineRange',
                        startLine: block.fileInput?.lineRange?.startLine ?? 0,
                        endLine: block.fileInput?.lineRange?.endLine ?? 1,
                    },
                },
            }
        }
        case NotebookBlockType.SYMBOL: {
            return {
                __typename: 'SymbolBlock',
                id: block.id,
                symbolInput: {
                    __typename: 'SymbolBlockInput',
                    repositoryName: block.symbolInput?.repositoryName ?? '',
                    filePath: block.symbolInput?.filePath ?? '',
                    revision: block.symbolInput?.revision ?? '',
                    lineContext: block.symbolInput?.lineContext ?? 3,
                    symbolName: block.symbolInput?.symbolName ?? '',
                    symbolContainerName: block.symbolInput?.symbolContainerName ?? '',
                    symbolKind: block.symbolInput?.symbolKind ?? SymbolKind.UNKNOWN,
                },
            }
        }
    }
}

const mockSymbolStreamEvents: SearchEvent[] = [
    {
        type: 'matches',
        data: [
            {
                type: 'symbol',
                repository: 'github.com/sourcegraph/sourcegraph',
                path: 'client/web/index.ts',
                commit: 'branch',
                symbols: [
                    {
                        name: 'func',
                        containerName: 'class',
                        url: 'https://sourcegraph.com/github.com/sourcegraph/sourcegraph@branch/-/blob/client/web/index.ts?L1:1-1:3',
                        kind: SymbolKind.FUNCTION,
                        line: 1,
                    },
                ],
            },
        ],
    },
    { type: 'done', data: {} },
]

const mockFilePathStreamEvents: SearchEvent[] = [
    {
        type: 'matches',
        data: [
            {
                type: 'path',
                repository: 'github.com/sourcegraph/sourcegraph',
                path: 'client/web/index.ts',
                commit: 'branch',
            },
        ],
    },
    { type: 'done', data: {} },
]

const commonSearchGraphQLResults: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
    ...commonWebGraphQlResults,
    ...highlightFileResult,
    ...viewerSettings,
    ResolveRepoRev: () => createResolveRepoRevisionResult('/github.com/sourcegraph/sourcegraph'),
    FetchNotebook: ({ id }) => ({
        node: notebookFixture(id, 'Notebook Title', [
            { __typename: 'MarkdownBlock', id: '1', markdownInput: '# Title' },
            { __typename: 'QueryBlock', id: '2', queryInput: 'query' },
        ]),
    }),
    UpdateNotebook: ({ id, notebook }) => ({
        updateNotebook: notebookFixture(id, notebook.title, notebook.blocks.map(GQLBlockInputToResponse)),
    }),
    ListNotebooks: () => ({
        notebooks: {
            totalCount: 1,
            nodes: [notebookFixture('id', 'Title', [])],
            pageInfo: { endCursor: null, hasNextPage: false },
        },
    }),
}

describe('Search Notebook', () => {
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
        testContext.overrideGraphQL(commonSearchGraphQLResults)
        testContext.overrideSearchStreamEvents(mixedSearchStreamEvents)
        testContext.overrideJsContext({
            temporarySettings: {
                __typename: 'TemporarySettings',
                contents: JSON.stringify({
                    'search.notebooks.gettingStartedTabSeen': true,
                }),
            },
        })
    })
    afterEachSaveScreenshotIfFailed(() => driver.page)
    afterEach(() => testContext?.dispose())

    const getBlockIds = () =>
        driver.page.evaluate(() => {
            const blockElements = [...document.querySelectorAll('[data-block-id]')] as HTMLElement[]
            return blockElements.map((block: HTMLElement) => {
                if (!block.dataset.blockId) {
                    throw new Error('Invalid block id')
                }
                return block.dataset.blockId
            })
        })

    const blockSelector = (id: string) => `[data-block-id="${id}"]`

    const selectBlock = (id: string) => driver.page.click(blockSelector(id))

    const runBlockMenuAction = (id: string, actionLabel: string) =>
        driver.page.click(`${blockSelector(id)} [data-testid="${actionLabel}"]`)

    const addNewBlock = (type: BlockType) =>
        driver.page.click(`[data-testid="notebook-command-palette"] [data-testid="add-${type}-block"]`)

    const getFileBlockHeaderText = async (fileBlockSelector: string) => {
        const fileBlockHeaderSelector = `${fileBlockSelector} [data-testid="file-block-header"]`
        await driver.page.waitForSelector(fileBlockHeaderSelector, { visible: true })
        const fileBlockHeaderText = await driver.page.evaluate(
            fileBlockHeaderSelector => document.querySelector<HTMLDivElement>(fileBlockHeaderSelector)?.textContent,
            fileBlockHeaderSelector
        )
        return fileBlockHeaderText
    }

    const getRenderedMarkdownText = async (mdBlockSelector: string) => {
        const markdownOutputSelector = `${mdBlockSelector} [data-testid="output"]`
        await driver.page.waitForSelector(markdownOutputSelector, { visible: true })
        return driver.page.evaluate(
            markdownOutputSelector => document.querySelector<HTMLElement>(markdownOutputSelector)?.textContent,
            markdownOutputSelector
        )
    }

    it('Should render a notebook', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks/n1')
        await driver.page.waitForSelector('[data-block-id]', { visible: true })
        const blockIds = await getBlockIds()
        expect(blockIds).toHaveLength(2)
        await percySnapshotWithVariants(driver.page, 'Search notebook')
        await accessibilityAudit(driver.page)
    })

    it('Should move, duplicate, and delete blocks', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks/n1')
        await driver.page.waitForSelector('[data-block-id]', { visible: true })
        const blockIds = await getBlockIds()

        await selectBlock(blockIds[0])
        await runBlockMenuAction(blockIds[0], 'Move Down')

        expect(await getBlockIds()).toStrictEqual([blockIds[1], blockIds[0]])

        await runBlockMenuAction(blockIds[0], 'Move Up')
        expect(await getBlockIds()).toStrictEqual(blockIds)

        await runBlockMenuAction(blockIds[0], 'Duplicate')
        const blockIdsAfterDuplicate = await getBlockIds()
        expect(await getBlockIds()).toHaveLength(3)

        for (const blockId of blockIdsAfterDuplicate) {
            await selectBlock(blockId)
            await runBlockMenuAction(blockId, 'Delete')
        }
        expect(await getBlockIds()).toHaveLength(0)
    })

    it('Should add markdown and query blocks, edit, and run them', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks/n1')
        await driver.page.waitForSelector('[data-block-id]', { visible: true })

        await addNewBlock('md')
        await addNewBlock('query')

        const blockIds = await getBlockIds()
        expect(blockIds).toHaveLength(4)

        const newMarkdownBlockSelector = blockSelector(blockIds[2])
        const newQueryBlockSelector = blockSelector(blockIds[3])

        // Edit and run new markdown block
        await driver.page.click(newMarkdownBlockSelector)
        await driver.replaceText({
            selector: `${newMarkdownBlockSelector} .cm-content`,
            newText: 'Replaced text',
            selectMethod: 'keyboard',
            enterTextMethod: 'paste',
        })
        await driver.page.click('[data-testid="Render"]')

        const renderedMarkdownText = await getRenderedMarkdownText(newMarkdownBlockSelector)
        expect(renderedMarkdownText?.trim()).toEqual('Replaced text')

        // Edit and run new query block
        const inputSelector = `${newQueryBlockSelector} [data-testid="notebook-query-block-input"]`
        await driver.page.click(inputSelector)
        await driver.replaceText({
            selector: inputSelector,
            newText: 'repo:test query',
            selectMethod: 'keyboard',
            enterTextMethod: 'paste',
        })
        await driver.page.click('[data-testid="Run search"]')

        const queryResultContainerSelector = `${newQueryBlockSelector} [data-testid="result-container"]`
        await driver.page.waitForSelector(queryResultContainerSelector, { visible: true })
        const isResultContainerVisible = await driver.page.evaluate(
            queryResultContainerSelector => document.querySelector(queryResultContainerSelector) !== null,
            queryResultContainerSelector
        )
        expect(isResultContainerVisible).toBeTruthy()
        await percySnapshotWithVariants(driver.page, 'Search notebook with markdown and query blocks')
        await accessibilityAudit(driver.page)
    })

    it('Should add file block and edit it', async () => {
        testContext.overrideSearchStreamEvents(mockFilePathStreamEvents)

        await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks/n1')
        await driver.page.waitForSelector('[data-block-id]', { visible: true })

        await addNewBlock('file')

        const blockIds = await getBlockIds()
        expect(blockIds).toHaveLength(3)

        const fileBlockSelector = blockSelector(blockIds[2])

        // Edit new file block
        const inputSelector = `${fileBlockSelector} [data-testid="notebook-file-block-input"]`
        await driver.page.click(inputSelector)
        await driver.replaceText({
            selector: inputSelector,
            newText: 'client/web/file.tsx',
            selectMethod: 'keyboard',
            enterTextMethod: 'paste',
        })

        // Wait for file suggestion button and click it
        await driver.page.waitForSelector(`${fileBlockSelector} [data-testid="file-suggestion-button"]`, {
            visible: true,
        })
        await driver.page.click(`${fileBlockSelector} [data-testid="file-suggestion-button"]`)

        await driver.replaceText({
            selector: `[id="${blockIds[2]}-line-range-input"]`,
            newText: '1-20',
            selectMethod: 'keyboard',
            enterTextMethod: 'paste',
        })

        // Wait for header to update to load
        await driver.page.waitForFunction(
            (fileBlockSelector: string) => {
                const fileBlockHeaderSelector = `${fileBlockSelector} [data-testid="file-block-header"]`
                return document.querySelector<HTMLDivElement>(fileBlockHeaderSelector)?.textContent?.includes('/')
            },
            {},
            fileBlockSelector
        )

        // Save the inputs
        await driver.page.click('[data-testid="Save"]')

        const fileBlockHeaderText = await getFileBlockHeaderText(fileBlockSelector)
        expect(fileBlockHeaderText).toEqual('sourcegraph/sourcegraph›client/web/index.ts')
    })

    it('Should add file block and auto-fill the inputs when pasting a file URL', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks/n1')
        await driver.page.waitForSelector('[data-block-id]', { visible: true })

        await addNewBlock('file')

        const blockIds = await getBlockIds()
        expect(blockIds).toHaveLength(3)

        const fileBlockSelector = blockSelector(blockIds[2])

        // Edit new file block
        await driver.page.click(fileBlockSelector)

        // Simulate pasting the file URL
        await driver.page.evaluate(
            (fileBlockSelector: string, fileURL: string) => {
                const dataTransfer = new DataTransfer()
                dataTransfer.setData('text', fileURL)
                const event = new ClipboardEvent('paste', {
                    clipboardData: dataTransfer,
                    bubbles: true,
                })
                const element = document.querySelector(fileBlockSelector)
                element?.dispatchEvent(event)
            },
            fileBlockSelector,
            'https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/blob/client/search/src/index.ts?L30-32'
        )

        // Wait for highlighted code to load
        await driver.page.waitForSelector(`${fileBlockSelector} td.line`, { visible: true })

        // Refocus the entire block (prevents jumping content for below actions)
        await driver.page.click(fileBlockSelector)

        // Save the inputs
        await driver.page.click('[data-testid="Save"]')

        const fileBlockHeaderText = await getFileBlockHeaderText(fileBlockSelector)
        expect(fileBlockHeaderText).toEqual('sourcegraph/sourcegraph›client/search/src/index.ts')
    })

    it('Should update the notebook title', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks/n1')
        await driver.page.waitForSelector('[data-block-id]', { visible: true })

        await driver.page.click('[data-testid="notebook-title-button"]')

        await driver.page.waitForSelector('[data-testid="notebook-title-input"]')
        await driver.enterText('type', ' Edited')

        await driver.page.keyboard.press('Enter')
        await driver.page.waitForSelector('[data-testid="notebook-title-button"]')
        const titleText = await driver.page.evaluate(
            () => document.querySelector<HTMLButtonElement>('[data-testid="notebook-title-button"]')?.textContent
        )
        expect(titleText).toEqual('Notebook Title Edited')
    })

    it('Should open the share dialog, switch the share option, and close the dialog', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks/n1')
        await driver.page.waitForSelector('[data-testid="share-notebook-button"]', { visible: true })

        await driver.page.click('[data-testid="share-notebook-button"]')
        await driver.page.waitForSelector('[data-testid="share-notebook-options-dropdown-toggle"]', { visible: true })

        await driver.page.click('[data-testid="share-notebook-options-dropdown-toggle"]')
        await driver.page.waitForSelector('[data-testid="share-notebook-option-test-true"]')
        await driver.page.click('[data-testid="share-notebook-option-test-true"]')

        await driver.page.click('[data-testid="share-notebook-done-button"]')
    })

    afterEach(() => {
        const exportedNotebookPath = path.resolve(downloadPath, 'Exported.snb.md')
        // eslint-disable-next-line no-sync
        if (fs.existsSync(exportedNotebookPath)) {
            // eslint-disable-next-line no-sync
            fs.unlinkSync(exportedNotebookPath)
        }
    })

    it.skip('Should export a notebook as Markdown file and import it back', async () => {
        testContext.overrideGraphQL({
            ...commonSearchGraphQLResults,
            FetchNotebook: ({ id }) => ({
                node: notebookFixture(id, 'Exported', [
                    { __typename: 'MarkdownBlock', id: '1', markdownInput: '# Title' },
                    { __typename: 'QueryBlock', id: '2', queryInput: 'query' },
                    {
                        __typename: 'FileBlock',
                        id: '3',
                        fileInput: {
                            __typename: 'FileBlockInput',
                            repositoryName: 'github.com/sourcegraph/sourcegraph',
                            filePath: 'client/web/index.ts',
                            revision: 'main',
                            lineRange: { __typename: 'FileBlockLineRange', startLine: 1, endLine: 10 },
                        },
                    },
                    {
                        __typename: 'SymbolBlock',
                        id: '4',
                        symbolInput: {
                            __typename: 'SymbolBlockInput',
                            repositoryName: 'github.com/sourcegraph/sourcegraph',
                            filePath: 'client/web/index.ts',
                            revision: 'branch',
                            symbolName: 'func',
                            symbolContainerName: 'class',
                            symbolKind: SymbolKind.FUNCTION,
                            lineContext: 3,
                        },
                    },
                ]),
            }),
            CreateNotebook: ({ notebook }) => ({
                createNotebook: notebookFixture(
                    'importedId',
                    notebook.title,
                    notebook.blocks.map(GQLBlockInputToResponse)
                ),
            }),
        })
        testContext.overrideSearchStreamEvents(mockSymbolStreamEvents)

        const expectedExportedMarkdown = `# Title

\`\`\`sourcegraph
query
\`\`\`

${process.env.SOURCEGRAPH_BASE_URL}/github.com/sourcegraph/sourcegraph@main/-/blob/client/web/index.ts?L2-10

${process.env.SOURCEGRAPH_BASE_URL}/github.com/sourcegraph/sourcegraph@branch/-/blob/client/web/index.ts?L1:1-1:3#symbolName=func&symbolContainerName=class&symbolKind=FUNCTION&lineContext=3
`

        await driver.page.client().send('Page.setDownloadBehavior', { behavior: 'allow', downloadPath })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks/n1')
        await driver.page.waitForSelector('[data-testid="export-notebook-markdown-button"]', { visible: true })

        await driver.page.click('[data-testid="export-notebook-markdown-button"]')
        // Wait for the download to complete.
        await driver.page.waitForTimeout(1000)

        const exportedNotebookPath = path.resolve(downloadPath, 'Exported.snb.md')
        // eslint-disable-next-line no-sync
        expect(fs.existsSync(exportedNotebookPath)).toBeTruthy()

        // eslint-disable-next-line no-sync
        const exportedNotebookFileContents = fs.readFileSync(exportedNotebookPath, 'utf-8')
        expect(exportedNotebookFileContents).toEqual(expectedExportedMarkdown)

        // Navigate to the notebooks list page to import the notebook.
        await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks')
        await driver.page.waitForSelector('[data-testid="import-markdown-notebook-file-input"]')
        const fileInputElement = await driver.page.$('[data-testid="import-markdown-notebook-file-input"]')
        await fileInputElement?.uploadFile(exportedNotebookPath)
        await fileInputElement?.evaluate(upload => upload.dispatchEvent(new Event('change', { bubbles: true })))

        // Should be redirected to the notebook page.
        await driver.page.waitForSelector('[data-block-id]', { visible: true })
        // Verify the redirected URL contains the imported notebook id.
        expect(driver.page.url()).toContain('/notebooks/importedId')
    })

    it('Should copy the notebook', async () => {
        testContext.overrideGraphQL({
            ...commonSearchGraphQLResults,
            CreateNotebook: ({ notebook }) => ({
                createNotebook: notebookFixture(
                    'copiedId',
                    notebook.title,
                    notebook.blocks.map(GQLBlockInputToResponse)
                ),
            }),
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks/n1')
        await driver.page.waitForSelector('[data-testid="copy-notebook-button"]', { visible: true })

        await Promise.all([
            // We should be redirected to the copied notebook page, wait for the navigation.
            driver.page.waitForNavigation({ waitUntil: 'networkidle0' }),
            driver.page.click('[data-testid="copy-notebook-button"]'),
        ])

        // Wait for blocks to load.
        await driver.page.waitForSelector('[data-block-id]', { visible: true })

        // Verify the redirected URL contains the copied notebook id.
        expect(driver.page.url()).toContain('/notebooks/copiedId')
    })

    it('Should add symbol block and edit it', async () => {
        testContext.overrideSearchStreamEvents(mockSymbolStreamEvents)

        await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks/n1')
        await driver.page.waitForSelector('[data-block-id]', { visible: true })

        await addNewBlock('symbol')

        const blockIds = await getBlockIds()
        expect(blockIds).toHaveLength(3)

        const symbolBlockSelector = blockSelector(blockIds[2])

        // Edit new symbol block
        const inputSelector = `${symbolBlockSelector} [data-testid="notebook-symbol-block-input"]`
        await driver.page.click(inputSelector)
        await driver.replaceText({
            selector: inputSelector,
            newText: 'func',
            selectMethod: 'keyboard',
            enterTextMethod: 'paste',
        })

        // Wait for symbol suggestion button and click it
        await driver.page.waitForSelector(`${symbolBlockSelector} [data-testid="symbol-suggestion-button"]`, {
            visible: true,
        })
        await driver.page.click(`${symbolBlockSelector} [data-testid="symbol-suggestion-button"]`)

        // Wait for highlighted code to load
        await driver.page.waitForSelector(`${symbolBlockSelector} td.line`, { visible: true })

        // Refocus the entire block (prevents jumping content for below actions)
        await driver.page.click(symbolBlockSelector)

        // Save the inputs
        await driver.page.click('[data-testid="Save"]')

        const symbolBlockSelectedSymbolNameSelector = `${symbolBlockSelector} [data-testid="selected-symbol-name"]`
        const selectedSymbolName = await driver.page.evaluate(
            symbolBlockSelectedSymbolNameSelector =>
                document.querySelector<HTMLDivElement>(symbolBlockSelectedSymbolNameSelector)?.textContent,
            symbolBlockSelectedSymbolNameSelector
        )
        expect(selectedSymbolName).toEqual('func')
    })

    it('Should add an empty markdown block through the command palette', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks/n1')
        await driver.page.waitForSelector('[data-block-id]', { visible: true })

        // Focus the input and use the slash command to filter available options
        await driver.page.click('[data-testid="command-palette-input"]')
        await driver.replaceText({
            selector: '[data-testid="command-palette-input"]',
            newText: '/markdown',
            selectMethod: 'keyboard',
            enterTextMethod: 'paste',
        })

        // Wait for the command palette option to show up and select it
        await driver.page.waitForSelector('[data-option-id="add-md-block"]', { visible: true })
        await driver.page.click('[data-option-id="add-md-block"]')

        // A new block should appear
        const blockIds = await getBlockIds()
        expect(blockIds).toHaveLength(3)
    })

    it('Should add a markdown block with initial input through the command palette', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks/n1')
        await driver.page.waitForSelector('[data-block-id]', { visible: true })

        // Focus the input and enter the markdown text
        await driver.page.click('[data-testid="command-palette-input"]')
        await driver.replaceText({
            selector: '[data-testid="command-palette-input"]',
            newText: 'Markdown Text',
            selectMethod: 'keyboard',
            enterTextMethod: 'paste',
        })

        // Wait for the command palette option to show up and select it
        await driver.page.waitForSelector('[data-option-id="add-md-block-with-input"]', { visible: true })
        await driver.page.click('[data-option-id="add-md-block-with-input"]')

        // A new block should appear
        const blockIds = await getBlockIds()
        expect(blockIds).toHaveLength(3)
        const mdBlockSelector = blockSelector(blockIds[2])

        // Render the markdown and check the output
        await driver.page.click('[data-testid="Render"]')
        const renderedMarkdownText = await getRenderedMarkdownText(mdBlockSelector)
        expect(renderedMarkdownText?.trim()).toEqual('Markdown Text')
    })

    it('Should add a pre-populated file block when pasting a file URL in command palette', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks/n1')
        await driver.page.waitForSelector('[data-block-id]', { visible: true })

        // Focus the input and paste the file URL
        await driver.page.click('[data-testid="command-palette-input"]')
        await driver.replaceText({
            selector: '[data-testid="command-palette-input"]',
            newText:
                'https://sourcegraph.com/github.com/sourcegraph/sourcegraph@main/-/blob/client/search/src/index.ts?L30-32',
            selectMethod: 'keyboard',
            enterTextMethod: 'paste',
        })

        // Wait for the command palette option to show up and select it
        await driver.page.waitForSelector('[data-option-id="add-file-from-url"]', { visible: true })
        await driver.page.click('[data-option-id="add-file-from-url"]')

        // A new block should appear
        const blockIds = await getBlockIds()
        expect(blockIds).toHaveLength(3)
        const fileBlockSelector = blockSelector(blockIds[2])

        // Wait for highlighted code to load
        await driver.page.waitForSelector(`${fileBlockSelector} td.line`, { visible: true })

        const fileBlockHeaderText = await getFileBlockHeaderText(fileBlockSelector)
        expect(fileBlockHeaderText).toEqual('sourcegraph/sourcegraph›client/search/src/index.ts')
    })

    const getHighlightedOutlineHeading = async () => {
        const selector = '[data-testid="notebook-outline"] [aria-current="true"]'
        await driver.page.waitForSelector(selector, { visible: true })
        return driver.page.evaluate(selector => document.querySelector<HTMLElement>(selector)?.dataset.id, selector)
    }

    it('Should highlight the top visible heading in the outline', async () => {
        const fillerText = 'Lorem ipsum dolor sit amet. '.repeat(256)
        // HTML should not affect heading ids.
        const markdownInput = `# Title <i>1</i>\n\n${fillerText}\n# Title <i>2</i>\n\n${fillerText}\n`

        testContext.overrideGraphQL({
            ...commonSearchGraphQLResults,
            FetchNotebook: ({ id }) => ({
                node: notebookFixture(id, 'Outline', [{ __typename: 'MarkdownBlock', id: 'id-1', markdownInput }]),
            }),
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks/n1')
        await driver.page.waitForSelector('[data-block-id]', { visible: true })
        // The first "Title 1" heading should be highlighted
        await driver.page.waitForSelector('[data-id="title-1-id-1"][aria-current="true"]')
        expect(await getHighlightedOutlineHeading()).toEqual('title-1-id-1')

        // Click on "Title 2" link in the outline
        await driver.page.click('[data-id="title-2-id-1"] a')

        // The "Title 2" heading should be visible
        await driver.page.waitForSelector('h1#title-2-id-1', { visible: true })

        // The "Title 2" heading in the outline should be highlighted
        await driver.page.waitForSelector('[data-id="title-2-id-1"][aria-current="true"]')
        expect(await getHighlightedOutlineHeading()).toEqual('title-2-id-1')
    })

    it('Should scroll to the heading specified in the hash on page load', async () => {
        const fillerText = 'Lorem ipsum dolor sit amet. '.repeat(256)
        // HTML should not affect heading ids.
        const markdownInput = `# Title <i>1</i>\n\n${fillerText}\n# Title <i>2</i>\n\n${fillerText}\n`

        testContext.overrideGraphQL({
            ...commonSearchGraphQLResults,
            FetchNotebook: ({ id }) => ({
                node: notebookFixture(id, 'Outline', [{ __typename: 'MarkdownBlock', id: 'id-1', markdownInput }]),
            }),
        })

        await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks/n1#title-2-id-1')
        await driver.page.waitForSelector('[data-block-id]', { visible: true })

        // The "Title 2" heading should be visible
        await driver.page.waitForSelector('h1#title-2-id-1', { visible: true })

        // The "Title 2" heading in the outline should be highlighted
        await driver.page.waitForSelector('[data-id="title-2-id-1"][aria-current="true"]')
        expect(await getHighlightedOutlineHeading()).toEqual('title-2-id-1')
    })

    it('Notebooks list page should be accessible', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks?tab=notebooks')
        await driver.page.waitForSelector('[data-testid="filtered-connection-nodes"]', { visible: true })
        await percySnapshotWithVariants(driver.page, 'Notebooks list')
        await accessibilityAudit(driver.page)
    })
})
