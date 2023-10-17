import { writeFile, mkdir } from 'mz/fs'
import pTimeout from 'p-timeout'
import type { Browser, WebWorker } from 'puppeteer'
import * as uuid from 'uuid'

import { logger } from '@sourcegraph/common'

import type { Driver } from './driver'

declare global {
    interface FileCoverage {
        /** Absolute path. */
        path: string
        hash: string
        // fnMap, branchMap, statementMap, s, f, b, _coverageSchema
    }

    // eslint-disable-next-line no-var
    var __coverage__: Record<string, FileCoverage> | undefined
}

let warnedNoCoverage = false

/**
 * Saves coverage recorded by the instrumented code in `.nyc_output` after each test.
 */
export function afterEachRecordCoverage(getDriver: () => Driver): void {
    afterEach(() => recordCoverage(getDriver().browser))
}

/**
 * Saves coverage recorded by the instrumented code in `.nyc_output`.
 */
export async function recordCoverage(browser: Browser): Promise<void> {
    await mkdir('.nyc_output', { recursive: true })
    // Get pages, web workers, background pages, etc.
    const targets = browser.targets()

    await Promise.all(
        targets.map(async target => {
            if (target.url() === 'about:blank') {
                return
            }
            const executionContext = (await target.worker()) ?? (await target.page())
            if (!executionContext) {
                return
            }
            const coverage: typeof __coverage__ | void = await pTimeout(
                (executionContext as WebWorker).evaluate(() => globalThis.__coverage__),
                2000,
                () => {
                    if (!warnedNoCoverage) {
                        logger.error(
                            `No coverage found in target ${target.url()}\n` +
                                'Run the dev Sourcegraph instance with COVERAGE_INSTRUMENT=true to track coverage.'
                        )
                        warnedNoCoverage = true
                    }
                }
            )

            if (coverage) {
                await writeFile(`.nyc_output/${uuid.v4()}.json`, JSON.stringify(coverage), { flag: 'wx' })
            }
        })
    )
}
