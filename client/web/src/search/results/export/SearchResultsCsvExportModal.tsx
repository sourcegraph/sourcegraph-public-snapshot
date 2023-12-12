import React, { useCallback, useMemo, useState } from 'react'

import { type ErrorLike, isErrorLike, logger } from '@sourcegraph/common'
import type { PlatformContext } from '@sourcegraph/shared/src/platform/context'
import { FilterKind, findFilter } from '@sourcegraph/shared/src/search/query/query'
import type { AggregateStreamingSearchResults, StreamSearchOptions } from '@sourcegraph/shared/src/search/stream'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Alert, Button, Code, H3, Modal, Text } from '@sourcegraph/wildcard'

import { LoaderButton } from '../../../components/LoaderButton'
import { useFeatureFlag } from '../../../featureFlags/useFeatureFlag'

import { downloadSearchResults, EXPORT_RESULT_DISPLAY_LIMIT } from './searchResultsExport'

interface SearchResultsCsvExportModalProps extends Pick<PlatformContext, 'sourcegraphURL'>, TelemetryProps {
    query?: string
    options: StreamSearchOptions
    results?: AggregateStreamingSearchResults
    onClose: () => void
}

const MODAL_LABEL_ID = 'search-results-export-csv-modal-id'

export const SearchResultsCsvExportModal: React.FunctionComponent<SearchResultsCsvExportModalProps> = ({
    telemetryService,
    telemetryRecorder,
    sourcegraphURL,
    query = '',
    options,
    results,
    onClose,
}) => {
    const searchCompleted = results?.state === 'complete' || results?.state === 'error' // Allow exporting results even if there was an error

    const shouldRerunSearch = useMemo(
        () => searchCompleted && results?.progress.skipped.some(skipped => skipped.reason === 'display'),
        [results?.progress.skipped, searchCompleted]
    )

    const noTypeFilter = useMemo(
        () => !findFilter(query, 'type', FilterKind.Global) && !findFilter(query, 'select', FilterKind.Global),
        [query]
    )

    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<ErrorLike | undefined>()

    const [enableRepositoryMetadata] = useFeatureFlag('repository-metadata', true)

    const downloadResults = useCallback(() => {
        if (!searchCompleted) {
            return
        }

        if (query.includes('select:file.owners')) {
            telemetryService.log('searchResults:ownershipCsv:exported')
            telemetryRecorder.recordEvent('searchResults.ownershipCsv', 'exported')
        }

        setLoading(true)
        setError(undefined)

        downloadSearchResults(
            sourcegraphURL,
            query,
            { ...options, enableRepositoryMetadata },
            results,
            shouldRerunSearch
        )
            .then(() => {
                onClose()
            })
            .catch(error => {
                logger.error(error)
                if (isErrorLike(error)) {
                    setError(error)
                } else {
                    setError(new Error('An unknown error occurred when trying to export your search results.'))
                }
            })
            .finally(() => {
                setLoading(false)
            })
    }, [
        searchCompleted,
        query,
        sourcegraphURL,
        options,
        enableRepositoryMetadata,
        results,
        shouldRerunSearch,
        telemetryService,
        telemetryRecorder,
        onClose,
    ])

    return (
        <Modal aria-labelledby={MODAL_LABEL_ID}>
            <H3 id={MODAL_LABEL_ID}>Export search results</H3>

            <Text>Your search results will be exported as a CSV file.</Text>

            {!searchCompleted && (
                <Alert variant="danger">Your search has not completed. Please wait for the search to finish.</Alert>
            )}

            {shouldRerunSearch && (
                <Alert variant="warning">
                    Your search reached the maximum number of results that are displayed in the Sourcegraph UI. The
                    search will be re-run to export all results. This may take a while and your exported data may not be
                    in the same order as the original search. A maximum of{' '}
                    {EXPORT_RESULT_DISPLAY_LIMIT.toLocaleString('en-US')} results will be exported.
                </Alert>
            )}

            {noTypeFilter && (
                <Alert variant="warning">
                    Your search does not have a global <Code>type:</Code> or <Code>select:</Code> filter. If your search
                    produced results of multiple types, your exported results will only include results of the same type
                    as the first result.
                </Alert>
            )}

            <div className="d-flex justify-content-end">
                <Button disabled={loading} onClick={onClose} variant="secondary" className="mr-2">
                    Cancel
                </Button>
                <LoaderButton
                    label="Export"
                    onClick={downloadResults}
                    loading={loading}
                    disabled={!searchCompleted || loading}
                    variant="primary"
                />
            </div>

            {error && (
                <Alert variant="danger" className="mt-2">
                    {error.message}
                </Alert>
            )}
        </Modal>
    )
}
