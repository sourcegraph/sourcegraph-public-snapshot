import {
    createHoverifier,
    findPositionsFromEvents,
    HoveredToken,
    HoveredTokenContext,
    HoverOverlay,
    HoverState,
} from '@sourcegraph/codeintellify'
import { getCodeElementsInRange, locateTarget } from '@sourcegraph/codeintellify/lib/token_position'
import * as H from 'history'
import { isEqual, pick } from 'lodash'
import * as React from 'react'
import { Link, LinkProps } from 'react-router-dom'
import { combineLatest, fromEvent, merge, Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, filter, map, share, switchMap, withLatestFrom } from 'rxjs/operators'
import { TextDocumentDecoration } from 'sourcegraph/module/protocol/plainTypes'
import { AbsoluteRepoFile, RenderMode } from '..'
import { getDecorations, getHover, getJumpURL, ModeSpec } from '../../backend/features'
import { LSPSelector, LSPTextDocumentPositionParams } from '../../backend/lsp'
import { isDiscussionsEnabled } from '../../discussions'
import { ExtensionsDocumentsProps } from '../../extensions/environment/ExtensionsEnvironment'
import {
    ConfigurationCascadeProps,
    ExtensionsControllerProps,
    ExtensionsProps,
} from '../../extensions/ExtensionsClientCommonContext'
import { eventLogger } from '../../tracking/eventLogger'
import { asError, ErrorLike, isErrorLike } from '../../util/errors'
import { isDefined, propertyIsDefined } from '../../util/types'
import { LineOrPositionOrRange, parseHash, toPositionOrRangeHash } from '../../util/url'
import { BlameLine } from './blame/BlameLine'
import { DiscussionsGutterOverlay } from './discussions/DiscussionsGutterOverlay'
import { LineDecorationAttachment } from './LineDecorationAttachment'

/**
 * toPortalID builds an ID that will be used for the blame portal containers.
 */
const toPortalID = (line: number) => `blame-portal-${line}`

interface BlobProps
    extends AbsoluteRepoFile,
        ModeSpec,
        ConfigurationCascadeProps,
        ExtensionsProps,
        ExtensionsDocumentsProps,
        ExtensionsControllerProps {
    /** The raw content of the blob. */
    content: string

    /** The trusted syntax-highlighted code as HTML */
    html: string

    location: H.Location
    history: H.History
    className: string
    wrapCode: boolean
    renderMode: RenderMode
}

interface BlobState extends HoverState {
    /** The desired position of the discussions gutter overlay */
    discussionsGutterOverlayPosition?: { left: number; top: number }

    /**
     * blameLineIDs is a map from line numbers with portal nodes created to portal IDs.
     * It's used to render the portals for blames. The line numbers are taken from the blob
     * so they are 1-indexed.
     */
    blameLineIDs: { [key: number]: string }

    /** The decorations to display in the blob. */
    decorationsOrError?: TextDocumentDecoration[] | null | ErrorLike
}

const logTelemetryEvent = (event: string, data?: any) => eventLogger.log(event, data)
const LinkComponent = (props: LinkProps) => <Link {...props} />

const domFunctions = {
    getCodeElementFromTarget: (target: HTMLElement): HTMLTableCellElement | null => {
        // If the target is part of the blame annotation (a.blame or span.blame__contents), return null.
        if (
            target.classList.contains('blame') ||
            target.classList.contains('blame__contents') ||
            target.classList.contains('line-decoration-attachment') ||
            target.classList.contains('line-decoration-attachment__contents')
        ) {
            return null
        }

        const row = target.closest('tr')
        if (!row) {
            return null
        }
        return row.cells[1]
    },
    getCodeElementFromLineNumber: (codeView: HTMLElement, line: number): HTMLTableCellElement | null => {
        const table = codeView.firstElementChild as HTMLTableElement
        const row = table.rows[line - 1]
        if (!row) {
            return null
        }
        return row.cells[1]
    },
    getLineNumberFromCodeElement: (codeCell: HTMLElement): number => {
        const row = codeCell.closest('tr')
        if (!row) {
            throw new Error('Could not find closest row for codeCell')
        }
        const numberCell = row.cells[0]
        if (!numberCell || !numberCell.dataset.line) {
            throw new Error('Could not find line number')
        }
        return parseInt(numberCell.dataset.line, 10)
    },
}

