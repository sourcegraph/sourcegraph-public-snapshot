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

import { TabPanelIndexContext, TabsSettingsContext, useTabsSettings } from './context'
import styles from './Tabs.module.scss'
import { useShouldPanelRender } from './useShouldPanelRender'

export interface TabsSettings {
    /**
     * Tab component font size.
     * Default is "small"
     */
    size?: 'small' | 'medium' | 'large'
    /**
     * true: only load the initial tab when tab component mounts
     * false: render all the TabPanel children when tab component mounts
     */
    lazy?: boolean
    /**
     * This prop is lazy dependant, only should be used when lazy is true
     * memoize: Once a selected tabPanel is rendered this will keep mounted
     * forceRender: Each time a tab is selected the associated tabPanel is mounted
     * and the rest is unmounted
     */
    behavior?: 'memoize' | 'forceRender'
}

export interface TabsProps extends ReachTabsProps, TabsSettings {
    className?: string
}

export interface TabListProps extends ReachTabListProps {
    /*
     * action is used to render content in the left side of
     * the component. e.g. a close button or a list of links.
     */
    actions?: React.ReactNode
}

export interface TabProps extends ReachTabProps {}
export interface TabPanelsProps extends ReachTabPanelsProps {}
export interface TabPanelProps extends ReachTabPanelProps, React.HTMLAttributes<HTMLDivElement> {}

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

    return (
        <TabsSettingsContext.Provider value={{ lazy, size, behavior }}>
            <ReachTabs
                className={classNames(styles.wildcardTabs, className)}
                data-testid="wildcard-tabs"
                {...reachProps}
            />
        </TabsSettingsContext.Provider>
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
    const { size = 'small' } = useTabsSettings()
    return (
        <ReachTab className={styles[size]} data-testid="wildcard-tab" {...props}>
            <span className={styles.tabLabel}>{props.children}</span>
        </ReachTab>
    )
}

export const TabPanels: React.FunctionComponent<TabPanelsProps> = ({ children }) => (
    <ReachTabPanels data-testid="wildcard-tab-panel-list">
        {React.Children.map(children, (child, index) => (
            <TabPanelIndexContext.Provider value={index}>{child}</TabPanelIndexContext.Provider>
        ))}
    </ReachTabPanels>
)

export const TabPanel: React.FunctionComponent<TabPanelProps> = ({ children, ...otherProps }) => {
    const shouldRender = useShouldPanelRender(children)

    return shouldRender ? (
        <ReachTabPanel data-testid="wildcard-tab-panel" {...otherProps}>
            {children}
        </ReachTabPanel>
    ) : null
}
