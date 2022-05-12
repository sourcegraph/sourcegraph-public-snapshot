import React, { useMemo, useState } from 'react'

import classNames from 'classnames'
import ChevronDownIcon from 'mdi-react/ChevronDownIcon'
import ChevronLeftIcon from 'mdi-react/ChevronLeftIcon'
import { catchError } from 'rxjs/operators'

import { gql } from '@sourcegraph/http-client'
import { LATEST_VERSION } from '@sourcegraph/shared/src/search/stream'
import { Icon, Typography, useObservable } from '@sourcegraph/wildcard'

import { SavedSearchesResult, SavedSearchesVariables, SearchPatternType } from '../../../../graphql-operations'
import { HistorySidebarProps } from '../HistorySidebarView'

import styles from '../../search/SearchSidebarView.module.scss'

const savedSearchQuery = gql`
    query SavedSearches {
        savedSearches {
            ...SavedSearchFields
        }
    }
    fragment SavedSearchFields on SavedSearch {
        id
        description
        notify
        notifySlack
        query
        namespace {
            __typename
            id
            namespaceName
        }
        slackWebhookURL
    }
`

export const SavedSearchesSection: React.FunctionComponent<React.PropsWithChildren<HistorySidebarProps>> = ({
    platformContext,
    extensionCoreAPI,
}) => {
    const itemsToLoad = 15
    const [collapsed, setCollapsed] = useState(false)

    const savedSearchesResult = useObservable(
        useMemo(
            () =>
                platformContext
                    .requestGraphQL<SavedSearchesResult, SavedSearchesVariables>({
                        request: savedSearchQuery,
                        variables: {},
                        mightContainPrivateInfo: true,
                    })
                    .pipe(
                        catchError(error => {
                            console.error('Error fetching saved searches', error)
                            return [null]
                        })
                    ),
            [platformContext]
        )
    )

    const savedSearches = savedSearchesResult?.data?.savedSearches

    if (!savedSearches || savedSearches.length === 0) {
        return null
    }

    const onSavedSearchClick = (query: string): void => {
        platformContext.telemetryService.log('VSCESidebarSavedSearchClick')
        extensionCoreAPI
            .streamSearch(query, {
                // Debt: using defaults here. The saved search should override these, though.
                caseSensitive: false,
                patternType: SearchPatternType.literal,
                version: LATEST_VERSION,
                trace: undefined,
            })
            .catch(error => {
                // TODO surface to user
                console.error('Error submitting search from Sourcegraph sidebar', error)
            })
    }

    return (
        <div className={styles.sidebarSection}>
            <button
                type="button"
                className={classNames('btn btn-outline-secondary', styles.sidebarSectionCollapseButton)}
                onClick={() => setCollapsed(!collapsed)}
                aria-label={`${collapsed ? 'Expand' : 'Collapse'} saved searches`}
            >
                <Typography.H5 className="flex-grow-1">Saved Searches</Typography.H5>
                <Icon
                    role="img"
                    aria-hidden={true}
                    className="mr-1"
                    as={collapsed ? ChevronLeftIcon : ChevronDownIcon}
                />
            </button>

            {!collapsed && savedSearches && (
                <div className={classNames('p-1', styles.sidebarSectionList)}>
                    {savedSearches
                        .filter((search, index) => index < itemsToLoad)
                        .map(search => (
                            <div key={search.id}>
                                <small className={styles.sidebarSectionListItem}>
                                    <button
                                        type="button"
                                        className="btn btn-link p-0 text-left text-decoration-none"
                                        onClick={() => onSavedSearchClick(search.query)}
                                    >
                                        {search.description}
                                    </button>
                                </small>
                            </div>
                        ))}
                </div>
            )}
        </div>
    )
}
