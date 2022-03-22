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
    let vscodeProccess: null | { kill: () => void } = null

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

        vscodeProccess = launchVSC(vscodeExecutablePath, userDataDirectory, extensionsDirectory, PORT)

        const browserURL = `http://127.0.0.1:${PORT}`
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

        await page.waitForSelector('[aria-label="Sourcegraph"]')
        await page.click('[aria-label="Sourcegraph"]')

        // The VSCode extension currently opens two web views, one is the sidebar content and the other the search page.
        // We want to run assertions on the search page.
        const outerFrameHandle = await page.waitForSelector(
            'iframe:not([name~="webviewview-sourcegraph-searchsidebar"])'
        )
        if (outerFrameHandle === null) {
            throw new Error('Could not find Sourcegraph search page iframe handle')
        }
        const outerFrame = await outerFrameHandle.contentFrame()
        if (outerFrame === null) {
            throw new Error('Could not find Sourcegraph search page iframe')
        }

        // The search page web view has another iframe inside it. ¯\_(ツ)_/¯
        const frameHandle = await outerFrame.waitForSelector('iframe')
        if (frameHandle === null) {
            throw new Error('Could not find inner Sourcegraph search page iframe handle')
        }
        const frame = await frameHandle.contentFrame()
        if (frame === null) {
            throw new Error('Could not find inner Sourcegraph search page iframe')
        }

        const context = await frame.executionContext()
        const textContent: string = await context.evaluate(() => document.body.textContent)

        if (!textContent.includes('Search your code and 2M+ open source repositories')) {
            throw new Error('Expected page content to contain a specific string')
        }

        console.log('+++ Test successful')
    } catch (error) {
        console.error('--- Failed to run tests')
        console.error(error)
        process.exit(1)
    } finally {
        setTimeout(() => {
            if (vscodeProccess !== null) {
                vscodeProccess.kill()
            }

            // eslint-disable-next-line no-void
            void delay(1000).then(() => {
                rimraf.sync(userDataDirectory)
                rimraf.sync(extensionsDirectory)
            })
        }, 1000)
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
