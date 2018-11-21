import { isEmpty } from 'lodash'
import * as React from 'react'
import { fromEvent, interval, merge, Observable, Subject, Subscription, zip } from 'rxjs'
import { catchError, debounceTime, filter, map, switchMap, take, takeUntil, tap } from 'rxjs/operators'

import { getPathExtension } from '../../../../../shared/src/languages'
import * as github from '../../libs/github/util'
import { fetchJumpURL, isEmptyHover, lspViaAPIXlang, SimpleProviderFns } from '../backend/lsp'
import {
    AbsoluteRepoFile,
    AbsoluteRepoFilePosition,
    CodeCell,
    MaybeDiffSpec,
    OpenInSourcegraphProps,
    PositionSpec,
} from '../repo'
import { fetchBlobContentLines } from '../repo/backend'
import {
    convertNode,
    createTooltips,
    getTableDataCell,
    hideTooltip,
    isOtherFileTooltipVisible,
    isTooltipVisible,
    TooltipData,
    updateTooltip,
} from '../repo/tooltips'
import { eventLogger } from '../util/context'
import { parseHash } from '../util/url'
import { OpenOnSourcegraph } from './OpenOnSourcegraph'

export interface ButtonProps {
    className: string
    style: React.CSSProperties
    iconStyle?: React.CSSProperties
}

interface Props extends AbsoluteRepoFile, Partial<PositionSpec> {
    fileElement: HTMLElement
    getTableElement: () => HTMLTableElement
    getCodeCells: () => CodeCell[]
    getTargetLineAndOffset: (
        target: HTMLElement,
        opt: MaybeDiffSpec
    ) => { line: number; character: number; word: string } | undefined
    findElementWithOffset: (
        cell: HTMLElement,
        line: number,
        offset: number,
        opt: MaybeDiffSpec
    ) => HTMLElement | undefined
    findTokenCell: (td: HTMLElement, target: HTMLElement) => HTMLElement
    filterTarget: (target: HTMLElement) => boolean
    getNodeToConvert: (td: HTMLTableDataCellElement) => HTMLElement | null
    isCommit: boolean
    isPullRequest: boolean
    isDelta?: boolean
    isSplitDiff: boolean
    isBase: boolean
    buttonProps: ButtonProps
    simpleProviderFns?: SimpleProviderFns
}

interface State {
    fixedTooltip?: TooltipData
    showOpenFileCTA?: boolean
}

export class BlobAnnotator extends React.Component<Props, State> {
    public fileExtension: string
    public isDelta: boolean
    private fixedTooltip = new Subject<Props>()
    private subscriptions = new Subscription()
    private cells = new Map<number, CodeCell>()
    private simpleProviderFns = lspViaAPIXlang

    constructor(props: Props) {
        super(props)

        this.state = {}

        this.fileExtension = getPathExtension(this.props.filePath)
        this.isDelta = this.props.isDelta || (this.props.isCommit || this.props.isPullRequest)
        this.updateCodeCells()
        this.simpleProviderFns = props.simpleProviderFns || lspViaAPIXlang
    }

    // BlobAnnotator will only ever recieve new props when it is being rendered
    // from a component that is monitoring for a table change(split diff <-> unified diff)
    public componentWillReceiveProps(nextProps: Props): void {
        this.subscriptions.unsubscribe()
        this.subscriptions = new Subscription()
        this.addTooltipEventListeners(nextProps.getTableElement())
    }

