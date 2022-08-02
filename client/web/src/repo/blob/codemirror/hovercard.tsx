import { Extension, Facet, RangeSet, StateEffect, StateEffectType, StateField } from '@codemirror/state'
import {
    Decoration,
    EditorView,
    PluginValue,
    repositionTooltips,
    showTooltip,
    Tooltip,
    TooltipView,
    ViewPlugin,
    ViewUpdate,
} from '@codemirror/view'
import { createRoot, Root } from 'react-dom/client'

import { addLineRangeQueryParameter, isErrorLike, toPositionOrRangeQueryParameter } from '@sourcegraph/common'

import { WebHoverOverlay, WebHoverOverlayProps } from '../../../components/WebHoverOverlay'
import { blobPropsFacet } from '.'
import { combineLatest, Observable, Subject, Subscription } from 'rxjs'
import { startWith, scan, distinctUntilChanged } from 'rxjs/operators'
import { BlobProps, updateBrowserHistoryIfChanged } from '../Blob'
import { Container } from './react-interop'
import webOverlayStyles from '../../../components/WebHoverOverlay/WebHoverOverlay.module.scss'
import { UIPositionSpec, UIRangeSpec } from '@sourcegraph/shared/src/util/url'
import { offsetToPosition } from './utils'
import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'

type HovercardData = Pick<WebHoverOverlayProps, 'hoverOrError' | 'actionsOrError'>
type HovercardRange = {
    // Line/column position
    range: UIRangeSpec['range']

    // CodeMirror document offsets
    from: number
    to: number
}

/**
 * A HovercardSource is a function that is passed a position and returns an
 * observable that provides hover information.
 */
export type HovercardSource = (view: EditorView, position: UIPositionSpec['position']) => Observable<HovercardData>

/**
 * Some style overrides to replicate the existing hovercard style.
 */
const hoverCardTheme = EditorView.theme({
    '.cm-code-intel-hovercard': {
        // Without this all text in the hovercard is monospace
        fontFamily: 'sans-serif',
    },
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
    },
})

/**
 * This field (and effects) are necessary for highlighting the hovered token
 * properly. Unfortunately we cannot create decorations directly from
 * hovercardRanges because we don't know whether any of these ranges have
 * information associated with them. So it's on the Hovercard instances to
 * inform this field about the ranges that have information.
 * This field also monitors hovercardRanges to remove any stray highlights.
 */
const addRange = StateEffect.define<HovercardRange>()
const removeRange = StateEffect.define<HovercardRange>()
const selectionHighlightDecoration = Decoration.mark({ class: 'selection-highlight' })

const validTokenRanges = StateField.define<HovercardRange[]>({
    create() {
        return []
    },

    update(value, transaction) {
        const availableRanges = transaction.state.facet(hovercardRanges)

        const newValues = [...value]
        let changed = false

        // Remove any values not in the current set of ranges. availableRanges and
        // value will be small, so processing them this way should be fine.
        for (const range of value) {
            if (!availableRanges.includes(range)) {
                newValues.splice(newValues.indexOf(range), 1)
                changed = true
            }
        }

        for (const effect of transaction.effects) {
            if (effect.is(addRange)) {
                if (availableRanges.includes(effect.value) && !newValues.includes(effect.value)) {
                    newValues.push(effect.value)
                    changed = true
                }
            }
            if (effect.is(removeRange)) {
                const index = newValues.indexOf(effect.value)
                if (index > -1) {
                    newValues.splice(index, 1)
                    changed = true
                }
            }
        }

        return changed ? newValues : value
    },

    provide(field) {
        return EditorView.decorations.from(field, ranges =>
            RangeSet.of(
                ranges.map(range => selectionHighlightDecoration.range(range.from, range.to)),
                true
            )
        )
    },
})

/**
 * This class is used to prevent flickering when a hovercard is pinned or when a
 * pinned and a hovered hovercard are rendered.
 * This class keeps track of for which ranges a tooltip exists and will
 * add/remove tooltips as necessary. Flickering is prevented by reusing existing
 * tooltip instances for existing ranges.
 */
class HovercardManager implements PluginValue {
    private tooltips: Map<string, Tooltip> = new Map()
    private hovercardRanges: readonly HovercardRange[] = []

    constructor(readonly view: EditorView, readonly setTooltips: StateEffectType<Tooltip[]>) {}

    public update(update: ViewUpdate) {
        const ranges = update.state.facet(hovercardRanges)
        if (this.hovercardRanges !== ranges) {
            this.hovercardRanges = ranges
            this.updateTooltips()
        }
    }

