import React, { useState, useEffect, useCallback, useMemo } from 'react'

import classNames from 'classnames'
import { useHistory } from 'react-router'
import { of } from 'rxjs'

import { smartSearchIconSvgPath } from '@sourcegraph/branded'
import { SearchMode } from '@sourcegraph/shared/src/search'
import { Icon, H3, H2, Text, Button, useObservable } from '@sourcegraph/wildcard'

<<<<<<<< HEAD:client/branded/src/search-ui/components/SmartSearchPreview.tsx
import { submitSearch } from '../../../../web/src/search/helpers'
import { useNavbarQueryState, setSearchMode } from '../../../../web/src/stores'
import { SearchPatternType } from '../../../../shared/src/graphql-operations'
import { LATEST_VERSION, aggregateStreamingSearch, ProposedQuery } from '../../../../shared/src/search/stream'
========
import { SearchPatternType } from '../../../../shared/src/graphql-operations'
import { LATEST_VERSION, aggregateStreamingSearch, ProposedQuery } from '../../../../shared/src/search/stream'
import { useNavbarQueryState, setSearchMode } from '../../stores'
import { submitSearch } from '../helpers'
>>>>>>>> 8332d924d0bef7df0af435bbc5eca09f7f4d8de9:client/web/src/search/suggestion/SmartSearchPreview.tsx

import { SmartSearchListItem } from './SmartSearchListItem'

import styles from './SmartSearchPreview.module.scss'

export const SmartSearchPreview: React.FunctionComponent<{}> = () => {
    const [resultNumber, setResultNumber] = useState<number | string>(0)

    const caseSensitive = useNavbarQueryState(state => state.searchCaseSensitivity)
    const query = useNavbarQueryState(state => state.searchQueryFromURL)

    const results = useObservable(
        useMemo(
            () =>
                aggregateStreamingSearch(of(query), {
                    version: LATEST_VERSION,
                    patternType: SearchPatternType.standard,
                    caseSensitive,
                    trace: undefined,
                    searchMode: SearchMode.SmartSearch,
                }),
            [query, caseSensitive]
        )
    )

    useEffect(() => {
        if (results?.alert?.proposedQueries) {
            const resultNum: number = results.alert.proposedQueries.reduce(
                (acc: number, proposedQuery: ProposedQuery): number => {
                    let proposedQueryResultCount = 0
                    const proposedQueryResultCountGroup = proposedQuery.annotations?.filter(
                        ({ name }) => name === 'ResultCount'
                    )

                    if (proposedQueryResultCountGroup) {
                        for (const result of proposedQueryResultCountGroup) {
                            proposedQueryResultCount += parseInt(result.value.replace(/\D/g, ''), 10)
                        }
                    }
                    acc += proposedQueryResultCount
                    return acc
                },
                0
            )

            setResultNumber(resultNum)
        }
        return
    }, [results])

    if (results?.state === 'complete' && !results?.alert?.proposedQueries) {
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
                        However, Smart Smart found {resultNumber >= 500 ? `${resultNumber}+` : resultNumber} results:
                    </H3>

                    <ul className={classNames('list-unstyled px-0 mb-2')}>
                        {results?.alert?.proposedQueries?.map(item => (
                            <SmartSearchListItem proposedQuery={item} previewStyle={true} key={item.query} />
                        ))}
                    </ul>

                    <EnableSmartSearch query={query} caseSensitive={caseSensitive} />
                </>
            )}
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
    }, [query, history, caseSensitive])

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
