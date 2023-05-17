import { mkdtempSync } from 'fs'
import { tmpdir } from 'os'
import * as path from 'path'

import { test as base } from '@playwright/test'
import { downloadAndUnzipVSCode } from '@vscode/test-electron'
import { _electron as electron } from 'playwright'

export const test = base.extend<{}>({
    page: async ({}, use) => {
        const codyRoot = path.resolve(__dirname, '..', '..')

        const vscodeExecutablePath = await downloadAndUnzipVSCode()
        const extensionDevelopmentPath = codyRoot

        const userDataDirectory = mkdtempSync(path.join(tmpdir(), 'cody-vsce'))
        const extensionsDirectory = mkdtempSync(path.join(tmpdir(), 'cody-vsce'))

        const app = await electron.launch({
            executablePath: vscodeExecutablePath,
            args: [
                // https://github.com/microsoft/vscode/issues/84238
                '--no-sandbox',
                // https://github.com/microsoft/vscode-test/issues/120
                '--disable-updates',
                '--skip-welcome',
                '--skip-release-notes',
                '--disable-workspace-trust',
                '--extensionDevelopmentPath=' + extensionDevelopmentPath,
                `--user-data-dir=${userDataDirectory}`,
                `--extensions-dir=${extensionsDirectory}`,
            ],
        })

        await waitUntil(() => app.windows().length > 0)

        const page = await app.firstWindow()
        await use(page)
    },
})

async function waitUntil(predicate: () => boolean | Promise<boolean>): Promise<void> {
    let delay = 10
    while (!(await predicate())) {
        await new Promise(resolve => setTimeout(resolve, delay))
        delay <<= 1
    }
}
