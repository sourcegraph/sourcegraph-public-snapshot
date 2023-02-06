/* eslint-disable jsdoc/check-indentation */
/**
 * This module uses various view plugins, facets and fields to implement
 * hovercard functionality.
 *
 * We currently have two "instances" of hovercards: The hovercard shown when
 * hovering over a token and a pinned hovercard, as derived from the URL.
 * The {@link hovercardState} field holds the information about the hovered and
 * pinned hovercards.
 * Its value is set/updated by two managers: {@link hoverManager} is responsible
 * for creating {@link Hovercard} instances when the mouse hovers over a token.
 * {@link pinManager} is responsible for creating {@link Hovercard} instances
 * for the position in the URL.
 *
 * {@link hovercardState} provides the input for the {@link showHovercard}
 * facet. The facet uses this information to populate multiple other
 * extension/facets:
 * - {@link showTooltip} for rendering the actual hovercard
 * - {@link EditorView.decorations} to highlight the corresponding token in the
 *   document
 * - {@link EditorView.domEventHandlers} for handling
 *   "click-to-go-to-definition"
 */
import { Facet, RangeSet, StateEffect, StateField } from '@codemirror/state'
import {
    Decoration,
    EditorView,
    getTooltip,
    PluginValue,
    repositionTooltips,
    showTooltip,
    Tooltip,
    TooltipView,
    ViewPlugin,
    ViewUpdate,
} from '@codemirror/view'
import classNames from 'classnames'
import { createRoot, Root } from 'react-dom/client'
import {
    BehaviorSubject,
    combineLatest,
    fromEvent,
    Observable,
    of,
    OperatorFunction,
    pipe,
    Subject,
    Subscription,
} from 'rxjs'
import {
    startWith,
    filter,
    scan,
    switchMap,
    shareReplay,
    map,
    distinctUntilChanged,
    debounceTime,
} from 'rxjs/operators'

import {
    addLineRangeQueryParameter,
    isErrorLike,
    LineOrPositionOrRange,
    toPositionOrRangeQueryParameter,
} from '@sourcegraph/common'
import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { UIPositionSpec, UIRangeSpec } from '@sourcegraph/shared/src/util/url'

import {
    getClickToGoToDefinition,
    getGoToURL,
    WebHoverOverlay,
    WebHoverOverlayProps,
} from '../../../components/WebHoverOverlay'
import { BlobProps, updateBrowserHistoryIfChanged } from '../Blob'

import { CodeMirrorContainer } from './react-interop'
import {
    preciseWordAtCoords,
    offsetToUIPosition,
    positionToOffset,
    preciseOffsetAtCoords,
    uiPositionToOffset,
    zeroToOneBasedPosition,
} from './utils'

import { blobPropsFacet } from './index'

import webOverlayStyles from '../../../components/WebHoverOverlay/WebHoverOverlay.module.scss'

type UIRange = UIRangeSpec['range']
type UIPosition = UIPositionSpec['position']

/**
 * Hover information received from a hover source.
 */
export type HoverData = Pick<WebHoverOverlayProps, 'hoverOrError' | 'actionsOrError'>

/**
 * A {@link Hovercard} represent a currently visible hovercard.
 */
interface Hovercard {
    // CodeMirror document offsets of the token under the cursor as determined
    // by CodeMirror
    from: number
    to: number

    // Line/column position
    range: UIRange

    // CodeMirror Tooltip associated with this hovercard.
    tooltip: Tooltip

    // Sometimes the range returned by the hover provider differs from the
    // "word" range determined by CodeMirror. Highlighting and
    // click-to-go-to-definition should use this range if available.
    providerOffset?: { from: number; to: number }

    // Used to provide "click to go to definition". The presence of this value
    // also indicates that the range should be decorated with a pointer cursor.
    onClick?: () => void

    // Whether or not this hovercard is considered "pinned". We only show a
    // close button for pinned hovercards
    pinned?: boolean
}

