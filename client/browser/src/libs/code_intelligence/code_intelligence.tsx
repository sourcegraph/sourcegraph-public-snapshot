import {
    ContextResolver,
    createHoverifier,
    DiffPart,
    DOMFunctions,
    findPositionsFromEvents,
    Hoverifier,
    HoverState,
    PositionAdjuster,
} from '@sourcegraph/codeintellify'
import { Selection } from '@sourcegraph/extension-api-types'
import * as H from 'history'
import { isEqual, uniqBy } from 'lodash'
import * as React from 'react'
import { render } from 'react-dom'
import { animationFrameScheduler, EMPTY, fromEvent, Observable, of, Subject, Subscription, Unsubscribable } from 'rxjs'
import {
    catchError,
    concatAll,
    concatMap,
    distinctUntilChanged,
    filter,
    map,
    mergeMap,
    observeOn,
    startWith,
    withLatestFrom,
} from 'rxjs/operators'
import { ActionItemAction } from '../../../../../shared/src/actions/ActionItem'
import { CodeEditorData } from '../../../../../shared/src/api/client/services/editorService'
import { WorkspaceRootWithMetadata } from '../../../../../shared/src/api/client/services/workspaceService'
import { HoverMerged } from '../../../../../shared/src/api/client/types/hover'
import { CommandListClassProps } from '../../../../../shared/src/commandPalette/CommandList'
import { Controller } from '../../../../../shared/src/extensions/controller'
import { registerHighlightContributions } from '../../../../../shared/src/highlight/contributions'
import { getHoverActions, registerHoverContributions } from '../../../../../shared/src/hover/actions'
import { HoverContext, HoverOverlay, HoverOverlayClassProps } from '../../../../../shared/src/hover/HoverOverlay'
import { getModeFromPath } from '../../../../../shared/src/languages'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { NOOP_TELEMETRY_SERVICE } from '../../../../../shared/src/telemetry/telemetryService'
import { isDefined, isInstanceOf, propertyIsDefined } from '../../../../../shared/src/util/types'
import {
    FileSpec,
    lprToSelectionsZeroIndexed,
    parseHash,
    PositionSpec,
    RepoSpec,
    ResolvedRevSpec,
    RevSpec,
    toRootURI,
    toURIWithPath,
    ViewStateSpec,
} from '../../../../../shared/src/util/url'
import { sendMessage } from '../../browser/runtime'
import { isInPage } from '../../context'
import { ERPRIVATEREPOPUBLICSOURCEGRAPHCOM } from '../../shared/backend/errors'
import { createLSPFromExtensions, toTextDocumentIdentifier } from '../../shared/backend/lsp'
import { ButtonProps, CodeViewToolbar, CodeViewToolbarClassProps } from '../../shared/components/CodeViewToolbar'
import { resolveRev, retryWhenCloneInProgressError } from '../../shared/repo/backend'
import { sourcegraphUrl } from '../../shared/util/context'
import { MutationRecordLike } from '../../shared/util/dom'
import { bitbucketServerCodeHost } from '../bitbucket/code_intelligence'
import { githubCodeHost } from '../github/code_intelligence'
import { gitlabCodeHost } from '../gitlab/code_intelligence'
import { phabricatorCodeHost } from '../phabricator/code_intelligence'
import { fetchFileContents, trackCodeViews } from './code_views'
import { applyDecorations, initializeExtensions, renderCommandPalette, renderGlobalDebug } from './extensions'
import { renderViewContextOnSourcegraph, ViewOnSourcegraphButtonClassProps } from './external_links'

registerHighlightContributions()

/**
 * Defines a type of code view a given code host can have. It tells us how to
 * look for the code view and how to do certain things when we find it.
 */
export interface CodeViewSpec {
    /** A selector used by `document.querySelectorAll` to find the code view. */
    selector: string
    /** The DOMFunctions for the code view. */
    dom: DOMFunctions
    /**
     * Finds or creates a DOM element where we should inject the
     * `CodeViewToolbar`. This function is responsible for ensuring duplicate
     * mounts aren't created.
     */
    getToolbarMount?: (codeView: HTMLElement) => HTMLElement
    /**
     * Resolves the file info for a given code view. It returns an observable
     * because some code hosts need to resolve this asynchronously. The
     * observable should only emit once.
     */
    resolveFileInfo: (codeView: HTMLElement) => Observable<FileInfo>
    /**
     * In some situations, we need to be able to adjust the position going into
     * and coming out of codeintellify. For example, Phabricator converts tabs
     * to spaces in it's DOM.
     */
    adjustPosition?: PositionAdjuster<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec>
    /** Props for styling the buttons in the `CodeViewToolbar`. */
    toolbarButtonProps?: ButtonProps

    isDiff?: boolean
}

