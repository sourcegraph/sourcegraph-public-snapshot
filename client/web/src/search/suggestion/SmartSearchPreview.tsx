import React, { useCallback, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'

import { mdiArrowRight } from '@mdi/js'
import classNames from 'classnames'

import { of } from 'rxjs'
import { tap } from 'rxjs/operators'

import { SyntaxHighlightedSearchQuery, smartSearchIconSvgPath } from '@sourcegraph/branded'
import { SearchPatternType } from '../../../../shared/src/graphql-operations'
import { formatSearchParameters } from '@sourcegraph/common'
import { LATEST_VERSION, aggregateStreamingSearch } from '../../../../shared/src/search/stream'
import { SearchMode } from '@sourcegraph/shared/src/search'
import { Link, Icon, H3, H2, Text, Button, createLinkUrl, useObservable } from '@sourcegraph/wildcard'

import { useNavbarQueryState, setSearchMode } from '../../stores'
import { submitSearch } from '../helpers'

import styles from './QuerySuggestion.module.scss'

interface SmartSearchPreviewProps extends Pick<RouteComponentProps, 'history'> {}

export const SmartSearchPreview: React.FunctionComponent<React.PropsWithChildren<SmartSearchPreviewProps>> = ({
    history,
}) => {
    let resultsNum = 2

    //BE related results count
    //Move 'did you mean' under alert
    const query = useNavbarQueryState(state => state.searchQueryFromURL)

    const results = useObservable(
        useMemo(() => {
            return aggregateStreamingSearch(of(query), {
                version: LATEST_VERSION,
                patternType: SearchPatternType.standard,
                caseSensitive: false,
                trace: undefined,
                searchMode: SearchMode.SmartSearch,
            }).pipe(tap(() => {}))
        }, [query])
    )

    return (
        <>
            {results?.state === 'loading' ? (
                <div className="mb-5">
                    <H2 as={H3}>Please wait. Smart Search is trying variations on your query...</H2>
                    <ul className={classNames(styles.container, 'px-0 mb-3')}>
                        {results?.alert?.proposedQueries?.map(item => (
                            <li key={item.query}>
                                <Link
                                    to={createLinkUrl({
                                        pathname: '/search',
                                        search: formatSearchParameters(new URLSearchParams({ q: item.query })),
                                    })}
                                    className={classNames(styles.link, 'px-0')}
                                >
                                    <span className="p-1 bg-code">
                                        <SyntaxHighlightedSearchQuery
                                            query={item.query}
                                            searchPatternType={SearchPatternType.standard}
                                        />
                                    </span>
                                    <Icon svgPath={mdiArrowRight} aria-hidden={true} className="ml-2 mr-1 text-body" />
                                    <span>
                                        {item.annotations
                                            ?.filter(({ name }) => name === 'ResultCount')
                                            ?.map(({ name, value }) => (
                                                <span key={name} className="text-muted">
                                                    {' '}
                                                    {value}
                                                </span>
                                            ))}
                                    </span>
                                </Link>
                            </li>
                        ))}
                    </ul>

                    <EnableSmartSearch query={query} history={history} />
                </div>
            ) : results?.state === 'complete' && !!results?.alert?.proposedQueries ? (
                <div className="mb-5">
                    <H2 as={H3}>However, Smart Smart found {resultsNum} results:</H2>
                    <ul className={classNames(styles.container, 'px-0 mb-3')}>
                        {results?.alert?.proposedQueries?.map(item => (
                            <li key={item.query}>
                                <Link
                                    to={createLinkUrl({
                                        pathname: '/search',
                                        search: formatSearchParameters(new URLSearchParams({ q: item.query })),
                                    })}
                                    className={classNames(styles.link, 'px-0')}
                                >
                                    <span className="p-1 bg-code">
                                        <SyntaxHighlightedSearchQuery
                                            query={item.query}
                                            searchPatternType={SearchPatternType.standard}
                                        />
                                    </span>
                                    <Icon svgPath={mdiArrowRight} aria-hidden={true} className="ml-2 mr-1 text-body" />
                                    <span>
                                        {item.annotations
                                            ?.filter(({ name }) => name === 'ResultCount')
                                            ?.map(({ name, value }) => (
                                                <span key={name} className="text-muted">
                                                    {' '}
                                                    {value}
                                                </span>
                                            ))}
                                    </span>
                                </Link>
                            </li>
                        ))}
                    </ul>

                    <EnableSmartSearch query={query} history={history} />
                </div>
            ) : null}
        </>
    )
}

interface EnableSmartSearchProps extends Pick<RouteComponentProps, 'history'> {
    query: string
}

const EnableSmartSearch: React.FunctionComponent<React.PropsWithChildren<EnableSmartSearchProps>> = ({
    history,
    query,
}) => {
    const patternType = useNavbarQueryState(state => state.searchPatternType)
    const caseSensitive = useNavbarQueryState(state => state.searchCaseSensitivity)

    const enableSmartSearch = useCallback((): void => {
        setSearchMode(SearchMode.SmartSearch)
        submitSearch({
            history,
            query,
            patternType,
            caseSensitive,
            searchMode: SearchMode.SmartSearch,
            source: 'smartSearchDisabled',
        })
    }, [])

    return (
        <Text className="text-muted d-flex align-items-center">
            <Icon
                aria-hidden={true}
                svgPath={smartSearchIconSvgPath}
                className={classNames(styles.smartIcon, 'text-white my-auto')}
            />
            <Button variant="link" className="px-0 mr-1" onClick={enableSmartSearch}>
                Enable Smart Search
            </Button>{' '}
            to find more related results.
        </Text>
    )
}
