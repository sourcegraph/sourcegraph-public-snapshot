import { MarkupKind, Range } from '@sourcegraph/extension-api-classes'
import { Hover as PlainHover, Range as PlainRange } from '@sourcegraph/extension-api-types'
import { Badged, Hover, MarkupContent, HoverAlert } from 'sourcegraph'

/** A hover that is merged from multiple Hover results and normalized. */
export interface HoverMerged {
    contents: Badged<MarkupContent>[]
    alerts?: Badged<HoverAlert>[]
    range?: PlainRange
}

/** Create a merged hover from the given individual hovers. */
export function fromHoverMerged(values: (Badged<Hover | PlainHover> | null | undefined)[]): HoverMerged | null {
    const contents: HoverMerged['contents'] = []
    const alerts: HoverMerged['alerts'] = []
    let range: PlainRange | undefined
    for (const result of values) {
        if (result) {
            if (result.contents?.value) {
                contents.push({
                    value: result.contents.value,
                    kind: result.contents.kind || MarkupKind.PlainText,
                    badge: result.badge,
                })
            }
            if (result.alerts) {
                alerts.push(...result.alerts)
            }

            if (result.range && !range) {
                // TODO(tj): Merge ranges so we highlight all provided ranges, not just the first range
                if (result.range instanceof Range) {
                    range = result.range.toPlain()
                } else {
                    range = result.range
                }
            }
        }
    }

    if (contents.length === 0) {
        return null
    }
    return range ? { contents, alerts, range } : { contents, alerts }
}