const HOVER_INFO_CACHE_SIZE = 30
export const HOVER_DEBOUNCE_TIME = 25 // ms

/**
 * Some style overrides to replicate the existing hovercard style.
 */
const hovercardTheme = EditorView.theme({
    [`.${webOverlayStyles.webHoverOverlay}`]: {
        // This is normally "position: 'absolute'". CodeMirror does the
        // positioning. Without this CodeMirror thinks the hover content is
        // empty.
        position: 'initial !important',
    },
    '.cm-tooltip': {
        // Reset CodeMirror's default style
        border: 'initial',
        backgroundColor: 'initial',
        // Needed to ensure that the hovercard is not covered by the reference
        // panel header.
        zIndex: 1024,
    },
    '.hover-gtd': {
        cursor: 'pointer',
    },
})

/**
 * Decorations for highlighting the range of the {@link Hovercard}.
 */
export const selectionHighlightDecoration = Decoration.mark({ class: 'selection-highlight' })
const selectionGoToDefinitionDecoration = Decoration.mark({ class: 'selection-highlight hover-gtd' })

/**
 * Facet to which an extension can add a value to show a {@link Hovercard}. This
 * facet enables a couple of other facets with this data:
 * - {@link showTooltip}, for rendering the actual hocercards (using
 *   {@link Hovercard.tooltip}).
 * - {@link EditorView.decorations}, for highlighting the corresponding token with a different
 *   color and optionally showing a pointer cursor for click-to-go-to-definition
 * - {@link EditorView.domEventHandlers} for executing click-to-go-to-definition
 */
const showHovercard = Facet.define<Hovercard>({
    enables: facet => [
        hovercardTheme,
        // Provide CodeMirror tooltips from hovercard objects
        showTooltip.computeN([facet], state => state.facet(facet).map(range => range.tooltip)),

        // Highlight token under cursor
        EditorView.decorations.compute([facet], state =>
            RangeSet.of(
                Array.from(state.facet(facet), range => {
                    const { from, to } = range.providerOffset ?? range
                    return range.onClick
                        ? selectionGoToDefinitionDecoration.range(from, to)
                        : selectionHighlightDecoration.range(from, to)
                }),
                true
            )
        ),

        // Handles click-to-go-to-definition if enabled (as determined by Hovercard)
        EditorView.domEventHandlers({
            click(event: MouseEvent, view: EditorView) {
                const ranges = view.state.facet(facet)
                if (ranges.length === 0) {
                    return false
                }

                // Ignore event when the click event is the result of the
                // user selecting text
                if (view.state.selection.main.from !== view.state.selection.main.to) {
                    return false
                }

                const offset = preciseOffsetAtCoords(view, event)
                if (offset === null) {
                    return false
                }

                for (const range of ranges.values()) {
                    if (isOffsetInHoverRange(offset, range)) {
                        range.onClick?.()
                        return true
                    }
                }
                return false
            },
        }),
    ],
})

/**
 * Effect for setting the hovercard for the currently hovered token.
 */
const setHoverHovercard = StateEffect.define<Hovercard | null>()
/**
 * Effect for setting the hovercard for the pinned token/range.
 */
const setPinnedHovercard = StateEffect.define<Hovercard | null>()
/**
 * Effect for pinning the current hover hovercard.
 */
const pinHovercard = StateEffect.define<void>()

/**
 * Field for storing hover and pinned hovercard information. Gets updated by
 * {@link pinManager} and {@link hoverManager} and provides input for
 * {@link hovercardSource}.
 */
