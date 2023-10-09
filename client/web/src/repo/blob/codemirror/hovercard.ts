import { Facet } from '@codemirror/state'
import type { EditorView, TooltipView } from '@codemirror/view'
import type { Observable } from 'rxjs'
import type { WebHoverOverlayProps } from 'src/components/WebHoverOverlay'

import type { UIRangeSpec } from '@sourcegraph/shared/src/util/url'

export type UIRange = UIRangeSpec['range']

/**
 * Hover information received from a hover source.
 */
export type HoverData = Pick<WebHoverOverlayProps, 'hoverOrError' | 'actionsOrError'>

export type IHovercard = new (
    view: EditorView,
    tokenRange: UIRange,
    pinned: boolean,
    hovercardData: Observable<HoverData>
) => TooltipView

export const hoverCardConstructor = Facet.define<IHovercard, IHovercard | null>({
    combine(values) {
        return values[0] ?? null
    },
})

export function createHovercard(
    view: EditorView,
    tokenRange: UIRange,
    pinned: boolean,
    hovercardData: Observable<HoverData>
): TooltipView {
    const constructor = view.state.facet(hoverCardConstructor)
    if (!constructor) {
        throw new Error('A Hovercard constructor has to be provided via the `hoverCardConstructor` facet.')
    }
    return new constructor(view, tokenRange, pinned, hovercardData)
}
