import { useEffect, useRef, useState } from 'react'

import { useTabsContext as useReachTabsContext } from '@reach/tabs'

import { useTablePanelIndex, useTabsSettings } from './context'

export function useShouldPanelRender(children: React.ReactNode): boolean {
    const { selectedIndex } = useReachTabsContext()
    const index = useTablePanelIndex()
    const { lazy, behavior } = useTabsSettings()
    const [wasRendered, setWasRendered] = useState(selectedIndex === index)
    const previousChildren = useRef(children)

    useEffect(() => {
        if (lazy) {
            // If children change, we should invalidate "cache" and unrender
            if (children !== previousChildren.current) {
                setWasRendered(index === selectedIndex)
                previousChildren.current = children
            } else if (index === selectedIndex) {
                setWasRendered(true)
            }
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
