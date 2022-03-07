import classNames from 'classnames'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import MagnifyIcon from 'mdi-react/MagnifyIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'
import { catchError, debounceTime, delay, startWith, switchMap } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { StreamingSearchResultsListProps } from '@sourcegraph/search-ui'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Page } from '@sourcegraph/web/src/components/Page'
import { PageTitle } from '@sourcegraph/web/src/components/PageTitle'
import {
    FeedbackBadge,
    LoadingSpinner,
    PageHeader,
    useEventObservable,
    useObservable,
    Alert,
} from '@sourcegraph/wildcard'

import { Block } from '..'
import { AuthenticatedUser } from '../../auth'
import { Timestamp } from '../../components/time/Timestamp'
import { NotebookFields, NotebookInput, Scalars } from '../../graphql-operations'
import { resolveRevision as _resolveRevision, fetchRepository as _fetchRepository } from '../../repo/backend'
import { SearchStreamingProps } from '../../search'
import {
    fetchNotebook as _fetchNotebook,
    updateNotebook as _updateNotebook,
    deleteNotebook as _deleteNotebook,
    createNotebookStar as _createNotebookStar,
    deleteNotebookStar as _deleteNotebookStar,
} from '../backend'
import { copyNotebook as _copyNotebook, CopyNotebookProps } from '../notebook'
import { blockToGQLInput, convertNotebookTitleToFileName, GQLBlockToGQLInput } from '../serialize'

import { NotebookContent } from './NotebookContent'
import styles from './NotebookPage.module.scss'
import { NotebookPageHeaderActions } from './NotebookPageHeaderActions'
import { NotebookTitle } from './NotebookTitle'

interface NotebookPageProps
    extends Pick<RouteComponentProps<{ id: Scalars['ID'] }>, 'match'>,
        SearchStreamingProps,
        ThemeProps,
        TelemetryProps,
        Omit<StreamingSearchResultsListProps, 'allExpanded' | 'extensionsController' | 'platformContext'>,
        PlatformContextProps<'requestGraphQL' | 'urlToFile' | 'settings' | 'forceUpdateTooltip'>,
        ExtensionsControllerProps<'extHostAPI' | 'executeCommand'> {
    authenticatedUser: AuthenticatedUser | null
    globbing: boolean
    isMacPlatform: boolean
    resolveRevision?: typeof _resolveRevision
    fetchRepository?: typeof _fetchRepository
    fetchNotebook?: typeof _fetchNotebook
    updateNotebook?: typeof _updateNotebook
    deleteNotebook?: typeof _deleteNotebook
    createNotebookStar?: typeof _createNotebookStar
    deleteNotebookStar?: typeof _deleteNotebookStar
    copyNotebook?: typeof _copyNotebook
}

const LOADING = 'loading' as const

function isNotebookLoaded(notebook: NotebookFields | Error | typeof LOADING | undefined): notebook is NotebookFields {
    return notebook !== undefined && !isErrorLike(notebook) && notebook !== LOADING
}

