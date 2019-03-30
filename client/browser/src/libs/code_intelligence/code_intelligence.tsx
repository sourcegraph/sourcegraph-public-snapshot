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
import { createPortal, render } from 'react-dom'
import { animationFrameScheduler, EMPTY, fromEvent, Observable, of, Subject, Subscription } from 'rxjs'
import {
    catchError,
    distinctUntilChanged,
    filter,
    map,
    mergeMap,
    observeOn,
    startWith,
    tap,
    withLatestFrom,
} from 'rxjs/operators'
import { registerHighlightContributions } from '../../../../../shared/src/highlight/contributions'

import { ActionItemProps } from '../../../../../shared/src/actions/ActionItem'
import { ActionNavItemsClassProps } from '../../../../../shared/src/actions/ActionsNavItems'
import { ViewComponentData, WorkspaceRootWithMetadata } from '../../../../../shared/src/api/client/model'
import { HoverMerged } from '../../../../../shared/src/api/client/types/hover'
import { Controller } from '../../../../../shared/src/extensions/controller'
import { getHoverActions, registerHoverContributions } from '../../../../../shared/src/hover/actions'
import { HoverContext, HoverOverlay } from '../../../../../shared/src/hover/HoverOverlay'
import { getModeFromPath } from '../../../../../shared/src/languages'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { TelemetryContext } from '../../../../../shared/src/telemetry/telemetryContext'
import { propertyIsDefined } from '../../../../../shared/src/util/types'
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
import { ButtonProps, CodeViewToolbar } from '../../shared/components/CodeViewToolbar'
import { resolveRev, retryWhenCloneInProgressError } from '../../shared/repo/backend'
import { eventLogger, sourcegraphUrl } from '../../shared/util/context'
import { MutationRecordLike } from '../../shared/util/dom'
import { bitbucketServerCodeHost } from '../bitbucket/code_intelligence'
import { githubCodeHost } from '../github/code_intelligence'
import { gitlabCodeHost } from '../gitlab/code_intelligence'
import { phabricatorCodeHost } from '../phabricator/code_intelligence'
import { fetchFileContents, trackCodeViews } from './code_views'
import { applyDecorations, initializeExtensions, injectCommandPalette, injectGlobalDebug } from './extensions'
import { injectViewContextOnSourcegraph } from './external_links'

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
    getToolbarMount?: (codeView: HTMLElement, part?: DiffPart) => HTMLElement
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
    selector: string
    resolveCodeViewSpec: (elem: HTMLElement) => CodeViewSpecWithOutSelector | null
}

interface OverlayPosition {
    top: number
    left: number
}

/**
 * A function that gets the mount location for elements being mounted to the DOM.
 */
export type MountGetter = () => HTMLElement

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
     * The mount location for the contextual link to Sourcegraph.
     */
    getViewContextOnSourcegraphMount?: () => HTMLElement | null
    /**
     * Optional class name for the contextual link to Sourcegraph.
     */
    contextButtonClassName?: string

    /**
     * Checks to see if the current context the code is running in is within
     * the given code host.
     */
    check: () => Promise<boolean> | boolean

    /**
     * Gets the mount location for the hover overlay. Defaults to a created `<div>`
     * that is appended to `document.body`. Use this control to remove the
     * tooltip when the portion of the page containing the code views is removed
     * (e.g. from a soft page reload).
     */
    getOverlayMount?: () => HTMLElement | null

    // Code views and code view resolver form a union

    /**
     * The list of types of code views to try to annotate.
     */
    codeViewSpecs?: CodeViewSpec[]

    /**
     * Resolve `CodeView`s from the DOM. This is useful when each code view type
     * doesn't have a distinct selector
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
     * Get the DOM element where we'll mount the command palette for extensions.
     */
    getCommandPaletteMount?: MountGetter

    /**
     * Get the DOM element where we'll mount the small global debug menu for extensions in the bottom right.
     */
    getGlobalDebugMount?: MountGetter

    /** Construct the URL to the specified file. */
    urlToFile?: (
        location: RepoSpec & RevSpec & FileSpec & Partial<PositionSpec> & Partial<ViewStateSpec> & { part?: DiffPart }
    ) => string

    /** Returns a stream representing the selections in the current code view */
    selectionsChanges?: () => Observable<Selection[]>

    /** Optional classes for ActionNavItems, useful to customize the style of buttons contributed to the code view toolbar */
    actionNavItemClassProps?: ActionNavItemsClassProps

    /** Optional class to set on the command palette popover element */
    commandPalettePopoverClassName?: string

    /** Optional class to set on the code view toolbar element */
    codeViewToolbarClassName?: string
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

    headHasFileContents?: boolean
    baseHasFileContents?: boolean

    content?: string
    baseContent?: string
}

