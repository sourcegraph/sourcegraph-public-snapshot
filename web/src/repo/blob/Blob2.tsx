import * as H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { concat, fromEvent, merge, Observable, of, Subject, Subscription } from 'rxjs'
import {
    catchError,
    debounceTime,
    delay,
    distinctUntilChanged,
    filter,
    first,
    map,
    share,
    switchMap,
    takeUntil,
    tap,
    withLatestFrom,
} from 'rxjs/operators'
import { Hover, Position } from 'vscode-languageserver-types'
import { AbsoluteRepoFile, RenderMode } from '..'
import { EMODENOTFOUND, fetchHover, fetchJumpURL, isEmptyHover } from '../../backend/lsp'
import { eventLogger } from '../../tracking/eventLogger'
import { scrollIntoView } from '../../util'
import { asError, ErrorLike, isErrorLike } from '../../util/errors'
import { isDefined, propertyIsDefined } from '../../util/types'
import { parseHash, toPositionOrRangeHash } from '../../util/url'
import { BlameLine } from './blame/BlameLine'
import { HoverOverlay, isJumpURL } from './HoverOverlay'
import { convertNode, findElementWithOffset, getTableDataCell, getTargetLineAndOffset } from './tooltips'

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

interface HightlightArgs {
    /** The `<code>` element */
    codeElement: HTMLElement
    /** The table row that represents the new line */
    line?: HTMLTableRowElement
}

/**
 * Sets a new line to be highlighted and unhighlights the previous highlighted line, if exists.
 */
const highlightLine = ({ line, codeElement }: HightlightArgs): void => {
    const current = codeElement.querySelector('.selected')
    if (current) {
        current.classList.remove('selected')
    }
    if (line === undefined) {
        return
    }

    line.classList.add('selected')
}

const getTextNodes = (node: Node): Node[] => {
    if (node.nodeType === node.TEXT_NODE) {
        return [node]
    }

    const nodes: Node[] = []
    for (const child of Array.from(node.childNodes)) {
        nodes.push(...getTextNodes(child))
    }

    return nodes
}

const findTokenToHighlight = (position: Position, node: Node): HTMLElement | null => {
    const textNodes = getTextNodes(node)

    let activeNode: Node | null = null

    let offset = 0
    for (let i = 0; i < textNodes.length; i++) {
        const n = textNodes[i]

        if (n.nodeValue) {
            offset += n.nodeValue.length
        }

        if (offset + 1 === position.character) {
            activeNode = textNodes[i + 1]
            break
        }
    }

    if (!activeNode) {
        return null
    }

    return activeNode.parentElement
}

const scrollToCenter = (blobElement: HTMLElement, codeElement: HTMLElement, tableRow: HTMLElement) => {
    // if theres a position hash on page load, scroll it to the center of the screen
    const blobBound = blobElement.getBoundingClientRect()
    const codeBound = codeElement.getBoundingClientRect()
    const rowBound = tableRow.getBoundingClientRect()
    const scrollTop = rowBound.top - codeBound.top - blobBound.height / 2 + rowBound.height / 2

    blobElement.scrollTop = scrollTop
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

const isHover = (val: any): val is Hover => typeof val === 'object' && val !== null && Array.isArray(val.contents)

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

    activeLine: number | null

    mouseIsMoving: boolean
}

/**
 * Returns true if the HoverOverlay component should be rendered according to the given state.
 * The HoverOverlay is rendered when there is either a non-empty hover result or a non-empty definition result.
 */
const shouldRenderHover = (state: BlobState): boolean =>
    !state.mouseIsMoving &&
    ((state.hoverOrError && !(isHover(state.hoverOrError) && isEmptyHover(state.hoverOrError))) ||
        isJumpURL(state.definitionURLOrError))

/** The time in ms after which to show a loader if the result has not returned yet */
const LOADER_DELAY = 100

/** The time in ms after the mouse has stopped moving in which to show the tooltip */
const TOOLTIP_DISPLAY_DELAY = 500

export class Blob2 extends React.Component<BlobProps, BlobState> {
    /** Emits with the latest Props on every componentDidUpdate and on componentDidMount */
    private componentUpdates = new Subject<BlobProps>()

    private componentStateUpdates = new Subject<BlobState>()

    /** Emits whenever the ref callback for the code element is called */
    private codeElements = new Subject<HTMLElement | null>()
    private nextCodeElement = (element: HTMLElement | null) => this.codeElements.next(element)

    /** Emits whenever the ref callback for the blob element is called */
    private blobElements = new Subject<HTMLElement | null>()
    private nextBlobElement = (element: HTMLElement | null) => this.blobElements.next(element)

    /** Emits whenever the ref callback for the hover element is called */
    private hoverOverlayElements = new Subject<HTMLElement | null>()
    private nextOverlayElement = (element: HTMLElement | null) => this.hoverOverlayElements.next(element)

