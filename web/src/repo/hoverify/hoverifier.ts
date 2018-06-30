import * as H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { combineLatest, concat, fromEvent, merge, Observable, of, Subject, Subscription, zip } from 'rxjs'
import {
    catchError,
    debounceTime,
    delay,
    distinctUntilChanged,
    filter,
    map,
    share,
    switchMap,
    takeUntil,
    tap,
    withLatestFrom,
} from 'rxjs/operators'
import { Key } from 'ts-key-enum'
import { Position } from 'vscode-languageserver-types'
import { FileSpec, RepoSpec, ResolvedRevSpec, RevSpec } from '..'
import { HoverMerged } from '../../backend/features'
import { EMODENOTFOUND } from '../../backend/lsp'
import { eventLogger } from '../../tracking/eventLogger'
import { asError, ErrorLike } from '../../util/errors'
import { propertyIsDefined } from '../../util/types'
import { LineOrPositionOrRange } from '../../util/url'
import {
    convertCodeCellIdempotent,
    convertNode,
    findElementWithOffset,
    getTableDataCell,
    HoveredToken,
    locateTarget,
} from '../blob/tooltips'
import {
    calculateOverlayPosition,
    getRowInCodeElement,
    getRowsInRange,
    getTokenAtPosition,
    LOADING,
    overlayUIHasContent,
    scrollIntoCenterIfNeeded,
} from './helpers'
import { HoverOverlayProps, isJumpURL } from './HoverOverlay'
import { createObservableStateContainer } from './state'

interface HoverifierOptions {
    /**
     * Emit the HoverOverlay element on this after it was rerendered when its content changed and it needs to be repositioned.
     */
    hoverOverlayRerenders: Observable<{ hoverOverlayElement: HTMLElement; scrollElement: HTMLElement }>

    /**
     * Emit on this Observable when the Go-To-Definition button in the HoverOverlay was clicked
     */
    goToDefinitionClicks: Observable<MouseEvent>

    /**
     * Emit on this Observable when the close button in the HoverOverlay was clicked
     */
    closeButtonClicks: Observable<MouseEvent>

    hoverOverlayElements: Observable<HTMLElement | null>

    // TODO make the code not depend on history
    getHistory: () => H.History

    fetchHover: HoverFetcher
    fetchJumpURL: JumpURLFetcher
}

/**
 * A Hoverifier is a function that hoverifies one code view element in the DOM.
 * It will do very dirty things to it. Only call it if you're into that.
 *
 * There can be multiple code views in the DOM, which will only show a single HoverOverlay if the same Hoverifier was used.
 */
export interface Hoverifier {
    /**
     * The current Hover state. You can use this to read the initial state synchronously.
     */
    hoverState: Readonly<HoverState>
    /**
     * This Observable is to notify that the state that is used to render the HoverOverlay needs to be updated.
     */
    hoverStateUpdates: Observable<Readonly<HoverState>>

    /**
     * Hoverifies a code view.
     */
    hoverify(options: HoverifyOptions): Subscription

    unsubscribe(): void
}

export interface HoverifyOptions {
    codeMouseMoves: Observable<React.MouseEvent<HTMLElement>>
    codeMouseOvers: Observable<React.MouseEvent<HTMLElement>>
    codeClicks: Observable<React.MouseEvent<HTMLElement>>

    /**
     * Emit on this Observable to trigger the overlay on a position in this code view.
     * This Observable is intended to be used to trigger a Hover after a URL change with a position.
     */
    positionJumps: Observable<{
        /**
         * The position within the code view to jump to
         */
        position: LineOrPositionOrRange
        /**
         * The code view
         */
        codeElement: HTMLElement
        /**
         * The element to scroll if the position is out of view
         */
        scrollElement: HTMLElement
    }>
    resolveContext: ContextResolver
}

/**
 * Output that contains the information needed to render the HoverOverlay.
 */
export interface HoverState {
    /**
     * The props to pass to `HoverOverlay`, or `undefined` if it should not be rendered.
     */
    hoverOverlayProps?: HoverOverlayProps

    /**
     * The currently selected position, if any.
     * Can be a single line number or a line range.
     * Highlighted with a background color.
     */
    selectedPosition?: LineOrPositionOrRange
}

interface InternalHoverifierState {
    hoverOrError?: typeof LOADING | HoverMerged | null | ErrorLike
    definitionURLOrError?: typeof LOADING | { jumpURL: string } | null | ErrorLike

