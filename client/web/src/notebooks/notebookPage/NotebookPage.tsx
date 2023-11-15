import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { mdiCheckCircle, mdiBookOutline } from '@mdi/js'
import classNames from 'classnames'
import { useParams } from 'react-router-dom'
import { useStickyBox } from 'react-sticky-box'
import type { Observable } from 'rxjs'
import { catchError, delay, startWith, switchMap } from 'rxjs/operators'

import type { StreamingSearchResultsListProps } from '@sourcegraph/branded'
import { Timestamp } from '@sourcegraph/branded/src/components/Timestamp'
import { asError, isErrorLike } from '@sourcegraph/common'
import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { LoadingSpinner, PageHeader, useEventObservable, useObservable, Alert, Icon } from '@sourcegraph/wildcard'

import type { Block } from '..'
import type { AuthenticatedUser } from '../../auth'
import { PageTitle } from '../../components/PageTitle'
import type { NotebookFields, NotebookInput } from '../../graphql-operations'
import type { OwnConfigProps } from '../../own/OwnConfigProps'
import type { SearchStreamingProps } from '../../search'
import {
    fetchNotebook as _fetchNotebook,
    updateNotebook as _updateNotebook,
    deleteNotebook as _deleteNotebook,
    createNotebookStar as _createNotebookStar,
    deleteNotebookStar as _deleteNotebookStar,
} from '../backend'
import { copyNotebook as _copyNotebook, type CopyNotebookProps } from '../notebook'
import { blockToGQLInput, convertNotebookTitleToFileName, GQLBlockToGQLInput } from '../serialize'

import { NotebookContent } from './NotebookContent'
import { NotebookPageHeaderActions } from './NotebookPageHeaderActions'
import { NotebookTitle } from './NotebookTitle'

import styles from './NotebookPage.module.scss'