const hovercardState = StateField.define<{ pinned: Hovercard | null; hover: Hovercard | null }>({
    create() {
        return {
            pinned: null,
            hover: null,
        }
    },

    compare(previous, next) {
        return previous.hover === next.hover && previous.pinned === next.pinned
    },

    update(state, transaction) {
        for (const effect of transaction.effects) {
            if (effect.is(setHoverHovercard)) {
                state = {
                    ...state,
                    hover: effect.value,
                }
            }
            if (effect.is(setPinnedHovercard)) {
                state = {
                    ...state,
                    pinned: effect.value,
                }
            }
            if (effect.is(pinHovercard) && state.hover) {
                return {
                    pinned: {
                        ...state.hover,
                        pinned: true,
                    },
                    hover: null,
                }
            }
        }
        return state
    },
    provide(field) {
        return showHovercard.computeN([field], state => {
            const hovercards: Hovercard[] = []
            const hovercardState = state.field(field)
            if (hovercardState.hover) {
                hovercards.push(hovercardState.hover)
            }
            if (hovercardState.pinned) {
                hovercards.push(hovercardState.pinned)
            }
            return hovercards
        })
    },
})

/**
 * This field is used by the blob component to sync the position from the URL to
 * the editor.
 */
export const [pin, updatePin] = createUpdateableField<LineOrPositionOrRange | null>(null)

/**
 * This view plugin listens to changes to the {@link pin} field, fetches hover
 * information when necessary and updates {@link hovercardState}.
 */
const pinManager = ViewPlugin.fromClass(
    class implements PluginValue {
        private nextPin: Subject<LineOrPositionOrRange | null>
        private subscription: Subscription

        constructor(view: EditorView) {
            this.nextPin = new BehaviorSubject(view.state.field(pin))
            this.subscription = this.nextPin
                .pipe(
                    map(pin => {
                        if (!pin || !pin.line || !pin.character) {
                            return null
                        }

                        const from = uiPositionToOffset(view.state.doc, { line: pin.line, character: pin.character })

                        if (from === null) {
                            return null
                        }

                        let to: number | null = null

                        if (pin.endLine && pin.endCharacter) {
                            to = uiPositionToOffset(view.state.doc, { line: pin.endLine, character: pin.endCharacter })
                        } else {
                            // To determine the end position we have to find the word at the
                            // start position
                            const word = view.state.wordAt(from)
                            if (!word) {
                                return null
                            }
                            to = word.to
                        }

                        if (to === null) {
                            return null
                        }

                        return { from, to }
                    }),
                    filter(pin => {
                        if (!pin) {
                            return true
                        }
                        // If we already have hovercard in the pinned state that contains the new
                        // pin range, do nothing. That means the hovercard was just transfered
                        // to pinned state.
                        const currentlyPinned = view.state.field(hovercardState).pinned
                        return !currentlyPinned || !isOffsetInHoverRange(pin.from, currentlyPinned)
                    }),
                    // Create hovercard object for token
                    tokenRangeToHovercard(view, true)
                )
                .subscribe(hovercardRange => {
                    // Scheduling the update for the next loop is necessary at the
                    // moment because we are triggering this effect in response to an
                    // editor update (pin field change) and you cannot synchronously
                    // trigger an update from an update.
                    window.requestAnimationFrame(() =>
                        view.dispatch({ effects: setPinnedHovercard.of(hovercardRange) })
                    )
                })
        }

        public update(update: ViewUpdate): void {
            if (update.startState.field(pin) !== update.state.field(pin)) {
                this.nextPin.next(update.state.field(pin))
            }
        }

        public destroy(): void {
            this.subscription.unsubscribe()
        }
    }
)

/**
 * The MouseEvent uses numbers to indicate which button was pressed.
 * See https://developer.mozilla.org/en-US/docs/Web/API/MouseEvent/buttons#value
 */
export const MOUSE_NO_BUTTON = 0

/**
 * Listens to mousemove events, determines whether the position under the mouse
 * cursor is eligible (whether a "word" is under the mouse cursor), fetches
 * hover information as necessary and updates {@link hovercardState}.
 */
