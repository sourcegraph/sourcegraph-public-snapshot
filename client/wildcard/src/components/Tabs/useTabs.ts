import { useReducer, useMemo } from 'react'

import { TabsContext } from './context'
import { reducer } from './reducer'

export interface TabsApi {
    size: 'small' | 'medium' | 'large'
    lazy?: boolean
    behavior?: 'memoize' | 'forceRender'
}

interface UseTabs {
    contextValue: TabsContext
}

export const useTabs = ({ lazy, size, behavior }: TabsApi): UseTabs => {
    const [state, dispatch] = useReducer(reducer, { lazy, size, behavior, current: 1, tabs: {} })
    const contextValue = useMemo(() => ({ state, dispatch }), [state, dispatch])

    return { contextValue }
}