interface NotebookPageProps
    extends SearchStreamingProps,
        TelemetryProps,
        TelemetryV2Props,
        Omit<StreamingSearchResultsListProps, 'allExpanded' | 'platformContext' | 'executedQuery'>,
        PlatformContextProps<'sourcegraphURL' | 'requestGraphQL' | 'urlToFile' | 'settings'>,
        OwnConfigProps {
    authenticatedUser: AuthenticatedUser | null
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

export const NotebookPage: React.FunctionComponent<React.PropsWithChildren<NotebookPageProps>> = ({
    fetchNotebook = _fetchNotebook,
    updateNotebook = _updateNotebook,
    deleteNotebook = _deleteNotebook,
    createNotebookStar = _createNotebookStar,
    deleteNotebookStar = _deleteNotebookStar,
    copyNotebook = _copyNotebook,
    streamSearch,
    telemetryService,
    telemetryRecorder,
    searchContextsEnabled,
    ownEnabled,
    isSourcegraphDotCom,
    fetchHighlightedFileLineRanges,
    authenticatedUser,
    settingsCascade,
    platformContext,
}) => {
    const { id: notebookId } = useParams()

    useEffect(() => {
        telemetryService.logPageView('SearchNotebookPage')
        telemetryRecorder.recordEvent('SearchNotebookPage', 'viewed')
    }, [telemetryService, telemetryRecorder])

    const [notebookTitle, setNotebookTitle] = useState('')
    const [updateQueue, setUpdateQueue] = useState<Partial<NotebookInput>[]>([])
    const outlineContainerElement = useRef<HTMLDivElement | null>(null)

    const exportedFileName = useMemo(
        () => `${notebookTitle ? convertNotebookTitleToFileName(notebookTitle) : 'notebook'}.snb.md`,
        [notebookTitle]
    )

    const notebookOrError = useObservable(
        useMemo(
            () =>
                fetchNotebook(notebookId!).pipe(
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
                    switchMap(notebook =>
                        updateNotebook({ id: notebookId!, notebook }).pipe(delay(300), startWith(LOADING))
                    ),
                    catchError(error => [asError(error)])
                ),
            [updateNotebook, notebookId]
        )
    )

    const latestNotebook = useMemo(
        () => updatedNotebookOrError || notebookOrError,
        [notebookOrError, updatedNotebookOrError]
    )

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
        (blocks: Block[]) => setUpdateQueue(queue => queue.concat([{ blocks: blocks.map(blockToGQLInput) }])),
        [setUpdateQueue]
    )

    const onUpdateTitle = useCallback(
        (title: string) => setUpdateQueue(queue => queue.concat([{ title }])),
        [setUpdateQueue]
    )

    const onUpdateVisibility = useCallback(
        (isPublic: boolean, namespace: string) =>
            setUpdateQueue(queue => queue.concat([{ public: isPublic, namespace }])),
        [setUpdateQueue]
    )

    const onCopyNotebook = useCallback(
        (props: Omit<CopyNotebookProps, 'title'>) => copyNotebook({ title: `Copy of ${notebookTitle}`, ...props }),
        [notebookTitle, copyNotebook]
    )

    const stickyBox = useStickyBox()

    return (
        <div className={classNames('w-100', styles.searchNotebookPage)}>
            <PageTitle title={notebookTitle || 'Notebook'} />
            <div className={styles.sideColumn}>
                <div
                    ref={element => {
                        outlineContainerElement.current = element
                        stickyBox(element)
                    }}
                />
            </div>
            <div className={styles.centerColumn}>
                <div className={styles.content}>
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
                                className="mt-2 px-3"
                                actions={
                                    <NotebookPageHeaderActions
                                        isSourcegraphDotCom={isSourcegraphDotCom}
                                        authenticatedUser={authenticatedUser}
                                        notebookId={notebookId!}
                                        viewerCanManage={notebookOrError.viewerCanManage}
                                        isPublic={notebookOrError.public}
                                        namespace={notebookOrError.namespace}
                                        onUpdateVisibility={onUpdateVisibility}
                                        deleteNotebook={deleteNotebook}
                                        starsCount={notebookOrError.stars.totalCount}
                                        viewerHasStarred={notebookOrError.viewerHasStarred}
                                        createNotebookStar={createNotebookStar}
                                        deleteNotebookStar={deleteNotebookStar}
                                        telemetryService={telemetryService}
                                        telemetryRecorder={telemetryRecorder}
                                    />
                                }
                            >
                                <PageHeader.Heading as="h2" styleAs="h1">
                                    <PageHeader.Breadcrumb
                                        icon={mdiBookOutline}
                                        to="/notebooks"
                                        aria-label="Notebooks"
                                    />
                                    <PageHeader.Breadcrumb>
                                        <NotebookTitle
                                            title={notebookOrError.title}
                                            viewerCanManage={notebookOrError.viewerCanManage}
                                            onUpdateTitle={onUpdateTitle}
                                            telemetryService={telemetryService}
                                            telemetryRecorder={telemetryRecorder}
                                        />
                                    </PageHeader.Breadcrumb>
                                </PageHeader.Heading>
                            </PageHeader>
                            <small className="d-flex align-items-center mt-2 px-3">
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
                                            <Icon
                                                aria-hidden={true}
                                                svgPath={mdiCheckCircle}
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
                            <hr className="mt-2 mb-3 mx-3" />
                        </>
                    )}
                    {isNotebookLoaded(notebookOrError) && (
                        <>
                            <NotebookContent
                                viewerCanManage={notebookOrError.viewerCanManage}
                                blocks={notebookOrError.blocks}
                                onUpdateBlocks={onUpdateBlocks}
                                onCopyNotebook={onCopyNotebook}
                                exportedFileName={exportedFileName}
                                streamSearch={streamSearch}
                                telemetryService={telemetryService}
                                telemetryRecorder={telemetryRecorder}
                                searchContextsEnabled={searchContextsEnabled}
                                ownEnabled={ownEnabled}
                                isSourcegraphDotCom={isSourcegraphDotCom}
                                fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                                authenticatedUser={authenticatedUser}
                                settingsCascade={settingsCascade}
                                platformContext={platformContext}
                                outlineContainerElement={outlineContainerElement.current}
                            />
                        </>
                    )}
                </div>
            </div>
        </div>
    )
}
