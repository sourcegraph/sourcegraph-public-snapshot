import React, { useCallback, useEffect, useState, useMemo } from 'react'

import { gql } from '@apollo/client'
import classNames from 'classnames'
import { of } from 'rxjs'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { scanSearchQuery } from '@sourcegraph/shared/src/search/query/scanner'
import { isRepoFilter } from '@sourcegraph/shared/src/search/query/validate'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Link, Text, useObservable } from '@sourcegraph/wildcard'

import { parseSearchURLQuery } from '..'
import { streamComputeQuery } from '../../../../shared/src/search/stream'
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
    authenticatedUser,
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
                <Text className="mb-1">Recently searched repositories will be displayed here.</Text>
                <Text className="mb-1">
                    Search in repositories with the <strong>repo:</strong> filter:
                </Text>
                <Text className="mb-1">
                    <SyntaxHighlightedSearchQuery query="repo:sourcegraph/sourcegraph" />
                </Text>
                <Text className="mb-1">Add the code host to scope to a single repository:</Text>
                <Text className="mb-1">
                    <SyntaxHighlightedSearchQuery query="repo:^git\.local/my/repo$" />
                </Text>
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

    // a constant to hold git commits history
    // call streamComputeQuery from stream

    const gitRepository = useObservable(useMemo(() =>
    authenticatedUser ? streamComputeQuery(`content:output((.|\n)* -> $repo) author:${authenticatedUser.email} type:commit after:"1 year ago" count:all`): of([]), [authenticatedUser]))
    console.log(gitRepository)
    let gitRepositoryParsedString
    if (gitRepository) {
        gitRepositoryParsedString = gitRepository.map(value => JSON.parse(value))
    }
    const gitReposList = gitRepositoryParsedString?.flat()
    /*
    Algorithm:
        1.Get the user's git history
        2.Get the user's search history
        3.If the user has git history,
            then show the git history
        4.If the user has no git history,
            then check if the user has search history
        5.If the user has search history,
            then show the search history
        6.If the user has no search history,
            then show the empty display
    */

    // A new display for git history
    const gitHistoryDisplay = (
        <div className="mt-2">
            <div className="d-flex mb-1">
                <small>Git history</small>
            </div>
            {gitRepositoryParsedString?.length && (
                <ul className="list-group">
                    {gitRepositoryParsedString.map(repo =>
                        <li key={`${repo.value}-`} className="text-monospace text-break mb-2">
                            <small>
                                <Link to={`/search?q=repo:${repo.value}`} onClick={logRepoClicked}>
                                    <SyntaxHighlightedSearchQuery query={`repo:${repo.value}`} />
                                </Link>
                            </small>
                        </li>
                    )}
                </ul>
            )}
            {searchEventLogs?.pageInfo.hasNextPage && (
                <ShowMoreButton className="test-repositories-panel-show-more" onClick={loadMoreItems} />
            )}
        </div>
    )
    // 1. Get the user's git history
    // create a SET object to hold the git commit history
    const gitSet = new Set<string>()
    if (gitReposList) {
        for (const git of gitReposList) {
            gitSet.add(git.repo)
        }
    }
    console.log(gitSet)

    // 2. Get the user's search history
    const codeSearchHistory = useMemo(() => processRepositories(searchEventLogs), [searchEventLogs])
    // 3. If the user has git history,
    // then show the git history
    /*
    if (gitSet.size > 0) {
        return gitHistoryDisplay
    }
        // 4. If the user has no git history,
        // then check if the user has search history
    if (codeSearchHistory && codeSearchHistory.length > 0) {
        // 5. If the user has search history,
        // then show the search history
        return contentDisplay
    } /* else {
        // 6. If the user has no search history,
        // then show the empty display
        return emptyDisplay
    }*/

    return (
        <PanelContainer
            className={classNames(className, 'repositories-panel')}
            title="Repositories"
            // state={repoFilterValues ? (repoFilterValues.length > 0 ? 'populated' : 'empty') ? (gitSet.size > 0 ? 'gitPopulated' : 'empty') : 'loading' : 'loading'}
            // condition ? exprIfTrue : exprIfFalse
            state={repoFilterValues ? (repoFilterValues.length > 0 ? 'populated' : 'empty') ? (gitSet.size > 0 ? 'gitPopulated' : 'empty') : 'loading' : 'loading'}
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
