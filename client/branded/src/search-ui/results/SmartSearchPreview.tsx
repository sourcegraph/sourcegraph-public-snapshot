import { mdiArrowRight } from '@mdi/js'
import { of } from 'rxjs'
import { tap, filter, map } from 'rxjs/operators'

import { SearchPatternType } from '../../../../shared/src/graphql-operations'
import {
    LATEST_VERSION,
    aggregateStreamingSearch,
    ProposedQuery,
    StreamSearchOptions,
} from '../../../../shared/src/search/stream'
import { SearchMode } from '@sourcegraph/shared/src/search'

import React, { useMemo } from 'react'
import { Icon, useObservable } from '@sourcegraph/wildcard'

interface SmartSearchPreviewProposedQueriesProps {
    proposedQueries: ProposedQuery[]
}
const SmartSearchPreviewProposedQueries: React.FunctionComponent<SmartSearchPreviewProposedQueriesProps> = ({
    proposedQueries,
}) => {
    return (
        <ul>
            {proposedQueries.map((proposedQuery, index) => (
                <li key={index}>
                    {/* <span>{proposedQuery.description}</span> */}
                    <span>{proposedQuery.query}</span>
                    <Icon className="" aria-hidden={true} svgPath={mdiArrowRight} />
                    <span>
                        {proposedQuery.annotations?.map(annotation => (
                            <span>{annotation.value}</span>
                        ))}
                    </span>
                </li>
            ))}
        </ul>
    )
}

interface SmartSearchPreviewProps {
    query: string
}

function smartSearchProposedQueriesSearch(query: string, options: StreamSearchOptions): Observable<ProposedQuery[]> {
    return aggregateStreamingSearch(of(query), options).pipe(
        filter(event => event && event.state === 'complete'),
        filter(event => event.alert?.kind === 'smart-search-pure-results'),
        map(event => event.alert?.proposedQueries),
        tap(results => {
            console.log('Smart search results:', results)
        })
    )
}

export const SmartSearchPreview: React.FunctionComponent<SmartSearchPreviewProps> = ({ query }) => {
    const options = {
        version: LATEST_VERSION,
        patternType: SearchPatternType.standard,
        caseSensitive: false,
        trace: undefined,
        searchMode: SearchMode.SmartSearch,
    }

    const proposedQueriesObservable = useMemo(() => smartSearchProposedQueriesSearch(query, options), [query])

    const proposedQueries = useObservable<ProposedQuery[]>(proposedQueriesObservable)

    // TODO: calculate the actual number from proposedQueries
    const totalResultsCount = 100

    return (
        <div>
            {proposedQueries ? (
                <>
                    <span>Smart Search found {totalResultsCount} results with variations of your query</span>
                    <SmartSearchPreviewProposedQueries proposedQueries={proposedQueries} />
                </>
            ) : (
                <div>Placeholder for loading indicator...</div>
            )}
        </div>
    )
}
