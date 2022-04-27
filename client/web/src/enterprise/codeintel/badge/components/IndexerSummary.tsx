import React from 'react'

import AlertIcon from 'mdi-react/AlertIcon'
import CheckIcon from 'mdi-react/CheckIcon'

import { isDefined } from '@sourcegraph/common'
import { Badge, Link } from '@sourcegraph/wildcard'

import { Timestamp } from '../../../../components/time/Timestamp'
import {
    LsifIndexFields,
    CodeIntelIndexerFields,
    LsifUploadFields,
    LSIFUploadState,
    LSIFIndexState,
} from '../../../../graphql-operations'
import {
    useRequestedLanguageSupportQuery as defaultUseRequestedLanguageSupportQuery,
    useRequestLanguageSupportQuery as defaultUseRequestLanguageSupportQuery,
} from '../hooks/useCodeIntelStatus'

import { RequestLink } from './RequestLink'

export interface IndexerSummaryProps {
    repoName: string
    summary: {
        name: string
        uploads: LsifUploadFields[]
        indexes: LsifIndexFields[]
        indexer?: CodeIntelIndexerFields
    }
    className?: string
    useRequestedLanguageSupportQuery: typeof defaultUseRequestedLanguageSupportQuery
    useRequestLanguageSupportQuery: typeof defaultUseRequestLanguageSupportQuery
    now?: () => Date
}

export const IndexerSummary: React.FunctionComponent<IndexerSummaryProps> = ({
    repoName,
    summary,
    className,
    useRequestedLanguageSupportQuery,
    useRequestLanguageSupportQuery,
    now,
}) => {
    const failedUploads = summary.uploads.filter(upload => upload.state === LSIFUploadState.ERRORED)
    const failedIndexes = summary.indexes.filter(index => index.state === LSIFIndexState.ERRORED)
    const finishedAtTimes = summary.uploads.map(upload => upload.finishedAt || undefined).filter(isDefined)
    const lastUpdated = finishedAtTimes.length === 0 ? undefined : finishedAtTimes.sort().reverse()[0]

    return (
        <div className="px-2 py-1">
            <div className="d-flex align-items-center">
                <div className="px-2 py-1 text-uppercase">
                    {summary.uploads.length + summary.indexes.length > 0 ? (
                        <Badge variant="success" className={className}>
                            Enabled
                        </Badge>
                    ) : summary.indexer?.url ? (
                        <Badge variant="secondary" className={className}>
                            Configurable
                        </Badge>
                    ) : (
                        <Badge variant="outlineSecondary" className={className}>
                            Unavailable
                        </Badge>
                    )}
                </div>

                <div className="px-2 py-1">
                    <p className="mb-1">{summary.indexer?.name || summary.name} precise intelligence</p>

                    {lastUpdated && (
                        <p className="mb-1 text-muted">
                            Last updated: <Timestamp date={lastUpdated} now={now} />
                        </p>
                    )}

                    {summary.uploads.length + summary.indexes.length === 0 ? (
                        summary.indexer?.url ? (
                            <Link to={summary.indexer?.url}>Set up for this repository</Link>
                        ) : (
                            <RequestLink
                                indexerName={summary.name}
                                useRequestedLanguageSupportQuery={useRequestedLanguageSupportQuery}
                                useRequestLanguageSupportQuery={useRequestLanguageSupportQuery}
                            />
                        )
                    ) : (
                        <>
                            {failedUploads.length === 0 && failedIndexes.length === 0 && (
                                <p className="mb-1 text-muted">
                                    <CheckIcon size={16} className="text-success" /> Looks good!
                                </p>
                            )}
                            {failedUploads.length > 0 && (
                                <p className="mb-1 text-muted">
                                    <AlertIcon size={16} className="text-danger" />{' '}
                                    <Link to={`/${repoName}/-/code-intelligence/uploads?filters=errored`}>
                                        Latest upload processing
                                    </Link>{' '}
                                    failed
                                </p>
                            )}
                            {failedIndexes.length > 0 && (
                                <p className="mb-1 text-muted">
                                    <AlertIcon size={16} className="text-danger" />{' '}
                                    <Link to={`/${repoName}/-/code-intelligence/indexes?filters=errored`}>
                                        Latest indexing
                                    </Link>{' '}
                                    failed
                                </p>
                            )}
                        </>
                    )}
                </div>
            </div>
        </div>
    )
}
