import { cleanup, render } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'

import { TemporarySettings } from '@sourcegraph/shared/src/settings/temporary/TemporarySettings'
import { MockTemporarySettings } from '@sourcegraph/shared/src/settings/temporary/testUtils'

import { IdeExtensionTracker } from './IdeExtensionTracker'

describe('IdeExtensionTracker', () => {
    afterAll(cleanup)

    const cases: [string, null | 'vscode' | 'jetbrains'][] = [
        [
            'https://sourcegraph.com/-/editor?remote_url=git%40github.com%3Asourcegraph%2Fsourcegraph-jetbrains.git&branch=main&file=src%2Fmain%2Fjava%2FOpenRevisionAction.java&editor=JetBrains&version=v1.2.2&start_row=68&start_col=26&end_row=68&end_col=26&utm_product_name=IntelliJ+IDEA&utm_product_version=2021.3.2',
            'jetbrains',
        ],
        [
            'https://sourcegraph.com/-/editor?remote_url=git@github.com:sourcegraph/sourcegraph.git&branch=ps/detect-ide-extensions&file=client/web/src/tracking/util.ts&editor=VSCode&version=2.0.9&start_row=13&start_col=22&end_row=13&end_col=22&utm_campaign=vscode-extension&utm_medium=direct_traffic&utm_source=vscode-extension&utm_content=vsce-commands',
            'vscode',
        ],
        [
            'https://sourcegraph.com/sign-up?editor=vscode&utm_medium=VSCODE&utm_source=sidebar&utm_campaign=vsce-sign-up&utm_content=sign-up',
            'vscode',
        ],
        ['https://sourcegraph.com/?something=different', null],
    ]
    test.each(cases)('Detects the proper extension for %p', (url, expectedResult) => {
        let latestSettings: TemporarySettings = {}
        const onSettingsChanged = (nextSettings: TemporarySettings) => (latestSettings = nextSettings)

        render(
            <MemoryRouter initialEntries={[url]}>
                <MockTemporarySettings settings={{}} onSettingsChanged={onSettingsChanged}>
                    <IdeExtensionTracker />
                </MockTemporarySettings>
            </MemoryRouter>
        )

        if (expectedResult === 'vscode') {
            expect(latestSettings).toHaveProperty(['integrations.vscode.lastDetectionTimestamp'])
        } else {
            expect(latestSettings).not.toHaveProperty(['integrations.vscode.lastDetectionTimestamp'])
        }

        if (expectedResult === 'jetbrains') {
            expect(latestSettings).toHaveProperty(['integrations.jetbrains.lastDetectionTimestamp'])
        } else {
            expect(latestSettings).not.toHaveProperty(['integrations.jetbrains.lastDetectionTimestamp'])
        }
    })
})
