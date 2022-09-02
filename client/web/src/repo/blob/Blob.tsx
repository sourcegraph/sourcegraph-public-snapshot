import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import classNames from 'classnames'
import { Remote } from 'comlink'
import * as H from 'history'
import iterate from 'iterare'
import { isEqual } from 'lodash'
import {
    BehaviorSubject,
    combineLatest,
    merge,
    EMPTY,
    from,
    fromEvent,
    of,
    ReplaySubject,
    Subscription,
    Subject,
} from 'rxjs'
import {
    catchError,
    concatMap,
    distinctUntilChanged,
    filter,
    first,
    map,
    mapTo,
    pairwise,
    switchMap,
    tap,
    throttleTime,
    withLatestFrom,
} from 'rxjs/operators'
import useDeepCompareEffect from 'use-deep-compare-effect'

import { HoverMerged } from '@sourcegraph/client-api'
import {
    getCodeElementsInRange,
    HoveredToken,
    locateTarget,
    findPositionsFromEvents,
    createHoverifier,
    HoverState,
} from '@sourcegraph/codeintellify'
import {
    asError,
    isErrorLike,
    isDefined,
    property,
    observeResize,
    LineOrPositionOrRange,
    lprToSelectionsZeroIndexed,
    toPositionOrRangeQueryParameter,
    addLineRangeQueryParameter,
    formatSearchParameters,
} from '@sourcegraph/common'
import { TextDocumentDecoration } from '@sourcegraph/extension-api-types'
import { ActionItemAction } from '@sourcegraph/shared/src/actions/ActionItem'
import { wrapRemoteObservable } from '@sourcegraph/shared/src/api/client/api/common'
import { FlatExtensionHostAPI } from '@sourcegraph/shared/src/api/contract'
import { groupDecorationsByLine } from '@sourcegraph/shared/src/api/extension/api/decorations'
import { haveInitialExtensionsLoaded } from '@sourcegraph/shared/src/api/features'
import { ViewerId } from '@sourcegraph/shared/src/api/viewerTypes'
import { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import { getHoverActions } from '@sourcegraph/shared/src/hover/actions'
import { HoverContext, PinOptions } from '@sourcegraph/shared/src/hover/HoverOverlay'
import { getModeFromPath } from '@sourcegraph/shared/src/languages'
import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Settings, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { codeCopiedEvent } from '@sourcegraph/shared/src/tracking/event-log-creators'
import {
    AbsoluteRepoFile,
    FileSpec,
    ModeSpec,
    UIPositionSpec,
    RepoSpec,
    ResolvedRevisionSpec,
    RevisionSpec,
    toURIWithPath,
    parseQueryAndHash,
} from '@sourcegraph/shared/src/util/url'
import { Code, useObservable } from '@sourcegraph/wildcard'

import { getHover, getDocumentHighlights } from '../../backend/features'
import { WebHoverOverlay } from '../../components/shared'
import { StatusBar } from '../../extensions/components/StatusBar'
import { BlameHunk } from '../blame/useBlameHunks'
import { HoverThresholdProps } from '../RepoContainer'

import { BlameColumn } from './BlameColumn'
import { LineDecorator } from './LineDecorator'

import styles from './Blob.module.scss'

/**
 * toPortalID builds an ID that will be used for the {@link LineDecorator} portal containers.
 */
const toPortalID = (line: number): string => `line-decoration-attachment-${line}`

export interface BlobProps
    extends SettingsCascadeProps,
        PlatformContextProps<'urlToFile' | 'requestGraphQL' | 'settings'>,
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
    'data-testid'?: string

    // Experimental reference panel
    disableStatusBar: boolean
    disableDecorations: boolean

    // If set, nav is called when a user clicks on a token highlighted by
    // WebHoverOverlay
    nav?: (url: string) => void
    role?: string
    ariaLabel?: string

    blameHunks?: BlameHunk[]
    onHandleFuzzyFinder?: React.Dispatch<React.SetStateAction<boolean>>
}

export interface BlobInfo extends AbsoluteRepoFile, ModeSpec {
    /** The raw content of the blob. */
    content: string

    /** The trusted syntax-highlighted code as HTML */
    html: string

    /** LSIF syntax-highlighting data */
    lsif?: string
}

