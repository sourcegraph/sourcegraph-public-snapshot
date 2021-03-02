import { createHoverifier, findPositionsFromEvents, HoveredToken, HoverState } from '@sourcegraph/codeintellify'
import { getCodeElementsInRange, locateTarget } from '@sourcegraph/codeintellify/lib/token_position'
import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import * as H from 'history'
import { isEqual, pick } from 'lodash'
import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { combineLatest, from, fromEvent, merge, Observable, ReplaySubject, Subject, Subscription } from 'rxjs'
import { catchError, distinctUntilChanged, filter, map, share, switchMap, withLatestFrom } from 'rxjs/operators'
import { ActionItemAction } from '../../../../shared/src/actions/ActionItem'
import { groupDecorationsByLine } from '../../../../shared/src/api/client/services/decoration'
import { HoverMerged } from '../../../../shared/src/api/client/types/hover'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { getHoverActions } from '../../../../shared/src/hover/actions'
import { HoverContext } from '../../../../shared/src/hover/HoverOverlay'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { asError, ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { isDefined, property } from '../../../../shared/src/util/types'
import {
    AbsoluteRepoFile,
    FileSpec,
    LineOrPositionOrRange,
    lprToSelectionsZeroIndexed,
    ModeSpec,
    parseHash,
    UIPositionSpec,
    RepoSpec,
    ResolvedRevisionSpec,
    RevisionSpec,
    toPositionOrRangeHash,
    toURIWithPath,
} from '../../../../shared/src/util/url'
import { getHover, getDocumentHighlights } from '../../backend/features'
import { WebHoverOverlay } from '../../components/shared'
import { ThemeProps } from '../../../../shared/src/theme'
import { LineDecorator } from './LineDecorator'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'
import { HoverThresholdProps } from '../RepoContainer'
import useDeepCompareEffect from 'use-deep-compare-effect'
import iterate from 'iterare'

/**
 * toPortalID builds an ID that will be used for the {@link LineDecorator} portal containers.
 */
const toPortalID = (line: number): string => `line-decoration-attachment-${line}`

interface BlobProps
    extends SettingsCascadeProps,
        PlatformContextProps,
        TelemetryProps,
        HoverThresholdProps,
        ExtensionsControllerProps,
        ThemeProps {
    location: H.Location
    history: H.History
    className: string
    wrapCode: boolean
    /** The current text document to be rendered and provided to extensions */
    blobInfo: BlobInfo
}

export interface BlobInfo extends AbsoluteRepoFile, ThemeProps, ModeSpec {
    /** The raw content of the blob. */
    content: string

    /** The trusted syntax-highlighted code as HTML */
    html: string
}

const domFunctions = {
    getCodeElementFromTarget: (target: HTMLElement): HTMLTableCellElement | null => {
        // If the target is part of the line decoration attachment, return null.
        if (
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

export const Blob: React.FunctionComponent<BlobProps> = props => {
    const { location, isLightTheme, extensionsController, blobInfo } = props

    // Element reference subjects passed to `hoverifier`. They must be `ReplaySubjects` because
    // the ref callback is called before hoverifier is created in `useEffect`
    const blobElements = useMemo(() => new ReplaySubject<HTMLElement | null>(1), [])
    const nextBlobElement = useCallback((blobElement: HTMLElement | null) => blobElements.next(blobElement), [
        blobElements,
    ])

    const hoverOverlayElements = useMemo(() => new ReplaySubject<HTMLElement | null>(1), [])
    const nextOverlayElement = useCallback(
        (overlayElement: HTMLElement | null) => hoverOverlayElements.next(overlayElement),
        [hoverOverlayElements]
    )

    const codeViewElements = useMemo(() => new ReplaySubject<HTMLElement | null>(1), [])
    const codeViewReference = useRef<HTMLElement | null>()
    const nextCodeViewElement = useCallback(
        (codeView: HTMLElement | null) => {
            codeViewReference.current = codeView
            codeViewElements.next(codeView)
        },
        [codeViewElements]
    )

    // Emits on position changes from URL hash
    const locationPositions = useMemo(() => new ReplaySubject<LineOrPositionOrRange>(1), [])
    const nextLocationPosition = useCallback(
        (lineOrPositionOrRange: LineOrPositionOrRange) => locationPositions.next(lineOrPositionOrRange),
        [locationPositions]
    )
    const parsedHash = useMemo(() => parseHash(location.hash), [location.hash])
    useDeepCompareEffect(() => {
        nextLocationPosition(parsedHash)
    }, [parsedHash])

    // Subject that emits on every render. Source for `hoverOverlayRerenders`
    const rerenders = useMemo(() => new ReplaySubject(1), [])
    useEffect(() => {
        rerenders.next()
    })

    // Emits on blob info changes to update extension host model
    const blobInfoChanges = useMemo(() => new ReplaySubject<BlobInfo>(1), [])
    const nextBlobInfoChange = useCallback((blobInfo: BlobInfo) => blobInfoChanges.next(blobInfo), [blobInfoChanges])
    useEffect(() => {
        nextBlobInfoChange(blobInfo)
    }, [blobInfo, nextBlobInfoChange])

    const closeButtonClicks = useMemo(() => new Subject<MouseEvent>(), [])
    const nextCloseButtonClick = useCallback((click: MouseEvent) => closeButtonClicks.next(click), [closeButtonClicks])

    /** Create hoverifier */
    // We don't want to recreate hoverifier on each render, so props can't be a dependency
    // in useEffect, but hoverifier needs a way to access the latest props.
    const propsReference = useRef<BlobProps>(props)
    propsReference.current = props

    const [hoverState, setHoverState] = useState<HoverState<HoverContext, HoverMerged, ActionItemAction>>({})

    const [decorationsOrError, setDecorationsOrError] = useState<TextDocumentDecoration[] | Error | null>()

    // This effect is meant to run only after first render, cleanup on unmount.
    // TODO: Create a hoverifier class
    useEffect(() => {
        const subscriptions = new Subscription()

        const singleClickGoToDefinition = Boolean(
            propsReference.current.settingsCascade.final &&
                !isErrorLike(propsReference.current.settingsCascade.final) &&
                propsReference.current.settingsCascade.final.singleClickGoToDefinition === true
        )

        const hoverifier = createHoverifier<HoverContext, HoverMerged, ActionItemAction>({
            closeButtonClicks,
            hoverOverlayElements,
            hoverOverlayRerenders: rerenders.pipe(
                withLatestFrom(hoverOverlayElements, blobElements),
                map(([, hoverOverlayElement, blobElement]) => ({
                    hoverOverlayElement,
                    relativeElement: blobElement,
                })),
                filter(property('relativeElement', isDefined)),
                // Can't reposition HoverOverlay if it wasn't rendered
                filter(property('hoverOverlayElement', isDefined))
            ),
            getHover: position =>
                // before, static methods could read from this.props
                {
                    console.log('asking for hover client blob', { ms: Date.now() })

                    /* before, static methods could read from this.props*/
                    return getHover(
                        getLSPTextDocumentPositionParameters(position, propsReference.current.blobInfo.mode),
                        propsReference.current
                    )
                },
            getDocumentHighlights: position =>
                getDocumentHighlights(
                    getLSPTextDocumentPositionParameters(position, propsReference.current.blobInfo.mode),
                    propsReference.current
                ),
            getActions: context => getHoverActions(propsReference.current, context),
            pinningEnabled: !singleClickGoToDefinition,
        })
        subscriptions.add(hoverifier)

        subscriptions.add(
            hoverifier.hoverify({
                positionEvents: codeViewElements.pipe(filter(isDefined), findPositionsFromEvents({ domFunctions })),
                positionJumps: locationPositions.pipe(
                    withLatestFrom(codeViewElements, blobElements),
                    map(([position, codeView, scrollElement]) => ({
                        position,
                        // locationPositions is derived from componentUpdates,
                        // so these elements are guaranteed to have been rendered.
                        codeView: codeView!,
                        scrollElement: scrollElement!,
                    }))
                ),
                resolveContext: () => {
                    const { repoName, revision, commitID, filePath } = propsReference.current.blobInfo
                    return {
                        repoName,
                        revision,
                        commitID,
                        filePath,
                    }
                },
                dom: domFunctions,
            })
        )

        let hoveredTokenElement: HTMLElement | undefined
        let goToDefinition: (event: MouseEvent) => void
        // Make latest hover state accessible to other callbacks in this scope
        // without re-initializing hoverifier. Reassign on each hoverStateUpdates emission
        let latestHoverState: typeof hoverState = {}
        subscriptions.add(
            hoverifier.hoverStateUpdates.subscribe(update => {
                if (singleClickGoToDefinition && hoveredTokenElement !== update.hoveredTokenElement && goToDefinition) {
                    if (hoveredTokenElement) {
                        hoveredTokenElement.style.cursor = 'auto'
                        hoveredTokenElement.removeEventListener('click', goToDefinition)
                    }

                    if (update.hoveredTokenElement) {
                        // Create new goToDefinition function that closes over latest hover state
                        goToDefinition = (event: MouseEvent): void => {
                            const goToDefinitionAction =
                                Array.isArray(update.actionsOrError) &&
                                update.actionsOrError.find(action => action.action.id === 'goToDefinition.preloaded')
                            if (goToDefinitionAction) {
                                propsReference.current.history.push(
                                    goToDefinitionAction.action.commandArguments![0] as string
                                )
                                event.stopPropagation()
                            }
                        }
                        update.hoveredTokenElement.style.cursor = 'pointer'
                        update.hoveredTokenElement.addEventListener('click', goToDefinition)
                    }
                    hoveredTokenElement = update.hoveredTokenElement
                }
                latestHoverState = update
                setHoverState(update)
            })
        )

        // TODO(tj): useObservable
        // When clicking a line, update the URL (which will in turn trigger a highlight of the line)
        subscriptions.add(
            codeViewElements
                .pipe(
                    filter(isDefined),
                    switchMap(codeView => fromEvent<MouseEvent>(codeView, 'click')),
                    // Ignore click events caused by the user selecting text
                    filter(() => !window.getSelection()?.toString())
                )
                .subscribe(event => {
                    // Prevent selecting text on shift click (click+drag to select will still work)
                    // Note that this is only called if the selection was empty initially (see above),
                    // so this only clears a selection caused by this click.
                    window.getSelection()!.removeAllRanges()

                    const position = locateTarget(event.target as HTMLElement, domFunctions)
                    let hash: string
                    if (
                        position &&
                        event.shiftKey &&
                        latestHoverState.selectedPosition &&
                        latestHoverState.selectedPosition.line !== undefined
                    ) {
                        hash = toPositionOrRangeHash({
                            range: {
                                start: {
                                    line: Math.min(latestHoverState.selectedPosition.line, position.line),
                                },
                                end: {
                                    line: Math.max(latestHoverState.selectedPosition.line, position.line),
                                },
                            },
                        })
                    } else {
                        hash = toPositionOrRangeHash({ position })
                    }

                    if (!hash.startsWith('#')) {
                        hash = '#' + hash
                    }

                    propsReference.current.history.push({ ...propsReference.current.location, hash })
                })
        )

        // TODO(tj): useObservable
        // Update selected line when position in hash changes
        subscriptions.add(
            locationPositions.pipe(withLatestFrom(codeViewElements)).subscribe(([position, codeView]) => {
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
                for (const { element } of codeCells) {
                    // Highlight row
                    const row = element.parentElement as HTMLTableRowElement
                    row.classList.add('selected')
                }
            })
        )

        // TODO(tj): useObservable
        // Update the Sourcegraph extensions model to reflect the current file.
        subscriptions.add(
            combineLatest([blobInfoChanges, locationPositions]).subscribe(([blobInfo, position]) => {
                const uri = toURIWithPath(blobInfo)
                if (!propsReference.current.extensionsController.services.model.hasModel(uri)) {
                    propsReference.current.extensionsController.services.model.addModel({
                        uri,
                        languageId: blobInfo.mode,
                        text: blobInfo.content,
                    })
                }

                extensionsController.extHostAPI.then(API => {
                    // ohhhh, might be since we have to unwrap the api... why don't we just do all
                    // 'hoverification' after unwrapping the api once?
                    console.log('request to add document', { ms: Date.now() })
                    API.addTextDocumentIfNotExists({
                        uri,
                        languageId: blobInfo.mode,
                        text: blobInfo.content,
                    })

                    API.addViewerIfNotExists({
                        type: 'CodeEditor' as const,
                        resource: uri,
                        selections: lprToSelectionsZeroIndexed(position),
                        isActive: true,
                    })
                })

                const selections = lprToSelectionsZeroIndexed(position)
                console.log('gonna set mode here', { ms: Date.now() })
                propsReference.current.extensionsController.services.viewer.removeAllViewers()
                // TODO(tj): set editor selections here, but add viewer and model in parent
                // before render completes to speed up extension host processing
                propsReference.current.extensionsController.services.viewer.addViewer({
                    type: 'CodeEditor' as const,
                    resource: uri,
                    selections,
                    isActive: true,
                })
            })
        )

        // TODO(tj): useObservable
        // Get decorations for the current file
        let lastBlobInfo: (AbsoluteRepoFile & ModeSpec) | undefined
        const decorations = blobInfoChanges.pipe(
            switchMap(blobInfo => {
                const blobInfoChanged = !isEqual(blobInfo, lastBlobInfo)
                lastBlobInfo = blobInfo // record so we can compute blobInfoChanged
                // Only clear decorations if the model changed. If only the extensions changed,
                // keep the old decorations until the new ones are available, to avoid UI jitter
                return merge(
                    blobInfoChanged ? [null] : [],
                    propsReference.current.extensionsController.services.textDocumentDecoration.getDecorations({
                        uri: `git://${blobInfo.repoName}?${blobInfo.commitID}#${blobInfo.filePath}`,
                    })
                )
            }),
            share()
        )

        subscriptions.add(
            decorations.pipe(catchError(error => [asError(error)])).subscribe(decorationsOrError => {
                setDecorationsOrError(decorationsOrError)
            })
        )

        return () => {
            subscriptions.unsubscribe()
        }
    }, [
        hoverOverlayElements,
        blobElements,
        codeViewElements,
        rerenders,
        locationPositions,
        blobInfoChanges,
        closeButtonClicks,
    ])

    // Memoize `groupedDecorations` to avoid clearing and setting decorations in `LineDecorator`s on renders in which
    // decorations haven't changed.
    const groupedDecorations = useMemo(
        () => decorationsOrError && !isErrorLike(decorationsOrError) && groupDecorationsByLine(decorationsOrError),
        [decorationsOrError]
    )

    // TODO(tj): Refactor to class component. use lodash memoize for line decorations

    return (
        <div className={`blob ${props.className}`} ref={nextBlobElement}>
            <code
                className={`blob__code ${props.wrapCode ? ' blob__code--wrapped' : ''} test-blob`}
                ref={nextCodeViewElement}
                dangerouslySetInnerHTML={{ __html: blobInfo.html }}
            />
            {hoverState.hoverOverlayProps && (
                <WebHoverOverlay
                    {...props}
                    {...hoverState.hoverOverlayProps}
                    hoverRef={nextOverlayElement}
                    onCloseButtonClick={nextCloseButtonClick}
                    extensionsController={extensionsController}
                />
            )}
            {groupedDecorations &&
                iterate(groupedDecorations)
                    .map(([line, decorations]) => {
                        const portalID = toPortalID(line)
                        return (
                            <LineDecorator
                                isLightTheme={isLightTheme}
                                key={`${portalID}-${blobInfo.filePath}`}
                                portalID={portalID}
                                getCodeElementFromLineNumber={domFunctions.getCodeElementFromLineNumber}
                                line={line}
                                decorations={decorations}
                                codeViewReference={codeViewReference}
                            />
                        )
                    })
                    .toArray()}
        </div>
    )
}

function getLSPTextDocumentPositionParameters(
    position: HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec,
    mode: string
): RepoSpec & RevisionSpec & ResolvedRevisionSpec & FileSpec & UIPositionSpec & ModeSpec {
    return {
        repoName: position.repoName,
        filePath: position.filePath,
        commitID: position.commitID,
        revision: position.revision,
        mode,
        position,
    }
}

interface BlobState extends HoverState<HoverContext, HoverMerged, ActionItemAction> {
    /**
     * lineDecorationAttachmentIDs is a map from line numbers with portal nodes created to portal IDs. It's used to
     * render the portals for {@link LineDecorationAttachment}. The line numbers are taken from the blob so they
     * are 1-indexed.
     */
    lineDecorationAttachmentIDs: { [key: number]: string }

    /** The decorations to display in the blob. */
    decorationsOrError?: TextDocumentDecoration[] | null | ErrorLike
}

class BlobClass extends React.Component<BlobProps, BlobState> {
    /** Emits with the latest Props on every componentDidUpdate and on componentDidMount */
    private componentUpdates = new Subject<BlobProps>()

    /** Emits whenever the ref callback for the code element is called */
    private codeViewElements = new ReplaySubject<HTMLElement | null>(1)
    private nextCodeViewElement = (element: HTMLElement | null): void => this.codeViewElements.next(element)

    /** Emits whenever the ref callback for the blob element is called */
    private blobElements = new Subject<HTMLElement | null>()
    private nextBlobElement = (element: HTMLElement | null): void => this.blobElements.next(element)

    /** Emits whenever the ref callback for the hover element is called */
    private hoverOverlayElements = new Subject<HTMLElement | null>()
    private nextOverlayElement = (element: HTMLElement | null): void => this.hoverOverlayElements.next(element)

    /** Emits when the close button was clicked */
    private closeButtonClicks = new Subject<MouseEvent>()
    private nextCloseButtonClick = (event: MouseEvent): void => this.closeButtonClicks.next(event)

    /** Subscriptions to be disposed on unmout */
    private subscriptions = new Subscription()

    constructor(props: BlobProps) {
        super(props)
        this.state = {
            lineDecorationAttachmentIDs: {},
        }

        /** Emits parsed positions found in the URL */
        const locationPositions: Observable<LineOrPositionOrRange> = this.componentUpdates.pipe(
            map(props => parseHash(props.location.hash)),
            distinctUntilChanged((a, b) => isEqual(a, b)),
            share()
        )

        const singleClickGoToDefinition = Boolean(
            this.props.settingsCascade.final &&
                !isErrorLike(this.props.settingsCascade.final) &&
                this.props.settingsCascade.final.singleClickGoToDefinition === true
        )

        const hoverifier = createHoverifier<
            RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec,
            HoverMerged,
            ActionItemAction
        >({
            closeButtonClicks: this.closeButtonClicks,
            hoverOverlayElements: this.hoverOverlayElements,
            hoverOverlayRerenders: this.componentUpdates.pipe(
                withLatestFrom(this.hoverOverlayElements, this.blobElements),
                // After componentDidUpdate, the blob element is guaranteed to have been rendered
                map(([, hoverOverlayElement, blobElement]) => ({ hoverOverlayElement, relativeElement: blobElement! })),
                // Can't reposition HoverOverlay if it wasn't rendered
                filter(property('hoverOverlayElement', isDefined))
            ),
            getHover: position => getHover(this.getLSPTextDocumentPositionParams(position), this.props),
            getDocumentHighlights: position =>
                getDocumentHighlights(this.getLSPTextDocumentPositionParams(position), this.props),
            getActions: context => getHoverActions(this.props, context),
            pinningEnabled: !singleClickGoToDefinition,
        })
        this.subscriptions.add(hoverifier)

        this.subscriptions.add(
            hoverifier.hoverify({
                positionEvents: this.codeViewElements.pipe(
                    filter(isDefined),
                    findPositionsFromEvents({ domFunctions })
                ),
                positionJumps: locationPositions.pipe(
                    withLatestFrom(this.codeViewElements, this.blobElements),
                    map(([position, codeView, scrollElement]) => ({
                        position,
                        // locationPositions is derived from componentUpdates,
                        // so these elements are guaranteed to have been rendered.
                        codeView,
                        scrollElement,
                    }))
                ),
                resolveContext: () => ({
                    repoName: this.props.blobInfo.repoName,
                    revision: this.props.blobInfo.revision,
                    commitID: this.props.blobInfo.commitID,
                    filePath: this.props.blobInfo.filePath,
                }),
                dom: domFunctions,
            })
        )
        const goToDefinition = (event: MouseEvent): void => {
            const goToDefinitionAction =
                Array.isArray(this.state.actionsOrError) &&
                this.state.actionsOrError.find(action => action.action.id === 'goToDefinition.preloaded')
            if (goToDefinitionAction) {
                this.props.history.push(goToDefinitionAction.action.commandArguments![0] as string)
                event.stopPropagation()
            }
        }

        let hoveredTokenElement: HTMLElement | undefined
        this.subscriptions.add(
            hoverifier.hoverStateUpdates.subscribe(update => {
                if (singleClickGoToDefinition && hoveredTokenElement !== update.hoveredTokenElement) {
                    if (hoveredTokenElement) {
                        hoveredTokenElement.style.cursor = 'auto'
                        hoveredTokenElement.removeEventListener('click', goToDefinition)
                    }
                    if (update.hoveredTokenElement) {
                        update.hoveredTokenElement.style.cursor = 'pointer'
                        update.hoveredTokenElement.addEventListener('click', goToDefinition)
                    }
                    hoveredTokenElement = update.hoveredTokenElement
                }
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
                    filter(() => !window.getSelection()?.toString())
                )
                .subscribe(event => {
                    // Prevent selecting text on shift click (click+drag to select will still work)
                    // Note that this is only called if the selection was empty initially (see above),
                    // so this only clears a selection caused by this click.
                    window.getSelection()!.removeAllRanges()

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
                    this.createLineDecorationAttachmentDOMNode(line, element)
                    // Highlight row
                    const row = element.parentElement as HTMLTableRowElement
                    row.classList.add('selected')
                }
            })
        )

        /** Emits when the URL's target blob (repository, revision, path, and content) changes. */
        const modelChanges: Observable<BlobInfo> = this.componentUpdates.pipe(
            map(({ blobInfo }) => blobInfo),
            distinctUntilChanged((a, b) => isEqual(a, b)),
            share()
        )

        // Update the Sourcegraph extensions model to reflect the current file.
        this.subscriptions.add(
            combineLatest([modelChanges, locationPositions]).subscribe(([model, position]) => {
                const uri = toURIWithPath(model)
                if (!this.props.extensionsController.services.model.hasModel(uri)) {
                    this.props.extensionsController.services.model.addModel({
                        uri,
                        languageId: model.mode,
                        text: model.content,
                    })
                }
                this.props.extensionsController.services.viewer.removeAllViewers()
                this.props.extensionsController.services.viewer.addViewer({
                    type: 'CodeEditor' as const,
                    resource: uri,
                    selections: lprToSelectionsZeroIndexed(position),
                    isActive: true,
                })
            })
        )

        /** Decorations */
        let lastModel: (AbsoluteRepoFile & ModeSpec) | undefined
        const decorations: Observable<TextDocumentDecoration[] | null> = modelChanges.pipe(
            switchMap(model => {
                const modelChanged = !isEqual(model, lastModel)
                lastModel = model // record so we can compute modelChanged

                // Only clear decorations if the model changed. If only the extensions changed, keep
                // the old decorations until the new ones are available, to avoid UI jitter.
                return merge(
                    modelChanged ? [null] : [],
                    this.props.extensionsController.services.textDocumentDecoration.getDecorations({
                        uri: `git://${model.repoName}?${model.commitID}#${model.filePath}`,
                    })
                )
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
            combineLatest([
                decorations.pipe(
                    map(decorations => decorations || []),
                    catchError(error => {
                        console.error(error)

                        // Treat decorations error as empty decorations.
                        return [[] as TextDocumentDecoration[]]
                    })
                ),
                this.codeViewElements,
            ]).subscribe(([decorations, codeView]) => {
                if (codeView) {
                    if (decoratedElements) {
                        // Clear previous decorations.
                        for (const element of decoratedElements) {
                            element.style.backgroundColor = ''
                            element.style.border = ''
                            element.style.borderColor = ''
                            element.style.borderWidth = ''
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
                        const style = decorationStyleForTheme(decoration, this.props.isLightTheme)
                        if (style.backgroundColor) {
                            row.style.backgroundColor = style.backgroundColor
                            decorated = true
                        }
                        if (style.border) {
                            row.style.border = style.border
                            decorated = true
                        }
                        if (style.borderColor) {
                            row.style.borderColor = style.borderColor
                            decorated = true
                        }
                        if (style.borderWidth) {
                            row.style.borderWidth = style.borderWidth
                            decorated = true
                        }
                        if (decorated) {
                            decoratedElements.push(row)
                        }

                        if (decoration.after) {
                            const codeCell = row.cells[1]
                            this.createLineDecorationAttachmentDOMNode(line, codeCell)
                        }
                    }
                } else {
                    decoratedElements = []
                }
            })
        )
    }

    private getLSPTextDocumentPositionParams(
        position: HoveredToken & RepoSpec & RevisionSpec & FileSpec & ResolvedRevisionSpec
    ): RepoSpec & RevisionSpec & ResolvedRevisionSpec & FileSpec & UIPositionSpec & ModeSpec {
        return {
            repoName: position.repoName,
            filePath: position.filePath,
            commitID: position.commitID,
            revision: position.revision,
            mode: this.props.mode,
            position,
        }
    }

    /**
     * Appends a {@link LineDecorationAttachment} portal DOM node to the given code cell if it doesn't contain one
     * already.
     *
     * @param line 1-indexed line number
     * @param codeCell The `<td class="code">` element
     */
    private createLineDecorationAttachmentDOMNode(line: number, codeCell: HTMLElement): void {
        if (codeCell.querySelector('.line-decoration-attachment-portal')) {
            return
        }
        const portalNode = document.createElement('div')

        const id = toPortalID(line)
        portalNode.id = id
        portalNode.classList.add('line-decoration-attachment-portal')

        codeCell.append(portalNode)

        this.setState(state => ({
            lineDecorationAttachmentIDs: {
                ...state.lineDecorationAttachmentIDs,
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
        const { decorationsOrError } = this.state
        const groupedDecorations =
            decorationsOrError && !isErrorLike(decorationsOrError) && groupDecorationsByLine(decorationsOrError)

        const rend =
            groupedDecorations &&
            iterate(groupedDecorations)
                .map(([line, decorations]) => {
                    const portalID = toPortalID(line)
                    return (
                        <LineDecorator
                            isLightTheme={this.props.isLightTheme}
                            key={`${portalID}-${this.props.blobInfo.filePath}`}
                            portalID={portalID}
                            getCodeElementFromLineNumber={domFunctions.getCodeElementFromLineNumber}
                            line={line}
                            decorations={decorations}
                            codeViewReference={codeViewReference}
                            codeViewElements={this.codeViewElements}
                        />
                    )
                })
                .toArray()

        return (
            <div className={`blob ${this.props.className}`} ref={this.nextBlobElement}>
                <code
                    className={`blob__code ${this.props.wrapCode ? ' blob__code--wrapped' : ''} test-blob`}
                    ref={this.nextCodeViewElement}
                    dangerouslySetInnerHTML={{ __html: this.props.html }}
                />
                {this.state.hoverOverlayProps && (
                    <WebHoverOverlay
                        {...this.props}
                        {...this.state.hoverOverlayProps}
                        hoverRef={this.nextOverlayElement}
                        telemetryService={this.props.telemetryService}
                        onCloseButtonClick={this.nextCloseButtonClick}
                    />
                )}
            </div>
        )
    }
}
