import { describe } from 'mocha'
import { createDriverForTest, Driver, percySnapshot } from '../../../shared/src/testing/driver'
import { createWebIntegrationTestContext, WebIntegrationTestContext } from './context'
import { commonSearchGraphQLResults } from './search.test'
import { getCommonBlobGraphQlResults } from './blob-viewer.test'

describe('Visual tests', () => {
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

    describe('Search page', () => {
        it('Renders landing page correctly', async () => {
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search')
            await driver.page.waitForSelector('#monaco-query-input', { visible: true })
            await percySnapshot(driver.page, 'Search page')
            await percySnapshot(driver.page, 'Search page', { theme: 'theme-dark' })
        })

        it('Renders search result page correctly', async () => {
            testContext.overrideGraphQL({
                ...commonSearchGraphQLResults,
            })
            await driver.page.goto(driver.sourcegraphBaseUrl + '/search?q=foo')
            await driver.page.waitForSelector('#monaco-query-input')
            await percySnapshot(driver.page, 'Search results page')
            await percySnapshot(driver.page, 'Search results page', { theme: 'theme-dark' })
        })
    })

    describe('Blob page', () => {
        const repositoryName = 'github.com/sourcegraph/jsonrpc2'
        const repositorySourcegraphUrl = `/${repositoryName}`
        const fileName = 'test.ts'
        const commonBlobGraphQlResults = getCommonBlobGraphQlResults(repositoryName, repositorySourcegraphUrl, fileName)

        beforeEach(() => {
            testContext.overrideGraphQL(commonBlobGraphQlResults)
        })

        it('Renders blob page correctly', async () => {
            await driver.page.goto(`${driver.sourcegraphBaseUrl}/${repositoryName}/-/blob/${fileName}`)
            await driver.page.waitForSelector('.test-repo-blob')
            await percySnapshot(driver.page, 'Blob page')
            await percySnapshot(driver.page, 'Blob page', { theme: 'theme-dark' })
        })
    })
})
