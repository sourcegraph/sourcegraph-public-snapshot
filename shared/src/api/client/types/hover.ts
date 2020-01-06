import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { Badged, Hover as PlainHover, Range } from '@sourcegraph/extension-api-types'
import { Hover, MarkupContent, BadgeAttachmentRenderOptions } from 'sourcegraph'

/** A hover that is merged from multiple Hover results and normalized. */
export interface HoverMerged {
    /**
     * @todo Make this type *just* {@link MarkupContent} when all consumers are updated.
     */
    contents: MarkupContent[]

    range?: Range

    badges?: BadgeAttachmentRenderOptions[]
}

/** Create a merged hover from the given individual hovers. */
export function fromHoverMerged(values: (Badged<Hover> | Badged<PlainHover> | null | undefined)[]): HoverMerged | null {
    const contents: HoverMerged['contents'] = []
    let range: Range | undefined
    const badges: BadgeAttachmentRenderOptions[] = []
    for (const result of values) {
        if (result) {
            if (result.contents && result.contents.value) {
                contents.push({
                    value: result.contents.value,
                    kind: result.contents.kind || MarkupKind.PlainText,
                })
            }
            const __backcompatContents = result.__backcompatContents
            if (__backcompatContents) {
                for (const content of Array.isArray(__backcompatContents)
                    ? __backcompatContents
                    : [__backcompatContents]) {
                    if (typeof content === 'string') {
                        if (content) {
                            contents.push({ value: content, kind: MarkupKind.Markdown })
                        }
                    } else if ('language' in content) {
                        if (content.language && content.value) {
                            contents.push({
                                value: toMarkdownCodeBlock(content.language, content.value),
                                kind: MarkupKind.Markdown,
                            })
                        }
                    }
                }
            }
            if (result.range && !range) {
                range = result.range
            }
            if (result.badge) {
                badges.push(result.badge)
            }
        }
    }
    const merged: HoverMerged | null = contents.length === 0 ? null : range ? { contents, range } : { contents }
    if (merged && badges.length > 0) {
        merged.badges = badges
    }
    return merged
}

function toMarkdownCodeBlock(language: string, value: string): string {
    return '```' + language + '\n' + value + '\n```\n'
}
