import childProcess from 'child_process'
import { mkdtempSync, createReadStream, createWriteStream } from 'fs'
import { tmpdir } from 'os'
import { join } from 'path'
import zlib from 'zlib'

import { downloadAndUnzipVSCode } from '@vscode/test-electron'
import puppeteer from 'puppeteer'
import request from 'request-promise-native'

function spawn(executablePath: string, userDataDirectory: string, extensionsDirectory: string, port: number): any {
    return childProcess.spawn(
        executablePath,
        [
            `--remote-debugging-port=${port || 9229}`,
            // '--disable-extension=sourcegraph.sourcegraph',
            `--user-data-dir=${userDataDirectory}`, // arbitrary not-mine datadir to get the welcome screen
            `--extensions-dir=${extensionsDirectory}`,
            // '--extensionDevelopmentPath=${path.join(defaultCachePath, 'extensions')}`)
            '--enable-logging',
        ],
        {
            detached: true,
            env: process.env,
            stdio: ['pipe', 'pipe', 'pipe'],
        }
    )
}

function installExtension(extensionDirectory: string, extension: string) {
    const unzip = zlib.createInflate()

    const read = createReadStream(extension)
    const write = createWriteStream(join(extensionDirectory, 'sourcegraph.sourcegraph-2.2.0'))

    read.pipe(unzip).pipe(write)
    console.log('unZipped Successfully')
}

function delay(timeout: number): Promise<void> {
    return new Promise(resolve => setTimeout(resolve, timeout))
}

async function run(): Promise<void> {
    try {
        const vscodeExecutablePath = await downloadAndUnzipVSCode()

        const userDataDirectory = mkdtempSync(join(tmpdir(), 'vsce'))
        const extensionsDirectory = mkdtempSync(join(tmpdir(), 'vsce'))

        installExtension(extensionsDirectory, '/Users/philipp/dev/sourcegraph/client/vscode/sourcegraph-2.2.0.vsix')

        console.log({ vscodeExecutablePath, userDataDirectory, extensionsDirectory })

        const port = 29378

        const proc = spawn(vscodeExecutablePath, userDataDirectory, extensionsDirectory, port)

        await delay(2000)

        const response = await request(`http://127.0.0.1:${port}/json/list`)
        const developmentToolsPages = JSON.parse(response)
        console.log({ developmentToolsPages })
        const endpoint = developmentToolsPages.find((p: any) => !p.title.match(/^sharedProcess/))

        const browser = await puppeteer.connect({
            browserWSEndpoint: endpoint.webSocketDebuggerUrl,
            defaultViewport: null, // used to bypass Chrome viewport issue, doesn't work w/ VS code.
            slowMo: 50,
        })

        await delay(1000)

        const page = (await browser.pages())[0]

        console.log(page.title)

        await page.click('[href="command:workbench.action.files.newUntitledFile"]')

        await page.type('.monaco-editor', 'Woo! I am automating Visual Studio Code with puppeteer!\n')
        await page.type('.monaco-editor', 'This would be a super cool way of generating foolproof demos.')

        setTimeout(() => proc.kill(), 1000)

        // const extensionDevelopmentPath = path.resolve(__dirname, '../')
        // const extensionTestsPath = path.resolve(__dirname, './e2e/index.js')
        // const workspace = path.resolve(__dirname, '../test-fixtures/workspace')

        // await runTests({
        //     extensionDevelopmentPath,
        //     extensionTestsPath,
        //     launchArgs: [workspace],
        // })
    } catch {
        console.error('Failed to run tests')
        process.exit(1)
    }
}

// eslint-disable-next-line no-void
void run()
