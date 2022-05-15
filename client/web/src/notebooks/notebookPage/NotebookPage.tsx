import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import classNames from 'classnames'
import BookOutlineIcon from 'mdi-react/BookOutlineIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import { RouteComponentProps } from 'react-router'
import { Observable } from 'rxjs'
import { catchError, delay, startWith, switchMap } from 'rxjs/operators'

import { asError, isErrorLike } from '@sourcegraph/common'
import { StreamingSearchResultsListProps } from '@sourcegraph/search-ui'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import {
    LoadingSpinner,
    PageHeader,
    useEventObservable,
    useObservable,
    Alert,
    ProductStatusBadge,
    Button,
    Icon,
    Typography,
} from '@sourcegraph/wildcard'

import { Block } from '..'
import { AuthenticatedUser } from '../../auth'
import { MarketingBlock } from '../../components/MarketingBlock'
import { PageTitle } from '../../components/PageTitle'
import { Timestamp } from '../../components/time/Timestamp'
import { NotebookFields, NotebookInput, Scalars } from '../../graphql-operations'
import { SearchStreamingProps } from '../../search'
import { NotepadIcon } from '../../search/Notepad'
import { ThemePreference } from '../../stores/themeState'
import { useTheme } from '../../theme'
import {
    fetchNotebook as _fetchNotebook,
    updateNotebook as _updateNotebook,
    deleteNotebook as _deleteNotebook,
    createNotebookStar as _createNotebookStar,
    deleteNotebookStar as _deleteNotebookStar,
} from '../backend'
import { NOTEPAD_ENABLED_EVENT } from '../listPage/NotebooksListPage'
import { copyNotebook as _copyNotebook, CopyNotebookProps } from '../notebook'
import { blockToGQLInput, convertNotebookTitleToFileName, GQLBlockToGQLInput } from '../serialize'

import { NotebookContent } from './NotebookContent'
import { NotebookPageHeaderActions } from './NotebookPageHeaderActions'
import { NotebookTitle } from './NotebookTitle'

import styles from './NotebookPage.module.scss'

interface NotebookPageProps
    extends Pick<RouteComponentProps<{ id: Scalars['ID'] }>, 'match'>,
        SearchStreamingProps,
        ThemeProps,
        TelemetryProps,
        Omit<
            StreamingSearchResultsListProps,
            'allExpanded' | 'extensionsController' | 'platformContext' | 'executedQuery'
        >,
        PlatformContextProps<'sourcegraphURL' | 'requestGraphQL' | 'urlToFile' | 'settings' | 'forceUpdateTooltip'>,
        ExtensionsControllerProps<'extHostAPI' | 'executeCommand'> {
    authenticatedUser: AuthenticatedUser | null
    globbing: boolean
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
    globbing,
    streamSearch,
    isLightTheme,
    telemetryService,
    searchContextsEnabled,
    isSourcegraphDotCom,
    fetchHighlightedFileLineRanges,
    authenticatedUser,
    showSearchContext,
    settingsCascade,
    platformContext,
    extensionsController,
    match,
}) => {
    useEffect(() => telemetryService.logPageView('SearchNotebookPage'), [telemetryService])

    const notebookId = match.params.id
    const [notebookTitle, setNotebookTitle] = useState('')
    const [updateQueue, setUpdateQueue] = useState<Partial<NotebookInput>[]>([])
    const outlineContainerElement = useRef<HTMLDivElement | null>(null)
    const [notepadCTASeen, setNotepadCTASeen] = useTemporarySetting('search.notepad.ctaSeen')
    const [notepadEnabled, setNotepadEnabled] = useTemporarySetting('search.notepad.enabled')

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
                    switchMap(notebook =>
                        updateNotebook({ id: notebookId, notebook }).pipe(delay(300), startWith(LOADING))
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
        (blocks: Block[]) => setUpdateQueue(queue => queue.concat([{ blocks: blocks.map(blockToGQLInput) }])),
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

    const showNotepadCTA = useMemo(
        () =>
            !notepadEnabled &&
            !notepadCTASeen &&
            isNotebookLoaded(latestNotebook) &&
            latestNotebook.blocks.length === 0,
        [latestNotebook, notepadCTASeen, notepadEnabled]
    )

    return (
        <div className={classNames('w-100', styles.searchNotebookPage)}>
            <PageTitle title={notebookTitle || 'Notebook'} />
            <div className={styles.sideColumn} ref={outlineContainerElement} />
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
                                className="mt-2"
                                path={[
                                    { to: '/notebooks', icon: BookOutlineIcon, ariaLabel: 'Notebooks' },
                                    {
                                        text: (
                                            <NotebookTitle
                                                title={notebookOrError.title}
                                                viewerCanManage={notebookOrError.viewerCanManage}
                                                onUpdateTitle={onUpdateTitle}
                                                telemetryService={telemetryService}
                                            />
                                        ),
                                    },
                                ]}
                                actions={
                                    <NotebookPageHeaderActions
                                        isSourcegraphDotCom={isSourcegraphDotCom}
                                        authenticatedUser={authenticatedUser}
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
                                        telemetryService={telemetryService}
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
                                globbing={globbing}
                                streamSearch={streamSearch}
                                isLightTheme={isLightTheme}
                                telemetryService={telemetryService}
                                searchContextsEnabled={searchContextsEnabled}
                                isSourcegraphDotCom={isSourcegraphDotCom}
                                fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges}
                                authenticatedUser={authenticatedUser}
                                showSearchContext={showSearchContext}
                                settingsCascade={settingsCascade}
                                platformContext={platformContext}
                                extensionsController={extensionsController}
                                outlineContainerElement={outlineContainerElement.current}
                            />
                        </>
                    )}
                </div>
                <div className={styles.spacer}>
                    {showNotepadCTA && (
                        <NotepadCTA
                            onEnable={() => {
                                telemetryService.log(NOTEPAD_ENABLED_EVENT)
                                setNotepadCTASeen(true)
                                setNotepadEnabled(true)
                            }}
                            onClose={() => setNotepadCTASeen(true)}
                        />
                    )}
                </div>
            </div>
        </div>
    )
}

interface NotepadCTAProps {
    onEnable: () => void
    onClose: () => void
}

const NotepadCTA: React.FunctionComponent<React.PropsWithChildren<NotepadCTAProps>> = ({ onEnable, onClose }) => {
    const assetsRoot = window.context?.assetsRoot || ''
    const isLightTheme = useTheme().enhancedThemePreference === ThemePreference.Light

    return (
        <MarketingBlock wrapperClassName={styles.notepadCta}>
            <aside className={styles.notepadCtaContent}>
                <Button
                    aria-label="Hide"
                    variant="icon"
                    onClick={onClose}
                    size="sm"
                    className={styles.notepadCtaCloseButton}
                >
                    <Icon as={CloseIcon} />
                </Button>
                <img
                    className="flex-shrink-0 mr-3"
                    src={`${assetsRoot}/img/notepad-illustration-${isLightTheme ? 'light' : 'dark'}.svg`}
                    alt=""
                />
                <div>
                    <Typography.H3 className="d-inline-block">
                        <NotepadIcon /> Enable notepad
                    </Typography.H3>{' '}
                    <ProductStatusBadge status="beta" />
                    <p>
                        The notepad adds a toolbar to the bottom right of search results and file pages to help you
                        create notebooks from your code navigation activities.
                    </p>
                    <p>
                        <Button variant="primary" onClick={onEnable} size="sm">
                            Enable notepad
                        </Button>
                    </p>
                </div>
            </aside>
        </MarketingBlock>
    )
}