export class Blob extends React.Component<BlobProps, BlobState> {
    /** Emits with the latest Props on every componentDidUpdate and on componentDidMount */
    private componentUpdates = new Subject<BlobProps>()

    /** Emits whenever the ref callback for the code element is called */
    private codeViewElements = new Subject<HTMLElement | null>()
    private nextCodeViewElement = (element: HTMLElement | null) => this.codeViewElements.next(element)

    /** Emits whenever the ref callback for the blob element is called */
    private blobElements = new Subject<HTMLElement | null>()
    private nextBlobElement = (element: HTMLElement | null) => this.blobElements.next(element)

    /** Emits whenever the ref callback for the hover element is called */
    private hoverOverlayElements = new Subject<HTMLElement | null>()
    private nextOverlayElement = (element: HTMLElement | null) => this.hoverOverlayElements.next(element)

    /** Emits when the go to definition button was clicked */
    private goToDefinitionClicks = new Subject<MouseEvent>()
    private nextGoToDefinitionClick = (event: MouseEvent) => this.goToDefinitionClicks.next(event)

    /** Emits when the close button was clicked */
    private closeButtonClicks = new Subject<MouseEvent>()
    private nextCloseButtonClick = (event: MouseEvent) => this.closeButtonClicks.next(event)

    /** Subscriptions to be disposed on unmout */
    private subscriptions = new Subscription()

