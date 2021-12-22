import classNames from 'classnames'
import * as H from 'history'
import * as React from 'react'
import { render as reactDOMRender } from 'react-dom'
import {
    asyncScheduler,
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
    fromEvent,
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
    take,
} from 'rxjs/operators'
import { NotificationType, HoverAlert } from 'sourcegraph'

import { TextDocumentDecoration, WorkspaceRoot } from '@sourcegraph/extension-api-types'
import { ActionItemAction, urlForClientCommandOpen } from '@sourcegraph/shared/src/actions/ActionItem'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { HoverMerged } from '@sourcegraph/shared/src/api/client/types/hover'
import { DecorationMapByLine } from '@sourcegraph/shared/src/api/extension/api/decorations'
import { CodeEditorData, CodeEditorWithPartialModel } from '@sourcegraph/shared/src/api/viewerTypes'
import { isRepoNotFoundErrorLike } from '@sourcegraph/shared/src/backend/errors'
import { isHTTPAuthError } from '@sourcegraph/shared/src/backend/fetch'
import {
    ContextResolver,
    createHoverifier,
    findPositionsFromEvents,
    Hoverifier,
    HoverState,
    MaybeLoadingResult,
} from '@sourcegraph/shared/src/codeintellify'
import { DiffPart } from '@sourcegraph/shared/src/codeintellify/tokenPosition'
import {
    CommandListClassProps,
    CommandListPopoverButtonClassProps,
} from '@sourcegraph/shared/src/commandPalette/CommandList'
import { ApplyLinkPreviewOptions } from '@sourcegraph/shared/src/components/linkPreviews/linkPreviews'
import { Controller } from '@sourcegraph/shared/src/extensions/controller'
import { registerHighlightContributions } from '@sourcegraph/shared/src/highlight/contributions'
import { getHoverActions, registerHoverContributions } from '@sourcegraph/shared/src/hover/actions'
import { HoverContext, HoverOverlay, HoverOverlayClassProps } from '@sourcegraph/shared/src/hover/HoverOverlay'
import { getModeFromPath } from '@sourcegraph/shared/src/languages'
import { URLToFileContext } from '@sourcegraph/shared/src/platform/context'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { isFirefox } from '@sourcegraph/shared/src/util/browserDetection'
import { asError } from '@sourcegraph/shared/src/util/errors'
import { asObservable } from '@sourcegraph/shared/src/util/rxjs/asObservable'
import { isDefined, isInstanceOf, property } from '@sourcegraph/shared/src/util/types'
import {
    FileSpec,
    UIPositionSpec,
    RawRepoSpec,
    RepoSpec,
    ResolvedRevisionSpec,
    RevisionSpec,
    toRootURI,
    toURIWithPath,
    ViewStateSpec,
} from '@sourcegraph/shared/src/util/url'

import { background } from '../../../browser-extension/web-extension-api/runtime'
import { observeStorageKey } from '../../../browser-extension/web-extension-api/storage'
import { BackgroundPageApi } from '../../../browser-extension/web-extension-api/types'
import { toTextDocumentPositionParameters } from '../../backend/extension-api-conversion'
import { CodeViewToolbar, CodeViewToolbarClassProps } from '../../components/CodeViewToolbar'
import { isExtension, isInPage } from '../../context'
import { SourcegraphIntegrationURLs, BrowserPlatformContext } from '../../platform/context'
import { resolveRevision, retryWhenCloneInProgressError, resolvePrivateRepo } from '../../repo/backend'
import { EventLogger, ConditionalTelemetryService } from '../../tracking/eventLogger'
import {
    DEFAULT_SOURCEGRAPH_URL,
    getPlatformName,
    isDefaultSourcegraphUrl,
    observeSourcegraphURL,
} from '../../util/context'
import { MutationRecordLike, querySelectorOrSelf } from '../../util/dom'
import { featureFlags } from '../../util/featureFlags'
import { shouldOverrideSendTelemetry, observeOptionFlag } from '../../util/optionFlags'
import { bitbucketCloudCodeHost } from '../bitbucket-cloud/codeHost'
import { bitbucketServerCodeHost } from '../bitbucket/codeHost'
import { gerritCodeHost } from '../gerrit/codeHost'
import { githubCodeHost } from '../github/codeHost'
import { gitlabCodeHost } from '../gitlab/codeHost'
import { phabricatorCodeHost } from '../phabricator/codeHost'

import styles from './codeHost.module.scss'
import { CodeView, trackCodeViews, fetchFileContentForDiffOrFileInfo } from './codeViews'
import { ContentView, handleContentViews } from './contentViews'
import { NotAuthenticatedError, RepoURLParseError } from './errors'
import { applyDecorations, initializeExtensions, renderCommandPalette, renderGlobalDebug } from './extensions'
import { createPrivateCodeHoverAlert, getActiveHoverAlerts, onHoverAlertDismissed } from './hoverAlerts'
import {
    handleNativeTooltips,
    NativeTooltip,
    nativeTooltipsEnabledFromSettings,
    registerNativeTooltipContributions,
} from './nativeTooltips'
import { SignInButton } from './SignInButton'
import { resolveRepoNamesForDiffOrFileInfo, defaultRevisionToCommitID } from './util/fileInfo'
import {
    ViewOnSourcegraphButtonClassProps,
    ViewOnSourcegraphButton,
    ConfigureSourcegraphButton,
} from './ViewOnSourcegraphButton'
import { delayUntilIntersecting, trackViews, ViewResolver } from './views'

