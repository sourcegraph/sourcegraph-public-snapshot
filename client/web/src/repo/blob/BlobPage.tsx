import React, { useState, useEffect, useCallback, useMemo } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Redirect } from 'react-router'
import { Observable } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap } from 'rxjs/operators'

import { ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { ErrorLike, isErrorLike, asError } from '@sourcegraph/common'
import { SearchContextProps } from '@sourcegraph/search'
import { StreamingSearchResultsListProps } from '@sourcegraph/search-ui'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { AbsoluteRepoFile, ModeSpec, parseQueryAndHash } from '@sourcegraph/shared/src/util/url'
import { Alert, Button, LoadingSpinner, useEventObservable } from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../auth'
import { BreadcrumbSetters } from '../../components/Breadcrumbs'
import { HeroPage } from '../../components/HeroPage'
import { PageTitle } from '../../components/PageTitle'
import { render as renderLsifHtml } from '../../lsif/html'
import { copyNotebook, CopyNotebookProps } from '../../notebooks/notebook'
import { SearchStreamingProps } from '../../search'
import { useNotepad, useExperimentalFeatures } from '../../stores'
import { basename } from '../../util/path'
import { toTreeURL } from '../../util/url'
import { FilePathBreadcrumbs } from '../FilePathBreadcrumbs'
import { HoverThresholdProps } from '../RepoContainer'
import { RepoHeaderContributionsLifecycleProps } from '../RepoHeader'
import { RepoHeaderContributionPortal } from '../RepoHeaderContributionPortal'

import { ToggleHistoryPanel } from './actions/ToggleHistoryPanel'
import { ToggleLineWrap } from './actions/ToggleLineWrap'
import { ToggleRenderedFileMode } from './actions/ToggleRenderedFileMode'
import { getModeFromURL } from './actions/utils'
import { fetchBlob } from './backend'
import { Blob, BlobInfo } from './Blob'
import { GoToRawAction } from './GoToRawAction'
import { useBlobPanelViews } from './panel/BlobPanel'
import { RenderedFile } from './RenderedFile'
import { RenderedNotebookMarkdown, SEARCH_NOTEBOOK_FILE_EXTENSION } from './RenderedNotebookMarkdown'

import styles from './BlobPage.module.scss'

interface Props
    extends AbsoluteRepoFile,
        ModeSpec,
        RepoHeaderContributionsLifecycleProps,
        SettingsCascadeProps,
        PlatformContextProps,
        TelemetryProps,
        ExtensionsControllerProps,
        ThemeProps,
        HoverThresholdProps,
        BreadcrumbSetters,
        SearchStreamingProps,
        Pick<SearchContextProps, 'searchContextsEnabled'>,
        Pick<StreamingSearchResultsListProps, 'fetchHighlightedFileLineRanges'> {
    location: H.Location
    history: H.History
    repoID: Scalars['ID']
    authenticatedUser: AuthenticatedUser | null
    globbing: boolean
    isMacPlatform: boolean
    isSourcegraphDotCom: boolean
    repoUrl: string
}

export const BlobPage: React.FunctionComponent<React.PropsWithChildren<Props>> = props => {
    const [wrapCode, setWrapCode] = useState(ToggleLineWrap.getValue())
    let renderMode = getModeFromURL(props.location)
    const { repoName, revision, commitID, filePath, isLightTheme, useBreadcrumb, mode, repoUrl } = props
    const showSearchNotebook = useExperimentalFeatures(features => features.showSearchNotebook)
    const showSearchContext = useExperimentalFeatures(features => features.showSearchContext ?? false)
    const lineOrRange = useMemo(() => parseQueryAndHash(props.location.search, props.location.hash), [
        props.location.search,
        props.location.hash,
    ])

    // Log view event whenever a new Blob, or a Blob with a different render mode, is visited.
    useEffect(() => {
        props.telemetryService.logViewEvent('Blob', { repoName, filePath })
    }, [repoName, commitID, filePath, renderMode, props.telemetryService])
    useNotepad(
        useMemo(
            () => ({
                type: 'file',
                path: filePath,
                repo: repoName,
                revision,
                // Need to subtract 1 because IHighlightLineRange is 0-based but
                // line information in the URL is 1-based.
                lineRange: lineOrRange.line
                    ? { startLine: lineOrRange.line - 1, endLine: (lineOrRange.endLine ?? lineOrRange.line) - 1 }
                    : null,
            }),
            [filePath, repoName, revision, lineOrRange.line, lineOrRange.endLine]
        )
    )

    useBreadcrumb(
        useMemo(() => {
            if (!filePath) {
                return
            }

            return {
                key: 'filePath',
                className: 'flex-shrink-past-contents',
                element: (
                    // TODO should these be "flattened" all using setBreadcrumb()?
                    <FilePathBreadcrumbs
                        key="path"
                        repoName={repoName}
                        revision={revision}
                        filePath={filePath}
                        isDir={false}
                        repoUrl={repoUrl}
                        telemetryService={props.telemetryService}
                    />
                ),
            }
        }, [filePath, revision, repoName, repoUrl, props.telemetryService])
    )

    // Bundle latest blob with all other file info to pass to `Blob`
    // Prevents https://github.com/sourcegraph/sourcegraph/issues/14965 by not allowing
    // components to use current file props while blob hasn't updated, since all information
    // is bundled in one object whose creation is blocked by `fetchBlob` emission.
    const [nextFetchWithDisabledTimeout, blobInfoOrError] = useEventObservable<
        void,
        (BlobInfo & { richHTML: string; aborted: boolean }) | null | ErrorLike
    >(
        useCallback(
            (clicks: Observable<void>) =>
                clicks.pipe(
                    mapTo(true),
                    startWith(false),
                    switchMap(disableTimeout =>
                        fetchBlob({
                            repoName,
                            commitID,
                            filePath,
                            disableTimeout,
                        })
                    ),
                    map(blob => {
                        if (blob === null) {
                            return blob
                        }

                        // Replace html with lsif generated HTML, if available
                        if (blob.highlight.lsif && blob.highlight.lsif !== '{}') {
                            blob.highlight.html = renderLsifHtml(blob.highlight.lsif, blob.content)
                        }

                        const blobInfo: BlobInfo & { richHTML: string; aborted: boolean } = {
                            content: blob.content,
                            html: blob.highlight.html,
                            repoName,
                            revision,
                            commitID,
                            filePath,
                            mode,
                            // Properties used in `BlobPage` but not `Blob`
                            richHTML: blob.richHTML,
                            aborted: blob.highlight.aborted,
                        }
                        return blobInfo
                    }),
                    catchError((error): [ErrorLike] => [asError(error)])
                ),
            [repoName, revision, commitID, filePath, mode]
        )
    )

    const onExtendTimeoutClick = useCallback(
        (event: React.MouseEvent): void => {
            event.preventDefault()
            nextFetchWithDisabledTimeout()
        },
        [nextFetchWithDisabledTimeout]
    )

    const getPageTitle = (): string => {
        const repoNameSplit = repoName.split('/')
        const repoString = repoNameSplit.length > 2 ? repoNameSplit.slice(1).join('/') : repoName
        if (filePath) {
            const fileOrDirectory = filePath.split('/').pop()!
            return `${fileOrDirectory} - ${repoString}`
        }
        return `${repoString}`
    }

    useBlobPanelViews(props)

    const isSearchNotebook =
        blobInfoOrError &&
        !isErrorLike(blobInfoOrError) &&
        blobInfoOrError.filePath.endsWith(SEARCH_NOTEBOOK_FILE_EXTENSION) &&
        showSearchNotebook

    const onCopyNotebook = useCallback(
        (props: Omit<CopyNotebookProps, 'title'>) => {
            const title =
                blobInfoOrError && !isErrorLike(blobInfoOrError) ? basename(blobInfoOrError.filePath) : 'Notebook'
            return copyNotebook({ title: `Copy of ${title}`, ...props })
        },
        [blobInfoOrError]
    )

    // If url explicitly asks for a certain rendering mode, renderMode is set to that mode, else it checks:
    // - If file contains richHTML and url does not include a line number: We render in richHTML.
    // - If file does not contain richHTML or the url includes a line number: We render in code view.
    if (!renderMode) {
        renderMode =
            blobInfoOrError && !isErrorLike(blobInfoOrError) && blobInfoOrError.richHTML && !lineOrRange.line
                ? 'rendered'
                : 'code'
    }

    // Always render these to avoid UI jitter during loading when switching to a new file.
    const alwaysRender = (
        <>
            <PageTitle title={getPageTitle()} />
            <RepoHeaderContributionPortal
                position="right"
                priority={20}
                id="toggle-blob-panel"
                repoHeaderContributionsLifecycleProps={props.repoHeaderContributionsLifecycleProps}
            >
                {context => (
                    <ToggleHistoryPanel
                        {...context}
                        key="toggle-blob-panel"
                        location={props.location}
                        history={props.history}
                    />
                )}
            </RepoHeaderContributionPortal>
            {renderMode === 'code' && (
                <RepoHeaderContributionPortal
                    position="right"
                    priority={99}
                    id="toggle-line-wrap"
                    repoHeaderContributionsLifecycleProps={props.repoHeaderContributionsLifecycleProps}
                >
                    {context => <ToggleLineWrap {...context} key="toggle-line-wrap" onDidUpdate={setWrapCode} />}
                </RepoHeaderContributionPortal>
            )}
            <RepoHeaderContributionPortal
                position="right"
                priority={30}
                id="raw-action"
                repoHeaderContributionsLifecycleProps={props.repoHeaderContributionsLifecycleProps}
            >
                {context => (
                    <GoToRawAction
                        {...context}
                        telemetryService={props.telemetryService}
                        key="raw-action"
                        repoName={repoName}
                        revision={props.revision}
                        filePath={filePath}
                    />
                )}
            </RepoHeaderContributionPortal>
        </>
    )

    if (isErrorLike(blobInfoOrError)) {
        // Be helpful if the URL was actually a tree and redirect.
        // Some extensions may optimistically construct blob URLs because
        // they cannot easily determine eagerly if a file path is a tree or a blob.
        // We don't have error names on GraphQL errors.
        if (/not a blob/i.test(blobInfoOrError.message)) {
            return <Redirect to={toTreeURL(props)} />
        }
        return (
            <>
                {alwaysRender}
                <HeroPage icon={AlertCircleIcon} title="Error" subtitle={<ErrorMessage error={blobInfoOrError} />} />
            </>
        )
    }

    if (blobInfoOrError === undefined) {
        // Render placeholder for layout before content is fetched.
        return (
            <div className={styles.placeholder}>
                {alwaysRender}
                <div className="d-flex mt-3 justify-content-center">
                    <LoadingSpinner />
                </div>
            </div>
        )
    }

    // File not found:
    if (blobInfoOrError === null) {
        return (
            <div className={styles.placeholder}>
                <HeroPage
                    icon={MapSearchIcon}
                    title="Not found"
                    subtitle={`${filePath} does not exist at this revision.`}
                />
            </div>
        )
    }

    return (
        <>
            {alwaysRender}
            {blobInfoOrError.richHTML && (
                <RepoHeaderContributionPortal
                    position="right"
                    priority={100}
                    id="toggle-rendered-file-mode"
                    repoHeaderContributionsLifecycleProps={props.repoHeaderContributionsLifecycleProps}
                >
                    {({ actionType }) => (
                        <ToggleRenderedFileMode
                            key="toggle-rendered-file-mode"
                            mode={renderMode || 'rendered'}
                            actionType={actionType}
                        />
                    )}
                </RepoHeaderContributionPortal>
            )}
            {isSearchNotebook && renderMode === 'rendered' && (
                <RenderedNotebookMarkdown
                    {...props}
                    markdown={blobInfoOrError.content}
                    onCopyNotebook={onCopyNotebook}
                    showSearchContext={showSearchContext}
                    exportedFileName={basename(blobInfoOrError.filePath)}
                />
            )}
            {!isSearchNotebook && blobInfoOrError.richHTML && renderMode === 'rendered' && (
                <RenderedFile dangerousInnerHTML={blobInfoOrError.richHTML} location={props.location} />
            )}
            {!blobInfoOrError.richHTML && blobInfoOrError.aborted && (
                <div>
                    <Alert variant="info">
                        Syntax-highlighting this file took too long. &nbsp;
                        <Button onClick={onExtendTimeoutClick} variant="primary" size="sm">
                            Try again
                        </Button>
                    </Alert>
                </div>
            )}
            {/* Render the (unhighlighted) blob also in the case highlighting timed out */}
            {renderMode === 'code' && (
                <Blob
                    className={classNames('test-repo-blob', styles.blob)}
                    blobInfo={blobInfoOrError}
                    wrapCode={wrapCode}
                    platformContext={props.platformContext}
                    extensionsController={props.extensionsController}
                    settingsCascade={props.settingsCascade}
                    onHoverShown={props.onHoverShown}
                    history={props.history}
                    isLightTheme={isLightTheme}
                    telemetryService={props.telemetryService}
                    location={props.location}
                    disableStatusBar={false}
                />
            )}
        </>
    )
}
