import DirectionalSignIcon from '@sourcegraph/icons/lib/DirectionalSign'
import * as H from 'history'
import isEmpty from 'lodash/isEmpty'
import isEqual from 'lodash/isEqual'
import omit from 'lodash/omit'
import pick from 'lodash/pick'
import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { combineLatest } from 'rxjs/observable/combineLatest'
import { fromEvent } from 'rxjs/observable/fromEvent'
import { interval } from 'rxjs/observable/interval'
import { merge } from 'rxjs/observable/merge'
import { catchError } from 'rxjs/operators/catchError'
import { debounceTime } from 'rxjs/operators/debounceTime'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { filter } from 'rxjs/operators/filter'
import { map } from 'rxjs/operators/map'
import { mapTo } from 'rxjs/operators/mapTo'
import { startWith } from 'rxjs/operators/startWith'
import { switchMap } from 'rxjs/operators/switchMap'
import { take } from 'rxjs/operators/take'
import { takeUntil } from 'rxjs/operators/takeUntil'
import { tap } from 'rxjs/operators/tap'
import { zip } from 'rxjs/operators/zip'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { Position, Range } from 'vscode-languageserver-types'
import { gql, queryGraphQL } from '../backend/graphql'
import { EMODENOTFOUND, fetchHover, fetchJumpURL, isEmptyHover } from '../backend/lsp'
import { triggerBlame } from '../blame'
import { HeroPage } from '../components/HeroPage'
import { PageTitle } from '../components/PageTitle'
import { Resizable } from '../components/Resizable'
import { ReferencesWidget } from '../references/ReferencesWidget'
import { eventLogger } from '../tracking/eventLogger'
import { getPathExtension, supportedExtensions } from '../util'
import { memoizeObservable } from '../util/memoize'
import { LineOrPositionOrRange, parseHash, toAbsoluteBlobURL, toPrettyBlobURL } from '../util/url'
import { OpenInEditorAction } from './actions/OpenInEditorAction'
import { ToggleLineWrap } from './actions/ToggleLineWrap'
import { ToggleRenderedFileMode } from './actions/ToggleRenderedFileMode'
import {
    AbsoluteRepoFile,
    AbsoluteRepoFilePosition,
    AbsoluteRepoFileRange,
    getCodeCell,
    getCodeCells,
    makeRepoURI,
    ParsedRepoURI,
} from './index'
import { RenderedFile } from './RenderedFile'
import { RepoHeaderActionPortal } from './RepoHeaderActionPortal'
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
 * The same as updateLine, but also scrolls the blob.l
 */
function updateAndScrollToLine(
    cell: HTMLElement | HTMLElement[],
    history: H.History,
    ctx: AbsoluteRepoFileRange,
    clickEvent?: MouseEvent
): void {
    if (!Array.isArray(cell)) {
        cell = [cell]
    }

    updateLine(cell, history, ctx, clickEvent)

    // Scroll to the line.
    scrollToCell(cell[0])
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
}

interface State {
    fixedTooltip?: TooltipData
}

class Blob extends React.Component<Props, State> {
    public state: State = {}
    private blobElement: HTMLElement | null = null
    private fixedTooltip = new Subject<Props>()
    private subscriptions = new Subscription()

