import classNames from 'classnames'
import React from 'react'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { StreamingProgressCount } from '@sourcegraph/branded/src/search/results/progress/StreamingProgressCount'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { AggregateStreamingSearchResults } from '@sourcegraph/shared/src/search/stream'

export const StreamingSearchResultFooter: React.FunctionComponent<{
    results?: AggregateStreamingSearchResults
    children?: React.ReactChild | React.ReactChild[]
    className?: string
}> = ({ results, children, className }) => (
    <div className={classNames(className, 'd-flex flex-column align-items-center')}>
        {(!results || results?.state === 'loading') && (
            <div className="text-center my-4" data-testid="loading-container">
                <LoadingSpinner className="icon-inline" />
            </div>
        )}

        {results?.state === 'complete' && results?.results.length > 0 && (
            <StreamingProgressCount progress={results.progress} state={results.state} className="mt-4 mb-2" />
        )}

        {results?.state === 'error' && (
            <ErrorAlert className="m-3" data-testid="search-results-list-error" error={results.error} />
        )}

        {results?.state === 'complete' && !results.alert && results?.results.length === 0 && (
            <div className="pr-3 mt-3 align-self-stretch">
                <div className="alert alert-info">
                    <p className="m-0">
                        <strong>No results matched your query</strong>
                        <br />
                        Use the tips below to improve your query.
                    </p>
                </div>
            </div>
        )}

        {results?.state === 'complete' && results.progress.skipped.some(skipped => skipped.reason.includes('-limit')) && (
            <div className="alert alert-info d-flex m-3">
                <p className="m-0">
                    <strong>Result limit hit.</strong> Modify your search with <code>count:</code> to return additional
                    items.
                </p>
            </div>
        )}

        {children}
    </div>
)
