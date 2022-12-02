import React from 'react'

import { mdiCheck, mdiAlert, mdiInformationOutline } from '@mdi/js'
import classNames from 'classnames'

import { isDefined } from '@sourcegraph/common'
import { Badge, Text, Icon } from '@sourcegraph/wildcard'

import { Timestamp } from '../../../../components/time/Timestamp'
import {
    LsifIndexFields,
    CodeIntelIndexerFields,
    LsifUploadFields,
    LSIFUploadState,
    LSIFIndexState,
} from '../../../../graphql-operations'
import { TelemetricLink } from '../../../../tracking/TelemetricLink'
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
        additionalIndexer: string[]
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
                    <Text className="mb-1">{summary.indexer?.name || summary.name} precise intelligence</Text>

                    {lastUpdated && (
                        <Text className="mb-1 text-muted">
                            Last updated: <Timestamp date={lastUpdated} now={now} />
                        </Text>
                    )}

                    {summary.uploads.length + summary.indexes.length === 0 ? (
                        summary.indexer?.url ? (
                            <TelemetricLink
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
                            {failedUploads.length === 0 &&
                                failedIndexes.length === 0 &&
                                summary.additionalIndexer.length === 0 && (
                                    <Text className="mb-1 text-muted">
                                        <Icon
                                            className="text-success"
                                            svgPath={mdiCheck}
                                            inline={false}
                                            aria-hidden={true}
                                            height={16}
                                            width={16}
                                        />{' '}
                                        Looks good!
                                    </Text>
                                )}
                            {summary.additionalIndexer.length > 0 && (
                                <Text className="mb-1 text-muted">
                                    <Icon
                                        svgPath={mdiInformationOutline}
                                        inline={false}
                                        aria-hidden={true}
                                        height={16}
                                        width={16}
                                    />{' '}
                                    Additional coverage available
                                </Text>
                            )}
                            {failedUploads.length > 0 && (
                                <Text className="mb-1 text-muted">
                                    <Icon
                                        className="text-danger"
                                        svgPath={mdiAlert}
                                        inline={false}
                                        aria-hidden={true}
                                        height={16}
                                        width={16}
                                    />{' '}
                                    <TelemetricLink
                                        to={`/${repoName}/-/code-graph/uploads?filters=errored`}
                                        label="Latest upload processing"
                                        alwaysShowLabel={true}
                                        eventName="CodeIntelligenceUploadErrorInvestigated"
                                        className={telemetricRedirectClassName}
                                    />{' '}
                                    failed
                                </Text>
                            )}
                            {failedIndexes.length > 0 && (
                                <Text className="mb-1 text-muted">
                                    <Icon
                                        className="text-danger"
                                        svgPath={mdiAlert}
                                        inline={false}
                                        aria-hidden={true}
                                        height={16}
                                        width={16}
                                    />{' '}
                                    <TelemetricLink
                                        to={`/${repoName}/-/code-graph/indexes?filters=errored`}
                                        label="Latest indexing"
                                        alwaysShowLabel={true}
                                        eventName="CodeIntelligenceIndexErrorInvestigated"
                                        className={telemetricRedirectClassName}
                                    />{' '}
                                    failed
                                </Text>
                            )}
                        </>
                    )}
                </div>
            </div>
        </div>
    )
}
