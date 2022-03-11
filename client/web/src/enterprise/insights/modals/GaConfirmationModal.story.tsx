import { gql } from '@apollo/client'
import { createMockClient } from '@apollo/client/testing'
import { Meta } from '@storybook/react'
import React from 'react'

import { TemporarySettingsContext } from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsProvider'
import {
    InMemoryMockSettingsBackend,
    TemporarySettingsStorage,
} from '@sourcegraph/shared/src/settings/temporary/TemporarySettingsStorage'

import { WebStory } from '../../../components/WebStory'
import { CodeInsightsBackendContext } from '../core/backend/code-insights-backend-context'
import { CodeInsightsGqlBackend } from '../core/backend/gql-api/code-insights-gql-backend'

import { GaConfirmationModal } from './GaConfirmationModal'

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

class CodeInsightExampleBackend extends CodeInsightsGqlBackend {
    public getUiFeatures = () => ({ licensed: false })
}
const api = new CodeInsightExampleBackend({} as any)

const Story: Meta = {
    title: 'web/insights/GaConfirmationModal',
    decorators: [story => <WebStory>{() => <div className="p-3 container web-content">{story()}</div>}</WebStory>],
}

export default Story

export const GaConfirmationModalExample: React.FunctionComponent = () => {
    const settingsStorage = new TemporarySettingsStorage(settingsClient, true)

    settingsStorage.setSettingsBackend(new InMemoryMockSettingsBackend({}))

    return (
        <CodeInsightsBackendContext.Provider value={api}>
            <TemporarySettingsContext.Provider value={settingsStorage}>
                <div>
                    <h2>Some content</h2>
                    <GaConfirmationModal />
                </div>
            </TemporarySettingsContext.Provider>
        </CodeInsightsBackendContext.Provider>
    )
}
