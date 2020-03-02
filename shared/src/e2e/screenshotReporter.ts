import mkdirp from 'mkdirp-promise'
import * as path from 'path'
import * as puppeteer from 'puppeteer'
import { afterEach } from 'mocha'

/**
 * Registers an `afterEach` hook (for use with Mocha) that takes a screenshot of
 * the browser when a test fails. It is used by e2e tests.
 */
export function saveScreenshotsUponFailures(getPage: () => puppeteer.Page): void {
    afterEach(async function() {
        if (this.currentTest && this.currentTest.state === 'failed') {
            await takeScreenshot({
                page: getPage(),
                repoRootDir: path.resolve(__dirname, '..', '..', '..'),
                screenshotDir: path.resolve(__dirname, '..', '..', '..', 'puppeteer'),
                testName: this.currentTest.fullTitle(),
            })
        }
    })
}

async function takeScreenshot({
    page,
    repoRootDir,
    screenshotDir,
    testName,
}: {
    page: puppeteer.Page
    repoRootDir: string
    screenshotDir: string
    testName: string
}): Promise<void> {
    await mkdirp(screenshotDir)
    const filePath = path.join(screenshotDir, testName.replace(/\W/g, '_') + '.png')
    await page.screenshot({ path: filePath })
    if (process.env.CI) {
        // Print image with ANSI escape code for Buildkite: https://buildkite.com/docs/builds/images-in-log-output.
        console.log(`\u001B]1338;url="artifact://${path.relative(repoRootDir, filePath)}";alt="Screenshot"\u0007`)
    } else {
        console.log(`📸  Saved screenshot of failure to ${path.relative(process.cwd(), filePath)}`)
    }
}
