import React, { useCallback, useMemo } from 'react'

import { mdiArrowRight } from '@mdi/js'
import classNames from 'classnames'
import { useLocation, useNavigate } from 'react-router-dom'
import { of } from 'rxjs'

import { formatSearchParameters } from '@sourcegraph/common'
import { SearchMode, type SubmitSearchParameters } from '@sourcegraph/shared/src/search'
import { Icon, H3, H2, Text, Button, Link, createLinkUrl, useObservable } from '@sourcegraph/wildcard'

import { SearchPatternType } from '../../../../shared/src/graphql-operations'
import { LATEST_VERSION, aggregateStreamingSearch, type ProposedQuery } from '../../../../shared/src/search/stream'
import { smartSearchIconSvgPath } from '../input/toggles/SmartSearchToggle'

import { SyntaxHighlightedSearchQuery } from './SyntaxHighlightedSearchQuery'

import styles from './SmartSearchPreview.module.scss'

interface SmartSearchPreviewProps {
    setSearchMode: (mode: SearchMode) => void
    submitSearch: (parameters: SubmitSearchParameters) => void
    searchQueryFromURL: string
    caseSensitive: boolean
}

export const SmartSearchPreview: React.FunctionComponent<SmartSearchPreviewProps> = ({
    setSearchMode,
    submitSearch,
    searchQueryFromURL,
    caseSensitive,
}) => {
    const results = useObservable(
        useMemo(
            () =>
                aggregateStreamingSearch(of(searchQueryFromURL), {
                    version: LATEST_VERSION,
                    patternType: SearchPatternType.standard,
                    caseSensitive,
                    trace: undefined,
                    searchMode: SearchMode.SmartSearch,
                }),
            [searchQueryFromURL, caseSensitive]
        )
    )

    const resultNumber = useMemo(() => {
        if (results?.alert?.proposedQueries) {
            return results.alert.proposedQueries.reduce((acc: number, proposedQuery: ProposedQuery): number => {
                let proposedQueryResultCount = 0
                const proposedQueryResultCountGroup = proposedQuery.annotations?.filter(
                    ({ name }) => name === 'ResultCount'
                )

                if (proposedQueryResultCountGroup) {
                    for (const result of proposedQueryResultCountGroup) {
                        proposedQueryResultCount += parseInt(result.value.replaceAll(/\D/g, ''), 10)
                    }
                }
                acc += proposedQueryResultCount
                return acc
            }, 0)
        }
        return
    }, [results])

    if (results?.state === 'complete' && (!results?.alert?.proposedQueries || resultNumber === 0)) {
        return null
    }

    return (
        <div className="mb-5">
            {results?.state === 'loading' && (
                <>
                    <H3 as={H2}>Please wait. Smart Search is trying variations on your query...</H3>

                    <div className={classNames(styles.shimmerContainer, 'rounded my-3 col-6')}>
                        <div className={classNames(styles.shimmerAnimate, 'absolute top-0 overflow-hidden')} />
                    </div>

                    <div className={classNames(styles.shimmerContainer, 'rounded mb-3 col-4')}>
                        <div className={classNames(styles.shimmerAnimateSlower, 'absolute top-0 overflow-hidden')} />
                    </div>
                </>
            )}

            {results?.state === 'complete' && !!results?.alert?.proposedQueries && (
                <>
                    <H3 as={H2}>
                        However, Smart Smart found {Number(resultNumber) >= 500 ? `${resultNumber}+` : resultNumber}{' '}
                        results:
                    </H3>

                    <ul className={classNames('list-unstyled px-0 mb-2')}>
                        {results?.alert?.proposedQueries?.map(item => (
                            <li key={item.query} className="py-2">
                                <Link
                                    to={createLinkUrl({
                                        pathname: '/search',
                                        search: formatSearchParameters(new URLSearchParams({ q: item.query })),
                                    })}
                                    className="text-decoration-none"
                                >
                                    <Text className="mb-0">
                                        <span className="text-muted">{processDescription(item.description || '')}</span>
                                        <Icon svgPath={mdiArrowRight} aria-hidden={true} className="mx-2 text-body" />
                                        <span className="p-1 bg-code'">
                                            <SyntaxHighlightedSearchQuery
                                                query={item.query}
                                                searchPatternType={SearchPatternType.standard}
                                            />
                                        </span>
                                        {item.annotations
                                            ?.filter(({ name }) => name === 'ResultCount')
                                            ?.map(({ name, value }) => (
                                                <span key={name} className="text-muted ml-2">
                                                    {' '}
                                                    ({value})
                                                </span>
                                            ))}
                                    </Text>
                                </Link>
                            </li>
                        ))}
                    </ul>

                    <EnableSmartSearch
                        setSearchMode={setSearchMode}
                        submitSearch={submitSearch}
                        query={searchQueryFromURL}
                        caseSensitive={caseSensitive}
                    />
                </>
            )}
        </div>
    )
}

const processDescription = (description: string): string => {
    const split = description.split(' ⚬ ')

    split[0] = split[0][0].toUpperCase() + split[0].slice(1)
    return split.join(', ')
}

interface EnableSmartSearchProps {
    query: string
    caseSensitive: boolean
    setSearchMode: (mode: SearchMode) => void
    submitSearch: (parameters: SubmitSearchParameters) => void
}

const EnableSmartSearch: React.FunctionComponent<React.PropsWithChildren<EnableSmartSearchProps>> = ({
    query,
    caseSensitive,
    setSearchMode,
    submitSearch,
}) => {
    const navigate = useNavigate()
    const location = useLocation()

    const enableSmartSearch = useCallback((): void => {
        setSearchMode(SearchMode.SmartSearch)
        submitSearch({
            historyOrNavigate: navigate,
            location,
            query,
            patternType: SearchPatternType.standard,
            caseSensitive,
            searchMode: SearchMode.SmartSearch,
            source: 'smartSearchDisabled',
        })
    }, [query, navigate, location, caseSensitive, setSearchMode, submitSearch])

    return (
        <Text className="text-muted d-flex align-items-center mt-2">
            <Icon
                aria-hidden={true}
                svgPath={smartSearchIconSvgPath}
                className={classNames(styles.smartIcon, 'my-auto')}
            />
            <Button variant="link" className="px-0 mr-1" onClick={enableSmartSearch}>
                Enable Smart Search
            </Button>{' '}
            to find more related results.
        </Text>
    )
}
