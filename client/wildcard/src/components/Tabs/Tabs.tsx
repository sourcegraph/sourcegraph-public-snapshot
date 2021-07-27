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
} from '@reach/tabs'
import React from 'react'

import { TabsContext, useTabsContext } from './context'
import styles from './Tabs.module.scss'
import { useTabPanelBehavior } from './useTabPanelBehavior'
import { useTabPanelsState } from './useTabPanelsState'
import { TabsApi, useTabs } from './useTabs'

interface TabsProps extends ReachTabsProps, TabsApi {}

interface TabListProps extends ReachTabListProps {
    actions?: React.ReactNode
}

interface TabProps extends ReachTabProps {}

interface TabPanelsProps extends ReachTabPanelsProps {}
interface TabPanelProps extends ReachTabPanelProps {
    forceRender?: boolean
}

export type { TabsProps, TabPanelsProps, TabPanelProps }

export const Tabs: React.FunctionComponent<TabsProps> = props => {
    const { lazy, size, behavior, ...reachProps } = props
    const { contextValue } = useTabs({ lazy, size, behavior })

    return (
        <TabsContext.Provider value={contextValue}>
            <div className={styles.wildcardTabs} data-testid="wildcard-tabs">
                <ReachTabs {...reachProps} />
            </div>
        </TabsContext.Provider>
    )
}

export const TabList: React.FunctionComponent<TabListProps> = props => {
    const { actions, ...reachProps } = props
    return (
        <div className={styles.tablistWrapper}>
            <ReachTabList data-testid="wildcard-tab-list" {...reachProps} />
            {actions}
        </div>
    )
}

export const Tab: React.FunctionComponent<TabProps> = props => {
    const { state } = useTabsContext()

    const { size = 'small' } = state
    const styleSize = styles[size] as keyof typeof styles

    return <ReachTab className={styleSize} data-testid="wildcard-tab" {...props} />
}

export const TabPanels: React.FunctionComponent<TabPanelsProps> = ({ children }) => {
    const { show, element } = useTabPanelsState(children)
    return show ? <ReachTabPanels data-testid="wildcard-tab-panels">{element}</ReachTabPanels> : null
}

export const TabPanel: React.FunctionComponent = ({ children }) => {
    const isMounted = useTabPanelBehavior()
    return <ReachTabPanel data-testid="wildcard-tab-panel">{isMounted ? children : null}</ReachTabPanel>
}
