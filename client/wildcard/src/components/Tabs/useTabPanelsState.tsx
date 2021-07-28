import React, { MutableRefObject, useEffect, useMemo, useRef } from 'react'

import { useTabsContext, TabsIndexContext } from './context'
import { Tabs } from './reducer'

interface UseTabPanelsState {
    /* Determines if the tab collection ca be rendered */
    show: boolean
    /**
     * A dynamic element with a context provider per child valid for getting
     * single values per child component. In this case, each ancestor will know
     * which index is associated with the current child handling events without
     * declaring props for every child.
     */
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
