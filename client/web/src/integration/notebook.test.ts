import expect from 'expect'

import { SharedGraphQlOperations } from '@sourcegraph/shared/src/graphql-operations'
import { Driver, createDriverForTest } from '@sourcegraph/shared/src/testing/driver'
import { afterEachSaveScreenshotIfFailed } from '@sourcegraph/shared/src/testing/screenshotReporter'

import { WebGraphQlOperations } from '../graphql-operations'
import { BlockType } from '../search/notebook'

import { WebIntegrationTestContext, createWebIntegrationTestContext } from './context'
import { commonWebGraphQlResults } from './graphQlResults'
import { siteGQLID, siteID } from './jscontext'
import { highlightFileResult, mixedSearchStreamEvents } from './streaming-search-mocks'

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

const commonSearchGraphQLResults: Partial<WebGraphQlOperations & SharedGraphQlOperations> = {
    ...commonWebGraphQlResults,
    ...highlightFileResult,
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

    const addNewBlock = (type: BlockType) => driver.page.click(`[data-testid="add-${type}-button"]`)

    it('Should render a notebook with two default blocks', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search/notebook')
        await driver.page.waitForSelector('[data-block-id]', { visible: true })
        const blockIds = await getBlockIds()
        expect(blockIds).toHaveLength(2)
    })

    it('Should move, duplicate, and delete blocks', async () => {
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search/notebook')
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
        await driver.page.goto(driver.sourcegraphBaseUrl + '/search/notebook')
        await driver.page.waitForSelector('[data-block-id]', { visible: true })

        await addNewBlock('md')
        await addNewBlock('query')

        const blockIds = await getBlockIds()
        expect(blockIds).toHaveLength(4)

        const newMarkdownBlockSelector = blockSelector(blockIds[2])
        const newQueryBlockSelector = blockSelector(blockIds[3])

        // Edit and run new markdown cell
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

        // Edit and run new query cell
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
    })
})
