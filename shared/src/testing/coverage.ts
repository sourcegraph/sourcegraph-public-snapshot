import { Driver } from './driver'
import { writeFile, mkdir } from 'mz/fs'
import { Browser } from 'puppeteer'
import * as uuid from 'uuid'

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
    afterEach('Record coverage', () => recordCoverage(getDriver().browser))
}

/**
 * Saves coverage recorded by the instrumented code in `.nyc_output`.
 */
export async function recordCoverage(browser: Browser): Promise<void> {
    await mkdir('.nyc_output', { recursive: true })
    // Get pages, web workers, background pages, etc.
    const targets = await browser.targets()
    await Promise.all(
        targets.map(async target => {
            const executionContext = (await target.worker()) ?? (await target.page())
            if (!executionContext) {
                return
            }
            const coverage: typeof __coverage__ = await executionContext.evaluate(() => globalThis.__coverage__)
            if (!coverage) {
                if (!warnedNoCoverage && target.url() !== 'about:blank') {
                    console.error(
                        `No coverage found in target ${target.url()}\n` +
                            'Run the dev Sourcegraph instance with COVERAGE_INSTRUMENT=true to track coverage.'
                    )
                    warnedNoCoverage = true
                }
                return
            }
            await Promise.all(
                Object.values(coverage).map(async fileCoverage => {
                    await writeFile(
                        `.nyc_output/${uuid.v4()}.json`,
                        JSON.stringify({ [fileCoverage.path]: fileCoverage }),
                        { flag: 'wx' }
                    )
                })
            )
        })
    )
}
