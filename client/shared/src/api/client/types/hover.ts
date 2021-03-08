import { MarkupKind, Range } from '@sourcegraph/extension-api-classes'
import { Hover as PlainHover, Range as PlainRange } from '@sourcegraph/extension-api-types'
import { Badged, Hover, MarkupContent, HoverAlert, AggregableTag } from 'sourcegraph'

/** A hover that is merged from multiple Hover results and normalized. */
export interface HoverMerged {
    contents: Badged<MarkupContent>[]
    alerts?: Badged<HoverAlert>[]
    range?: PlainRange

    /** Sorted and de-duplicated set of tags in all source hover values. */
    aggregatedTags?: AggregableTag[]
}

/** Create a merged hover from the given individual hovers. */
export function fromHoverMerged(values: (Badged<Hover | PlainHover> | null | undefined)[]): HoverMerged | null {
    const contents: HoverMerged['contents'] = []
    const alerts: HoverMerged['alerts'] = []
    const aggregatedTags = new Map<string, AggregableTag>()
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

            const tags = [result, ...(result.alerts || [])].flatMap(badged => badged.aggregableTags || [])

            for (const tag of tags) {
                aggregatedTags.set(tag.text, tag)
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

    return {
        contents,
        alerts,
        ...(range ? { range } : {}),
        ...(aggregatedTags.size > 0
            ? { aggregatedTags: [...aggregatedTags.values()].sort((a, b) => a.text.localeCompare(b.text)) }
            : {}),
    }
}
