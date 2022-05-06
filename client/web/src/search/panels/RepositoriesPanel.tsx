import React, { useCallback, useEffect, useState } from 'react'

import { gql } from '@apollo/client'
import classNames from 'classnames'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { isRepoFilter } from '@sourcegraph/shared/src/search/query/validate'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Link } from '@sourcegraph/wildcard'

import { parseSearchURLQuery } from '..'
import { AuthenticatedUser } from '../../auth'
import { RecentlySearchedRepositoriesFragment } from '../../graphql-operations'
import { EventLogResult } from '../backend'

import { EmptyPanelContainer } from './EmptyPanelContainer'
import { HomePanelsFetchMore, RECENTLY_SEARCHED_REPOSITORIES_TO_LOAD } from './HomePanels'
import { LoadingPanelView } from './LoadingPanelView'
import { PanelContainer } from './PanelContainer'
import { ShowMoreButton } from './ShowMoreButton'

interface Props extends TelemetryProps {
    className?: string
    authenticatedUser: AuthenticatedUser | null
    recentlySearchedRepositories: RecentlySearchedRepositoriesFragment | null
    fetchMore: HomePanelsFetchMore
}

export const recentlySearchedRepositoriesFragment = gql`
    fragment RecentlySearchedRepositoriesFragment on User {
        recentlySearchedRepositoriesLogs: eventLogs(
            first: $firstRecentlySearchedRepositories
            eventName: "SearchResultsQueried"
        ) {
            nodes {
                argument
                timestamp
                url
            }
            pageInfo {
                hasNextPage
            }
            totalCount
        }
    }
`

export const RepositoriesPanel: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    className,
    telemetryService,
    recentlySearchedRepositories,
    fetchMore,
}) => {
    const [searchEventLogs, setSearchEventLogs] = useState<
        null | RecentlySearchedRepositoriesFragment['recentlySearchedRepositoriesLogs']
    >(recentlySearchedRepositories?.recentlySearchedRepositoriesLogs ?? null)
    useEffect(() => setSearchEventLogs(recentlySearchedRepositories?.recentlySearchedRepositoriesLogs ?? null), [
        recentlySearchedRepositories?.recentlySearchedRepositoriesLogs,
    ])

    const [itemsToLoad, setItemsToLoad] = useState(RECENTLY_SEARCHED_REPOSITORIES_TO_LOAD)

    const logRepoClicked = useCallback(() => telemetryService.log('RepositoriesPanelRepoFilterClicked'), [
        telemetryService,
    ])

    const loadingDisplay = <LoadingPanelView text="Loading recently searched repositories" />

    const emptyDisplay = (
        <EmptyPanelContainer className="text-muted">
            <small className="mb-2">
                <p className="mb-1">Recently searched repositories will be displayed here.</p>
                <p className="mb-1">
                    Search in repositories with the <strong>repo:</strong> filter:
                </p>
                <p className="mb-1">
                    <SyntaxHighlightedSearchQuery query="repo:sourcegraph/sourcegraph" />
                </p>
                <p className="mb-1">Add the code host to scope to a single repository:</p>
                <p className="mb-1">
                    <SyntaxHighlightedSearchQuery query="repo:^git\.local/my/repo$" />
                </p>
            </small>
        </EmptyPanelContainer>
    )

    const [repoFilterValues, setRepoFilterValues] = useState<string[] | null>(null)

    useEffect(() => {
        if (searchEventLogs) {
            const recentlySearchedRepos = processRepositories(searchEventLogs)
            setRepoFilterValues(recentlySearchedRepos)
        }
    }, [searchEventLogs])

    useEffect(() => {
        // Only log the first load (when items to load is equal to the page size)
        if (repoFilterValues && itemsToLoad === RECENTLY_SEARCHED_REPOSITORIES_TO_LOAD) {
            telemetryService.log(
                'RepositoriesPanelLoaded',
                { empty: repoFilterValues.length === 0 },
                { empty: repoFilterValues.length === 0 }
            )
        }
    }, [repoFilterValues, telemetryService, itemsToLoad])

    async function loadMoreItems(): Promise<void> {
        telemetryService.log('RepositoriesPanelShowMoreClicked')
        const newItemsToLoad = itemsToLoad + RECENTLY_SEARCHED_REPOSITORIES_TO_LOAD
        setItemsToLoad(newItemsToLoad)

        const { data } = await fetchMore({
            firstRecentlySearchedRepositories: newItemsToLoad,
        })

        if (data === undefined) {
            return
        }
        const node = data.node
        if (node === null || node.__typename !== 'User') {
            return
        }
        setSearchEventLogs(node.recentlySearchedRepositoriesLogs)
    }

    const contentDisplay = (
        <div className="mt-2">
            <div className="d-flex mb-1">
                <small>Search</small>
            </div>
            {repoFilterValues?.length && (
                <ul className="list-group">
                    {repoFilterValues.map((repoFilterValue, index) => (
                        <li key={`${repoFilterValue}-${index}`} className="text-monospace text-break mb-2">
                            <small>
                                <Link to={`/search?q=repo:${repoFilterValue}`} onClick={logRepoClicked}>
                                    <SyntaxHighlightedSearchQuery query={`repo:${repoFilterValue}`} />
                                </Link>
                            </small>
                        </li>
                    ))}
                </ul>
            )}
            {searchEventLogs?.pageInfo.hasNextPage && (
                <ShowMoreButton className="test-repositories-panel-show-more" onClick={loadMoreItems} />
            )}
        </div>
    )

    return (
        <PanelContainer
            className={classNames(className, 'repositories-panel')}
            title="Repositories"
            state={repoFilterValues ? (repoFilterValues.length > 0 ? 'populated' : 'empty') : 'loading'}
            loadingContent={loadingDisplay}
            populatedContent={contentDisplay}
            emptyContent={emptyDisplay}
        />
    )
}

function processRepositories(eventLogResult: EventLogResult): string[] | null {
    if (!eventLogResult) {
        return null
    }

    const recentlySearchedRepos: string[] = []

    for (const node of eventLogResult.nodes) {
        if (node.url) {
            const url = new URL(node.url)
            const queryFromURL = parseSearchURLQuery(url.search)
            const scannedQuery = scanSearchQuery(queryFromURL || '')
            if (scannedQuery.type === 'success') {
                for (const token of scannedQuery.term) {
                    if (isRepoFilter(token) && token.value && !recentlySearchedRepos.includes(token.value.value)) {
                        recentlySearchedRepos.push(token.value.value)
                    }
                }
            }
        }
    }
    return recentlySearchedRepos
}
