import { downloadAndUnzipVSCode } from '@vscode/test-electron'

import { mixedSearchStreamEvents, highlightFileResult } from '@sourcegraph/search'
import { Settings } from '@sourcegraph/shared/src/settings/settings'
import { setupExtensionMocking } from '@sourcegraph/shared/src/testing/integration/mockExtension'

import { createVSCodeIntegrationTestContext, VSCodeIntegrationTestContext } from './context'
import { getVSCodeWebviewFrames } from './getWebview'
import { launchVsCode, VSCodeTestDriver } from './launch'

const sourcegraphBaseUrl = 'https://sourcegraph.com'

describe('VS Code extension', () => {
    let vsCodeDriver: VSCodeTestDriver
    before(async () => {
        const vscodeExecutablePath = await downloadAndUnzipVSCode()
        vsCodeDriver = await launchVsCode(vscodeExecutablePath)
    })
    after(() => vsCodeDriver?.dispose())

    let testContext: VSCodeIntegrationTestContext

    beforeEach(async function () {
        testContext = await createVSCodeIntegrationTestContext(
            {
                currentTest: this.currentTest!,
                directory: __dirname,
            },
            vsCodeDriver.page
        )
    })

    // Debt: reset VS Code extension state between test cases in `afterEach` once we
    // have multiple tests. This will likely involve just closing the search panel.
    // afterEach(async () => {})

    it('works', async () => {
        const { Extensions } = setupExtensionMocking({
            pollyServer: testContext.server,
            sourcegraphBaseUrl,
        })

        const userSettings: Settings = {
            extensions: {},
        }

        testContext.overrideGraphQL({
            Extensions,

            ...highlightFileResult,
            ViewerSettings: () => ({
                viewerSettings: {
                    __typename: 'SettingsCascade',
                    final: JSON.stringify(userSettings),
                    subjects: [
                        {
                            __typename: 'User',
                            displayName: 'Test User',
                            id: 'TestUserSettingsID',
                            latestSettings: {
                                id: 123,
                                contents: JSON.stringify(userSettings),
                            },
                            username: 'test',
                            viewerCanAdminister: true,
                            settingsURL: '/users/test/settings',
                        },
                    ],
                },
            }),
        })

        testContext.overrideSearchStreamEvents([...mixedSearchStreamEvents])

        const { searchPanelFrame } = await getVSCodeWebviewFrames(vsCodeDriver.page)

        // Focus search box
        await searchPanelFrame.waitForSelector('.monaco-editor .view-lines')
        await searchPanelFrame.click('.monaco-editor .view-lines')

        await vsCodeDriver.page.keyboard.type('test', { delay: 50 })
        // Submit search
        await searchPanelFrame.waitForSelector('.test-search-button', { visible: true })
        await searchPanelFrame.click('.test-search-button')

        try {
            await searchPanelFrame.waitForSelector('.test-search-result', { visible: true })
        } catch {
            throw new Error('Timeout waiting for search results to render')
        }
    })

    // Potential future test cases:
    // - Clicking search result opens remote files
    // - Sourcegraph extensions work on remote files
    // - Clicking sidebar filter updates query (and executes for some?)
})
