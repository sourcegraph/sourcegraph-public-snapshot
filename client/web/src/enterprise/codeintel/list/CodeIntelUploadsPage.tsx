import classNames from 'classnames'
import React, { FunctionComponent, useCallback, useEffect, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import { of } from 'rxjs'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '../../../components/FilteredConnection'
import { PageTitle } from '../../../components/PageTitle'
import { LsifUploadFields, LSIFUploadState } from '../../../graphql-operations'
import { fetchLsifUploads as defaultFetchLsifUploads } from '../shared/backend'

import { fetchCommitGraphMetadata as defaultFetchCommitGraphMetadata } from './backend'
import { CodeIntelUploadNode, CodeIntelUploadNodeProps } from './CodeIntelUploadNode'
import styles from './CodeIntelUploadsPage.module.scss'
import { CommitGraphMetadata } from './CommitGraphMetadata'
import { EmptyUploads } from './EmptyUploads'

export interface CodeIntelUploadsPageProps extends RouteComponentProps<{}>, TelemetryProps {
    repo?: { id: string }
    fetchLsifUploads?: typeof defaultFetchLsifUploads
    fetchCommitGraphMetadata?: typeof defaultFetchCommitGraphMetadata
    now?: () => Date
}

const filters: FilteredConnectionFilter[] = [
    {
        id: 'filters',
        label: 'Upload state',
        type: 'select',
        values: [
            {
                label: 'All',
                value: 'all',
                tooltip: 'Show all uploads',
                args: {},
            },
            {
                label: 'Current',
                value: 'current',
                tooltip: 'Show current uploads only',
                args: { isLatestForRepo: true },
            },
            {
                label: 'Completed',
                value: 'completed',
                tooltip: 'Show completed uploads only',
                args: { state: LSIFUploadState.COMPLETED },
            },
            {
                label: 'Errored',
                value: 'errored',
                tooltip: 'Show errored uploads only',
                args: { state: LSIFUploadState.ERRORED },
            },
            {
                label: 'Processing',
                value: 'processing',
                tooltip: 'Show processing uploads only',
                args: { state: LSIFUploadState.PROCESSING },
            },
            {
                label: 'Queued',
                value: 'queued',
                tooltip: 'Show queued uploads only',
                args: { state: LSIFUploadState.QUEUED },
            },
            {
                label: 'Uploading',
                value: 'uploading',
                tooltip: 'Show uploading uploads only',
                args: { state: LSIFUploadState.UPLOADING },
            },
            {
                label: 'Deleting',
                value: 'deleting',
                tooltip: 'Show uploads queued for deletion',
                args: { state: LSIFUploadState.DELETING },
            },
        ],
    },
]

export const CodeIntelUploadsPage: FunctionComponent<CodeIntelUploadsPageProps> = ({
    repo,
    fetchLsifUploads = defaultFetchLsifUploads,
    fetchCommitGraphMetadata = defaultFetchCommitGraphMetadata,
    now,
    telemetryService,
    ...props
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelUploads'), [telemetryService])

    const queryUploads = useCallback(
        (args: FilteredConnectionQueryArguments) => fetchLsifUploads({ repository: repo?.id, ...args }),
        [repo?.id, fetchLsifUploads]
    )

    const commitGraphMetadata = useObservable(
        useMemo(() => (repo ? fetchCommitGraphMetadata({ repository: repo?.id }) : of(undefined)), [
            repo,
            fetchCommitGraphMetadata,
        ])
    )

    return (
        <div className="code-intel-uploads">
            <PageTitle title="Precise code intelligence uploads" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Precise code intelligence uploads' }]}
                description={`LSIF indexes uploaded to Sourcegraph from CI or from auto-indexing ${
                    repo ? 'for this repository' : 'over all repositories'
                }.`}
                className="mb-3"
            />

            {repo && commitGraphMetadata && (
                <Container className="mb-2">
                    <CommitGraphMetadata
                        stale={commitGraphMetadata.stale}
                        updatedAt={commitGraphMetadata.updatedAt}
                        className="mb-0"
                        now={now}
                    />
                </Container>
            )}

            <Container>
                <div className="list-group position-relative">
                    <FilteredConnection<LsifUploadFields, Omit<CodeIntelUploadNodeProps, 'node'>>
                        listComponent="div"
                        listClassName={classNames(styles.grid, 'mb-3')}
                        noun="upload"
                        pluralNoun="uploads"
                        nodeComponent={CodeIntelUploadNode}
                        nodeComponentProps={{ now }}
                        queryConnection={queryUploads}
                        history={props.history}
                        location={props.location}
                        cursorPaging={true}
                        filters={filters}
                        emptyElement={<EmptyUploads />}
                    />
                </div>
            </Container>
        </div>
    )
}