    public componentDidMount(): void {
        createTooltips()
        this.addTooltipEventListeners(this.props.getTableElement())
        fetchBlobContentLines({
            repoPath: this.props.repoPath,
            commitID: this.props.commitID,
            filePath: this.props.filePath,
        }).subscribe(
            () => {
                this.setState(() => ({ showOpenFileCTA: true }))
            },
            err => {
                this.setState(() => ({ showOpenFileCTA: false }))
            }
        )

        if (this.props.position) {
            const cell = this.getCodeCell(this.props.position.line)
            if (this.props.position.character) {
                const el = this.props.findElementWithOffset(
                    cell,
                    this.props.position.line,
                    this.props.position.character,
                    this.diffSpec()
                )
                if (el) {
                    el.classList.add('selection-highlight-sticky')
                }
            }
        }

        document.addEventListener('sourcegraph:dismissTooltip', this.onTooltipDismissed)
        window.addEventListener('hashchange', this.handleHashChange)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
        document.removeEventListener('sourcegraph:dismissTooltip', this.onTooltipDismissed)
        window.removeEventListener('hashchange', this.handleHashChange)
    }

    public componentDidUpdate(): void {
        createTooltips()
    }

    public render(): JSX.Element | null {
        let props: OpenInSourcegraphProps
        if (this.isDelta) {
            props = {
                repoPath: this.props.repoPath!,
                filePath: this.props.filePath,
                rev: this.props.commitID!,
                query: {
                    diff: {
                        rev: this.props.commitID,
                    },
                },
            }
        } else {
            props = {
                repoPath: this.props.repoPath,
                filePath: this.props.filePath,
                rev: this.props.rev!,
            }
        }

        let label = 'View File'
        if (this.isDelta) {
            label += this.props.isBase ? ' (base)' : ' (head)'
        }

        return (
            <div style={{ display: 'inline-flex', verticalAlign: 'middle', alignItems: 'center' }}>
                {this.state.showOpenFileCTA && (
                    <OpenOnSourcegraph
                        label={label}
                        ariaLabel="View file on Sourcegraph"
                        openProps={props}
                        className={this.props.buttonProps.className}
                        style={this.props.buttonProps.style}
                        iconStyle={this.props.buttonProps.iconStyle}
                    />
                )}
            </div>
        )
    }

