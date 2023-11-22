import { RangeSetBuilder, type Extension } from '@codemirror/state'
import { Decoration, type EditorView, ViewPlugin, type DecorationSet } from '@codemirror/view'
import prism from 'prismjs'

import { SyntaxKind } from '$lib/shared'

import 'prismjs/components/prism-go'

/**
 * Implements experimental client side syntax highlighting for Go files.
 * The idea is that downloading only the raw text from the server results in
 * smaller payload and avoids memory allocation for highlighting information,
 * which can be quite significant for large files.
 * Instead only the visible range is highlighted on demand. Here this is done
 * with client side syntax highlighting.
 * However it's broken if a token starts outside of the visible range, e.g.
 * multiline strings.
 */
export function highlight(): Extension {
    return ViewPlugin.define(
        view => {
            const decorationCache: Partial<Record<SyntaxKind, Decoration>> = {}
            function decorate(view: EditorView): DecorationSet {
                const builder = new RangeSetBuilder<Decoration>()
                const tokens = prism.tokenize(
                    view.state.sliceDoc(view.viewport.from, view.viewport.to),
                    prism.languages.go
                )
                let position = view.viewport.from
                for (const token of tokens) {
                    const from = position
                    const to = from + token.length
                    position = to
                    if (typeof token !== 'string') {
                        const kind = getType(token.type)
                        if (kind) {
                            const decoration =
                                decorationCache[kind] ??
                                (decorationCache[kind] = Decoration.mark({ class: `hl-typed-${SyntaxKind[kind]}` }))
                            builder.add(from, to, decoration)
                        }
                    }
                }
                return builder.finish()
            }

            return {
                decorations: decorate(view),
                update(update) {
                    if (update.viewportChanged) {
                        this.decorations = decorate(update.view)
                    }
                },
            }
        },
        {
            decorations(plugin) {
                return plugin.decorations
            },
        }
    )
}

function getType(type: string): SyntaxKind | null {
    switch (type) {
        case 'comment': {
            return SyntaxKind.Comment
        }
        case 'keyword': {
            return SyntaxKind.IdentifierKeyword
        }
        case 'builtin': {
            return SyntaxKind.IdentifierBuiltin
        }
        case 'class-name':
        case 'function': {
            return SyntaxKind.IdentifierFunction
        }
        case 'boolean': {
            return SyntaxKind.BooleanLiteral
        }
        case 'number': {
            return SyntaxKind.NumericLiteral
        }
        case 'string': {
            return SyntaxKind.StringLiteral
        }
        case 'char': {
            return SyntaxKind.CharacterLiteral
        }
        case 'variable':
        case 'symbol': {
            return SyntaxKind.Identifier
        }
        case 'constant': {
            return SyntaxKind.IdentifierConstant
        }
        case 'property': {
            return SyntaxKind.IdentifierAttribute
        }
        case 'punctuation': {
            return SyntaxKind.PunctuationDelimiter
        }
        case 'operator': {
            return SyntaxKind.IdentifierOperator
        }
        case 'regex':
        case 'url': {
            return SyntaxKind.UnspecifiedSyntaxKind
        }
    }
    return null
}
