import React, { useCallback, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'

import { mdiArrowRight } from '@mdi/js'

import { of } from 'rxjs'
import { tap } from 'rxjs/operators'

import { SyntaxHighlightedSearchQuery, smartSearchIconSvgPath } from '@sourcegraph/branded'
import { SearchPatternType } from '../../../../shared/src/graphql-operations'
import { formatSearchParameters } from '@sourcegraph/common'
import { LATEST_VERSION, aggregateStreamingSearch } from '../../../../shared/src/search/stream'
import { SearchMode } from '@sourcegraph/shared/src/search'
import { Link, Icon, H1, H2, Text, Button, createLinkUrl, useObservable } from '@sourcegraph/wildcard'

import { useNavbarQueryState, setSearchMode } from '../../stores'
import { submitSearch } from '../helpers'

import styles from './QuerySuggestion.module.scss'

interface SmartSearchPreviewProps extends Pick<RouteComponentProps, 'history'> {}

export const SmartSearchPreview: React.FunctionComponent<React.PropsWithChildren<SmartSearchPreviewProps>> = ({
    history,
}) => {
    let resultsNum = 2

    //BE related results count
    //Loading state - If results.state == 'loading' else if 'complete'
    //Polish styling
    //If no SS results, show nothing
    const query = useNavbarQueryState(state => state.searchQueryFromURL)
    const patternType = useNavbarQueryState(state => state.searchPatternType)
    const caseSensitive = useNavbarQueryState(state => state.searchCaseSensitivity)

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
        <div>
            <H2 className={styles.title} as={H1}>
                However, Smart Smart found {resultsNum} related results:
            </H2>
            <ul className={styles.container}>
                {results?.alert?.proposedQueries?.map(item => (
                    <li key={item.query} className={styles.listItem}>
                        <Link
                            to={createLinkUrl({
                                pathname: '/search',
                                search: formatSearchParameters(new URLSearchParams({ q: item.query })),
                            })}
                            className={styles.link}
                        >
                            <span className={styles.suggestion}>
                                <SyntaxHighlightedSearchQuery
                                    query={item.query}
                                    searchPatternType={SearchPatternType.standard}
                                />
                            </span>
                            <Icon svgPath={mdiArrowRight} aria-hidden={true} className="mx-2 text-body" />
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

            <span className="text-muted d-inline-block">
                <Icon aria-hidden={true} svgPath={smartSearchIconSvgPath} className={styles.smartIcon} />
                <Button variant="link" size="sm" className={styles.disableButton} onClick={enableSmartSearch}>
                    Enable Smart Search
                </Button>{' '}
                to find more related results.
            </span>
        </div>
    )
}
