import React from 'react'

import classNames from 'classnames'
import AlertIcon from 'mdi-react/AlertIcon'
import CheckIcon from 'mdi-react/CheckIcon'

import { isDefined } from '@sourcegraph/common'
import { Badge } from '@sourcegraph/wildcard'

import { Timestamp } from '../../../../components/time/Timestamp'
import {
    LsifIndexFields,
    CodeIntelIndexerFields,
    LsifUploadFields,
    LSIFUploadState,
    LSIFIndexState,
} from '../../../../graphql-operations'
import { TelemetricRedirect } from '../../../../tracking/TelemetricRedirect'
import {
    useRequestedLanguageSupportQuery as defaultUseRequestedLanguageSupportQuery,
    useRequestLanguageSupportQuery as defaultUseRequestLanguageSupportQuery,
} from '../hooks/useCodeIntelStatus'

import { RequestLink } from './RequestLink'

import styles from './IndexerSummary.module.scss'

export interface IndexerSummaryProps {
    repoName: string
    summary: {
        name: string
        uploads: LsifUploadFields[]
        indexes: LsifIndexFields[]
        indexer?: CodeIntelIndexerFields
    }
    className?: string
    now?: () => Date
    useRequestedLanguageSupportQuery: typeof defaultUseRequestedLanguageSupportQuery
    useRequestLanguageSupportQuery: typeof defaultUseRequestLanguageSupportQuery
}

export const IndexerSummary: React.FunctionComponent<React.PropsWithChildren<IndexerSummaryProps>> = ({
    repoName,
    summary,
    className,
    now,
    useRequestedLanguageSupportQuery,
    useRequestLanguageSupportQuery,
}) => {
    const failedUploads = summary.uploads.filter(upload => upload.state === LSIFUploadState.ERRORED)
    const failedIndexes = summary.indexes.filter(index => index.state === LSIFIndexState.ERRORED)
    const finishedAtTimes = summary.uploads.map(upload => upload.finishedAt || undefined).filter(isDefined)
    const lastUpdated = finishedAtTimes.length === 0 ? undefined : finishedAtTimes.sort().reverse()[0]

    const telemetricRedirectClassName = classNames('m-0 p-0', styles.telemetricRedirect)

    return (
        <div className={classNames('px-2 py-1', styles.badgeWrapper)}>
            <div className="d-flex align-items-center">
                <div className="px-2 py-1 text-uppercase">
                    {summary.uploads.length + summary.indexes.length > 0 ? (
                        <Badge variant="success" small={true} className={className}>
                            Enabled
                        </Badge>
                    ) : summary.indexer?.url ? (
                        <Badge variant="secondary" small={true} className={className}>
                            Configurable
                        </Badge>
                    ) : (
                        <Badge variant="outlineSecondary" small={true} className={className}>
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
                            <TelemetricRedirect
                                to={summary.indexer.url}
                                label="Set up for this repository"
                                alwaysShowLabel={true}
                                eventName="CodeIntelligenceIndexerSetupInvestigated"
                                className={telemetricRedirectClassName}
                            />
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
                                    <TelemetricRedirect
                                        to={`/${repoName}/-/code-intelligence/uploads?filters=errored`}
                                        label="Latest upload processing"
                                        alwaysShowLabel={true}
                                        eventName="CodeIntelligenceUploadErrorInvestigated"
                                        className={telemetricRedirectClassName}
                                    />{' '}
                                    failed
                                </p>
                            )}
                            {failedIndexes.length > 0 && (
                                <p className="mb-1 text-muted">
                                    <AlertIcon size={16} className="text-danger" />{' '}
                                    <TelemetricRedirect
                                        to={`/${repoName}/-/code-intelligence/indexes?filters=errored`}
                                        label="Latest indexing"
                                        alwaysShowLabel={true}
                                        eventName="CodeIntelligenceIndexErrorInvestigated"
                                        className={telemetricRedirectClassName}
                                    />{' '}
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