    /** Emits whenever the ref callback for the blob element is called */
    private highlightedElements = new Subject<HTMLElement | null>()
    private nextHighlightedElement = (element: HTMLElement | null) => this.highlightedElements.next(element)

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
    private closeButtonClicks = new Subject<void>()
    private nextCloseButtonClick = () => this.closeButtonClicks.next()

    /** Subscriptions to be disposed on unmout */
    private subscriptions = new Subscription()

    /** logs a hover event, if the hover is valid */
    private logHover(hoverOrError?: typeof LOADING | Hover | null | ErrorLike): void {
        if (hoverOrError && hoverOrError !== LOADING && !isErrorLike(hoverOrError)) {
            eventLogger.log('SymbolHovered')
        }
    }

    constructor(props: BlobProps) {
        super(props)
        this.state = {
            hoverOverlayIsFixed: false,
            clickedGoToDefinition: false,
            blameLineIDs: {},
            activeLine: null,
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

        /**
         * lineClickElements gets the full row that was clicked, the line number cell, the code cell,
         * and the line number itself.
         */
        const lineClickElements = this.codeClicks.pipe(
            map(({ target }) => target as HTMLElement),
            map(target => {
                let row: HTMLElement | null = target
                while (row.parentElement && row.tagName !== 'TR') {
                    row = row.parentElement
                }
                return { target, row }
            }),
            filter(propertyIsDefined('row')),
            map(({ target, row }) => ({
                target,
                row: row as HTMLElement,
                lineNumCell: row.children.item(0) as HTMLElement,
                codeCell: row.children.item(1) as HTMLElement,
            })),
            map(({ lineNumCell, ...rest }) => {
                let lineNum: number | null = null

                const data = lineNumCell.dataset
                lineNum = parseInt(data.line!, 10)

                return {
                    lineNum,
                    lineNumCell,
                    ...rest,
                }
            })
        )

        // Highlight the clicked row
        this.subscriptions.add(
            lineClickElements
                .pipe(
                    withLatestFrom(this.codeElements),
                    map(([{ row }, codeElement]) => ({ line: row, codeElement })),
                    filter(propertyIsDefined('codeElement'))
                )
                .subscribe(highlightLine)
        )

        // Unhighlight old highlighted token when new tokens are hovered over
        this.subscriptions.add(
            this.codeMouseOvers
                .pipe(withLatestFrom(this.highlightedElements.pipe(filter(isDefined))))
                .subscribe(([, highlightedToken]) => {
                    const highlighted = document.querySelectorAll('.selection-highlight')
                    for (const h of Array.from(highlighted)) {
                        if (this.state.hoverOverlayIsFixed && h === highlightedToken) {
                            continue
                        }
                        h.classList.remove('selection-highlight')
                    }
                })
        )

        // When clicking a line, update the URL (which will in turn trigger a highlight of the line)
        this.subscriptions.add(
            lineClickElements
                .pipe(
                    withLatestFrom(this.codeElements),
                    map(([{ target, lineNum }, codeElement]) => ({ target, lineNum, codeElement })),
                    filter(propertyIsDefined('codeElement')),
                    map(({ target, lineNum, codeElement }) => ({
                        lineNum,
                        position: getTargetLineAndOffset(target, codeElement!, false),
                    }))
                )
                .subscribe(({ position, lineNum }) => {
                    let hash: string
                    if (position !== undefined) {
                        hash = toPositionOrRangeHash({ position })
                    } else {
                        hash = `#L${lineNum}`
                    }

                    if (!hash.startsWith('#')) {
                        hash = '#' + hash
                    }

                    this.props.history.push(hash)
                })
        )

        const codeClickTargets = this.codeClicks.pipe(
            map(event => event.target as HTMLElement),
            withLatestFrom(this.codeElements),
            // If there was a click, there _must_ have been a blob element
            map(([target, codeElement]) => ({ target, codeElement: codeElement! })),
            share()
        )

        /** Emits new positions found in the URL */
        const positionsFromLocationHash: Observable<Position> = this.componentUpdates.pipe(
            map(props => parseHash(props.location.hash)),
            filter(Position.is),
            map(position => ({ line: position.line, character: position.character })),
            distinctUntilChanged((a, b) => isEqual(a, b)),
            share()
        )

        // Fix tooltip if a location includes a position
        this.subscriptions.add(
            positionsFromLocationHash.subscribe(() => {
                this.setState({ hoverOverlayIsFixed: true })
            })
        )

        /** Emits DOM elements at new positions found in the URL */
        const targetsFromLocationHash: Observable<{
            target: HTMLElement
            codeElement: HTMLElement
        }> = positionsFromLocationHash.pipe(
            withLatestFrom(this.codeElements),
            map(([position, codeElement]) => ({ position, codeElement })),
            filter(propertyIsDefined('codeElement')),
            map(({ position, codeElement }) => {
                const table = codeElement.firstElementChild as HTMLTableElement
                const row = table.rows[position.line - 1]
                if (!row) {
                    alert(`Could not find line ${position.line} in file`)
                    return { codeElement }
                }
                const cell = row.cells[1]
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

        // Add a dom node for the blame portals
        this.subscriptions.add(lineClickElements.subscribe(this.createBlameDomNode))

        // Set the currently active line from hover
        this.subscriptions.add(
            lineClickElements.pipe(map(({ lineNum }) => lineNum)).subscribe(activeLine => {
                this.setState({ activeLine })
            })
        )

        /**
         * Emits with the position at which a new tooltip is to be shown from a mouseover, click or location change.
         * Emits `undefined` when a target was hovered/clicked that does not correspond to a position (e.g. after the end of the line).
         */
        const filteredTargetPositions: Observable<Position | undefined> = merge(
            // When the location changes and and includes a line/column pair, use that position
            positionsFromLocationHash,
            merge(
                // mouseovers should only trigger a new hover when the overlay is not fixed
                codeMouseOverTargets.pipe(filter(() => !this.state.hoverOverlayIsFixed)),
                // clicks should trigger a new hover when the overlay is fixed
                codeClickTargets.pipe(filter(() => this.state.hoverOverlayIsFixed))
            ).pipe(
                // Find out the position that was hovered over
                map(({ target, codeElement }) => getTargetLineAndOffset(target, codeElement, false)),
                map(position => position && { line: position.line, character: position.character })
            )
        ).pipe(share())

        // HOVER FETCH
        // On every new hover position, fetch new hover contents and update the state
        this.subscriptions.add(
            filteredTargetPositions
                .pipe(
                    distinctUntilChanged((a, b) => isEqual(a, b)),
                    switchMap(position => {
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
                    })
                )
                .subscribe(hoverOrError => {
                    this.logHover(hoverOrError)
                    this.setState(state => ({
                        hoverOrError,
                        // Reset the hover position, it's gonna be repositioned after the hover was rendered
                        hoverOverlayPosition: undefined,
                    }))
                })
        )

        // GO TO DEFINITION FETCH
        // On every new hover position, (pre)fetch definition and update the state
        this.subscriptions.add(
            filteredTargetPositions
                .pipe(
                    // Fetch the definition location for that position
                    switchMap(position => {
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
                this.setState({
                    // This causes an error/loader/not found UI to get shown if needed
                    // Remember if ctrl/cmd was pressed to determine whether the definition should be opened in a new tab once loaded
                    clickedGoToDefinition: event.ctrlKey || event.metaKey ? 'new-tab' : 'same-tab',
                })
                // If we don't have a result yet, prevent default link behaviour (jump will occur dynamically once finished)
                if (!isJumpURL(this.state.definitionURLOrError)) {
                    event.preventDefault()
                }
            })
        )
        this.subscriptions.add(
            filteredTargetPositions.subscribe(hoveredTokenPosition => {
                this.setState({
                    hoveredTokenPosition,
                    // On every new target (from mouseover or click) hide the j2d loader/error/not found UI again
                    clickedGoToDefinition: false,
                })
            })
        )

        // HOVER OVERLAY PINNING ON CLICK
        this.subscriptions.add(
            codeClickTargets.subscribe(({ target, codeElement }) => {
                this.setState({
                    // If a token inside a code cell was clicked, pin the hover
                    // Otherwise if empty space was clicked, unpin it
                    hoverOverlayIsFixed: !target.matches('td'),
                })
            })
        )

        this.subscriptions.add(
            codeClickTargets
                .pipe(withLatestFrom(this.highlightedElements.pipe(filter(isDefined))))
                .subscribe(([{ target }, highlightedToken]) => {
                    if (target !== highlightedToken || !target.contains(highlightedToken)) {
                        highlightedToken.classList.remove('selection-highlight')
                    }
                })
        )

        // When the close button is clicked, unpin, hide and reset the hover
        this.subscriptions.add(
            this.closeButtonClicks.subscribe(() => {
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

        // When the blob loads, highlight the active line and scroll it to center of viewport
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(props => {
                        const pos = parseHash(props.location.hash)

                        return { line: pos.line, character: pos.character }
                    }),
                    distinctUntilChanged(),
                    filter(propertyIsDefined('line')),
                    withLatestFrom(this.codeElements.pipe(filter(isDefined))),
                    map(([position, codeElement]) => ({ position, codeElement })),
                    map(({ position, codeElement }) => {
                        const lineElem = codeElement.querySelector(`td[data-line="${position.line}"]`)
                        if (lineElem && lineElem.parentElement) {
                            return { position, codeElement, tableRow: lineElem.parentElement as HTMLTableRowElement }
                        }

                        return { position, codeElement, tableRow: undefined }
                    }),
                    filter(propertyIsDefined('tableRow')),
                    first(({ tableRow }) => !!tableRow),
                    withLatestFrom(this.blobElements.pipe(filter(isDefined)))
                )
                .subscribe(([{ position, tableRow, codeElement }, blobElement]) => {
                    highlightLine({ codeElement, line: tableRow })
                    this.createBlameDomNode({
                        lineNum: position.line,
                        codeCell: tableRow.children.item(1) as HTMLElement,
                    })

                    // if theres a position hash on page load, scroll it to the center of the screen
                    scrollToCenter(blobElement, codeElement, tableRow)
                    this.setState({ activeLine: position.line, hoverOverlayIsFixed: position.character !== undefined })
                })
        )

        // When the line in the location changes, scroll to it
        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    map(props => parseHash(props.location.hash).line),
                    distinctUntilChanged(),
                    withLatestFrom(this.blobElements.pipe(filter(isDefined)), this.codeElements.pipe(filter(isDefined)))
                )
                .subscribe(([line, blobElement, codeElement]) => {
                    const tableElement = codeElement.firstElementChild as HTMLTableElement
                    if (line !== undefined) {
                        const row = tableElement.rows[line - 1]
                        if (!row) {
                            return
                        }

                        if (this.state.clickedGoToDefinition) {
                            scrollToCenter(blobElement, codeElement, row)
                            highlightLine({ codeElement, line: row })
                        } else {
                            scrollIntoView(blobElement, row)
                        }
                    }
                })
        )

        this.subscriptions.add(
            this.componentStateUpdates
                .pipe(
                    filter(propertyIsDefined('hoveredTokenPosition')),
                    filter(propertyIsDefined('hoverOrError')),
                    filter(propertyIsDefined('hoverOverlayPosition')),
                    filter(({ hoverOrError }) => !(hoverOrError instanceof Error)),
                    map(({ hoveredTokenPosition, hoverOverlayIsFixed }) => ({
                        position: hoveredTokenPosition,
                        isFixed: hoverOverlayIsFixed,
                    })),
                    withLatestFrom(this.codeElements.pipe(filter(isDefined))),
                    map(([state, codeElement]) => {
                        const lineElem = codeElement.querySelector(`[data-line="${state.position.line}"`)
                        if (!lineElem) {
                            return null
                        }

                        const codeCell = lineElem.nextElementSibling
                        if (!codeCell) {
                            return null
                        }
                        return { token: findTokenToHighlight(state.position, codeCell), isFixed: state.isFixed }
                    }),
                    filter(isDefined),
                    filter(propertyIsDefined('token'))
                )
                .subscribe(({ token, isFixed }) => {
                    if (!isFixed) {
                        const highlighted = document.querySelectorAll('.selection-highlight')
                        for (const h of Array.from(highlighted)) {
                            if (h !== token) {
                                h.classList.remove('selection-highlight')
                            }
                        }
                    }
                    if (!token.textContent || !token.textContent.trim().length) {
                        return
                    }
                    token.classList.add('selection-highlight')
                    this.nextHighlightedElement(token)
                })
        )

        // Close tooltip when the user presses 'escape'.
        this.subscriptions.add(
            fromEvent<KeyboardEvent>(window, 'keydown')
                .pipe(filter(event => event.keyCode === 27))
                .subscribe(event => {
                    event.preventDefault()
                    this.closeButtonClicks.next()
                })
        )

        // Telemetry
        this.subscriptions.add(this.goToDefinitionClicks.subscribe(() => eventLogger.log('GoToDefClicked')))
    }

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public shouldComponentUpdate(nextProps: Readonly<BlobProps>, nextState: Readonly<BlobState>): boolean {
        return !isEqual(this.props, nextProps) || !isEqual(this.state, nextState)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
        this.componentStateUpdates.next(this.state)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): React.ReactNode {
        const blameLineNum = this.state.activeLine
        let blamePortalID: string | null = null

        if (blameLineNum) {
            blamePortalID = this.state.blameLineIDs[blameLineNum]
        }

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
                {blameLineNum &&
                    blamePortalID && (
                        <BlameLine key={blamePortalID} portalID={blamePortalID} line={blameLineNum} {...this.props} />
                    )}
            </div>
        )
    }

    private createBlameDomNode = ({ lineNum, codeCell }: { lineNum: number; codeCell: HTMLElement }): void => {
        const portalNode = document.createElement('span')

        const id = toPortalID(lineNum)
        portalNode.id = id
        portalNode.classList.add('blame-portal')

        codeCell.appendChild(portalNode)

        this.setState({
            blameLineIDs: {
                ...this.state.blameLineIDs,
                [lineNum]: id,
            },
        })
    }
}
