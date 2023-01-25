import React, { useState, useEffect, useCallback, useMemo } from 'react'
import { useHistory } from 'react-router'

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

import shimmerStyle from './SmartSearchPreview.module.scss'
import styles from './QuerySuggestion.module.scss'

export const SmartSearchPreview: React.FunctionComponent<{}> = () => {
    const [resultNumber, setResultNumber] = useState<number | string>(0)

    const caseSensitive = useNavbarQueryState(state => state.searchCaseSensitivity)
    const query = useNavbarQueryState(state => state.searchQueryFromURL)

    const results = useObservable(
        useMemo(() => {
            return aggregateStreamingSearch(of(query), {
                version: LATEST_VERSION,
                patternType: SearchPatternType.standard,
                caseSensitive,
                trace: undefined,
                searchMode: SearchMode.SmartSearch,
            }).pipe()
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
                        r => (proposedQueryResultCount += parseInt(r.value.replace(/\D/g, '')))
                    )
                    acc += proposedQueryResultCount
                    return acc
                },
                0
            )

            setResultNumber(resultNum)
        }
        return
    }, [results])

    if (results?.state === 'complete' && !!!results?.alert?.proposedQueries) {
        return null
    }

    return (
        <div className="mb-5">
            {results?.state === 'loading' && (
                <>
                    <H3 as={H2}>Please wait. Smart Search is trying variations on your query...</H3>

                    <div className={classNames(shimmerStyle.shimmerContainer, 'rounded my-3 col-6')}>
                        <div className={classNames(shimmerStyle.shimmerAnimate, 'absolute top-0 overflow-hidden')} />
                    </div>

                    <div className={classNames(shimmerStyle.shimmerContainer, 'rounded mb-3 col-4')}>
                        <div
                            className={classNames(shimmerStyle.shimmerAnimateSlower, 'absolute top-0 overflow-hidden')}
                        />
                    </div>
                </>
            )}

            {results?.state === 'complete' && !!results?.alert?.proposedQueries && (
                <>
                    <H3 as={H2}>
                        However, Smart Smart found {resultNumber >= 500 ? `${resultNumber}+` : resultNumber} results:
                    </H3>

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
                                        <SyntaxHighlightedSearchQuery query={item.query} />
                                    </span>
                                    <span className="ml-2 text-muted">({item.description})</span>
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
                </>
            )}

            <EnableSmartSearch query={query} caseSensitive={caseSensitive} />
        </div>
    )
}

interface EnableSmartSearchProps {
    query: string
    caseSensitive: boolean
}

const EnableSmartSearch: React.FunctionComponent<React.PropsWithChildren<EnableSmartSearchProps>> = ({
    query,
    caseSensitive,
}) => {
    const history = useHistory()

    const enableSmartSearch = useCallback((): void => {
        setSearchMode(SearchMode.SmartSearch)
        submitSearch({
            history,
            query,
            patternType: SearchPatternType.standard,
            caseSensitive,
            searchMode: SearchMode.SmartSearch,
            source: 'smartSearchDisabled',
        })
    }, [history])

    return (
        <Text className="text-muted d-flex align-items-center mt-2">
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
