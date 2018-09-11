import { Hover, MarkupContent, MarkupKind } from 'sourcegraph'
import { Hover as PlainHover, Range } from '../../protocol/plainTypes'

/** A hover that is merged from multiple Hover results and normalized. */
export interface HoverMerged {
    /**
     * @todo Make this type *just* {@link MarkupContent} when all consumers are updated.
     */
    contents:
        | MarkupContent
        | string
        | { language: string; value: string }
        | (MarkupContent | string | { language: string; value: string })[]

    range?: Range
}

export namespace HoverMerged {
    /** Create a merged hover from the given individual hovers. */
    export function from(values: (Hover | PlainHover | null | undefined)[]): HoverMerged | null {
        const contents: HoverMerged['contents'] = []
        let range: HoverMerged['range']
        for (const result of values) {
            if (result) {
                if (result.contents && result.contents.value) {
                    contents.push({
                        value: result.contents.value,
                        kind: result.contents.kind || ('plaintext' as MarkupKind),
                    })
                }
                const __backcompatContents = result.__backcompatContents // tslint:disable-line deprecation
                if (__backcompatContents) {
                    for (const content of Array.isArray(__backcompatContents)
                        ? __backcompatContents
                        : [__backcompatContents]) {
                        if (typeof content === 'string') {
                            if (content) {
                                contents.push(content)
                            }
                        } else if ('language' in content) {
                            if (content.language && content.value) {
                                contents.push(content)
                            }
                        } else if ('value' in content) {
                            if (content.value) {
                                contents.push(content.value)
                            }
                        }
                    }
                }
                if (result.range && !range) {
                    range = result.range
                }
            }
        }
        return contents.length === 0 ? null : range ? { contents, range } : { contents }
    }
}