const hoverManager = ViewPlugin.fromClass(
    class HoverManager implements PluginValue {
        private nextOffset = new Subject<number | null>()
        private subscription: Subscription

        constructor(private readonly view: EditorView) {
            this.subscription = fromEvent<MouseEvent>(this.view.dom, 'mousemove')
                .pipe(
                    // Debounce events so that users can move over tokens without triggering hovercards immediately
                    debounceTime(HOVER_DEBOUNCE_TIME),

                    // Ignore some events
                    filter(event => {
                        // Ignore events when hovering over an existing hovercard.
                        // This causes existing hovercards to stay open.
                        if (
                            (event.target as HTMLElement | null)?.closest(
                                '.cm-code-intel-hovercard:not(.cm-code-intel-hovercard-pinned)'
                            )
                        ) {
                            return false
                        }

                        // We have to forward any move events that also have a
                        // button pressed. User is probably selecting text and
                        // hovercards should be hidden.
                        if (event.buttons !== MOUSE_NO_BUTTON) {
                            return true
                        }

                        // Ignore events inside the current hover range. Without this
                        // hovercards flicker when the active range is wider than the
                        // word-under-cursor range. For example, hovering over
                        //
                        // import ( "io/fs" )
                        //
                        // will detect `io` and `fs` as separate words (and would
                        // therefore trigger two individual word lookups), but the
                        // hover information returned by the server is for the whole
                        // `io/fs` range.
                        const offset = preciseOffsetAtCoords(view, event)
                        if (offset === null) {
                            return true
                        }
                        const ranges = view.state.facet(showHovercard)
                        return ranges.every(range => !isOffsetInHoverRange(offset, range))
                    }),

                    // To make it easier to reach the hovercard with the mouse, we determine
                    // in which direction the mouse moves and only hide the hovercard when
                    // the mouse moves away from it.
                    scan(
                        (
                            previous: {
                                x: number
                                y: number
                                target: EventTarget | null
                                buttons: number
                                direction?: 'towards' | 'away' | undefined
                            },
                            next
                        ) => {
                            const currentTooltip = view.state.field(hovercardState).hover?.tooltip
                            if (!currentTooltip) {
                                return next
                            }

                            const tooltipView = getTooltip(view, currentTooltip)
                            if (!tooltipView) {
                                return next
                            }

                            const direction = computeMouseDirection(
                                tooltipView.dom.getBoundingClientRect(),
                                previous,
                                next
                            )
                            return { x: next.x, y: next.y, buttons: next.buttons, target: next.target, direction }
                        }
                    ),

                    // Determine the precise location of the word under the cursor.
                    switchMap(position => {
                        // Hide any hovercard when
                        // - the mouse is over an element that is not part of
                        //   the content. This seems necessary to make hovercards
                        //   not appear and hide open hovercards when the mouse
                        //   moves over the editor's search panel.
                        // - the user starts to select text
                        if (
                            position.buttons !== MOUSE_NO_BUTTON ||
                            !position.target ||
                            !this.view.contentDOM.contains(position.target as Node)
                        ) {
                            return of('HIDE' as const)
                        }
                        return of(preciseWordAtCoords(this.view, position)).pipe(
                            tokenRangeToHovercard(this.view),
                            map(hovercard => ({ position, hovercard }))
                        )
                    })
                )
                .subscribe(next => {
                    if (next === 'HIDE') {
                        this.view.dispatch({
                            effects: setHoverHovercard.of(null),
                        })
                        return
                    }

                    // We only change the hovercard when
                    // a) There is a new hovercord at the position (hovercard !== null)
                    // b) there is no hovercard and the mouse is moving away from the hovercard
                    if (next.hovercard || next.position.direction !== 'towards') {
                        this.view.dispatch({
                            effects: setHoverHovercard.of(next.hovercard),
                        })
                    }
                })

            this.view.dom.addEventListener('mouseleave', this.mouseleave)
        }

        private mouseleave = (): void => {
            this.nextOffset.next(null)
        }

        public destroy(): void {
            this.view.dom.removeEventListener('mouseleave', this.mouseleave)
            this.subscription.unsubscribe()
        }
    }
)

