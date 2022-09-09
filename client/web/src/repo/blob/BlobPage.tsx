import React, { useState, useEffect, useCallback, useMemo } from 'react'

import classNames from 'classnames'
import * as H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import { Redirect } from 'react-router'
import { Observable, of } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap } from 'rxjs/operators'
import { Optional } from 'utility-types'

import { ErrorMessage } from '@sourcegraph/branded/src/components/alerts'
import { ErrorLike, isErrorLike, asError } from '@sourcegraph/common'
import {
    useCurrentSpan,
    TraceSpanProvider,
    createActiveSpan,
    reactManualTracer,
} from '@sourcegraph/observability-client'
import { SearchContextProps } from '@sourcegraph/search'
import { StreamingSearchResultsListProps } from '@sourcegraph/search-ui'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { HighlightResponseFormat, Scalars } from '@sourcegraph/shared/src/graphql-operations'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { lazyComponent } from '@sourcegraph/shared/src/util/lazyComponent'
import { RepoFile, ModeSpec, parseQueryAndHash } from '@sourcegraph/shared/src/util/url'
import { Alert, Button, LoadingSpinner, useEventObservable, useObservable } from '@sourcegraph/wildcard'

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
import { useBlameHunks } from '../blame/useBlameHunks'
import { FilePathBreadcrumbs } from '../FilePathBreadcrumbs'
import { HoverThresholdProps } from '../RepoContainer'
import { RepoHeaderContributionsLifecycleProps } from '../RepoHeader'
import { RepoHeaderContributionPortal } from '../RepoHeaderContributionPortal'

import { ToggleHistoryPanel } from './actions/ToggleHistoryPanel'
import { ToggleLineWrap } from './actions/ToggleLineWrap'
import { ToggleRenderedFileMode } from './actions/ToggleRenderedFileMode'
import { getModeFromURL } from './actions/utils'
import { fetchBlob } from './backend'
import { Blob, BlobInfo, BlobProps } from './Blob'
import { Blob as CodeMirrorBlob } from './CodeMirrorBlob'
import { GoToRawAction } from './GoToRawAction'
import { BlobPanel } from './panel/BlobPanel'
import { RenderedFile } from './RenderedFile'

import styles from './BlobPage.module.scss'

const SEARCH_NOTEBOOK_FILE_EXTENSION = '.snb.md'
const RenderedNotebookMarkdown = lazyComponent(() => import('./RenderedNotebookMarkdown'), 'RenderedNotebookMarkdown')

interface BlobPageProps
    extends RepoFile,
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
        Pick<BlobProps, 'onHandleFuzzyFinder'>,
        Pick<SearchContextProps, 'searchContextsEnabled'>,
        Pick<StreamingSearchResultsListProps, 'fetchHighlightedFileLineRanges'> {
    location: H.Location
    history: H.History
    authenticatedUser: AuthenticatedUser | null
    globbing: boolean
    isMacPlatform: boolean
    isSourcegraphDotCom: boolean
    repoID?: Scalars['ID']
    repoUrl?: string
}

/**
 * Blob data including specific properties used in `BlobPage` but not `Blob`
 */
interface BlobPageInfo extends Optional<BlobInfo, 'commitID'> {
    richHTML: string
    aborted: boolean
}

