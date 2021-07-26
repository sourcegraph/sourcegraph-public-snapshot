import React, { useRef, useReducer, MutableRefObject, useMemo, useEffect } from 'react'
import {
    Tab as ReachTab,
    TabList as ReachTabList,
    TabListProps as ReachTabListProps,
    TabPanel as ReachTabPanel,
    TabPanelProps as ReachTabPanelProps,
    TabPanels as ReachTabPanels,
    TabPanelsProps as ReachTabPanelsProps,
    TabProps as ReachTabProps,
    Tabs as ReachTabs,
    TabsProps as ReachTabsProps,
    useTabsContext as useReachTabsContext,
} from '@reach/tabs'

import { reducer, Tabs as TabsInterface } from './reducer'
import { useTabsContext, TabsContext, TabsIndexContext, useTabsIndexContext } from './context'
interface TabsProps extends ReachTabsProps {
    size?: 'small' | 'default' | 'large'
    lazy?: boolean
    behavior?: 'memoize' | 'forceRender'
}

interface TabListProps extends ReachTabListProps {}

interface TabProps extends ReachTabProps {}

interface TabPanelsProps extends ReachTabPanelsProps {}
interface TabPanelProps extends ReachTabPanelProps {
    forceRender?: boolean
}

export type { TabsProps, TabPanelsProps, TabPanelProps }

// const traverse = (node, visitor) => {
//     return _traverse(node, visitor, { level: 0, parent: null, index: 0 })
// }

// const _traverse = (node, visitor, state) => {
//     console.log(state)
//     visitor(node, state)

//     if (!node.props) return

//     const children = React.Children.toArray(node.props.children)

//     children.forEach((child, i) =>
//         _traverse(child, visitor, {
//             level: state.level + 1,
//             parent: node,
//             index: i,
//         })
//     )
// }

export const Tabs: React.FunctionComponent<TabsProps> = props => {
    const { lazy, size, behavior } = props
    // How know the amount of tab components in the compound component
    // a traverse inspection?

    const [state, dispatch] = useReducer(reducer, { lazy, size, behavior, current: 1, tabs: {} })
    const contextValue = useMemo(() => ({ state, dispatch }), [state, dispatch])
    return (
        <TabsContext.Provider value={contextValue}>
            <ReachTabs data-testid="wildcard-tabs" {...props} />
        </TabsContext.Provider>
    )
}

export const TabList: React.FunctionComponent<TabListProps> = props => {
    return <ReachTabList data-testid="wildcard-tab-list" {...props} />
}

export const Tab: React.FunctionComponent<TabProps> = props => {
    return <ReachTab data-testid="wildcard-tab" {...props} />
}

export const TabPanels: React.FunctionComponent<TabPanelsProps> = ({ children }) => {
    const { dispatch } = useTabsContext()
    const tabPanelArray = React.Children.toArray(children)
    const renderTogo = useRef(false)

    const tabCollection: MutableRefObject<() => TabsInterface> = useRef(() =>
        tabPanelArray.reduce((accumulator: TabsInterface, _current, index) => {
            accumulator[index] = {
                index: index,
                mounted: false,
            }
            return accumulator
        }, {})
    )

    useEffect(() => {
        dispatch({ type: 'SET_TABS', payload: { tabs: tabCollection.current() } })
        renderTogo.current = true
    }, [dispatch, tabCollection])

    console.log('tab collection', tabCollection.current())
    const element = useMemo(
        () =>
            React.Children.map(children, (panel, index) => {
                return <TabsIndexContext.Provider value={index}>{panel}</TabsIndexContext.Provider>
            }),
        [children]
    )

    return renderTogo.current ? <ReachTabPanels data-testid="wildcard-tab-panels">{element}</ReachTabPanels> : null
}

export const TabPanel: React.FunctionComponent<TabPanelProps> = ({ forceRender, children }) => {
    const index = useTabsIndexContext()
    const { selectedIndex } = useReachTabsContext()
    const { state, dispatch } = useTabsContext()

    console.log('entra')

    const setMountedTab = (mounted: boolean, index: number) =>
        useMemo(() => dispatch({ type: 'MOUNTED_TAB', payload: { index, mounted } }), [dispatch, index, mounted])

    const { lazy, behavior, tabs } = state
    const currentTab = tabs[index]

    let element = null

    if (lazy) {
        if (behavior === 'forceRender') {
            if (selectedIndex === currentTab.index) {
                element = children
            } else {
                element = null
            }
        }

        if (behavior === 'memoize') {
            // current tab render
            if (selectedIndex === currentTab.index || currentTab.mounted) {
                setMountedTab(true, index)
                element = children
                // not current tab which is not mounted avoid render
            } else if (!currentTab.mounted) {
                setMountedTab(false, index)
                element = null
            }
        }
    } else {
        element = children
    }

    return <ReachTabPanel data-testid="wildcard-tab-panel">{element}</ReachTabPanel>
}
