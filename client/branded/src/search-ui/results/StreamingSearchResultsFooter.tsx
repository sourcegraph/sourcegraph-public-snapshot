import React from 'react'

import classNames from 'classnames'

import { AggregateStreamingSearchResults } from '@sourcegraph/shared/src/search/stream'
import { TelemetryService } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Alert, LoadingSpinner, Code, Text, Link, ErrorAlert } from '@sourcegraph/wildcard'

import { StreamingProgressCount } from './progress/StreamingProgressCount'

import styles from './StreamingSearchResultsList.module.scss'

export const StreamingSearchResultFooter: React.FunctionComponent<
    React.PropsWithChildren<{
        results?: AggregateStreamingSearchResults
        children?: React.ReactChild | React.ReactChild[]
        telemetryService: TelemetryService
    }>
> = ({ results, children, telemetryService }) => (
    <div className={classNames(styles.contentCentered, 'd-flex flex-column align-items-center')}>
        {(!results || results?.state === 'loading') && (
            <div className="text-center my-4" data-testid="loading-container">
                <LoadingSpinner />
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
                <Alert variant="info">
                    <Text className="m-0">
                        <strong>No results matched your query</strong>
                        <br />
                        Learn more about how to search{' '}
                        <Link
                            to="https://docs.sourcegraph.com/code_search/explanations/features"
                            rel="noopener noreferrer"
                            target="_blank"
                            onClick={() => telemetryService.log('ClickedOnDocs')}
                        >
                            in our docs
                        </Link>
                        , or use the tips below to improve your query.
                    </Text>
                </Alert>
            </div>
        )}

        {results?.state === 'complete' &&
            results.progress.skipped.some(skipped => skipped.reason.includes('-limit')) && (
                <Alert className="d-flex m-3" variant="info">
                    <Text className="m-0">
                        <strong>Result limit hit.</strong> Modify your search with <Code>count:</Code> to return
                        additional items.
                    </Text>
                </Alert>
            )}

        {children}
    </div>
)