export const NotebookPage: React.FunctionComponent<NotebookPageProps> = ({
    fetchRepository = _fetchRepository,
    resolveRevision = _resolveRevision,
    fetchNotebook = _fetchNotebook,
    updateNotebook = _updateNotebook,
    deleteNotebook = _deleteNotebook,
    createNotebookStar = _createNotebookStar,
    deleteNotebookStar = _deleteNotebookStar,
    copyNotebook = _copyNotebook,
    ...props
}) => {
    useEffect(() => props.telemetryService.logViewEvent('SearchNotebookPage'), [props.telemetryService])

    const notebookId = props.match.params.id
    const [notebookTitle, setNotebookTitle] = useState('')
    const [updateQueue, setUpdateQueue] = useState<Partial<NotebookInput>[]>([])

    const exportedFileName = useMemo(
        () => `${notebookTitle ? convertNotebookTitleToFileName(notebookTitle) : 'notebook'}.snb.md`,
        [notebookTitle]
    )

    const notebookOrError = useObservable(
        useMemo(
            () =>
                fetchNotebook(notebookId).pipe(
                    startWith(LOADING),
                    catchError(error => [asError(error)])
                ),
            [fetchNotebook, notebookId]
        )
    )

    const [onUpdateNotebook, updatedNotebookOrError] = useEventObservable(
        useCallback(
            (update: Observable<NotebookInput>) =>
                update.pipe(
                    debounceTime(400),
                    switchMap(notebook =>
                        updateNotebook({ id: notebookId, notebook }).pipe(delay(400), startWith(LOADING))
                    ),
                    catchError(error => [asError(error)])
                ),
            [updateNotebook, notebookId]
        )
    )

    const latestNotebook = useMemo(() => updatedNotebookOrError || notebookOrError, [
        notebookOrError,
        updatedNotebookOrError,
    ])

    useEffect(() => {
        if (isNotebookLoaded(latestNotebook)) {
            setNotebookTitle(latestNotebook.title)
        }
    }, [latestNotebook, setNotebookTitle])

    useEffect(() => {
        // Update the notebook if there are some updates in the queue and the notebook has fully loaded (i.e. there is no update
        // currently in progress).
        if (updateQueue.length > 0 && isNotebookLoaded(latestNotebook) && latestNotebook.namespace) {
            // Aggregate partial updates from the queue into a single update.
            const updateInput = updateQueue.reduce((input, value) => ({ ...input, ...value }))
            // Clear the queue for new updates and save the changes to the backend.
            setUpdateQueue([])
            onUpdateNotebook({
                // Use current notebook state as defaults.
                title: latestNotebook.title,
                blocks: latestNotebook.blocks.map(GQLBlockToGQLInput),
                public: latestNotebook.public,
                namespace: latestNotebook.namespace.id,
                // Apply updates.
                ...updateInput,
            })
        }
    }, [updateQueue, latestNotebook, onUpdateNotebook, setUpdateQueue])

    const onUpdateBlocks = useCallback(
        (blocks: Block[]) =>
            setUpdateQueue(queue =>
                queue.concat([
                    { blocks: blocks.flatMap(block => (block.type === 'compute' ? [] : [blockToGQLInput(block)])) },
                ])
            ),
        [setUpdateQueue]
    )

    const onUpdateTitle = useCallback((title: string) => setUpdateQueue(queue => queue.concat([{ title }])), [
        setUpdateQueue,
    ])

    const onUpdateVisibility = useCallback(
        (isPublic: boolean, namespace: string) =>
            setUpdateQueue(queue => queue.concat([{ public: isPublic, namespace }])),
        [setUpdateQueue]
    )

    const onCopyNotebook = useCallback(
        (props: Omit<CopyNotebookProps, 'title'>) => copyNotebook({ title: `Copy of ${notebookTitle}`, ...props }),
        [notebookTitle, copyNotebook]
    )

    return (
        <div className={classNames('w-100 p-2', styles.searchNotebookPage)}>
            <PageTitle title={notebookTitle || 'Notebook'} />
            <Page>
                {isErrorLike(notebookOrError) && (
                    <Alert variant="danger">
                        Error while loading the notebook: <strong>{notebookOrError.message}</strong>
                    </Alert>
                )}
                {isErrorLike(updatedNotebookOrError) && (
                    <Alert variant="danger">
                        Error while updating the notebook: <strong>{updatedNotebookOrError.message}</strong>
                    </Alert>
                )}
                {notebookOrError === LOADING && (
                    <div className="d-flex justify-content-center">
                        <LoadingSpinner />
                    </div>
                )}
                {isNotebookLoaded(notebookOrError) && (
                    <>
                        <PageHeader
                            annotation={
                                <FeedbackBadge status="beta" feedback={{ mailto: 'support@sourcegraph.com' }} />
                            }
                            path={[
                                { icon: MagnifyIcon, to: '/search' },
                                { to: '/notebooks', text: 'Notebooks' },
                                {
                                    text: (
                                        <NotebookTitle
                                            title={notebookOrError.title}
                                            viewerCanManage={notebookOrError.viewerCanManage}
                                            onUpdateTitle={onUpdateTitle}
                                            telemetryService={props.telemetryService}
                                        />
                                    ),
                                },
                            ]}
                            actions={
                                <NotebookPageHeaderActions
                                    isSourcegraphDotCom={props.isSourcegraphDotCom}
                                    authenticatedUser={props.authenticatedUser}
                                    notebookId={notebookId}
                                    viewerCanManage={notebookOrError.viewerCanManage}
                                    isPublic={notebookOrError.public}
                                    namespace={notebookOrError.namespace}
                                    onUpdateVisibility={onUpdateVisibility}
                                    deleteNotebook={deleteNotebook}
                                    starsCount={notebookOrError.stars.totalCount}
                                    viewerHasStarred={notebookOrError.viewerHasStarred}
                                    createNotebookStar={createNotebookStar}
                                    deleteNotebookStar={deleteNotebookStar}
                                    telemetryService={props.telemetryService}
                                />
                            }
                        />
                        <small className="d-flex align-items-center mt-2">
                            <div className="mr-2">
                                Created{' '}
                                {notebookOrError.creator && (
                                    <span>
                                        by <strong>@{notebookOrError.creator.username}</strong>
                                    </span>
                                )}{' '}
                                <Timestamp date={notebookOrError.createdAt} />
                            </div>
                            <div className="d-flex align-items-center">
                                {latestNotebook === LOADING && (
                                    <>
                                        <LoadingSpinner className={classNames('m-1', styles.autoSaveIndicator)} />{' '}
                                        Autosaving notebook...
                                    </>
                                )}
                                {isNotebookLoaded(latestNotebook) && (
                                    <>
                                        <CheckCircleIcon
                                            className={classNames('text-success m-1', styles.autoSaveIndicator)}
                                        />
                                        <span>
                                            Last updated{' '}
                                            {latestNotebook.updater && (
                                                <span>
                                                    by <strong>@{latestNotebook.updater.username}</strong>
                                                </span>
                                            )}
                                            &nbsp;
                                        </span>
                                        <Timestamp date={latestNotebook.updatedAt} />
                                    </>
                                )}
                            </div>
                        </small>
                        <hr className="mt-2 mb-3" />
                        <NotebookContent
                            {...props}
                            viewerCanManage={notebookOrError.viewerCanManage}
                            blocks={notebookOrError.blocks}
                            onUpdateBlocks={onUpdateBlocks}
                            fetchRepository={fetchRepository}
                            resolveRevision={resolveRevision}
                            onCopyNotebook={onCopyNotebook}
                            exportedFileName={exportedFileName}
                        />
                        <div className={styles.spacer} />
                    </>
                )}
            </Page>
        </div>
    )
}
