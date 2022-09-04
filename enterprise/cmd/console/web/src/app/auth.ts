import React from 'react'

import { SettingsProps } from './useSettings'

export interface AuthProvider {
    name: string
    signInComponent: React.ComponentType<SettingsProps>
}
