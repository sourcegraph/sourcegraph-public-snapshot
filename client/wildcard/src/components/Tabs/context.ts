import React from 'react'

import { State, Action } from './reducer'

export interface TabsContext {
    state: State
    dispatch: React.Dispatch<Action>
}

export const TabsContext = React.createContext<TabsContext | null>(null)
TabsContext.displayName = 'TabsContext'

export const TabsIndexContext = React.createContext<number>(0)
TabsIndexContext.displayName = 'TabsIndexContext'

export const useTabsContext = (): TabsContext => {
    const context = React.useContext(TabsContext)
    if (!context) {
        throw new Error('Tabs compound components can not be rendered outside the <Tabs> component')
    }

    return context
}

export const useTabsIndexContext = (): number => {
    const context = React.useContext(TabsIndexContext)
    if (context === undefined) {
        throw new Error('TabsIndexContext can not be used outside the <Tabs> component scope')
    }
    return context
}