    public componentWillReceiveProps(nextProps: Props): void {
        const hash = parseHash(nextProps.location.hash)
        if (this.props.location.pathname !== nextProps.location.pathname && !hash.line) {
            if (this.blobElement) {
                this.blobElement.scrollTop = 0
            }
        }
        const thisHash = parseHash(this.props.location.hash)
        const nextHash = parseHash(nextProps.location.hash)
        if (thisHash.modal !== nextHash.modal && this.props.location.pathname === nextProps.location.pathname) {
            // When updating references mode in the same file, scroll. Wait just a moment to make sure the references
            // panel is shown, since the scroll offset is calculated based on the height of the blob.
            setTimeout(() => this.scrollToLine(nextProps), 10)
        }

        if (
            this.props.location.pathname === nextProps.location.pathname &&
            (thisHash.character !== nextHash.character ||
                thisHash.line !== nextHash.line ||
                thisHash.endCharacter !== nextHash.endCharacter ||
                thisHash.endLine !== nextHash.endLine ||
                thisHash.modal !== nextHash.modal)
        ) {
            if (!nextHash.modal) {
                this.fixedTooltip.next(nextProps)
            } else {
                // If showing modal, remove any tooltip then highlight the element for the given start position.
                this.setFixedTooltip()
                if (nextHash.line && nextHash.character) {
                    const el = findElementWithOffset(
                        getCodeCell(nextHash.line!).childNodes[1]! as HTMLElement,
                        nextHash.character!
                    )
                    if (el) {
                        el.classList.add('selection-highlight-sticky')
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
                this.addEventListeners(this.blobElement)
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
        if (!parsedHash.modal) {
            // Show fixed tooltip if necessary iff not showing a modal.
            this.fixedTooltip.next(this.props)
        }
        // The HTML contents were updated on a mounted component, e.g. from a 'back' or 'forward' event,
        // or a jump-to-def.
        this.scrollToLine(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
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
            this.addEventListeners(ref)
            const parsedHash = parseHash(this.props.location.hash)
            if (parsedHash.line && parsedHash.character) {
                this.fixedTooltip.next(this.props)
            }
        }
    }

    private addEventListeners = (ref: HTMLElement): void => {
        const isSupportedExtension = supportedExtensions.has(getPathExtension(this.props.filePath))
        if (isSupportedExtension) {
            this.subscriptions.add(
                this.fixedTooltip
                    .pipe(
                        filter(props => {
                            const parsed = parseHash(props.location.hash)
                            if (parsed.line && parsed.character) {
                                const td = getCodeCell(parsed.line).childNodes[1] as HTMLTableDataCellElement
                                if (td && !td.classList.contains('annotated')) {
                                    td.classList.add('annotated')
                                    convertNode(td)
                                }
                                if (!parsed.modal) {
                                    return true
                                }
                                // Don't show a tooltip when there is a modal (but do highlight the token)
                                // TODO(john): this can probably be simplified.
                                const el = findElementWithOffset(
                                    getCodeCell(parsed.line!).childNodes[1]! as HTMLElement,
                                    parsed.character!
                                )
                                if (el) {
                                    el.classList.add('selection-highlight-sticky')
                                    return false
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
                                zip(this.getDefinition(ctx)),
                                map(
                                    ([tooltip, defUrl]) => ({ ...tooltip, defUrl: defUrl || undefined } as TooltipData)
                                ),
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
                            const tooltipWithJ2D: Observable<TooltipData> = tooltip.pipe(
                                zip(this.getDefinition(ctx)),
                                map(([tooltip, defUrl]) => ({ ...tooltip, defUrl: defUrl || undefined }))
                            )
                            const loading = this.getLoadingTooltip(target, ctx, tooltip)
                            return merge(loading, tooltip, tooltipWithJ2D).pipe(
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

    private setFixedTooltip = (data?: TooltipData) => {
        for (const el of document.querySelectorAll('.blob .selection-highlight')) {
            el.classList.remove('selection-highlight')
        }
        for (const el of document.querySelectorAll('.blob .selection-highlight-sticky')) {
            el.classList.remove('selection-highlight-sticky')
        }
        if (data) {
            eventLogger.log('TooltipDocked')
            data.target.classList.add('selection-highlight-sticky')
        } else {
            hideTooltip()
        }
        this.setState({ fixedTooltip: data || undefined })
    }

    private scrollToLine = (props: Props) => {
        const parsed = parseHash(props.location.hash)
        const { line, character, endLine, endCharacter, modalMode } = parsed
        if (line) {
            updateAndScrollToLine(getCodeCells(line, endLine), props.history, {
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
                referencesMode: modalMode,
            })
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
        this.props.history.push(toPrettyBlobURL({ ...ctx, rev: this.props.rev, referencesMode: 'local' }))
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
                eventLogger.log('SymbolHovered', { hoverHasDefUrl: data.defUrl !== undefined })
            }
        }
    }

    private tooltipActions = (ctx: AbsoluteRepoFilePosition) => ({
        definition: this.handleGoToDefinition,
        references: this.handleFindReferences,
        dismiss: this.handleDismiss,
    })
}

export function fetchBlobCacheKey(parsed: ParsedRepoURI & { isLightTheme: boolean; disableTimeout: boolean }): string {
    return makeRepoURI(parsed) + parsed.isLightTheme + parsed.disableTimeout
}

const fetchBlob = memoizeObservable(
    (args: {
        repoPath: string
        commitID: string
        filePath: string
        isLightTheme: boolean
        disableTimeout: boolean
    }): Observable<GQL.IFile> =>
        queryGraphQL(
            gql`
                query Blob(
                    $repoPath: String!
                    $commitID: String!
                    $filePath: String!
                    $isLightTheme: Boolean!
                    $disableTimeout: Boolean!
                ) {
                    repository(uri: $repoPath) {
                        commit(rev: $commitID) {
                            file(path: $filePath) {
                                richHTML
                                highlight(disableTimeout: $disableTimeout, isLightTheme: $isLightTheme) {
                                    aborted
                                    html
                                }
                            }
                        }
                    }
                }
            `,
            args
        ).pipe(
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.repository ||
                    !data.repository.commit ||
                    !data.repository.commit.file ||
                    !data.repository.commit.file.highlight
                ) {
                    throw Object.assign(
                        'Could not fetch blob content: ' + new Error((errors || []).map(e => e.message).join('\n')),
                        { errors }
                    )
                }
                return data.repository.commit.file
            })
        ),
    fetchBlobCacheKey
)

interface BlobPageProps {
    location: H.Location
    history: H.History
    isLightTheme: boolean
    repoPath: string
    rev: string | undefined
    commitID: string
    filePath: string
}

interface BlobPageState {
    loading: boolean
    error?: string
    wrapCode: boolean

    /**
     * Whether to show the references panel.
     */
    showRefs: boolean

    /**
     * The blob data.
     */
    blob?: GQL.IFile
}

export class BlobPage extends React.PureComponent<BlobPageProps, BlobPageState> {
    private propsUpdates = new Subject<BlobPageProps>()
    private extendHighlightingTimeoutClicks = new Subject<void>()
    private subscriptions = new Subscription()

    constructor(props: BlobPageProps) {
        super(props)

        this.state = {
            loading: true,
            wrapCode: ToggleLineWrap.getValue(),
            showRefs: parseHash(props.location.hash).modal === 'references',
        }
    }

    private logViewEvent(): void {
        eventLogger.logViewEvent('Blob', { fileShown: true, referencesShown: this.state.showRefs })
    }

    public componentDidMount(): void {
        this.logViewEvent()

        // Fetch repository revision.
        this.subscriptions.add(
            combineLatest(
                this.propsUpdates.pipe(
                    map(props => pick(props, 'repoPath', 'commitID', 'filePath', 'isLightTheme')),
                    distinctUntilChanged((a, b) => isEqual(a, b))
                ),
                this.extendHighlightingTimeoutClicks.pipe(mapTo(true), startWith(false))
            )
                .pipe(
                    tap(() => this.setState({ loading: true, blob: undefined, error: undefined })),
                    switchMap(([{ repoPath, commitID, filePath, isLightTheme }, extendHighlightingTimeout]) =>
                        fetchBlob({
                            repoPath,
                            commitID,
                            filePath,
                            isLightTheme,
                            disableTimeout: extendHighlightingTimeout,
                        }).pipe(
                            catchError(error => {
                                console.error(error)
                                this.setState({ loading: false, error })
                                return []
                            })
                        )
                    )
                )
                .subscribe(blob => this.setState({ loading: false, blob, error: undefined }), err => console.error(err))
        )

        this.propsUpdates.next(this.props)
    }

    public componentWillReceiveProps(newProps: BlobPageProps): void {
        this.propsUpdates.next(newProps)
        if (
            newProps.repoPath !== this.props.repoPath ||
            newProps.commitID !== this.props.commitID ||
            newProps.filePath !== this.props.filePath ||
            ToggleRenderedFileMode.getModeFromURL(newProps.location) !==
                ToggleRenderedFileMode.getModeFromURL(this.props.location)
        ) {
            this.logViewEvent()
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): React.ReactNode {
        if (this.state.loading) {
            // Render placeholder for layout before content is fetched.
            return <div className="blob-page__placeholder" />
        }

        if (!this.state.blob) {
            return (
                <HeroPage
                    icon={DirectionalSignIcon}
                    title="404: Not Found"
                    subtitle="The requested file was not found."
                />
            )
        }

        const renderMode = ToggleRenderedFileMode.getModeFromURL(this.props.location)
        const hash = parseHash(this.props.location.hash)

        return [
            <PageTitle key="page-title" title={this.getPageTitle()} />,
            <RepoHeaderActionPortal
                position="right"
                key="open-in-editor"
                element={
                    <OpenInEditorAction
                        key="open-in-editor"
                        repoPath={this.props.repoPath}
                        commitID={this.props.commitID}
                        filePath={this.props.filePath}
                        location={this.props.location}
                    />
                }
            />,
            <RepoHeaderActionPortal
                position="right"
                key="toggle-line-wrap"
                element={<ToggleLineWrap key="toggle-line-wrap" onDidUpdate={this.onDidUpdateLineWrap} />}
            />,
            this.state.blob.richHTML && (
                <RepoHeaderActionPortal
                    key="toggle-rendered-file-mode"
                    position="right"
                    element={
                        <ToggleRenderedFileMode
                            key="toggle-rendered-file-mode"
                            mode={renderMode}
                            location={this.props.location}
                        />
                    }
                />
            ),
            this.state.blob.richHTML &&
                renderMode === 'rendered' && (
                    <RenderedFile key="rendered-file" dangerousInnerHTML={this.state.blob.richHTML} />
                ),
            (renderMode === 'code' || !this.state.blob.richHTML) &&
                !this.state.blob.highlight.aborted && (
                    <Blob
                        key="blob"
                        className="blob-page__blob"
                        repoPath={this.props.repoPath}
                        commitID={this.props.commitID}
                        filePath={this.props.filePath}
                        html={this.state.blob.highlight.html}
                        rev={this.props.rev}
                        wrapCode={this.state.wrapCode}
                        location={this.props.location}
                        history={this.props.history}
                    />
                ),
            !this.state.blob.richHTML &&
                this.state.blob.highlight.aborted && (
                    <div className="blob-page__aborted" key="aborted">
                        <div className="alert alert-notice">
                            Syntax-highlighting this file took too long. &nbsp;
                            <button onClick={this.onExtendHighlightingTimeoutClick} className="btn btn-sm btn-primary">
                                Try again
                            </button>
                        </div>
                    </div>
                ),
            hash.modal === 'references' &&
                hash.line && (
                    <Resizable
                        key="blob-page-references"
                        className="blob-page__panel--resizable"
                        handlePosition="top"
                        defaultSize={350}
                        storageKey="blob-page-references"
                        element={
                            <ReferencesWidget
                                key="refs"
                                repoPath={this.props.repoPath}
                                commitID={this.props.commitID}
                                rev={this.props.rev}
                                referencesMode={hash.modalMode}
                                filePath={this.props.filePath}
                                position={{ line: hash.line, character: hash.character || 0 }}
                                location={this.props.location}
                                history={this.props.history}
                                isLightTheme={this.props.isLightTheme}
                            />
                        }
                    />
                ),
        ]
    }

    private onDidUpdateLineWrap = (value: boolean) => this.setState({ wrapCode: value })

    private onExtendHighlightingTimeoutClick = () => this.extendHighlightingTimeoutClicks.next()

    private getPageTitle(): string {
        const repoPathSplit = this.props.repoPath.split('/')
        const repoStr = repoPathSplit.length > 2 ? repoPathSplit.slice(1).join('/') : this.props.repoPath
        if (this.props.filePath) {
            const fileOrDir = this.props.filePath.split('/').pop()
            return `${fileOrDir} - ${repoStr}`
        }
        return `${repoStr}`
    }
}