interface CodeIntelligenceProps
    extends PlatformContextProps<'forceUpdateTooltip' | 'urlToFile' | 'sideloadedExtensionURL'> {
    codeHost: CodeHost
    extensionsController: Controller
    showGlobalDebug?: boolean
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
}: CodeIntelligenceProps): Hoverifier<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec, HoverMerged, ActionItemProps> {
    const { getHover } = createLSPFromExtensions(extensionsController)

    /** Emits when the close button was clicked */
    const closeButtonClicks = new Subject<MouseEvent>()
    const nextCloseButtonClick = (event: MouseEvent) => closeButtonClicks.next(event)

    /** Emits whenever the ref callback for the hover element is called */
    const hoverOverlayElements = new Subject<HTMLElement | null>()
    const nextOverlayElement = (element: HTMLElement | null) => hoverOverlayElements.next(element)

    const relativeElement = document.body

    const containerComponentUpdates = new Subject<void>()

    registerHoverContributions({ extensionsController, platformContext, history: H.createBrowserHistory() })

    const hoverifier = createHoverifier<RepoSpec & RevSpec & FileSpec & ResolvedRevSpec, HoverMerged, ActionItemProps>({
        closeButtonClicks,
        hoverOverlayElements,
        hoverOverlayRerenders: containerComponentUpdates.pipe(
            withLatestFrom(hoverOverlayElements),
            map(([, hoverOverlayElement]) => ({ hoverOverlayElement, relativeElement })),
            filter(propertyIsDefined('hoverOverlayElement'))
        ),
        getHover: ({ line, character, part, ...rest }) =>
            getHover({ ...rest, position: { line, character } }).pipe(
                map(hover => (hover ? (hover as HoverMerged) : hover))
            ),
        getActions: context => getHoverActions({ extensionsController, platformContext }, context),
    })

    const classNames = ['hover-overlay-mount', `hover-overlay-mount__${codeHost.name}`]

    const createOverlayContainerMount = () => {
        const overlayMount = document.createElement('div')
        overlayMount.style.height = '0px'
        overlayMount.classList.add('overlay-mount-container')
        document.body.appendChild(overlayMount)
        return overlayMount
    }

    const overlayContainerMount = document.querySelector('.overlay-mount-container') || createOverlayContainerMount()

    const getOverlayMount = (): HTMLElement => {
        let mount: HTMLElement | null = document.querySelector('.sg-overlay-mount')
        if (mount) {
            mount.parentElement!.removeChild(mount)
        }

        if (codeHost.getOverlayMount) {
            mount = codeHost.getOverlayMount()
        }

        if (!mount) {
            mount = document.createElement('div')
            overlayContainerMount.appendChild(mount)
        }

        mount.classList.add('sg-overlay-mount')
        for (const className of classNames) {
            mount.classList.add(className)
        }

        return mount
    }

    class HoverOverlayContainer extends React.Component<{}, HoverState<HoverContext, HoverMerged, ActionItemProps>> {
        private portal: HTMLElement | null = null

        constructor(props: {}) {
            super(props)
            this.state = hoverifier.hoverState
            hoverifier.hoverStateUpdates.subscribe(update => this.setState(update))
        }
        public componentDidMount(): void {
            containerComponentUpdates.next()
        }
        public componentDidUpdate(): void {
            if (!this.portal || !document.body.contains(this.portal)) {
                this.portal = getOverlayMount()
            }

            containerComponentUpdates.next()
        }
        public render(): JSX.Element | null {
            const hoverOverlayProps = this.getHoverOverlayProps()
            return hoverOverlayProps && this.portal
                ? createPortal(
                      <HoverOverlay
                          {...hoverOverlayProps}
                          hoverRef={nextOverlayElement}
                          extensionsController={extensionsController!}
                          platformContext={platformContext!}
                          location={H.createLocation(window.location)}
                          onCloseButtonClick={nextCloseButtonClick}
                      />,
                      this.portal
                  )
                : null
        }
        private getHoverOverlayProps(): HoverState<HoverContext, HoverMerged, ActionItemProps>['hoverOverlayProps'] {
            if (!this.state.hoverOverlayProps) {
                return undefined
            }

            let { overlayPosition, ...rest } = this.state.hoverOverlayProps
            if (overlayPosition && codeHost.adjustOverlayPosition) {
                overlayPosition = codeHost.adjustOverlayPosition(overlayPosition)
            }

            return {
                ...rest,
                overlayPosition,
            }
        }
    }

    render(
        <TelemetryContext.Provider value={eventLogger}>
            <HoverOverlayContainer />
        </TelemetryContext.Provider>,
        overlayContainerMount
    )

    return hoverifier
}

