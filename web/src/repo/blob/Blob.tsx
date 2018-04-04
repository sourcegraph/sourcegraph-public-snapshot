import * as H from 'history'
import isEmpty from 'lodash/isEmpty'
import isEqual from 'lodash/isEqual'
import omit from 'lodash/omit'
import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { fromEvent } from 'rxjs/observable/fromEvent'
import { interval } from 'rxjs/observable/interval'
import { merge } from 'rxjs/observable/merge'
import { catchError } from 'rxjs/operators/catchError'
import { debounceTime } from 'rxjs/operators/debounceTime'
import { filter } from 'rxjs/operators/filter'
import { map } from 'rxjs/operators/map'
import { switchMap } from 'rxjs/operators/switchMap'
import { take } from 'rxjs/operators/take'
import { takeUntil } from 'rxjs/operators/takeUntil'
import { tap } from 'rxjs/operators/tap'
import { zip } from 'rxjs/operators/zip'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { Position, Range } from 'vscode-languageserver-types'
import { EMODENOTFOUND, fetchHover, fetchJumpURL, isEmptyHover } from '../../backend/lsp'
import { eventLogger } from '../../tracking/eventLogger'
import { getPathExtension, scrollIntoView, supportedExtensions } from '../../util'
import { asError } from '../../util/errors'
import { LineOrPositionOrRange, parseHash, toAbsoluteBlobURL, toPrettyBlobURL } from '../../util/url'
import {
    AbsoluteRepoFile,
    AbsoluteRepoFilePosition,
    AbsoluteRepoFileRange,
    BlobViewState,
    getCodeCell,
    getCodeCells,
    RenderMode,
} from '../index'
import { triggerBlame } from './blame'
import {
    convertNode,
    createTooltips,
    findElementWithOffset,
    getTableDataCell,
    getTargetLineAndOffset,
    hideTooltip,
    TooltipData,
    updateTooltip,
} from './tooltips'

/**
 * Highlights a <td> element and updates the page URL if necessary.
 */
function updateLine(
    cells: HTMLElement | HTMLElement[],
    history: H.History,
    ctx: AbsoluteRepoFileRange,
    clickEvent?: MouseEvent
): void {
    if (!Array.isArray(cells)) {
        cells = [cells]
    }

    triggerBlame(ctx, clickEvent)

    const currentlyHighlighted = document.querySelectorAll('.sg-highlighted') as NodeListOf<HTMLElement>
    for (const cellElem of currentlyHighlighted) {
        cellElem.classList.remove('sg-highlighted')
        cellElem.style.backgroundColor = 'inherit'
    }

    for (const cell of cells) {
        cell.style.backgroundColor = 'rgb(34, 44, 58)'
        cell.classList.add('sg-highlighted')
    }

    // Check URL change first, since this function can be called in response to
    // onhashchange.
    const newUrl = toPrettyBlobURL(ctx)
    if (newUrl === window.location.pathname + window.location.hash) {
        // Avoid double-pushing the same URL
        return
    }

    history.push(toPrettyBlobURL(ctx))
}

/**
 * The same as updateLine, but also scrolls the blob.
 */
function updateAndScrollToLine(
    cell: HTMLElement | HTMLElement[],
    history: H.History,
    ctx: AbsoluteRepoFileRange,
    clickEvent?: MouseEvent,
    scrollIfNeeded?: boolean
): void {
    if (!cell) {
        return
    }
    if (!Array.isArray(cell)) {
        cell = [cell]
    }
    if (cell.length === 0) {
        return
    }

    updateLine(cell, history, ctx, clickEvent)
    // Scroll to the line if outside viewport.
    if (scrollIfNeeded) {
        const scrollingElement = document.querySelector('.blob')! as HTMLElement
        scrollIntoView(scrollingElement, cell[0])
    } else {
        scrollToCell(cell[0])
    }
}

