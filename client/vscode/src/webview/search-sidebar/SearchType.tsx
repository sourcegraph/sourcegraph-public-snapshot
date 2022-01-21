import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import React, { useState } from 'react'
import { UseStore } from 'zustand'
import shallow from 'zustand/shallow'

import { getSearchTypeLinks } from '@sourcegraph/branded/src/search/results/sidebar/SearchTypeLink'
import { SearchQueryState } from '@sourcegraph/shared/src/search/searchQueryState'

import { SearchPatternType } from '../../graphql-operations'
import { WebviewPageProps } from '../platform/context'

import styles from './HistorySidebar.module.scss'

interface SearchTypesProps extends Pick<WebviewPageProps, 'sourcegraphVSCodeExtensionAPI'> {
    caseSensitive: boolean
    useQueryState: UseStore<SearchQueryState>
    forceButton?: boolean
    patternType: SearchPatternType
}

const selectFromQueryState = ({
    queryState: { query },
    setQueryState,
    submitSearch,
}: SearchQueryState): {
    query: string
    setQueryState: SearchQueryState['setQueryState']
    submitSearch: SearchQueryState['submitSearch']
} => ({
    query,
    setQueryState,
    submitSearch,
})

export const SearchTypes: React.FunctionComponent<SearchTypesProps> = ({
    forceButton,
    useQueryState,
    patternType,
    caseSensitive,
}) => {
    const [collapsed, setCollapsed] = useState(false)
    const { query, setQueryState } = useQueryState(selectFromQueryState, shallow)

    return (
        <div className={styles.sidebarSection}>
            <button
                type="button"
                className={classNames('btn btn-outline-secondary', styles.sidebarSectionCollapseButton)}
                onClick={() => setCollapsed(!collapsed)}
            >
                <h5 className="flex-grow-1">Search Types</h5>
                {collapsed ? (
                    <ChevronLeftIcon className="icon-inline mr-1" />
                ) : (
                    <ChevronDownIcon className="icon-inline mr-1" />
                )}
            </button>
            {!collapsed && (
                <div className={classNames('p-1', styles.sidebarSectionList)}>
                    <small>
                        {getSearchTypeLinks({
                            caseSensitive,
                            onNavbarQueryChange: setQueryState,
                            patternType,
                            query,
                            selectedSearchContextSpec: undefined,
                            forceButton,
                        })}
                    </small>
                </div>
            )}
        </div>
    )
}