export function computeMouseDirection(
    rect: DOMRect,
    position1: { x: number; y: number },
    position2: { x: number; y: number }
): 'towards' | 'away' {
    if (
        // Moves away from the top
        (position2.y < position1.y && position2.y < rect.top) ||
        // Moves away from the bottom
        (position2.y > position1.y && position2.y > rect.bottom) ||
        // Moves away from the left
        (position2.x < position1.x && position2.x < rect.left) ||
        // Moves away from the right
        (position2.x > position1.x && position2.x > rect.right)
    ) {
        return 'away'
    }

    return 'towards'
}

// WebHoverOverlay expects to be passed the overlay position. Since CodeMirror
// positions the element we always use the same value.
const dummyOverlayPosition = { left: 0, bottom: 0 }

/**
 * This class is responsible for rendering a WebHoverOverlay component as a
 * CodeMirror tooltip. When constructed the instance subscribes to the hovercard
 * data source and the component props, and updates the component as it receives
 * changes.
 */
export class HovercardView implements TooltipView {
    public dom: HTMLElement
    private root: Root | null = null
    private nextContainer = new Subject<HTMLElement>()
    private nextProps = new Subject<BlobProps>()
    private props: BlobProps | null = null
    public overlap = true
    private nextPinned = new Subject<boolean>()
    private subscription: Subscription

    constructor(
        private readonly view: EditorView,
        private readonly tokenRange: UIRange,
        pinned: boolean,
        hovercardData: Observable<HoverData>
    ) {
        this.dom = document.createElement('div')

        this.subscription = combineLatest([
            this.nextContainer,
            hovercardData,
            this.nextProps.pipe(startWith(view.state.facet(blobPropsFacet))),
            this.nextPinned.pipe(startWith(pinned)),
        ]).subscribe(([container, hovercardData, props, pinned]) => {
            if (!this.root) {
                this.root = createRoot(container)
            }
            this.render(this.root, hovercardData, props, pinned)
        })
    }

    public mount(): void {
        this.nextContainer.next(this.dom)
    }

    public update(update: ViewUpdate): void {
        // Update component when props change
        const props = update.state.facet(blobPropsFacet)
        if (this.props !== props) {
            this.props = props
            this.nextProps.next(props)
        }
    }

    public destroy(): void {
        this.subscription.unsubscribe()
        this.root?.unmount()
    }

    private render(root: Root, { hoverOrError, actionsOrError }: HoverData, props: BlobProps, pinned: boolean): void {
        const hoverContext = {
            commitID: props.blobInfo.commitID,
            filePath: props.blobInfo.filePath,
            repoName: props.blobInfo.repoName,
            revision: props.blobInfo.revision,
        }

        let hoveredToken: Exclude<WebHoverOverlayProps['hoveredToken'], undefined> = {
            ...hoverContext,
            ...this.tokenRange.start,
        }

        if (hoverOrError && hoverOrError !== 'loading' && !isErrorLike(hoverOrError) && hoverOrError.range) {
            hoveredToken = {
                ...hoveredToken,
                ...zeroToOneBasedPosition(hoverOrError.range.start),
            }
        }

        root.render(
            <CodeMirrorContainer onRender={() => repositionTooltips(this.view)} history={props.history}>
                <div
                    className={classNames({
                        'cm-code-intel-hovercard': true,
                        'cm-code-intel-hovercard-pinned': pinned,
                    })}
                >
                    <WebHoverOverlay
                        // Blob props
                        location={props.location}
                        onHoverShown={props.onHoverShown}
                        isLightTheme={props.isLightTheme}
                        platformContext={props.platformContext}
                        settingsCascade={props.settingsCascade}
                        telemetryService={props.telemetryService}
                        extensionsController={props.extensionsController}
                        // Hover props
                        actionsOrError={actionsOrError}
                        hoverOrError={hoverOrError}
                        // CodeMirror handles the positioning but a
                        // non-nullable value must be passed for the
                        // hovercard to render
                        overlayPosition={dummyOverlayPosition}
                        hoveredToken={hoveredToken}
                        onAlertDismissed={() => repositionTooltips(this.view)}
                        pinOptions={{
                            showCloseButton: pinned,
                            onCloseButtonClick: () => {
                                const parameters = new URLSearchParams(props.location.search)
                                parameters.delete('popover')

                                updateBrowserHistoryIfChanged(props.navigate, props.location, parameters)
                                this.nextPinned.next(false)
                            },
                            onCopyLinkButtonClick: async () => {
                                if (!pinned) {
                                    // This needs to happen before updating the URL so that we avoid re-creating the hovercard
                                    this.view.dispatch({
                                        effects: pinHovercard.of(),
                                    })
                                }
                                const { line, character } = hoveredToken
                                const position = { line, character }

                                const search = new URLSearchParams(location.search)
                                search.set('popover', 'pinned')

                                updateBrowserHistoryIfChanged(
                                    props.navigate,
                                    props.location,
                                    // It may seem strange to set start and end to the same value, but that what's the old blob view is doing as well
                                    addLineRangeQueryParameter(
                                        search,
                                        toPositionOrRangeQueryParameter({
                                            position,
                                            range: { start: position, end: position },
                                        })
                                    )
                                )
                                await navigator.clipboard.writeText(window.location.href)

                                this.nextPinned.next(true)
                            },
                        }}
                        hoverOverlayContainerClassName="position-relative"
                    />
                </div>
            </CodeMirrorContainer>
        )
    }
}

