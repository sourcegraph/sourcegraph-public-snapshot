import React, { useState, useEffect, useCallback, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'

import { mdiArrowRight } from '@mdi/js'
import classNames from 'classnames'

import { of } from 'rxjs'
import { tap } from 'rxjs/operators'

import { SyntaxHighlightedSearchQuery, smartSearchIconSvgPath } from '@sourcegraph/branded'
import { SearchPatternType } from '../../../../shared/src/graphql-operations'
import { formatSearchParameters } from '@sourcegraph/common'
import { LATEST_VERSION, aggregateStreamingSearch, ProposedQuery } from '../../../../shared/src/search/stream'
import { SearchMode } from '@sourcegraph/shared/src/search'
import { Link, Icon, H3, H2, Text, Button, createLinkUrl, useObservable } from '@sourcegraph/wildcard'

import { useNavbarQueryState, setSearchMode } from '../../stores'
import { submitSearch } from '../helpers'

import styles from './QuerySuggestion.module.scss'

interface SmartSearchPreviewProps extends Pick<RouteComponentProps, 'history'> {}

export const SmartSearchPreview: React.FunctionComponent<React.PropsWithChildren<SmartSearchPreviewProps>> = ({
    history,
}) => {
    const [resultNumber, setResultNumber] = useState<number | string>(0)

    const query = useNavbarQueryState(state => state.searchQueryFromURL)
    const patternType = useNavbarQueryState(state => state.searchPatternType)
    const caseSensitive = useNavbarQueryState(state => state.searchCaseSensitivity)

    const results = useObservable(
        useMemo(() => {
            return aggregateStreamingSearch(of(query), {
                version: LATEST_VERSION,
                patternType,
                caseSensitive,
                trace: undefined,
                searchMode: SearchMode.SmartSearch,
            }).pipe(tap(() => {}))
        }, [query])
    )

    useEffect(() => {
        if (results?.alert?.proposedQueries) {
            const resultNum: number = results.alert.proposedQueries.reduce(
                (acc: number, proposedQuery: ProposedQuery): number => {
                    let proposedQueryResultCount = 0
                    const proposedQueryResultCountGroup = proposedQuery.annotations?.filter(
                        ({ name }) => name === 'ResultCount'
                    )
                    proposedQueryResultCountGroup?.forEach(
                        r => (proposedQueryResultCount += Number(r.value.replace(/\D/g, '')))
                    )
                    acc += proposedQueryResultCount
                    return acc
                },
                0
            )

            // SmartSearch count stops after 500
            setResultNumber(resultNum > 500 ? '500+' : resultNum)
        }
        return
    }, [results])

    return (
        <>
            {results?.state === 'loading' ? (
                <div className="mb-5">
                    <H2 as={H3}>Please wait. Smart Search is trying variations on your query...</H2>

                    <div className={classNames(styles.shimmerContainer, 'rounded my-3 col-6')}>
                        <div className={classNames(styles.shimmerAnimate, 'absolute top-0 overflow-hidden')} />
                    </div>

                    <div className={classNames(styles.shimmerContainer, 'rounded mb-3 col-4')}>
                        <div className={classNames(styles.shimmerAnimateSlower, 'absolute top-0 overflow-hidden')} />
                    </div>

                    <EnableSmartSearch
                        query={query}
                        history={history}
                        patternType={patternType}
                        caseSensitive={caseSensitive}
                    />
                </div>
            ) : results?.state === 'complete' && !!results?.alert?.proposedQueries ? (
                <div className="mb-5">
                    <H2 as={H3}>However, Smart Smart found {resultNumber} results:</H2>

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
                                            query={query}
                                            history={history}
                                            patternType={patternType}
                                            caseSensitive={caseSensitive}
                                        />
                                    </span>
                                    <Icon svgPath={mdiArrowRight} aria-hidden={true} className="ml-2 mr-1 text-body" />
                                    <span>
                                        {item.annotations
                                            ?.filter(({ name }) => name === 'ResultCount')
                                            ?.map(({ name, value }) => (
                                                <span key={name} className="text-muted">
                                                    {' '}
                                                    {value.replace('additional ', '')}
                                                </span>
                                            ))}
                                    </span>
                                </Link>
                            </li>
                        ))}
                    </ul>

                    <EnableSmartSearch
                        query={query}
                        history={history}
                        patternType={patternType}
                        caseSensitive={caseSensitive}
                    />
                </div>
            ) : null}
        </>
    )
}

interface EnableSmartSearchProps extends Pick<RouteComponentProps, 'history'> {
    query: string
    patternType: SearchPatternType
    caseSensitive: boolean
}

const EnableSmartSearch: React.FunctionComponent<React.PropsWithChildren<EnableSmartSearchProps>> = ({
    history,
    query,
    patternType,
    caseSensitive,
}) => {
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
