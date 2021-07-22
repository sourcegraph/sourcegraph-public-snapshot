import React, { useContext } from 'react'
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
    useTabsContext,
} from '@reach/tabs'

interface TabsProps extends ReachTabsProps {}

interface TabListProps extends ReachTabListProps {}

interface TabProps extends ReachTabProps {}

interface TabPanelsProps extends ReachTabPanelsProps {}
interface TabPanelProps extends ReachTabPanelProps {
    forceRender?: boolean
}

const TabsIndexContext = React.createContext(0)

export type { TabsProps, TabPanelsProps, TabPanelProps }

export const Tabs: React.FunctionComponent<TabsProps> = props => {
    return <ReachTabs data-testid="wildcard-tabs" {...props} />
}

export const TabList: React.FunctionComponent<TabListProps> = props => {
    return <ReachTabList data-testid="wildcard-tab-list" {...props} />
}

export const Tab: React.FunctionComponent<TabProps> = props => {
    return <ReachTab data-testid="wildcard-tab" {...props} />
}

export const TabPanels: React.FunctionComponent<TabPanelsProps> = ({ children }) => {
    const element = React.Children.map(children, (panel, index) => {
        return <TabsIndexContext.Provider value={index}>{panel}</TabsIndexContext.Provider>
    })

    return <ReachTabPanels data-testid="wildcard-tab-panels">{element}</ReachTabPanels>
}

export const TabPanel: React.FunctionComponent<TabPanelProps> = ({ forceRender, children }) => {
    const index = useContext(TabsIndexContext)
    const { selectedIndex } = useTabsContext()
    let element = null

    if (forceRender) {
        if (selectedIndex === index) {
            element = children
        } else {
            element = null
        }
    } else {
        element = children
    }

    return <ReachTabPanel data-testid="wildcard-tab-panel">{element}</ReachTabPanel>
}
