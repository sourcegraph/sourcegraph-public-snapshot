import React from 'react'

import type { TabsSettings } from '.'

export interface TabsState {
    settings: Required<TabsSettings>
    activeIndex: number
}

export const TabsStateContext = React.createContext<TabsState | null>(null)
TabsStateContext.displayName = 'TabsStateContext'

export const useTabsState = (): TabsState => {
    const context = React.useContext(TabsStateContext)
    if (!context) {
        throw new Error('useTabsState or Tabs inner components cannot be used outside <Tabs> sub-tree')
    }
    return context
}

export const TabPanelIndexContext = React.createContext<number>(0)
TabPanelIndexContext.displayName = 'TabPanelIndexContext'

export const useTablePanelIndex = (): number => {
    const context = React.useContext(TabPanelIndexContext)
    if (context === undefined) {
        throw new Error('TabPanelIndexContext cannot be used outside <Tabs> sub-tree')
    }
    return context
}
