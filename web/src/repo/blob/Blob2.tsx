import * as H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { merge, of, Subject, Subscription } from 'rxjs'
import {
    catchError,
    debounceTime,
    delay,
    filter,
    map,
    share,
    switchMap,
    takeUntil,
    tap,
    withLatestFrom,
} from 'rxjs/operators'
import { Hover } from 'vscode-languageserver-types'
import { AbsoluteRepoFile, RenderMode } from '..'
import { fetchHover } from '../../backend/lsp'
import { asError, ErrorLike } from '../../util/errors'
import { HoverOverlay } from './HoverOverlay'
import { convertNode, getTableDataCell, getTargetLineAndOffset } from './tooltips'

/**
 * Calculates the desired position of the hover overlay depending on the container,
 * the hover target and the size of the hover overlay
 *
 * @param scrollable The closest container that is scrollable
 * @param target The DOM Node that was hovered
 * @param tooltip The DOM Node of the tooltip
 */
const calculateOverlayPosition = (
    scrollable: HTMLElement,
    target: HTMLElement,
    tooltip: HTMLElement
): { left: number; top: number } => {
    // The scrollable element is the one with scrollbars. The scrolling element is the one with the content.
    const scrollableBounds = scrollable.getBoundingClientRect()
    const scrollingElement = scrollable.firstElementChild! // table that we're positioning tooltips relative to.
    const scrollingBounds = scrollingElement.getBoundingClientRect() // tables bounds
    const targetBound = target.getBoundingClientRect() // our target elements bounds

    // Anchor it horizontally, prior to rendering to account for wrapping
    // changes to vertical height if the tooltip is at the edge of the viewport.
    const relLeft = targetBound.left - scrollingBounds.left

    // Anchor the tooltip vertically.
    const tooltipBound = tooltip.getBoundingClientRect()
    const relTop = targetBound.top + scrollable.scrollTop - scrollableBounds.top
    const margin = 0
    let tooltipTop = relTop - (tooltipBound.height + margin)
    if (tooltipTop - scrollable.scrollTop < 0) {
        // Tooltip wouldn't be visible from the top, so display it at the
        // bottom.
        const relBottom = targetBound.bottom + scrollable.scrollTop - scrollableBounds.top
        tooltipTop = relBottom + margin
    }
    return { left: relLeft, top: tooltipTop }
}

interface BlobProps extends AbsoluteRepoFile {
    /** The trusted syntax-highlighted code as HTML */
    html: string

    location: H.Location
    history: H.History
    className: string
    wrapCode: boolean
    renderMode: RenderMode
}

const LOADING: 'loading' = 'loading'

interface BlobState {
    hoverOrError?: typeof LOADING | Hover | null | ErrorLike
    hoverIsFixed: boolean
    hoverPosition?: { left: number; top: number }
}

export class Blob2 extends React.Component<BlobProps, BlobState> {
    /** Emits with the latest Props on every componentDidUpdate and on componentDidMount */
    private componentUpdates = new Subject<BlobProps>()

    /** Emits whenever the ref callback for the code element is called */
    private codeElements = new Subject<HTMLElement | null>()
    private nextCodeElement = (element: HTMLElement | null) => this.codeElements.next(element)

    /** Emits whenever the ref callback for the hover element is called */
    private hoverOverlayElements = new Subject<HTMLElement | null>()
    private nextOverlayElement = (element: HTMLElement | null) => this.hoverOverlayElements.next(element)

    /** Emits whenever something is hovered in the code */
    private codeMouseOvers = new Subject<React.MouseEvent<HTMLElement>>()
    private nextBlobMouseOver = (event: React.MouseEvent<HTMLElement>) => this.codeMouseOvers.next(event)

    /** Emits whenever something is clicked in the code */
    private codeClicks = new Subject<React.MouseEvent<HTMLElement>>()
    private nextCodeClick = (event: React.MouseEvent<HTMLElement>) => this.codeClicks.next(event)

    /** Subscriptions to be disposed on unmout */
    private subscriptions = new Subscription()

