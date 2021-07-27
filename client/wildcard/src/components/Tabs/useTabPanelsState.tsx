import React, { MutableRefObject, useEffect, useMemo, useRef } from 'react'

import { useTabsContext, TabsIndexContext } from './context'
import { Tabs } from './reducer'

interface UseTabPanelsState {
    show: boolean
    element: React.ReactNodeArray | null | undefined
}

export const useTabPanelsState = (children: React.ReactNode): UseTabPanelsState => {
    const { dispatch } = useTabsContext()
    const tabPanelArray = React.Children.toArray(children)
    const renderTogo = useRef(false)

    const tabCollection: MutableRefObject<() => Tabs> = useRef(() =>
        tabPanelArray.reduce((accumulator: Tabs, _current, index) => {
            accumulator[index] = {
                index,
                mounted: false,
            }
            return accumulator
        }, {})
    )

    useEffect(() => {
        dispatch({ type: 'SET_TABS', payload: { tabs: tabCollection.current() } })
        renderTogo.current = true
    }, [dispatch])

    const element = useMemo(
        () =>
            React.Children.map(children, (panel, index) => (
                <TabsIndexContext.Provider value={index}>{panel}</TabsIndexContext.Provider>
            )),
        [children]
    )

    return { show: renderTogo.current, element }
}
