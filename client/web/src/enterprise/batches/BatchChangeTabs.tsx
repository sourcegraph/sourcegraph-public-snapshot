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

const BATCH_CHANGE_TAB_NAMES = ['changesets', 'chart', 'spec', 'archived'] as const

interface BatchChangeTabsProps {
    history: H.History
    location: H.Location
}

export const BatchChangeTabs: React.FunctionComponent<BatchChangeTabsProps> = ({ children, history, location }) => {
    const onChange = useCallback(
        (newIndex: number): void => {
            const newTab = BATCH_CHANGE_TAB_NAMES[newIndex]

            const urlParameters = new URLSearchParams(location.search)
            urlParameters.delete('visible')
            urlParameters.delete('first')
            urlParameters.delete('after')

            if (newTab === 'changesets') {
                urlParameters.delete('tab')
            } else {
                urlParameters.set('tab', newTab)
            }

            if (location.search !== urlParameters.toString()) {
                history.replace({ ...location, search: urlParameters.toString() })
            }
        },
        [history, location]
    )

    return (
        <Tabs className={styles.batchChangeTabs} onChange={onChange}>
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
    return (
        <Tab className="nav-item">
            {/* eslint-disable-next-line jsx-a11y/anchor-is-valid */}
            <a className={classNames('nav-link', { active: selectedIndex === index })} role="button">
                {children}
            </a>
        </Tab>
    )
}

export { BatchChangeTabPanel, BatchChangeTabPanels }
