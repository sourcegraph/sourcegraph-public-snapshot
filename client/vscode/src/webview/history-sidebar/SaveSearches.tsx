import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import React, { useState } from 'react'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/branded/src/components/SyntaxHighlightedSearchQuery'
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
    const [showMore, setShowMore] = useState(false)
    const [itemsToLoad, setItemsToLoad] = useState(5)

    function loadMoreItems(): void {
        setItemsToLoad(current => current + 5)
        telemetryService.log('RecentSearchesPanelShowMoreClicked')
    }
    // const [processedResults, setProcessedResults] = useState<string[] | null>(null)
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
                        .filter((search, index) => search.namespace.__typename === 'User')
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
                                        <SyntaxHighlightedSearchQuery query={search.query} />
                                    </Link>
                                </small>
                            </div>
                        ))}
                    {showMore && <ShowMoreButton onClick={loadMoreItems} />}
                </div>
            )}
        </div>
    )
}

const ShowMoreButton: React.FunctionComponent<{ onClick: () => void }> = ({ onClick }) => (
    <div className="text-center py-3">
        <button type="button" className={classNames('btn', styles.sidebarSectionButtonLink)} onClick={onClick}>
            Show more
        </button>
    </div>
)
