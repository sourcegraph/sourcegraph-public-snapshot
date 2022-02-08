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
import { As, PropsWithAs } from '@reach/utils'
import classNames from 'classnames'
import React from 'react'

import { TabPanelIndexContext, TabsSettingsContext, useTabsSettings } from './context'
import styles from './Tabs.module.scss'
import { useShouldPanelRender } from './useShouldPanelRender'

export { useTabsContext }

export interface TabsSettings {
    /**
     * Tab component font size.
     * Default is "small"
     */
    size?: 'small' | 'medium' | 'large'
    /**
     * true: only load the initial tab when tab component mounts
     * false: render all the TabPanel children when tab component mounts
     * Default is false
     */
    lazy?: boolean
    /**
     * This prop is lazy dependant, only should be used when lazy is true
     * memoize: Once a selected tabPanel is rendered this will keep mounted
     * forceRender: Each time a tab is selected the associated tabPanel is mounted
     * and the rest is unmounted.
     * Default is "forceRender"
     */
    behavior?: 'memoize' | 'forceRender'
}

export interface TabsProps extends PropsWithAs<As, ReachTabsProps & TabsSettings> {
    className?: string
}

export interface TabListProps extends PropsWithAs<As, ReachTabListProps>, React.HTMLAttributes<HTMLDivElement> {
    /*
     * action is used to render content in the left side of
     * the component. e.g. a close button or a list of links.
     */
    actions?: React.ReactNode
}

export interface TabProps extends PropsWithAs<As, ReachTabProps>, React.HTMLAttributes<HTMLDivElement> {}
export interface TabPanelsProps extends PropsWithAs<As, ReachTabPanelsProps>, React.HTMLAttributes<HTMLDivElement> {}
export interface TabPanelProps extends PropsWithAs<As, ReachTabPanelProps>, React.HTMLAttributes<HTMLDivElement> {}

/**
 * reach UI tabs component with steroids, this tabs handles how the data should be loaded
 * in terms of a11y tabs are following all the WAI-ARIA Tabs Design Pattern.
 *
 * See: https://reach.tech/tabs/
 */
export const Tabs: React.FunctionComponent<TabsProps> = React.forwardRef((props, reference) => {
    const { lazy = false, size, behavior = 'forceRender', className, as = 'div', ...reachProps } = props
    return (
        <TabsSettingsContext.Provider value={{ lazy, size, behavior }}>
            <ReachTabs
                className={classNames(styles.wildcardTabs, className)}
                data-testid="wildcard-tabs"
                ref={reference}
                as={as}
                {...reachProps}
            />
        </TabsSettingsContext.Provider>
    )
})

export const TabList: React.FunctionComponent<TabListProps> = React.forwardRef((props, reference) => {
    const { actions, as = 'div', ...reachProps } = props
    return (
        <div className={styles.tablistWrapper}>
            <ReachTabList data-testid="wildcard-tab-list" as={as} ref={reference} {...reachProps} />
            {actions}
        </div>
    )
})

export const Tab: React.FunctionComponent<TabProps> = React.forwardRef((props, reference) => {
    const { as = 'button', ...reachProps } = props
    const { size = 'small' } = useTabsSettings()
    return (
        <ReachTab className={styles[size]} data-testid="wildcard-tab" as={as} ref={reference} {...reachProps}>
            <span className={styles.tabLabel}>{props.children}</span>
        </ReachTab>
    )
})

export const TabPanels: React.FunctionComponent<TabPanelsProps> = React.forwardRef((props, reference) => {
    const { as = 'div', ...reachProps } = props
    return (
        <ReachTabPanels data-testid="wildcard-tab-panel-list" as={as} ref={reference} {...reachProps}>
            {React.Children.map(props.children, (child, index) => (
                <TabPanelIndexContext.Provider value={index}>{child}</TabPanelIndexContext.Provider>
            ))}
        </ReachTabPanels>
    )
})

export const TabPanel: React.FunctionComponent<TabPanelProps> = React.forwardRef((props, reference) => {
    const { as = 'div', children, ...reachProps } = props
    const shouldRender = useShouldPanelRender(children)
    return (
        <ReachTabPanel data-testid="wildcard-tab-panel" as={as} ref={reference} {...reachProps}>
            {shouldRender ? children : null}
        </ReachTabPanel>
    )
})
