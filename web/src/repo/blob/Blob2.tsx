import * as H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { concat, fromEvent, merge, Observable, ObservableInput, of, Subject, Subscription } from 'rxjs'
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
import { Hover, Position } from 'vscode-languageserver-types'
import { AbsoluteRepoFile, RenderMode } from '..'
import { EMODENOTFOUND, fetchHover, fetchJumpURL, isEmptyHover, isHover } from '../../backend/lsp'
import { eventLogger } from '../../tracking/eventLogger'
import { asError, ErrorLike, isErrorLike } from '../../util/errors'
import { isDefined, propertyIsDefined } from '../../util/types'
import { LineOrPositionOrRange, parseHash, toPositionOrRangeHash } from '../../util/url'
import { BlameLine } from './blame/BlameLine'
import { HoverOverlay, isJumpURL } from './HoverOverlay'
import { convertNode, findElementWithOffset, getTableDataCell, locateTarget } from './tooltips'

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
    return findElementWithOffset(row, position.character)
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
 * Returns true if the HoverOverlay component should be rendered according to the given state.
 * The HoverOverlay is rendered when there is either a non-empty hover result or a non-empty definition result.
 */
const shouldRenderHover = (state: BlobState): boolean =>
    !(!state.hoverOverlayIsFixed && state.mouseIsMoving) &&
    ((state.hoverOrError && !(isHover(state.hoverOrError) && isEmptyHover(state.hoverOrError))) ||
        isJumpURL(state.definitionURLOrError))

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

    /** Emits whenever something is clicked in the code */
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

        // Mouse is moving, don't show the tooltip
        this.subscriptions.add(
            this.codeMouseOvers.subscribe(() => {
                this.setState({ mouseIsMoving: true })
            })
        )

        // Mouse stopped over a token for TOOLTIP_DISPLAY_DELAY, show tooltip
        this.subscriptions.add(
            this.codeMouseOvers.pipe(debounceTime(TOOLTIP_DISPLAY_DELAY)).subscribe(() => {
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

        // When clicking a line, update the URL (which will in turn trigger a highlight of the line)
        this.subscriptions.add(
            this.codeClicks
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

        const codeClickTargets = this.codeClicks.pipe(
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
            filter(Position.is),
            map(position => ({ line: position.line, character: position.character })),
            distinctUntilChanged((a, b) => isEqual(a, b)),
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
                    withLatestFrom(
                        merge(
                            codeMouseOverTargets.pipe(map(data => ({ ...data, source: 'mouseover' as 'mouseover' }))),
                            codeClickTargets.pipe(map(data => ({ ...data, source: 'click' as 'click' }))),
                            targetsFromLocationHash.pipe(map(data => ({ ...data, source: 'location' as 'location' })))
                        )
                    ),
                    map(([, { target, codeElement, source }]) => ({ target, codeElement, source })),
                    // When the new target came from a mouseover, only reposition the hover if it is not fixed
                    filter(({ source }) => source !== 'mouseover' || !this.state.hoverOverlayIsFixed),
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

        /**
         * Emits with the 1-indexed position at which a new tooltip is to be shown from a mouseover, click or location change.
         * Emits `undefined` when a target was hovered/clicked that does not correspond to a position (e.g. after the end of the line).
         */
        const filteredTargetPositions: Observable<{ position?: Position; target: HTMLElement }> = merge(
            merge(
                // When the location changes and and includes a line/column pair, use that target
                targetsFromLocationHash,
                // mouseovers should only trigger a new hover when the overlay is not fixed
                codeMouseOverTargets.pipe(filter(() => !this.state.hoverOverlayIsFixed)),
                // clicks should trigger a new hover when the overlay is fixed
                codeClickTargets.pipe(filter(() => this.state.hoverOverlayIsFixed))
            ).pipe(
                // Find out the position that was hovered over
                map(({ target, codeElement }) => {
                    const hoveredToken = locateTarget(target, codeElement, false)
                    return {
                        target,
                        position: Position.is(hoveredToken) ? hoveredToken : undefined,
                    }
                }),
                distinctUntilChanged((a, b) => isEqual(a.position, b.position))
            )
        ).pipe(share())

        // HOVER FETCH
        // On every new hover position, fetch new hover contents
        const hovers = filteredTargetPositions.pipe(
            switchMap(({ position }): ObservableInput<undefined | Hover | null | ErrorLike | typeof LOADING> => {
                if (!position) {
                    return [undefined]
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
                // Show a loader if it hasn't returned after 100ms
                return merge(hoverFetch, of(LOADING).pipe(delay(LOADER_DELAY), takeUntil(hoverFetch)))
            }),
            share()
        )
        // Update the state
        this.subscriptions.add(
            hovers.subscribe(hoverOrError => {
                this.setState(state => ({
                    hoverOrError,
                    // Reset the hover position, it's gonna be repositioned after the hover was rendered
                    hoverOverlayPosition: undefined,
                    // After the hover is fetched, if the overlay was pinned, unpin it if the hover is empty
                    hoverOverlayIsFixed: state.hoverOverlayIsFixed
                        ? !!hoverOrError || !isHover(hoverOrError) || !isEmptyHover(hoverOrError)
                        : false,
                }))
                // Telemetry
                if (hoverOrError && hoverOrError !== LOADING && !isErrorLike(hoverOrError)) {
                    eventLogger.log('SymbolHovered')
                }
            })
        )
        // Highlight the hover range returned by the language server
        this.subscriptions.add(
            hovers.pipe(withLatestFrom(this.codeElements)).subscribe(([hoverOrError, codeElement]) => {
                const currentHighlighted = codeElement!.querySelector('.selection-highlight')
                if (currentHighlighted) {
                    currentHighlighted.classList.remove('selection-highlight')
                }
                if (!isHover(hoverOrError) || !hoverOrError.range) {
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

        // GO TO DEFINITION FETCH
        // On every new hover position, (pre)fetch definition and update the state
        this.subscriptions.add(
            filteredTargetPositions
                .pipe(
                    // Fetch the definition location for that position
                    switchMap(({ position }) => {
                        if (!position) {
                            return [undefined]
                        }
                        return concat(
                            [LOADING],
                            fetchJumpURL({
                                repoPath: this.props.repoPath,
                                commitID: this.props.commitID,
                                filePath: this.props.filePath,
                                rev: this.props.rev,
                                position,
                            }).pipe(
                                map(url => (url !== null ? { jumpURL: url } : null)),
                                catchError(error => [asError(error)])
                            )
                        )
                    })
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
                // Telemetry
                eventLogger.log('GoToDefClicked')

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
            filteredTargetPositions.subscribe(({ position }) => {
                this.setState({
                    hoveredTokenPosition: position,
                    // On every new target (from mouseover or click) hide the j2d loader/error/not found UI again
                    clickedGoToDefinition: false,
                })
            })
        )

        // HOVER OVERLAY PINNING ON CLICK
        this.subscriptions.add(
            codeClickTargets.subscribe(({ target }) => {
                this.setState({
                    // If a token inside a code cell was clicked, pin the hover
                    // Otherwise if empty space (the cell itself or the <code> element) was clicked, unpin it
                    hoverOverlayIsFixed: !!target.closest('td'),
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
                        // Pin overlay if a concrete position or range is given
                        hoverOverlayIsFixed: position.character !== undefined,
                    })
                    const rows = getRowsInRange(codeElement, position)
                    // Remove existing highlighting
                    for (const selected of codeElement.querySelectorAll('.selected')) {
                        selected.classList.remove('selected')
                    }
                    for (const { line, element } of rows) {
                        const codeCell = element.cells[1]!
                        convertNode(codeCell)
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

        this.setState({
            blameLineIDs: {
                ...this.state.blameLineIDs,
                [line]: id,
            },
        })
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
                    onMouseOver={this.nextCodeMouseOver}
                    data-e2e="blob"
                />
                {shouldRenderHover(this.state) && (
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
