import { gql } from '@apollo/client'
import { createMockClient } from '@apollo/client/testing'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { TemporarySettingsContext } from '../../../settings/temporary/TemporarySettingsProvider'
import {
    InMemoryMockSettingsBackend,
    TemporarySettingsStorage,
} from '../../../settings/temporary/TemporarySettingsStorage'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { BetaConfirmationModal, BetaConfirmationModalContent } from './BetaConfirmationModal'

const { add } = storiesOf('web/insights/BetaConfirmationModal', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

const mockClient = createMockClient(
    { contents: JSON.stringify({}) },
    gql`
        query {
            temporarySettings {
                contents
            }
        }
    `
)

add('Code Insights Beta modal UI', () => {
    const settingsStorage = new TemporarySettingsStorage(mockClient, true)

    settingsStorage.setSettingsBackend(new InMemoryMockSettingsBackend({}))

    return (
        <EnterpriseWebStory>
            {() => (
                <TemporarySettingsContext.Provider value={settingsStorage}>
                    <div>
                        <h2>Some content</h2>
                        <BetaConfirmationModal />
                    </div>
                </TemporarySettingsContext.Provider>
            )}
        </EnterpriseWebStory>
    )
})

add('Code Insights modal content', () => (
    <EnterpriseWebStory>{() => <BetaConfirmationModalContent onAccept={noop} onDismiss={noop} />}</EnterpriseWebStory>
))
