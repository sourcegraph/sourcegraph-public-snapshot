import React, { useCallback, useState, useMemo } from 'react'

import { mdiChartLine } from '@mdi/js'
import * as H from 'history'
import { of } from 'rxjs'
import { catchError } from 'rxjs/operators'

import { asError, isErrorLike, ErrorLike } from '@sourcegraph/common'
import { Badge, Button, LoadingSpinner, useObservable, Alert, Icon, Code, H2, Input, Form } from '@sourcegraph/wildcard'

import { querySearchResultsStats } from './backend'
import { SearchStatsLanguages } from './SearchStatsLanguages'

interface Props {
    location: H.Location
    history: H.History

    /** Mockable in tests. */
    _querySearchResultsStats?: typeof querySearchResultsStats
}

/**
 * Shows statistics about the results for a search query.
 */
export const SearchStatsPage: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    location,
    history,
    _querySearchResultsStats = querySearchResultsStats,
}) => {
    const query = new URLSearchParams(location.search).get('q') || ''
    const [uncommittedQuery, setUncommittedQuery] = useState(query)
    const onUncommittedQueryChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setUncommittedQuery(event.currentTarget.value)
    }, [])
    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        event => {
            event.preventDefault()
            history.push({ ...location, search: new URLSearchParams({ q: uncommittedQuery }).toString() })
        },
        [history, location, uncommittedQuery]
    )

    const DEFAULT_COUNT = 1000
    const queryWithCount = query.includes('count:') ? query : `${query} count:${DEFAULT_COUNT}`

    // TODO(sqs): reuse the user's current patternType
    const stats = useObservable(
        useMemo(
            () => _querySearchResultsStats(queryWithCount).pipe(catchError(error => of<ErrorLike>(asError(error)))),
            [queryWithCount, _querySearchResultsStats]
        )
    )

    return (
        <div className="search-stats-page container mt-4">
            <header className="d-flex align-items-center justify-content-between mb-3">
                <H2 className="d-flex align-items-center mb-0">
                    <Icon aria-hidden={true} className="mr-2" svgPath={mdiChartLine} /> Code statistics{' '}
                    <Badge variant="secondary" className="text-uppercase ml-2" as="small">
                        Experimental
                    </Badge>
                </H2>
            </header>
            <Form onSubmit={onSubmit} className="form">
                <div className="form-group d-flex">
                    <Input
                        id="stats-page__query"
                        className="mb-0 w-100"
                        type="search"
                        placeholder="Enter a Sourcegraph search query"
                        value={uncommittedQuery}
                        onChange={onUncommittedQueryChange}
                        autoCapitalize="off"
                        spellCheck={false}
                        autoCorrect="off"
                        autoComplete="off"
                    />
                    {uncommittedQuery !== query && (
                        <Button type="submit" className="ml-2" variant="primary">
                            Update
                        </Button>
                    )}
                </div>
            </Form>
            <hr className="my-3" />
            {stats === undefined ? (
                <LoadingSpinner />
            ) : isErrorLike(stats) ? (
                <Alert variant="danger">{stats.message}</Alert>
            ) : stats.limitHit ? (
                <Alert variant="warning">
                    Limit hit. Add <Code>count:{DEFAULT_COUNT * 5}</Code> (or an even larger number) to your query to
                    retry with a higher limit.
                </Alert>
            ) : (
                <SearchStatsLanguages query={query} stats={stats} />
            )}
        </div>
    )
}
