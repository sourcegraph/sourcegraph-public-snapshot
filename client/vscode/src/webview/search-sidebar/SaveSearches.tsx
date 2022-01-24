import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import React, { useState } from 'react'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { ISavedSearch } from '@sourcegraph/shared/src/graphql/schema'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebviewPageProps } from '../platform/context'

import styles from './HistorySidebar.module.scss'

interface SaveSearchesProps extends WebviewPageProps, TelemetryProps {
    savedSearches: ISavedSearch[]
}

export const SaveSearches: React.FunctionComponent<SaveSearchesProps> = ({
    savedSearches,
    sourcegraphVSCodeExtensionAPI,
    telemetryService,
    platformContext,
}) => {
    const itemsToLoad = 15
    const [collapsed, setCollapsed] = useState(false)

    return (
        <div className={styles.sidebarSection}>
            <button
                type="button"
                className={classNames('btn btn-outline-secondary', styles.sidebarSectionCollapseButton)}
                onClick={() => setCollapsed(!collapsed)}
            >
                <h5 className="flex-grow-1">Saved Searches</h5>
                {collapsed ? (
                    <ChevronLeftIcon className="icon-inline mr-1" />
                ) : (
                    <ChevronDownIcon className="icon-inline mr-1" />
                )}
            </button>
            {!collapsed && savedSearches && (
                <div className={classNames('p-1', styles.sidebarSectionList)}>
                    {savedSearches
                        // .filter((search, index) => search.namespace.__typename === 'User')
                        .filter((search, index) => index < itemsToLoad)
                        .map((search, index) => (
                            <div key={index}>
                                <small key={index} className={styles.sidebarSectionListItem}>
                                    <Link
                                        to="/"
                                        onClick={() =>
                                            sourcegraphVSCodeExtensionAPI.setActiveWebviewQueryState({
                                                query: search.query,
                                            })
                                        }
                                    >
                                        {search.description}
                                    </Link>
                                </small>
                            </div>
                        ))}
                </div>
            )}
        </div>
    )
}
