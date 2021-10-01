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
import classNames from 'classnames'
import React from 'react'

import { TabsContext, useTabsContext } from './context'
import styles from './Tabs.module.scss'
import { useTabPanelBehavior } from './useTabPanelBehavior'
import { useTabPanelsState } from './useTabPanelsState'
import { TabsState, useTabs } from './useTabs'

interface TabsProps extends ReachTabsProps, TabsState {
    className?: string
}

interface TabListProps extends ReachTabListProps {
    /*
     * action is used to render content in the left side of
     * the component. e.g. a close button or a list of links.
     */
    actions?: React.ReactNode
}

interface TabProps extends ReachTabProps {}
interface TabPanelsProps extends ReachTabPanelsProps {}
interface TabPanelProps extends ReachTabPanelProps {}

export type { TabsProps, TabPanelsProps, TabPanelProps }

/**
 *
 * reach UI tabs component with steroids, this tabs handles how the data should be loaded
 * in terms of a11y tabs are following all the WAI-ARIA Tabs Design Pattern.
 *
 * See: https://reach.tech/tabs/
 *
 */
export const Tabs: React.FunctionComponent<TabsProps> = props => {
    const { lazy, size, behavior, className, ...reachProps } = props
    const { contextValue } = useTabs({ lazy, size, behavior })

    return (
        <TabsContext.Provider value={contextValue}>
            <div className={classNames(styles.wildcardTabs, className)} data-testid="wildcard-tabs">
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
    const styleSize = styles[size]

    return (
        <ReachTab className={styleSize} data-testid="wildcard-tab" {...props}>
            <span className={styles.tabLabel}>{props.children}</span>
        </ReachTab>
    )
}

export const TabPanels: React.FunctionComponent<TabPanelsProps> = ({ children }) => {
    const { show, element } = useTabPanelsState(children)
    return show ? <ReachTabPanels data-testid="wildcard-tab-panels">{element}</ReachTabPanels> : null
}

export const TabPanel: React.FunctionComponent = ({ children }) => {
    const { isMounted } = useTabPanelBehavior()
    return <ReachTabPanel data-testid="wildcard-tab-panel">{isMounted ? children : null}</ReachTabPanel>
}
