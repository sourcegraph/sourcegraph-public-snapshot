import { FunctionComponent, useCallback, useEffect, useMemo, useState } from 'react'

import { useApolloClient } from '@apollo/client'
import classNames from 'classnames'
import { useLocation } from 'react-router'
import { of, Subject } from 'rxjs'

import { isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Container, PageHeader, useObservable, ErrorAlert } from '@sourcegraph/wildcard'

import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import { DeleteLsifUploadsVariables, LsifUploadFields, LSIFUploadState } from '../../../../graphql-operations'
import { FlashMessage } from '../../configuration/components/FlashMessage'
import { queryCommitGraphMetadata as defaultQueryCommitGraphMetadata } from '../../indexes/hooks/queryCommitGraphMetadata'
import { CodeIntelUploadNode, CodeIntelUploadNodeProps } from '../components/CodeIntelUploadNode'
import { CommitGraphMetadata } from '../components/CommitGraphMetadata'
import { EmptyUploads } from '../components/EmptyUploads'
import { queryLsifUploadsByRepository as defaultQueryLsifUploadsByRepository } from '../hooks/queryLsifUploadsByRepository'
import { queryLsifUploadsList as defaultQueryLsifUploadsList } from '../hooks/queryLsifUploadsList'
import { useDeleteLsifUpload } from '../hooks/useDeleteLsifUpload'
import { useDeleteLsifUploads } from '../hooks/useDeleteLsifUploads'

import styles from './CodeIntelUploadsPage.module.scss'

export interface CodeIntelUploadsPageProps extends TelemetryProps {
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
    ...props
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelUploads'), [telemetryService])
    const location = useLocation<{ message: string; modal: string }>()

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
        useMemo(
            () => (repo ? queryCommitGraphMetadata(repo?.id, apolloClient) : of(undefined)),
            [repo, queryCommitGraphMetadata, apolloClient]
        )
    )

    const [deleteStatus, setDeleteStatus] = useState({ isDeleting: false, message: '', state: '' })
    useEffect(() => {
        if (location.state) {
            setDeleteStatus({
                isDeleting: true,
                message: location.state.message,
                state: location.state.modal,
            })
        }
    }, [location.state])

    const [args, setArgs] = useState<DeleteLsifUploadsVariables>()

    // selection has the same type as CodeIntelUploadNode's prop because there is no CodeIntelUploadNodeProps
    const [selection, setSelection] = useState<CodeIntelUploadNodeProps['selection']>(new Set())
    const onCheckboxToggle = useCallback<CodeIntelUploadNodeProps['onCheckboxToggle']>((id, checked) => {
        setSelection(selection => {
            if (selection === 'all') {
                return selection
            }
            if (checked) {
                selection.add(id)
            } else {
                selection.delete(id)
            }
            return new Set(selection)
        })
    }, [])

    const { handleDeleteLsifUpload, deleteError } = useDeleteLsifUpload()
    const { handleDeleteLsifUploads, deletesError } = useDeleteLsifUploads()

    const deletes = useMemo(() => new Subject<undefined>(), [])

    const queryConnection = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            setArgs({
                query: args.query ?? null,
                // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-explicit-any
                state: (args as any).state ?? null,
                // eslint-disable-next-line @typescript-eslint/no-unsafe-assignment, @typescript-eslint/no-unsafe-member-access, @typescript-eslint/no-explicit-any
                isLatestForRepo: (args as any).isLatestForRepo ?? null,
                repository: repo?.id ?? null,
            })
            setSelection(new Set())
            return queryLsifUploads(args)
        },
        [queryLsifUploads, repo?.id]
    )

    return (
        <div className="code-intel-uploads">
            <PageTitle title="Code graph data uploads" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Code graph data uploads' }]}
                description={`Indexes uploaded to Sourcegraph from CI or from auto-indexing ${
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
                <div className="mb-3">
                    <CommitGraphMetadata
                        stale={commitGraphMetadata.stale}
                        updatedAt={commitGraphMetadata.updatedAt}
                        className="mb-0"
                        now={now}
                    />
                </div>
            )}

            <Container>
                <div className="mb-3">
                    <Button
                        className="mr-2"
                        variant="primary"
                        disabled={selection !== 'all' && selection.size === 0}
                        // eslint-disable-next-line @typescript-eslint/no-misused-promises
                        onClick={async () => {
                            if (selection === 'all') {
                                if (args === undefined) {
                                    return
                                }

                                if (
                                    !confirm(
                                        `Delete all uploads matching the filter criteria?\n\n${Object.entries(args)
                                            .map(([key, value]) => `${key}: ${value}`)
                                            .join('\n')}`
                                    )
                                ) {
                                    return
                                }

                                await handleDeleteLsifUploads({
                                    variables: args,
                                    update: cache => cache.modify({ fields: { node: () => {} } }),
                                })

                                deletes.next()

                                return
                            }

                            for (const id of selection) {
                                await handleDeleteLsifUpload({
                                    variables: { id },
                                    update: cache => cache.modify({ fields: { node: () => {} } }),
                                })
                            }

                            deletes.next()
                        }}
                    >
                        Delete {selection === 'all' ? 'matching' : selection.size === 0 ? '' : selection.size}
                    </Button>
                    <Button
                        variant="secondary"
                        onClick={() => setSelection(selection => (selection === 'all' ? new Set() : 'all'))}
                    >
                        {selection === 'all' ? 'Deselect' : 'Select matching'}
                    </Button>
                </div>

                {isErrorLike(deleteError) && <ErrorAlert prefix="Error deleting LSIF upload" error={deleteError} />}
                {isErrorLike(deletesError) && <ErrorAlert prefix="Error deleting LSIF uploads" error={deletesError} />}

                <div className="list-group position-relative">
                    <FilteredConnection<LsifUploadFields, Omit<CodeIntelUploadNodeProps, 'node'>>
                        listComponent="div"
                        listClassName={classNames(styles.grid, 'mb-3')}
                        inputClassName="flex-1"
                        noun="upload"
                        pluralNoun="uploads"
                        nodeComponent={CodeIntelUploadNode}
                        nodeComponentProps={{ now, selection, onCheckboxToggle }}
                        queryConnection={queryConnection}
                        cursorPaging={true}
                        filters={filters}
                        emptyElement={<EmptyUploads />}
                        updates={deletes}
                    />
                </div>
            </Container>
        </div>
    )
}