export type CodeViewSpecWithOutSelector = Pick<CodeViewSpec, Exclude<keyof CodeViewSpec, 'selector'>>

export interface CodeViewSpecResolver {
    /**
     * Selector that is used to find code views on the page with `querySelectorAll()`.
     */
    selector: string

    /**
     * Function that is called for each element that was found with `selector` to determine which code view the element is.
     */
    resolveCodeViewSpec: (elem: HTMLElement) => CodeViewSpecWithOutSelector | null
}

interface OverlayPosition {
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
export type CodeHostContext = RepoSpec & Partial<RevSpec>

/** Information for adding code intelligence to code views on arbitrary code hosts. */
export interface CodeHost {
    /**
     * The name of the code host. This will be added as a className to the overlay mount.
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
     * The list of types of code views to try to annotate.
     *
     * The set of code views tracked on a page is the union of all code views found using `codeViewSpecs` and `codeViewResolver`.
     */
    codeViewSpecs?: CodeViewSpec[]

    /**
     * Resolve `CodeView`s from the DOM. This is useful when each code view type
     * doesn't have a distinct selector.
     *
     * The set of code views tracked on a page is the union of all code views found using `codeViewSpecs` and `codeViewResolver`.
     */
    codeViewSpecResolver?: CodeViewSpecResolver

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

    /** Construct the URL to the specified file. */
    urlToFile?: (
        location: RepoSpec & RevSpec & FileSpec & Partial<PositionSpec> & Partial<ViewStateSpec> & { part?: DiffPart }
    ) => string

    /** Returns a stream representing the selections in the current code view */
    selectionsChanges?: () => Observable<Selection[]>

    /**
     * CSS classes for the command palette to customize styling
     */
    commandPaletteClassProps?: CommandListClassProps

