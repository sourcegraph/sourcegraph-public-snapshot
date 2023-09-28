import { startTransition } from 'react'

import create from 'zustand'
import { persist } from 'zustand/middleware'

import { countOverrides } from '../devsettings/utils'

interface DeveloperSettingsState {
    showDialog: boolean
    selectedTab: number
    featureFlags: {
        view: string
        filter: string
    }
    temporarySettings: {
        view: string
        filter: string
    }
}

export const useDeveloperSettings = create<DeveloperSettingsState>(
    persist<DeveloperSettingsState>(
        () => ({
            showDialog: false,
            selectedTab: 0,
            featureFlags: {
                view: 'All',
                filter: '',
            },
            temporarySettings: {
                view: 'All',
                filter: '',
            },
        }),
        {
            name: 'developerSettingsDialog',
        }
    )
)

export function setDeveloperSettingsFeatureFlags(settings: Partial<DeveloperSettingsState['featureFlags']>): void {
    useDeveloperSettings.setState(state => ({
        featureFlags: {
            ...state.featureFlags,
            ...settings,
        },
    }))
}

export function setDeveloperSettingsTemporarySettings(
    settings: Partial<DeveloperSettingsState['temporarySettings']>
): void {
    useDeveloperSettings.setState(state => ({
        temporarySettings: {
            ...state.temporarySettings,
            ...settings,
        },
    }))
}

export function toggleDevSettingsDialog(show?: boolean): void {
    // startTransition is needed because the dialog is/should be
    // lazy loaded. Without it an error is thrown.
    startTransition(() => {
        useDeveloperSettings.setState(state => ({
            showDialog: show ?? !state.showDialog,
        }))
    })
}

export const useOverrideCounter = create<{ featureFlags: number; temporarySettings: number }>(() => countOverrides())

export function updateOverrideCounter(): void {
    useOverrideCounter.setState(countOverrides())
}
