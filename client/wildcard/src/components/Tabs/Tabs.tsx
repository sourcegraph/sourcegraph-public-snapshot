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

import { TabsContext } from './context'
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
            <ReachTabs data-testid="wildcard-tabs" {...reachProps} />
        </TabsContext.Provider>
    )
}

export const TabList: React.FunctionComponent<TabListProps> = props => (
    <div>
        <ReachTabList data-testid="wildcard-tab-list" {...props} />
        {props.actions}
    </div>
)

export const Tab: React.FunctionComponent<TabProps> = props => <ReachTab data-testid="wildcard-tab" {...props} />

export const TabPanels: React.FunctionComponent<TabPanelsProps> = ({ children }) => {
    const { show, element } = useTabPanelsState(children)
    return show ? <ReachTabPanels data-testid="wildcard-tab-panels">{element}</ReachTabPanels> : null
}

export const TabPanel: React.FunctionComponent = ({ children }) => {
    const isMounted = useTabPanelBehavior()
    return <ReachTabPanel data-testid="wildcard-tab-panel">{isMounted ? children : null}</ReachTabPanel>
}