export const BlobPage: React.FunctionComponent<React.PropsWithChildren<BlobPageProps>> = props => {
    const { span } = useCurrentSpan()
    const [wrapCode, setWrapCode] = useState(ToggleLineWrap.getValue())
    let renderMode = getModeFromURL(props.location)
    const { repoName, revision, repoID, commitID, filePath, isLightTheme, useBreadcrumb, mode } = props
    const showSearchNotebook = useExperimentalFeatures(features => features.showSearchNotebook)
    const showSearchContext = useExperimentalFeatures(features => features.showSearchContext ?? false)
    const enableCodeMirror = useExperimentalFeatures(features => features.enableCodeMirrorFileView ?? false)
    const enableLazyBlobSyntaxHighlighting = useExperimentalFeatures(
        features => features.enableLazyBlobSyntaxHighlighting ?? false
    )

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
                        telemetryService={props.telemetryService}
                    />
                ),
            }
        }, [filePath, revision, repoName, props.telemetryService])
    )

    /**
     * Fetches formatted, but un-highlighted, blob content.
     * Intention is to use this whilst we wait for syntax highlighting,
     * so the user has useful content rather than a loading spinner
     */
    const formattedBlobInfoOrError = useObservable(
        useMemo(() => {
            // Note: Lazy syntax highlighting is currently buggy in CodeMirror.
            // GitHub issue to fix: https://github.com/sourcegraph/sourcegraph/issues/41413
            if (!enableLazyBlobSyntaxHighlighting || enableCodeMirror) {
                return of(undefined)
            }

            return createActiveSpan(
                reactManualTracer,
                { name: 'formattedBlobInfoOrError', parentSpan: span },
                fetchSpan =>
                    fetchBlob({ repoName, revision, filePath, format: HighlightResponseFormat.HTML_PLAINTEXT }).pipe(
                        map(blob => {
                            if (blob === null) {
                                return blob
                            }

                            const blobInfo: BlobPageInfo = {
                                content: blob.content,
                                html: blob.highlight.html ?? '',
                                repoName,
                                revision,
                                filePath,
                                mode,
                                // Properties used in `BlobPage` but not `Blob`
                                richHTML: blob.richHTML,
                                aborted: false,
                            }

                            fetchSpan.end()

                            return blobInfo
                        })
                    )
            )
        }, [enableCodeMirror, enableLazyBlobSyntaxHighlighting, filePath, mode, repoName, revision, span])
    )

    // Bundle latest blob with all other file info to pass to `Blob`
    // Prevents https://github.com/sourcegraph/sourcegraph/issues/14965 by not allowing
    // components to use current file props while blob hasn't updated, since all information
    // is bundled in one object whose creation is blocked by `fetchBlob` emission.
    const [nextFetchWithDisabledTimeout, highlightedBlobInfoOrError] = useEventObservable<
        void,
        BlobPageInfo | null | ErrorLike
    >(
        useCallback(
            (clicks: Observable<void>) =>
                clicks.pipe(
                    mapTo(true),
                    startWith(false),
                    switchMap(disableTimeout =>
                        fetchBlob({
                            repoName,
                            revision,
                            filePath,
                            disableTimeout,
                            format: enableCodeMirror
                                ? HighlightResponseFormat.JSON_SCIP
                                : HighlightResponseFormat.HTML_HIGHLIGHT,
                        })
                    ),
                    map(blob => {
                        if (blob === null) {
                            return blob
                        }

                        // Replace html with lsif generated HTML, if available
                        if (!enableCodeMirror && !blob.highlight.html && blob.highlight.lsif) {
                            const html = renderLsifHtml({ lsif: blob.highlight.lsif, content: blob.content })
                            if (html) {
                                blob.highlight.html = html
                            }
                        }

                        const blobInfo: BlobPageInfo = {
                            content: blob.content,
                            html: blob.highlight.html ?? '',
                            lsif: blob.highlight.lsif ?? '',
                            repoName,
                            revision,
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
            [repoName, revision, filePath, mode, enableCodeMirror]
        )
    )

    const blobInfoOrError = enableLazyBlobSyntaxHighlighting
        ? // Fallback to formatted blob whilst we do not have the highlighted blob
          highlightedBlobInfoOrError || formattedBlobInfoOrError
        : highlightedBlobInfoOrError

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

    const blameDecorations = useBlameHunks({ repoName, revision, filePath }, props.platformContext.sourcegraphURL)

    const isSearchNotebook = Boolean(
        blobInfoOrError &&
            !isErrorLike(blobInfoOrError) &&
            blobInfoOrError.filePath.endsWith(SEARCH_NOTEBOOK_FILE_EXTENSION) &&
            showSearchNotebook
    )

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
                {!enableLazyBlobSyntaxHighlighting && (
                    <div className="d-flex mt-3 justify-content-center">
                        <LoadingSpinner />
                    </div>
                )}
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

    const BlobComponent = enableCodeMirror ? CodeMirrorBlob : Blob

    return (
        <>
            {alwaysRender}
            {repoID && commitID && <BlobPanel {...props} repoID={repoID} commitID={commitID} />}
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
                <React.Suspense fallback={<LoadingSpinner />}>
                    <RenderedNotebookMarkdown
                        {...props}
                        markdown={blobInfoOrError.content}
                        onCopyNotebook={onCopyNotebook}
                        showSearchContext={showSearchContext}
                        exportedFileName={basename(blobInfoOrError.filePath)}
                        className={styles.border}
                    />
                </React.Suspense>
            )}
            {!isSearchNotebook && blobInfoOrError.richHTML && renderMode === 'rendered' && (
                <RenderedFile
                    dangerousInnerHTML={blobInfoOrError.richHTML}
                    location={props.location}
                    className={styles.border}
                />
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
            {renderMode === 'code' && commitID && (
                <TraceSpanProvider
                    name="Blob"
                    attributes={{
                        isSearchNotebook,
                        renderMode,
                        enableCodeMirror,
                        enableLazyBlobSyntaxHighlighting,
                    }}
                >
                    <BlobComponent
                        data-testid="repo-blob"
                        className={classNames(styles.blob, styles.border)}
                        blobInfo={{ ...blobInfoOrError, commitID }}
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
                        disableDecorations={false}
                        role="region"
                        ariaLabel="File blob"
                        blameHunks={blameDecorations}
                        onHandleFuzzyFinder={props.onHandleFuzzyFinder}
                    />
                </TraceSpanProvider>
            )}
        </>
    )
}
