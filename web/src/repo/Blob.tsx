import * as H from 'history'
import * as React from 'react'
import 'rxjs/add/observable/fromEvent'
import 'rxjs/add/observable/fromPromise'
import 'rxjs/add/observable/interval'
import 'rxjs/add/operator/catch'
import 'rxjs/add/operator/do'
import 'rxjs/add/operator/filter'
import 'rxjs/add/operator/map'
import 'rxjs/add/operator/merge'
import 'rxjs/add/operator/switchMap'
import 'rxjs/add/operator/take'
import 'rxjs/add/operator/takeUntil'
import 'rxjs/add/operator/zip'
import { Observable } from 'rxjs/Observable'
import { Subscription } from 'rxjs/Subscription'
import { EEMPTYTOOLTIP, fetchJumpURL, getTooltip } from 'sourcegraph/backend/lsp'
import { triggerBlame } from 'sourcegraph/blame'
import { AbsoluteRepoFile, AbsoluteRepoFilePosition, getCodeCell } from 'sourcegraph/repo'
import { convertNode, createTooltips, getTableDataCell, getTargetLineAndOffset, hideTooltip, TooltipData, updateTooltip } from 'sourcegraph/repo/tooltips'
import { events } from 'sourcegraph/tracking/events'
import { getPathExtension, supportedExtensions } from 'sourcegraph/util'
import { parseHash, toBlobPositionURL, toPrettyBlobPositionURL } from 'sourcegraph/util/url'

/**
 * Highlights a <td> element and updates the page URL if necessary.
 */
function updateLine(cell: HTMLElement, history: H.History, ctx: AbsoluteRepoFilePosition, userTriggered?: React.MouseEvent<HTMLDivElement>): void {
    triggerBlame({
        time: new Date(),
        repoURI: ctx.repoPath,
        commitID: ctx.commitID,
        path: ctx.filePath,
        line: ctx.position.line
    }, userTriggered)

    const currentlyHighlighted = document.querySelectorAll('.sg-highlighted') as NodeListOf<HTMLElement>
    for (const cellElem of currentlyHighlighted) {
        cellElem.classList.remove('sg-highlighted')
        cellElem.style.backgroundColor = 'inherit'
    }

    cell.style.backgroundColor = 'rgb(34, 44, 58)'
    cell.classList.add('sg-highlighted')

    // Check URL change first, since this function can be called in response to
    // onhashchange.
    const newUrl = toPrettyBlobPositionURL(ctx)
    if (newUrl === (window.location.pathname + window.location.hash)) {
        // Avoid double-pushing the same URL
        return
    }

    history.push(toPrettyBlobPositionURL(ctx))
}

/**
 * The same as updateLine, but also scrolls the blob.l
 */
function updateAndScrollToLine(cell: HTMLElement, history: H.History, ctx: AbsoluteRepoFilePosition, userTriggered?: React.MouseEvent<HTMLDivElement>): void {
    updateLine(cell, history, ctx, userTriggered)

    // Scroll to the line.
    const scrollingElement = document.querySelector('.blob')!
    const viewportBound = scrollingElement.getBoundingClientRect()
    const blobTable = document.querySelector('.blob > table')! // table that we're positioning tooltips relative to.
    const tableBound = blobTable.getBoundingClientRect() // tables bounds
    const targetBound = cell.getBoundingClientRect() // our target elements bounds

    scrollingElement.scrollTop = targetBound.top - tableBound.top - (viewportBound.height / 2) + (targetBound.height / 2)
}

interface Props extends AbsoluteRepoFile {
    html: string
    location: H.Location
    history: H.History
}

interface State {
    fixedTooltip?: TooltipData
}

export class Blob extends React.Component<Props, State> {
    public state: State = {}
    private tooltip: TooltipData | null
    private blobElement: HTMLElement | null = null
    private subscriptions = new Subscription()

    public componentWillReceiveProps(nextProps: Props): void {
        const hash = parseHash(nextProps.location.hash)
        if (this.props.location.pathname !== nextProps.location.pathname && !hash.line) {
            if (this.blobElement) {
                this.blobElement.scrollTop = 0
            }
        }

        if (this.props.html !== nextProps.html) {
            // Hide the previous tooltip, if it exists.
            hideTooltip()

            this.subscriptions.unsubscribe()
            this.subscriptions = new Subscription()
            if (this.blobElement) {
                this.addTooltipEventListeners(this.blobElement)
            }
            this.setState({ fixedTooltip: undefined })
        }
    }

