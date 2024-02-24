import React from 'react'

import classNames from 'classnames'

import type { AggregateStreamingSearchResults } from '@sourcegraph/shared/src/search/stream'
import { Alert, LoadingSpinner, Code, Text, H2, H3, ErrorAlert } from '@sourcegraph/wildcard'

import { StreamingProgressCount } from './progress/StreamingProgressCount'

import styles from './StreamingSearchResultsFooter.module.scss'

export const StreamingSearchResultFooter: React.FunctionComponent<
    React.PropsWithChildren<{
        results?: AggregateStreamingSearchResults
        children?: React.ReactChild | React.ReactChild[]
        className?: string
    }>
> = ({ results, children, className }) => {
    const skippedDisplay =
        results?.state === 'complete' && results.progress.skipped.find(skipped => skipped.reason.includes('display'))
    const resultLimitHit =
        results?.state === 'complete' && results.progress.skipped.some(skipped => skipped.reason.includes('-limit'))

    return (
        <div className={classNames(className, styles.root)}>
            {(!results || results?.state === 'loading') && (
                <div className="text-center" data-testid="loading-container">
                    <LoadingSpinner />
                </div>
            )}

            {results?.state === 'complete' && results?.results.length > 0 && (
                <StreamingProgressCount progress={results.progress} state={results.state} />
            )}

            {results?.state === 'error' && (
                <ErrorAlert className="m-3" data-testid="search-results-list-error" error={results.error} />
            )}

            {results?.state === 'complete' && !results.alert && results?.results.length === 0 && (
                <Alert variant="info">
                    <H3 as={H2} className="m-0 py-1">
                        No results matched your search.
                    </H3>
                </Alert>
            )}

            {(skippedDisplay || resultLimitHit) && (
                <Alert className="d-flex flex-column" variant="info">
                    {skippedDisplay && (
                        <Text className="m-0">
                            <strong>Display limit hit.</strong> {skippedDisplay.message}
                        </Text>
                    )}
                    {resultLimitHit && (
                        <Text className="m-0">
                            <strong>Result limit hit.</strong> Modify your query with <Code>count:</Code> to search for
                            more items.
                        </Text>
                    )}
                </Alert>
            )}

            {children}
        </div>
    )
}
