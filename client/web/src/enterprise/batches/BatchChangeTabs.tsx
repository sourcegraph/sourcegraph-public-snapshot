import {
    Tab,
    TabList,
    TabPanel as BatchChangeTabPanel,
    TabPanels as BatchChangeTabPanels,
    Tabs,
    useTabsContext,
} from '@reach/tabs'
import classNames from 'classnames'
import * as H from 'history'
import React, { useCallback } from 'react'

import styles from './BatchChangeTabs.module.scss'

interface BatchChangeTabsProps {
    history: H.History
    location: H.Location
    /** The ordered, short names for each tab, used to read and write from the URL parameter `?tab=xxx` */
    tabNames: readonly [string, string, ...string[]]
}

export const BatchChangeTabs: React.FunctionComponent<BatchChangeTabsProps> = ({
    children,
    history,
    location,
    tabNames,
}) => {
    const defaultTab = tabNames[0]
    const urlParameters = new URLSearchParams(location.search)
    const tabParameter = urlParameters.get('tab') || defaultTab
    const initialTabIndex = tabNames.indexOf(tabParameter) || 0

    const onChange = useCallback(
        (newIndex: number): void => {
            const newTab = tabNames[newIndex]

            const urlParameters = new URLSearchParams(location.search)
            urlParameters.delete('visible')
            urlParameters.delete('first')
            urlParameters.delete('after')

            if (newTab === defaultTab) {
                urlParameters.delete('tab')
            } else {
                urlParameters.set('tab', newTab)
            }

            if (location.search !== urlParameters.toString()) {
                history.replace({ ...location, search: urlParameters.toString() })
            }
        },
        [defaultTab, history, location, tabNames]
    )

    return (
        <Tabs className={styles.batchChangeTabs} defaultIndex={initialTabIndex} onChange={onChange}>
            {children}
        </Tabs>
    )
}

export const BatchChangeTabsList: React.FunctionComponent = ({ children }) => (
    <div className="overflow-auto mb-2">
        <TabList className="nav nav-tabs d-inline-flex d-sm-flex flex-nowrap text-nowrap">{children}</TabList>
    </div>
)

interface BatchChangeTabProps {
    index: number
}

export const BatchChangeTab: React.FunctionComponent<BatchChangeTabProps> = ({ children, index }) => {
    const { selectedIndex } = useTabsContext()
    return <Tab className={classNames('nav-link', styles.navLink, { active: selectedIndex === index })}>{children}</Tab>
}

export { BatchChangeTabPanel, BatchChangeTabPanels }