    public shouldComponentUpdate(nextProps: Props): boolean {
        // Only update the blob if the inner HTML content changes.
        if (this.props.html !== nextProps.html) {
            return true
        }

        const prevHash = parseHash(this.props.location.hash)
        const nextHash = parseHash(nextProps.location.hash)
        if (prevHash.line !== nextHash.line && nextProps.history.action === 'POP') {
            // If we don't need an update (the file hasn't changed, and we will *not* get into componentDidUpdate).
            // We still want to scroll if the hash is changed, but only on 'back' and 'forward' browser events (and not e.g. on each line click).
            this.scrollToLine(nextProps)
        }
        return false
    }

    public componentDidUpdate(prevProps: Props, prevState: State): void {
        hideTooltip()
        createTooltips()
        if (this.props.history.action === 'POP') {
            // The contents were updated on a mounted component and we did a 'back' or 'forward' event;
            // scroll to the appropariate line after the new table is created.
            this.scrollToLine(this.props)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className='blob' onClick={this.handleBlobClick} ref={this.onBlobRef} dangerouslySetInnerHTML={{ __html: this.props.html }} />
        )
    }

    private onBlobRef = (ref: HTMLElement | null) => {
        this.blobElement = ref
        if (ref) {
            // This is the first time the component is ever mounted. We need to set initial scroll.
            this.scrollToLine(this.props)
            createTooltips()
            if (supportedExtensions.has(getPathExtension(this.props.filePath))) {
                this.addTooltipEventListeners(ref)
            }
        }
    }

    private addTooltipEventListeners = (ref: HTMLElement): void => {
        this.subscriptions.add(
            Observable.fromEvent<MouseEvent>(ref, 'mouseover')
                .map(e => e.target as HTMLElement)
                .do(target => {
                    const td = getTableDataCell(target)
                    if (td && !td.classList.contains('annotated')) {
                        td.classList.add('annotated')
                        td.setAttribute('data-sg-line-number', (td.previousSibling as HTMLTableDataCellElement).getAttribute('data-line') || '')
                        convertNode(td)
                    }
                })
                .map(target => ({ target, loc: getTargetLineAndOffset(target, false) }))
                .filter(data => Boolean(data.loc))
                .map(data => ({ target: data.target, ctx: { ...this.props, position: data.loc! } }))
                .switchMap(({ target, ctx }) => {
                    const tooltip = this.getTooltip(target, ctx)
                    const tooltipWithJ2D: Observable<TooltipData> = tooltip.zip(this.getDefinition(ctx))
                        .map(([tooltip, defUrl]) => ({ ...tooltip, defUrl: defUrl || undefined }))
                    const loading = this.getLoadingTooltip(target, ctx, tooltip)
                    return Observable.merge(loading, tooltip, tooltipWithJ2D).catch(e => {
                        if (e.code !== EEMPTYTOOLTIP) {
                            console.error(e)
                        }
                        const data: TooltipData = { target, ctx }
                        return [data]
                    })
                })
                .subscribe(data => {
                    const lastTooltip = this.tooltip
                    if (lastTooltip && lastTooltip.target.classList.contains('selection-highlight')) {
                        lastTooltip.target.classList.remove('selection-highlight')
                    }
                    this.tooltip = data
                    if (data.title) {
                        data.target.classList.add('selection-highlight')
                    }
                    if (!this.state.fixedTooltip) {
                        updateTooltip(data, false, this.tooltipActions(data.ctx))
                    }
                })
        )
        this.subscriptions.add(
            Observable.fromEvent<MouseEvent>(ref, 'mouseout')
                .subscribe(e => {
                    for (const el of document.querySelectorAll('.blob .selection-highlight')) {
                        el.classList.remove('selection-highlight')
                    }
                    if (!this.state.fixedTooltip) {
                        hideTooltip()
                    }
                })
        )
        this.subscriptions.add(
            Observable.fromEvent<MouseEvent>(ref, 'click')
                .map(e => e.target as HTMLElement)
                .filter(target => {
                    if (!target) {
                        return false
                    }
                    const tooltip = document.querySelector('.sg-tooltip')
                    if (tooltip && tooltip.contains(target)) {
                        return false
                    }
                    return true
                })
                .map(target => {
                    const data = { target, loc: getTargetLineAndOffset(target, false) }
                    if (!data.loc) {
                        return null
                    }
                    const ctx = { ...this.props, position: data.loc! }
                    return { target: data.target, ctx }
                })
                .switchMap(data => {
                    if (data === null) {
                        return [null]
                    }
                    const { target, ctx } = data
                    const tooltipWithJ2D: Observable<TooltipData> = this.getTooltip(target, ctx)
                        .zip(this.getDefinition(ctx))
                        .map(([tooltip, defUrl]) => ({ ...tooltip, defUrl: defUrl || undefined }))
                    return tooltipWithJ2D.catch(e => {
                        if (e.code !== EEMPTYTOOLTIP) {
                            console.error(e)
                        }
                        const data: TooltipData = { target, ctx }
                        return [data]
                    })
                })
                .subscribe(data => {
                    this.tooltip = data
                    if (!data || !data.title) {
                        this.setState({ fixedTooltip: undefined }, hideTooltip)
                    } else {
                        this.setState({ fixedTooltip: data }, () => updateTooltip(data, true, this.tooltipActions(data.ctx)))
                    }
                })
        )
    }