registerHighlightContributions()

export interface OverlayPosition {
    top: number
    left: number
}

export type ObserveMutations = (
    target: Node,
    options?: MutationObserverInit,
    paused?: Subject<boolean>
) => Observable<MutationRecordLike[]>

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
export type CodeHostContext = RawRepoSpec & Partial<RevisionSpec> & { privateRepository: boolean }

export type CodeHostType = 'github' | 'phabricator' | 'bitbucket-server' | 'bitbucket-cloud' | 'gitlab' | 'gerrit'

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
    getContext?: () => Promise<CodeHostContext>

    /**
     * An Observable for whether the code host is in light theme (vs dark theme).
     * Defaults to always light theme.
     */
    isLightTheme?: Observable<boolean>

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
     * Configuration for built-in search input enhancement
     */
    searchEnhancement?: {
        /** Search input element resolver */
        searchViewResolver: ViewResolver<{ element: HTMLElement }>
        /** Search result element resolver */
        resultViewResolver: ViewResolver<{ element: HTMLElement }>
        /** Callback to trigger on input element change */
        onChange: (args: { value: string; searchURL: string; resultElement: HTMLElement }) => void
    }

    /**
     * Resolve {@link ContentView}s from the DOM.
     */
    contentViewResolvers?: ViewResolver<ContentView>[]

    /**
     * Resolves {@link NativeTooltip}s from the DOM.
     */
    nativeTooltipResolvers?: ViewResolver<NativeTooltip>[]

    /**
     * Override of `observeMutations`, used where a MutationObserve is not viable, such as in the shadow DOMs in Gerrit.
     */
    observeMutations?: ObserveMutations

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
        target: RepoSpec & RawRepoSpec & RevisionSpec & FileSpec & Partial<UIPositionSpec> & Partial<ViewStateSpec>,
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
     * Whether or not code views need to be tokenized. Defaults to false.
     */
    codeViewsRequireTokenization?: boolean
}

/**
 * A blob (single file `FileInfo`) or a diff (with a head `FileInfo` and/or base `FileInfo`)
 */
export type DiffOrBlobInfo<T extends FileInfo = FileInfo> = BlobInfo<T> | DiffInfo<T>
export interface BlobInfo<T extends FileInfo = FileInfo> {
    blob: T
}
export type DiffInfo<T extends FileInfo = FileInfo> =
    // `base?: undefined` avoids making `{ head: T; base: T }` assignable to this type
    { head: T; base?: undefined } | { base: T; head?: undefined } | { head: T; base: T }

export interface FileInfo {
    /**
     * The path for the repo the file belongs to.
     */
    rawRepoName: string
    /**
     * The path for the file path for a given `codeView`.
     */
    filePath: string
    /**
     * The commit that the code view is at.
     */
    commitID: string
    /**
     * The revision the code view is at.
     */
    revision?: string
}

export interface FileInfoWithRepoName extends FileInfo, RepoSpec {}
export interface FileInfoWithContent extends FileInfoWithRepoName {
    content?: string
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
        | 'sourcegraphURL'
    >
    codeHost: CodeHost
    extensionsController: Controller
    showGlobalDebug?: boolean
}

export const createOverlayMount = (codeHostName: string, container: HTMLElement): HTMLElement => {
    const mount = document.createElement('div')
    mount.classList.add('hover-overlay-mount', `hover-overlay-mount__${codeHostName}`)
    container.append(mount)
    return mount
}

