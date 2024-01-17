import React from 'react'

import {
    Tab as ReachTab,
    TabList as ReachTabList,
    type TabListProps as ReachTabListProps,
    TabPanel as ReachTabPanel,
    type TabPanelProps as ReachTabPanelProps,
    TabPanels as ReachTabPanels,
    type TabPanelsProps as ReachTabPanelsProps,
    type TabProps as ReachTabProps,
    Tabs as ReachTabs,
    type TabsProps as ReachTabsProps,
    useTabsContext,
} from '@reach/tabs'
import classNames from 'classnames'
import { isFunction } from 'lodash'

import { useElementObscuredArea } from '../../hooks'
import type { ForwardReferenceComponent } from '../../types'

import { TabPanelIndexContext, type TabsState, TabsStateContext, useTabsState } from './context'
import { useScrollBackToActive } from './useScrollBackToActive'
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

export interface TabPanelContext {
    shouldRender: boolean
}

export interface TabPanelProps extends Omit<ReachTabPanelProps, 'children'> {
    children?: React.ReactNode | ((tabContext: TabPanelContext) => React.ReactNode)
}

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
        onChange,
        ...reachProps
    } = props

    const [activeIndex, setActiveIndex] = React.useState(props.defaultIndex || 0)
    const tabsStateContext: TabsState = React.useMemo(
        () => ({
            settings: {
                size,
                lazy,
                behavior,
                longTabList,
            },
            activeIndex,
        }),
        [activeIndex, behavior, lazy, longTabList, size]
    )

    const onChangePersistIndex = React.useCallback(
        (index: number) => {
            setActiveIndex(index)
            if (onChange) {
                onChange(index)
            }
        },
        [onChange]
    )

    return (
        <TabsStateContext.Provider value={tabsStateContext}>
            <ReachTabs
                className={classNames(styles.wildcardTabs, className)}
                data-testid="wildcard-tabs"
                ref={reference}
                as={as}
                onChange={onChangePersistIndex}
                {...reachProps}
            />
        </TabsStateContext.Provider>
    )
}) as ForwardReferenceComponent<'div', TabsProps>

export const TabList = React.forwardRef((props, reference) => {
    const {
        settings: { longTabList },
    } = useTabsState()

    if (longTabList === 'scroll') {
        return <TabListScrolled ref={reference} {...props} />
    }

    return <TabListPlain ref={reference} {...props} />
}) as ForwardReferenceComponent<'div', TabListProps>

const TabListScrolled = React.forwardRef((props, passedReference) => {
    const ownReference = React.useRef<HTMLDivElement | null>(null)

    // This is required because ref can be passed as a ref object
    // or callback. We need to support both cases
    const saveAndPassReference = React.useCallback(
        (element: HTMLDivElement) => {
            ownReference.current = element
            if (!passedReference) {
                return
            }
            if ('current' in passedReference) {
                passedReference.current = element
            }
            if (typeof passedReference === 'function') {
                passedReference(element)
            }
        },
        [passedReference]
    )

    const obscuredArea = useElementObscuredArea(ownReference)

    const extraWrapperClasses = [
        obscuredArea.left > 0 ? styles.tablistWrapperObscuredLeft : undefined,
        obscuredArea.right > 0 ? styles.tablistWrapperObscuredRight : undefined,
    ]

    useScrollBackToActive(ownReference)

    return (
        <TabListPlain
            extraClasses={[styles.tabListScroll]}
            extraWrapperClasses={extraWrapperClasses}
            ref={saveAndPassReference}
            {...props}
        />
    )
}) as ForwardReferenceComponent<'div', TabListProps>

const TabListPlain = React.forwardRef((props, reference) => {
    const {
        as,
        actions,
        className,
        wrapperClassName,
        extraClasses = [],
        extraWrapperClasses = [],
        ...restProps
    } = props

    return (
        <div className={classNames(styles.tablistWrapper, wrapperClassName, ...extraWrapperClasses)}>
            <ReachTabList
                data-testid="wildcard-tab-list"
                as={as}
                ref={reference}
                className={classNames(className, styles.tabList, ...extraClasses)}
                {...restProps}
            />
            {actions}
        </div>
    )
}) as ForwardReferenceComponent<
    'div',
    TabListProps & {
        extraClasses?: (string | undefined)[]
        extraWrapperClasses?: (string | undefined)[]
    }
>

export const Tab = React.forwardRef((props, reference) => {
    const { as = 'button', className, ...reachProps } = props
    const {
        settings: { size, longTabList },
    } = useTabsState()

    return (
        <ReachTab
            className={classNames(styles[size], longTabList === 'scroll' && styles.tabNowrap, className)}
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
            {isFunction(children) ? children({ shouldRender }) : shouldRender ? children : null}
        </ReachTabPanel>
    )
}) as ForwardReferenceComponent<'div', TabPanelProps>
