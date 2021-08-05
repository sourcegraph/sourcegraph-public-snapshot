import { useTabsContext as useReachTabsContext } from '@reach/tabs'
import { useEffect, useMemo } from 'react'

import { useTabsContext, useTabsIndexContext } from './context'

export const useTabPanelBehavior = (): { isMounted: boolean } => {
    const index = useTabsIndexContext()
    const { selectedIndex } = useReachTabsContext()
    const { state, dispatch } = useTabsContext()

    const { setMountedTab } = useMemo(
        () => ({
            setMountedTab: (mounted: boolean, index: number) =>
                dispatch({ type: 'MOUNTED_TAB', payload: { index, mounted } }),
        }),
        [dispatch]
    )

    const { lazy, behavior, tabs } = state
    const currentTab = tabs[index]

    useEffect(() => {
        if (lazy) {
            if (behavior === 'forceRender') {
                if (selectedIndex === currentTab.index) {
                    setMountedTab(true, currentTab.index)
                } else {
                    setMountedTab(false, currentTab.index)
                }
            }

            if (behavior === 'memoize') {
                // current tab render
                if (selectedIndex === currentTab.index || currentTab.mounted) {
                    setMountedTab(true, currentTab.index)
                    // not current tab which is not mounted avoid render
                } else if (!currentTab.mounted) {
                    setMountedTab(false, currentTab.index)
                }
            }
        } else {
            setMountedTab(true, currentTab.index)
        }
    }, [behavior, currentTab.index, currentTab.mounted, lazy, selectedIndex, setMountedTab])

    return { isMounted: currentTab.mounted }
}
