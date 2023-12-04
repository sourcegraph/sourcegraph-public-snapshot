import React from 'react'

import { gql } from '@apollo/client'
import { createMockClient } from '@apollo/client/testing'
import type { Meta } from '@storybook/react'

import { TemporarySettingsContext } from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsProvider'
import {
    InMemoryMockSettingsBackend,
    TemporarySettingsStorage,
} from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsStorage'
import { H2 } from '@sourcegraph/wildcard'

import { WebStory } from '../../../components/WebStory'
import { CodeInsightsBackendContext, CodeInsightsGqlBackend } from '../core'
import type { DashboardPermissions } from '../pages/dashboards/dashboard-view/utils/get-dashboard-permissions'

import { GaConfirmationModal } from './GaConfirmationModal'

const settingsClient = createMockClient(
    { contents: JSON.stringify({}) },
    gql`
        query TemporarySettings {
            temporarySettings {
                contents
            }
        }
    `
)

class CodeInsightExampleBackend extends CodeInsightsGqlBackend {
    public getUiFeatures = () => {
        const permissions: DashboardPermissions = { isConfigurable: true }
        return { licensed: false, permissions }
    }
}
const api = new CodeInsightExampleBackend({} as any)

const Story: Meta = {
    title: 'web/insights/GaConfirmationModal',
    decorators: [story => <WebStory>{() => <div className="p-3 container web-content">{story()}</div>}</WebStory>],
}

export default Story

export const GaConfirmationModalExample: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => {
    const settingsStorage = new TemporarySettingsStorage(settingsClient, true)

    settingsStorage.setSettingsBackend(new InMemoryMockSettingsBackend({}))

    return (
        <CodeInsightsBackendContext.Provider value={api}>
            <TemporarySettingsContext.Provider value={settingsStorage}>
                <div>
                    <H2>Some content</H2>
                    <GaConfirmationModal />
                </div>
            </TemporarySettingsContext.Provider>
        </CodeInsightsBackendContext.Provider>
    )
}
