import { describe as mochaDescribe, test as mochaTest, before as mochaBefore, after as mochaAfter } from 'mocha'
import { Subscription } from 'rxjs'
import { snakeCase, noop } from 'lodash'
import { readFile, writeFile, exists } from 'mz/fs'
import { createDriverForTest, Driver } from '../../../shared/src/e2e/driver'
import * as path from 'path'
import mkdirp from 'mkdirp-promise'
import express from 'express'
import { isDefined } from '../../../shared/src/util/types'
import { readEnvBoolean } from '../../../shared/src/e2e/e2e-test-utils'

const FIXTURES_DIRECTORY = `${__dirname}/__fixtures__`
const ASSETS_DIRECTORY = `${__dirname}/../../../ui/assets`

type IntegrationTestInitGeneration = () => Promise<{
    driver: Driver
    sourcegraphBaseUrl: string
    subscriptions?: Subscription
}>

type IntegrationTest = (
    testName: string,
    run: (options: { sourcegraphBaseUrl: string; driver: Driver }) => Promise<void>
) => void

type IntegrationTestBeforeGeneration = (
    setupLogic: (options: { sourcegraphBaseUrl: string; driver: Driver }) => Promise<void>
) => void

type IntegrationTestDescribe = (
    description: string,
    suite: (helpers: {
        before: IntegrationTestBeforeGeneration
        test: IntegrationTest
        describe: IntegrationTestDescribe
    }) => void
) => void

type IntegrationTestSuite = (helpers: {
    /**
     * Registers a calback that will be run before test data is generated,
     * responsible for creating the {@link Driver} and providing the Sourcegraph URL.
     */
    initGeneration: (setupLogic: IntegrationTestInitGeneration) => void
    test: IntegrationTest
    describe: IntegrationTestDescribe
}) => void

interface RequestStubs {
    [requestKey: string]: {
        status: number
        contentType: string
        body: string
    }
}

const isAssetsURL = (url: URL): boolean => url.pathname.startsWith('/.assets')

/**
 * Generates the fixture data for an integration test suite
 * by running the tests in Puppeteer against a real backend.
 *
 * Captured responses will be saved to a file in the `__fixtures__` directory.
 */
function generateIntegrationTestsData(description: string, suite: IntegrationTestSuite): void {
    mochaDescribe(description, () => {
        let driver: Driver
        let sourcegraphBaseUrl: string
        const subscriptions = new Subscription()
        const test = (prefixes: string[]): IntegrationTest => (testName, run) => {
            mochaTest(testName, async () => {
                await driver.newPage()
                const requestStubs: RequestStubs = {}

                // eslint-disable-next-line @typescript-eslint/no-misused-promises
                driver.page.on('response', async response => {
                    // Intercept and save responses.
                    const contentType = response.headers()['content-type']
                    if (!contentType || contentType === 'image/svg+xml') {
                        return
                    }
                    const requestURL = new URL(response.request().url())
                    if (isAssetsURL(requestURL)) {
                        return
                    }
                    if (requestURL.pathname.startsWith('/.api') && requestURL.search === '?logEvent') {
                        return
                    }
                    const requestKey = [
                        requestURL.pathname,
                        requestURL.search,
                        requestURL.hash,
                        response.request().postData(),
                    ]
                        .filter(isDefined)
                        .map(s => s.replace(/\s+/g, ''))
                        .join('-')
                    requestStubs[requestKey] = {
                        status: response.status(),
                        contentType,
                        body: await response.text(),
                    }
                })
                await run({ sourcegraphBaseUrl, driver })
                await driver.page.close()

                // Save all responses as JSON in the __fixtures__ directory.
                const fixtureDirectory = path.join(FIXTURES_DIRECTORY, ...prefixes.map(snakeCase))
                await mkdirp(fixtureDirectory)
                await writeFile(
                    path.join(fixtureDirectory, `${snakeCase(testName)}.json`),
                    JSON.stringify(requestStubs, null, 4)
                )
            })
        }
        const before: IntegrationTestBeforeGeneration = setupLogic => {
            mochaBefore(() => setupLogic({ driver, sourcegraphBaseUrl }))
        }
        const describe = (prefixes: string[]): IntegrationTestDescribe => (description, suite) => {
            mochaDescribe(description, () => {
                suite({
                    describe: describe([...prefixes, description]),
                    before,
                    test: test([...prefixes, description]),
                })
            })
        }
        let initGenerationCallback: IntegrationTestInitGeneration | null = null
        const initGeneration = (logic: IntegrationTestInitGeneration): void => {
            initGenerationCallback = logic
        }
        suite({ initGeneration, test: test([description]), describe: describe([description]) })
        before(async () => {
            if (!initGenerationCallback) {
                throw new Error('initGeneration() was never called')
            }
            const setupResult = await initGenerationCallback()
            driver = setupResult.driver
            sourcegraphBaseUrl = setupResult.sourcegraphBaseUrl
            if (setupResult.subscriptions) {
                subscriptions.add(setupResult.subscriptions)
            }
        })
        after(async () => {
            await driver.close()
            subscriptions.unsubscribe()
        })
    })
}