    constructor(props: BlobProps) {
        super(props)
        this.state = {
            blameLineIDs: {},
        }

        /** Emits parsed positions found in the URL */
        const locationPositions: Observable<LineOrPositionOrRange> = this.componentUpdates.pipe(
            map(props => parseHash(props.location.hash)),
            distinctUntilChanged((a, b) => isEqual(a, b)),
            share()
        )

        const hoverifier = createHoverifier({
            closeButtonClicks: this.closeButtonClicks,
            goToDefinitionClicks: this.goToDefinitionClicks,
            hoverOverlayElements: this.hoverOverlayElements,
            hoverOverlayRerenders: this.componentUpdates.pipe(
                withLatestFrom(this.hoverOverlayElements, this.blobElements),
                // After componentDidUpdate, the blob element is guaranteed to have been rendered
                map(([, hoverOverlayElement, blobElement]) => ({ hoverOverlayElement, relativeElement: blobElement! })),
                // Can't reposition HoverOverlay if it wasn't rendered
                filter(propertyIsDefined('hoverOverlayElement'))
            ),
            pushHistory: path => this.props.history.push(path),
            logTelemetryEvent,
            fetchHover: position => getHover(this.getLSPTextDocumentPositionParams(position), this.props),
            fetchJumpURL: position => getJumpURL(this.getLSPTextDocumentPositionParams(position), this.props),
        })
        this.subscriptions.add(hoverifier)

        const resolveContext = () => ({
            repoPath: this.props.repoPath,
            rev: this.props.rev,
            commitID: this.props.commitID,
            filePath: this.props.filePath,
        })
        this.subscriptions.add(
            hoverifier.hoverify({
                positionEvents: this.codeViewElements.pipe(
                    filter(isDefined),
                    findPositionsFromEvents(domFunctions)
                ),
                positionJumps: locationPositions.pipe(
                    withLatestFrom(this.codeViewElements, this.blobElements),
                    map(([position, codeView, scrollElement]) => ({
                        position,
                        // locationPositions is derived from componentUpdates,
                        // so these elements are guaranteed to have been rendered.
                        codeView: codeView!,
                        scrollElement: scrollElement!,
                    }))
                ),
                resolveContext,
                dom: domFunctions,
            })
        )
        this.subscriptions.add(
            hoverifier.hoverStateUpdates.subscribe(update => {
                this.setState(update)
            })
        )

        // When clicking a line, update the URL (which will in turn trigger a highlight of the line)
        this.subscriptions.add(
            this.codeViewElements
                .pipe(
                    filter(isDefined),
                    switchMap(codeView => fromEvent<MouseEvent>(codeView, 'click')),
                    // Ignore click events caused by the user selecting text
                    filter(() => window.getSelection().toString() === '')
                )
                .subscribe(event => {
                    // Prevent selecting text on shift click (click+drag to select will still work)
                    // Note that this is only called if the selection was empty initially (see above),
                    // so this only clears a selection caused by this click.
                    window.getSelection().removeAllRanges()

                    const position = locateTarget(event.target as HTMLElement, domFunctions)
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

        // LOCATION CHANGES
        this.subscriptions.add(
            locationPositions.pipe(withLatestFrom(this.codeViewElements)).subscribe(([position, codeView]) => {
                codeView = codeView! // locationPositions is derived from componentUpdates, so this is guaranteed to exist
                const codeCells = getCodeElementsInRange({
                    codeView,
                    position,
                    getCodeElementFromLineNumber: domFunctions.getCodeElementFromLineNumber,
                })
                // Remove existing highlighting
                for (const selected of codeView.querySelectorAll('.selected')) {
                    selected.classList.remove('selected')
                }
                for (const { line, element } of codeCells) {
                    this.createBlameDomNode(line, element)
                    // Highlight row
                    const row = element.parentElement as HTMLTableRowElement
                    row.classList.add('selected')
                }

                // Update overlay position for discussions gutter icon.
                if (codeCells.length > 0) {
                    const blobBounds = codeView.parentElement!.getBoundingClientRect()
                    const row = codeCells[0].element.parentElement as HTMLTableRowElement
                    const targetBounds = row.cells[0].getBoundingClientRect()
                    const left = targetBounds.left - blobBounds.left
                    const top = targetBounds.top + codeView.parentElement!.scrollTop - blobBounds.top
                    this.setState({ discussionsGutterOverlayPosition: { left, top } })
                }
            })
        )

        // EXPERIMENTAL: DECORATIONS

        /** Emits when the URL's target blob (repository, revision, path, and content) changes. */
        const modelChanges: Observable<
            AbsoluteRepoFile & LSPSelector & Pick<BlobProps, 'content'>
        > = this.componentUpdates.pipe(
            map(props => pick(props, 'repoPath', 'rev', 'commitID', 'filePath', 'mode', 'content')),
            distinctUntilChanged((a, b) => isEqual(a, b)),
            share()
        )

        // Update the Sourcegraph extensions environment to reflect the current file.
        this.subscriptions.add(
            combineLatest(modelChanges, locationPositions).subscribe(([model, pos]) => {
                this.props.extensionsOnVisibleTextDocumentsChange([
                    {
                        uri: `git://${model.repoPath}?${model.commitID}#${model.filePath}`,
                        languageId: model.mode,
                        text: model.content,
                    },
                ])
            })
        )
        // Clear the Sourcegraph extensions environment's component when the blob is no longer shown.
        this.subscriptions.add(() => this.props.extensionsOnVisibleTextDocumentsChange(null))

        /** Decorations */
        let lastModel: (AbsoluteRepoFile & LSPSelector) | undefined
        const decorations: Observable<TextDocumentDecoration[] | null> = combineLatest(modelChanges).pipe(
            switchMap(([model]) => {
                const modelChanged = !isEqual(model, lastModel)
                lastModel = model // record so we can compute modelChanged

                // Only clear decorations if the model changed. If only the extensions changed, keep
                // the old decorations until the new ones are available, to avoid UI jitter.
                return merge(modelChanged ? [null] : [], getDecorations(model, this.props))
            }),
            share()
        )
        this.subscriptions.add(
            decorations
                .pipe(catchError(error => [asError(error)]))
                .subscribe(decorationsOrError => this.setState({ decorationsOrError }))
        )

        /** Render decorations. */
        let decoratedElements: HTMLElement[] = []
        this.subscriptions.add(
            combineLatest(
                decorations.pipe(
                    map(decorations => decorations || []),
                    catchError(error => {
                        console.error(error)

                        // Treat decorations error as empty decorations.
                        return [[] as TextDocumentDecoration[]]
                    })
                ),
                this.codeViewElements
            ).subscribe(([decorations, codeView]) => {
                if (codeView) {
                    if (decoratedElements) {
                        // Clear previous decorations.
                        for (const element of decoratedElements) {
                            element.style.backgroundColor = null
                        }
                    }

                    for (const decoration of decorations) {
                        const line = decoration.range.start.line + 1
                        const codeCell = domFunctions.getCodeElementFromLineNumber(codeView, line)
                        if (!codeCell) {
                            continue
                        }
                        const row = codeCell.parentElement as HTMLTableRowElement
                        let decorated = false
                        if (decoration.backgroundColor) {
                            row.style.backgroundColor = decoration.backgroundColor
                            decorated = true
                        }
                        if (decoration.border) {
                            row.style.border = decoration.border
                            decorated = true
                        }
                        if (decoration.borderColor) {
                            row.style.borderColor = decoration.borderColor
                            decorated = true
                        }
                        if (decoration.borderWidth) {
                            row.style.borderWidth = decoration.borderWidth
                            decorated = true
                        }
                        if (decorated) {
                            decoratedElements.push(row)
                        }

                        if (decoration.after) {
                            const codeCell = row.cells[1]!
                            this.createBlameDomNode(line, codeCell)
                        }
                    }
                } else {
                    decoratedElements = []
                }
            })
        )
    }

    private getLSPTextDocumentPositionParams(
        position: HoveredToken & HoveredTokenContext
    ): LSPTextDocumentPositionParams {
        return {
            repoPath: position.repoPath,
            filePath: position.filePath,
            commitID: position.commitID,
            rev: position.rev,
            mode: this.props.mode,
            position,
        }
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
            <div className={`blob ${this.props.className}`} ref={this.nextBlobElement}>
                <code
                    className={`blob__code ${this.props.wrapCode ? ' blob__code--wrapped' : ''} e2e-blob`}
                    ref={this.nextCodeViewElement}
                    dangerouslySetInnerHTML={{ __html: this.props.html }}
                />
                {this.state.hoverOverlayProps && (
                    <HoverOverlay
                        {...this.state.hoverOverlayProps}
                        logTelemetryEvent={logTelemetryEvent}
                        linkComponent={LinkComponent}
                        hoverRef={this.nextOverlayElement}
                        onGoToDefinitionClick={this.nextGoToDefinitionClick}
                        onCloseButtonClick={this.nextCloseButtonClick}
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
                {this.state.decorationsOrError &&
                    !isErrorLike(this.state.decorationsOrError) &&
                    this.state.decorationsOrError
                        .filter(d => !!d.after && this.state.blameLineIDs[d.range.start.line + 1])
                        .map((d, i) => {
                            const line = d.range.start.line + 1
                            return (
                                <LineDecorationAttachment
                                    key={this.state.blameLineIDs[line]}
                                    portalID={this.state.blameLineIDs[line]}
                                    line={line}
                                    attachment={d.after!}
                                    {...this.props}
                                />
                            )
                        })}
                {isDiscussionsEnabled(this.props.configurationCascade) &&
                    this.state.selectedPosition &&
                    this.state.selectedPosition.line !== undefined && (
                        <DiscussionsGutterOverlay
                            overlayPosition={this.state.discussionsGutterOverlayPosition}
                            selectedPosition={this.state.selectedPosition}
                            {...this.props}
                        />
                    )}
            </div>
        )
    }
}