function scrollToCell(cell: HTMLElement): void {
    // Scroll to the line.
    const scrollingElement = document.querySelector('.blob')!
    const viewportBound = scrollingElement.getBoundingClientRect()
    const blobTable = document.querySelector('.blob > table')! // table that we're positioning tooltips relative to.
    const tableBound = blobTable.getBoundingClientRect() // tables bounds
    const targetBound = cell.getBoundingClientRect() // our target elements bounds

    scrollingElement.scrollTop = targetBound.top - tableBound.top - viewportBound.height / 2 + targetBound.height / 2
}

interface Props extends AbsoluteRepoFile {
    location: H.Location
    history: H.History
    className: string
    html: string
    wrapCode: boolean
    renderMode: RenderMode
    isLightTheme: boolean
}

interface State {
    fixedTooltip?: TooltipData
}

export class Blob extends React.Component<Props, State> {
    public state: State = {}
    private blobElement: HTMLElement | null = null
    private fixedTooltip = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentWillReceiveProps(nextProps: Props): void {
        const nextHash = parseHash(nextProps.location.hash)
        if (this.props.location.pathname !== nextProps.location.pathname && !nextHash.line) {
            if (this.blobElement) {
                this.blobElement.scrollTop = 0
            }
        }
        const thisHash = parseHash(this.props.location.hash)
        if (thisHash.viewState !== nextHash.viewState && this.props.location.pathname === nextProps.location.pathname) {
            // When updating references mode in the same file, scroll. Wait just a moment to make sure the panel is
            // shown, since the scroll offset is calculated based on the height of the blob.
            setTimeout(() => this.scrollToLine(nextProps), 10)
        }

        if (
            this.props.location.pathname === nextProps.location.pathname &&
            (thisHash.character !== nextHash.character ||
                thisHash.line !== nextHash.line ||
                thisHash.endCharacter !== nextHash.endCharacter ||
                thisHash.endLine !== nextHash.endLine ||
                thisHash.viewState !== nextHash.viewState)
        ) {
            if (!nextHash.viewState) {
                this.fixedTooltip.next(nextProps)
                this.scrollToLine(nextProps, true)
            } else {
                // If showing panel, remove any tooltip then highlight the element for the given start position.
                this.setFixedTooltip()
                if (nextHash.line) {
                    this.addSelectionHighlightSticky(nextHash.line, nextHash.character)
                    if (getCodeCell(nextHash.line)) {
                        this.scrollToLine(nextProps)
                    }
                }
            }
        }

        if (this.props.html !== nextProps.html) {
            // Hide the previous tooltip, if it exists.
            hideTooltip()

            this.subscriptions.unsubscribe()
            this.subscriptions = new Subscription()
            if (this.blobElement) {
                this.addEventListeners(this.blobElement, nextProps)
            }
            this.setFixedTooltip()
        }
    }

    public shouldComponentUpdate(nextProps: Props): boolean {
        // Update the blob if the inner HTML content changes.
        if (this.props.html !== nextProps.html) {
            return true
        }

        // Update the blob if wrapCode changes value.
        if (this.props.wrapCode !== nextProps.wrapCode) {
            return true
        }

        if (isEqual(omit(this.props, 'rev'), omit(nextProps, 'rev'))) {
            // nextProps is a new location, but we don't have new HTML.
            // We *only* want lifeycle hooks when the html is changed.
            // This prevents e.g. scrolling to a line that doesn't exist
            // yet when file has changed but html hasn't been resolved.
            return false
        }

        const prevHash = parseHash(this.props.location.hash)
        const nextHash = parseHash(nextProps.location.hash)
        if (
            (prevHash.line !== nextHash.line || prevHash.endLine !== nextHash.endLine) &&
            nextProps.history.action === 'POP'
        ) {
            // If we don't need an update (the file hasn't changed, and we will *not* get into componentDidUpdate).
            // We still want to scroll if the hash is changed, but only on 'back' and 'forward' browser events (and not e.g. on each line click).
            this.scrollToLine(nextProps)
        }
        return false
    }