    private handleBlobClick: React.MouseEventHandler<HTMLDivElement> = e => {
        const row = (e.target as Element).closest('tr') as HTMLTableRowElement | null
        if (!row) {
            return
        }
        const line = parseInt(row.firstElementChild!.getAttribute('data-line')!, 10)
        updateLine(row.lastChild as HTMLElement, this.props.history, {
            repoPath: this.props.repoPath,
            rev: this.props.rev,
            commitID: this.props.commitID,
            filePath: this.props.filePath,
            position: { line }
        }, e)
    }

    private scrollToLine = (props: Props) => {
        const { line, char, modalMode } = parseHash(props.location.hash)
        if (line) {
            updateAndScrollToLine(getCodeCell(line), props.history, {
                repoPath: props.repoPath,
                rev: props.rev,
                commitID: props.commitID,
                filePath: props.filePath,
                position: { line, char },
                referencesMode: modalMode
            })
        }
    }

    /**
     * getTooltip wraps the asynchronous fetch of tooltip data from the Sourcegraph API.
     * This Observable will emit exactly one value before it completes. If the resolved
     * tooltip is defined, it will update the target styling.
     */
    private getTooltip(target: HTMLElement, ctx: AbsoluteRepoFilePosition): Observable<TooltipData> {
        return Observable.fromPromise(getTooltip(ctx))
            .do(data => {
                if (data && data.title) {
                    target.style.cursor = 'pointer'
                }
            })
            .map(data => ({ target, ctx, ...data }))
    }
    /**
     * getDefinition wraps the asynchronous fetch of tooltip data from the Sourcegraph API.
     * This Observable will emit exactly one value before it completes.
     */
    private getDefinition(ctx: AbsoluteRepoFilePosition): Observable<string | null> {
        return Observable.fromPromise(fetchJumpURL(ctx))
    }

    /**
     * getLoadingTooltip emits "loading" tooltip data after a delay,
     * iff the other Observable hasn't already emitted a value.
     */
    private getLoadingTooltip(target: HTMLElement, ctx: AbsoluteRepoFilePosition, tooltip: Observable<TooltipData>): Observable<TooltipData> {
        return Observable.interval(500)
            .take(1)
            .takeUntil(tooltip)
            .map(() => ({ target, ctx, loading: true }))
    }

    private handleGoToDefinition = (defCtx: AbsoluteRepoFilePosition) =>
        (e: MouseEvent) => {
            events.GoToDefClicked.log()
            if (e.shiftKey || e.ctrlKey || e.altKey || e.metaKey) {
                return
            }
            e.preventDefault()
            const lastHash = parseHash(this.props.location.hash)
            hideTooltip()
            if (defCtx.position.line && this.props.repoPath === defCtx.repoPath && this.props.rev === defCtx.rev && lastHash.line !== defCtx.position.line) {
                // Handles URL update + scroll to file (for j2d within same file)
                updateAndScrollToLine(getCodeCell(defCtx.position.line), this.props.history, defCtx)
            } else {
                this.setState({ fixedTooltip: undefined }, () => this.props.history.push(toBlobPositionURL(defCtx)))
            }
        }

    private handleFindReferences = (ctx: AbsoluteRepoFilePosition) =>
        (e: MouseEvent) => {
            events.FindRefsClicked.log()
            e.preventDefault()
            this.props.history.push(toBlobPositionURL({ ...ctx, referencesMode: 'local' }))
            hideTooltip()
        }

    private handleDismiss = () => {
        this.setState({ fixedTooltip: undefined })
        hideTooltip()
    }

    private tooltipActions = (ctx: AbsoluteRepoFilePosition) =>
        ({ definition: this.handleGoToDefinition, references: this.handleFindReferences, dismiss: this.handleDismiss })
}
