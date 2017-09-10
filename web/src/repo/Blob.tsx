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
import { triggerReferences } from 'sourcegraph/references'
import { AbsoluteRepoPosition } from 'sourcegraph/repo'
import { convertNode, createTooltips, getTableDataCell, getTargetLineAndOffset, hideTooltip, TooltipData, updateTooltip } from 'sourcegraph/repo/tooltips'
import { events } from 'sourcegraph/tracking/events'
import { getCodeCellsForAnnotation, getPathExtension, highlightAndScrollToLine, highlightLine, supportedExtensions } from 'sourcegraph/util'
import * as url from 'sourcegraph/util/url'

interface BlobProps {
    html: string
    repoPath: string
    filePath: string
    commitID: string
    rev?: string
    location: H.Location
    history: H.History
}

interface State {
    fixedTooltip?: TooltipData
}

export class Blob extends React.Component<BlobProps, State> {
    public state: State = {}
    private tooltip: TooltipData | null
    private blobElement: HTMLElement | null = null
    private subscriptions = new Subscription()

    public componentWillReceiveProps(nextProps: BlobProps): void {
        const hash = url.parseHash(nextProps.location.hash)
        if (this.props.location.pathname !== nextProps.location.pathname && !hash.line) {
            if (this.blobElement) {
                this.blobElement.scrollTop = 0
            }
        }

        if (this.props.html !== nextProps.html) {
            this.subscriptions.unsubscribe()
            this.subscriptions = new Subscription()
            if (this.blobElement) {
                this.addTooltipEventListeners(this.blobElement)
            }
            createTooltips()
            this.setState({ fixedTooltip: undefined })
        }
    }

    public shouldComponentUpdate(nextProps: BlobProps): boolean {
        return this.props.html !== nextProps.html
    }

    public componentDidUpdate(): void {
        createTooltips()
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
        highlightLine(this.props.history, this.props.repoPath, this.props.commitID, this.props.filePath, line, getCodeCellsForAnnotation(), e)
    }

    private scrollToLine = (props: BlobProps) => {
        const line = url.parseHash(props.location.hash).line
        if (line) {
            highlightAndScrollToLine(props.history, props.repoPath,
                props.commitID, props.filePath, line, getCodeCellsForAnnotation())
        }
    }

    /**
     * getTooltip wraps the asynchronous fetch of tooltip data from the Sourcegraph API.
     * This Observable will emit exactly one value before it completes. If the resolved
     * tooltip is defined, it will update the target styling.
     */
    private getTooltip(target: HTMLElement, ctx: AbsoluteRepoPosition): Observable<TooltipData> {
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
    private getDefinition(ctx: AbsoluteRepoPosition): Observable<string | null> {
        return Observable.fromPromise(fetchJumpURL(ctx))
    }

    /**
     * getLoadingTooltip emits "loading" tooltip data after a delay,
     * iff the other Observable hasn't already emitted a value.
     */
    private getLoadingTooltip(target: HTMLElement, ctx: AbsoluteRepoPosition, tooltip: Observable<TooltipData>): Observable<TooltipData> {
        return Observable.interval(500)
            .take(1)
            .takeUntil(tooltip)
            .map(() => ({ target, ctx }))
    }

    private handleGoToDefinition = (ctx: AbsoluteRepoPosition) =>
        (e: MouseEvent) => {
            events.GoToDefClicked.log()
            if (e.shiftKey || e.ctrlKey || e.altKey || e.metaKey) {
                return
            }
            e.preventDefault()
            const lastHash = url.parseHash(this.props.location.hash)
            hideTooltip()
            if (ctx.position.line && this.props.repoPath === ctx.repoPath && this.props.rev === ctx.rev && lastHash.line !== ctx.position.line) {
                // Handles URL update + scroll to file (for j2d within same file)
                highlightAndScrollToLine(this.props.history, ctx.repoPath,
                    ctx.commitID, ctx.filePath, ctx.position.line, getCodeCellsForAnnotation())
            } else {
                this.setState({ fixedTooltip: undefined }, () => this.props.history.push(url.toBlobV2(ctx)))
            }
        }

    private handleFindReferences = (ctx: AbsoluteRepoPosition) =>
        (e: MouseEvent) => {
            events.FindRefsClicked.log()
            e.preventDefault()
            triggerReferences(ctx)
            this.props.history.push(`/${ctx.repoPath}@${ctx.commitID}/-/blob/${ctx.filePath}#L${ctx.position.line}:${ctx.position.char}$references`)
            hideTooltip()
        }

    private handleDismiss = () => {
        this.setState({ fixedTooltip: undefined })
        hideTooltip()
    }

    private tooltipActions = (ctx: AbsoluteRepoPosition) =>
        ({ definition: this.handleGoToDefinition, references: this.handleFindReferences, dismiss: this.handleDismiss })
}
