import {
    ContextResolver,
    createHoverifier,
    findPositionsFromEvents,
    Hoverifier,
    HoverState,
    MaybeLoadingResult,
} from '@sourcegraph/codeintellify'
import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import * as H from 'history'
import * as React from 'react'
import { render as reactDOMRender } from 'react-dom'
import {
    animationFrameScheduler,
    combineLatest,
    EMPTY,
    from,
    Observable,
    of,
    Subject,
    Subscription,
    Unsubscribable,
    concat,
    BehaviorSubject,
} from 'rxjs'
import {
    catchError,
    concatAll,
    concatMap,
    filter,
    finalize,
    map,
    mergeMap,
    observeOn,
    switchMap,
    withLatestFrom,
    tap,
    startWith,
    distinctUntilChanged,
    retryWhen,
    mapTo,
} from 'rxjs/operators'
import { ActionItemAction } from '../../../../shared/src/actions/ActionItem'
import { DecorationMapByLine } from '../../../../shared/src/api/client/services/decoration'
import { CodeEditorData, CodeEditorWithPartialModel } from '../../../../shared/src/api/client/services/editorService'
import {
    isPrivateRepoPublicSourcegraphComErrorLike,
    isRepoNotFoundErrorLike,
} from '../../../../shared/src/backend/errors'
import {
    CommandListClassProps,
    CommandListPopoverButtonClassProps,
} from '../../../../shared/src/commandPalette/CommandList'
import { CompletionWidgetClassProps } from '../../../../shared/src/components/completion/CompletionWidget'
import { asObservable } from '../../../../shared/src/util/rxjs/asObservable'
import { ApplyLinkPreviewOptions } from '../../../../shared/src/components/linkPreviews/linkPreviews'
import { Controller } from '../../../../shared/src/extensions/controller'
import { registerHighlightContributions } from '../../../../shared/src/highlight/contributions'
import { getHoverActions, registerHoverContributions } from '../../../../shared/src/hover/actions'
import {
    HoverAlert,
    HoverContext,
    HoverData,
    HoverOverlay,
    HoverOverlayClassProps,
} from '../../../../shared/src/hover/HoverOverlay'
import { getModeFromPath } from '../../../../shared/src/languages'
import { URLToFileContext } from '../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'
import { isDefined, isInstanceOf, property } from '../../../../shared/src/util/types'
import {
    FileSpec,
    UIPositionSpec,
    RawRepoSpec,
    RepoSpec,
    ResolvedRevSpec,
    RevSpec,
    toRootURI,
    toURIWithPath,
    ViewStateSpec,
} from '../../../../shared/src/util/url'
import { observeStorageKey } from '../../browser/storage'
import { isInPage } from '../../context'
import { SourcegraphIntegrationURLs, BrowserPlatformContext } from '../../platform/context'
import { toTextDocumentIdentifier, toTextDocumentPositionParams } from '../../shared/backend/ext-api-conversion'
import { CodeViewToolbar, CodeViewToolbarClassProps } from '../../shared/components/CodeViewToolbar'
import { resolveRev, retryWhenCloneInProgressError } from '../../shared/repo/backend'
import { EventLogger } from '../../shared/tracking/eventLogger'
import { MutationRecordLike, querySelectorOrSelf } from '../../shared/util/dom'
import { featureFlags } from '../../shared/util/featureFlags'
import { bitbucketServerCodeHost } from '../bitbucket/code_intelligence'
import { githubCodeHost } from '../github/code_intelligence'
import { gitlabCodeHost } from '../gitlab/code_intelligence'
import { phabricatorCodeHost } from '../phabricator/code_intelligence'
import { CodeView, fetchFileContents, trackCodeViews } from './code_views'
import { ContentView, handleContentViews } from './content_views'
import { applyDecorations, initializeExtensions, renderCommandPalette, renderGlobalDebug } from './extensions'
import { ViewOnSourcegraphButtonClassProps, ViewOnSourcegraphButton } from './external_links'
import { ExtensionHoverAlertType, getActiveHoverAlerts, onHoverAlertDismissed } from './hover_alerts'
import {
    handleNativeTooltips,
    NativeTooltip,
    nativeTooltipsEnabledFromSettings,
    registerNativeTooltipContributions,
} from './native_tooltips'
import { handleTextFields, TextField } from './text_fields'
import { resolveRepoNames } from './util/file_info'
import { delayUntilIntersecting, ViewResolver } from './views'

import { IS_LIGHT_THEME } from './consts'
import { NotificationType } from 'sourcegraph'
import { failedWithHTTPStatus } from '../../../../shared/src/backend/fetch'
import { asError } from '../../../../shared/src/util/errors'

registerHighlightContributions()

export interface OverlayPosition {
    top: number
    left: number
}

/**
 * A function that gets the mount location for elements being mounted to the DOM.
 *
 * - If the mount doesn't belong into the container, it must return `null`.
 * - If the mount already exists in the container, it must return the existing mount.
 * - If the mount does not exist yet in the container, it must create and return it.
 *
 * Caveats:
 * - The passed element might be the mount itself
 * - The passed element might be an element _within_ the mount
 */
