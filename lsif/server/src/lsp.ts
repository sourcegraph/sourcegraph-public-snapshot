import { Hover, MarkupContent, MarkupKind } from 'vscode-languageserver-types'

/**
 * Normalizes an LSP hover so it always uses MarkupContent and no union types.
 * DGraph does not support union types.
 */
export function normalizeHover(hover: Hover): Hover {
    const contents = Array.isArray(hover.contents) ? hover.contents : [hover.contents]
    return {
        ...hover,
        contents: {
            kind: MarkupKind.Markdown,
            value: contents
                .map(content => {
                    if (MarkupContent.is(content)) {
                        // Assume it's markdown. To be correct, markdown would need to be escaped for non-markdown kinds.
                        return content.value
                    }
                    if (typeof content === 'string') {
                        return content
                    }
                    if (!content.value) {
                        return ''
                    }
                    return '```' + content.language + '\n' + content.value + '\n```'
                })
                .filter(str => !!str.trim())
                .join('\n\n---\n\n'),
        },
    }
}
