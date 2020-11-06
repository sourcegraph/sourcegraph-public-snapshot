import { createHoverifier, findPositionsFromEvents, HoveredToken, HoverState } from '@sourcegraph/codeintellify'
import { getCodeElementsInRange, locateTarget } from '@sourcegraph/codeintellify/lib/token_position'
import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import * as H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { combineLatest, fromEvent, merge, Observable, ReplaySubject, Subject, Subscription } from 'rxjs'
import { catchError, filter, map, share, switchMap, withLatestFrom } from 'rxjs/operators'
import { ActionItemAction } from '../../../../shared/src/actions/ActionItem'
import { groupDecorationsByLine } from '../../../../shared/src/api/client/services/decoration'
import { HoverMerged } from '../../../../shared/src/api/client/types/hover'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { getHoverActions } from '../../../../shared/src/hover/actions'
import { HoverContext } from '../../../../shared/src/hover/HoverOverlay'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { asError, isErrorLike } from '../../../../shared/src/util/errors'
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
    model: ModelProps
}

export interface ModelProps extends AbsoluteRepoFile, ThemeProps, ModeSpec {
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
    const { location, isLightTheme, extensionsController, model } = props

    // Element reference subjects passed to `hoverifier`. They must be `ReplaySubjects` because
    // the ref callback is called before hoverifier is created in `useEffect`
    const [blobElements, nextBlobElement] = useSubject<HTMLElement | null>(replaySubjectFactory)
    const [hoverOverlayElements, nextOverlayElement] = useSubject<HTMLElement | null>(replaySubjectFactory)

    const [codeViewElements, nextCodeViewElement] = useSubject<HTMLElement | null>(replaySubjectFactory)
    const codeViewReference = React.useRef<HTMLElement | null>()
    const updateCodeView = React.useCallback(
        (codeView: HTMLElement | null) => {
            codeViewReference.current = codeView
            nextCodeViewElement(codeView)
        },
        [nextCodeViewElement]
    )

    // Emits on position changes from URL hash
    const [locationPositions, nextLocationPosition] = useSubject<LineOrPositionOrRange>(replaySubjectFactory)
    const parsedHash = React.useMemo(() => parseHash(location.hash), [location.hash])
    useDeepCompareEffect(() => {
        nextLocationPosition(parsedHash)
    }, [parsedHash])

    // Subject that emits on every render. Source for `hoverOverlayRerenders`
    const [rerenders, nextRerender] = useSubject<void>(replaySubjectFactory)
    React.useEffect(() => {
        nextRerender()
    })

    // Emits on model
    const [modelChanges, nextModelChange] = useSubject<ModelProps>(replaySubjectFactory)
    React.useEffect(() => {
        nextModelChange(model)
    }, [model, nextModelChange])

    const [closeButtonClicks, nextCloseButtonClick] = useSubject<MouseEvent>()

    /** Create hoverifier */
    // We don't want to recreate hoverifier on each render, so props can't be a dependency
    // in useEffect, but hoverifier needs a way to access the latest props.
    const propsReference = React.useRef<BlobProps>(props)
    propsReference.current = props

    const [hoverState, setHoverState] = React.useState<HoverState<HoverContext, HoverMerged, ActionItemAction>>({})

    const [decorationsOrError, setDecorationsOrError] = React.useState<TextDocumentDecoration[] | Error | null>()

    // This effect is meant to run only after first render, cleanup on unmount.
    // TODO: Create a hoverifier React hook
    React.useEffect(() => {
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
                getHover(
                    getLSPTextDocumentPositionParameters(position, propsReference.current.model.mode),
                    propsReference.current
                ),
            getDocumentHighlights: position =>
                getDocumentHighlights(
                    getLSPTextDocumentPositionParameters(position, propsReference.current.model.mode),
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
                    const { repoName, revision, commitID, filePath } = propsReference.current.model
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

        // Update the Sourcegraph extensions model to reflect the current file.
        subscriptions.add(
            combineLatest([modelChanges, locationPositions]).subscribe(([model, position]) => {
                const uri = toURIWithPath(model)
                if (!propsReference.current.extensionsController.services.model.hasModel(uri)) {
                    propsReference.current.extensionsController.services.model.addModel({
                        uri,
                        languageId: model.mode,
                        text: model.content,
                    })
                }
                propsReference.current.extensionsController.services.viewer.removeAllViewers()
                propsReference.current.extensionsController.services.viewer.addViewer({
                    type: 'CodeEditor' as const,
                    resource: uri,
                    selections: lprToSelectionsZeroIndexed(position),
                    isActive: true,
                })
            })
        )

        // Get decorations for the current file
        let lastModel: (AbsoluteRepoFile & ModeSpec) | undefined
        const decorations = modelChanges.pipe(
            switchMap(model => {
                const modelChanged = !isEqual(model, lastModel)
                lastModel = model // record so we can compute modelChanged
                // Only clear decorations if the model changed. If only the extensions changed,
                // keep the old decorations until the new ones are available, to avoid UI jitter
                return merge(
                    modelChanged ? [null] : [],
                    propsReference.current.extensionsController.services.textDocumentDecoration.getDecorations({
                        uri: `git://${model.repoName}?${model.commitID}#${model.filePath}`,
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
        modelChanges,
        closeButtonClicks,
    ])

    // Memoize `groupedDecorations` to avoid clearing and setting decorations in `LineDecorator`s on renders in which
    // decorations haven't changed.
    const groupedDecorations = React.useMemo(
        () => decorationsOrError && !isErrorLike(decorationsOrError) && groupDecorationsByLine(decorationsOrError),
        [decorationsOrError]
    )

    return (
        <div className={`blob ${props.className}`} ref={nextBlobElement}>
            <code
                className={`blob__code ${props.wrapCode ? ' blob__code--wrapped' : ''} test-blob`}
                ref={updateCodeView}
                dangerouslySetInnerHTML={{ __html: props.model.html }}
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
                                key={`${portalID}-${model.filePath}`}
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

/**
 * Creates a ReplaySubject with a buffer size of 1
 */
function replaySubjectFactory<T>(): ReplaySubject<T> {
    return new ReplaySubject<T>(1)
}

/**
 * Unlike `useObservable` or `useEventObservable`, `useSubject` doesn't return the latest value
 * emitted by the Subscribable. Rather, it returns the subject instantiated by the factory
 * as an observable, and a stable `next` function.
 *
 * @param subjectFactory A function that returns a Subject. This must be wrapped in useCallback if it defined inside
 * your component
 */
function useSubject<T>(subjectFactory?: () => Subject<T>): [Observable<T>, (nextValue: T) => void] {
    const subject = React.useMemo(() => (typeof subjectFactory === 'function' ? subjectFactory() : new Subject<T>()), [
        subjectFactory,
    ])
    const observable = React.useMemo(() => subject.asObservable(), [subject])
    const next = React.useCallback((element: T): void => subject.next(element), [subject])

    return [observable, next]
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