    constructor(props: BlobProps) {
        super(props)
        this.state = {
            hoverIsFixed: false,
        }

        const codeMouseOverTargets = this.codeMouseOvers.pipe(
            map(event => event.target as HTMLElement),
            // Casting is okay here, we know these are HTMLElements
            withLatestFrom(this.codeElements),
            // If there was a mouseover, there _must_ have been a blob element
            map(([target, codeElement]) => ({ target, codeElement: codeElement! })),
            debounceTime(50),
            // SIDE EFFECT (but idempotent)
            // If not done for this cell, wrap the tokens in this cell to enable finding the precise positioning.
            // This may be possible in other ways (looking at mouse position and rendering characters), but it works
            tap(({ target, codeElement }) => {
                const td = getTableDataCell(target, codeElement)
                if (td && !td.classList.contains('annotated')) {
                    convertNode(td)
                    td.classList.add('annotated')
                }
            }),
            share()
        )

        const codeClickTargets = this.codeClicks.pipe(
            map(event => event.target as HTMLElement),
            withLatestFrom(this.codeElements),
            // If there was a click, there _must_ have been a blob element
            map(([target, codeElement]) => ({ target, codeElement: codeElement! }))
        )

        // On every componentDidUpdate (after the component was rerendered, e.g. from a hover state update) resposition
        // the tooltip
        // It's important to add this subscription first so that withLatestFrom will be guaranteed to have gotten the
        // latest hover target by the time componentDidUpdate is triggered from the setState() in the second chain
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    filter(() => !this.state.hoverIsFixed),
                    withLatestFrom(this.hoverOverlayElements, codeMouseOverTargets),
                    map(([, hoverElement, { target, codeElement }]) => ({ hoverElement, target, codeElement })),
                    filter(
                        (data): data is typeof data & { hoverElement: NonNullable<typeof data['hoverElement']> } =>
                            !!data.hoverElement
                    )
                )
                .subscribe(({ codeElement, hoverElement, target }) => {
                    const hoverPosition = calculateOverlayPosition(codeElement, target, hoverElement)
                    this.setState({ hoverPosition })
                })
        )

        // On every new mouse over, fetch new hover contents and update the state
        this.subscriptions.add(
            merge(
                codeMouseOverTargets.pipe(filter(() => !this.state.hoverIsFixed)),
                codeClickTargets.pipe(filter(() => this.state.hoverIsFixed))
            )
                .pipe(
                    switchMap(({ target, codeElement }) => {
                        // Find out the position that was hovered over
                        const position = getTargetLineAndOffset(target, codeElement, false)
                        if (!position) {
                            return []
                        }
                        // Fetch the hover for that position
                        const hoverFetch = fetchHover({
                            repoPath: this.props.repoPath,
                            commitID: this.props.commitID,
                            filePath: this.props.filePath,
                            position,
                        }).pipe(catchError(error => [asError(error)]), share())
                        // Show a loader if it hasn't returned after 100ms
                        return merge(hoverFetch, of(LOADING).pipe(delay(100), takeUntil(hoverFetch)))
                    })
                )
                .subscribe(hoverOrError => {
                    // Reset the hover position, it's gonna be repositioned after the hover was rendered
                    this.setState({ hoverOrError, hoverPosition: undefined })
                })
        )

        // Whenever something is clicked in the code, fix/unfix the hover
        this.subscriptions.add(
            this.codeClicks.subscribe(() => {
                this.setState(
                    prevState =>
                        prevState.hoverIsFixed
                            ? {
                                  hoverIsFixed: false,
                                  hoverOrError: undefined,
                                  hoverPosition: undefined,
                              }
                            : {
                                  hoverIsFixed: true,
                              }
                )
            })
        )
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public shouldComponentUpdate?(nextProps: Readonly<BlobProps>, nextState: Readonly<BlobState>): boolean {
        return !isEqual(this.props, nextProps) || !isEqual(this.state, nextState)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): React.ReactNode {
        return (
            <div className={`blob2 ${this.props.className}`}>
                <code
                    className={`blob2__code ${this.props.wrapCode ? ' blob2__code--wrapped' : ''} `}
                    ref={this.nextCodeElement}
                    dangerouslySetInnerHTML={{ __html: this.props.html }}
                    onClick={this.nextCodeClick}
                    onMouseOver={this.nextBlobMouseOver}
                />
                {this.state.hoverOrError && (
                    <HoverOverlay
                        hoverRef={this.nextOverlayElement}
                        hoverOrError={this.state.hoverOrError}
                        position={this.state.hoverPosition}
                        isFixed={this.state.hoverIsFixed}
                    />
                )}
            </div>
        )
    }
}
