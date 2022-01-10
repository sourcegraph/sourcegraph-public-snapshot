import { subDays } from 'date-fns'
import expect from 'expect'

import { NotebookBlockType, SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { WebGraphQlOperations } from '../graphql-operations'
import { BlockType } from '../search/notebook'

import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import { createRepositoryRedirectResult } from './graphQlResponseHelpers'
import { commonWebGraphQlResults } from './graphQlResults'
import { siteGQLID, siteID } from './jscontext'
import { highlightFileResult, mixedSearchStreamEvents } from './streaming-search-mocks'
import { percySnapshotWithVariants } from './utils'

const viewerSettings: Partial<WebGraphQlOperations> = {
    ViewerSettings: () => ({
        viewerSettings: {
            __typename: 'SettingsCascade',
            subjects: [
                {
                    __typename: 'DefaultSettings',
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

const commonSearchGraphQLResults: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
    ...commonWebGraphQlResults,
    ...highlightFileResult,
    RepositoryRedirect: ({ repoName }) => createRepositoryRedirectResult(repoName),
    FetchNotebook: ({ id }) => ({
        node: {
            __typename: 'Notebook',
            id,
            title: 'Notebook Title',
            createdAt: subDays(now, 5).toISOString(),
            updatedAt: subDays(now, 5).toISOString(),
            public: true,
            viewerCanManage: true,
            creator: { __typename: 'User', username: 'user1' },
            blocks: [
                { __typename: 'MarkdownBlock', id: '1', markdownInput: '# Title' },
                { __typename: 'QueryBlock', id: '2', queryInput: 'query' },
            ],
        },
    }),
    UpdateNotebook: ({ id, notebook }) => ({
        updateNotebook: {
            __typename: 'Notebook',
            id,
            title: notebook.title,
            createdAt: subDays(now, 5).toISOString(),
            updatedAt: subDays(now, 5).toISOString(),
            public: notebook.public,
            viewerCanManage: true,
            creator: { __typename: 'User', username: 'user1' },
            blocks: notebook.blocks.map(block => {
                switch (block.type) {
                    case NotebookBlockType.MARKDOWN:
                        return { __typename: 'MarkdownBlock', id: block.id, markdownInput: block.markdownInput ?? '' }
                    case NotebookBlockType.QUERY:
                        return { __typename: 'QueryBlock', id: block.id, queryInput: block.queryInput ?? '' }
                    case NotebookBlockType.FILE:
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
            }),
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
        testContext.overrideGraphQL({ ...commonSearchGraphQLResults, ...viewerSettings })
        testContext.overrideSearchStreamEvents(mixedSearchStreamEvents)
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
        driver.page.click(`[data-testid="always-visible-add-block-buttons"] [data-testid="add-${type}-button"]`)

    it('Should render a notebook', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks/n1')
        await driver.page.waitForSelector('[data-block-id]', { visible: true })
        const blockIds = await getBlockIds()
        expect(blockIds).toHaveLength(2)
        await percySnapshotWithVariants(driver.page, 'Search notebook')
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
        await driver.page.click('[data-testid="Edit"]')
        await driver.replaceText({
            selector: `${newMarkdownBlockSelector} .monaco-editor`,
            newText: 'Replaced text',
            selectMethod: 'keyboard',
            enterTextMethod: 'paste',
        })
        await driver.page.click('[data-testid="Render"]')

        const markdownOutputSelector = `${newMarkdownBlockSelector} [data-testid="output"]`
        await driver.page.waitForSelector(markdownOutputSelector, { visible: true })
        const renderedMarkdownText = await driver.page.evaluate(
            markdownOutputSelector => document.querySelector<HTMLElement>(markdownOutputSelector)?.textContent,
            markdownOutputSelector
        )
        expect(renderedMarkdownText?.trim()).toEqual('Replaced text')

        // Edit and run new query block
        await driver.page.click(`${newQueryBlockSelector} .monaco-editor`)
        await driver.replaceText({
            selector: `${newQueryBlockSelector} .monaco-editor`,
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
    })

    it('Should add file block and edit it', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/notebooks/n1')
        await driver.page.waitForSelector('[data-block-id]', { visible: true })

        await addNewBlock('file')

        const blockIds = await getBlockIds()
        expect(blockIds).toHaveLength(3)

        const fileBlockSelector = blockSelector(blockIds[2])

        // Edit new file block
        await driver.page.click(fileBlockSelector)

        await driver.replaceText({
            selector: `${fileBlockSelector} [data-testid="file-block-repository-name-input"]`,
            newText: 'github.com/sourcegraph/sourcegraph',
            selectMethod: 'keyboard',
            enterTextMethod: 'paste',
        })
        // Wait for input to validate
        await driver.page.waitForSelector(
            `${fileBlockSelector} [data-testid="file-block-repository-name-input"].is-valid`
        )

        await driver.replaceText({
            selector: `${fileBlockSelector} [data-testid="file-block-file-path-input"]`,
            newText: 'client/web/file.tsx',
            selectMethod: 'keyboard',
            enterTextMethod: 'paste',
        })
        // Wait for input to validate
        await driver.page.waitForSelector(`${fileBlockSelector} [data-testid="file-block-file-path-input"].is-valid`)

        // Wait for highlighted code to load
        await driver.page.waitForSelector(`${fileBlockSelector} td.line`, { visible: true })

        // Refocus the entire block (prevents jumping content for below actions)
        await driver.page.click(fileBlockSelector)

        // Save the inputs
        await driver.page.click('[data-testid="Save"]')

        const fileBlockHeaderSelector = `${fileBlockSelector} [data-testid="file-block-header"]`
        await driver.page.waitForSelector(fileBlockHeaderSelector, { visible: true })
        const fileBlockHeaderText = await driver.page.evaluate(
            fileBlockHeaderSelector => document.querySelector<HTMLDivElement>(fileBlockHeaderSelector)?.textContent,
            fileBlockHeaderSelector
        )
        expect(fileBlockHeaderText).toEqual('github.com/sourcegraph/sourcegraph/client/web/file.tsx')
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
})
