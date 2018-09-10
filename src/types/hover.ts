import { Hover, MarkupContent } from 'sourcegraph'
import { Hover as PlainHover, Range } from '../protocol/plainTypes'

/** A hover that is merged from multiple Hover results and normalized. */
export interface HoverMerged {
    contents: MarkupContent[]
    range?: Range
}

export namespace HoverMerged {
    /** Create a merged hover from the given individual hovers. */
    export function from(values: (Hover | PlainHover | null)[]): HoverMerged | null {
        const contents: HoverMerged['contents'] = []
        let range: HoverMerged['range']
        for (const result of values) {
            if (result) {
                if (Array.isArray(result.contents)) {
                    contents.push(...result.contents)
                } else {
                    contents.push(result.contents)
                }
                if (result.range && !range) {
                    range = result.range
                }
            }
        }
        return contents.length === 0 ? null : range ? { contents, range } : { contents }
    }
}
