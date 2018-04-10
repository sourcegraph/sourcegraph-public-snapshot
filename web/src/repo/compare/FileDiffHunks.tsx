import * as H from 'history'
import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { fromEvent } from 'rxjs/observable/fromEvent'
import { interval } from 'rxjs/observable/interval'
import { merge } from 'rxjs/observable/merge'
import { catchError } from 'rxjs/operators/catchError'
import { debounceTime } from 'rxjs/operators/debounceTime'
import { filter } from 'rxjs/operators/filter'
import { map } from 'rxjs/operators/map'
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { take } from 'rxjs/operators/take'
import { takeUntil } from 'rxjs/operators/takeUntil'
import { tap } from 'rxjs/operators/tap'
import { zip } from 'rxjs/operators/zip'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { AbsoluteRepoFilePosition } from '..'
import { EMODENOTFOUND, fetchHover, fetchJumpURL, isEmptyHover } from '../../backend/lsp'
import { eventLogger } from '../../tracking/eventLogger'
import { asError } from '../../util/errors'
import { toAbsoluteBlobURL, toPrettyBlobURL } from '../../util/url'
import {
    convertNode,
    createTooltips,
    getTableDataCell,
    getTargetLineAndOffset,
    hideTooltip,
    logTelemetryOnTooltip,
    TooltipData,
    updateTooltip,
} from '../blob/tooltips'

const DiffBoundary: React.SFC<{
    /** The "lines" property is set for end boundaries (only for start boundaries and between hunks). */
    oldRange: { startLine: number; lines?: number }
    newRange: { startLine: number; lines?: number }

    section: string | null

    lineNumberClassName: string
    contentClassName: string

    lineNumbers: boolean
}> = props => (
    <tr className="diff-boundary">
        {props.lineNumbers && <td className={`diff-boundary__num ${props.lineNumberClassName}`} colSpan={2} />}
        <td className={`diff-boundary__content ${props.contentClassName}`}>
            {props.oldRange.lines !== undefined &&
                props.newRange.lines !== undefined && (
                    <code>
                        @@ -{props.oldRange.startLine},{props.oldRange.lines} +{props.newRange.startLine},{
                            props.newRange.lines
                        }{' '}
                        {props.section && `@@ ${props.section}`}
                    </code>
                )}
        </td>
    </tr>
)

const DiffHunk: React.SFC<{
    hunk: GQL.IFileDiffHunk
    lineNumbers: boolean
}> = ({ hunk, lineNumbers }) => {
    let oldLine = hunk.oldRange.startLine
    let newLine = hunk.newRange.startLine
    return (
        <>
            <DiffBoundary
                {...hunk}
                lineNumberClassName="diff-hunk__num--both"
                contentClassName="diff-hunk__content"
                lineNumbers={lineNumbers}
            />
            {hunk.body
                .split('\n')
                .slice(0, -1)
                .map((line, i) => {
                    if (line[0] !== '+') {
                        oldLine++
                    }
                    if (line[0] !== '-') {
                        newLine++
                    }
                    return (
                        <tr
                            key={i}
                            className={`diff-hunk__line ${line[0] === ' ' ? 'diff-hunk__line--both' : ''} ${
                                line[0] === '-' ? 'diff-hunk__line--deletion' : ''
                            } ${line[0] === '+' ? 'diff-hunk__line--addition' : ''}`}
                        >
                            {lineNumbers && (
                                <>
                                    {line[0] !== '+' ? (
                                        <td className="diff-hunk__num" data-line={oldLine - 1} data-part="old" />
                                    ) : (
                                        <td className="diff-hunk__num diff-hunk__num--empty" />
                                    )}
                                    {line[0] !== '-' ? (
                                        <td className="diff-hunk__num" data-line={newLine - 1} data-part="new" />
                                    ) : (
                                        <td className="diff-hunk__num diff-hunk__num--empty" />
                                    )}
                                </>
                            )}
                            <td className="diff-hunk__content">{line}</td>
                        </tr>
                    )
                })}
        </>
    )
}

interface Props {
    /** The base repository, revision, and file. */
    base: { repoPath: string; repoID: GQLID; rev: string; commitID: string; filePath: string | null }

    /** The head repository, revision, and file. */
    head: { repoPath: string; repoID: GQLID; rev: string; commitID: string; filePath: string | null }

    /** The file's hunks. */
    hunks: GQL.IFileDiffHunk[]

    /** Whether to show line numbers. */
    lineNumbers: boolean

    className: string
    history: H.History
}

interface State {
    fixedTooltip?: TooltipData
}

