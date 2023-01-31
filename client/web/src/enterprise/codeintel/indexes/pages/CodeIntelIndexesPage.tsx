import { FunctionComponent, useCallback, useEffect, useMemo, useState } from 'react'

import { useApolloClient } from '@apollo/client'
import classNames from 'classnames'
import { RouteComponentProps, useLocation } from 'react-router'
import { Subject } from 'rxjs'

import { isErrorLike } from '@sourcegraph/common'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Container, PageHeader, ErrorAlert } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../../auth'
import {
    FilteredConnection,
    FilteredConnectionFilter,
    FilteredConnectionQueryArguments,
} from '../../../../components/FilteredConnection'
import { PageTitle } from '../../../../components/PageTitle'
import { LsifIndexFields, LSIFIndexState } from '../../../../graphql-operations'
import { FlashMessage } from '../../configuration/components/FlashMessage'
import { CodeIntelIndexNode, CodeIntelIndexNodeProps } from '../components/CodeIntelIndexNode'
import { EmptyAutoIndex } from '../components/EmptyAutoIndex'
import { EnqueueForm } from '../components/EnqueueForm'
import { queryLsifIndexList as defaultQueryLsifIndexList } from '../hooks/queryLsifIndexList'
import { queryLsifIndexListByRepository as defaultQueryLsifIndexListByRepository } from '../hooks/queryLsifIndexListByRepository'
import { useDeleteLsifIndex } from '../hooks/useDeleteLsifIndex'
import { useDeleteLsifIndexes } from '../hooks/useDeleteLsifIndexes'
import { useReindexLsifIndex } from '../hooks/useReindexLsifIndex'
import { useReindexLsifIndexes } from '../hooks/useReindexLsifIndexes'

import styles from './CodeIntelIndexesPage.module.scss'

export interface CodeIntelIndexesPageProps extends RouteComponentProps<{}>, TelemetryProps {
    authenticatedUser: AuthenticatedUser | null
    repo?: { id: string }
    queryLsifIndexListByRepository?: typeof defaultQueryLsifIndexListByRepository
    queryLsifIndexList?: typeof defaultQueryLsifIndexList
    now?: () => Date
}

const filters: FilteredConnectionFilter[] = [
    {
        id: 'filters',
        label: 'Index state',
        type: 'select',
        values: [
            {
                label: 'All',
                value: 'all',
                tooltip: 'Show all indexes',
                args: {},
            },
            {
                label: 'Completed',
                value: 'completed',
                tooltip: 'Show completed indexes only',
                args: { state: LSIFIndexState.COMPLETED },
            },
            {
                label: 'Errored',
                value: 'errored',
                tooltip: 'Show errored indexes only',
                args: { state: LSIFIndexState.ERRORED },
            },
            {
                label: 'Processing',
                value: 'processing',
                tooltip: 'Show processing indexes only',
                args: { state: LSIFIndexState.PROCESSING },
            },
            {
                label: 'Queued',
                value: 'queued',
                tooltip: 'Show queued indexes only',
                args: { state: LSIFIndexState.QUEUED },
            },
        ],
    },
]

export const CodeIntelIndexesPage: FunctionComponent<CodeIntelIndexesPageProps> = ({
    authenticatedUser,
    repo,
    queryLsifIndexListByRepository = defaultQueryLsifIndexListByRepository,
    queryLsifIndexList = defaultQueryLsifIndexList,
    now,
    telemetryService,
}) => {
    useEffect(() => telemetryService.logViewEvent('CodeIntelIndexes'), [telemetryService])
    const location = useLocation<{ message: string; modal: string }>()

    const apolloClient = useApolloClient()
    const queryIndexes = useCallback(
        (args: FilteredConnectionQueryArguments) => {
            if (repo?.id) {
                return queryLsifIndexListByRepository(args, repo?.id, apolloClient)
            }

            return queryLsifIndexList(args, apolloClient)
        },
        [repo?.id, queryLsifIndexListByRepository, queryLsifIndexList, apolloClient]
    )

    const querySubject = useMemo(() => new Subject<string>(), [])

    const [args, setArgs] = useState<any>()

    // selection has the same type as CodeIntelUploadNode's prop because there is no CodeIntelUploadNodeProps
    const [selection, setSelection] = useState<CodeIntelIndexNodeProps['selection']>(new Set())
    const onCheckboxToggle = useCallback<CodeIntelIndexNodeProps['onCheckboxToggle']>((id, checked) => {
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

    const { handleDeleteLsifIndex, deleteError } = useDeleteLsifIndex()
    const { handleDeleteLsifIndexes, deletesError } = useDeleteLsifIndexes()
    const { handleReindexLsifIndex, reindexError } = useReindexLsifIndex()
    const { handleReindexLsifIndexes, reindexesError } = useReindexLsifIndexes()

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
            return queryIndexes(args)
        },
        [queryIndexes, repo?.id]
    )

    return (
        <div className="code-intel-indexes">
            <PageTitle title="Auto-indexing jobs" />
            <PageHeader
                headingElement="h2"
                path={[{ text: 'Auto-indexing jobs' }]}
                description={`Auto-indexing jobs ${repo ? 'for this repository' : 'over all repositories'}.`}
                className="mb-3"
            />

            {!!repo && !!authenticatedUser?.siteAdmin && (
                <Container className="mb-2">
                    <EnqueueForm repoId={repo.id} querySubject={querySubject} />
                </Container>
            )}

            {!!location.state && <FlashMessage state={location.state.modal} message={location.state.message} />}

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

                                await handleDeleteLsifIndexes({
                                    variables: args,
                                    update: cache => cache.modify({ fields: { node: () => {} } }),
                                })

                                deletes.next()

                                return
                            }

                            for (const id of selection) {
                                await handleDeleteLsifIndex({
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
                                        `Reindex all uploads matching the filter criteria?\n\n${Object.entries(args)
                                            .map(([key, value]) => `${key}: ${value}`)
                                            .join('\n')}`
                                    )
                                ) {
                                    return
                                }

                                await handleReindexLsifIndexes({
                                    variables: args,
                                    update: cache => cache.modify({ fields: { node: () => {} } }),
                                })

                                return
                            }

                            for (const id of selection) {
                                await handleReindexLsifIndex({
                                    variables: { id },
                                    update: cache => cache.modify({ fields: { node: () => {} } }),
                                })
                            }
                        }}
                    >
                        Reindex {selection === 'all' ? 'matching' : selection.size === 0 ? '' : selection.size}
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
                {isErrorLike(reindexError) && <ErrorAlert prefix="Error reindexing LSIF upload" error={reindexError} />}
                {isErrorLike(reindexesError) && (
                    <ErrorAlert prefix="Error reindexing LSIF uploads" error={reindexesError} />
                )}

                <div className="position-relative">
                    <FilteredConnection<LsifIndexFields, Omit<CodeIntelIndexNodeProps, 'node'>>
                        listComponent="div"
                        inputClassName="flex-1"
                        listClassName={classNames('list-group', styles.grid, 'mb-3')}
                        noun="index"
                        pluralNoun="indexes"
                        querySubject={querySubject}
                        nodeComponent={CodeIntelIndexNode}
                        nodeComponentProps={{ now, selection, onCheckboxToggle }}
                        queryConnection={queryConnection}
                        cursorPaging={true}
                        filters={filters}
                        emptyElement={<EmptyAutoIndex />}
                        updates={deletes}
                    />
                </div>
            </Container>
        </div>
    )
}
