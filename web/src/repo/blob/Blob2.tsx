import { createHoverifier, HoverOverlay, HoverState } from '@sourcegraph/codeintellify'
import { getRowInCodeElement, getRowsInRange } from '@sourcegraph/codeintellify/lib/token_position'
import * as H from 'history'
import { isEqual, pick } from 'lodash'
import * as React from 'react'
import { Link, LinkProps } from 'react-router-dom'
import { combineLatest, fromEvent, merge, Observable, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, filter, map, share, switchMap, withLatestFrom } from 'rxjs/operators'
import { Position } from 'vscode-languageserver-types'
import { AbsoluteRepoFile, RenderMode } from '..'
import { ExtensionsProps, getDecorations, getHover, getJumpURL, ModeSpec } from '../../backend/features'
import { LSPSelector, LSPTextDocumentPositionParams, TextDocumentDecoration } from '../../backend/lsp'
import { eventLogger } from '../../tracking/eventLogger'
import { asError, ErrorLike, isErrorLike } from '../../util/errors'
import { toNativeEvent } from '../../util/react'
import { isDefined, propertyIsDefined } from '../../util/types'
import { LineOrPositionOrRange, parseHash, toPositionOrRangeHash } from '../../util/url'
import { BlameLine } from './blame/BlameLine'
import { DiscussionsGutterOverlay } from './discussions/DiscussionsGutterOverlay'
import { LineDecorationAttachment } from './LineDecorationAttachment'
import { locateTarget } from './tooltips'

/**
 * toPortalID builds an ID that will be used for the blame portal containers.
 */
const toPortalID = (line: number) => `blame-portal-${line}`

interface BlobProps extends AbsoluteRepoFile, ModeSpec, ExtensionsProps {
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
    decorationsOrError?: TextDocumentDecoration[] | ErrorLike
}

