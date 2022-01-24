import React from 'react'

import { TabsSettings } from '.'

export const TabsSettingsContext = React.createContext<TabsSettings | null>(null)
TabsSettingsContext.displayName = 'TabsSettingsContext'

export const useTabsSettings = (): TabsSettings => {
    const context = React.useContext(TabsSettingsContext)
    if (!context) {
        throw new Error('useTabsSettingsContext or Tabs inner components cannot be used outside <Tabs>')
    }

    return context
}

export const TabPanelIndexContext = React.createContext<number>(0)
TabPanelIndexContext.displayName = 'TabPanelIndexContext'

export const useTablePanelIndex = (): number => {
    const context = React.useContext(TabPanelIndexContext)
    if (context === undefined) {
        throw new Error('TabsIndexContext can not be used outside the <Tabs> component scope')
    }
    return context
}
