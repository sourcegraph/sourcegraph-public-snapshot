import createPersistedState from 'use-persisted-state'

import { GoogleAuthSettings } from '../services/google/GoogleSignIn'

export interface Settings {
    googleAuth?: GoogleAuthSettings
}

export interface SettingsProps {
    settings: Settings
    setSettings: React.Dispatch<React.SetStateAction<Settings>>
}

const useSettingsState = createPersistedState<Settings>('settings')

export const useSettings = (): [SettingsProps['settings'], SettingsProps['setSettings']] => useSettingsState({})