export type MountGetter = (container: HTMLElement) => HTMLElement | null

/**
 * The context the code host is in on the current page.
 */
export type CodeHostContext = RawRepoSpec & Partial<RevSpec> & { privateRepository: boolean }

type CodeHostType = 'github' | 'phabricator' | 'bitbucket-server' | 'gitlab'

/** Information for adding code intelligence to code views on arbitrary code hosts. */
export interface CodeHost extends ApplyLinkPreviewOptions {
    /**
     * The type of the code host. This will be added as a className to the overlay mount.
     * Use {@link CodeHost#name} if you need a human-readable name for the code host to display in the UI.
     */
    type: CodeHostType

    /**
     * A human-readable name for the code host, to be displayed in the UI.
     */
    name: string

    /**
     * Basic contextual information for the current code host.
     */
    getContext?: () => CodeHostContext

    /**
     * Mount getter for the repository "View on Sourcegraph" button.
     *
     * If undefined, the "View on Sourcegraph" button won't be rendered on the code host.
     */
    getViewContextOnSourcegraphMount?: MountGetter

    /**
     * Optional class name for the contextual link to Sourcegraph.
     */
    viewOnSourcegraphButtonClassProps?: ViewOnSourcegraphButtonClassProps

    /**
     * Checks to see if the current context the code is running in is within
     * the given code host.
     */
    check: () => boolean

    /**
     * CSS classes for ActionItem buttons in the hover overlay to customize styling
     */
    hoverOverlayClassProps?: HoverOverlayClassProps

    /**
     * Resolve {@link CodeView}s from the DOM.
     */
    codeViewResolvers: ViewResolver<CodeView>[]

    /**
     * Resolve {@link ContentView}s from the DOM.
     */
    contentViewResolvers?: ViewResolver<ContentView>[]

    /**
     * Resolve {@link TextField}s from the DOM.
     */
    textFieldResolvers?: ViewResolver<TextField>[]

    /**
     * Resolves {@link NativeTooltip}s from the DOM.
     */
    nativeTooltipResolvers?: ViewResolver<NativeTooltip>[]

    /**
     * Adjust the position of the hover overlay. Useful for fixed headers or other
     * elements that throw off the position of the tooltip within the relative
     * element.
     */
    adjustOverlayPosition?: (position: OverlayPosition) => OverlayPosition

    // Extensions related input

    /**
     * Mount getter for the command palette button for extensions.
     *
     * If undefined, the command palette button won't be rendered on the code host.
     */
    getCommandPaletteMount?: MountGetter

    /**
     * Returns a selector used to determine the mount location of the hover overlay in the DOM.
     *
     * If undefined, or when null is returned, the hover overlay container will be mounted to <body>.
     */
    getHoverOverlayMountLocation?: () => string | null

    /**
     * Construct the URL to the specified file.
     *
     * @param sourcegraphURL The URL of the Sourcegraph instance.
     * @param target The target to build a URL for.
     * @param context Context information about this invocation.
     */
    urlToFile?: (
        sourcegraphURL: string,
        target: RepoSpec & RawRepoSpec & RevSpec & FileSpec & Partial<UIPositionSpec> & Partial<ViewStateSpec>,
        context: URLToFileContext
    ) => string

    notificationClassNames: Record<NotificationType, string>

    /**
     * CSS classes for the command palette to customize styling
     */
    commandPaletteClassProps?: CommandListPopoverButtonClassProps & CommandListClassProps

    /**
     * CSS classes for the code view toolbar to customize styling
     */
    codeViewToolbarClassProps?: CodeViewToolbarClassProps

    /**
     * CSS classes for the completion widget to customize styling
     */
    completionWidgetClassProps?: CompletionWidgetClassProps

    /**
     * Whether or not code views need to be tokenized. Defaults to false.
     */
    codeViewsRequireTokenization?: boolean
}

export interface FileInfo {
    /**
     * The path for the repo the file belongs to. If a `baseRepoName` is provided, this value
     * is treated as the head repo name.
     */
    rawRepoName: string
    /**
     * The path for the file path for a given `codeView`. If a `baseFilePath` is provided, this value
     * is treated as the head file path.
     */
    filePath: string
    /**
     * The commit that the code view is at. If a `baseCommitID` is provided, this value is treated
     * as the head commit ID.
     */
    commitID: string
    /**
     * The revision the code view is at. If a `baseRev` is provided, this value is treated as the head rev.
     */
    rev?: string
    /**
     * The repo name for the BASE side of a diff. This is useful for Phabricator
     * staging areas since they are separate repos.
     */
    baseRawRepoName?: string
    /**
     * The base file path.
     */
    baseFilePath?: string
    /**
     * Commit ID for the BASE side of the diff.
     */
    baseCommitID?: string
    /**
     * Revision for the BASE side of the diff.
     */
    baseRev?: string
}