    hoverOverlayIsFixed: boolean

    /** The desired position of the hover overlay */
    hoverOverlayPosition?: { left: number; top: number }

    /**
     * Whether the user has clicked the go to definition button for the current overlay yet,
     * and whether he pressed Ctrl/Cmd while doing it to open it in a new tab or not.
     */
    clickedGoToDefinition: false | 'same-tab' | 'new-tab'

    /** The currently hovered token */
    hoveredToken?: HoveredToken & HoveredTokenContext

    mouseIsMoving: boolean

    /**
     * The currently selected position, if any.
     * Can be a single line number or a line range.
     * Highlighted with a background color.
     */
    selectedPosition?: LineOrPositionOrRange
}

/**
 * Returns true if the HoverOverlay component should be rendered according to the given state.
 */
const shouldRenderOverlay = (state: InternalHoverifierState): boolean =>
    !(!state.hoverOverlayIsFixed && state.mouseIsMoving) && overlayUIHasContent(state)

/**
 * Maps internal HoverifierState to the publicly exposed HoverState
 */
const internalToExternalState = (internalState: InternalHoverifierState): HoverState => ({
    selectedPosition: internalState.selectedPosition,
    hoverOverlayProps: shouldRenderOverlay(internalState)
        ? {
              overlayPosition: internalState.hoverOverlayPosition,
              hoverOrError: internalState.hoverOrError,
              definitionURLOrError:
                  // always modify the href, but only show error/loader/not found after the button was clicked
                  isJumpURL(internalState.definitionURLOrError) || internalState.clickedGoToDefinition
                      ? internalState.definitionURLOrError
                      : undefined,
              hoveredToken: internalState.hoveredToken,
              showCloseButton: internalState.hoverOverlayIsFixed,
          }
        : undefined,
})

/** The time in ms after which to show a loader if the result has not returned yet */
const LOADER_DELAY = 300

/** The time in ms after the mouse has stopped moving in which to show the tooltip */
const TOOLTIP_DISPLAY_DELAY = 100

type HoverFetcher = (position: HoveredToken & HoveredTokenContext) => Observable<HoverMerged | null>
type JumpURLFetcher = (position: HoveredToken & HoveredTokenContext) => Observable<string | null>
export type ContextResolver = (hoveredToken: HoveredToken) => HoveredTokenContext

export interface HoveredTokenContext extends RepoSpec, RevSpec, ResolvedRevSpec, FileSpec {}