    private updateTooltips() {
        // Remove removed tooltips
        for (const [key, tooltip] of this.tooltips) {
            if (!this.hovercardRanges.some(range => range.from === tooltip.pos && range.to === tooltip.end)) {
                this.tooltips.delete(key)
            }
        }

        // Add new ranges
        for (const range of this.hovercardRanges) {
            const key = this.toKey(range)
            if (!this.tooltips.has(key)) {
                this.tooltips.set(key, {
                    pos: range.from,
                    end: range.to,
                    above: true,
                    create: view => Hovercard.create(view, range),
                })
            }
        }

        // We cannot directly dispatch a transaction within an update cycle
        window.requestAnimationFrame(() =>
            this.view.dispatch({ effects: this.setTooltips.of(Array.from(this.tooltips.values())) })
        )
    }

    private toKey(range: HovercardRange): string {
        return `${range.from}:${range.to}`
    }
}

function hovercardManager(): Extension {
    const [tooltips, , setTooltips] = createUpdateableField<Tooltip[]>([], field =>
        showTooltip.computeN([field], state => {
            return state.field(field)
        })
    )

    return [tooltips, ViewPlugin.define(view => new HovercardManager(view, setTooltips))]
}

/**
 * Facet to which an extension can add a value to show a hovercard.
 */
export const hovercardRanges = Facet.define<HovercardRange>({
    enables: [
        hoverCardTheme,
        // Compute CodeMirror tooltips from hovercard ranges
        hovercardManager(),
        // Highlight hovered token(s)
        validTokenRanges,
    ],
})

/**
 * Facet with which an extension can provide a hovercard source. For simplicity
 * only one source can be provided, others are ignored (in practice there is
 * only one source at the moment anyway).
 */
export const hovercardSource = Facet.define<HovercardSource, HovercardSource>({
    combine: sources => sources[0],
    enables: hovercard(),
})

/**
 * Converts mousemove events to hovercard ranges.
 */
class HoverPlugin implements PluginValue {
    position: { from: number; to: number } | null = null
    subscription: Subscription = new Subscription()

    constructor(
        readonly view: EditorView,
        readonly setHovercardPosition: StateEffectType<HovercardRange | null>,
        nextOffset: Subject<number | null>
    ) {
        this.subscription.add(
            nextOffset
                .pipe(
                    scan((position: { from: number; to: number } | null, offset) => {
                        if (offset === null) {
                            return null
                        }
                        if (position && position.from <= offset && position.to >= offset) {
                            // Nothing to do, return same value
                            return position
                        }

                        {
                            const word = this.view.state.wordAt(offset)
                            if (word) {
                                return { from: word.from, to: word.to }
                            }
                        }

                        return null
                    }, null),
                    distinctUntilChanged()
                )
                .subscribe(position => {
                    this.view.dispatch({
                        effects: setHovercardPosition.of(
                            position
                                ? {
                                      ...position,
                                      range: offsetToPosition(this.view.state.doc, position.from, position.to),
                                  }
                                : null
                        ),
                    })
                })
        )
    }

    destroy() {
        this.subscription.unsubscribe()
    }
}

function hovercard(): Extension {
    const [hovercardRange, , setHovercardRange] = createUpdateableField<HovercardRange | null>(null, field =>
        hovercardRanges.computeN([field], state => {
            const range = state.field(field)
            return range ? [range] : []
        })
    )
    const nextOffset = new Subject<number | null>()

    return [
        hovercardRange,
        ViewPlugin.define(view => new HoverPlugin(view, setHovercardRange, nextOffset), {
            eventHandlers: {
                mousemove(event: MouseEvent, view) {
                    let pos = view.posAtCoords(event)

                    if (pos) {
                        // It seems CodeMirror returns the document position _closest_ to mouse position.
                        // This has the unfortunate effect that hovering over empty parts a line will find
                        // the closest word next to it, which is not something we want. To ensure that we
                        // only consider positions of actual words/characters we perform the inverse
                        // conversion and compare the results. This is also done by CodeMirror's own hover
                        // tooltip plugin.
                        const posCoords = view.coordsAtPos(pos)
                        if (
                            posCoords == null ||
                            event.y < posCoords.top ||
                            event.y > posCoords.bottom ||
                            event.x < posCoords.left - this.view.defaultCharacterWidth ||
                            event.x > posCoords.right + this.view.defaultCharacterWidth
                        ) {
                            pos = null
                        }
                    }

                    nextOffset.next(pos)
                },
            },
        }),
    ]
}

// WebHoverOverlay requires to be passed an element representing the currently
// hovered token.  Since we don't have/want that for CodeMirror we are passing a
// dummy element.
const dummyHoveredElement = document.createElement('span')
// WebHoverOverlay expects to be passed the overlay position. Since CodeMirror
// positions the element we always use the same value.
const dummyOverlayPosition = { left: 0, bottom: 0 }