export const createGlobalDebugMount = (): HTMLElement => {
    const mount = document.createElement('div')
    mount.dataset.globalDebug = 'true'
    document.body.append(mount)
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
    privateCloudErrors,
}: Pick<CodeIntelligenceProps, 'codeHost' | 'platformContext' | 'extensionsController' | 'telemetryService'> & {
    render: typeof reactDOMRender
    hoverAlerts: Observable<HoverAlert>[]
    mutations: Observable<MutationRecordLike[]>
    privateCloudErrors: Observable<boolean>
}): {
    hoverifier: Hoverifier<RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec, HoverMerged, ActionItemAction>
    subscription: Unsubscribable
} {
    const subscription = new Subscription()

    /** Emits whenever the ref callback for the hover element is called */
    const hoverOverlayElements = new Subject<HTMLElement | null>()

    const relativeElement = document.body

    const containerComponentUpdates = new Subject<void>()

    subscription.add(
        registerHoverContributions({
            extensionsController,
            platformContext,
            history: H.createBrowserHistory(),
            locationAssign: location.assign.bind(location),
        })
    )

    // Code views come and go, but there is always a single hoverifier on the page
    const hoverifier = createHoverifier<
        RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec,
        HoverMerged,
        ActionItemAction
    >({
        hoverOverlayElements,
        hoverOverlayRerenders: containerComponentUpdates.pipe(
            withLatestFrom(hoverOverlayElements),
            map(([, hoverOverlayElement]) => ({ hoverOverlayElement, relativeElement })),
            filter(property('hoverOverlayElement', isDefined))
        ),
        getHover: ({ line, character, part, ...rest }) =>
            concat(
                [{ isLoading: true, result: null }],
                combineLatest([
                    from(extensionsController.extHostAPI).pipe(
                        withLatestFrom(privateCloudErrors),
                        switchMap(([extensionHost, hasPrivateCloudError]) =>
                            // Prevent GraphQL requests that we know will result in error/null when the repo is private (and not added to Cloud)
                            hasPrivateCloudError
                                ? of({ isLoading: true, result: null })
                                : wrapRemoteObservable(
                                      extensionHost.getHover(
                                          toTextDocumentPositionParameters({ ...rest, position: { line, character } })
                                      )
                                  )
                        )
                    ),
                    getActiveHoverAlerts([
                        ...hoverAlerts,
                        privateCloudErrors.pipe(
                            distinctUntilChanged(),
                            map(showAlert => (showAlert ? createPrivateCodeHoverAlert(codeHost) : undefined)),
                            filter(isDefined)
                        ),
                    ]),
                ]).pipe(
                    map(
                        ([{ isLoading, result: hoverMerged }, alerts]): MaybeLoadingResult<HoverMerged | null> => ({
                            isLoading,
                            result: hoverMerged || alerts?.length ? { contents: [], ...hoverMerged, alerts } : null,
                        })
                    )
                )
            ),
        getDocumentHighlights: ({ line, character, part, ...rest }) =>
            from(extensionsController.extHostAPI).pipe(
                withLatestFrom(privateCloudErrors),
                switchMap(([extensionHost, hasPrivateCloudError]) =>
                    // Prevent GraphQL requests that we know will result in error/null when the repo is private (and not added to Cloud)
                    hasPrivateCloudError
                        ? of([])
                        : wrapRemoteObservable(
                              extensionHost.getDocumentHighlights(
                                  toTextDocumentPositionParameters({ ...rest, position: { line, character } })
                              )
                          )
                )
            ),
        getActions: context =>
            // Prevent GraphQL requests that we know will result in error/null when the repo is private (and not added to Cloud)
            privateCloudErrors.pipe(
                take(1),
                switchMap(hasPrivateCloudError =>
                    hasPrivateCloudError ? of([]) : getHoverActions({ extensionsController, platformContext }, context)
                )
            ),
        tokenize: codeHost.codeViewsRequireTokenization,
    })

    class HoverOverlayContainer extends React.Component<
        {},
        HoverState<HoverContext, HoverMerged, ActionItemAction> & ThemeProps
    > {
        private subscription = new Subscription()
        private nextOverlayElement = hoverOverlayElements.next.bind(hoverOverlayElements)

        constructor(props: {}) {
            super(props)
            this.state = {
                ...hoverifier.hoverState,
                isLightTheme: true,
            }
        }
        public componentDidMount(): void {
            this.subscription.add(
                hoverifier.hoverStateUpdates.subscribe(update => {
                    this.setState(update)
                })
            )
            this.subscription.add(
                hoverifier.hoverStateUpdates
                    .pipe(
                        switchMap(({ hoveredTokenElement: token, hoverOverlayProps }) => {
                            if (token === undefined) {
                                return EMPTY
                            }
                            if (hoverOverlayProps === undefined) {
                                return EMPTY
                            }

                            const { actionsOrError } = hoverOverlayProps
                            const definitionAction =
                                Array.isArray(actionsOrError) &&
                                actionsOrError.find(a => a.action.id === 'goToDefinition.preloaded' && !a.disabledWhen)

                            const referenceAction =
                                Array.isArray(actionsOrError) &&
                                actionsOrError.find(a => a.action.id === 'findReferences' && !a.disabledWhen)

                            const action = definitionAction || referenceAction
                            if (!action) {
                                return EMPTY
                            }

                            const def = urlForClientCommandOpen(action.action, window.location.hash)
                            if (!def) {
                                return EMPTY
                            }

                            const oldCursor = token.style.cursor
                            token.style.cursor = 'pointer'

                            return fromEvent(token, 'click').pipe(
                                tap(() => {
                                    const selection = window.getSelection()
                                    if (selection !== null && selection.toString() !== '') {
                                        return
                                    }

                                    const actionType = action === definitionAction ? 'definition' : 'reference'
                                    telemetryService.log(`${actionType}CodeHost.click`)
                                    window.location.href = def
                                }),
                                finalize(() => (token.style.cursor = oldCursor))
                            )
                        })
                    )
                    .subscribe()
            )
            if (codeHost.isLightTheme) {
                this.subscription.add(
                    codeHost.isLightTheme.subscribe(isLightTheme => {
                        this.setState({ isLightTheme })
                    })
                )
            }
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
                    className={classNames(styles.hoverOverlay, codeHost.hoverOverlayClassProps?.className)}
                    telemetryService={telemetryService}
                    isLightTheme={this.state.isLightTheme}
                    hoverRef={this.nextOverlayElement}
                    extensionsController={extensionsController}
                    platformContext={platformContext}
                    location={H.createLocation(window.location)}
                    onAlertDismissed={onHoverAlertDismissed}
                    useBrandedLogo={true}
                />
            ) : null
        }
        private getHoverOverlayProps(): HoverState<HoverContext, HoverMerged, ActionItemAction>['hoverOverlayProps'] {
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
 * 1. User visits https://gitlab.com/gitlab-org/gitaly/-/merge_requests/1575. `div.tab-pane.diffs` doesn't exist yet (it'll be lazy-loaded) -> Mount the  hover overlay is to `document.body`.
 * 2. User visits the 'Changes' tab -> Unmount from `document.body`, mount to `div.tab-pane.diffs`.
 * 3. User visits the 'Overview' tab again -> `div.tab-pane.diffs` is hidden, and as a result so is the hover overlay.
 * 4. User navigates away from the merge request (soft-reload), `div.tab-pane.diffs` is removed -> Mount to `document.body` again.
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
    render: typeof reactDOMRender
    minimalUI: boolean
    hideActions?: boolean
    background: Pick<BackgroundPageApi, 'notifyPrivateCloudError' | 'openOptionsPage'>
}

/**
 * Opens extension options page
 */
const onConfigureSourcegraphClick: React.MouseEventHandler<HTMLAnchorElement> = async event => {
    event.preventDefault()
    if (isExtension) {
        await background.openOptionsPage()
    }
}

/**
 * @returns boolean indicating whether it is safe to continue initialization
 *
 * Returns
 * - "false": if could not parse repository name or if detected repository is private is not cloned in Sourcegraph Cloud
 * - "true" in all other cases
 *
 * Side-effect:
 * - Notifies background about private cloud error
 * - Renders "Configure Sourcegraph" or "Sign In" button
 */
const isSafeToContinueCodeIntel = async ({
    sourcegraphURL,
    codeHost,
    requestGraphQL,
    render,
}: Pick<HandleCodeHostOptions, 'render' | 'codeHost'> &
    Pick<HandleCodeHostOptions['platformContext'], 'requestGraphQL' | 'sourcegraphURL'>): Promise<boolean> => {
    if (!isDefaultSourcegraphUrl(sourcegraphURL) || !codeHost.getContext) {
        return true
    }
    // This is only when connected to Sourcegraph Cloud and code host either GitLab or GitHub
    try {
        const { privateRepository, rawRepoName } = await codeHost.getContext()
        if (!privateRepository) {
            // We can auto-clone public repos
            return true
        }

        const isRepoCloned = await resolvePrivateRepo({
            rawRepoName,
            requestGraphQL,
        }).toPromise()

        return isRepoCloned
    } catch (error) {
        // Ignore non-repository pages
        if (error instanceof RepoURLParseError) {
            return false
        }

        if (isExtension) {
            // Notify to show extension alert-icon
            background.notifyPrivateCloudError(true).catch(error => {
                console.error('Error notifying background page of private cloud.', error)
            })
        }

        if (!codeHost.getViewContextOnSourcegraphMount) {
            console.warn('Repository is not cloned or you are not authenticated to Sourcegraph.', error)
            return false
        }

        if (error instanceof NotAuthenticatedError) {
            // Show "Sign In" button
            console.warn('Not authenticated to Sourcegraph.', error)
            render(
                <SignInButton {...{ ...codeHost.viewOnSourcegraphButtonClassProps }} sourcegraphURL={sourcegraphURL} />,
                codeHost.getViewContextOnSourcegraphMount(document.body)
            )
        } else {
            // Show "Configure Sourcegraph" button
            console.warn('Repository is not cloned.', error)
            render(
                <ConfigureSourcegraphButton
                    {...codeHost.viewOnSourcegraphButtonClassProps}
                    className={classNames('open-on-sourcegraph', codeHost.viewOnSourcegraphButtonClassProps?.className)}
                    codeHostType={codeHost.type}
                    onConfigureSourcegraphClick={isInPage ? undefined : onConfigureSourcegraphClick}
                />,
                codeHost.getViewContextOnSourcegraphMount(document.body)
            )
        }

        return false
    }
}

export async function handleCodeHost({
    mutations,
    codeHost,
    extensionsController,
    platformContext,
    showGlobalDebug,
    telemetryService,
    render,
    minimalUI,
    hideActions,
    background,
}: HandleCodeHostOptions): Promise<Subscription> {
    const history = H.createBrowserHistory()
    const subscriptions = new Subscription()
    const { requestGraphQL, sourcegraphURL } = platformContext

    const addedElements = mutations.pipe(
        concatAll(),
        concatMap(mutation => mutation.addedNodes),
        filter(isInstanceOf(HTMLElement))
    )

    // Handle theming
    subscriptions.add(
        (codeHost.isLightTheme ?? of(true)).subscribe(isLightTheme => {
            document.body.classList.toggle('theme-light', isLightTheme)
            document.body.classList.toggle('theme-dark', !isLightTheme)
        })
    )
    const nativeTooltipsEnabled = codeHost.nativeTooltipResolvers
        ? nativeTooltipsEnabledFromSettings(platformContext.settings)
        : of(false)

    const hoverAlerts: Observable<HoverAlert>[] = []

    /**
     * A stream that emits a boolean that signifies
     * whether any request for the current repository has failed on the basis
     * that it is a private repository that has not been added to Sourcegraph Cloud
     * (only emits `true` when the Sourcegraph instance is Cloud).
     * If the current state is `true`, we can short circuit subsequent requests.
     * */
    const privateCloudErrors = new BehaviorSubject<boolean>(false)
    // Set by `ViewOnSourcegraphButton` (cleans up and sets to `false` whenever it is unmounted).
    const setPrivateCloudError = privateCloudErrors.next.bind(privateCloudErrors)

    /**
     * Checks whether the error occured because the repository
     * is a private repository that hasn't been added to Sourcegraph Cloud
     * (no side effects, doesn't notify `privateCloudErrors`)
     * */
    const checkPrivateCloudError = async (error: any): Promise<boolean> =>
        !!(
            isRepoNotFoundErrorLike(error) &&
            isDefaultSourcegraphUrl(sourcegraphURL) &&
            (await codeHost.getContext?.())?.privateRepository
        )

    if (codeHost.searchEnhancement) {
        subscriptions.add(initializeSearchEnhancement(codeHost.searchEnhancement, sourcegraphURL, mutations))
    }

    if (!(await isSafeToContinueCodeIntel({ sourcegraphURL, requestGraphQL, codeHost, render }))) {
        // Stop initializing code intelligence
        return subscriptions
    }

    if (codeHost.nativeTooltipResolvers) {
        const { subscription, nativeTooltipsAlert } = handleNativeTooltips(
            mutations,
            nativeTooltipsEnabled,
            codeHost,
            privateCloudErrors
        )
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
        privateCloudErrors,
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
        const repoExistsOrErrors = combineLatest([signInCloses.pipe(startWith(null)), from(getContext())]).pipe(
            switchMap(([, { rawRepoName, revision }]) =>
                resolveRevision({ repoName: rawRepoName, revision, requestGraphQL }).pipe(
                    retryWhenCloneInProgressError(),
                    mapTo(true),
                    startWith(undefined)
                )
            ),
            catchError(error => {
                if (isRepoNotFoundErrorLike(error) || error instanceof RepoURLParseError) {
                    return [false]
                }
                return [asError(error)]
            })
        )
        const onPrivateCloudError = (hasPrivateCloudError: boolean): void => {
            setPrivateCloudError(hasPrivateCloudError)
            if (isExtension) {
                background.notifyPrivateCloudError(hasPrivateCloudError).catch(error => {
                    console.error('Error notifying background page of private cloud error:', error)
                })
            }
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
                from(getContext()),
            ]).subscribe(([repoExistsOrError, mount, showSignInButton, context]) => {
                render(
                    <ViewOnSourcegraphButton
                        {...viewOnSourcegraphButtonClassProps}
                        codeHostType={codeHost.type}
                        context={context}
                        minimalUI={minimalUI}
                        sourcegraphURL={sourcegraphURL}
                        repoExistsOrError={repoExistsOrError}
                        showSignInButton={showSignInButton}
                        // The bound function is constant
                        onSignInClose={nextSignInClose}
                        onConfigureSourcegraphClick={isInPage ? undefined : onConfigureSourcegraphClick}
                        onPrivateCloudError={onPrivateCloudError}
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
                mergeMap(diffOrBlobInfo =>
                    resolveRepoNamesForDiffOrFileInfo(
                        diffOrBlobInfo,
                        checkPrivateCloudError,
                        platformContext.requestGraphQL
                    )
                ),
                mergeMap(diffOrBlobInfo =>
                    fetchFileContentForDiffOrFileInfo(
                        diffOrBlobInfo,
                        checkPrivateCloudError,
                        platformContext.requestGraphQL
                    ).pipe(
                        map(diffOrBlobInfo => ({
                            diffOrBlobInfo,
                            ...codeViewEvent,
                        }))
                    )
                ),
                catchError(error =>
                    // Ignore private Cloud RepoNotFound errors (don't initialize those code views)
                    from(checkPrivateCloudError(error)).pipe(hasPrivateCloudError => {
                        if (hasPrivateCloudError) {
                            return EMPTY
                        }
                        throw error
                    })
                ),
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
                                    // The bound function is constant
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
                            if (!isHTTPAuthError(error)) {
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
        observeOn(asyncScheduler)
    )

    /** Map from workspace URI to number of editors referencing it */
    const rootReferenceCounts = new Map<string, number>()

    /**
     * Adds root referenced by a code editor to the worskpace.
     */
    const addRootReference = async (uri: string, inputRevision: string | undefined): Promise<void> => {
        rootReferenceCounts.set(uri, (rootReferenceCounts.get(uri) || 0) + 1)
        if (rootReferenceCounts.get(uri) === 1) {
            const workspaceRoot: WorkspaceRoot = { uri, inputRevision }
            return extensionsController.extHostAPI
                .then(extensionHostAPI => extensionHostAPI.addWorkspaceRoot({ uri, inputRevision }))
                .catch(error =>
                    console.error('Sourcegraph: error adding workspace root', { error: asError(error), workspaceRoot })
                )
        }
    }

    /**
     * Deletes a reference to a workspace root from a code editor.
     */
    const deleteRootReference = async (uri: string): Promise<void> => {
        const currentReferenceCount = rootReferenceCounts.get(uri)
        if (!currentReferenceCount) {
            throw new Error(`No preexisting root refs for uri ${uri}`)
        }
        const updatedReferenceCount = currentReferenceCount - 1
        rootReferenceCounts.set(uri, updatedReferenceCount)
        if (updatedReferenceCount === 0) {
            return extensionsController.extHostAPI
                .then(extensionHostAPI => extensionHostAPI.removeWorkspaceRoot(uri))
                .catch(error =>
                    console.error('Sourcegraph: error removing workspace root', { error: asError(error), uri })
                )
        }
    }

    subscriptions.add(
        codeViews.subscribe(codeViewEvent => {
            console.log('Code view added')
            // This code view could have left the DOM between the time that
            // 1) it entered the DOM
            // 2) requests to Sourcegraph instance for repo name + file info fulfilled
            // When the code view leaves the DOM, the codeViewEvent's subscription is
            // unsubscribed, and new additions to the subscription are immediately invoked.
            // Check `wasRemoved` to prevent doing unnecessary work
            let wasRemoved = false
            codeViewEvent.subscriptions.add(() => {
                wasRemoved = true
                console.log('Code view removed')
            })

            if (wasRemoved) {
                return
            }

            ;(async () => {
                const {
                    element,
                    diffOrBlobInfo,
                    getPositionAdjuster,
                    getToolbarMount,
                    toolbarButtonProps,
                    overrideTokenize,
                } = codeViewEvent

                const initializeModelAndViewerForFileInfo = async (
                    fileInfo: FileInfoWithContent & FileInfoWithRepoName
                ): Promise<CodeEditorWithPartialModel> => {
                    const uri = toURIWithPath(fileInfo)

                    // Model
                    const languageId = getModeFromPath(fileInfo.filePath)
                    const model = { uri, languageId, text: fileInfo.content }

                    // Viewer
                    const editorData: CodeEditorData = {
                        type: 'CodeEditor' as const,
                        resource: uri,
                        selections: codeViewEvent.getSelections
                            ? codeViewEvent.getSelections(codeViewEvent.element)
                            : [],
                        isActive: true,
                    }

                    const extensionHostAPI = await extensionsController.extHostAPI

                    const rootURI = toRootURI(fileInfo)
                    const [, viewerId] = await Promise.all([
                        // Only add the model if it doesn't exist
                        // (there may be several code views on the page pointing to the same model)
                        extensionHostAPI.addTextDocumentIfNotExists(model),
                        extensionHostAPI.addViewerIfNotExists(editorData),
                        addRootReference(rootURI, fileInfo.revision),
                    ])

                    // Subscribe for removal
                    codeViewEvent.subscriptions.add(() => {
                        Promise.all([deleteRootReference(rootURI), extensionHostAPI.removeViewer(viewerId)]).catch(
                            error =>
                                console.error('Sourcegraph: error removing viewer and workspace root', {
                                    error: asError(error),
                                })
                        )
                    })

                    return {
                        ...editorData,
                        ...viewerId,
                        model,
                    }
                }

                const initializeModelAndViewerForDiffOrFileInfo = async (
                    diffOrFileInfo: DiffOrBlobInfo<FileInfoWithContent>
                ): Promise<DiffOrBlobInfo<FileInfoWithContent & { editor: CodeEditorWithPartialModel }>> => {
                    if ('blob' in diffOrFileInfo) {
                        return {
                            blob: {
                                ...diffOrFileInfo.blob,
                                editor: await initializeModelAndViewerForFileInfo(diffOrFileInfo.blob),
                            },
                        }
                    }
                    if (diffOrFileInfo.head && diffOrFileInfo.base) {
                        // For diffs, both editors are created (for head and base)
                        // but only one of them is passed into
                        // the `scope` of the CodeViewToolbar component.
                        // Both are used to listen for text decorations.
                        const [headEditor, baseEditor] = await Promise.all([
                            initializeModelAndViewerForFileInfo(diffOrFileInfo.head),
                            initializeModelAndViewerForFileInfo(diffOrFileInfo.base),
                        ])
                        return {
                            head: {
                                ...diffOrFileInfo.head,
                                editor: headEditor,
                            },
                            base: {
                                ...diffOrFileInfo.base,
                                editor: baseEditor,
                            },
                        }
                    }
                    if (diffOrFileInfo.base) {
                        return {
                            base: {
                                ...diffOrFileInfo.base,
                                editor: await initializeModelAndViewerForFileInfo(diffOrFileInfo.base),
                            },
                            head: undefined,
                        }
                    }
                    return {
                        head: {
                            ...diffOrFileInfo.head,
                            editor: await initializeModelAndViewerForFileInfo(diffOrFileInfo.head),
                        },
                        base: undefined,
                    }
                }

                const diffOrFileInfoWithEditor = await initializeModelAndViewerForDiffOrFileInfo(diffOrBlobInfo)

                let scopeEditor: CodeEditorWithPartialModel

                if ('blob' in diffOrFileInfoWithEditor) {
                    scopeEditor = diffOrFileInfoWithEditor.blob.editor
                } else if (diffOrFileInfoWithEditor.head) {
                    scopeEditor = diffOrFileInfoWithEditor.head.editor
                } else {
                    scopeEditor = diffOrFileInfoWithEditor.base.editor
                }

                if (wasRemoved) {
                    return
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

                const applyDecorationsForFileInfo = (editor: CodeEditorWithPartialModel, diffPart?: DiffPart): void => {
                    let decorationsByLine: DecorationMapByLine = new Map()
                    let previousIsLightTheme = true
                    const update = (decorations?: TextDocumentDecoration[] | null, isLightTheme?: boolean): void => {
                        try {
                            decorationsByLine = applyDecorations(
                                domFunctions,
                                element,
                                decorations ?? [],
                                decorationsByLine,
                                isLightTheme ?? true,
                                previousIsLightTheme,
                                diffPart
                            )
                            previousIsLightTheme = isLightTheme ?? true
                        } catch (error) {
                            console.error('Could not apply decorations to code view', codeViewEvent.element, error)
                        }
                    }

                    codeViewEvent.subscriptions.add(
                        combineLatest([
                            from(extensionsController.extHostAPI).pipe(
                                switchMap(extensionHostAPI =>
                                    wrapRemoteObservable(
                                        extensionHostAPI.getTextDecorations({
                                            viewerId: editor.viewerId,
                                        })
                                    )
                                )
                            ),
                            codeHost.isLightTheme ?? of(true),
                        ])
                            // Make sure extensions get cleaned up un unsubscription
                            .pipe(finalize(update))
                            // The nested subscribe cannot be replaced with a switchMap()
                            // We manage the subscription correctly.
                            // eslint-disable-next-line rxjs/no-nested-subscribe
                            .subscribe(([decorations, isLightTheme]) => update(decorations, isLightTheme))
                    )
                }

                // Apply decorations coming from extensions
                if (!minimalUI) {
                    if ('blob' in diffOrFileInfoWithEditor) {
                        applyDecorationsForFileInfo(diffOrFileInfoWithEditor.blob.editor)
                    } else {
                        if (diffOrFileInfoWithEditor.head) {
                            applyDecorationsForFileInfo(diffOrFileInfoWithEditor.head.editor, 'head')
                        }
                        if (diffOrFileInfoWithEditor.base) {
                            applyDecorationsForFileInfo(diffOrFileInfoWithEditor.base.editor, 'base')
                        }
                    }
                }

                // Add hover code intelligence
                const resolveContext: ContextResolver<RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec> = ({
                    part,
                }) => {
                    if ('blob' in diffOrBlobInfo) {
                        return defaultRevisionToCommitID(diffOrBlobInfo.blob)
                    }
                    if (diffOrBlobInfo.head && part === 'head') {
                        return defaultRevisionToCommitID(diffOrBlobInfo.head)
                    }
                    if (diffOrBlobInfo.base && part === 'base') {
                        return defaultRevisionToCommitID(diffOrBlobInfo.base)
                    }
                    throw new Error(`Could not resolve context for diff part ${JSON.stringify(part)}`)
                }

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
                                        tokenize: !!(typeof overrideTokenize === 'boolean'
                                            ? overrideTokenize
                                            : codeHost.codeViewsRequireTokenization),
                                    })
                                ),
                                resolveContext,
                                adjustPosition,
                                scrollBoundaries: codeViewEvent.getScrollBoundaries
                                    ? codeViewEvent.getScrollBoundaries(codeViewEvent.element)
                                    : [],
                                overrideTokenize,
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
                            actionItemClass={
                                codeViewEvent.toolbarButtonProps?.actionItemClass ??
                                codeHost.codeViewToolbarClassProps?.actionItemClass
                            }
                            hideActions={hideActions}
                            fileInfoOrError={diffOrBlobInfo}
                            sourcegraphURL={sourcegraphURL}
                            telemetryService={telemetryService}
                            platformContext={platformContext}
                            extensionsController={extensionsController}
                            buttonProps={toolbarButtonProps}
                            location={H.createLocation(window.location)}
                            scope={scopeEditor}
                            // The bound function is constant
                            onSignInClose={nextSignInClose}
                        />,
                        mount
                    )
                }
            })().catch(error => {
                console.error('Sourcegraph: uncaught error handling code view', asError(error))
            })
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

    return subscriptions
}

const SHOW_DEBUG = (): boolean => localStorage.getItem('debug') !== null

const CODE_HOSTS: CodeHost[] = [
    bitbucketServerCodeHost,
    bitbucketCloudCodeHost,
    githubCodeHost,
    gitlabCodeHost,
    phabricatorCodeHost,
    gerritCodeHost,
]

const CLOUD_CODE_HOST_HOSTS = ['github.com', 'gitlab.com']

export const determineCodeHost = (sourcegraphURL?: string): CodeHost | undefined => {
    const codeHost = CODE_HOSTS.find(codeHost => codeHost.check())

    if (!codeHost) {
        return undefined
    }

    // Prevent repo lookups for code hosts that we know cannot have repositories
    // cloned on sourcegraph.com. Repo lookups trigger cloning, which will
    // inevitably fail in this case.
    if (isDefaultSourcegraphUrl(sourcegraphURL)) {
        const { hostname } = new URL(location.href)
        const validCodeHost = CLOUD_CODE_HOST_HOSTS.some(cloudHost => cloudHost === hostname)
        if (!validCodeHost) {
            console.log(
                `Sourcegraph code host integration: stopped initialization since ${hostname} is not a supported code host when Sourcegraph URL is ${DEFAULT_SOURCEGRAPH_URL}.\n List of supported code hosts on ${DEFAULT_SOURCEGRAPH_URL}: ${CLOUD_CODE_HOST_HOSTS.join(
                    ', '
                )}`
            )
            return undefined
        }
    }

    return codeHost
}

function initializeSearchEnhancement(
    searchEnhancement: NonNullable<CodeHost['searchEnhancement']>,
    sourcegraphURL: string,
    mutations: Observable<MutationRecordLike[]>
): Subscription {
    const { searchViewResolver, resultViewResolver, onChange } = searchEnhancement
    const searchURL = new URL('/search', sourcegraphURL)
    searchURL.searchParams.append('utm_source', getPlatformName())
    searchURL.searchParams.append('utm_campaign', 'global-search')

    const searchView = mutations.pipe(
        trackViews([searchViewResolver]),
        switchMap(({ element }) =>
            fromEvent(element, 'input').pipe(
                map(event => (event.target as HTMLInputElement).value),
                startWith((element as HTMLInputElement).value)
            )
        ),
        map(value => ({
            value,
            searchURL: searchURL.href,
        })),
        observeOn(asyncScheduler)
    )
    const resultView = mutations.pipe(trackViews([resultViewResolver]), observeOn(asyncScheduler))

    return combineLatest([searchView, resultView])
        .pipe(map(([search, { element: resultElement }]) => ({ ...search, resultElement })))
        .subscribe(onChange)
}

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

    subscriptions.add(extensionsController)

    const overrideSendTelemetry = observeSourcegraphURL(isExtension).pipe(
        map(sourcegraphUrl => shouldOverrideSendTelemetry(isFirefox(), isExtension, sourcegraphUrl))
    )

    const observeSendTelemetry = combineLatest([overrideSendTelemetry, observeOptionFlag('sendTelemetry')]).pipe(
        map(([override, sendTelemetry]) => {
            if (override) {
                return true
            }
            return sendTelemetry
        })
    )

    const innerTelemetryService = new EventLogger(isExtension, requestGraphQL)
    const telemetryService = new ConditionalTelemetryService(innerTelemetryService, observeSendTelemetry)
    subscriptions.add(telemetryService)

    let codeHostSubscription: Subscription
    // In the browser extension, observe whether the `disableExtension` storage flag is set.
    // In the native integration, this flag does not exist.
    const extensionDisabled = isExtension ? observeStorageKey('sync', 'disableExtension') : of(false)

    // RFC 68: hide some UI features in the GitLab native integration.
    // This can be overridden using the `sourcegraphMinimalUI` local storage flag.
    const minimalUIStorageFlag = localStorage.getItem('sourcegraphMinimalUI')
    const minimalUI =
        minimalUIStorageFlag !== null ? minimalUIStorageFlag === 'true' : codeHost.type === 'gitlab' && !isExtension
    // Flag to hide the actions in the code view toolbar (hide ActionNavItems) leaving only the "Open on Sourcegraph" button in the toolbar.
    const hideActions = codeHost.type === 'gerrit'
    subscriptions.add(
        // eslint-disable-next-line rxjs/no-async-subscribe, @typescript-eslint/no-misused-promises
        extensionDisabled.subscribe(async disableExtension => {
            if (disableExtension) {
                // We don't need to unsubscribe if the extension starts with disabled state.
                if (codeHostSubscription) {
                    codeHostSubscription.unsubscribe()
                }
                console.log('Browser extension is disabled')
            } else {
                codeHostSubscription = await handleCodeHost({
                    mutations,
                    codeHost,
                    extensionsController,
                    platformContext,
                    showGlobalDebug,
                    telemetryService,
                    render: reactDOMRender,
                    minimalUI,
                    hideActions,
                    background,
                })
                subscriptions.add(codeHostSubscription)
                console.log(`${isExtension ? 'Browser extension' : 'Native integration'} is enabled`)
            }
        })
    )
    return subscriptions
}
