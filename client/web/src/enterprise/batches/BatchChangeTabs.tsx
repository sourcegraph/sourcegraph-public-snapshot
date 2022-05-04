import React, { useCallback, useEffect, useReducer, useState } from 'react'

import classNames from 'classnames'
import * as H from 'history'

import {
    Tab,
    TabList,
    TabPanel as BatchChangeTabPanel,
    TabPanels as BatchChangeTabPanels,
    Tabs,
} from '@sourcegraph/wildcard'

import { resetFilteredConnectionURLQuery } from '../../components/FilteredConnection'

import styles from './BatchChangeTabs.module.scss'

interface TabDetails {
    index: number
    customPath?: string
}

/**
 * Record of tab names and details with indices in the order that they appear in the UI,
 * which is derived from props on each `BatchChangeTab` and kept in context so that the
 * parent `BatchChangeTabs` can read and write from the URL location.
 */
type TabsState = Record<string, TabDetails>
interface TabsActionPayload {
    name: string
    details: TabDetails
}

const TabsStateContext = React.createContext<TabsState | undefined>(undefined)
const TabsDispatchContext = React.createContext<React.Dispatch<TabsActionPayload> | undefined>(undefined)

const tabsReducer = (state: TabsState, action: TabsActionPayload): TabsState => ({
    ...state,
    [action.name]: action.details,
})

const useTabsContext = (): TabsState => {
    const context = React.useContext(TabsStateContext)
    if (context === undefined) {
        throw new Error('useTabsContext must be used within a TabNamesProvider')
    }
    return context
}

const useTabsDispatch = (): React.Dispatch<TabsActionPayload> => {
    const context = React.useContext(TabsDispatchContext)
    if (context === undefined) {
        throw new Error('useTabsDispatch must be used within a TabNamesProvider')
    }
    return context
}

interface BatchChangeTabsProps {
    history: H.History
    location: H.Location
    /** The name of the tab that should be initially open */
    initialTab?: string
}

const BatchChangeTabsInternal: React.FunctionComponent<React.PropsWithChildren<BatchChangeTabsProps>> = ({
    children,
    history,
    location,
    initialTab,
}) => {
    // We are required to track the current tab locally in order to also control it from
    // the URL parameter.
    const [tabIndex, setTabIndex] = useState(0)
    const [customPath, setCustomPath] = useState<string | undefined>(undefined)
    const tabs = useTabsContext()
    const defaultTabName = Object.keys(tabs).find(tabName => tabs[tabName].index === 0)

    // Determine the initial tab from the URL parameters and react to changes in the URL.
    useEffect(() => {
        const initialTabName = new URLSearchParams(location.search).get('tab') || initialTab || defaultTabName
        if (initialTabName) {
            setTabIndex(tabs[initialTabName]?.index || 0)
            setCustomPath(tabs[initialTabName]?.customPath)
        }
    }, [initialTab, defaultTabName, location.search, tabs])

    const onChange = useCallback(
        (newIndex: number): void => {
            setTabIndex(newIndex)
            const newTabName = Object.keys(tabs).find(tabName => tabs[tabName].index === newIndex) || defaultTabName
            if (!newTabName) {
                return
            }

            const newTab = tabs[newTabName]

            const urlParameters = new URLSearchParams(location.search)
            resetFilteredConnectionURLQuery(urlParameters)

            // If the tab uses a custom path, we match the new URL path to it.
            if (newTab.customPath) {
                if (location.pathname.includes(newTab.customPath)) {
                    return
                }
                // Remember our custom path, so that we can remove it later
                setCustomPath(newTab.customPath)
                history.replace(location.pathname + newTab.customPath)
            } else {
                if (newTabName === defaultTabName) {
                    urlParameters.delete('tab')
                } else {
                    urlParameters.set('tab', newTabName)
                }
                // Make sure to unset any custom URL path.
                const newLocation = customPath
                    ? { ...location, pathname: location.pathname.replace(customPath, '') }
                    : location
                setCustomPath(undefined)

                history.replace({ ...newLocation, search: urlParameters.toString() })
            }
        },
        [defaultTabName, history, location, tabs, customPath]
    )

    return (
        <Tabs className={styles.batchChangeTabs} index={tabIndex} onChange={onChange} lazy={true}>
            {children}
        </Tabs>
    )
}

/** Wrapper of Wildcards's `Tabs` with built-in logic for reading and writing to the URL tab parameter */
export const BatchChangeTabs: React.FunctionComponent<React.PropsWithChildren<BatchChangeTabsProps>> = props => {
    const [state, dispatch] = useReducer(tabsReducer, {})
    return (
        <TabsStateContext.Provider value={state}>
            <TabsDispatchContext.Provider value={dispatch}>
                <BatchChangeTabsInternal {...props} />
            </TabsDispatchContext.Provider>
        </TabsStateContext.Provider>
    )
}

export const BatchChangeTabList: React.FunctionComponent<React.PropsWithChildren<unknown>> = ({ children }) => (
    <div className="overflow-auto mb-2">
        <TabList
            className={classNames(styles.batchChangeTabList, 'nav d-inline-flex d-sm-flex flex-nowrap text-nowrap')}
        >
            {children}
        </TabList>
    </div>
)

interface BatchChangeTabProps {
    index: number
    name: string
    /**
     * Optionally, if the tab should use a custom URL path, instead of the normal `tab`
     * URL parameter. Only supports the form "/customPath".
     */
    customPath?: string
}

export const BatchChangeTab: React.FunctionComponent<React.PropsWithChildren<BatchChangeTabProps>> = ({
    children,
    index,
    name,
    customPath,
}) => {
    const dispatch = useTabsDispatch()

    useEffect(() => {
        dispatch({ name, details: { index, customPath } })
    }, [index, name, customPath, dispatch])

    return <Tab>{children}</Tab>
}

export { BatchChangeTabPanels, BatchChangeTabPanel }
