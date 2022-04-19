/* eslint-disable @typescript-eslint/no-unsafe-member-access */
/* eslint-disable @typescript-eslint/no-unsafe-assignment */
/* eslint-disable @typescript-eslint/explicit-function-return-type */
import childProcess from 'child_process'
import { mkdtempSync, readFileSync } from 'fs'
import { tmpdir } from 'os'
import { join } from 'path'

import { downloadAndUnzipVSCode } from '@vscode/test-electron'
import puppeteer, { Page } from 'puppeteer'
import rimraf from 'rimraf'

import { installExtension } from './installExtension'

const verbose = process.argv.includes('-v') || process.argv.includes('--verbose')

const PORT = 29378

async function run(): Promise<void> {
    let vscodeProcess: null | { kill: () => void } = null

    function cleanupVSCode(runAfter?: () => void): void {
        setTimeout(() => {
            if (vscodeProcess !== null) {
                vscodeProcess.kill()
            }

            // eslint-disable-next-line no-void
            void delay(1000).then(() => {
                rimraf.sync(userDataDirectory)
                rimraf.sync(extensionsDirectory)

                runAfter?.()
            })
        }, 1000)
    }

    const userDataDirectory = mkdtempSync(join(tmpdir(), 'vsce'))
    const extensionsDirectory = mkdtempSync(join(tmpdir(), 'vsce'))
    try {
        const vscodeExecutablePath = await downloadAndUnzipVSCode()

        console.log('Starting VS Code', { verbose, vscodeExecutablePath, userDataDirectory, extensionsDirectory })

        const extensionVersion: string = JSON.parse(readFileSync('package.json').toString()).version
        if (typeof extensionVersion !== 'string' || extensionVersion === '') {
            throw new Error('Could not extract extension version from client/vscode/package.json')
        }

        const extensionPackage = join(process.cwd(), 'dist', `sourcegraph-${extensionVersion}.vsix`)

        await installExtension(extensionPackage, extensionsDirectory)

        vscodeProcess = launchVSC(vscodeExecutablePath, userDataDirectory, extensionsDirectory, PORT)

        const browserURL = `http://127.0.0.1:${PORT}`
        await delay(2000)
        console.log(`VSCode started in debug mode on ${browserURL}`)

        const browser = await puppeteer.connect({
            browserURL,
            defaultViewport: null, // used to bypass Chrome viewport issue, doesn't work w/ VS code.
            slowMo: 50,
        })

        await delay(1000)

        // We look for the VSCode frontend process
        const pages = await browser.pages()
        let page: null | Page = null
        for (const _page of pages) {
            const title = await _page.title()
            if (title.includes('Get Started')) {
                page = _page
            }
        }
        if (page === null) {
            throw new Error('Could not find VS Code frontend page')
        }

        const frame = await getSearchPanelWebview(page)

        const context = await frame.executionContext()
        const textContent: string = await context.evaluate(() => document.body.textContent)

        if (!textContent.includes('Search your code and 2M+ open source repositories')) {
            throw new Error('Expected page content to contain a specific string')
        }

        console.log('+++ Test successful')
        cleanupVSCode()
    } catch (error) {
        console.error('--- Failed to run tests')
        console.error(error)
        cleanupVSCode(() => process.exit(1))
    }
}

// eslint-disable-next-line no-void
void run()

function launchVSC(executablePath: string, userDataDirectory: string, extensionsDirectory: string, port: number) {
    return childProcess.spawn(
        executablePath,
        [
            `--remote-debugging-port=${port || 9229}`,
            `--user-data-dir=${userDataDirectory}`,
            `--extensions-dir=${extensionsDirectory}`,
            '--enable-logging',
            '--skip-release-notes',
            // https://github.com/microsoft/vscode-test/issues/120
            '--disable-updates',
            // https://github.com/microsoft/vscode/issues/84238
            '--no-sandbox',
            '--disable-gpu',
        ],
        {
            env: process.env,
            stdio: verbose ? 'inherit' : ['pipe', 'pipe', 'pipe'],
        }
    )
}

function delay(timeout: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, timeout))
}

/**
 * The VSCode extension currently opens two web views, one is the sidebar content and the other the search page.
 * We want to run assertions on the search page.
 * To be reused for all search panel test cases.
 *
 * @param page The VS Code frontend page.
 * @returns Search panel webview frame.
 */
async function getSearchPanelWebview(page: puppeteer.Page): Promise<puppeteer.Frame> {
    await page.waitForSelector('[aria-label="Sourcegraph"]')
    await page.click('[aria-label="Sourcegraph"]')

    // In the release of VS Code at the time this test harness was written, there was no
    // stable unique selector for the search panel webview's outermost iframe ancestor.
    // Try all iframes, from last to first. Fail when we can't find the search panel webview.
    const outerFrameHandles = (await page?.$$('div[id^="webview"] iframe')).reverse()

    let searchPanelWebviewFrame: puppeteer.Frame | null = null

    // Throw error for the latest step in which a failure occurred.
    const errorMessages: string[] = []

    if (outerFrameHandles.length === 0) {
        errorMessages[0] = 'Could not find Sourcegraph search page iframe handle'
    }

    for (const outerFrameHandle of outerFrameHandles) {
        const outerFrame = await outerFrameHandle.contentFrame()
        if (!outerFrame) {
            errorMessages[1] = 'Could not find Sourcegraph search page iframe'
            continue
        }
        // The search page web view has another iframe inside it. ¯\_(ツ)_/¯
        const frameHandle = await outerFrame.waitForSelector('iframe')
        if (frameHandle === null) {
            errorMessages[2] = 'Could not find inner Sourcegraph search page iframe handle'
            continue
        }
        const frame = await frameHandle.contentFrame()
        if (frame === null) {
            errorMessages[3] = 'Could not find inner Sourcegraph search page iframe'
            continue
        }

        try {
            const brandHeader = await frame.waitForSelector('[data-testid="brand-header"]')
            if (!brandHeader) {
                throw new Error('Expected search panel to render brand header')
            }
        } catch {
            errorMessages[4] = 'Expected search panel to render brand header'
            continue
        }

        searchPanelWebviewFrame = frame
        break
    }

    if (!searchPanelWebviewFrame) {
        throw new Error(errorMessages.slice(-1)[0])
    }

    return searchPanelWebviewFrame
}
