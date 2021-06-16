import React from 'react'

import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { AggregateStreamingSearchResults } from '@sourcegraph/shared/src/search/stream'

import { ErrorAlert } from '../../components/alerts'

import { StreamingProgressCount } from './progress/StreamingProgressCount'

export const StreamingSearchResultFooter: React.FunctionComponent<{ results?: AggregateStreamingSearchResults }> = ({
    results,
}) => (
    <div className="d-flex flex-column align-items-center">
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
            <div className="alert alert-info d-flex m-3">
                <p className="m-0">No results</p>
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
    </div>
)
