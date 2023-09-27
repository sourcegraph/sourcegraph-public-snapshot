import { countOverrides } from "../devsettings/utils"
import create from "zustand"
import { persist } from 'zustand/middleware'

interface DeveloperSettingsState {
    showDialog: boolean
    selectedTab: number
    selectedView: {
        featureFlags: string
        temporarySettings: string
    }
}

export const useDeveloperSettings = create<DeveloperSettingsState>(persist<DeveloperSettingsState>(
    () => {
        return {
            showDialog: false,
            selectedTab: 0,
            selectedView: {
                featureFlags: 'All',
                temporarySettings: 'All',
            }
        }
    },
    {
        name: 'developerSettingsDialog',
    }
))

export function setDeveloperSettingsFeatureFlagsView(view: string): void {
    useDeveloperSettings.setState(state => ({
        selectedView: {
            ...state.selectedView,
            featureFlags: view
        }
    }))
}

export function setDeveloperSettingsTemporarySettingsView(view: string): void {
    useDeveloperSettings.setState(state => ({
        selectedView: {
            ...state.selectedView,
            temporarySettings: view
        }
    }))
}

export function toggleDevSettingsDialog(show?: boolean): void {
    useDeveloperSettings.setState(state => ({
        showDialog: show ?? !state.showDialog
    }))
}

export const useOverrideCounter = create<{featureFlags: number, temporarySettings: number}>(() => {
    return countOverrides()
})

export function updateOverrideCounter(): void {
    useOverrideCounter.setState(countOverrides())
}
