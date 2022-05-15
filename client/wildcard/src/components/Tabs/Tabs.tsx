import React from 'react'

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
import classNames from 'classnames'

import { ForwardReferenceComponent } from '../../types'

import { TabPanelIndexContext, TabsSettingsContext, useTabsSettings } from './context'
import { useShouldPanelRender } from './useShouldPanelRender'

import styles from './Tabs.module.scss'

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

    /**
     * Controls the behavior in case tab list doesn't fit into a container.
     * Default is "wrap"
     */
    longTabList?: 'wrap' | 'scroll'
}

export interface TabsProps extends ReachTabsProps, TabsSettings {
    className?: string
}

export interface TabListProps extends ReachTabListProps {
    wrapperClassName?: string
    /*
     * action is used to render content in the left side of
     * the component. e.g. a close button or a list of links.
     */
    actions?: React.ReactNode
}

export interface TabProps extends ReachTabProps {}

export interface TabPanelsProps extends ReachTabPanelsProps {}

export interface TabPanelProps extends ReachTabPanelProps {}

/**
 * reach UI tabs component with steroids, this tabs handles how the data should be loaded
 * in terms of a11y tabs are following all the WAI-ARIA Tabs Design Pattern.
 *
 * See: https://reach.tech/tabs/
 */
export const Tabs = React.forwardRef((props, reference) => {
    const {
        lazy = false,
        size = 'small',
        behavior = 'forceRender',
        className,
        as = 'div',
        longTabList = 'wrap',
        ...reachProps
    } = props
    return (
        <TabsSettingsContext.Provider value={{ lazy, size, behavior, longTabList }}>
            <ReachTabs
                className={classNames(styles.wildcardTabs, className)}
                data-testid="wildcard-tabs"
                ref={reference}
                as={as}
                {...reachProps}
            />
        </TabsSettingsContext.Provider>
    )
}) as ForwardReferenceComponent<'div', TabsProps>

export const TabList = React.forwardRef((props, reference) => {
    const { actions, as = 'div', wrapperClassName, className, ...reachProps } = props
    const { longTabList } = useTabsSettings()

    return (
        <div className={classNames(styles.tablistWrapper, wrapperClassName)}>
            <ReachTabList
                data-testid="wildcard-tab-list"
                as={as}
                ref={reference}
                className={classNames(className, styles.tabList, longTabList === 'scroll' && styles.tabListScroll)}
                {...reachProps}
            />
            {actions}
        </div>
    )
}) as ForwardReferenceComponent<'div', TabListProps>

export const Tab = React.forwardRef((props, reference) => {
    const { as = 'button', ...reachProps } = props
    const { size } = useTabsSettings()
    const { longTabList } = useTabsSettings()

    return (
        <ReachTab
            className={classNames(styles[size], longTabList === 'scroll' && styles.tabNowrap)}
            data-testid="wildcard-tab"
            as={as}
            ref={reference}
            {...reachProps}
        >
            <span className={styles.tabLabel}>{props.children}</span>
        </ReachTab>
    )
}) as ForwardReferenceComponent<'button', TabProps>

export const TabPanels = React.forwardRef((props, reference) => {
    const { as = 'div', ...reachProps } = props
    return (
        <ReachTabPanels data-testid="wildcard-tab-panel-list" as={as} ref={reference} {...reachProps}>
            {React.Children.map(props.children, (child, index) => (
                <TabPanelIndexContext.Provider value={index}>{child}</TabPanelIndexContext.Provider>
            ))}
        </ReachTabPanels>
    )
}) as ForwardReferenceComponent<'div', TabPanelsProps>

export const TabPanel = React.forwardRef((props, reference) => {
    const { as = 'div', children, ...reachProps } = props
    const shouldRender = useShouldPanelRender(children)
    return (
        <ReachTabPanel data-testid="wildcard-tab-panel" as={as} ref={reference} {...reachProps}>
            {shouldRender ? children : null}
        </ReachTabPanel>
    )
}) as ForwardReferenceComponent<'div', TabPanelProps>
