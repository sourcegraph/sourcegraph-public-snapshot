import {
    Tab,
    TabList,
    TabPanel,
    TabPanelProps,
    TabPanels as BatchChangeTabPanels,
    Tabs,
    useTabsContext,
} from '@reach/tabs'
import classNames from 'classnames'
import * as H from 'history'
import React, { useCallback, useEffect, useReducer, useState } from 'react'

import { resetFilteredConnectionURLQuery } from '../../components/FilteredConnection'

import styles from './BatchChangeTabs.module.scss'

/**
 * Record of tab names and the indices for the order that they appear in the UI, which is
 * derived from props on each `BatchChangeTab` and kept in context so that the parent
 * `BatchChangeTabs` can read and write from the URL parameters
 */
type TabNamesState = Record<string, number>
interface TabNamesAction {
    tabName: string
    tabIndex: number
}

const TabNamesStateContext = React.createContext<TabNamesState | undefined>(undefined)
const TabNamesDispatchContext = React.createContext<React.Dispatch<TabNamesAction> | undefined>(undefined)

const tabsReducer = (state: TabNamesState, action: TabNamesAction): TabNamesState => ({
    ...state,
    [action.tabName]: action.tabIndex,
})

const useTabNamesContext = (): TabNamesState => {
    const context = React.useContext(TabNamesStateContext)
    if (context === undefined) {
        throw new Error('useTabNamesContext must be used within a TabNamesProvider')
    }
    return context
}

const useTabNamesDispatch = (): React.Dispatch<TabNamesAction> => {
    const context = React.useContext(TabNamesDispatchContext)
    if (context === undefined) {
        throw new Error('useTabNamesDispatch must be used within a TabNamesProvider')
    }
    return context
}

interface BatchChangeTabsProps {
    history: H.History
    location: H.Location
}

const BatchChangeTabs_: React.FunctionComponent<BatchChangeTabsProps> = ({ children, history, location }) => {
    // We are required to track the current tab locally in order to also control it from the URL parameter
    const [tabIndex, setTabIndex] = useState(0)
    const tabNames = useTabNamesContext()
    const defaultTabName = Object.keys(tabNames).find(key => tabNames[key] === 0)

    // Determine the initial tab from the URL parameters
    useEffect(() => {
        const initialTabName = new URLSearchParams(location.search).get('tab') || defaultTabName
        setTabIndex(initialTabName ? tabNames[initialTabName] || 0 : 0)
    }, [defaultTabName, location.search, tabNames])

    const onChange = useCallback(
        (newIndex: number): void => {
            setTabIndex(newIndex)
            const newTabName = Object.keys(tabNames).find(key => tabNames[key] === newIndex) || defaultTabName

            const urlParameters = new URLSearchParams(location.search)
            resetFilteredConnectionURLQuery(urlParameters)

            if (!newTabName || newTabName === defaultTabName) {
                urlParameters.delete('tab')
            } else {
                urlParameters.set('tab', newTabName)
            }

            if (location.search !== urlParameters.toString()) {
                history.replace({ ...location, search: urlParameters.toString() })
            }
        },
        [defaultTabName, history, location, tabNames]
    )

    return (
        <Tabs className={styles.batchChangeTabs} index={tabIndex} onChange={onChange}>
            {children}
        </Tabs>
    )
}

/** Wrapper of ReachUI's `Tabs` with built-in logic for reading and writing to the URL tab parameter */
export const BatchChangeTabs: React.FunctionComponent<BatchChangeTabsProps> = props => {
    const [state, dispatch] = useReducer(tabsReducer, {})
    return (
        <TabNamesStateContext.Provider value={state}>
            <TabNamesDispatchContext.Provider value={dispatch}>
                <BatchChangeTabs_ {...props} />
            </TabNamesDispatchContext.Provider>
        </TabNamesStateContext.Provider>
    )
}

export const BatchChangeTabList: React.FunctionComponent = ({ children }) => (
    <div className="overflow-auto mb-2">
        <TabList className="nav nav-tabs d-inline-flex d-sm-flex flex-nowrap text-nowrap">{children}</TabList>
    </div>
)

interface BatchChangeTabProps {
    index: number
    name: string
}

export const BatchChangeTab: React.FunctionComponent<BatchChangeTabProps> = ({ children, index, name }) => {
    const { selectedIndex } = useTabsContext()
    const dispatch = useTabNamesDispatch()

    useEffect(() => {
        dispatch({ tabName: name, tabIndex: index })
    }, [index, name, dispatch])

    return <Tab className={classNames('nav-link', { active: selectedIndex === index })}>{children}</Tab>
}

/** Wrapper of ReachUI's `TabPanel`, but that only renders its children if the tab is active */
export const BatchChangeTabPanel: React.FunctionComponent<TabPanelProps & { index: number }> = ({
    children,
    index,
}) => {
    const { selectedIndex } = useTabsContext()
    return <TabPanel>{selectedIndex === index ? children : null}</TabPanel>
}

export { BatchChangeTabPanels }
