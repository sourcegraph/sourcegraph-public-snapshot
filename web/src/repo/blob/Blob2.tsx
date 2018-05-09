import * as H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { concat, merge, of, Subject, Subscription } from 'rxjs'
import {
    catchError,
    debounceTime,
    delay,
    distinctUntilChanged,
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
import { fetchHover, fetchJumpURL, isEmptyHover } from '../../backend/lsp'
import { asError, ErrorLike } from '../../util/errors'
import { parseHash } from '../../util/url'
import { HoverOverlay, isJumpURL } from './HoverOverlay'
import { convertNode, getTableDataCell, getTargetLineAndOffset, HoveredToken } from './tooltips'

/**
 * `padding-top` of the blob element in px.
 * TODO find a way to remove the need for this.
 */
const BLOB_PADDING_TOP = 8

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
    // This is the padding-top of the blob element
    let tooltipTop = relTop - (tooltipBound.height - BLOB_PADDING_TOP)
    if (tooltipTop - scrollable.scrollTop < 0) {
        // Tooltip wouldn't be visible from the top, so display it at the
        // bottom.
        const relBottom = targetBound.bottom + scrollable.scrollTop - scrollableBounds.top
        tooltipTop = relBottom + BLOB_PADDING_TOP
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

const isHover = (val: any): val is Hover => typeof val === 'object' && val !== null && Array.isArray(val.contents)

function isDefined<T>(val: T): val is NonNullable<T> {
    return val !== undefined && val !== null
}

interface BlobState {
    hoverOrError?: typeof LOADING | Hover | null | ErrorLike
    definitionURLOrError?: typeof LOADING | { jumpURL: string } | null | ErrorLike
    hoverIsFixed: boolean

    /** The desired position of the hover overlay */
    overlayPosition?: { left: number; top: number }

    /**
     * Whether the user has clicked the go to definition button for the current overlay yet,
     * and whether he pressed Ctrl/Cmd while doing it to open it in a new tab or not.
     */
    clickedGoToDefinition: false | 'same-tab' | 'new-tab'

    /** The currently hovered token */
    hoveredToken?: HoveredToken
}

export class Blob2 extends React.Component<BlobProps, BlobState> {
    /** Emits with the latest Props on every componentDidUpdate and on componentDidMount */
    private componentUpdates = new Subject<BlobProps>()

    /** Emits whenever the ref callback for the code element is called */
    private codeElements = new Subject<HTMLElement | null>()
    private nextCodeElement = (element: HTMLElement | null) => this.codeElements.next(element)

    /** Emits whenever the ref callback for the blob element is called */
    private blobElements = new Subject<HTMLElement | null>()
    private nextBlobElement = (element: HTMLElement | null) => this.blobElements.next(element)

    /** Emits whenever the ref callback for the hover element is called */
    private hoverOverlayElements = new Subject<HTMLElement | null>()
    private nextOverlayElement = (element: HTMLElement | null) => this.hoverOverlayElements.next(element)

    /** Emits whenever something is hovered in the code */
    private codeMouseOvers = new Subject<React.MouseEvent<HTMLElement>>()
    private nextBlobMouseOver = (event: React.MouseEvent<HTMLElement>) => this.codeMouseOvers.next(event)

    /** Emits whenever something is clicked in the code */
    private codeClicks = new Subject<React.MouseEvent<HTMLElement>>()
    private nextCodeClick = (event: React.MouseEvent<HTMLElement>) => this.codeClicks.next(event)

    /** Emits when the go to definition button was clicked */
    private goToDefinitionClicks = new Subject<React.MouseEvent<HTMLElement>>()
    private nextGoToDefinitionClick = (event: React.MouseEvent<HTMLElement>) => this.goToDefinitionClicks.next(event)

    /** Subscriptions to be disposed on unmout */
    private subscriptions = new Subscription()

    constructor(props: BlobProps) {
        super(props)
        this.state = {
            hoverIsFixed: false,
            clickedGoToDefinition: false,
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
                    const overlayPosition = calculateOverlayPosition(codeElement, target, hoverElement)
                    this.setState({ overlayPosition })
                })
        )

        /** Emits with the position at which a new tooltip is to be shown */
        const filteredTargetPositions = merge(
            codeMouseOverTargets.pipe(filter(() => !this.state.hoverIsFixed)),
            codeClickTargets.pipe(filter(() => this.state.hoverIsFixed))
        ).pipe(
            // Find out the position that was hovered over
            map(({ target, codeElement }) => getTargetLineAndOffset(target, codeElement, false)),
            filter(isDefined),
            share()
        )

        // On every new hover position, fetch new hover contents and update the state
        this.subscriptions.add(
            filteredTargetPositions
                .pipe(
                    switchMap(position => {
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
                    this.setState({ hoverOrError, overlayPosition: undefined })
                })
        )

        // On every new hover position, (pre)fetch definition and update the state
        this.subscriptions.add(
            filteredTargetPositions
                .pipe(
                    // Fetch the hover for that position
                    switchMap(position =>
                        concat(
                            [LOADING],
                            fetchJumpURL({
                                repoPath: this.props.repoPath,
                                commitID: this.props.commitID,
                                filePath: this.props.filePath,
                                position,
                            }).pipe(
                                map(url => (url !== null ? { jumpURL: url } : null)),
                                catchError(error => [asError(error)])
                            )
                        )
                    )
                )
                .subscribe(definitionURLOrError => {
                    this.setState({ definitionURLOrError })
                    // If the j2d button was already clicked and we now have the result, jump to it
                    if (this.state.clickedGoToDefinition && isJumpURL(definitionURLOrError)) {
                        switch (this.state.clickedGoToDefinition) {
                            case 'same-tab':
                                this.props.history.push(definitionURLOrError.jumpURL)
                                break
                            case 'new-tab':
                                window.open(definitionURLOrError.jumpURL, '_blank')
                                break
                        }
                    }
                })
        )

        // On every click on a go to definition button, reveal loader/error/not found UI
        this.subscriptions.add(
            this.goToDefinitionClicks.subscribe(event => {
                // This causes an error/loader/not found UI to get shown if needed
                // Remember if ctrl/cmd was pressed to determine whether the definition should be opened in a new tab once loaded
                this.setState({ clickedGoToDefinition: event.ctrlKey || event.metaKey ? 'new-tab' : 'same-tab' })
                // If we don't have a result yet, prevent default link behaviour (jump will occur dynamically once finished)
                if (!isJumpURL(this.state.definitionURLOrError)) {
                    event.preventDefault()
                }
            })
        )
        this.subscriptions.add(
            filteredTargetPositions.subscribe(hoveredToken => {
                this.setState({
                    hoveredToken,
                    // On every new target (from mouseover or click) hide the j2d loader/error/not found UI again
                    clickedGoToDefinition: false,
                })
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
                                  overlayPosition: undefined,
                              }
                            : {
                                  hoverIsFixed: true,
                              }
                )
            })
        )

        // When the location, reset and hide any existing overlay
        this.subscriptions.add(
            this.componentUpdates.pipe(map(props => props.location), distinctUntilChanged()).subscribe(() => {
                this.setState({
                    clickedGoToDefinition: false,
                    definitionURLOrError: undefined,
                    hoveredToken: undefined,
                    hoverIsFixed: false,
                    hoverOrError: undefined,
                    overlayPosition: undefined,
                })
            })
        )

        // When the line in the location changes, scroll to it
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(props => parseHash(props.location.hash).line),
                    distinctUntilChanged(),
                    filter(isDefined),
                    withLatestFrom(this.blobElements.pipe(filter(isDefined)), this.codeElements.pipe(filter(isDefined)))
                )
                .subscribe(([line, blobElement, codeElement]) => {
                    const blobBound = blobElement.getBoundingClientRect()
                    const codeBound = codeElement.getBoundingClientRect()
                    const tableElement = codeElement.firstElementChild as HTMLTableElement
                    const row = tableElement.rows[line - 1]!
                    const rowBound = row.getBoundingClientRect()
                    const scrollTop = rowBound.top - codeBound.top - blobBound.height / 2 + rowBound.height / 2
                    blobElement.scrollTop = scrollTop
                })
        )
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public shouldComponentUpdate(nextProps: Readonly<BlobProps>, nextState: Readonly<BlobState>): boolean {
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
            <div className={`blob2 ${this.props.className}`} ref={this.nextBlobElement}>
                <code
                    className={`blob2__code ${this.props.wrapCode ? ' blob2__code--wrapped' : ''} `}
                    ref={this.nextCodeElement}
                    dangerouslySetInnerHTML={{ __html: this.props.html }}
                    onClick={this.nextCodeClick}
                    onMouseOver={this.nextBlobMouseOver}
                />
                {this.state.hoverOrError &&
                    !(isHover(this.state.hoverOrError) && isEmptyHover(this.state.hoverOrError)) && (
                        <HoverOverlay
                            hoverRef={this.nextOverlayElement}
                            definitionURLOrError={
                                // always modify the href, but only show error/loader/not found after the button was clicked
                                isJumpURL(this.state.definitionURLOrError) || this.state.clickedGoToDefinition
                                    ? this.state.definitionURLOrError
                                    : undefined
                            }
                            onGoToDefinitionClick={this.nextGoToDefinitionClick}
                            hoverOrError={this.state.hoverOrError}
                            hoveredToken={this.state.hoveredToken}
                            overlayPosition={this.state.overlayPosition}
                            isFixed={this.state.hoverIsFixed}
                            {...this.props}
                        />
                    )}
            </div>
        )
    }
}
