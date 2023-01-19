import React, { useCallback, useMemo } from 'react'

import { mdiArrowRight } from '@mdi/js'

import { of } from 'rxjs'
import { tap } from 'rxjs/operators'

import { SyntaxHighlightedSearchQuery, smartSearchIconSvgPath } from '@sourcegraph/branded'
import { SearchPatternType } from '../../../../shared/src/graphql-operations'
import { formatSearchParameters } from '@sourcegraph/common'
import { LATEST_VERSION, aggregateStreamingSearch } from '../../../../shared/src/search/stream'
import { SearchMode } from '@sourcegraph/shared/src/search'
import { Link, Icon, H1, H2, Text, Button, createLinkUrl, useObservable } from '@sourcegraph/wildcard'

import { submitSearch } from '../helpers'
import { useNavbarQueryState } from '../../stores'

import styles from './QuerySuggestion.module.scss'

interface SmartSearchPreviewProps {
    // alert: Required<AggregateStreamingSearchResults>['alert'] | undefined
    // onDisableSmartSearch: () => void
}

export const SmartSearchPreview: React.FunctionComponent<React.PropsWithChildren<SmartSearchPreviewProps>> = () => {
    let resultsNum = 2

    //TODO: Enable SmartSearch setting > submitSearch (parent props??)
    //BE related results count
    //Loading state - If results.state == 'loading' else if 'complete'
    //Polish styling
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
    console.log('END', results)

    const enableSmartSearch = useCallback(
        () =>
            submitSearch({
                // ...props,
                // caseSensitive,
                patternType: SearchPatternType.standard,
                query: query,
                source: 'smartSearchEnabled',
            }),
        [query]
    )

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