    private addTooltipEventListeners = (ref: HTMLElement): void => {
        if (!ref) {
            return
        }
        this.subscriptions.add(
            this.fixedTooltip
                .pipe(
                    tap(() => {
                        // always hide any existing tooltip when change
                        hideTooltip()
                    }),
                    filter(props => {
                        const position = props.position
                        if (!position || !isCodeIntelligenceEnabled(this.props.filePath)) {
                            this.setFixedTooltip()
                            return false
                        }
                        if (position.line && position.character) {
                            const cell = this.getCodeCell(position.line)
                            const el = this.props.findElementWithOffset(
                                cell,
                                position.line,
                                position.character!,
                                this.diffSpec()
                            )
                            if (el) {
                                const tokenCell = this.props.findTokenCell(cell, el)
                                tokenCell.classList.add('selection-highlight-sticky')
                                return true
                            }
                        }
                        this.setFixedTooltip()
                        return false
                    }),
                    map(props => props.position!),
                    map(pos =>
                        this.props.findElementWithOffset(
                            this.getCodeCell(pos.line),
                            pos.line,
                            pos.character,
                            this.diffSpec()
                        )
                    ),
                    filter((el): el is HTMLElement => !!el),
                    map((target: HTMLElement) => {
                        const loc = this.props.getTargetLineAndOffset(target!, this.diffSpec())
                        const data = { target, loc }
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
                        return zip(
                            this.getTooltip(target, ctx).pipe(
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
                                    // Show the tooltip with j2d and findRef buttons. We don't want to wait for
                                    // a response from getDefinition before showing this, as the response can take some time
                                    // for JS and TS when private packages are used.
                                    updateTooltip(tooltip, true, this.tooltipActions(ctx), this.props.isBase)
                                })
                            ),
                            this.getDefinition(ctx)
                        ).pipe(
                            map(([tooltip, defUrl]) => ({ ...tooltip, defUrl: defUrl || undefined } as TooltipData)),
                            catchError(e => {
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

                    // Hide the previous placeholder tooltip.
                    hideTooltip()
                    this.setFixedTooltip(data)
                    updateTooltip(data, true, this.tooltipActions(data.ctx), this.props.isBase)
                })
        )

        this.subscriptions.add(
            fromEvent<MouseEvent>(ref, 'mouseover', { passive: true })
                .pipe(
                    debounceTime(50),
                    map(e => e.target as HTMLElement),
                    filter(() => isCodeIntelligenceEnabled(this.props.filePath)),
                    filter(this.props.filterTarget),
                    filter(target => {
                        if (!target.lastChild) {
                            return false
                        }
                        if (!target.lastChild.textContent || target.lastChild.textContent.trim().length === 0) {
                            return false
                        }
                        return true
                    }),
                    map(target => {
                        const td = getTableDataCell(target)
                        if (!td) {
                            return target
                        }

                        const cell = this.props.getNodeToConvert(td)
                        if (cell && !cell.classList.contains('annotated')) {
                            cell.classList.add('annotated')
                            convertNode(cell)
                        }
                        return target.classList.contains('wrapped-node') ? target : this.props.findTokenCell(td, target)
                    })
                )
                .pipe(
                    map(target => ({ target, loc: this.props.getTargetLineAndOffset(target, this.diffSpec()) })),
                    filter(data => Boolean(data.loc)),
                    map(data => ({ target: data.target, ctx: { ...this.props, position: data.loc! } })),
                    switchMap(({ target, ctx }) => {
                        const tooltip = this.getTooltip(target, ctx)
                        this.subscriptions.add(tooltip.subscribe(this.logTelemetryOnTooltip))
                        const tooltipWithJ2D: Observable<TooltipData> = zip(tooltip, this.getDefinition(ctx)).pipe(
                            map(([tooltip, defUrl]) => ({ ...tooltip, defUrl: defUrl || undefined }))
                        )
                        const loading = this.getLoadingTooltip(target, ctx, tooltip)
                        return merge(loading, tooltip, tooltipWithJ2D).pipe(
                            catchError(e => {
                                const data: TooltipData = { target, ctx }
                                return [data]
                            })
                        )
                    })
                )
                .subscribe(data => {
                    if (isTooltipVisible(this.props, !this.props.isBase) || isOtherFileTooltipVisible(this.props)) {
                        // If another tooltip is visible for the diff, ignore this mouseover.
                        return
                    }
                    if (!this.state.fixedTooltip) {
                        updateTooltip(data, false, this.tooltipActions(data.ctx), this.props.isBase)
                    }
                })
        )
        this.subscriptions.add(
            fromEvent<MouseEvent>(ref, 'mouseout', { passive: true }).subscribe(e => {
                for (const el of this.props.fileElement.querySelectorAll('.selection-highlight')) {
                    el.classList.remove('selection-highlight')
                }
                if (isTooltipVisible(this.props, !this.props.isBase)) {
                    // If another tooltip is visible for the diff, ignore this mouseover.
                    return
                }
                if (!this.state.fixedTooltip) {
                    if (isTooltipVisible(this.props, this.props.isBase)) {
                        hideTooltip()
                    }
                }
            })
        )
        this.subscriptions.add(
            fromEvent<MouseEvent>(ref, 'mouseup', { passive: true })
                .pipe(
                    debounceTime(50),
                    map(e => e.target as HTMLElement),
                    filter(() => isCodeIntelligenceEnabled(this.props.filePath)),
                    filter(this.props.filterTarget),
                    filter(target => {
                        if (!target) {
                            return false
                        }
                        const tooltip = document.querySelector('.sg-tooltip')
                        if (tooltip && tooltip.contains(target)) {
                            return false
                        }
                        if (!target.lastChild) {
                            return false
                        }
                        if (!target.lastChild.textContent || target.lastChild.textContent.trim().length === 0) {
                            return false
                        }

                        return true
                    })
                )
                .subscribe(target => {
                    const row = (target as Element).closest('tr') as HTMLTableRowElement | null
                    if (!row) {
                        return
                    }
                    for (const el of document.querySelectorAll('.highlighted')) {
                        el.classList.remove('highlighted')
                    }

                    // TODO(isaac): make a prop method for this. The .cov css class is specific to Phabricator
                    // so theres no reason to have that in here, though this doesn't break github.
                    let td = row.lastElementChild
                    if (td) {
                        // Get the last element in the row that is not code coverage annotations in Phabricator.
                        while (td!.classList.contains('cov')) {
                            td = td!.previousElementSibling
                        }

                        td!.classList.add('highlighted')
                    }

                    const position = this.props.getTargetLineAndOffset(target, this.diffSpec())
                    if (position && !this.isDelta) {
                        const hash = '#L' + position.line + (position.character ? ':' + position.character : '')
                        const url = new URL(window.location.href)
                        url.hash = hash
                        if (url.href !== window.location.href && !window.SOURCEGRAPH_PHABRICATOR_EXTENSION) {
                            history.pushState(history.state, '', url.href)
                        }
                    }
                    const nextProps = { ...this.props, position }
                    this.fixedTooltip.next(nextProps)
                })
        )
    }

    private logTelemetryOnTooltip = (data: TooltipData) => {
        // If another tooltip is visible for the diff, ignore this mouseover.
        if (isTooltipVisible(this.props, !this.props.isBase) || isOtherFileTooltipVisible(this.props)) {
            return
        }
        // Only log an event if there is no fixed tooltip docked, we have a
        // target element, and we have tooltip contents
        if (!this.state.fixedTooltip && data.target && !isEmpty(data.contents)) {
            eventLogger.logCodeIntelligenceEvent()
        }
    }

    private diffSpec = () => {
        const spec: MaybeDiffSpec = {
            isDelta: this.isDelta,
            isBase: this.props.isBase,
            isSplitDiff: this.props.isSplitDiff,
        }
        return spec
    }

    private handleHashChange = (e: HashChangeEvent): void => {
        if (e.newURL) {
            const hashIndex = e.newURL.indexOf('#')
            const parsed = parseHash(e.newURL.substr(hashIndex))

            if (!parsed.line) {
                return
            }

            for (const el of document.querySelectorAll('.highlighted')) {
                el.classList.remove('highlighted')
            }
            for (const el of document.querySelectorAll('.selection-highlight-sticky')) {
                el.classList.remove('selection-highlight-sticky')
            }
            const cell = this.getCodeCell(parsed.line) as HTMLElement
            if (parsed.character) {
                const el = this.props.findElementWithOffset(cell, parsed.line, parsed.character, this.diffSpec())
                if (el) {
                    el.classList.add('selection-highlight-sticky')
                }
            }

            // Don't use `this.setFixedTooltip` b/c it will remove sticky selection highlight.
            this.setState({ fixedTooltip: undefined })
            if (isTooltipVisible(this.props, this.props.isBase)) {
                hideTooltip()
            }
        }
    }

    /**
     * getTooltip wraps the asynchronous fetch of tooltip data from the Sourcegraph API.
     * This Observable will emit exactly one value before it completes. If the resolved
     * tooltip is defined, it will update the target styling.
     */
    private getTooltip(target: HTMLElement, ctx: AbsoluteRepoFilePosition): Observable<TooltipData> {
        return this.simpleProviderFns.fetchHover(ctx).pipe(
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
        return fetchJumpURL(this.simpleProviderFns.fetchDefinition, ctx)
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
        return interval(500).pipe(
            take(1),
            takeUntil(tooltip),
            map(() => ({ target, ctx, loading: true }))
        )
    }

    private setFixedTooltip = (data?: TooltipData) => {
        for (const el of this.props.fileElement.querySelectorAll('.selection-highlight')) {
            el.classList.remove('selection-highlight')
        }
        for (const el of this.props.fileElement.querySelectorAll('.selection-highlight-sticky')) {
            el.classList.remove('selection-highlight-sticky')
        }
        if (data) {
            const td = this.getCodeCell(data.ctx.position.line)
            this.props.findTokenCell(td, data.target).classList.add('selection-highlight-sticky')
        } else {
            if (isTooltipVisible(this.props, this.props.isBase)) {
                hideTooltip()
            }
        }
        this.setState({ fixedTooltip: data || undefined })
    }

    private handleGoToDefinition = (defCtx: AbsoluteRepoFilePosition) => (e: MouseEvent) => {
        eventLogger.logCodeIntelligenceEvent()
        if (!window.SOURCEGRAPH_PHABRICATOR_EXTENSION) {
            // Assume it is GitHub, make default j2d be within github pages.
            if (e.shiftKey || e.ctrlKey || e.altKey || e.metaKey) {
                return
            }
            e.preventDefault()
            if (isTooltipVisible(this.props, this.props.isBase)) {
                hideTooltip()
            }
            this.setState({ fixedTooltip: undefined })
            // TODO(john): unsetting fixed tooltip won't work b/c it removes sticky highlight for identity j2d?
            // this.setFixedTooltip()

            // Jump to definition inside of a pull request if the file exists in the PR.
            const sameRepo = this.props.repoPath === defCtx.repoPath
            if (sameRepo && this.props.isPullRequest) {
                const containers = github.getFileContainers()
                for (const container of Array.from(containers)) {
                    const header = container.querySelector('.file-header') as HTMLElement
                    const anchorPath = header.dataset.path
                    if (anchorPath === defCtx.filePath) {
                        const anchorUrl = header.dataset.anchor
                        const url = `${window.location.origin}${window.location.pathname}#${anchorUrl}${
                            this.props.isBase ? 'L' : 'R'
                        }${defCtx.position.line}`
                        window.location.href = url
                        return
                    }
                }
            }

            const rev = sameRepo
                ? this.props.commitID === defCtx.commitID
                    ? this.props.rev
                    : defCtx.commitID || defCtx.rev
                : defCtx.commitID || defCtx.rev
            // tslint:disable-next-line
            const url = `https://${defCtx.repoPath}/blob/${rev || 'HEAD'}/${defCtx.filePath}#L${defCtx.position.line}${
                defCtx.position.character ? ':' + defCtx.position.character : ''
            }`
            window.location.href = url
        }
    }

    private handleFindReferences = (ctx: AbsoluteRepoFilePosition) => (e: MouseEvent) => {
        eventLogger.logCodeIntelligenceEvent()
    }

    private handleDismiss = () => {
        this.setFixedTooltip()
    }

    private updateCodeCells = () => {
        this.cells.clear()
        for (const cell of this.props.getCodeCells()) {
            this.cells.set(cell.line, cell)
        }
    }

    private tooltipActions = (ctx: AbsoluteRepoFilePosition) => ({
        definition: this.handleGoToDefinition,
        references: this.handleFindReferences,
        dismiss: this.handleDismiss,
    })

    private getCodeCell(line: number): HTMLElement {
        let cell: HTMLElement
        // TODO(john): this is less efficient than it needs to be;
        // non-expanding file views don't need to recalculate code cells.
        // if (this.isDelta) {
        this.updateCodeCells()
        cell = this.cells.get(line)!.cell
        // } else {
        //     const table = this.props.getTableElement()
        //     const row = table.rows[line - 1]
        //     cell = row.children[1] as HTMLElement
        // }
        if (!cell.classList.contains('annotated')) {
            convertNode(cell)
            cell.classList.add('annotated')
        }
        return cell
    }

    private onTooltipDismissed = () => {
        if (this.state.fixedTooltip) {
            this.setState({ fixedTooltip: undefined })
        }
    }
}

function isCodeIntelligenceEnabled(filePath: string): boolean {
    const disabledFiles = localStorage.getItem('disabledCodeIntelligenceFiles') || '{}'
    return !JSON.parse(disabledFiles)[`${window.location.pathname}:${filePath}`]
}
