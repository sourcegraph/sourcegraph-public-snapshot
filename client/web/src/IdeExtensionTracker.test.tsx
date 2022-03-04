import { gql } from '@apollo/client'
import { createMockClient } from '@apollo/client/testing'
import { cleanup, render } from '@testing-library/react'
import React from 'react'

import { TemporarySettingsContext } from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsProvider'
import {
    InMemoryMockSettingsBackend,
    TemporarySettingsStorage,
} from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsStorage'

import { IdeExtensionTracker } from './IdeExtensionTracker'

const settingsClient = createMockClient(
    { contents: JSON.stringify({}) },
    gql`
        query {
            temporarySettings {
                contents
            }
        }
    `
)

type ExpectedResult = null | 'vscode' | 'jetbrains'

describe('IdeExtensionTracker', () => {
    afterAll(cleanup)

    const cases: [string, ExpectedResult][] = [
        [
            'https://sourcegraph.com/-/editor?remote_url=git%40github.com%3Asourcegraph%2Fsourcegraph-jetbrains.git&branch=main&file=src%2Fmain%2Fjava%2FOpenRevisionAction.java&editor=JetBrains&version=v1.2.2&start_row=68&start_col=26&end_row=68&end_col=26&utm_product_name=IntelliJ+IDEA&utm_product_version=2021.3.2',
            'jetbrains',
        ],
        [
            'https://sourcegraph.com/-/editor?remote_url=git@github.com:sourcegraph/sourcegraph.git&branch=ps/detect-ide-extensions&file=client/web/src/tracking/util.ts&editor=VSCode&version=2.0.9&start_row=13&start_col=22&end_row=13&end_col=22&utm_campaign=vscode-extension&utm_medium=direct_traffic&utm_source=vscode-extension&utm_content=vsce-commands',
            'vscode',
        ],
        [
            'https://sourcegraph.com/sign-up?editor=vscode&utm_medium=VSCIDE&utm_source=sidebar&utm_campaign=vsce-sign-up&utm_content=sign-up',
            'vscode',
        ],
        ['https://sourcegraph.com/?something=different', null],
    ]
    test.each(cases)('Detects the proper extension for %p', async (url, expectedResult) => {
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        delete window.location
        // eslint-disable-next-line @typescript-eslint/ban-ts-comment
        // @ts-ignore
        window.location = new URL(url)

        const settingsStorage = new TemporarySettingsStorage(settingsClient, true)
        const settingsBackend = new InMemoryMockSettingsBackend({})
        settingsStorage.setSettingsBackend(settingsBackend)

        render(
            <TemporarySettingsContext.Provider value={settingsStorage}>
                <IdeExtensionTracker />
            </TemporarySettingsContext.Provider>
        )

        const settings = await settingsBackend.load().toPromise()

        if (expectedResult === 'vscode') {
            expect(settings).toHaveProperty(['integrations.vscode.lastDetectionTimestamp'])
        } else {
            expect(settings).not.toHaveProperty(['integrations.vscode.lastDetectionTimestamp'])
        }

        if (expectedResult === 'jetbrains') {
            expect(settings).toHaveProperty(['integrations.jetbrains.lastDetectionTimestamp'])
        } else {
            expect(settings).not.toHaveProperty(['integrations.jetbrains.lastDetectionTimestamp'])
        }
    })
})
