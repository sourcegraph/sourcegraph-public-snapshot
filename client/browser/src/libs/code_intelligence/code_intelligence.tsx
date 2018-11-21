import {
    ContextResolver,
    createHoverifier,
    DiffPart,
    DOMFunctions,
    findPositionsFromEvents,
    Hoverifier,
    HoverOverlay,
    HoverState,
    LinkComponent,
    PositionAdjuster,
} from '@sourcegraph/codeintellify'
import { propertyIsDefined } from '@sourcegraph/codeintellify/lib/helpers'
import { HoverMerged } from '@sourcegraph/codeintellify/lib/types'
import { toPrettyBlobURL } from '@sourcegraph/codeintellify/lib/url'
import * as H from 'history'
import * as React from 'react'
import { render } from 'react-dom'
import { animationFrameScheduler, BehaviorSubject, Observable, of, Subject, Subscription, Unsubscribable } from 'rxjs'
import { filter, map, mergeMap, observeOn, withLatestFrom } from 'rxjs/operators'

import { Environment } from '../../../../../shared/src/api/client/environment'
import { TextDocumentItem } from '../../../../../shared/src/api/client/types/textDocument'
import { getModeFromPath } from '../../../../../shared/src/languages'
import {
    createJumpURLFetcher,
    createLSPFromExtensions,
    JumpURLLocation,
    lspViaAPIXlang,
    toTextDocumentIdentifier,
} from '../../shared/backend/lsp'
import { ButtonProps, CodeViewToolbar } from '../../shared/components/CodeViewToolbar'
import { AbsoluteRepo, AbsoluteRepoFile } from '../../shared/repo'
import { eventLogger, sourcegraphUrl, useExtensions } from '../../shared/util/context'
import { bitbucketServerCodeHost } from '../bitbucket/code_intelligence'
import { githubCodeHost } from '../github/code_intelligence'
import { gitlabCodeHost } from '../gitlab/code_intelligence'
import { phabricatorCodeHost } from '../phabricator/code_intelligence'
import { findCodeViews, getContentOfCodeView } from './code_views'
import { applyDecoration, Controllers, initializeExtensions } from './extensions'

/**
 * Defines a type of code view a given code host can have. It tells us how to
 * look for the code view and how to do certain things when we find it.
 */
export interface CodeView {
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
    adjustPosition?: PositionAdjuster
    /** Props for styling the buttons in the `CodeViewToolbar`. */
    toolbarButtonProps?: ButtonProps

    isDiff?: boolean

    /** Gets the 1-indexed range of the code view */
    getLineRanges?: (
        codeView: HTMLElement,
        part?: DiffPart
    ) => {
        /** The first line shown in the code view. */
        start: number
        /** The last line shown in the code view. */
        end: number
    }[]
}

export type CodeViewWithOutSelector = Pick<CodeView, Exclude<keyof CodeView, 'selector'>>

export interface CodeViewResolver {
    selector: string
    resolveCodeView: (elem: HTMLElement) => CodeViewWithOutSelector | null
}

interface OverlayPosition {
    top: number
    left: number
}

/**
 * A function that gets the mount location for elements being mounted to the DOM.
 */
export type MountGetter = () => HTMLElement

/** Information for adding code intelligence to code views on arbitrary code hosts. */
export interface CodeHost {
    /**
     * The name of the code host. This will be added as a className to the overlay mount.
     */
    name: string

    /**
     * Checks to see if the current context the code is running in is within
     * the given code host.
     */
    check: () => Promise<boolean> | boolean

    /**
     * The list of types of code views to try to annotate.
     */
    codeViews?: CodeView[]

    /**
     * Resolve `CodeView`s from the DOM. This is useful when each code view type
     * doesn't have a distinct selector for
     */
    codeViewResolver?: CodeViewResolver

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

    /** Build the J2D url from the location. */
    buildJumpURLLocation?: (def: JumpURLLocation) => string
}

