import { mkdir, mkdtempSync, rmdirSync, writeFile } from 'fs'
import { tmpdir } from 'os'
import * as path from 'path'

import { Frame, Page, test as base } from '@playwright/test'
import { downloadAndUnzipVSCode } from '@vscode/test-electron'
import { _electron as electron } from 'playwright'

import { run } from '../fixtures/mock-server'

export const test = base
    .extend<{}>({
        page: async ({ page: _page }, use, testInfo) => {
            void _page

            const codyRoot = path.resolve(__dirname, '..', '..')

            const vscodeExecutablePath = await downloadAndUnzipVSCode('1.78.2')
            const extensionDevelopmentPath = codyRoot

            const userDataDirectory = mkdtempSync(path.join(tmpdir(), 'cody-vsce'))
            const extensionsDirectory = mkdtempSync(path.join(tmpdir(), 'cody-vsce'))
            const videoDirectory = path.join(codyRoot, '..', '..', 'playwright', escapeToPath(testInfo.title))

            const workspaceDirectory = path.join(codyRoot, 'test', 'fixtures', 'workspace')

            await buildWorkSpaceSettings(workspaceDirectory)

            // See: https://github.com/microsoft/vscode-test/blob/main/lib/runTest.ts
            const app = await electron.launch({
                executablePath: vscodeExecutablePath,
                env: {
                    ...process.env,
                    CODY_TESTING: 'true',
                },
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
                // Record a video that can be picked up by Buildkite. Since there is no way right
                recordVideo: {
                    dir: videoDirectory,
                },
            })

            await waitUntil(() => app.windows().length > 0)

            const page = await app.firstWindow()

            // Bring the cody sidebar to the foreground
            await page.click('[aria-label="Sourcegraph Cody"]')
            // Ensure that we remove the hover from the activity icon
            await page.getByRole('heading', { name: 'Sourcegraph Cody: Chat' }).hover()
            // Wait for Cody to become activated
            // TODO(philipp-spiess): Figure out which playwright matcher we can use that works for
            // the signed-in and signed-out cases
            await new Promise(resolve => setTimeout(resolve, 500))

            const sidebar = await getCodySidebar(page)

            await run(async () => {
                // Ensure we're logged out
                // TODO(philipp-spiess): Find a way to access the extension host via the injected
                // electron process so we can run the hidden command instead.
                if (await page.isVisible('[aria-label="Settings"]')) {
                    await page.click('[aria-label="Settings"]')
                    await sidebar.getByRole('button', { name: 'Logout' }).click()
                }

                await use(page)
            })

            await app.close()

            // Delete the recorded video if the test passes
            if (testInfo.status === 'passed') {
                rmdirSync(videoDirectory, { recursive: true })
            }

            rmdirSync(userDataDirectory, { recursive: true })
            rmdirSync(extensionsDirectory, { recursive: true })
        },
    })
    .extend<{ sidebar: Frame }>({
        sidebar: async ({ page }, use) => {
            const sidebar = await getCodySidebar(page)
            await use(sidebar)
        },
    })

export async function getCodySidebar(page: Page): Promise<Frame> {
    async function findCodySidebarFrame(): Promise<null | Frame> {
        for (const frame of page.frames()) {
            try {
                const title = await frame.title()
                if (title === 'Cody') {
                    return frame
                }
            } catch (error: any) {
                // Skip over frames that were detached in the meantime.
                if (error.message.indexOf('Frame was detached') === -1) {
                    throw error
                }
            }
        }
        return null
    }
    await waitUntil(async () => (await findCodySidebarFrame()) !== null)
    return (await findCodySidebarFrame()) || page.mainFrame()
}

async function waitUntil(predicate: () => boolean | Promise<boolean>): Promise<void> {
    let delay = 10
    while (!(await predicate())) {
        await new Promise(resolve => setTimeout(resolve, delay))
        delay <<= 1
    }
}

function escapeToPath(text: string): string {
    return text.replace(/\W/g, '_')
}

// Build a workspace settings file that enables the experimental inline mode
export async function buildWorkSpaceSettings(workspaceDirectory: string): Promise<void> {
    const settings = {
        'cody.experimental.inline': true,
        'cody.serverEndpoint': 'http://localhost:49300',
    }
    // create a temporary directory with settings.json and add to the workspaceDirectory
    const workspaceSettingsPath = path.join(workspaceDirectory, '.vscode', 'settings.json')
    const workspaceSettingsDirectory = path.join(workspaceDirectory, '.vscode')
    mkdir(workspaceSettingsDirectory, { recursive: true }, () => {})
    await new Promise<void>((resolve, reject) => {
        writeFile(workspaceSettingsPath, JSON.stringify(settings), error => {
            if (error) {
                reject(error)
            } else {
                resolve()
            }
        })
    })
}
