import { Hover as PlainHover, Range } from '@sourcegraph/extension-api-types'
import { Hover, MarkupContent, MarkupKind } from 'sourcegraph'

/** A hover that is merged from multiple Hover results and normalized. */
export interface HoverMerged {
    /**
     * @todo Make this type *just* {@link MarkupContent} when all consumers are updated.
     */
    contents: MarkupContent[]

    range?: Range
}

function hoverPriority(value: Hover | PlainHover | null | undefined): number | undefined {
    return value && 'priority' in value ? value.priority : undefined
}

export namespace HoverMerged {
    /** Create a merged hover from the given individual hovers. */
    export function from(values: (Hover | PlainHover | null | undefined)[]): HoverMerged | null {
        // Sort by priority.
        values = values.sort((a, b) => {
            const ap = hoverPriority(a)
            const bp = hoverPriority(b)
            if (ap === undefined && bp === undefined) {
                return 0
            } else if (ap === undefined) {
                return 1
            } else if (bp === undefined) {
                return -1
            }
            return bp - ap
        })

        const maxPriority = values.reduce((max: undefined | number, v: Hover | PlainHover | null | undefined) => {
            const priority = hoverPriority(v)
            if (typeof priority === 'number' && (max === undefined || priority > max)) {
                return priority
            }
            return max
        }, undefined)

        const contents: HoverMerged['contents'] = []
        let range: Range | undefined
        for (const result of values) {
            if (result) {
                if (
                    typeof result.priority === 'number' &&
                    typeof maxPriority === 'number' &&
                    result.priority < 0 &&
                    result.priority < maxPriority
                ) {
                    continue
                }

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
                                contents.push({ value: content, kind: 'plaintext' as MarkupKind })
                            }
                        } else if ('language' in content) {
                            if (content.language && content.value) {
                                contents.push({
                                    value: toMarkdownCodeBlock(content.language, content.value),
                                    kind: 'markdown' as MarkupKind,
                                })
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

function toMarkdownCodeBlock(language: string, value: string): string {
    return '```' + language + '\n' + value + '\n```\n'
}
