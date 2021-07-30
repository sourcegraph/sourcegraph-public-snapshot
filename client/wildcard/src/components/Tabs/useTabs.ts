import { useReducer, useMemo } from 'react'

import { TabsContext } from './context'
import { reducer } from './reducer'

export interface TabsState {
    /* Tab component font size */
    size: 'small' | 'medium' | 'large'
    /* true: only load the initial tab when tab component mounts
     * false: render all the TabPanel children when tab component mounts */
    lazy?: boolean
    /**
     * This prop is lazy dependant, only should be used when lazy is true
     * memoize: Once a selected tabPanel is rendered this will keep mounted
     * forceRender: Each time a tab is selected the associated tabPanel is mounted
     * and the rest is unmounted
     */
    behavior?: 'memoize' | 'forceRender'
}

interface UseTabs {
    contextValue: TabsContext
}

export const useTabs = ({ lazy, size, behavior }: TabsState): UseTabs => {
    const [state, dispatch] = useReducer(reducer, { lazy, size, behavior, current: 1, tabs: {} })
    const contextValue = useMemo(() => ({ state, dispatch }), [state, dispatch])

    return { contextValue }
}