/**
 * This class is responsible for rendering a WebHoverOverlay component as a
 * CodeMirror tooltip. When constructed the instance subscribes to the hovercard
 * data source and the component props, and updates the component as it receives
 * changes.
 */
class Hovercard implements TooltipView {
    public dom: HTMLElement
    private root: Root | null = null
    private nextRoot = new Subject<Root>()
    private nextProps = new Subject<BlobProps>()
    private props: BlobProps | null = null
    public overlap = true
    private subscription: Subscription

    static create(view: EditorView, range: HovercardRange): Hovercard {
        return new Hovercard(view, range)
    }

    constructor(readonly view: EditorView, readonly range: HovercardRange) {
        this.dom = document.createElement('div')

        this.subscription = combineLatest([
            this.nextRoot,
            this.view.state.facet(hovercardSource)(view, range.range.start),
            this.nextProps.pipe(startWith(view.state.facet(blobPropsFacet))),
        ]).subscribe(([root, hovercardData, props]) => {
            this.render(root, hovercardData, props)
        })
    }

    public mount() {
        this.root = createRoot(this.dom)
        this.nextRoot.next(this.root)
    }

    public update(update: ViewUpdate) {
        // Umount React components when tooltip range does exist anymore
        if (
            !update.state
                .facet(hovercardRanges)
                .some(range => range.from === this.range.from && range.to === this.range.to)
        ) {
            this.root?.unmount()
            this.subscription.unsubscribe()
            return
        }

        // Update component when props change
        const props = update.state.facet(blobPropsFacet)
        if (this.props !== props) {
            this.props = props
            this.nextProps.next(props)
        }
    }

    private addRange() {
        window.requestAnimationFrame(() => {
            this.view.dispatch({ effects: addRange.of(this.range) })
        })
    }

    private removeRange() {
        window.requestAnimationFrame(() => {
            this.view.dispatch({ effects: removeRange.of(this.range) })
        })
    }

    private render(root: Root, { hoverOrError, actionsOrError }: HovercardData, props: BlobProps) {
        let hoverContext = {
            commitID: props.blobInfo.commitID,
            filePath: props.blobInfo.filePath,
            repoName: props.blobInfo.repoName,
            revision: props.blobInfo.revision,
        }

        let hoveredToken: WebHoverOverlayProps['hoveredToken'] = {
            ...hoverContext,
            ...this.range.range.start,
        }

        if (hoverOrError && hoverOrError !== 'loading' && !isErrorLike(hoverOrError) && hoverOrError.range) {
            hoveredToken = {
                ...hoveredToken,
                line: hoverOrError.range.start.line + 1,
                character: hoverOrError.range.start.character + 1,
            }
        }

        if (hoverOrError) {
            this.addRange()

            root.render(
                <Container onRender={() => repositionTooltips(this.view)} history={props.history}>
                    <div className="cm-code-intel-hovercard">
                        <WebHoverOverlay
                            // Blob props
                            location={props.location}
                            onHoverShown={props.onHoverShown}
                            isLightTheme={props.isLightTheme}
                            platformContext={props.platformContext}
                            settingsCascade={props.settingsCascade}
                            telemetryService={props.telemetryService}
                            extensionsController={props.extensionsController}
                            nav={props.nav ?? (url => props.history.push(url))}
                            // Hover props
                            actionsOrError={actionsOrError}
                            hoverOrError={hoverOrError}
                            // CodeMirror handles the positioning but a
                            // non-nullable value must be passed for the
                            // hovercard to render
                            overlayPosition={dummyOverlayPosition}
                            hoveredToken={hoveredToken}
                            hoveredTokenElement={dummyHoveredElement}
                            pinOptions={{
                                showCloseButton: true,
                                onCloseButtonClick: () => {
                                    const parameters = new URLSearchParams(props.location.search)
                                    parameters.delete('popover')

                                    updateBrowserHistoryIfChanged(props.history, props.location, parameters)
                                },
                                onCopyLinkButtonClick: async () => {
                                    const context = {
                                        position: this.range.range.start,
                                        range: {
                                            start: this.range.range.start,
                                            end: this.range.range.start,
                                        },
                                    }
                                    const search = new URLSearchParams(location.search)
                                    search.set('popover', 'pinned')
                                    updateBrowserHistoryIfChanged(
                                        props.history,
                                        props.location,
                                        addLineRangeQueryParameter(search, toPositionOrRangeQueryParameter(context))
                                    )
                                    await navigator.clipboard.writeText(window.location.href)
                                },
                            }}
                        />
                    </div>
                </Container>
            )
        } else {
            this.removeRange()
            root.render([])
        }
    }
}
