import * as H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { combineLatest, concat, fromEvent, merge, Observable, of, Subject, Subscription, zip } from 'rxjs'
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
    withLatestFrom,
} from 'rxjs/operators'
import { Hover, Position } from 'vscode-languageserver-types'
import { AbsoluteRepoFile, RenderMode } from '..'
import { EMODENOTFOUND, fetchHover, fetchJumpURL, isEmptyHover } from '../../backend/lsp'
import { eventLogger } from '../../tracking/eventLogger'
import { asError, ErrorLike } from '../../util/errors'
import { isDefined, propertyIsDefined } from '../../util/types'
import { LineOrPositionOrRange, parseHash, toPositionOrRangeHash } from '../../util/url'
import { BlameLine } from './blame/BlameLine'
import { HoverOverlay, isJumpURL } from './HoverOverlay'
import { findElementWithOffset, locateTarget } from './tooltips'

/**
 * @param codeElement The `<code>` element
 * @param line 1-indexed line number
 * @return The `<tr>` element
 */
const getRowInCodeElement = (codeElement: HTMLElement, line: number): HTMLTableRowElement | undefined => {
    const table = codeElement.firstElementChild as HTMLTableElement
    return table.rows[line - 1]
}

/**
 * Returns a list of `<tr>` elements that are contained in the given range
 *
 * @param position 1-indexed line, position or inclusive range
 */
const getRowsInRange = (
    codeElement: HTMLElement,
    position?: LineOrPositionOrRange
): {
    /** 1-indexed line number */
    line: number
    /** The `<tr>` element */
    element: HTMLTableRowElement
}[] => {
    if (!position || position.line === undefined) {
        return []
    }
    const tableElement = codeElement.firstElementChild as HTMLTableElement
    const rows: { line: number; element: HTMLTableRowElement }[] = []
    for (let line = position.line; line <= (position.endLine || position.line); line++) {
        const element = tableElement.rows[line - 1]
        if (!element) {
            break
        }
        rows.push({ line, element })
    }
    return rows
}

/**
 * Returns the token `<span>` element in a `<code>` element for a given 1-indexed position.
 *
 * @param codeElement The `<code>` element
 * @param position 1-indexed position
 */
const getTokenAtPosition = (codeElement: HTMLElement, position: Position): HTMLElement | undefined => {
    const row = getRowInCodeElement(codeElement, position.line)
    if (!row) {
        return undefined
    }
    const [, codeCell] = row.cells
    return findElementWithOffset(codeCell, position.character)
}

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
        tooltipTop = relBottom
    } else {
        tooltipTop -= BLOB_PADDING_TOP
    }
    return { left: relLeft, top: tooltipTop }
}

/**
 * Scrolls an element to the center if it is out of view.
 * Does nothing if the element is in view.
 *
 * @param container The scrollable container (that has `overflow: auto`)
 * @param content The content child that is being scrolled
 * @param target The element that should be scrolled into view
 */
const scrollIntoCenterIfNeeded = (container: HTMLElement, content: HTMLElement, target: HTMLElement): void => {
    const blobRect = container.getBoundingClientRect()
    const rowRect = target.getBoundingClientRect()
    if (rowRect.top <= blobRect.top || rowRect.bottom >= blobRect.bottom) {
        const blobRect = container.getBoundingClientRect()
        const contentRect = content.getBoundingClientRect()
        const rowRect = target.getBoundingClientRect()
        const scrollTop = rowRect.top - contentRect.top - blobRect.height / 2 + rowRect.height / 2
        container.scrollTop = scrollTop
    }
}

/**
 * toPortalID builds an ID that will be used for the blame portal containers.
 */