/**
 * A small ring cache for hover data. Clears itself when the document changes.
 */
const hoverdataCache = ViewPlugin.fromClass(
    class implements PluginValue {
        private values: [string, Observable<HoverData>][] = []
        private nextIndex = 0

        public add(key: { from: number; to: number }, value: Observable<HoverData>): void {
            this.values[this.nextIndex] = [this.keyToString(key), value]
            this.nextIndex = this.nextIndex + (1 % HOVER_INFO_CACHE_SIZE)
        }

        public get(key: { from: number; to: number }): Observable<HoverData> | undefined {
            const strKey = this.keyToString(key)
            return this.values.find(([key]) => strKey === key)?.[1]
        }

        public update(update: ViewUpdate): void {
            if (update.docChanged) {
                this.values = []
                this.nextIndex = 0
            }
        }

        private keyToString({ from, to }: { from: number; to: number }): string {
            return `${from}:${to}`
        }
    }
)

/**
 * Helper operator for requesting hover information and creating a {@link Hovercard}
 * object.
 */
function tokenRangeToHovercard(
    view: EditorView,
    pinned: boolean = false
): OperatorFunction<{ from: number; to: number } | null, Hovercard | null> {
    return pipe(
        // Request hover information for the provided token.
        switchMap(token => {
            if (token) {
                const lineCharacterRange = offsetToUIPosition(view.state.doc, token.from, token.to)

                const hoverDataSource = view.state.facet(hovercardSource)
                if (!hoverDataSource) {
                    return of(null)
                }
                // We request hover information and create a hot observable. We use the observable to determine
                // whether/when to show a tooltip. The observable is also passed to the tooltip so it can update itself
                // as new information arrives.
                let hoverData: Observable<HoverData>
                const cache = view.plugin(hoverdataCache)
                const cachedHoverData = cache?.get(token)
                if (cachedHoverData) {
                    hoverData = cachedHoverData
                } else {
                    hoverData = hoverDataSource(view, lineCharacterRange.start).pipe(shareReplay(1))
                    if (cache) {
                        cache.add(token, hoverData)
                    }
                }

                return hoverData.pipe(
                    // `hoverInformation` emits multiple times. We need to
                    // update the hovercard object as new data comes in. We use
                    // `scan` to keep state.
                    scan((hovercard: Hovercard | null, { hoverOrError, actionsOrError }) => {
                        // Only render if we either have something for hover or actions. Adapted
                        // from shouldRenderOverlay in codeintellify/src/hoverifier.ts
                        if (
                            !(
                                (hoverOrError && hoverOrError !== 'loading') ||
                                (actionsOrError &&
                                    actionsOrError !== 'loading' &&
                                    (isErrorLike(actionsOrError) || actionsOrError.length > 0))
                            )
                        ) {
                            return null
                        }

                        // As new information arrives we update click-to-go-to-definition availability
                        // and reposition the hovercard if necessary

                        let providerOffset = hovercard?.providerOffset
                        let onClick = hovercard?.onClick

                        if (!onClick) {
                            const props = view.state.facet(blobPropsFacet)
                            // Adaption of the "click to go to definition" code inside
                            // WebHoverOverlay
                            if (getClickToGoToDefinition(props.settingsCascade)) {
                                const urlAndType = getGoToURL(actionsOrError, props.location)
                                if (urlAndType) {
                                    const { url, actionType } = urlAndType
                                    onClick = () => {
                                        props.telemetryService.log(`${actionType}HoverOverlay.click`)
                                        if (props.nav) {
                                            props.nav(url)
                                        } else {
                                            props.history.push(url)
                                        }
                                    }
                                }
                            }
                        }

                        if (!providerOffset) {
                            if (
                                hoverOrError &&
                                hoverOrError !== 'loading' &&
                                !isErrorLike(hoverOrError) &&
                                hoverOrError.range
                            ) {
                                const from = positionToOffset(view.state.doc, hoverOrError.range.start)
                                const to = positionToOffset(view.state.doc, hoverOrError.range.end)
                                if (from !== null && to !== null) {
                                    providerOffset = { from, to }
                                }
                            }
                        }

                        if (!hovercard) {
                            hovercard = {
                                ...token,
                                range: lineCharacterRange,
                                tooltip: {
                                    pos: providerOffset?.from ?? token.from,
                                    end: providerOffset?.to ?? token.to,
                                    above: true,
                                    create: view => new HovercardView(view, lineCharacterRange, pinned, hoverData),
                                },
                                onClick,
                                providerOffset,
                                pinned,
                            }
                        } else if (hovercard.onClick !== onClick) {
                            hovercard = { ...hovercard, onClick }
                        } else if (providerOffset && hovercard.providerOffset !== providerOffset) {
                            hovercard = {
                                ...hovercard,
                                providerOffset,
                                tooltip: {
                                    // By updating the position of the tooltip only we are triggering CodeMirror to
                                    // reposition the hovercard with recreating it.
                                    // This causes the hovercard to be aligned with the token range returned by server
                                    // and not with the range determined by CodeMirror.
                                    // Example: import( "io/fs" )
                                    // CodeMirror will determine `fs` as the token (word) but the server returns a
                                    // range that covers `io/fs`.
                                    ...hovercard.tooltip,
                                    pos: providerOffset.from,
                                    end: providerOffset.to,
                                },
                            }
                        }
                        return hovercard
                    }, null),
                    distinctUntilChanged()
                )
            }
            return of(null)
        })
    )
}

/**
 * A HovercardSource is a function that is passed a position and returns an
 * observable that provides hover information.
 */
export type HovercardSource = (view: EditorView, position: UIPosition) => Observable<HoverData>

/**
 * Facet with which an extension can provide a hovercard source. For simplicity
 * only one source can be provided, others are ignored (in practice there is
 * only one source at the moment anyway).
 */
export const hovercardSource = Facet.define<HovercardSource, HovercardSource | undefined>({
    combine: sources => sources[0],
    enables: [
        hoverdataCache,
        hovercardState,
        // By enabling the managers when the hovercardSource is present we
        // ensure that the source exist when they try to access it.
        hoverManager,
        pinManager,
    ],
})

function isOffsetInHoverRange(offset: number, range: Hovercard): boolean {
    const offsets = range.providerOffset ?? range
    return offsets.from <= offset && offset <= offsets.to
}