export interface FileInfoWithRepoNames extends FileInfo, RepoSpec {
    baseRepoName?: string
}

export interface CodeIntelligenceProps extends TelemetryProps {
    platformContext: Pick<
        BrowserPlatformContext,
        | 'forceUpdateTooltip'
        | 'urlToFile'
        | 'sideloadedExtensionURL'
        | 'requestGraphQL'
        | 'settings'
        | 'refreshSettings'
    >
    codeHost: CodeHost
    extensionsController: Controller
    showGlobalDebug?: boolean
}

export const createOverlayMount = (codeHostName: string, container: HTMLElement): HTMLElement => {
    const mount = document.createElement('div')
    mount.classList.add('hover-overlay-mount', `hover-overlay-mount__${codeHostName}`, 'theme-light')
    container.appendChild(mount)
    return mount
}

export const createGlobalDebugMount = (): HTMLElement => {
    const mount = document.createElement('div')
    mount.className = 'global-debug'
    document.body.appendChild(mount)
    return mount
}

/**
 * Prepares the page for code intelligence. It creates the hoverifier, injects
 * and mounts the hover overlay and then returns the hoverifier.
 */
function initCodeIntelligence({
    mutations,
    codeHost,
    platformContext,
    extensionsController,
    render,
    telemetryService,
    hoverAlerts,
}: Pick<CodeIntelligenceProps, 'codeHost' | 'platformContext' | 'extensionsController' | 'telemetryService'> & {
    render: typeof reactDOMRender
    hoverAlerts: Observable<HoverAlert<ExtensionHoverAlertType>>[]
    mutations: Observable<MutationRecordLike[]>
}): {
    hoverifier: Hoverifier<
        RepoSpec & RevSpec & FileSpec & ResolvedRevSpec,
        HoverData<ExtensionHoverAlertType>,
        ActionItemAction
    >
    subscription: Unsubscribable
} {
    const subscription = new Subscription()

    /** Emits when the close button was clicked */
    const closeButtonClicks = new Subject<MouseEvent>()
    const nextCloseButtonClick = closeButtonClicks.next.bind(closeButtonClicks)

    /** Emits whenever the ref callback for the hover element is called */
    const hoverOverlayElements = new Subject<HTMLElement | null>()
    const nextOverlayElement = hoverOverlayElements.next.bind(hoverOverlayElements)

    const relativeElement = document.body

    const containerComponentUpdates = new Subject<void>()

    subscription.add(
        registerHoverContributions({ extensionsController, platformContext, history: H.createBrowserHistory() })
    )

    // Code views come and go, but there is always a single hoverifier on the page
    const hoverifier = createHoverifier<
        RepoSpec & RevSpec & FileSpec & ResolvedRevSpec,
        HoverData<ExtensionHoverAlertType>,
        ActionItemAction
    >({
        closeButtonClicks,
        hoverOverlayElements,
        hoverOverlayRerenders: containerComponentUpdates.pipe(
            withLatestFrom(hoverOverlayElements),
            map(([, hoverOverlayElement]) => ({ hoverOverlayElement, relativeElement })),
            filter(property('hoverOverlayElement', isDefined))
        ),
        getHover: ({ line, character, part, ...rest }) =>
            combineLatest([
                extensionsController.services.textDocumentHover.getHover(
                    toTextDocumentPositionParams({ ...rest, position: { line, character } })
                ),
                getActiveHoverAlerts(hoverAlerts),
            ]).pipe(
                map(
                    ([{ isLoading, result: hoverMerged }, alerts]): MaybeLoadingResult<HoverData<
                        ExtensionHoverAlertType
                    > | null> => ({
                        isLoading,
                        result: hoverMerged ? { ...hoverMerged, alerts } : null,
                    })
                )
            ),
        getActions: context => getHoverActions({ extensionsController, platformContext }, context),
        pinningEnabled: true,
        tokenize: codeHost.codeViewsRequireTokenization,
    })

    class HoverOverlayContainer extends React.Component<
        {},
        HoverState<HoverContext, HoverData<ExtensionHoverAlertType>, ActionItemAction>
    > {
        private subscription = new Subscription()
        constructor(props: {}) {
            super(props)
            this.state = hoverifier.hoverState
            this.subscription.add(
                hoverifier.hoverStateUpdates.subscribe(update => {
                    this.setState(update)
                })
            )
        }
        public componentDidMount(): void {
            containerComponentUpdates.next()
        }
        public componentWillUnmount(): void {
            this.subscription.unsubscribe()
        }
        public componentDidUpdate(): void {
            containerComponentUpdates.next()
        }
        public render(): JSX.Element | null {
            const hoverOverlayProps = this.getHoverOverlayProps()
            return hoverOverlayProps ? (
                <HoverOverlay
                    {...hoverOverlayProps}
                    {...codeHost.hoverOverlayClassProps}
                    telemetryService={telemetryService}
                    isLightTheme={IS_LIGHT_THEME}
                    hoverRef={nextOverlayElement}
                    extensionsController={extensionsController}
                    platformContext={platformContext}
                    location={H.createLocation(window.location)}
                    onCloseButtonClick={nextCloseButtonClick}
                    onAlertDismissed={onHoverAlertDismissed}
                />
            ) : null
        }
        private getHoverOverlayProps(): HoverState<
            HoverContext,
            HoverData<ExtensionHoverAlertType>,
            ActionItemAction
        >['hoverOverlayProps'] {
            if (!this.state.hoverOverlayProps) {
                return undefined
            }
            let { overlayPosition, ...rest } = this.state.hoverOverlayProps
            // TODO: is adjustOverlayPosition needed or could it be solved with a better relativeElement?
            if (overlayPosition && codeHost.adjustOverlayPosition) {
                overlayPosition = codeHost.adjustOverlayPosition(overlayPosition)
            }
            return { ...rest, overlayPosition }
        }
    }

    const { getHoverOverlayMountLocation } = codeHost
    if (!getHoverOverlayMountLocation) {
        // This renders to document.body, which we can assume is never removed,
        // so we don't need to subscribe to mutations.
        const overlayMount = createOverlayMount(codeHost.type, document.body)
        render(<HoverOverlayContainer />, overlayMount)
    } else {
        let previousMount: HTMLElement | null = null
        subscription.add(
            observeHoverOverlayMountLocation(getHoverOverlayMountLocation, mutations).subscribe(mountLocation => {
                // Remove the previous mount if it exists,
                // to avoid displaying duplicate hovers.
                if (previousMount) {
                    previousMount.remove()
                }
                const mount = createOverlayMount(codeHost.type, mountLocation)
                previousMount = mount
                render(<HoverOverlayContainer />, mount)
            })
        )
    }

    return { hoverifier, subscription }
}