export interface FileInfo {
    /**
     * The path for the repo the file belongs to. If a `baseRepoPath` is provided, this value
     * is treated as the head repo path.
     */
    repoPath: string
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
     * The repo bath for the BASE side of a diff. This is useful for Phabricator
     * staging areas since they are separate repos.
     */
    baseRepoPath?: string
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

/**
 * Prepares the page for code intelligence. It creates the hoverifier, injects
 * and mounts the hover overlay and then returns the hoverifier.
 *
 * @param codeHost
 */
function initCodeIntelligence(
    codeHost: CodeHost,
    environment: BehaviorSubject<Pick<Environment, 'roots' | 'visibleTextDocuments'>>
): {
    hoverifier: Hoverifier
    controllers: Partial<Controllers>
} {
    const { platformContext, extensionsController }: Partial<Controllers> =
        useExtensions && codeHost.getCommandPaletteMount
            ? initializeExtensions(codeHost.getCommandPaletteMount, environment)
            : {}
    const simpleProviderFns = extensionsController ? createLSPFromExtensions(extensionsController) : lspViaAPIXlang

    /** Emits when the go to definition button was clicked */
    const goToDefinitionClicks = new Subject<MouseEvent>()
    const nextGoToDefinitionClick = (event: MouseEvent) => goToDefinitionClicks.next(event)

    /** Emits when the close button was clicked */
    const closeButtonClicks = new Subject<MouseEvent>()
    const nextCloseButtonClick = (event: MouseEvent) => closeButtonClicks.next(event)

    /** Emits whenever the ref callback for the hover element is called */
    const hoverOverlayElements = new Subject<HTMLElement | null>()
    const nextOverlayElement = (element: HTMLElement | null) => hoverOverlayElements.next(element)

    const classNames = ['hover-overlay-mount', `hover-overlay-mount__${codeHost.name}`]

    const createMount = () => {
        const overlayMount = document.createElement('div')
        overlayMount.style.height = '0px'
        for (const className of classNames) {
            overlayMount.classList.add(className)
        }
        document.body.appendChild(overlayMount)
        return overlayMount
    }

    const overlayMount = document.querySelector(`.${classNames.join('.')}`) || createMount()

    const relativeElement = document.body

    const fetchJumpURL = createJumpURLFetcher(
        simpleProviderFns.fetchDefinition,
        codeHost.buildJumpURLLocation || toPrettyBlobURL
    )

    const containerComponentUpdates = new Subject<void>()

    const hoverifier = createHoverifier({
        closeButtonClicks,
        goToDefinitionClicks,
        hoverOverlayElements,
        hoverOverlayRerenders: containerComponentUpdates.pipe(
            withLatestFrom(hoverOverlayElements),
            map(([, hoverOverlayElement]) => ({ hoverOverlayElement, relativeElement })),
            filter(propertyIsDefined('hoverOverlayElement'))
        ),
        pushHistory: path => {
            location.href = path
        },
        fetchHover: ({ line, character, part, ...rest }) =>
            simpleProviderFns
                .fetchHover({ ...rest, position: { line, character } })
                .pipe(map(hover => (hover ? (hover as HoverMerged) : hover))),
        fetchJumpURL,
        logTelemetryEvent: () => eventLogger.logCodeIntelligenceEvent(),
    })

    const Link: LinkComponent = ({ to, children, ...rest }) => (
        <a href={new URL(to, sourcegraphUrl).href} {...rest}>
            {children}
        </a>
    )

    class HoverOverlayContainer extends React.Component<{}, HoverState> {
        constructor(props: {}) {
            super(props)
            this.state = hoverifier.hoverState
            hoverifier.hoverStateUpdates.subscribe(update => this.setState(update))
        }
        public componentDidMount(): void {
            containerComponentUpdates.next()
        }
        public componentDidUpdate(): void {
            containerComponentUpdates.next()
        }
        public render(): JSX.Element | null {
            const hoverOverlayProps = this.getHoverOverlayProps()
            return hoverOverlayProps ? (
                <HoverOverlay
                    {...hoverOverlayProps}
                    linkComponent={Link}
                    logTelemetryEvent={this.log}
                    hoverRef={nextOverlayElement}
                    onGoToDefinitionClick={nextGoToDefinitionClick}
                    onCloseButtonClick={nextCloseButtonClick}
                />
            ) : null
        }
        private log = () => eventLogger.logCodeIntelligenceEvent()
        private getHoverOverlayProps(): HoverState['hoverOverlayProps'] {
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

    render(<HoverOverlayContainer />, overlayMount)

    return { hoverifier, controllers: { platformContext, extensionsController } }
}

/**
 * ResolvedCodeView attaches an actual code view DOM element that was found on
 * the page to the CodeView type being passed around by this file.
 */
export interface ResolvedCodeView extends CodeViewWithOutSelector {
    /** The code view DOM element. */
    codeView: HTMLElement
}

function handleCodeHost(codeHost: CodeHost): Subscription {
    const environmentSubject = new BehaviorSubject<Pick<Environment, 'roots' | 'visibleTextDocuments'>>({
        roots: null,
        visibleTextDocuments: null,
    })
    const {
        hoverifier,
        controllers: { platformContext, extensionsController },
    } = initCodeIntelligence(codeHost, environmentSubject)

    const subscriptions = new Subscription()

    subscriptions.add(hoverifier)

    // Keeps track of all documents on the page since calling this function (should be once per page).
    let visibleTextDocuments: TextDocumentItem[] = []

    subscriptions.add(
        of(document.body)
            .pipe(
                findCodeViews(codeHost),
                mergeMap(({ codeView, resolveFileInfo, ...rest }) =>
                    resolveFileInfo(codeView).pipe(map(info => ({ info, codeView, ...rest })))
                ),
                observeOn(animationFrameScheduler)
            )
            .subscribe(
                ({
                    codeView,
                    info,
                    isDiff,
                    getLineRanges,
                    dom,
                    adjustPosition,
                    getToolbarMount,
                    toolbarButtonProps,
                }) => {
                    const toRootURI = (ctx: AbsoluteRepo) => `git://${ctx.repoPath}?${ctx.commitID}`
                    const toURIWithPath = (ctx: AbsoluteRepoFile) =>
                        `git://${ctx.repoPath}?${ctx.commitID}#${ctx.filePath}`

                    const originalDOM = dom
                    dom = {
                        ...dom,
                        // If any parent element has the sourcegraph-extension-element
                        // class then that element does not have any code. We
                        // must check for "any parent element" because extensions
                        // create their DOM changes before the blob is tokenized
                        // into multiple elements.
                        getCodeElementFromTarget: (target: HTMLElement): HTMLElement | null =>
                            target.closest('.sourcegraph-extension-element') !== null
                                ? null
                                : originalDOM.getCodeElementFromTarget(target),
                    }
                    if (extensionsController) {
                        let content = info.content
                        let baseContent = info.baseContent

                        if (!content) {
                            if (!getLineRanges) {
                                throw new Error('Must either provide a line range getter or provide file contents')
                            }

                            const contents = getContentOfCodeView(codeView, { isDiff, getLineRanges, dom })

                            content = contents.content
                            baseContent = contents.baseContent
                        }

                        visibleTextDocuments = [
                            // All the currently open documents
                            ...visibleTextDocuments,
                            // Either a normal file, or HEAD when codeView is a diff
                            {
                                uri: toURIWithPath(info),
                                languageId: getModeFromPath(info.filePath) || 'could not determine mode',
                                text: content!,
                            },
                        ]
                        const roots: Environment['roots'] = [{ uri: toRootURI(info) }]

                        // When codeView is a diff, add BASE too.
                        if (baseContent! && info.baseRepoPath && info.baseCommitID && info.baseFilePath) {
                            visibleTextDocuments.push({
                                uri: toURIWithPath({
                                    repoPath: info.baseRepoPath,
                                    commitID: info.baseCommitID,
                                    filePath: info.baseFilePath,
                                }),
                                languageId: getModeFromPath(info.filePath) || 'could not determine mode',
                                text: baseContent!,
                            })
                            roots.push({
                                uri: toRootURI({
                                    repoPath: info.baseRepoPath,
                                    commitID: info.baseCommitID,
                                }),
                            })
                        }

                        if (extensionsController && !info.baseCommitID) {
                            let oldDecorations: Unsubscribable[] = []

                            extensionsController.registries.textDocumentDecoration
                                .getDecorations(toTextDocumentIdentifier(info))
                                .subscribe(decorations => {
                                    for (const old of oldDecorations) {
                                        old.unsubscribe()
                                    }
                                    oldDecorations = []
                                    for (const decoration of decorations || []) {
                                        try {
                                            oldDecorations.push(
                                                applyDecoration(dom, {
                                                    codeView,
                                                    decoration,
                                                })
                                            )
                                        } catch (e) {
                                            console.warn(e)
                                        }
                                    }
                                })
                        }

                        environmentSubject.next({ roots, visibleTextDocuments })
                    }

                    const resolveContext: ContextResolver = ({ part }) => ({
                        repoPath: part === 'base' ? info.baseRepoPath || info.repoPath : info.repoPath,
                        commitID: part === 'base' ? info.baseCommitID! : info.commitID,
                        filePath: part === 'base' ? info.baseFilePath || info.filePath : info.filePath,
                        rev: part === 'base' ? info.baseRev || info.baseCommitID! : info.rev || info.commitID,
                    })

                    subscriptions.add(
                        hoverifier.hoverify({
                            dom,
                            positionEvents: of(codeView).pipe(findPositionsFromEvents(dom)),
                            resolveContext,
                            adjustPosition,
                        })
                    )

                    codeView.classList.add('sg-mounted')

                    if (!getToolbarMount) {
                        return
                    }

                    const mount = getToolbarMount(codeView)

                    render(
                        <CodeViewToolbar
                            {...info}
                            platformContext={platformContext}
                            extensionsController={extensionsController}
                            buttonProps={
                                toolbarButtonProps || {
                                    className: '',
                                    style: {},
                                }
                            }
                            simpleProviderFns={
                                extensionsController ? createLSPFromExtensions(extensionsController) : lspViaAPIXlang
                            }
                            location={H.createLocation(window.location)}
                        />,
                        mount
                    )
                }
            )
    )

    return subscriptions
}

async function injectCodeIntelligenceToCodeHosts(codeHosts: CodeHost[]): Promise<Subscription> {
    const subscriptions = new Subscription()

    for (const codeHost of codeHosts) {
        const isCodeHost = await Promise.resolve(codeHost.check())
        if (isCodeHost) {
            subscriptions.add(handleCodeHost(codeHost))
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
export async function injectCodeIntelligence(): Promise<Subscription> {
    const codeHosts: CodeHost[] = [bitbucketServerCodeHost, githubCodeHost, gitlabCodeHost, phabricatorCodeHost]

    return await injectCodeIntelligenceToCodeHosts(codeHosts)
}
