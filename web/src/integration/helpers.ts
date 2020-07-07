import { describe as mochaDescribe, test as mochaTest, before as mochaBefore } from 'mocha'
import { Subscription } from 'rxjs'
import { snakeCase } from 'lodash'
import { createDriverForTest, Driver } from '../../../shared/src/testing/driver'
import * as path from 'path'
import mkdirp from 'mkdirp-promise'
import express from 'express'
import { Polly } from '@pollyjs/core'
import { PuppeteerAdapter } from './polly/PuppeteerAdapter'
import FSPersister from '@pollyjs/persister-fs'

Polly.register(PuppeteerAdapter as any)
Polly.register(FSPersister)

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

/**
 * Describes an integration test suite using wrappers over Mocha primitives.
 *
 * To record test data, set the RECORD environment variable to a truthy value.
 * When recording, the tests will be run in Puppeteer against a real backend,
 * and captured response fixtures will be saved to the `__fixtures__` directory.
 *
 * When running the tests, static CSS/JS assets will be served from the `ui/assets` directory.
 * Other requests (for instance to the GraphQL API) will be stubbed using response
 * stubs from the `__fixtures__` directory, through Puppeteer's request interception.
 *
 */
export function describeIntegration(description: string, testSuite: IntegrationTestSuite): void {
    const record = Boolean(process.env.RECORD)
    mochaDescribe(description, () => {
        let driver: Driver
        let sourcegraphBaseUrl: string
        const subscriptions = new Subscription()
        const test = (prefixes: string[]): IntegrationTest => (testName, run) => {
            mochaTest(testName, async () => {
                await driver.newPage()
                await driver.page.setRequestInterception(true)
                const recordingsDirectory = path.join(FIXTURES_DIRECTORY, ...prefixes.map(snakeCase))
                if (record) {
                    await mkdirp(recordingsDirectory)
                }
                const polly = new Polly(snakeCase(testName), {
                    adapters: ['puppeteer'],
                    adapterOptions: {
                        puppeteer: { page: driver.page, requestResourceTypes: ['xhr', 'fetch', 'document'] },
                    },
                    persister: 'fs',
                    persisterOptions: {
                        fs: {
                            recordingsDir: recordingsDirectory,
                        },
                    },
                    expiryStrategy: 'warn',
                    recordIfMissing: record,
                    matchRequestsBy: {
                        method: true,
                        body: true,
                        order: true,
                        // Origin header will change when running against a test instance
                        headers: false,
                        url: {
                            pathname: true,
                            query: true,
                            hash: true,
                            // Allow recording tests against https://sourcegraph.test
                            // but running them against http:://localhost:8000
                            protocol: false,
                            port: false,
                            hostname: false,
                            username: false,
                            password: false,
                        },
                    },
                    mode: record ? 'record' : 'replay',
                    logging: false,
                })
                const { server } = polly
                server.get('/.assets/*path').passthrough()
                // Filter out 'server' header filled in by Caddy before persisting responses,
                // otherwise tests will hang when replayed from recordings.
                server
                    .any()
                    .on(
                        'beforePersist',
                        (request, recording: { response: { headers: { name: string; value: string }[] } }) => {
                            recording.response.headers = recording.response.headers.filter(
                                ({ name }) => name !== 'server'
                            )
                        }
                    )
                await run({ sourcegraphBaseUrl, driver })
                await polly.stop()
                await driver.page.close()
            })
        }
        const before: IntegrationTestBeforeGeneration = setupLogic => {
            if (record) {
                mochaBefore(() => setupLogic({ driver, sourcegraphBaseUrl }))
            }
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
        testSuite({ initGeneration, test: test([description]), describe: describe([description]) })
        mochaBefore(async () => {
            if (!initGenerationCallback) {
                throw new Error('initGeneration() was never called')
            }
            if (record) {
                const setupResult = await initGenerationCallback()
                driver = setupResult.driver
                sourcegraphBaseUrl = setupResult.sourcegraphBaseUrl
                if (setupResult.subscriptions) {
                    subscriptions.add(setupResult.subscriptions)
                }
            } else {
                sourcegraphBaseUrl = 'http://localhost:8000'
                driver = await createDriverForTest({
                    sourcegraphBaseUrl,
                    logBrowserConsole: false,
                })
                // Serve static assets from `ui/assets`
                const app = express()
                app.use('/.assets', express.static(ASSETS_DIRECTORY))
                const server = app.listen(8000)
                subscriptions.add(() => server.close())
            }
        })
        after(async () => {
            await driver.close()
            subscriptions.unsubscribe()
        })
    })
}