const domFunctions = {
    getCodeElementFromTarget: (target: HTMLElement): HTMLTableCellElement | null => {
        // If the target is part of the line decoration attachment, return null.
        if (
            target.hasAttribute('data-line-decoration-attachment') ||
            target.hasAttribute('data-line-decoration-attachment-content')
        ) {
            return null
        }

        const row = target.closest('tr')
        if (!row) {
            return null
        }
        return row.querySelector('td.code')
    },
    getCodeElementFromLineNumber: (codeView: HTMLElement, line: number): HTMLTableCellElement | null => {
        const table = codeView.firstElementChild as HTMLTableElement
        const row = table.rows[line - 1]
        if (!row) {
            return null
        }
        return row.querySelector('td.code')
    },
    getLineNumberFromCodeElement: (codeCell: HTMLElement): number => {
        const row = codeCell.closest('tr')
        if (!row) {
            throw new Error('Could not find closest row for codeCell')
        }
        const numberCell = row.querySelector<HTMLTableCellElement>('td.line')
        if (!numberCell || !numberCell.dataset.line) {
            throw new Error('Could not find line number')
        }
        return parseInt(numberCell.dataset.line, 10)
    },
}

const STATUS_BAR_HORIZONTAL_GAP_VAR = '--blob-status-bar-horizontal-gap'
const STATUS_BAR_VERTICAL_GAP_VAR = '--blob-status-bar-vertical-gap'

/**
 * Renders a code view augmented by Sourcegraph extensions
 *
 * Documentation:
 *
 * What is the difference between blobInfoChanges and viewerUpdates?
 *
 * - blobInfoChanges: emits when document info has loaded from the backend (including raw HTML)
 * - viewerUpdates: emits when the extension host confirms that it knows about the current viewer.
 * message to extension host is sent on each blobInfo change, and when that message receives a response
 * with the viewerId (handle to viewer on extension host side), viewerUpdates emits it along with
 * other data (such as subscriptions, extension host API) relevant to observers for this viewer.
 *
 * The possible states that Blob can be in:
 * - "extension host bootstrapping": Initial page load, the initial set of extensions
 * haven't been loaded yet. Regardless of whether or not the extension host knows about
 * the current viewer, users can't interact with extensions yet.
 * - "extension host ready": Extensions have loaded, extension host knows about the current viewer
 * - "extension host loading viewer": Extensions have loaded, but the extension host
 * doesn't know about the current viewer yet. We know that we are in this state
 * when blobInfo changes. On entering this state, clear resources from
 * previous viewer (e.g. hoverifier subscription, line decorations). If we don't remove extension features
 * in this state, hovers can lead to errors like `DocumentNotFoundError`.
 */