export const createHoverifier = ({
    goToDefinitionClicks,
    closeButtonClicks,
    hoverOverlayRerenders,
    getHistory,
    fetchHover,
    fetchJumpURL,
}: HoverifierOptions): Hoverifier => {
    // Internal state that is not exposed to the caller
    // Shared between all hoverified code views
    const container = createObservableStateContainer<InternalHoverifierState>({
        hoverOverlayIsFixed: false,
        clickedGoToDefinition: false,
        definitionURLOrError: undefined,
        hoveredToken: undefined,
        hoverOrError: undefined,
        hoverOverlayPosition: undefined,
        mouseIsMoving: false,
        selectedPosition: undefined,
    })

    interface MouseEventTrigger {
        event: React.MouseEvent<HTMLElement>
        resolveContext: ContextResolver
    }

    // These Subjects aggregate all events from all hoverified code views
    const allCodeMouseMoves = new Subject<MouseEventTrigger>()
    const allCodeMouseOvers = new Subject<MouseEventTrigger>()
    const allCodeClicks = new Subject<MouseEventTrigger>()
    const allPositionJumps = new Subject<{
        position: LineOrPositionOrRange
        codeElement: HTMLElement
        scrollElement: HTMLElement
        resolveContext: ContextResolver
    }>()

    const subscription = new Subscription()

    /**
     * click events on the code element, ignoring click events caused by the user selecting text.
     * Selecting text should not mess with the hover, hover pinning nor the URL.
     */
    const codeClicksWithoutSelections = allCodeClicks.pipe(filter(() => window.getSelection().toString() === ''))

    // Mouse is moving, don't show the tooltip
    subscription.add(
        merge(
            allCodeMouseMoves.pipe(
                map(({ event }) => event.target),
                // Make sure a move of the mouse from the go-to-definition button
                // back to the same target doesn't cause the tooltip to briefly disappear
                distinctUntilChanged(),
                map(() => true)
            ),

            // When the mouse stopped for TOOLTIP_DISPLAY_DELAY, show tooltip
            // Don't use mouseover for this because it is only fired once per token,
            // not continuously while moving the mouse
            allCodeMouseMoves.pipe(debounceTime(TOOLTIP_DISPLAY_DELAY), map(() => false))
        ).subscribe(mouseIsMoving => {
            container.update({ mouseIsMoving })
        })
    )

    const codeMouseOverTargets = allCodeMouseOvers.pipe(
        map(({ event, ...rest }) => ({
            target: event.target as HTMLElement,
            codeElement: event.currentTarget,
            ...rest,
        })),
        // SIDE EFFECT (but idempotent)
        // If not done for this cell, wrap the tokens in this cell to enable finding the precise positioning.
        // This may be possible in other ways (looking at mouse position and rendering characters), but it works
        tap(({ target, codeElement }) => {
            const td = getTableDataCell(target, codeElement)
            if (td !== undefined) {
                convertCodeCellIdempotent(td)
            }
        }),
        debounceTime(50),
        // Do not consider mouseovers while overlay is pinned
        filter(() => !container.values.hoverOverlayIsFixed),
        share()
    )

    const codeClickTargets = codeClicksWithoutSelections.pipe(
        map(({ event, ...rest }) => ({
            target: event.target as HTMLElement,
            codeElement: event.currentTarget,
            ...rest,
        })),
        share()
    )

    /** Emits DOM elements at new positions found in the URL */
    const jumpTargets: Observable<{
        target: HTMLElement
        codeElement: HTMLElement
        resolveContext: ContextResolver
    }> = allPositionJumps.pipe(
        // Only use line and character for comparison
        map(({ position: { line, character }, ...rest }) => ({ position: { line, character }, ...rest })),
        // Ignore same values
        // It's important to do this before filtering otherwise navigating from
        // a position, to a line-only position, back to the first position would get ignored
        distinctUntilChanged((a, b) => isEqual(a, b)),
        // Ignore undefined or partial positions (e.g. line only)
        filter((jump): jump is typeof jump & { position: Position } => Position.is(jump.position)),
        map(({ position, codeElement, ...rest }) => {
            const row = getRowInCodeElement(codeElement, position.line)
            if (!row) {
                return { target: undefined, codeElement, ...rest }
            }
            const cell = row.cells[1]!
            const target = findElementWithOffset(cell, position.character)
            if (!target) {
                console.warn('Could not find target for position in file', position)
            }
            return { target, codeElement, ...rest }
        }),
        filter(propertyIsDefined('target'))
    )

    // REPOSITIONING
    // On every componentDidUpdate (after the component was rerendered, e.g. from a hover state update) resposition
    // the tooltip
    // It's important to add this subscription first so that withLatestFrom will be guaranteed to have gotten the
    // latest hover target by the time componentDidUpdate is triggered from the setState() in the second chain
    subscription.add(
        // Take every rerender
        hoverOverlayRerenders
            .pipe(
                // with the latest target that came from either a mouseover, click or location change (whatever was the most recent)
                withLatestFrom(merge(codeMouseOverTargets, codeClickTargets, jumpTargets)),
                map(([{ hoverOverlayElement, scrollElement }, { target }]) =>
                    calculateOverlayPosition(scrollElement, target, hoverOverlayElement)
                )
            )
            .subscribe(hoverOverlayPosition => {
                container.update({ hoverOverlayPosition })
            })
    )
    /** Emits new positions at which a tooltip needs to be shown from clicks, mouseovers and URL changes. */
    const positions: Observable<{
        /**
         * The 1-indexed position at which a new tooltip is to be shown,
         * or undefined when a target was hovered/clicked that does not correspond to a position (e.g. after the end of the line)
         */
        position?: HoveredToken & HoveredTokenContext
        /**
         * True if the tooltip should be pinned once the hover came back and is non-empty.
         * This depends on what triggered the new position.
         * We remember it because the pinning is deferred to when we have a result,
         * so we don't pin empty (i.e. invisible) hovers.
         */
        pinIfNonEmpty: boolean
        codeElement: HTMLElement
    }> = merge(
        // Should unpin the tooltip even if hover cames back non-empty
        codeMouseOverTargets.pipe(map(data => ({ ...data, pinIfNonEmpty: false }))),
        // When the location changes and includes a line/column pair, use that target
        // Should pin the tooltip if hover cames back non-empty
        jumpTargets.pipe(map(data => ({ ...data, pinIfNonEmpty: true }))),
        // Should pin the tooltip if hover cames back non-empty
        codeClickTargets.pipe(map(data => ({ ...data, pinIfNonEmpty: true })))
    ).pipe(
        // Find out the position that was hovered over
        map(({ target, codeElement, resolveContext, ...rest }) => {
            const hoveredToken = locateTarget(target, codeElement, false)
            const position = Position.is(hoveredToken)
                ? { ...hoveredToken, ...resolveContext(hoveredToken) }
                : undefined
            return { position, codeElement, ...rest }
        }),
        share()
    )

    /**
     * For every position, emits an Observable with new values for the `hoverOrError` state.
     * This is a higher-order Observable (Observable that emits Observables).
     */
    const hoverObservables = positions.pipe(
        map(({ position, codeElement }) => {
            if (!position) {
                return of({ codeElement, hoverOrError: undefined })
            }
            // Fetch the hover for that position
            const hoverFetch = fetchHover(position).pipe(
                catchError(error => {
                    if (error && error.code === EMODENOTFOUND) {
                        return [undefined]
                    }
                    return [asError(error)]
                }),
                share()
            )
            // 1. Reset the hover content, so no old hover content is displayed at the new position while fetching
            // 2. Show a loader if the hover fetch hasn't returned after 100ms
            // 3. Show the hover once it returned
            return merge([undefined], of(LOADING).pipe(delay(LOADER_DELAY), takeUntil(hoverFetch)), hoverFetch).pipe(
                map(hoverOrError => ({ hoverOrError, codeElement }))
            )
        }),
        share()
    )
    // Highlight the hover range returned by the language server
    subscription.add(
        hoverObservables
            .pipe(switchMap(hoverObservable => hoverObservable))
            .subscribe(({ hoverOrError, codeElement }) => {
                container.update({
                    hoverOrError,
                    // Reset the hover position, it's gonna be repositioned after the hover was rendered
                    hoverOverlayPosition: undefined,
                })
                const currentHighlighted = codeElement!.querySelector('.selection-highlight')
                if (currentHighlighted) {
                    currentHighlighted.classList.remove('selection-highlight')
                }
                if (!HoverMerged.is(hoverOrError) || !hoverOrError.range) {
                    return
                }
                // LSP is 0-indexed, the code in the webapp currently is 1-indexed
                const { line, character } = hoverOrError.range.start
                const token = getTokenAtPosition(codeElement!, { line: line + 1, character: character + 1 })
                if (!token) {
                    return
                }
                token.classList.add('selection-highlight')
            })
    )
    // Telemetry for hovers
    subscription.add(
        zip(positions, hoverObservables)
            .pipe(
                distinctUntilChanged(([positionA], [positionB]) => isEqual(positionA, positionB)),
                switchMap(([position, hoverObservable]) => hoverObservable),
                filter(HoverMerged.is)
            )
            .subscribe(() => {
                eventLogger.log('SymbolHovered')
            })
    )

    /**
     * For every position, emits an Observable that emits new values for the `definitionURLOrError` state.
     * This is a higher-order Observable (Observable that emits Observables).
     */
    const definitionObservables = positions.pipe(
        // Fetch the definition location for that position
        map(({ position }) => {
            if (!position) {
                return of(undefined)
            }
            return concat(
                [LOADING],
                fetchJumpURL(position).pipe(
                    map(url => (url !== null ? { jumpURL: url } : null)),
                    catchError(error => [asError(error)])
                )
            )
        })
    )

    // GO TO DEFINITION FETCH
    // On every new hover position, (pre)fetch definition and update the state
    subscription.add(
        definitionObservables
            // flatten inner Observables
            .pipe(switchMap(definitionObservable => definitionObservable), share())
            .subscribe(definitionURLOrError => {
                container.update({ definitionURLOrError })
                // If the j2d button was already clicked and we now have the result, jump to it
                // TODO move this logic into HoverOverlay
                if (container.values.clickedGoToDefinition && isJumpURL(definitionURLOrError)) {
                    switch (container.values.clickedGoToDefinition) {
                        case 'same-tab':
                            getHistory().push(definitionURLOrError.jumpURL)
                            break
                        case 'new-tab':
                            window.open(definitionURLOrError.jumpURL, '_blank')
                            break
                    }
                }
            })
    )

    // DEFERRED HOVER OVERLAY PINNING
    // If the new position came from a click or the URL,
    // if either the hover or the definition turn out non-empty, pin the tooltip.
    // If they both turn out empty, unpin it so we don't end up with an invisible tooltip.
    //
    // zip together a position and the hover and definition fetches it triggered
    subscription.add(
        zip(positions, hoverObservables, definitionObservables)
            .pipe(
                switchMap(([{ pinIfNonEmpty }, hoverObservable, definitionObservable]) => {
                    // If the position was triggered by a mouseover, never pin
                    if (!pinIfNonEmpty) {
                        return [false]
                    }
                    // combine the latest values for them, so we have access to both values
                    // and can reevaluate our pinning decision whenever one of the two updates,
                    // independent of the order in which they emit
                    return combineLatest(hoverObservable, definitionObservable).pipe(
                        map(([{ hoverOrError }, definitionURLOrError]) =>
                            overlayUIHasContent({ hoverOrError, definitionURLOrError })
                        )
                    )
                }),
                share()
            )
            .subscribe(hoverOverlayIsFixed => {
                container.update({ hoverOverlayIsFixed })
            })
    )

    // On every click on a go to definition button, reveal loader/error/not found UI
    subscription.add(
        goToDefinitionClicks.subscribe(event => {
            // Telemetry
            eventLogger.log('GoToDefClicked')

            // If we don't have a result yet that would be jumped to by the native <a> tag...
            if (!isJumpURL(container.values.definitionURLOrError)) {
                // Prevent default link behaviour (jump will be done programmatically once finished)
                event.preventDefault()
            }
        })
    )

    // When the close button is clicked, unpin, hide and reset the hover
    subscription.add(
        merge(
            closeButtonClicks,
            fromEvent<KeyboardEvent>(window, 'keydown').pipe(filter(event => event.key === Key.Escape))
        ).subscribe(event => {
            event.preventDefault()
            container.update({
                hoverOverlayIsFixed: false,
                hoverOverlayPosition: undefined,
                hoverOrError: undefined,
                hoveredToken: undefined,
                definitionURLOrError: undefined,
                clickedGoToDefinition: false,
            })
        })
    )

    // LOCATION CHANGES
    subscription.add(
        // It's important to not filter partial positions out here
        // so that selectedPosition still gets updated for partial positions
        // (e.g. to show blame info)
        allPositionJumps.subscribe(({ position, scrollElement, codeElement }) => {
            container.update({
                // Remember active position in state for blame and range expansion
                selectedPosition: position,
            })
            const rows = getRowsInRange(codeElement, position)
            for (const { element } of rows) {
                convertNode(element.cells[1]!)
            }
            // Scroll into view
            if (rows.length > 0) {
                scrollIntoCenterIfNeeded(scrollElement, codeElement, rows[0].element)
            }
        })
    )
    subscription.add(
        positions.subscribe(({ position }) => {
            container.update({
                hoveredToken: position,
                // On every new target (from mouseover or click) hide the j2d loader/error/not found UI again
                clickedGoToDefinition: false,
            })
        })
    )
    subscription.add(
        goToDefinitionClicks.subscribe(event => {
            container.update({ clickedGoToDefinition: event.ctrlKey || event.metaKey ? 'new-tab' : 'same-tab' })
        })
    )

    return {
        get hoverState(): Readonly<HoverState> {
            return internalToExternalState(container.values)
        },
        hoverStateUpdates: container.updates.pipe(
            map(internalToExternalState),
            distinctUntilChanged((a, b) => isEqual(a, b))
        ),
        hoverify({
            codeMouseMoves,
            codeMouseOvers,
            codeClicks,
            positionJumps,
            resolveContext,
        }: HoverifyOptions): Subscription {
            const subscription = new Subscription()
            const eventWithContextResolver = map((event: React.MouseEvent<HTMLElement>) => ({ event, resolveContext }))
            // Broadcast all events from this code view
            subscription.add(codeMouseMoves.pipe(eventWithContextResolver).subscribe(allCodeMouseMoves))
            subscription.add(codeMouseOvers.pipe(eventWithContextResolver).subscribe(allCodeMouseOvers))
            subscription.add(codeClicks.pipe(eventWithContextResolver).subscribe(allCodeClicks))
            subscription.add(positionJumps.pipe(map(jump => ({ ...jump, resolveContext }))).subscribe(allPositionJumps))
            return subscription
        },
        unsubscribe(): void {
            subscription.unsubscribe()
        },
    }
}
