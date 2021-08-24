import classNames from 'classnames'
import ChevronRightIcon from 'mdi-react/ChevronRightIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { FunctionComponent, useCallback, useEffect, useMemo } from 'react'
import { RouteComponentProps } from 'react-router'
import { of } from 'rxjs'

import { Link } from '@sourcegraph/shared/src/components/Link'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { useObservable } from '@sourcegraph/shared/src/util/useObservable'
import { Timestamp } from '@sourcegraph/web/src/components/time/Timestamp'
import { Container, PageHeader } from '@sourcegraph/wildcard'

import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '../../../components/FilteredConnection'
import { PageTitle } from '../../../components/PageTitle'
import { LsifUploadFields, LSIFUploadState } from '../../../graphql-operations'
import { fetchLsifUploads as defaultFetchLsifUploads } from '../shared/backend'
import { CodeIntelState } from '../shared/CodeIntelState'
import { CodeIntelUploadOrIndexCommit } from '../shared/CodeIntelUploadOrIndexCommit'
import { CodeIntelUploadOrIndexRepository } from '../shared/CodeIntelUploadOrIndexerRepository'
import { CodeIntelUploadOrIndexIndexer } from '../shared/CodeIntelUploadOrIndexIndexer'
import { CodeIntelUploadOrIndexLastActivity } from '../shared/CodeIntelUploadOrIndexLastActivity'
import { CodeIntelUploadOrIndexRoot } from '../shared/CodeIntelUploadOrIndexRoot'

import { fetchCommitGraphMetadata as defaultFetchCommitGraphMetadata } from './backend'

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
            <PageHeader headingElement="h2" path={[{ text: 'Precise code intelligence uploads' }]} className="mb-3" />

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
                        listClassName="codeintel-uploads__grid mb-3"
                        noun="upload"
                        pluralNoun="uploads"
                        nodeComponent={CodeIntelUploadNode}
                        nodeComponentProps={{ now }}
                        queryConnection={queryUploads}
                        history={props.history}
                        location={props.location}
                        cursorPaging={true}
                        filters={filters}
                        emptyElement={<EmptyLSIFUploadsElement />}
                    />
                </div>
            </Container>
        </div>
    )
}

const EmptyLSIFUploadsElement: React.FunctionComponent = () => (
    <p className="text-muted text-center w-100 mb-0 mt-1">
        <MapSearchIcon className="mb-2" />
        <br />
        No uploads yet. Enable precise code intelligence by{' '}
        <a
            href="https://docs.sourcegraph.com/code_intelligence/explanations/precise_code_intelligence"
            target="_blank"
            rel="noreferrer noopener"
        >
            uploading LSIF data
        </a>
        .
    </p>
)

interface CodeIntelUploadNodeProps {
    node: LsifUploadFields
    now?: () => Date
}

const CodeIntelUploadNode: FunctionComponent<CodeIntelUploadNodeProps> = ({ node, now }) => (
    <>
        <span className="codeintel-upload-node__separator" />

        <div className="d-flex flex-column codeintel-upload-node__information">
            <div className="m-0">
                <h3 className="m-0 d-block d-md-inline">
                    <CodeIntelUploadOrIndexRepository node={node} />
                </h3>
            </div>

            <div>
                <span className="mr-2 d-block d-mdinline-block">
                    Directory <CodeIntelUploadOrIndexRoot node={node} /> indexed at commit{' '}
                    <CodeIntelUploadOrIndexCommit node={node} /> by <CodeIntelUploadOrIndexIndexer node={node} />
                </span>

                <small className="text-mute">
                    <CodeIntelUploadOrIndexLastActivity node={{ ...node, queuedAt: null }} now={now} />
                </small>
            </div>
        </div>

        <span className="d-none d-md-inline codeintel-upload-node__state">
            <CodeIntelState node={node} className="d-flex flex-column align-items-center" />
        </span>
        <span>
            <Link to={`./uploads/${node.id}`}>
                <ChevronRightIcon />
            </Link>
        </span>
    </>
)

interface CommitGraphMetadataProps {
    stale: boolean
    updatedAt: Date | null
    className?: string
    now?: () => Date
}

const CommitGraphMetadata: FunctionComponent<CommitGraphMetadataProps> = ({ stale, updatedAt, className, now }) => (
    <>
        <div className={classNames('alert', stale ? 'alert-primary' : 'alert-success', className)}>
            {stale ? <StaleRepository /> : <FreshRepository />}{' '}
            {updatedAt && <LastUpdated updatedAt={updatedAt} now={now} />}
        </div>
    </>
)

const FreshRepository: FunctionComponent<{}> = () => <>Repository commit graph is currently up to date.</>

const StaleRepository: FunctionComponent<{}> = () => (
    <>
        Repository commit graph is currently stale and is queued to be refreshed. Refreshing the commit graph updates
        which uploads are visible from which commits.
    </>
)

interface LastUpdatedProps {
    updatedAt: Date
    now?: () => Date
}

const LastUpdated: FunctionComponent<LastUpdatedProps> = ({ updatedAt, now }) => (
    <>
        Last refreshed <Timestamp date={updatedAt} now={now} />.
    </>
)