    /**
     * CSS classes for the code view toolbar to customize styling
     */
    codeViewToolbarClassProps?: CodeViewToolbarClassProps
}

export interface FileInfo {
    /**
     * The path for the repo the file belongs to. If a `baseRepoName` is provided, this value
     * is treated as the head repo name.
     */
    repoName: string
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
    baseRepoName?: string
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

interface CodeIntelligenceProps
    extends PlatformContextProps<'forceUpdateTooltip' | 'urlToFile' | 'sideloadedExtensionURL'> {
    codeHost: CodeHost
    extensionsController: Controller
    showGlobalDebug?: boolean
}

export const createOverlayMount = (codeHostName: string): HTMLElement => {
    const mount = document.createElement('div')
    mount.classList.add('hover-overlay-mount', `hover-overlay-mount__${codeHostName}`)
    document.body.appendChild(mount)
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
 *
 * @param codeHost
 */
export function initCodeIntelligence({
    codeHost,
    platformContext,
    extensionsController,
}: Pick<CodeIntelligenceProps, 'codeHost' | 'platformContext' | 'extensionsController'>): {
    hoverifier: Hoverifier<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec, HoverMerged, ActionItemAction>
    subscription: Unsubscribable
} {
    const subscription = new Subscription()

    const { getHover } = createLSPFromExtensions(extensionsController)

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
    const hoverifier = createHoverifier<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec, HoverMerged, ActionItemAction>(
        {
            closeButtonClicks,
            hoverOverlayElements,
            hoverOverlayRerenders: containerComponentUpdates.pipe(
                withLatestFrom(hoverOverlayElements),
                map(([, hoverOverlayElement]) => ({ hoverOverlayElement, relativeElement })),
                filter(propertyIsDefined('hoverOverlayElement'))
            ),
            getHover: ({ line, character, part, ...rest }) => getHover({ ...rest, position: { line, character } }),
            getActions: context => getHoverActions({ extensionsController, platformContext }, context),
        }
    )

    class HoverOverlayContainer extends React.Component<{}, HoverState<HoverContext, HoverMerged, ActionItemAction>> {
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
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                    hoverRef={nextOverlayElement}
                    extensionsController={extensionsController}
                    platformContext={platformContext}
                    location={H.createLocation(window.location)}
                    onCloseButtonClick={nextCloseButtonClick}
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

    // This renders to document.body, which we can assume is never removed,
    // so we don't need to subscribe to mutations.
    const overlayMount = createOverlayMount(codeHost.name)
    render(<HoverOverlayContainer />, overlayMount)

    return { hoverifier, subscription }
}

/**
 * ResolvedCodeView attaches an actual code view DOM element that was found on
 * the page to the CodeView type being passed around by this file.
 */
export interface ResolvedCodeView extends CodeViewSpecWithOutSelector {
    /** The code view DOM element. */
    element: HTMLElement
}

export function handleCodeHost({
    mutations,
    codeHost,
    extensionsController,
    platformContext,
    showGlobalDebug,
}: CodeIntelligenceProps & { mutations: Observable<MutationRecordLike[]> }): Subscription {
    const history = H.createBrowserHistory()
    const subscriptions = new Subscription()

    const ensureRepoExists = (context: CodeHostContext) =>
        resolveRev(context).pipe(
            retryWhenCloneInProgressError(),
            map(rev => !!rev),
            catchError(() => [false])
        )

    const openOptionsMenu = () => {
        sendMessage({ type: 'openOptionsPage' })
    }

    const addedElements = mutations.pipe(
        concatAll(),
        concatMap(mutation => mutation.addedNodes),
        filter(isInstanceOf(HTMLElement))
    )

    const { hoverifier, subscription } = initCodeIntelligence({ codeHost, extensionsController, platformContext })
    subscriptions.add(hoverifier)
    subscriptions.add(subscription)

    // Inject UI components
    // Render command palette
    if (codeHost.getCommandPaletteMount) {
        subscriptions.add(
            addedElements
                .pipe(
                    map(codeHost.getCommandPaletteMount),
                    filter(isDefined)
                )
                .subscribe(
                    renderCommandPalette({
                        extensionsController,
                        history,
                        platformContext,
                        ...codeHost.commandPaletteClassProps,
                    })
                )
        )
    }

    // Render extension debug menu
    // This renders to document.body, which we can assume is never removed,
    // so we don't need to subscribe to mutations.
    if (showGlobalDebug) {
        const mount = createGlobalDebugMount()
        renderGlobalDebug({ extensionsController, platformContext, history })(mount)
    }

    // Render view on Sourcegraph button
    if (codeHost.getViewContextOnSourcegraphMount && codeHost.getContext) {
        const { getContext, viewOnSourcegraphButtonClassProps } = codeHost
        subscriptions.add(
            addedElements
                .pipe(
                    map(codeHost.getViewContextOnSourcegraphMount),
                    filter(isDefined)
                )
                .subscribe(
                    renderViewContextOnSourcegraph({
                        sourcegraphUrl,
                        getContext,
                        viewOnSourcegraphButtonClassProps,
                        ensureRepoExists,
                        onConfigureSourcegraphClick: isInPage ? undefined : openOptionsMenu,
                    })
                )
        )
    }

    // A stream of selections for the current code view. By default, selections
    // are parsed from the location hash, but the codeHost can provide an alternative implementation.
    const selectionsChanges: Observable<Selection[]> = codeHost.selectionsChanges
        ? codeHost.selectionsChanges()
        : fromEvent(window, 'hashchange').pipe(
              map(() => lprToSelectionsZeroIndexed(parseHash(window.location.hash))),
              distinctUntilChanged(isEqual),
              startWith([])
          )

    /** A stream of added or removed code views */
    const codeViews = mutations.pipe(
        trackCodeViews(codeHost),
        mergeMap(codeViewEvent =>
            codeViewEvent.type === 'added'
                ? codeViewEvent.resolveFileInfo(codeViewEvent.element).pipe(
                      mergeMap(fileInfo =>
                          fetchFileContents(fileInfo).pipe(
                              map(fileInfoWithContents => ({
                                  fileInfo: fileInfoWithContents,
                                  ...codeViewEvent,
                              }))
                          )
                      )
                  )
                : [codeViewEvent]
        ),
        catchError(err => {
            if (err.name === ERPRIVATEREPOPUBLICSOURCEGRAPHCOM) {
                return EMPTY
            }
            throw err
        }),
        observeOn(animationFrameScheduler)
    )

    interface CodeViewState {
        subscriptions: Subscription
        editors: CodeEditorData[]
        roots: WorkspaceRootWithMetadata[]
    }
    /** Map from code view element to the state associated with it (to be updated or removed) */
    const codeViewStates = new Map<Element, CodeViewState>()

    // Update code editors as selections change
    subscriptions.add(
        selectionsChanges.subscribe(selections => {
            extensionsController.services.editor.editors.next(
                [...codeViewStates.values()].flatMap(state => state.editors).map(editor => ({ ...editor, selections }))
            )
        })
    )

    subscriptions.add(
        codeViews.pipe(withLatestFrom(selectionsChanges)).subscribe(([codeViewEvent, selections]) => {
            console.log(`Code view ${codeViewEvent.type}`)

            // Handle added or removed view component, workspace root and subscriptions
            if (codeViewEvent.type === 'added' && !codeViewStates.has(codeViewEvent.element)) {
                const { element, fileInfo, adjustPosition, getToolbarMount, toolbarButtonProps } = codeViewEvent
                const codeViewState: CodeViewState = {
                    subscriptions: new Subscription(),
                    editors: [
                        {
                            type: 'CodeEditor' as const,
                            item: {
                                uri: toURIWithPath(fileInfo),
                                languageId: getModeFromPath(fileInfo.filePath),
                                text: fileInfo.content,
                            },
                            selections,
                            isActive: true,
                        },
                    ],
                    roots: [{ uri: toRootURI(fileInfo), inputRevision: fileInfo.rev || '' }],
                }
                codeViewStates.set(element, codeViewState)

                // When codeView is a diff (and not an added file), add BASE too.
                if (fileInfo.baseContent && fileInfo.baseRepoName && fileInfo.baseCommitID && fileInfo.baseFilePath) {
                    codeViewState.editors.push({
                        type: 'CodeEditor' as const,
                        item: {
                            uri: toURIWithPath({
                                repoName: fileInfo.baseRepoName,
                                commitID: fileInfo.baseCommitID,
                                filePath: fileInfo.baseFilePath,
                            }),
                            languageId: getModeFromPath(fileInfo.filePath),
                            text: fileInfo.baseContent,
                        },
                        // There is no notion of a selection on diff views yet, so this is empty.
                        selections: [],
                        isActive: true,
                    })
                    codeViewState.roots.push({
                        uri: toRootURI({
                            repoName: fileInfo.baseRepoName,
                            commitID: fileInfo.baseCommitID,
                        }),
                        inputRevision: fileInfo.baseRev || '',
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
                let decoratedLines: number[] = []
                if (!fileInfo.baseCommitID) {
                    codeViewState.subscriptions.add(
                        extensionsController.services.textDocumentDecoration
                            .getDecorations(toTextDocumentIdentifier(fileInfo))
                            // The nested subscribe cannot be replaced with a switchMap()
                            // We manage the subscription correctly.
                            // tslint:disable-next-line: rxjs-no-nested-subscribe
                            .subscribe(decorations => {
                                decoratedLines = applyDecorations(
                                    domFunctions,
                                    element,
                                    decorations || [],
                                    decoratedLines
                                )
                            })
                    )
                }

                // Add hover code intelligence
                const resolveContext: ContextResolver<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec> = ({
                    part,
                }) => ({
                    repoName: part === 'base' ? fileInfo.baseRepoName || fileInfo.repoName : fileInfo.repoName,
                    commitID: part === 'base' ? fileInfo.baseCommitID! : fileInfo.commitID,
                    filePath: part === 'base' ? fileInfo.baseFilePath || fileInfo.filePath : fileInfo.filePath,
                    rev:
                        part === 'base'
                            ? fileInfo.baseRev || fileInfo.baseCommitID!
                            : fileInfo.rev || fileInfo.commitID,
                })
                codeViewState.subscriptions.add(
                    hoverifier.hoverify({
                        dom: domFunctions,
                        positionEvents: of(element).pipe(findPositionsFromEvents(domFunctions)),
                        resolveContext,
                        adjustPosition,
                    })
                )

                element.classList.add('sg-mounted')

                // Render toolbar
                if (getToolbarMount) {
                    const mount = getToolbarMount(element)
                    render(
                        <CodeViewToolbar
                            {...fileInfo}
                            {...codeHost.codeViewToolbarClassProps}
                            telemetryService={NOOP_TELEMETRY_SERVICE}
                            platformContext={platformContext}
                            extensionsController={extensionsController}
                            buttonProps={toolbarButtonProps}
                            location={H.createLocation(window.location)}
                            scope={codeViewState.editors[0]}
                        />,
                        mount
                    )
                }
            } else if (codeViewEvent.type === 'removed') {
                const codeViewState = codeViewStates.get(codeViewEvent.element)
                if (codeViewState) {
                    codeViewState.subscriptions.unsubscribe()
                    codeViewStates.delete(codeViewEvent.element)
                }
            }

            // Apply added/removed roots/editors
            extensionsController.services.editor.editors.next(
                [...codeViewStates.values()].flatMap(state => state.editors)
            )
            extensionsController.services.workspace.roots.next(
                uniqBy([...codeViewStates.values()].flatMap(state => state.roots), root => root.uri)
            )
        })
    )

    return subscriptions
}

const SHOW_DEBUG = () => localStorage.getItem('debug') !== null

const CODE_HOSTS = [bitbucketServerCodeHost, githubCodeHost, gitlabCodeHost, phabricatorCodeHost]
export const determineCodeHost = (): CodeHost | undefined => CODE_HOSTS.find(codeHost => codeHost.check())

export async function injectCodeIntelligenceToCodeHost(
    mutations: Observable<MutationRecordLike[]>,
    codeHost: CodeHost,
    showGlobalDebug = SHOW_DEBUG()
): Promise<Subscription> {
    const subscriptions = new Subscription()
    const { platformContext, extensionsController } = initializeExtensions(codeHost)
    subscriptions.add(extensionsController)
    subscriptions.add(
        handleCodeHost({
            mutations,
            codeHost,
            extensionsController,
            platformContext,
            showGlobalDebug,
        })
    )
    return subscriptions
}
