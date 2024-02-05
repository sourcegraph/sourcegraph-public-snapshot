import { useLayoutEffect, useState, type ReactNode } from 'react'

import { useTabsContext as useReachTabsContext } from '@reach/tabs'

import { useTablePanelIndex, useTabsState } from './context'

type FunctionLikeChildren = ReactNode | ((context: any) => ReactNode)

export function useShouldPanelRender(children: FunctionLikeChildren): boolean {
    const { selectedIndex } = useReachTabsContext()
    const index = useTablePanelIndex()
    const {
        settings: { lazy, behavior },
    } = useTabsState()
    const [wasRendered, setWasRendered] = useState(selectedIndex === index)

    useLayoutEffect(() => {
        if (lazy && index === selectedIndex) {
            setWasRendered(true)
        }
    }, [lazy, children, index, selectedIndex])

    if (lazy) {
        if (behavior === 'forceRender') {
            return selectedIndex === index
        }
        if (behavior === 'memoize') {
            return wasRendered
        }
    }

    return true
}