const logTelemetryEvent = (event: string, data?: any) => eventLogger.log(event, data)
const LinkComponent = (props: LinkProps) => <Link {...props} />

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
    private nextCodeClick = (event: React.MouseEvent<HTMLElement>) => this.codeClicks.next(event)

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
            blameLineIDs: {},
        }

        /** Emits parsed positions found in the URL */
        const locationPositions: Observable<LineOrPositionOrRange> = this.componentUpdates.pipe(
            map(props => parseHash(props.location.hash)),
            distinctUntilChanged((a, b) => isEqual(a, b)),
            share()
        )

        const hoverifier = createHoverifier({
            closeButtonClicks: this.closeButtonClicks.pipe(map(toNativeEvent)),
            goToDefinitionClicks: this.goToDefinitionClicks.pipe(map(toNativeEvent)),
            hoverOverlayElements: this.hoverOverlayElements,
            hoverOverlayRerenders: this.componentUpdates.pipe(
                withLatestFrom(this.hoverOverlayElements, this.blobElements),
                // After componentDidUpdate, the blob element is guaranteed to have been rendered
                map(([, hoverOverlayElement, blobElement]) => ({ hoverOverlayElement, scrollElement: blobElement! })),
                // Can't reposition HoverOverlay if it wasn't rendered
                filter(propertyIsDefined('hoverOverlayElement'))
            ),
            pushHistory: path => this.props.history.push(path),
            logTelemetryEvent,
            fetchHover: position => getHover(this.getLSPTextDocumentPositionParams(position), this.props.extensions),
            fetchJumpURL: position =>
                getJumpURL(this.getLSPTextDocumentPositionParams(position), this.props.extensions),
        })
        this.subscriptions.add(hoverifier)

        // Get the native event objects by using fromEvent directly on the element,
        // as React does dark magic with event objects that messes with the hoverify logic
        // (currentTarget can have unexpected values)
        const fromCodeElementEvent = (eventName: string) =>
            this.codeElements.pipe(
                filter(isDefined),
                switchMap(codeElement => fromEvent<MouseEvent>(codeElement, eventName))
            )

        const resolveContext = () => ({
            repoPath: this.props.repoPath,
            rev: this.props.rev,
            commitID: this.props.commitID,
            filePath: this.props.filePath,
        })
        this.subscriptions.add(
            hoverifier.hoverify({
                codeMouseMoves: fromCodeElementEvent('mousemove'),
                codeMouseOvers: fromCodeElementEvent('mouseover'),
                codeClicks: fromCodeElementEvent('click'),
                positionJumps: locationPositions.pipe(
                    withLatestFrom(this.codeElements, this.blobElements),
                    map(([position, codeElement, scrollElement]) => ({
                        position,
                        // locationPositions is derived from componentUpdates,
                        // so these elements are guaranteed to have been rendered.
                        codeElement: codeElement!,
                        scrollElement: scrollElement!,
                        ...resolveContext(),
                    }))
                ),
                resolveContext,
            })
        )
        this.subscriptions.add(
            hoverifier.hoverStateUpdates.subscribe(update => {
                this.setState(update)
            })
        )

        // When clicking a line, update the URL (which will in turn trigger a highlight of the line)
        this.subscriptions.add(
            this.codeClicks
                .pipe(
                    // Ignore click events caused by the user selecting text
                    filter(() => window.getSelection().toString() === ''),
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

        // LOCATION CHANGES
        this.subscriptions.add(
            locationPositions.pipe(withLatestFrom(this.codeElements)).subscribe(([position, codeElement]) => {
                codeElement = codeElement! // locationPositions is derived from componentUpdates, so this is guaranteed to exist
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

                // Update overlay position for discussions gutter icon.
                if (rows.length > 0) {
                    const blobBounds = codeElement.parentElement!.getBoundingClientRect()
                    const targetBounds = rows[0].element.cells[0].getBoundingClientRect()
                    const left = targetBounds.left - blobBounds.left
                    const top = targetBounds.top + codeElement.parentElement!.scrollTop - blobBounds.top
                    this.setState({ discussionsGutterOverlayPosition: { left, top } })
                }
            })
        )

        // EXPERIMENTAL: DECORATIONS

        /** Emits the extensions when they change. */
        const extensionsChanges = this.componentUpdates.pipe(
            map(({ extensions }) => extensions),
            distinctUntilChanged(isEqual),
            share()
        )

        /** Emits when the URL's target blob (repository, revision, and path) changes. */
        const modelChanges: Observable<AbsoluteRepoFile & LSPSelector> = this.componentUpdates.pipe(
            map(props => pick(props, 'repoPath', 'rev', 'commitID', 'filePath', 'mode')),
            distinctUntilChanged((a, b) => isEqual(a, b)),
            share()
        )

        /** Decorations */
        let lastModel: (AbsoluteRepoFile & LSPSelector) | undefined
        const decorations: Observable<TextDocumentDecoration[] | undefined> = combineLatest(
            modelChanges,
            extensionsChanges
        ).pipe(
            switchMap(([model, extensions]) => {
                const modelChanged = !isEqual(model, lastModel)
                lastModel = model // record so we can compute modelChanged

                // Only clear decorations if the model changed. If only the extensions changed, keep
                // the old decorations until the new ones are available, to avoid UI jitter.
                return merge(modelChanged ? [undefined] : [], getDecorations(model, extensions))
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
                this.codeElements
            )
                .pipe(map(([decorations, codeElement]) => ({ decorations, codeElement })))
                .subscribe(({ decorations, codeElement }) => {
                    if (codeElement) {
                        if (decoratedElements) {
                            // Clear previous decorations.
                            for (const e of decoratedElements) {
                                e.style.backgroundColor = null
                            }
                        }

                        for (const d of decorations) {
                            const lineElement = getRowInCodeElement(codeElement, d.range.start.line + 1)
                            if (lineElement && d.backgroundColor) {
                                lineElement.style.backgroundColor = d.backgroundColor
                                decoratedElements.push(lineElement)
                            }
                        }
                    } else {
                        decoratedElements = []
                    }
                })
        )
    }

    private getLSPTextDocumentPositionParams(position: Position): LSPTextDocumentPositionParams {
        return {
            repoPath: this.props.repoPath,
            filePath: this.props.filePath,
            commitID: this.props.commitID,
            mode: this.props.mode,
            rev: this.props.rev,
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
            <div className={`blob2 ${this.props.className}`} ref={this.nextBlobElement}>
                <code
                    className={`blob2__code ${this.props.wrapCode ? ' blob2__code--wrapped' : ''} e2e-blob`}
                    ref={this.nextCodeElement}
                    dangerouslySetInnerHTML={{ __html: this.props.html }}
                    onClick={this.nextCodeClick}
                    onMouseOver={this.nextCodeMouseOver}
                    onMouseMove={this.nextCodeMouseMove}
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
                {window.context.discussionsEnabled &&
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