const toPortalID = (line: number) => `blame-portal-${line}`

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
    definitionURLOrError?: typeof LOADING | { jumpURL: string } | null | ErrorLike
    hoverOverlayIsFixed: boolean

    /** The desired position of the hover overlay */
    hoverOverlayPosition?: { left: number; top: number }

    /**
     * Whether the user has clicked the go to definition button for the current overlay yet,
     * and whether he pressed Ctrl/Cmd while doing it to open it in a new tab or not.
     */
    clickedGoToDefinition: false | 'same-tab' | 'new-tab'

    /** The currently hovered token */
    hoveredTokenPosition?: Position

    /**
     * blameLineIDs is a map from line numbers with portal nodes created to portal IDs.
     * It's used to render the portals for blames. The line numbers are taken from the blob
     * so they are 1-indexed.
     */
    blameLineIDs: { [key: number]: string }

    /**
     * The currently selected position, if any.
     * Can be a single line number or a line range.
     * Highlighted with a background color.
     */
    selectedPosition?: LineOrPositionOrRange

    mouseIsMoving: boolean
}

/**
 * Returns true if the HoverOverlay would have anything to show according to the given hover and definition states.
 */
const overlayUIHasContent = (state: Pick<BlobState, 'hoverOrError' | 'definitionURLOrError'>): boolean =>
    (state.hoverOrError && !(Hover.is(state.hoverOrError) && isEmptyHover(state.hoverOrError))) ||
    isJumpURL(state.definitionURLOrError)

/**
 * Returns true if the HoverOverlay component should be rendered according to the given state.
 */
const shouldRenderOverlay = (state: BlobState): boolean =>
    !(!state.hoverOverlayIsFixed && state.mouseIsMoving) && overlayUIHasContent(state)

/** The time in ms after which to show a loader if the result has not returned yet */
const LOADER_DELAY = 300

