import { mkdtempSync, rmdirSync } from 'fs'
import { tmpdir } from 'os'
import * as path from 'path'

import { Frame, Page, test as base } from '@playwright/test'
import { downloadAndUnzipVSCode } from '@vscode/test-electron'
import { _electron as electron } from 'playwright'

import { run } from '../fixtures/mock-server'

export const test = base.extend<{}>({
    page: async ({ page: _page }, use) => {
        void _page

        const codyRoot = path.resolve(__dirname, '..', '..')

        const vscodeExecutablePath = await downloadAndUnzipVSCode()
        const extensionDevelopmentPath = codyRoot

        const userDataDirectory = mkdtempSync(path.join(tmpdir(), 'cody-vsce'))
        const extensionsDirectory = mkdtempSync(path.join(tmpdir(), 'cody-vsce'))

        const workspaceDirectory = path.join(codyRoot, 'test', 'fixtures', 'workspace')

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
                workspaceDirectory,
            ],
        })

        await waitUntil(() => app.windows().length > 0)

        const page = await app.firstWindow()

        await run(async () => {
            // Bring the cody sidebar to the foreground
            await page.click('[aria-label="Sourcegraph Cody"]')

            await use(page)
        })

        rmdirSync(userDataDirectory, { recursive: true })
        rmdirSync(extensionsDirectory, { recursive: true })
    },
})

export async function getCodySidebar(page: Page): Promise<Frame> {
    async function findCodySidebarFrame(): Promise<null | Frame> {
        for (const frame of page.frames()) {
            const title = await frame.title()
            if (title === 'Cody') {
                return frame
            }
        }
        return null
    }
    await waitUntil(async () => (await findCodySidebarFrame()) !== null)
    return (await findCodySidebarFrame())!
}

async function waitUntil(predicate: () => boolean | Promise<boolean>): Promise<void> {
    let delay = 10
    while (!(await predicate())) {
        await new Promise(resolve => setTimeout(resolve, delay))
        delay <<= 1
    }
}
