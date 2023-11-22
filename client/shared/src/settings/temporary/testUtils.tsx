import React from 'react'

import { gql } from '@apollo/client'
import { createMockClient } from '@apollo/client/testing'

import type { TemporarySettings } from './TemporarySettings'
import { TemporarySettingsContext } from './TemporarySettingsProvider'
import { InMemoryMockSettingsBackend, TemporarySettingsStorage } from './TemporarySettingsStorage'

export const MockTemporarySettings: React.FunctionComponent<
    React.PropsWithChildren<{
        settings: TemporarySettings
        onSettingsChanged?: (settings: TemporarySettings) => void
    }>
> = ({ settings, onSettingsChanged, children }) => {
    const mockClient = createMockClient(
        null,
        gql`
            query TemporarySettings {
                temporarySettings {
                    contents
                }
            }
        `
    )

    const settingsBackend = new InMemoryMockSettingsBackend(settings, onSettingsChanged)
    const settingsStorage = new TemporarySettingsStorage(mockClient, false)
    settingsStorage.setSettingsBackend(settingsBackend)

    return <TemporarySettingsContext.Provider value={settingsStorage}>{children}</TemporarySettingsContext.Provider>
}