export const Blob: React.FunctionComponent<React.PropsWithChildren<BlobProps>> = props => {
    const {
        location,
        isLightTheme,
        extensionsController,
        blobInfo,
        platformContext,
        settingsCascade,
        role,
        ariaLabel,
        'data-testid': dataTestId,
    } = props

    const settingsChanges = useMemo(() => new BehaviorSubject<Settings | null>(null), [])
    useEffect(() => {
        if (
            settingsCascade.final &&
            !isErrorLike(settingsCascade.final) &&
            (!settingsChanges.value || !isEqual(settingsChanges.value, settingsCascade.final))
        ) {
            settingsChanges.next(settingsCascade.final)
        }
    }, [settingsCascade, settingsChanges])

    // Element reference subjects passed to `hoverifier`
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
        // We dangerousSetInnerHTML and modify the <code> element.
        // We need to listen to blobInfo to ensure that we correctly
        // respond whenever this element updates.
        // eslint-disable-next-line react-hooks/exhaustive-deps
        [codeViewElements, blobInfo.html]
    )

    // Emits on changes from URL search params
    const urlSearchParameters = useMemo(() => new ReplaySubject<URLSearchParams>(1), [])
    const nextUrlSearchParameters = useCallback((value: URLSearchParams) => urlSearchParameters.next(value), [
        urlSearchParameters,
    ])
    useEffect(() => {
        nextUrlSearchParameters(new URLSearchParams(location.search))
    }, [nextUrlSearchParameters, location.search])

    // Emits on position changes from URL hash
    const locationPositions = useMemo(() => new ReplaySubject<LineOrPositionOrRange>(1), [])
    const nextLocationPosition = useCallback(
        (lineOrPositionOrRange: LineOrPositionOrRange) => locationPositions.next(lineOrPositionOrRange),
        [locationPositions]
    )
    const parsedHash = useMemo(() => parseQueryAndHash(location.search, location.hash), [
        location.search,
        location.hash,
    ])
    useDeepCompareEffect(() => {
        nextLocationPosition(parsedHash)
    }, [parsedHash])

    // Subject that emits on every render. Source for `hoverOverlayRerenders`, used to
    // reposition hover overlay if needed when `Blob` rerenders
    const rerenders = useMemo(() => new ReplaySubject(1), [])
    useEffect(() => {
        rerenders.next()
    })

    // Emits on blob info changes to update extension host model
    const blobInfoChanges = useMemo(() => new ReplaySubject<BlobInfo>(1), [])
    const nextBlobInfoChange = useCallback((blobInfo: BlobInfo) => blobInfoChanges.next(blobInfo), [blobInfoChanges])

    const viewerUpdates = useMemo(
        () =>
            new BehaviorSubject<{
                viewerId: ViewerId
                blobInfo: BlobInfo
                extensionHostAPI: Remote<FlatExtensionHostAPI>
                subscriptions: Subscription
            } | null>(null),
        []
    )

    useEffect(() => {
        nextBlobInfoChange(blobInfo)
        return () => {
            // Clean up for any resources used by the previous viewer.
            // We can't wait for + don't care about the round trip of
            // client (blobInfo change) -> ext host (add viewer) -> client (receive viewerId)
            // that viewerUpdates emits after.
            viewerUpdates.value?.subscriptions.unsubscribe()

            // Clear viewerUpdates to signify that we are in the "extension host loading viewer" state
            viewerUpdates.next(null)
        }
    }, [blobInfo, nextBlobInfoChange, viewerUpdates])

    const [decorationsOrError, setDecorationsOrError] = useState<TextDocumentDecoration[] | Error | undefined>()

    const popoverCloses = useMemo(() => new Subject<void>(), [])
    const nextPopoverClose = useCallback((click: void) => popoverCloses.next(click), [popoverCloses])

    useObservable(
        useMemo(
            () =>
                popoverCloses.pipe(
                    withLatestFrom(urlSearchParameters),
                    tap(([, parameters]) => {
                        parameters.delete('popover')
                        updateBrowserHistoryIfChanged(props.history, location, parameters)
                    })
                ),
            [location, popoverCloses, props.history, urlSearchParameters]
        )
    )

    const popoverParameter = useMemo(() => urlSearchParameters.pipe(map(parameters => parameters.get('popover'))), [
        urlSearchParameters,
    ])

    const hoverifier = useMemo(
        () =>
            createHoverifier<HoverContext, HoverMerged, ActionItemAction>({
                pinOptions: {
                    pins: popoverParameter.pipe(
                        filter(value => value === 'pinned'),
                        mapTo(undefined)
                    ),
                    closeButtonClicks: merge(
                        popoverCloses,
                        popoverParameter.pipe(
                            pairwise(),
                            filter(([previous, next]) => previous === 'pinned' && next !== 'pinned'),
                            mapTo(undefined)
                        )
                    ),
                },
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
                getHover: context =>
                    getHover(getLSPTextDocumentPositionParameters(context, getModeFromPath(context.filePath)), {
                        extensionsController,
                    }),
                getDocumentHighlights: context =>
                    getDocumentHighlights(
                        getLSPTextDocumentPositionParameters(context, getModeFromPath(context.filePath)),
                        { extensionsController }
                    ),
                getActions: context => getHoverActions({ extensionsController, platformContext }, context),
            }),
        [
            popoverParameter,
            popoverCloses,
            hoverOverlayElements,
            rerenders,
            blobElements,
            extensionsController,
            platformContext,
        ]
    )

    // Update URL when clicking on a line (which will trigger the line highlighting defined below)
    useObservable(
        useMemo(
            () =>
                codeViewElements.pipe(
                    filter(isDefined),
                    switchMap(codeView => fromEvent<MouseEvent>(codeView, 'click')),
                    // Ignore click events caused by the user selecting text
                    filter(() => !window.getSelection()?.toString()),
                    tap(event => {
                        // Prevent selecting text on shift click (click+drag to select will still work)
                        // Note that this is only called if the selection was empty initially (see above),
                        // so this only clears a selection caused by this click.
                        window.getSelection()!.removeAllRanges()

                        const position = locateTarget(event.target as HTMLElement, domFunctions)
                        let query: string | undefined
                        let replace = false
                        if (
                            position &&
                            event.shiftKey &&
                            hoverifier.hoverState.selectedPosition &&
                            hoverifier.hoverState.selectedPosition.line !== undefined
                        ) {
                            // Compare with previous selections (maintained by hoverifier)
                            query = toPositionOrRangeQueryParameter({
                                range: {
                                    start: {
                                        line: Math.min(hoverifier.hoverState.selectedPosition.line, position.line),
                                    },
                                    end: {
                                        line: Math.max(hoverifier.hoverState.selectedPosition.line, position.line),
                                    },
                                },
                            })
                        } else {
                            query = toPositionOrRangeQueryParameter({ position })

                            // Replace the current history entry instead of
                            // adding a new one if the newly selected line is
                            // within 10 lines of the currently selected one.
                            // If the current position is a range a new entry
                            // will always be added.
                            const currentPosition = parseQueryAndHash(location.search, location.hash)
                            replace = Boolean(
                                position &&
                                    currentPosition.line &&
                                    !currentPosition.endLine &&
                                    Math.abs(position.line - currentPosition.line) < 11
                            )
                        }

                        const parameters = new URLSearchParams(location.search)
                        parameters.delete('popover')

                        if (position && !('character' in position)) {
                            // Only change the URL when clicking on blank space on the line (not on
                            // characters). Otherwise, this would interfere with go to definition.
                            updateBrowserHistoryIfChanged(
                                props.history,
                                location,
                                addLineRangeQueryParameter(parameters, query),
                                replace
                            )
                        }
                    }),
                    mapTo(undefined)
                ),
            [codeViewElements, hoverifier.hoverState.selectedPosition, location, props.history]
        )
    )

    // Line highlighting when position in hash changes
    useEffect(() => {
        if (codeViewReference.current) {
            const codeCells = getCodeElementsInRange({
                codeView: codeViewReference.current,
                position: parsedHash,
                getCodeElementFromLineNumber: domFunctions.getCodeElementFromLineNumber,
            })
            // Remove existing highlighting
            for (const selected of codeViewReference.current.querySelectorAll('.selected')) {
                selected.classList.remove('selected')
            }
            for (const { element } of codeCells) {
                // Highlight row
                const row = element.parentElement as HTMLTableRowElement
                row.classList.add('selected')
            }
        }
        // It looks like `parsedHash` is updated _before_ `blobInfo` when
        // navigating between files. That means we have to make this effect
        // dependent on `blobInfo` even if it is not used inside the effect,
        // otherwise the highlighting would not be updated when the new file
        // content is available.
    }, [parsedHash, blobInfo])

    // EXTENSION FEATURES

    // Data source for `viewerUpdates`
    useObservable(
        useMemo(
            () =>
                extensionsController !== null
                    ? combineLatest([
                          blobInfoChanges,
                          // Use the initial position when the document is opened.
                          // Don't want to create new viewers on position change
                          locationPositions.pipe(first()),
                          from(extensionsController.extHostAPI),
                          settingsChanges,
                      ]).pipe(
                          concatMap(([blobInfo, initialPosition, extensionHostAPI]) => {
                              const uri = toURIWithPath(blobInfo)
                              return from(
                                  Promise.all([
                                      // This call should be made before adding viewer, but since
                                      // messages to web worker are handled in order, we can use Promise.all
                                      extensionHostAPI.addTextDocumentIfNotExists({
                                          uri,
                                          languageId: blobInfo.mode,
                                          text: blobInfo.content,
                                      }),
                                      extensionHostAPI.addViewerIfNotExists({
                                          type: 'CodeEditor' as const,
                                          resource: uri,
                                          selections: lprToSelectionsZeroIndexed(initialPosition),
                                          isActive: true,
                                      }),
                                  ])
                              ).pipe(map(([, viewerId]) => ({ viewerId, blobInfo, extensionHostAPI })))
                          }),
                          tap(({ viewerId, blobInfo, extensionHostAPI }) => {
                              const subscriptions = new Subscription()

                              // Cleanup on navigation between/away from viewers
                              subscriptions.add(() => {
                                  extensionHostAPI
                                      .removeViewer(viewerId)
                                      .catch(error => console.error('Error removing viewer from extension host', error))
                              })

                              viewerUpdates.next({ viewerId, blobInfo, extensionHostAPI, subscriptions })
                          }),
                          mapTo(undefined)
                      )
                    : EMPTY,
            [blobInfoChanges, locationPositions, extensionsController, settingsChanges, viewerUpdates]
        )
    )

    // Hoverify
    useObservable(
        useMemo(
            () =>
                viewerUpdates.pipe(
                    filter(isDefined),
                    tap(viewerData => {
                        const subscription = hoverifier.hoverify({
                            positionEvents: codeViewElements.pipe(
                                filter(isDefined),
                                findPositionsFromEvents({ domFunctions })
                            ),
                            positionJumps: locationPositions.pipe(
                                withLatestFrom(
                                    codeViewElements.pipe(filter(isDefined)),
                                    blobElements.pipe(filter(isDefined))
                                ),
                                map(([position, codeView, scrollElement]) => ({
                                    position,
                                    // locationPositions is derived from componentUpdates,
                                    // so these elements are guaranteed to have been rendered.
                                    codeView,
                                    scrollElement,
                                }))
                            ),
                            resolveContext: () => {
                                const { repoName, revision, commitID, filePath } = viewerData.blobInfo
                                return {
                                    repoName,
                                    revision,
                                    commitID,
                                    filePath,
                                }
                            },
                            dom: domFunctions,
                        })
                        viewerData.subscriptions.add(() => subscription.unsubscribe())
                    }),
                    mapTo(undefined)
                ),
            [hoverifier, viewerUpdates, codeViewElements, blobElements, locationPositions]
        )
    )

    // Update position/selections on extension host (extensions use selections to set line decorations)
    useObservable(
        useMemo(
            () =>
                viewerUpdates.pipe(
                    switchMap(viewerData => {
                        if (!viewerData) {
                            return EMPTY
                        }

                        // We can't skip the initial position since we can't guarantee that user hadn't
                        // changed selection between sending the initial message to extension host
                        // for viewer initialization -> receiving viewerId.
                        // The extension host will ensure that extensions are only notified when
                        // selection values have actually changed.
                        return locationPositions.pipe(
                            tap(position => {
                                viewerData.extensionHostAPI
                                    .setEditorSelections(viewerData.viewerId, lprToSelectionsZeroIndexed(position))
                                    .catch(error =>
                                        console.error('Error updating editor selections on extension host', error)
                                    )
                            })
                        )
                    }),
                    mapTo(undefined)
                ),
            [viewerUpdates, locationPositions]
        )
    )

    // Listen for line decorations from extensions
    useObservable(
        useMemo(
            () =>
                props.disableDecorations
                    ? of(undefined)
                    : viewerUpdates.pipe(
                          switchMap(viewerData => {
                              if (!viewerData) {
                                  return EMPTY
                              }

                              // Schedule decorations to be cleared when this viewer is removed.
                              // We store decoration state independent of this observable since we want to clear decorations
                              // immediately on viewer change. If we wait for the latest emission of decorations from the
                              // extension host, decorations from the previous viewer will be visible for a noticeable amount of time
                              // on the current viewer
                              viewerData.subscriptions.add(() => setDecorationsOrError(undefined))
                              return wrapRemoteObservable(
                                  viewerData.extensionHostAPI.getTextDecorations(viewerData.viewerId)
                              )
                          }),
                          catchError(error => [asError(error)]),
                          tap(setDecorationsOrError),
                          mapTo(undefined)
                      ),
            [props.disableDecorations, viewerUpdates]
        )
    )

    // Warm cache for references panel. Eventually display a loading indicator
    useObservable(
        useMemo(
            () =>
                extensionsController !== null ? haveInitialExtensionsLoaded(extensionsController.extHostAPI) : EMPTY,
            [extensionsController]
        )
    )

    // Memoize `groupedDecorations` to avoid clearing and setting decorations in `LineDecorator`s on renders in which
    // decorations haven't changed.
    const groupedDecorations = useMemo(
        () => decorationsOrError && !isErrorLike(decorationsOrError) && groupDecorationsByLine(decorationsOrError),
        [decorationsOrError]
    )

    // Passed to HoverOverlay
    const hoverState: Readonly<HoverState<HoverContext, HoverMerged, ActionItemAction>> =
        useObservable(hoverifier.hoverStateUpdates) || {}

    // Status bar
    const getStatusBarItems = useCallback(
        () =>
            viewerUpdates.pipe(
                switchMap(viewerData => {
                    if (!viewerData) {
                        return of('loading' as const)
                    }

                    return wrapRemoteObservable(viewerData.extensionHostAPI.getStatusBarItems(viewerData.viewerId))
                })
            ),
        [viewerUpdates]
    )

    const statusBarElements = useMemo(() => new ReplaySubject<HTMLDivElement | null>(1), [])
    const nextStatusBarElement = useCallback(
        (statusBarElement: HTMLDivElement | null) => statusBarElements.next(statusBarElement),
        [statusBarElements]
    )

    // Floating status bar: add scrollbar size with "base" gaps to achieve
    // our desired gap between the scrollbar and status bar
    useObservable(
        useMemo(
            () =>
                combineLatest([blobElements, statusBarElements]).pipe(
                    switchMap(([blobElement, statusBarElement]) => {
                        if (!(blobElement && statusBarElement)) {
                            return EMPTY
                        }

                        // ResizeObserver doesn't reliably fire when navigating between documents
                        // in Firefox, so recalculate on blobInfoChanges as well.
                        return merge(observeResize(blobElement), blobInfoChanges).pipe(
                            // Throttle reflow without losing final value.
                            throttleTime(100, undefined, { leading: true, trailing: true }),
                            map(() => {
                                // Read
                                const blobRightScrollbarWidth = blobElement.offsetWidth - blobElement.clientWidth
                                const blobBottomScollbarHeight = blobElement.offsetHeight - blobElement.clientHeight

                                return { blobRightScrollbarWidth, blobBottomScollbarHeight }
                            }),
                            distinctUntilChanged((a, b) => isEqual(a, b)),
                            tap(({ blobRightScrollbarWidth, blobBottomScollbarHeight }) => {
                                // Write
                                statusBarElement.style.right = `calc(var(${STATUS_BAR_HORIZONTAL_GAP_VAR}) + ${blobRightScrollbarWidth}px)`
                                statusBarElement.style.bottom = `calc(var(${STATUS_BAR_VERTICAL_GAP_VAR}) + ${blobBottomScollbarHeight}px)`

                                // Maintain an equal gap with the left side of the container when the status bar is overflowing.
                                statusBarElement.style.maxWidth = `calc(100% - ((2 * var(${STATUS_BAR_HORIZONTAL_GAP_VAR})) + ${blobRightScrollbarWidth}px))`
                            })
                        )
                    }),
                    mapTo(undefined),
                    catchError(() => EMPTY)
                ),
            [blobElements, statusBarElements, blobInfoChanges]
        )
    )

    const pinOptions = useMemo<PinOptions>(
        () => ({
            showCloseButton: true,
            onCloseButtonClick: nextPopoverClose,
            onCopyLinkButtonClick: async () => {
                const line = hoverifier.hoverState.hoveredToken?.line
                const character = hoverifier.hoverState.hoveredToken?.character
                if (line === undefined || character === undefined) {
                    return
                }
                const point = { line, character }
                const range = { start: point, end: point }
                const context = { position: point, range }
                const search = new URLSearchParams(location.search)
                search.set('popover', 'pinned')
                updateBrowserHistoryIfChanged(
                    props.history,
                    location,
                    addLineRangeQueryParameter(search, toPositionOrRangeQueryParameter(context))
                )
                await navigator.clipboard.writeText(window.location.href)
            },
        }),
        [
            hoverifier.hoverState.hoveredToken?.line,
            hoverifier.hoverState.hoveredToken?.character,
            location,
            nextPopoverClose,
            props.history,
        ]
    )

    // Add top and bottom spacers to improve code readability.
    useEffect(() => {
        const subscription = codeViewElements.subscribe(codeView => {
            if (codeView) {
                const table = codeView.firstElementChild as HTMLTableElement
                const firstRow = table.rows[0]
                const lastRow = table.rows[table.rows.length - 1]

                if (firstRow) {
                    for (const cell of firstRow.cells) {
                        if (!cell.querySelector('.top-spacer')) {
                            const spacer = document.createElement('div')
                            spacer.classList.add('top-spacer')
                            cell.prepend(spacer)
                        }
                    }
                }

                if (lastRow) {
                    for (const cell of lastRow.cells) {
                        if (!cell.querySelector('.bottom-spacer')) {
                            const spacer = document.createElement('div')
                            spacer.classList.add('bottom-spacer')
                            cell.append(spacer)
                        }
                    }
                }
            }
        })

        return () => {
            subscription.unsubscribe()
        }
    }, [codeViewElements])

    const logEventOnCopy = useCallback(() => {
        props.telemetryService.log(...codeCopiedEvent('blob'))
    }, [props.telemetryService])

    return (
        <>
            <div
                data-testid={dataTestId}
                className={classNames(props.className, styles.blob)}
                ref={nextBlobElement}
                tabIndex={-1}
                role={role}
                aria-label={ariaLabel}
            >
                <Code
                    className={classNames('test-blob', styles.blobCode, props.wrapCode && styles.blobCodeWrapped)}
                    ref={nextCodeViewElement}
                    onCopy={logEventOnCopy}
                    dangerouslySetInnerHTML={{
                        __html: blobInfo.html,
                    }}
                />
                {hoverState.hoverOverlayProps && extensionsController !== null && (
                    <WebHoverOverlay
                        {...props}
                        {...hoverState.hoverOverlayProps}
                        nav={url => (props.nav ? props.nav(url) : props.history.push(url))}
                        hoveredTokenElement={hoverState.hoveredTokenElement}
                        hoverRef={nextOverlayElement}
                        pinOptions={pinOptions}
                        extensionsController={extensionsController}
                    />
                )}

                {props.blameHunks && <BlameColumn blameHunks={props.blameHunks} codeViewElements={codeViewElements} />}

                {groupedDecorations &&
                    iterate(groupedDecorations)
                        .map(([line, items]) => {
                            const portalID = toPortalID(line)
                            return (
                                <LineDecorator
                                    isLightTheme={isLightTheme}
                                    key={`${portalID}-${blobInfo.filePath}`}
                                    portalID={portalID}
                                    getCodeElementFromLineNumber={domFunctions.getCodeElementFromLineNumber}
                                    line={line}
                                    decorations={items}
                                    codeViewElements={codeViewElements}
                                />
                            )
                        })
                        .toArray()}
            </div>
            {!props.disableStatusBar && extensionsController !== null && window.context.enableLegacyExtensions && (
                <StatusBar
                    getStatusBarItems={getStatusBarItems}
                    extensionsController={extensionsController}
                    uri={toURIWithPath(blobInfo)}
                    location={location}
                    className={styles.blobStatusBarBody}
                    statusBarRef={nextStatusBarElement}
                    hideWhileInitializing={true}
                    isBlobPage={true}
                />
            )}
        </>
    )
}

/**
 * Adds an entry to the browser history only if new search parameters differ
 * from the current ones. This prevents adding a new entry when e.g. the user
 * clicks the same line multiple times.
 */
export function updateBrowserHistoryIfChanged(
    history: H.History,
    location: H.Location,
    newSearchParameters: URLSearchParams,
    /** If set to true replace the current history entry instead of adding a new one. */
    replace: boolean = false
): void {
    const currentSearchParameters = [...new URLSearchParams(location.search).entries()]

    // Update history if the number of search params changes or if any parameter
    // value changes. This will also work for file position changes, which are
    // encoded as parameter without a value. The old file position will be a
    // non-existing key in the new search parameters and thus return `null`
    // (whereas it returns an empty string in the current search parameters).
    const needsUpdate =
        currentSearchParameters.length !== [...newSearchParameters.keys()].length ||
        currentSearchParameters.some(([key, value]) => newSearchParameters.get(key) !== value)

    if (needsUpdate) {
        const entry = {
            ...location,
            search: formatSearchParameters(newSearchParameters),
        }
        if (replace) {
            history.replace(entry)
        } else {
            history.push(entry)
        }
    }
}

export function getLSPTextDocumentPositionParameters(
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