    public componentDidUpdate(prevProps: Props, prevState: State): void {
        hideTooltip()
        createTooltips()

        const parsedHash = parseHash(this.props.location.hash)
        if (!parsedHash.viewState) {
            // Show fixed tooltip if necessary iff not showing a panel.
            this.fixedTooltip.next(this.props)
        }
        // The HTML contents were updated on a mounted component, e.g. from a 'back' or 'forward' event,
        // or a jump-to-def.
        this.scrollToLine(this.props)

        // Update highlighted range.
        if (this.props.location.hash !== prevProps.location.hash) {
            if (parsedHash.line) {
                this.addSelectionHighlightSticky(parsedHash.line, parsedHash.character)
            }
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private addSelectionHighlightSticky(line: number, character?: number): void {
        const cell = getCodeCell(line)
        if (cell) {
            const el = findElementWithOffset(cell.childNodes[1]! as HTMLElement, character || 0)
            if (el) {
                el.classList.add('selection-highlight-sticky')
            }
        }
    }

    public render(): JSX.Element | null {
        return (
            <code
                className={`blob ${this.props.wrapCode ? ' blob--wrapped' : ''} ${this.props.className}`}
                ref={this.onBlobRef}
                dangerouslySetInnerHTML={{ __html: this.props.html }}
            />
        )
    }

    private onBlobRef = (ref: HTMLElement | null) => {
        this.blobElement = ref
        if (ref) {
            // This is the first time the component is ever mounted. We need to set initial scroll.
            this.scrollToLine(this.props)
            createTooltips()
            this.addEventListeners(ref, this.props)
            const parsedHash = parseHash(this.props.location.hash)
            if (parsedHash.line && parsedHash.character) {
                this.fixedTooltip.next(this.props)
            }
        }
    }

    private addEventListeners = (ref: HTMLElement, props: Props): void => {
        const isSupportedExtension = supportedExtensions.has(getPathExtension(props.filePath))
        if (isSupportedExtension) {
            this.subscriptions.add(
                this.fixedTooltip
                    .pipe(
                        filter(props => {
                            const parsed = parseHash(props.location.hash)
                            if (parsed.line && parsed.character) {
                                const cell = getCodeCell(parsed.line)
                                if (cell) {
                                    const td = cell.childNodes[1] as HTMLTableDataCellElement
                                    if (td && !td.classList.contains('annotated')) {
                                        td.classList.add('annotated')
                                        convertNode(td)
                                    }
                                }
                                if (!parsed.viewState) {
                                    return true
                                }
                                // Don't show a tooltip when there is a panel (but do highlight the token)
                                // TODO(john): this can probably be simplified.
                                if (cell) {
                                    const el = findElementWithOffset(
                                        cell.childNodes[1]! as HTMLElement,
                                        parsed.character!
                                    )
                                    if (el) {
                                        el.classList.add('selection-highlight-sticky')
                                        return false
                                    }
                                }
                            }
                            this.setFixedTooltip()
                            return false
                        }),
                        map(props => parseHash(props.location.hash)),
                        map(pos =>
                            findElementWithOffset(getCodeCell(pos.line!).childNodes[1]! as HTMLElement, pos.character!)
                        ),
                        filter((el: HTMLElement | undefined): el is HTMLElement => !!el),
                        map((target: HTMLElement) => {
                            const data = { target, loc: getTargetLineAndOffset(target!, false) }
                            if (!data.loc) {
                                return null
                            }
                            const ctx = { ...this.props, position: data.loc! }
                            return { target: data.target, ctx }
                        }),
                        switchMap(data => {
                            if (data === null) {
                                return [null]
                            }
                            const { target, ctx } = data
                            return this.getTooltip(target, ctx).pipe(
                                tap(tooltip => {
                                    if (!tooltip) {
                                        this.setFixedTooltip()
                                        return
                                    }

                                    const contents = tooltip.contents
                                    if (!contents || isEmptyHover({ contents })) {
                                        this.setFixedTooltip()
                                        return
                                    }

                                    this.setFixedTooltip(tooltip)
                                    updateTooltip(tooltip, true, this.tooltipActions(ctx))
                                }),
                                zip(this.getDefinition(ctx).pipe(catchError(err => [asError(err)]))),
                                map(([tooltip, defResponse]) => ({
                                    ...tooltip,
                                    defUrlOrError: defResponse || undefined,
                                })),
                                catchError(err => {
                                    if (err.code !== EMODENOTFOUND) {
                                        console.error(err)
                                    }
                                    const data: TooltipData = { target, ctx }
                                    return [data]
                                })
                            )
                        })
                    )
                    .subscribe(data => {
                        if (!data) {
                            this.setFixedTooltip()
                            return
                        }

                        const contents = data.contents
                        if (!contents || isEmptyHover({ contents })) {
                            this.setFixedTooltip()
                            return
                        }

                        this.setFixedTooltip(data)
                        updateTooltip(data, true, this.tooltipActions(data.ctx))
                    })
            )
            this.subscriptions.add(
                fromEvent<MouseEvent>(ref, 'mouseover')
                    .pipe(
                        debounceTime(50),
                        map(e => e.target as HTMLElement),
                        tap(target => {
                            const td = getTableDataCell(target)
                            if (td && !td.classList.contains('annotated')) {
                                td.classList.add('annotated')
                                convertNode(td)
                            }
                        }),
                        map(target => ({ target, loc: getTargetLineAndOffset(target, false) })),
                        filter(data => Boolean(data.loc)),
                        map(data => ({ target: data.target, ctx: { ...this.props, position: data.loc! } })),
                        switchMap(({ target, ctx }) => {
                            const tooltip = this.getTooltip(target, ctx)
                            const loading = this.getLoadingTooltip(target, ctx, tooltip)

                            // Preemptively fetch the symbol's definition, but no need to pass it on to the hover
                            // (getDefinition is called again when the hover is docked).
                            this.getDefinition(ctx)

                            return merge(loading, tooltip).pipe(
                                catchError(err => {
                                    if (err.code !== EMODENOTFOUND) {
                                        console.error(err)
                                    }
                                    const data: TooltipData = { target, ctx }
                                    return [data]
                                })
                            )
                        })
                    )
                    .subscribe(data => {
                        this.logTelemetryOnTooltip(data)
                        if (!this.state.fixedTooltip) {
                            updateTooltip(data, false, this.tooltipActions(data.ctx))
                        }
                    })
            )
        }

        this.subscriptions.add(
            fromEvent<MouseEvent>(ref, 'mouseout').subscribe(e => {
                for (const el of document.querySelectorAll('.blob .selection-highlight')) {
                    el.classList.remove('selection-highlight')
                }
                if (isSupportedExtension && !this.state.fixedTooltip) {
                    hideTooltip()
                }
            })
        )
        // When the user presses 'esc', dismiss tooltip.
        this.subscriptions.add(
            fromEvent<KeyboardEvent>(window, 'keydown')
                .pipe(filter(event => event.keyCode === 27))
                .subscribe(event => {
                    event.preventDefault()
                    this.handleDismiss()
                })
        )
        this.subscriptions.add(
            fromEvent<MouseEvent>(ref, 'click')
                .pipe(
                    filter(e => {
                        const target = e.target as HTMLElement
                        if (!target) {
                            return false
                        }
                        const tooltip = document.querySelector('.sg-tooltip')
                        if (tooltip && tooltip.contains(target)) {
                            return false
                        }
                        return true
                    })
                )
                .subscribe(e => {
                    const target = e.target as HTMLElement
                    const row = (target as Element).closest('tr') as HTMLTableRowElement | null
                    if (!row) {
                        return
                    }

                    const clickedLineNumber = target && target.classList.contains('line')

                    const targetLine = parseInt(row.firstElementChild!.getAttribute('data-line')!, 10)
                    const data = { target, loc: getTargetLineAndOffset(target, false) }
                    const targetPos: Position = { line: targetLine, character: data.loc ? data.loc.character : 0 }

                    // Expand selection if shift-click on line number.
                    const shouldGrowSelectionLines = e.shiftKey && clickedLineNumber
                    const currentRange: LineOrPositionOrRange = parseHash(this.props.location.hash)
                    let newRange: Range
                    if (shouldGrowSelectionLines) {
                        // Always select entire lines when selecting by line (don't allow multi-line selection
                        // to/from specific characters on the line).
                        let start: Position = { line: currentRange.line || 1, character: 0 }
                        let end: Position = { line: currentRange.endLine || start.line, character: 0 }

                        // Ensure currentRange's start line is before its end.
                        if (start.line > end.line || (start.line === end.line && start.character > end.character)) {
                            const tmp = end
                            end = start
                            start = tmp
                        }

                        // TODO(sqs): remember selection anchor point to grow correctly, instead of
                        // always growing from start
                        if (
                            targetPos.line < start.line ||
                            (targetPos.line === start.line && targetPos.character < start.character)
                        ) {
                            newRange = { start: targetPos, end }
                        } else {
                            newRange = { start, end: targetPos }
                        }
                    } else {
                        newRange = { start: targetPos, end: targetPos }
                    }

                    const rows = getCodeCells(newRange.start.line, newRange.end.line)
                    if (!data.loc) {
                        return updateLine(
                            rows,
                            this.props.history,
                            {
                                repoPath: this.props.repoPath,
                                rev: this.props.rev,
                                commitID: this.props.commitID,
                                filePath: this.props.filePath,
                                range: newRange,
                                renderMode: this.props.renderMode,
                            },
                            e
                        )
                    }
                    const ctx = {
                        ...this.props,
                        range: newRange,
                    }
                    updateLine(rows, this.props.history, ctx, e)
                })
        )
    }

    /**
     * A fixed tooltip is one that is docked. In the web UI, this means the user has
     * clicked on the symbol corresponding to the tooltip. getTooltip and getDefinition
     * is called on the current fixedTooltip, so this should be called whenever there is
     * a new symbol clicked/the tooltip we need information for changes.
     */
    private setFixedTooltip = (data?: TooltipData) => {
        for (const el of document.querySelectorAll('.blob .selection-highlight')) {
            el.classList.remove('selection-highlight')
        }
        for (const el of document.querySelectorAll('.blob .selection-highlight-sticky')) {
            el.classList.remove('selection-highlight-sticky')
        }
        if (data) {
            if (data.defUrlOrError === undefined) {
                eventLogger.log('TooltipDocked', { hoverHasDefUrl: false })
            } else {
                eventLogger.log('TooltipDockedWithDefinition', { hoverHasDefUrl: true })
            }
            data.target.classList.add('selection-highlight-sticky')
        } else {
            hideTooltip()
        }
        this.setState({ fixedTooltip: data || undefined })
    }

    private scrollToLine = (props: Props, scrollOnlyIfNeeded?: boolean) => {
        const parsed = parseHash<BlobViewState>(props.location.hash)
        const { line, character, endLine, endCharacter, viewState } = parsed
        if (line) {
            const cells = getCodeCells(line, endLine)
            updateAndScrollToLine(
                cells,
                props.history,
                {
                    repoPath: props.repoPath,
                    rev: props.rev,
                    commitID: props.commitID,
                    filePath: props.filePath,
                    range: {
                        start: { line, character: character || 0 },
                        end: endLine
                            ? { line: endLine, character: endCharacter || 0 }
                            : { line, character: character || 0 },
                    },
                    viewState,
                    renderMode: props.renderMode,
                },
                undefined,
                scrollOnlyIfNeeded
            )
        }
    }

    /**
     * getTooltip wraps the asynchronous fetch of tooltip data from the Sourcegraph API.
     * This Observable will emit exactly one value before it completes. If the resolved
     * tooltip is defined, it will update the target styling.
     */
    private getTooltip(target: HTMLElement, ctx: AbsoluteRepoFilePosition): Observable<TooltipData> {
        return fetchHover(ctx).pipe(
            tap(data => {
                if (isEmptyHover(data)) {
                    // short-cirtuit, no tooltip data
                    return
                }
                target.style.cursor = 'pointer'
                target.classList.add('selection-highlight')
            }),
            map(data => ({ target, ctx, ...data }))
        )
    }
    /**
     * getDefinition wraps the asynchronous fetch of tooltip data from the Sourcegraph API.
     * This Observable will emit exactly one value before it completes.
     */
    private getDefinition(ctx: AbsoluteRepoFilePosition): Observable<string | null> {
        return fetchJumpURL(ctx)
    }

    /**
     * getLoadingTooltip emits "loading" tooltip data after a delay,
     * iff the other Observable hasn't already emitted a value.
     */
    private getLoadingTooltip(
        target: HTMLElement,
        ctx: AbsoluteRepoFilePosition,
        tooltip: Observable<TooltipData>
    ): Observable<TooltipData> {
        return interval(500).pipe(take(1), takeUntil(tooltip), map(() => ({ target, ctx, loading: true })))
    }

    private handleGoToDefinition = (defCtx: AbsoluteRepoFilePosition) => (e: MouseEvent) => {
        eventLogger.log('GoToDefClicked')
        if (e.shiftKey || e.ctrlKey || e.altKey || e.metaKey) {
            return
        }
        e.preventDefault()
        const lastHash = parseHash(this.props.location.hash)
        hideTooltip()
        if (
            defCtx.position.line &&
            this.props.repoPath === defCtx.repoPath &&
            (this.props.rev === defCtx.rev || this.props.commitID === defCtx.commitID) &&
            this.props.filePath === defCtx.filePath &&
            (lastHash.line !== defCtx.position.line ||
                lastHash.character !== defCtx.position.character ||
                lastHash.endLine !== defCtx.position.line ||
                lastHash.endCharacter !== defCtx.position.character)
        ) {
            // Handles URL update + scroll to file (for j2d within same file).
            // Since the defCtx rev/commitID may be undefined, use the resolved rev
            // for the current file.
            const ctx = {
                ...defCtx,
                commitID: this.props.commitID,
                range: {
                    start: { line: defCtx.position.line, character: defCtx.position.character || 0 },
                    end: { line: defCtx.position.line, character: defCtx.position.character || 0 },
                },
            } as AbsoluteRepoFileRange
            updateAndScrollToLine(getCodeCell(ctx.range.start.line), this.props.history, ctx)
        } else {
            this.setFixedTooltip()
            this.props.history.push(toAbsoluteBlobURL(defCtx))
        }
    }

    private handleFindReferences = (ctx: AbsoluteRepoFilePosition) => (e: MouseEvent) => {
        eventLogger.log('FindRefsClicked')
        if (e.shiftKey || e.ctrlKey || e.altKey || e.metaKey) {
            return
        }
        e.preventDefault()
        this.props.history.push(toPrettyBlobURL({ ...ctx, rev: this.props.rev, viewState: 'references' }))
        hideTooltip()
        scrollToCell(getCodeCell(ctx.position.line))
    }

    private handleDismiss = () => {
        const parsed = parseHash(this.props.location.hash)
        if (parsed.line) {
            // Remove the character position so the fixed tooltip goes away.
            const ctx = {
                ...this.props,
                range: {
                    start: { line: parsed.line, character: 0 },
                    end: parsed.endLine ? { line: parsed.endLine, character: 0 } : { line: parsed.line, character: 0 },
                },
            } as AbsoluteRepoFileRange
            this.props.history.push(toPrettyBlobURL(ctx))
        } else {
            // Unset fixed tooltip if it exists (no URL update necessary).
            this.setFixedTooltip()
        }
    }

    private logTelemetryOnTooltip = (data: TooltipData) => {
        // Only log an event if there is no fixed tooltip docked, we have a target element
        if (!this.state.fixedTooltip && data.target) {
            if (data.loading) {
                eventLogger.log('SymbolHoveredLoading')
                // Don't log tooltips with no content
            } else if (!isEmpty(data.contents)) {
                eventLogger.log('SymbolHovered')
            }
        }
    }

    private tooltipActions = (ctx: AbsoluteRepoFilePosition) => ({
        definition: this.handleGoToDefinition,
        references: this.handleFindReferences,
        dismiss: this.handleDismiss,
    })
}