/**
 * Runs an integration test suite using Puppeteer and Mocha.
 *
 * Static CSS/JS assets will be served from the `ui/assets` directory.
 *
 * Other requests (for instance to the GraphQL API) will be stubbed using response
 * stubs from the `__fixtures__` directory, through Puppeteer's request interception.
 */
function runIntegrationTests(description: string, suite: IntegrationTestSuite): void {
    const sourcegraphBaseUrl = 'http://localhost:8000'
    mochaDescribe(description, () => {
        let driver: Driver
        let app: express.Express
        let stopServer: () => void
        mochaBefore(async () => {
            driver = await createDriverForTest({
                sourcegraphBaseUrl,
            })
            // Serve static assets from `ui/assets`
            app = express()
            app.use('/.assets', express.static(ASSETS_DIRECTORY))
            const server = app.listen(8000)
            stopServer = () => server.close()
        })
        mochaAfter(async () => {
            stopServer()
            await driver.close()
        })
        const test = (prefixes: string[]): IntegrationTest => (testName, run) => {
            mochaTest(testName, async () => {
                const fixture = path.join(FIXTURES_DIRECTORY, ...prefixes.map(snakeCase), `${snakeCase(testName)}.json`)
                const fixtureExists = await exists(fixture)
                if (!fixtureExists) {
                    throw new Error(`no fixture file for test ${JSON.stringify(testName)}`)
                }
                const requestStubs: RequestStubs = JSON.parse((await readFile(fixture)).toString())
                await driver.newPage()

                // Intercept requests and provide stubbed responses
                await driver.page.setRequestInterception(true)
                // eslint-disable-next-line @typescript-eslint/no-misused-promises
                driver.page.on('request', async request => {
                    const requestURL = new URL(request.url())
                    if (requestURL.pathname.startsWith('/.assets')) {
                        await request.continue()
                        return
                    }
                    const requestKey = [requestURL.pathname, requestURL.search, requestURL.hash, request.postData()]
                        .filter(isDefined)
                        .map(s => s.replace(/\s+/g, ''))
                        .join('-')
                    const requestStub = requestStubs[requestKey]
                    if (requestStub) {
                        await request.respond(requestStub)
                    }
                })
                await run({ sourcegraphBaseUrl, driver })
                await driver.page.close()
            })
        }
        const describe = (prefixes: string[]): IntegrationTestDescribe => (description, suite) => {
            mochaDescribe(description, () => {
                suite({
                    describe: describe([...prefixes, description]),
                    before: noop,
                    test: test([...prefixes, description]),
                })
            })
        }
        suite({ initGeneration: noop, test: test([description]), describe: describe([description]) })
    })
}

/**
 *
 */
export function describeIntegration(description: string, testSuite: IntegrationTestSuite): void {
    if (readEnvBoolean({ variable: 'GENERATE' })) {
        generateIntegrationTestsData(description, testSuite)
    } else {
        runIntegrationTests(description, testSuite)
    }
}
