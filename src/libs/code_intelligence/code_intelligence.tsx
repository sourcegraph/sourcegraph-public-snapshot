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
import * as React from 'react'
import { render } from 'react-dom'
import { animationFrameScheduler, Observable, of, Subject, Subscription } from 'rxjs'
import { filter, map, mergeMap, observeOn, withLatestFrom } from 'rxjs/operators'

import { createJumpURLFetcher } from '../../shared/backend/lsp'
import { lspViaAPIXlang } from '../../shared/backend/lsp'
import { ButtonProps, CodeViewToolbar } from '../../shared/components/CodeViewToolbar'
import { eventLogger, sourcegraphUrl } from '../../shared/util/context'
import { githubCodeHost } from '../github/code_intelligence'
import { phabricatorCodeHost } from '../phabricator/code_intelligence'
import { findCodeViews } from './code_views'

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
}

export type CodeViewWithOutSelector = Pick<CodeView, Exclude<keyof CodeView, 'selector'>>

export interface CodeViewResolver {
    selector: string
    resolveCodeView: (elem: HTMLElement) => CodeViewWithOutSelector
}

/** Information for adding code intelligence to code views on arbitrary code hosts. */
export interface CodeHost {
    /**
     * The name of the code host. This will be added as a className to the overlay mount.
     */
    name: string

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
     * Checks to see if the current context the code is running in is within
     * the given code host.
     */
    check: () => Promise<boolean> | boolean
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
}

/**
 * Prepares the page for code intelligence. It creates the hoverifier, injects
 * and mounts the hover overlay and then returns the hoverifier.
 *
 * @param codeHost
 */
function initCodeIntelligence(codeHost: CodeHost): { hoverifier: Hoverifier } {
    /** Emits when the go to definition button was clicked */
    const goToDefinitionClicks = new Subject<MouseEvent>()
    const nextGoToDefinitionClick = (event: MouseEvent) => goToDefinitionClicks.next(event)

    /** Emits when the close button was clicked */
    const closeButtonClicks = new Subject<MouseEvent>()
    const nextCloseButtonClick = (event: MouseEvent) => closeButtonClicks.next(event)

    /** Emits whenever the ref callback for the hover element is called */
    const hoverOverlayElements = new Subject<HTMLElement | null>()
    const nextOverlayElement = (element: HTMLElement | null) => hoverOverlayElements.next(element)

    const overlayMount = document.createElement('div')
    overlayMount.style.height = '0px'
    overlayMount.classList.add('hover-overlay-mount')
    overlayMount.classList.add(`hover-overlay-mount__${codeHost.name}`)
    document.body.appendChild(overlayMount)

    const relativeElement = document.body

    const fetchJumpURL = createJumpURLFetcher(lspViaAPIXlang.fetchDefinition, toPrettyBlobURL)

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
            lspViaAPIXlang
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
            return this.state.hoverOverlayProps ? (
                <HoverOverlay
                    {...this.state.hoverOverlayProps}
                    linkComponent={Link}
                    logTelemetryEvent={this.log}
                    hoverRef={nextOverlayElement}
                    onGoToDefinitionClick={nextGoToDefinitionClick}
                    onCloseButtonClick={nextCloseButtonClick}
                />
            ) : null
        }
        private log = () => eventLogger.logCodeIntelligenceEvent()
    }

    render(<HoverOverlayContainer />, overlayMount)

    return { hoverifier }
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
    const { hoverifier } = initCodeIntelligence(codeHost)

    const subscriptions = new Subscription()

    subscriptions.add(
        of(document.body)
            .pipe(
                findCodeViews(codeHost),
                mergeMap(({ codeView, resolveFileInfo, ...rest }) =>
                    resolveFileInfo(codeView).pipe(map(info => ({ info, codeView, ...rest })))
                ),
                observeOn(animationFrameScheduler)
            )
            .subscribe(({ codeView, info, dom, adjustPosition, getToolbarMount, toolbarButtonProps }) => {
                const resolveContext: ContextResolver = ({ part }) => ({
                    repoPath: part === 'base' ? info.baseRepoPath || info.repoPath : info.repoPath,
                    commitID: part === 'base' ? info.baseCommitID! : info.commitID,
                    filePath: part === 'base' ? info.baseFilePath! : info.filePath,
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
                        buttonProps={
                            toolbarButtonProps || {
                                className: '',
                                style: {},
                            }
                        }
                        simpleProviderFns={lspViaAPIXlang}
                    />,
                    mount
                )
            })
    )

    return subscriptions
}

async function injectCodeIntelligenceToCodeHosts(codeHosts: CodeHost[]): Promise<Subscription> {
    const subscriptions = new Subscription()

    for (const codeHost of codeHosts) {
        const isCodeHost = await Promise.resolve(codeHost.check())
        if (isCodeHost) {
            subscriptions.add(handleCodeHost(codeHost))
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
    const codeHosts: CodeHost[] = [githubCodeHost, phabricatorCodeHost]

    return await injectCodeIntelligenceToCodeHosts(codeHosts)
}
