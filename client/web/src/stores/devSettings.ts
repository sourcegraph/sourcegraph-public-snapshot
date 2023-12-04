import { startTransition } from 'react'

import create from 'zustand'
import { persist } from 'zustand/middleware'

import type { AuthenticatedUser } from '../auth'
import { countOverrides } from '../devsettings/utils'

interface DeveloperSettingsState {
    enabled: boolean
    showDialog: boolean
    selectedTab: number
    zoekt: {
        searchOptions: string
    }
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
            enabled: false,
            showDialog: false,
            selectedTab: 0,
            zoekt: {
                searchOptions: '',
            },
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

export function setDeveloperSettingsSearchOptions(settings: Partial<DeveloperSettingsState['zoekt']>): void {
    useDeveloperSettings.setState(state => ({
        zoekt: {
            ...state.zoekt,
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

/**
 * Show or hide the developer settings dialog.
 */
export function toggleDevSettingsDialog(show?: boolean): void {
    // startTransition is needed because the dialog is/should be
    // lazy loaded. Without it an error is thrown.
    startTransition(() => {
        useDeveloperSettings.setState(state => ({
            showDialog: show ?? !state.showDialog,
        }))
    })
}

/**
 * Show or hide the developer settings dialog button in the main navbar.
 */
export function enableDevSettings(enable?: boolean): void {
    useDeveloperSettings.setState(state => ({
        enabled: enable ?? !state.enabled,
    }))
}

export const useOverrideCounter = create<{ featureFlags: number; temporarySettings: number }>(() => countOverrides())

export function updateOverrideCounter(): void {
    useOverrideCounter.setState(countOverrides())
}

export function isSourcegraphDev(authenticatedUser: Pick<AuthenticatedUser, 'emails'> | null): boolean {
    return (
        authenticatedUser?.emails?.some(email => email.verified && email.email?.endsWith('@sourcegraph.com')) ?? false
    )
}