/**
 * Returns an Observable that emits the element where
 * the hover overlay mount should be appended, taking account
 * mutations and {@link CodeHost#getHoverOverlayMountLocation}.
 *
 * The caller is responsible for removing the previous mount if it exists.
 *
 * This is useful to mount the hover overlay to a different container than document.body,
 * so that it is affected by the visibility changes of that container.
 *
 * Related issue: https://gitlab.com/gitlab-org/gitlab/issues/193433
 *
 * Example use case on GitLab:
 * 1. User visits https://gitlab.com/gitlab-org/gitaly/-/merge_requests/1575. div.tab-pane.diffs doesn't exist yet (it'll be lazy-loaded)
 *      -> Mount the  hover overlay is to document.body.
 * 2. User visits the 'Changes' tab
 *      -> Unmount from document.body, mount to div.tab-pane.diffs
 * 3. User visits the 'Overview' tab again
 *      -> div.tab-pane.diffs is hidden, and as a result so is the hover overlay.
 * 4. User navigates away from the merge request (soft-reload), div.tab-pane.diffs is removed
 *      -> Mount to document.body again
 */
export function observeHoverOverlayMountLocation(
    getMountLocationSelector: NonNullable<CodeHost['getHoverOverlayMountLocation']>,
    mutations: Observable<MutationRecordLike[]>
): Observable<HTMLElement> {
    return mutations.pipe(
        concatAll(),
        map(({ addedNodes, removedNodes }): HTMLElement | null => {
            // If no selector can be used to determine the mount location
            // return document.body as the mount location.
            const selector = getMountLocationSelector()
            if (selector === null) {
                return document.body
            }
            // If any of the added nodes match the selector, return it
            // as the new mount location.
            for (const addedNode of addedNodes) {
                if (!(addedNode instanceof HTMLElement)) {
                    continue
                }
                const mountLocation = querySelectorOrSelf<HTMLElement>(addedNode, selector)
                if (mountLocation) {
                    return mountLocation
                }
            }
            // If any of the removed nodes match the selector,
            // return document.body as the new mount location.
            for (const removedNode of removedNodes) {
                if (!(removedNode instanceof HTMLElement)) {
                    continue
                }
                if (querySelectorOrSelf<HTMLElement>(removedNode, selector)) {
                    return document.body
                }
            }
            // Neither added nodes nor removed nodes match the selector,
            // don't return a new mount location.
            return null
        }),
        filter(isDefined),
        startWith(document.body),
        distinctUntilChanged()
    )
}

export interface HandleCodeHostOptions extends CodeIntelligenceProps {
    mutations: Observable<MutationRecordLike[]>
    sourcegraphURL: string
    render: typeof reactDOMRender
    minimalUI: boolean
}