/** The time in ms after the mouse has stopped moving in which to show the tooltip */
const TOOLTIP_DISPLAY_DELAY = 100

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
    private nextCodeMouseOver = (event: React.MouseEvent<HTMLElement>) => this.codeMouseOvers.next(event)

    /** Emits whenever something is hovered in the code */
    private codeMouseMoves = new Subject<React.MouseEvent<HTMLElement>>()
    private nextCodeMouseMove = (event: React.MouseEvent<HTMLElement>) => this.codeMouseMoves.next(event)

    /**
     * Emits whenever something is clicked in the code.
     * Note that this also fires when the user selects text, see `codeClicksWithoutSelection` further down.
     */
    private codeClicks = new Subject<React.MouseEvent<HTMLElement>>()
    private nextCodeClick = (event: React.MouseEvent<HTMLElement>) => {
        event.persist()
        this.codeClicks.next(event)
    }

    /** Emits when the go to definition button was clicked */
    private goToDefinitionClicks = new Subject<React.MouseEvent<HTMLElement>>()
    private nextGoToDefinitionClick = (event: React.MouseEvent<HTMLElement>) => this.goToDefinitionClicks.next(event)

    /** Emits when the close button was clicked */
    private closeButtonClicks = new Subject<React.MouseEvent<HTMLElement>>()
    private nextCloseButtonClick = (event: React.MouseEvent<HTMLElement>) => this.closeButtonClicks.next(event)

    /** Subscriptions to be disposed on unmout */
    private subscriptions = new Subscription()

    constructor(props: BlobProps) {
        super(props)
        this.state = {
            hoverOverlayIsFixed: false,
            clickedGoToDefinition: false,
            blameLineIDs: {},
            mouseIsMoving: false,
        }

        /**
         * click events on the code element, ignoring click events caused by the user selecting text.
         * Selecting text should not mess with the hover, hover pinning nor the URL.
         */
        const codeClicksWithoutSelections = this.codeClicks.pipe(filter(() => window.getSelection().toString() === ''))

        // Mouse is moving, don't show the tooltip
        this.subscriptions.add(
            this.codeMouseMoves
                .pipe(
                    map(event => event.target),
                    // Make sure a move of the mouse from the go-to-definition button
                    // back to the same target doesn't cause the tooltip to briefly disappear
                    distinctUntilChanged()
                )
                .subscribe(() => {
                    this.setState({ mouseIsMoving: true })
                })
        )

        // When the mouse stopped for TOOLTIP_DISPLAY_DELAY, show tooltip
        // Don't use mouseover for this because it is only fired once per token,
        // not continuously while moving the mouse
        this.subscriptions.add(
            this.codeMouseMoves.pipe(debounceTime(TOOLTIP_DISPLAY_DELAY)).subscribe(() => {
                this.setState({ mouseIsMoving: false })
            })
        )

        const codeMouseOverTargets = this.codeMouseOvers.pipe(
            map(event => event.target as HTMLElement),
            // Casting is okay here, we know these are HTMLElements
            withLatestFrom(this.codeElements),
            // If there was a mouseover, there _must_ have been a blob element
            map(([target, codeElement]) => ({ target, codeElement: codeElement! })),
            debounceTime(50),
            // Do not consider mouseovers while overlay is pinned
            filter(() => !this.state.hoverOverlayIsFixed),
            share()
        )

        // When clicking a line, update the URL (which will in turn trigger a highlight of the line)
        this.subscriptions.add(
            codeClicksWithoutSelections
                .pipe(
                    withLatestFrom(this.codeElements),
                    map(([event, codeElement]) => ({
                        event,
                        position: locateTarget(event.target as HTMLElement, codeElement!, false),
                    }))
                )
                .subscribe(({ event, position }) => {
                    let hash: string
                    if (
                        position &&
                        event.shiftKey &&
                        this.state.selectedPosition &&
                        this.state.selectedPosition.line !== undefined
                    ) {
                        hash = toPositionOrRangeHash({
                            range: {
                                start: {
                                    line: Math.min(this.state.selectedPosition.line, position.line),
                                },
                                end: {
                                    line: Math.max(this.state.selectedPosition.line, position.line),
                                },
                            },
                        })
                    } else {
                        hash = toPositionOrRangeHash({ position })
                    }

                    if (!hash.startsWith('#')) {
                        hash = '#' + hash
                    }

                    this.props.history.push({ ...this.props.location, hash })
                })
        )

        const codeClickTargets = codeClicksWithoutSelections.pipe(
            map(event => event.target as HTMLElement),
            withLatestFrom(this.codeElements),
            // If there was a click, there _must_ have been a blob element
            map(([target, codeElement]) => ({ target, codeElement: codeElement! })),
            share()
        )

        /** Emits parsed positions found in the URL */
        const locationPositions = this.componentUpdates.pipe(map(props => parseHash(props.location.hash)), share())

        /** Emits DOM elements at new positions found in the URL */
        const targetsFromLocationHash: Observable<{
            target: HTMLElement
            codeElement: HTMLElement
        }> = locationPositions.pipe(
            // Make sure to pick only the line and character to compare
            map(position => ({ line: position.line, character: position.character })),
            // Ignore same values
            // It's important to do this before filtering otherwise navigating from
            // a position, to a line-only position, back to the first position would get ignored
            distinctUntilChanged((a, b) => isEqual(a, b)),
            // Ignore undefined or partial positions (e.g. line only)
            filter(Position.is),
            withLatestFrom(this.codeElements),
            map(([position, codeElement]) => ({ position, codeElement })),
            filter(propertyIsDefined('codeElement')),
            map(({ position, codeElement }) => {
                const row = getRowInCodeElement(codeElement, position.line)
                if (!row) {
                    return { codeElement }
                }
                const cell = row.cells[1]!
                const target = findElementWithOffset(cell, position.character)
                if (!target) {
                    console.warn('Could not find target for position in file', position)
                }
                return { target, codeElement }
            }),
            filter(propertyIsDefined('target'))
        )

        // REPOSITIONING
        // On every componentDidUpdate (after the component was rerendered, e.g. from a hover state update) resposition
        // the tooltip
        // It's important to add this subscription first so that withLatestFrom will be guaranteed to have gotten the
        // latest hover target by the time componentDidUpdate is triggered from the setState() in the second chain
        this.subscriptions.add(
            // Take every rerender
            this.componentUpdates
                .pipe(
                    // with the latest target that came from either a mouseover, click or location change (whatever was the most recent)
                    withLatestFrom(merge(codeMouseOverTargets, codeClickTargets, targetsFromLocationHash)),
                    map(([, { target, codeElement }]) => ({ target, codeElement })),
                    withLatestFrom(this.hoverOverlayElements),
                    map(([{ target, codeElement }, hoverElement]) => ({ target, hoverElement, codeElement })),
                    filter(propertyIsDefined('hoverElement'))
                )
                .subscribe(({ codeElement, hoverElement, target }) => {
                    const hoverOverlayPosition = calculateOverlayPosition(
                        codeElement.parentElement!, // ! because we know its there
                        target,
                        hoverElement
                    )
                    this.setState({ hoverOverlayPosition })
                })
        )

        /** Emits new positions at which a tooltip needs to be shown from clicks, mouseovers and URL changes. */
        const positions: Observable<{
            /**
             * The 1-indexed position at which a new tooltip is to be shown,
             * or undefined when a target was hovered/clicked that does not correspond to a position (e.g. after the end of the line)
             */
            position?: Position
            /**
             * True if the tooltip should be pinned once the hover came back and is non-empty.
             * This depends on what triggered the new position.
             * We remember it because the pinning is deferred to when we have a result,
             * so we don't pin empty (i.e. invisible) hovers.
             */
            pinIfNonEmpty: boolean
        }> = merge(
            merge(
                // Should unpin the tooltip even if hover cames back non-empty
                codeMouseOverTargets.pipe(map(data => ({ ...data, pinIfNonEmpty: false }))),
                // When the location changes and includes a line/column pair, use that target
                // Should pin the tooltip if hover cames back non-empty
                targetsFromLocationHash.pipe(map(data => ({ ...data, pinIfNonEmpty: true }))),
                // Should pin the tooltip if hover cames back non-empty
                codeClickTargets.pipe(map(data => ({ ...data, pinIfNonEmpty: true })))
            ).pipe(
                // Find out the position that was hovered over
                map(({ target, codeElement, pinIfNonEmpty }) => {
                    const hoveredToken = locateTarget(target, codeElement, false)
                    const position = Position.is(hoveredToken) ? hoveredToken : undefined
                    return { position, pinIfNonEmpty }
                })
            )
        ).pipe(share())

        /**
         * For every position, emits an Observable with new values for the `hoverOrError` state.
         * This is a higher-order Observable (Observable that emits Observables).
         */
        const hoverObservables = positions.pipe(
            map(({ position }) => {
                if (!position) {
                    return of(undefined)
                }
                // Fetch the hover for that position
                const hoverFetch = fetchHover({
                    repoPath: this.props.repoPath,
                    commitID: this.props.commitID,
                    filePath: this.props.filePath,
                    rev: this.props.rev,
                    position,
                }).pipe(
                    catchError(error => {
                        if (error && error.code === EMODENOTFOUND) {
                            return [undefined]
                        }
                        return [asError(error)]
                    }),
                    share()
                )
                // 1. Reset the hover content, so no old hover content is displayed at the new position while fetching
                // 2. Show a loader if the hover fetch hasn't returned after 100ms
                // 3. Show the hover once it returned
                return merge([undefined], of(LOADING).pipe(delay(LOADER_DELAY), takeUntil(hoverFetch)), hoverFetch)
            }),
            share()
        )
        /** Flattened `hoverObservables` */
        const hovers = hoverObservables.pipe(switchMap(hoverObservable => hoverObservable), share())

        this.subscriptions.add(
            hovers.subscribe(hoverOrError => {
                // Update the state
                this.setState({
                    hoverOrError,
                    // Reset the hover position, it's gonna be repositioned after the hover was rendered
                    hoverOverlayPosition: undefined,
                })
            })
        )
        // Highlight the hover range returned by the language server
        this.subscriptions.add(
            hovers.pipe(withLatestFrom(this.codeElements)).subscribe(([hoverOrError, codeElement]) => {
                const currentHighlighted = codeElement!.querySelector('.selection-highlight')
                if (currentHighlighted) {
                    currentHighlighted.classList.remove('selection-highlight')
                }
                if (!Hover.is(hoverOrError) || !hoverOrError.range) {
                    return
                }
                // LSP is 0-indexed, the code in the webapp currently is 1-indexed
                const { line, character } = hoverOrError.range.start
                const token = getTokenAtPosition(codeElement!, { line: line + 1, character: character + 1 })
                if (!token) {
                    return
                }
                token.classList.add('selection-highlight')
            })
        )
        // Telemetry for hovers
        this.subscriptions.add(
            zip(positions, hoverObservables)
                .pipe(
                    distinctUntilChanged(([positionA], [positionB]) => isEqual(positionA, positionB)),
                    switchMap(([position, hoverObservable]) => hoverObservable),
                    filter(Hover.is)
                )
                .subscribe(() => {
                    eventLogger.log('SymbolHovered')
                })
        )

        /**
         * For every position, emits an Observable that emits new values for the `definitionURLOrError` state.
         * This is a higher-order Observable (Observable that emits Observables).
         */
        const definitionObservables = positions.pipe(
            // Fetch the definition location for that position
            map(({ position }) => {
                if (!position) {
                    return of(undefined)
                }
                return concat(
                    [LOADING],
                    fetchJumpURL({
                        repoPath: this.props.repoPath,
                        commitID: this.props.commitID,
                        filePath: this.props.filePath,
                        rev: this.props.rev,
                        position,
                    }).pipe(map(url => (url !== null ? { jumpURL: url } : null)), catchError(error => [asError(error)]))
                )
            })
        )

        // GO TO DEFINITION FETCH
        // On every new hover position, (pre)fetch definition and update the state
        this.subscriptions.add(
            definitionObservables
                // flatten inner Observables
                .pipe(switchMap(definitionObservable => definitionObservable))
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

        // DEFERRED HOVER OVERLAY PINNING
        // If the new position came from a click or the URL,
        // if either the hover or the definition turn out non-empty, pin the tooltip.
        // If they both turn out empty, unpin it so we don't end up with an invisible tooltip.
        //
        // zip together a position and the hover and definition fetches it triggered
        zip(positions, hoverObservables, definitionObservables)
            .pipe(
                switchMap(([{ pinIfNonEmpty }, hoverObservable, definitionObservable]) => {
                    // If the position was triggered by a mouseover, never pin
                    if (!pinIfNonEmpty) {
                        return [false]
                    }
                    // combine the latest values for them, so we have access to both values
                    // and can reevaluate our pinning decision whenever one of the two updates,
                    // independent of the order in which they emit
                    return combineLatest(hoverObservable, definitionObservable).pipe(
                        map(([hoverOrError, definitionURLOrError]) =>
                            overlayUIHasContent({ hoverOrError, definitionURLOrError })
                        )
                    )
                })
            )
            .subscribe(hoverOverlayIsFixed => {
                this.setState({ hoverOverlayIsFixed })
            })

        // On every click on a go to definition button, reveal loader/error/not found UI
        this.subscriptions.add(
            this.goToDefinitionClicks.subscribe(event => {
                // Telemetry
                eventLogger.log('GoToDefClicked')

                // If we don't have a result yet that would be jumped to by the native <a> tag...
                if (!isJumpURL(this.state.definitionURLOrError)) {
                    // Prevent default link behaviour (jump will be done programmatically once finished)
                    event.preventDefault()

                    // Remember if ctrl/cmd was pressed to determine whether the definition should be opened in a new tab once loaded
                    // Also causes an error/loader/not found UI to get shown if needed
                    this.setState({ clickedGoToDefinition: event.ctrlKey || event.metaKey ? 'new-tab' : 'same-tab' })
                }
            })
        )
        this.subscriptions.add(
            positions.subscribe(({ position }) => {
                this.setState({
                    hoveredTokenPosition: position,
                    // On every new target (from mouseover or click) hide the j2d loader/error/not found UI again
                    clickedGoToDefinition: false,
                })
            })
        )

        // When the close button is clicked, unpin, hide and reset the hover
        this.subscriptions.add(
            merge(
                this.closeButtonClicks,
                fromEvent<KeyboardEvent>(window, 'keydown').pipe(filter(event => event.key === 'Escape'))
            ).subscribe(event => {
                event.preventDefault()
                this.setState({
                    hoverOverlayIsFixed: false,
                    hoverOverlayPosition: undefined,
                    hoverOrError: undefined,
                    hoveredTokenPosition: undefined,
                    definitionURLOrError: undefined,
                    clickedGoToDefinition: false,
                })
            })
        )

        // LOCATION CHANGES
        this.subscriptions.add(
            locationPositions
                .pipe(
                    distinctUntilChanged((a, b) => isEqual(a, b)),
                    withLatestFrom(this.blobElements.pipe(filter(isDefined)), this.codeElements.pipe(filter(isDefined)))
                )
                .subscribe(([position, blobElement, codeElement]) => {
                    this.setState({
                        // Remember active position in state for blame and range expansion
                        selectedPosition: position,
                    })
                    const rows = getRowsInRange(codeElement, position)
                    // Remove existing highlighting
                    for (const selected of codeElement.querySelectorAll('.selected')) {
                        selected.classList.remove('selected')
                    }
                    for (const { line, element } of rows) {
                        const codeCell = element.cells[1]!
                        this.createBlameDomNode(line, codeCell)
                        // Highlight row
                        element.classList.add('selected')
                    }
                    // Scroll into view
                    if (rows.length > 0) {
                        scrollIntoCenterIfNeeded(blobElement, codeElement, rows[0].element)
                    }
                })
        )
    }

    /**
     * Appends a blame portal DOM node to the given code cell if it doesn't contain one already.
     *
     * @param line 1-indexed line number
     * @param codeCell The `<td class="code">` element
     */
    private createBlameDomNode(line: number, codeCell: HTMLElement): void {
        if (codeCell.querySelector('.blame-portal')) {
            return
        }
        const portalNode = document.createElement('span')

        const id = toPortalID(line)
        portalNode.id = id
        portalNode.classList.add('blame-portal')

        codeCell.appendChild(portalNode)

        this.setState(state => ({
            blameLineIDs: {
                ...state.blameLineIDs,
                [line]: id,
            },
        }))
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
                    className={`blob2__code ${this.props.wrapCode ? ' blob2__code--wrapped' : ''} e2e-blob`}
                    ref={this.nextCodeElement}
                    dangerouslySetInnerHTML={{ __html: this.props.html }}
                    onClick={this.nextCodeClick}
                    onMouseOver={this.nextCodeMouseOver}
                    onMouseMove={this.nextCodeMouseMove}
                />
                {shouldRenderOverlay(this.state) && (
                    <HoverOverlay
                        hoverRef={this.nextOverlayElement}
                        definitionURLOrError={
                            // always modify the href, but only show error/loader/not found after the button was clicked
                            isJumpURL(this.state.definitionURLOrError) || this.state.clickedGoToDefinition
                                ? this.state.definitionURLOrError
                                : undefined
                        }
                        onGoToDefinitionClick={this.nextGoToDefinitionClick}
                        onCloseButtonClick={this.nextCloseButtonClick}
                        hoverOrError={this.state.hoverOrError}
                        hoveredTokenPosition={this.state.hoveredTokenPosition}
                        overlayPosition={this.state.hoverOverlayPosition}
                        showCloseButton={this.state.hoverOverlayIsFixed}
                        {...this.props}
                    />
                )}
                {this.state.selectedPosition &&
                    this.state.selectedPosition.line !== undefined &&
                    this.state.blameLineIDs[this.state.selectedPosition.line] && (
                        <BlameLine
                            key={this.state.blameLineIDs[this.state.selectedPosition.line]}
                            portalID={this.state.blameLineIDs[this.state.selectedPosition.line]}
                            line={this.state.selectedPosition.line}
                            {...this.props}
                        />
                    )}
            </div>
        )
    }
}
