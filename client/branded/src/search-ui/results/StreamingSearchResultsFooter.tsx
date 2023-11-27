import React from 'react'

import { mdiBookSearch } from '@mdi/js'
import classNames from 'classnames'

import type { AggregateStreamingSearchResults } from '@sourcegraph/shared/src/search/stream'
import { Alert, LoadingSpinner, Code, Text, H2, H3, ErrorAlert, Icon, Button } from '@sourcegraph/wildcard'

import { StreamingProgressCount } from './progress/StreamingProgressCount'

import styles from './StreamingSearchResultsFooter.module.scss'

export const StreamingSearchResultFooter: React.FunctionComponent<
    React.PropsWithChildren<{
        results?: AggregateStreamingSearchResults
        children?: React.ReactChild | React.ReactChild[]
    }>
> = ({ results, children }) => (
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
            <div className="pr-3 mt-5 d-flex flex-column align-items-center justify-content-center">
                <Icon aria-hidden={true} svgPath={mdiBookSearch} size="md" className={styles.searchIcon} />
                <H3 as={H2}>We couldn't find a match for your search query.</H3>
                <Button variant="primary">Update Search Query</Button>
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