export function handleCodeHost({
    mutations,
    codeHost,
    extensionsController,
    platformContext,
    showGlobalDebug,
    sourcegraphURL,
    telemetryService,
    render,
    minimalUI,
}: HandleCodeHostOptions): Subscription {
    const history = H.createBrowserHistory()
    const subscriptions = new Subscription()
    const { requestGraphQL } = platformContext

    const addedElements = mutations.pipe(
        concatAll(),
        concatMap(mutation => mutation.addedNodes),
        filter(isInstanceOf(HTMLElement))
    )

    const nativeTooltipsEnabled = codeHost.nativeTooltipResolvers
        ? nativeTooltipsEnabledFromSettings(platformContext.settings)
        : of(false)

    const hoverAlerts: Observable<HoverAlert<ExtensionHoverAlertType>>[] = []

    if (codeHost.nativeTooltipResolvers) {
        const { subscription, nativeTooltipsAlert } = handleNativeTooltips(mutations, nativeTooltipsEnabled, codeHost)
        subscriptions.add(subscription)
        hoverAlerts.push(nativeTooltipsAlert)
        subscriptions.add(registerNativeTooltipContributions(extensionsController))
    }

    const { hoverifier, subscription } = initCodeIntelligence({
        codeHost,
        extensionsController,
        platformContext,
        telemetryService,
        render,
        hoverAlerts,
        mutations,
    })
    subscriptions.add(hoverifier)
    subscriptions.add(subscription)

    // Inject UI components
    // Render command palette
    if (codeHost.getCommandPaletteMount && !minimalUI) {
        subscriptions.add(
            addedElements.pipe(map(codeHost.getCommandPaletteMount), filter(isDefined)).subscribe(
                renderCommandPalette({
                    extensionsController,
                    history,
                    platformContext,
                    telemetryService,
                    render,
                    ...codeHost.commandPaletteClassProps,
                    notificationClassNames: codeHost.notificationClassNames,
                })
            )
        )
    }

    // Render extension debug menu
    // This renders to document.body, which we can assume is never removed,
    // so we don't need to subscribe to mutations.
    if (showGlobalDebug) {
        const mount = createGlobalDebugMount()
        renderGlobalDebug({ extensionsController, platformContext, history, sourcegraphURL, render })(mount)
    }

    const signInCloses = new Subject<void>()
    const nextSignInClose = signInCloses.next.bind(signInCloses)

    // Try to fetch settings and refresh them when a sign in tab was closed
    subscriptions.add(
        concat([null], signInCloses)
            .pipe(
                switchMap(() =>
                    from(platformContext.refreshSettings()).pipe(
                        catchError(error => {
                            console.error('Refreshing settings failed', error)
                            return []
                        })
                    )
                )
            )
            .subscribe()
    )

    /** The number of code views that were detected on the page (not necessarily initialized) */
    const codeViewCount = new BehaviorSubject<number>(0)

    // Render view on Sourcegraph button
    if (codeHost.getViewContextOnSourcegraphMount && codeHost.getContext) {
        const { getContext, viewOnSourcegraphButtonClassProps } = codeHost

        /** Whether or not the repo exists on the configured Sourcegraph instance. */
        const repoExistsOrErrors = signInCloses.pipe(
            startWith(null),
            switchMap(() => {
                const { rawRepoName, rev } = getContext()
                return resolveRev({ repoName: rawRepoName, rev, requestGraphQL }).pipe(
                    retryWhenCloneInProgressError(),
                    mapTo(true),
                    catchError(error => {
                        if (isRepoNotFoundErrorLike(error)) {
                            return [false]
                        }
                        return [asError(error)]
                    }),
                    startWith(undefined)
                )
            })
        )
        const onConfigureSourcegraphClick: React.MouseEventHandler<HTMLAnchorElement> = async event => {
            event.preventDefault()
            await browser.runtime.sendMessage({ type: 'openOptionsPage' })
        }

        subscriptions.add(
            combineLatest([
                repoExistsOrErrors,
                addedElements.pipe(map(codeHost.getViewContextOnSourcegraphMount), filter(isDefined)),
                // Only show sign in button when there is no other code view on the page that is displaying it
                codeViewCount.pipe(
                    map(count => count === 0),
                    distinctUntilChanged()
                ),
            ]).subscribe(([repoExistsOrError, mount, showSignInButton]) => {
                render(
                    <ViewOnSourcegraphButton
                        {...viewOnSourcegraphButtonClassProps}
                        codeHostType={codeHost.type}
                        getContext={getContext}
                        minimalUI={minimalUI}
                        sourcegraphURL={sourcegraphURL}
                        repoExistsOrError={repoExistsOrError}
                        showSignInButton={showSignInButton}
                        onSignInClose={nextSignInClose}
                        onConfigureSourcegraphClick={isInPage ? undefined : onConfigureSourcegraphClick}
                    />,
                    mount
                )
            })
        )
    }

    /** A stream of added or removed code views with the resolved file info */
    const codeViews = mutations.pipe(
        trackCodeViews(codeHost),
        tap(codeViewEvent => {
            codeViewCount.next(codeViewCount.value + 1)
            codeViewEvent.subscriptions.add(() => codeViewCount.next(codeViewCount.value - 1))
        }),
        // Delay emitting code views until they are in the viewport, or within 4000 vertical
        // pixels of the viewport's top or bottom edges.
        delayUntilIntersecting({ rootMargin: '4000px 0px' }),
        mergeMap(codeViewEvent =>
            asObservable(() =>
                codeViewEvent.resolveFileInfo(codeViewEvent.element, platformContext.requestGraphQL)
            ).pipe(
                mergeMap(fileInfo => resolveRepoNames(fileInfo, platformContext.requestGraphQL)),
                mergeMap(fileInfo =>
                    fetchFileContents(fileInfo, platformContext.requestGraphQL).pipe(
                        map(fileInfoWithContents => ({
                            fileInfo: fileInfoWithContents,
                            ...codeViewEvent,
                        }))
                    )
                ),
                catchError(err => {
                    // Ignore PrivateRepoPublicSourcegraph errors (don't initialize those code views)
                    if (isPrivateRepoPublicSourcegraphComErrorLike(err)) {
                        return EMPTY
                    }
                    throw err
                }),
                tap({
                    error: error => {
                        if (codeViewEvent.getToolbarMount) {
                            const mount = codeViewEvent.getToolbarMount(codeViewEvent.element)
                            render(
                                <CodeViewToolbar
                                    {...codeHost.codeViewToolbarClassProps}
                                    fileInfoOrError={error}
                                    sourcegraphURL={sourcegraphURL}
                                    telemetryService={telemetryService}
                                    platformContext={platformContext}
                                    extensionsController={extensionsController}
                                    buttonProps={codeViewEvent.toolbarButtonProps}
                                    onSignInClose={nextSignInClose}
                                    location={H.createLocation(window.location)}
                                />,
                                mount
                            )
                        }
                    },
                }),
                // Retry auth errors after the user closed a sign-in tab
                retryWhen(errors =>
                    errors.pipe(
                        // Don't swallow non-auth errors
                        tap(error => {
                            if (!failedWithHTTPStatus(error, 401)) {
                                throw error
                            }
                        }),
                        switchMap(() => signInCloses)
                    )
                ),
                catchError(error => {
                    // Log errors but don't break the handling of other code views
                    console.error('Could not resolve file info for code view', error)
                    return []
                })
            )
        ),
        observeOn(animationFrameScheduler)
    )

    /** Map from workspace URI to number of editors referencing it */
    const rootRefCounts = new Map<string, number>()

    /**
     * Adds root referenced by a code editor to the worskpace.
     *
     * Will only cause `workspace.roots` to emit if no root with
     * the given `uri` existed.
     */
    const addRootRef = (uri: string, inputRevision: string | undefined): void => {
        rootRefCounts.set(uri, (rootRefCounts.get(uri) || 0) + 1)
        if (rootRefCounts.get(uri) === 1) {
            extensionsController.services.workspace.roots.next([
                ...extensionsController.services.workspace.roots.value,
                { uri, inputRevision },
            ])
        }
    }

    /**
     * Deletes a reference to a workspace root from a code editor.
     *
     * Will only cause `workspace.roots` to emit if the root
     * with the given `uri` has no more references.
     */
    const deleteRootRef = (uri: string): void => {
        const currentRefCount = rootRefCounts.get(uri)
        if (!currentRefCount) {
            throw new Error(`No preexisting root refs for uri ${uri}`)
        }
        const updatedRefCount = currentRefCount - 1
        if (updatedRefCount === 0) {
            extensionsController.services.workspace.roots.next(
                extensionsController.services.workspace.roots.value.filter(root => root.uri !== uri)
            )
        } else {
            rootRefCounts.set(uri, updatedRefCount)
        }
    }

    subscriptions.add(
        codeViews.subscribe(codeViewEvent => {
            console.log('Code view added')
            codeViewEvent.subscriptions.add(() => console.log('Code view removed'))

            const { element, fileInfo, getPositionAdjuster, getToolbarMount, toolbarButtonProps } = codeViewEvent
            const uri = toURIWithPath(fileInfo)
            const languageId = getModeFromPath(fileInfo.filePath)
            const model = { uri, languageId, text: fileInfo.content }
            // Only add the model if it doesn't exist
            // (there may be several code views on the page pointing to the same model)
            if (!extensionsController.services.model.hasModel(uri)) {
                extensionsController.services.model.addModel(model)
            }
            const editorData: CodeEditorData = {
                type: 'CodeEditor' as const,
                resource: uri,
                selections: codeViewEvent.getSelections ? codeViewEvent.getSelections(codeViewEvent.element) : [],
                isActive: true,
            }
            const editorId = extensionsController.services.editor.addEditor(editorData)
            const scope: CodeEditorWithPartialModel = {
                ...editorData,
                ...editorId,
                model,
            }
            const rootURI = toRootURI(fileInfo)
            addRootRef(rootURI, fileInfo.rev)
            codeViewEvent.subscriptions.add(() => {
                deleteRootRef(rootURI)
                extensionsController.services.editor.removeEditor(editorId)
            })

            if (codeViewEvent.observeSelections) {
                codeViewEvent.subscriptions.add(
                    // This nested subscription is necessary, it is managed correctly through `codeViewEvent.subscriptions`
                    // eslint-disable-next-line rxjs/no-nested-subscribe
                    codeViewEvent.observeSelections(codeViewEvent.element).subscribe(selections => {
                        extensionsController.services.editor.setSelections(editorId, selections)
                    })
                )
            }

            // When codeView is a diff (and not an added file), add BASE too.
            if (fileInfo.baseContent && fileInfo.baseRepoName && fileInfo.baseCommitID && fileInfo.baseFilePath) {
                const uri = toURIWithPath({
                    repoName: fileInfo.baseRepoName,
                    commitID: fileInfo.baseCommitID,
                    filePath: fileInfo.baseFilePath,
                })
                // Only add the model if it doesn't exist
                // (there may be several code views on the page pointing to the same model)
                if (!extensionsController.services.model.hasModel(uri)) {
                    extensionsController.services.model.addModel({
                        uri,
                        languageId: getModeFromPath(fileInfo.baseFilePath),
                        text: fileInfo.baseContent,
                    })
                }
                const editor = extensionsController.services.editor.addEditor({
                    type: 'CodeEditor' as const,
                    resource: uri,
                    // There is no notion of a selection on diff views yet, so this is empty.
                    selections: [],
                    isActive: true,
                })
                const baseRootURI = toRootURI({
                    repoName: fileInfo.baseRepoName,
                    commitID: fileInfo.baseCommitID,
                })
                addRootRef(baseRootURI, fileInfo.baseRev)
                codeViewEvent.subscriptions.add(() => {
                    deleteRootRef(baseRootURI)
                    extensionsController.services.editor.removeEditor(editor)
                })
            }

            const domFunctions = {
                ...codeViewEvent.dom,
                // If any parent element has the sourcegraph-extension-element
                // class then that element does not have any code. We
                // must check for "any parent element" because extensions
                // create their DOM changes before the blob is tokenized
                // into multiple elements.
                getCodeElementFromTarget: (target: HTMLElement): HTMLElement | null =>
                    target.closest('.sourcegraph-extension-element') !== null
                        ? null
                        : codeViewEvent.dom.getCodeElementFromTarget(target),
            }

            // Apply decorations coming from extensions
            if (!minimalUI) {
                let decorationsByLine: DecorationMapByLine = new Map()
                const update = (decorations?: TextDocumentDecoration[] | null): void => {
                    try {
                        decorationsByLine = applyDecorations(
                            domFunctions,
                            element,
                            decorations || [],
                            decorationsByLine,
                            fileInfo.baseCommitID ? 'head' : undefined
                        )
                    } catch (err) {
                        console.error('Could not apply head decorations to code view', codeViewEvent.element, err)
                    }
                }
                codeViewEvent.subscriptions.add(
                    extensionsController.services.textDocumentDecoration
                        .getDecorations(toTextDocumentIdentifier(fileInfo))
                        // Make sure extensions get cleaned up un unsubscription
                        .pipe(finalize(update))
                        // The nested subscribe cannot be replaced with a switchMap()
                        // We manage the subscription correctly.
                        // eslint-disable-next-line rxjs/no-nested-subscribe
                        .subscribe(update)
                )
            }
            if (fileInfo.baseCommitID && fileInfo.baseFilePath) {
                let decorationsByLine: DecorationMapByLine = new Map()
                const update = (decorations?: TextDocumentDecoration[] | null): void => {
                    try {
                        decorationsByLine = applyDecorations(
                            domFunctions,
                            element,
                            decorations || [],
                            decorationsByLine,
                            'base'
                        )
                    } catch (err) {
                        console.error('Could not apply base decorations to code view', codeViewEvent.element, err)
                    }
                }
                codeViewEvent.subscriptions.add(
                    extensionsController.services.textDocumentDecoration
                        .getDecorations(
                            toTextDocumentIdentifier({
                                repoName: fileInfo.baseRepoName || fileInfo.repoName, // not sure if all code hosts set baseRepoName
                                commitID: fileInfo.baseCommitID,
                                filePath: fileInfo.baseFilePath,
                            })
                        )
                        // Make sure decorations get cleaned up on unsubscription
                        .pipe(finalize(update))
                        // The nested subscribe cannot be replaced with a switchMap()
                        // We manage the subscription correctly.
                        // eslint-disable-next-line rxjs/no-nested-subscribe
                        .subscribe(update)
                )
            }

            // Add hover code intelligence
            const resolveContext: ContextResolver<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec> = ({ part }) => ({
                repoName: part === 'base' ? fileInfo.baseRepoName || fileInfo.repoName : fileInfo.repoName,
                commitID: part === 'base' ? fileInfo.baseCommitID! : fileInfo.commitID,
                filePath: part === 'base' ? fileInfo.baseFilePath || fileInfo.filePath : fileInfo.filePath,
                rev: part === 'base' ? fileInfo.baseRev || fileInfo.baseCommitID! : fileInfo.rev || fileInfo.commitID,
            })
            const adjustPosition = getPositionAdjuster?.(platformContext.requestGraphQL)
            let hoverSubscription = new Subscription()
            codeViewEvent.subscriptions.add(
                // eslint-disable-next-line rxjs/no-nested-subscribe
                nativeTooltipsEnabled.subscribe(useNativeTooltips => {
                    hoverSubscription.unsubscribe()
                    if (!useNativeTooltips) {
                        hoverSubscription = hoverifier.hoverify({
                            dom: domFunctions,
                            positionEvents: of(element).pipe(
                                findPositionsFromEvents({
                                    domFunctions,
                                    tokenize: codeHost.codeViewsRequireTokenization !== false,
                                })
                            ),
                            resolveContext,
                            adjustPosition,
                            scrollBoundaries: codeViewEvent.getScrollBoundaries
                                ? codeViewEvent.getScrollBoundaries(codeViewEvent.element)
                                : [],
                        })
                    }
                })
            )
            codeViewEvent.subscriptions.add(hoverSubscription)

            element.classList.add('sg-mounted')

            // Render toolbar
            if (getToolbarMount && !minimalUI) {
                const mount = getToolbarMount(element)
                render(
                    <CodeViewToolbar
                        {...codeHost.codeViewToolbarClassProps}
                        fileInfoOrError={fileInfo}
                        sourcegraphURL={sourcegraphURL}
                        telemetryService={telemetryService}
                        platformContext={platformContext}
                        extensionsController={extensionsController}
                        buttonProps={toolbarButtonProps}
                        location={H.createLocation(window.location)}
                        scope={scope}
                        onSignInClose={nextSignInClose}
                    />,
                    mount
                )
            }
        })
    )

    // Show link previews on content views (feature-flagged).
    subscriptions.add(
        handleContentViews(
            from(featureFlags.isEnabled('experimentalLinkPreviews')).pipe(
                switchMap(enabled => (enabled ? mutations : []))
            ),
            { extensionsController },
            codeHost
        )
    )

    // Show completions in text fields (feature-flagged).
    subscriptions.add(
        handleTextFields(
            from(featureFlags.isEnabled('experimentalTextFieldCompletion')).pipe(
                switchMap(enabled => (enabled ? mutations : []))
            ),
            { extensionsController },
            codeHost
        )
    )

    return subscriptions
}

