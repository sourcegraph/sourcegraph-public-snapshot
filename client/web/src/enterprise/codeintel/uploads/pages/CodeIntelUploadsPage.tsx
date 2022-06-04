import { FunctionComponent, useCallback, useEffect, useMemo, useState } from 'react'

import { useApolloClient } from '@apollo/client'
import classNames from 'classnames'
import { RouteComponentProps } from 'react-router'
import { of } from 'rxjs'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Container, PageHeader, useObservable } from '@sourcegraph/wildcard'

import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import { LsifUploadFields, LSIFUploadState } from '../../../../graphql-operations'
import { FlashMessage } from '../../configuration/components/FlashMessage'
import { queryCommitGraphMetadata as defaultQueryCommitGraphMetadata } from '../../indexes/hooks/queryCommitGraphMetadata'
import { CodeIntelUploadNode, CodeIntelUploadNodeProps } from '../components/CodeIntelUploadNode'
import { CommitGraphMetadata } from '../components/CommitGraphMetadata'
import { EmptyUploads } from '../components/EmptyUploads'
import { queryLsifUploadsByRepository as defaultQueryLsifUploadsByRepository } from '../hooks/queryLsifUploadsByRepository'
import { queryLsifUploadsList as defaultQueryLsifUploadsList } from '../hooks/queryLsifUploadsList'

import styles from './CodeIntelUploadsPage.module.scss'

export interface CodeIntelUploadsPageProps extends RouteComponentProps<{}>, TelemetryProps {
    repo?: { id: string }
    queryLsifUploadsByRepository?: typeof defaultQueryLsifUploadsByRepository
    queryLsifUploadsList?: typeof defaultQueryLsifUploadsList
    queryCommitGraphMetadata?: typeof defaultQueryCommitGraphMetadata
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

export const CodeIntelUploadsPage: FunctionComponent<React.PropsWithChildren<CodeIntelUploadsPageProps>> = ({
    repo,
    queryLsifUploadsByRepository = defaultQueryLsifUploadsByRepository,
    queryLsifUploadsList = defaultQueryLsifUploadsList,
    queryCommitGraphMetadata = defaultQueryCommitGraphMetadata,
    now,
    telemetryService,
    history,
    ...props
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelUploads'), [telemetryService])

    const apolloClient = useApolloClient()
    const queryLsifUploads = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            if (repo?.id) {
                return queryLsifUploadsByRepository({ ...args }, repo?.id, apolloClient)
            }
            return queryLsifUploadsList({ ...args }, apolloClient)
        },
        [repo?.id, queryLsifUploadsByRepository, queryLsifUploadsList, apolloClient]
    )

    const commitGraphMetadata = useObservable(
        useMemo(() => (repo ? queryCommitGraphMetadata(repo?.id, apolloClient) : of(undefined)), [
            repo,
            queryCommitGraphMetadata,
            apolloClient,
        ])
    )

    const [deleteStatus, setDeleteStatus] = useState({ isDeleting: false, message: '', state: '' })
    useEffect(() => {
        if (history.location.state) {
            setDeleteStatus({
                isDeleting: true,
                message: history.location.state.message,
                state: history.location.state.modal,
            })
        }
    }, [history.location.state])

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

            {deleteStatus.isDeleting && (
                <Container className="mb-2">
                    <FlashMessage className="mb-0" state={deleteStatus.state} message={deleteStatus.message} />
                </Container>
            )}

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
                        inputClassName="w-auto"
                        noun="upload"
                        pluralNoun="uploads"
                        nodeComponent={CodeIntelUploadNode}
                        nodeComponentProps={{ now }}
                        queryConnection={queryLsifUploads}
                        history={history}
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