/**
 * ResolvedCodeView attaches an actual code view DOM element that was found on
 * the page to the CodeView type being passed around by this file.
 */
export interface ResolvedCodeView extends CodeViewSpecWithOutSelector {
    /** The code view DOM element. */
    codeViewElement: HTMLElement
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
            catchError((err: Error) => {
                if (err.name === ERPRIVATEREPOPUBLICSOURCEGRAPHCOM) {
                    return [false]
                }

                return [true]
            })
        )

    const openOptionsMenu = () => {
        sendMessage({ type: 'openOptionsPage' })
    }

    const hoverifier = initCodeIntelligence({ codeHost, extensionsController, platformContext, showGlobalDebug })
    subscriptions.add(hoverifier)

    // Inject UI components
    subscriptions.add(
        mutations.subscribe(mutations => {
            // We don't need to inspect the mutations because the functions are idempotent
            // TODO optimize
            // Maybe getMount() would have to be replaced by a function that determines if a new mount was added from a given mutation record

            injectCommandPalette({
                extensionsController,
                platformContext,
                history,
                getMount: codeHost.getCommandPaletteMount,
                popoverClassName: codeHost.commandPalettePopoverClassName,
            })
            injectGlobalDebug({
                extensionsController,
                platformContext,
                getMount: codeHost.getGlobalDebugMount,
                history,
                showGlobalDebug,
            })
            injectViewContextOnSourcegraph(
                sourcegraphUrl,
                codeHost,
                ensureRepoExists,
                isInPage ? undefined : openOptionsMenu
            )
        })
    )

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
        tap(codeView => console.log('CodeView ' + codeView.type, codeView.codeViewElement)),
        mergeMap(codeViewEvent =>
            codeViewEvent.type === 'added'
                ? codeViewEvent.resolveFileInfo(codeViewEvent.codeViewElement).pipe(
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
        visibleViewComponents: ViewComponentData[]
        roots: WorkspaceRootWithMetadata[]
    }
    /** Map from code view element to the state associated with it (to be updated or removed) */
    const codeViewStates = new Map<Element, CodeViewState>()

    // Update model as selections change
    subscriptions.add(
        selectionsChanges.subscribe(selections => {
            extensionsController.services.model.model.next({
                ...extensionsController.services.model.model.value,
                visibleViewComponents: [...codeViewStates.values()]
                    .flatMap(state => state.visibleViewComponents)
                    .map(visibleViewComponent => ({ ...visibleViewComponent, selections })),
            })
        })
    )

    subscriptions.add(
        codeViews.pipe(withLatestFrom(selectionsChanges)).subscribe(([codeViewEvent, selections]) => {
            // Handle added or removed view component, workspace root and subscriptions
            if (codeViewEvent.type === 'added' && !codeViewStates.has(codeViewEvent.codeViewElement)) {
                const { codeViewElement, fileInfo, adjustPosition, getToolbarMount, toolbarButtonProps } = codeViewEvent
                const codeViewState: CodeViewState = {
                    subscriptions: new Subscription(),
                    visibleViewComponents: [
                        {
                            type: 'textEditor' as 'textEditor',
                            item: {
                                uri: toURIWithPath(fileInfo),
                                languageId: getModeFromPath(fileInfo.filePath) || 'could not determine mode',
                                text: fileInfo.content,
                            },
                            selections,
                            isActive: true,
                        },
                    ],
                    roots: [{ uri: toRootURI(fileInfo), inputRevision: fileInfo.rev || '' }],
                }
                codeViewStates.set(codeViewElement, codeViewState)

                // When codeView is a diff, add BASE too.
                if (fileInfo.baseContent && fileInfo.baseRepoName && fileInfo.baseCommitID && fileInfo.baseFilePath) {
                    codeViewState.visibleViewComponents.push({
                        type: 'textEditor',
                        item: {
                            uri: toURIWithPath({
                                repoName: fileInfo.baseRepoName,
                                commitID: fileInfo.baseCommitID,
                                filePath: fileInfo.baseFilePath,
                            }),
                            languageId: getModeFromPath(fileInfo.filePath) || 'could not determine mode',
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
                            .subscribe(decorations => {
                                decoratedLines = applyDecorations(
                                    domFunctions,
                                    codeViewElement,
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
                        positionEvents: of(codeViewElement).pipe(findPositionsFromEvents(domFunctions)),
                        resolveContext,
                        adjustPosition,
                    })
                )

                codeViewElement.classList.add('sg-mounted')

                // Render toolbar
                if (getToolbarMount) {
                    const mount = getToolbarMount(codeViewElement)
                    render(
                        <TelemetryContext.Provider value={eventLogger}>
                            <CodeViewToolbar
                                {...fileInfo}
                                {...codeHost.actionNavItemClassProps}
                                platformContext={platformContext}
                                extensionsController={extensionsController}
                                buttonProps={
                                    toolbarButtonProps || {
                                        className: '',
                                        style: {},
                                    }
                                }
                                location={H.createLocation(window.location)}
                                className={codeHost.codeViewToolbarClassName}
                            />
                        </TelemetryContext.Provider>,
                        mount
                    )
                }
            } else if (codeViewEvent.type === 'removed') {
                const codeViewState = codeViewStates.get(codeViewEvent.codeViewElement)
                if (codeViewState) {
                    codeViewState.subscriptions.unsubscribe()
                    codeViewStates.delete(codeViewEvent.codeViewElement)
                }
            }

            // Apply added/removed roots/visibleViewComponents
            extensionsController.services.model.model.next({
                roots: uniqBy([...codeViewStates.values()].flatMap(state => state.roots), root => root.uri),
                visibleViewComponents: [...codeViewStates.values()].flatMap(state => state.visibleViewComponents),
            })
        })
    )

    return subscriptions
}

const SHOW_DEBUG = () => localStorage.getItem('debug') !== null

export async function injectCodeIntelligenceToCodeHosts(
    mutations: Observable<MutationRecordLike[]>,
    codeHosts: CodeHost[],
    showGlobalDebug = SHOW_DEBUG()
): Promise<Subscription> {
    const subscriptions = new Subscription()

    // Find the right code host
    for (const codeHost of codeHosts) {
        const isCodeHost = await Promise.resolve(codeHost.check())
        if (isCodeHost) {
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
            break
        }
    }

    return subscriptions
}

/**
 * Injects all code hosts into the page.
 *
 * @returns A promise with a subscription containing all subscriptions for code
 * intelligence. Unsubscribing will clean up subscriptions for hoverify and any
 * incomplete setup requests.
 */
export async function injectCodeIntelligence(mutations: Observable<MutationRecordLike[]>): Promise<Subscription> {
    const codeHosts: CodeHost[] = [bitbucketServerCodeHost, githubCodeHost, gitlabCodeHost, phabricatorCodeHost]

    return await injectCodeIntelligenceToCodeHosts(mutations, codeHosts)
}