const SHOW_DEBUG = (): boolean => localStorage.getItem('debug') !== null

const CODE_HOSTS: CodeHost[] = [bitbucketServerCodeHost, githubCodeHost, gitlabCodeHost, phabricatorCodeHost]
export const determineCodeHost = (): CodeHost | undefined => CODE_HOSTS.find(codeHost => codeHost.check())

export function injectCodeIntelligenceToCodeHost(
    mutations: Observable<MutationRecordLike[]>,
    codeHost: CodeHost,
    { sourcegraphURL, assetsURL }: SourcegraphIntegrationURLs,
    isExtension: boolean,
    showGlobalDebug = SHOW_DEBUG()
): Subscription {
    const subscriptions = new Subscription()
    const { platformContext, extensionsController } = initializeExtensions(
        codeHost,
        { sourcegraphURL, assetsURL },
        isExtension
    )
    const { requestGraphQL } = platformContext
    const telemetryService = new EventLogger(isExtension, requestGraphQL)
    subscriptions.add(extensionsController)

    let codeHostSubscription: Subscription
    // In the browser extension, observe whether the `disableExtension` storage flag is set.
    // In the native integration, this flag does not exist.
    const extensionDisabled = isExtension ? observeStorageKey('sync', 'disableExtension') : of(false)

    // RFC 68: hide some UI features in the GitLab native integration.
    // This can be overridden using the `sourcegraphMinimalUI` local storage flag.
    const minimalUIStorageFlag = localStorage.getItem('sourcegraphMinimalUI')
    const minimalUI =
        minimalUIStorageFlag !== null ? minimalUIStorageFlag === 'true' : codeHost.type === 'gitlab' && !isExtension
    subscriptions.add(
        extensionDisabled.subscribe(disableExtension => {
            if (disableExtension) {
                // We don't need to unsubscribe if the extension starts with disabled state.
                if (codeHostSubscription) {
                    codeHostSubscription.unsubscribe()
                }
                console.log('Browser extension is disabled')
            } else {
                codeHostSubscription = handleCodeHost({
                    mutations,
                    codeHost,
                    extensionsController,
                    platformContext,
                    showGlobalDebug,
                    sourcegraphURL,
                    telemetryService,
                    render: reactDOMRender,
                    minimalUI,
                })
                subscriptions.add(codeHostSubscription)
                console.log(`${isExtension ? 'Browser extension' : 'Native integration'} is enabled`)
            }
        })
    )
    return subscriptions
}