/** Displays hunks in a unified file diff. */
export class FileDiffHunks extends React.PureComponent<Props, State> {
    public state: State = {}

    private refSubscriptions: Subscription | undefined
    private fixedTooltip = new Subject<TooltipData>()
    private subscriptions = new Subscription()

    private setElement = (ref: HTMLElement | null): void => {
        if (ref === null) {
            if (this.refSubscriptions) {
                this.refSubscriptions.unsubscribe()
                this.subscriptions.remove(this.refSubscriptions)
                this.refSubscriptions = undefined
            }
            return
        }

        this.refSubscriptions = new Subscription()
        this.subscriptions.add(this.refSubscriptions)

        this.subscriptions.add(
            this.fixedTooltip
                .pipe(
                    startWith(this.state.fixedTooltip || null),
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

        this.refSubscriptions.add(
            merge(fromEvent<MouseEvent>(ref, 'mouseover'), fromEvent<MouseEvent>(ref, 'click'))
                .pipe(
                    debounceTime(50),
                    map(e => ({ type: e.type as 'mouseover' | 'click', target: e.target as HTMLElement })),
                    tap(({ target }) => {
                        createTooltips(ref)

                        const td = getTableDataCell(target, ref)
                        if (td && !td.classList.contains('annotated')) {
                            td.classList.add('annotated')
                            convertNode(td)
                        }
                    }),
                    map(({ type, target }) => ({ type, target, loc: getTargetLineAndOffset(target, ref, true) })),
                    filter(data => !!data.loc && Boolean(data.loc.part)),
                    map(({ type, target, loc }) => ({
                        type,
                        target,
                        ctx: {
                            ...(loc!.part! === 'old' ? this.props.base : this.props.head),
                            position: loc!,
                        } as AbsoluteRepoFilePosition,
                    })),
                    switchMap(({ type, target, ctx }) => {
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
                                return [{ target, ctx } as TooltipData]
                            }),
                            map(data => ({ ...data, type }))
                        )
                    })
                )
                .subscribe(data => {
                    const click = data.type === 'click'
                    logTelemetryOnTooltip(data, click)
                    if (click) {
                        this.fixedTooltip.next(data)
                    } else if (!this.state.fixedTooltip) {
                        updateTooltip(data, false, this.tooltipActions(data.ctx))
                    }
                })
        )

        this.subscriptions.add(
            fromEvent<MouseEvent>(ref, 'mouseout').subscribe(e => {
                for (const el of ref.querySelectorAll('.selection-highlight')) {
                    el.classList.remove('selection-highlight')
                }
                if (!this.state.fixedTooltip) {
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
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <div className={`file-diff-hunks ${this.props.className}`} ref={this.setElement}>
                <div className="file-diff-hunks__container">
                    <table className="file-diff-hunks__table">
                        {this.props.lineNumbers && (
                            <colgroup>
                                <col width="40" />
                                <col width="40" />
                                <col />
                            </colgroup>
                        )}
                        <tbody>
                            {this.props.hunks.map((hunk, i) => (
                                <DiffHunk key={i} hunk={hunk} lineNumbers={this.props.lineNumbers} />
                            ))}
                        </tbody>
                    </table>
                </div>
            </div>
        )
    }

    /**
     * A fixed tooltip is one that is docked. In the web UI, this means the user has
     * clicked on the symbol corresponding to the tooltip. getTooltip and getDefinition
     * is called on the current fixedTooltip, so this should be called whenever there is
     * a new symbol clicked/the tooltip we need information for changes.
     */
    private setFixedTooltip = (data?: TooltipData) => {
        for (const el of document.querySelectorAll('.selection-highlight')) {
            el.classList.remove('selection-highlight')
        }
        for (const el of document.querySelectorAll('.selection-highlight-sticky')) {
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
        hideTooltip()
        this.setFixedTooltip()
        this.props.history.push(toAbsoluteBlobURL(defCtx))
    }

    private handleFindReferences = (ctx: AbsoluteRepoFilePosition) => (e: MouseEvent) => {
        eventLogger.log('FindRefsClicked')
        if (e.shiftKey || e.ctrlKey || e.altKey || e.metaKey) {
            return
        }
        e.preventDefault()
        this.props.history.push(toPrettyBlobURL({ ...ctx, viewState: 'references' }))
    }

    private handleDismiss = () => {
        this.setFixedTooltip()
    }

    private tooltipActions = (ctx: AbsoluteRepoFilePosition) => ({
        definition: this.handleGoToDefinition,
        references: this.handleFindReferences,
        dismiss: this.handleDismiss,
    })
}
