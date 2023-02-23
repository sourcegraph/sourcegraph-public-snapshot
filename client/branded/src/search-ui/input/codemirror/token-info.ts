import { Extension, MapMode, StateEffect, StateField } from '@codemirror/state'
import { Decoration, EditorView, hoverTooltip, TooltipView } from '@codemirror/view'

import { renderMarkdown } from '@sourcegraph/common'
import type { DecoratedToken } from '@sourcegraph/shared/src/search/query/decoratedToken'
import { toHover } from '@sourcegraph/shared/src/search/query/hover'
import { Node } from '@sourcegraph/shared/src/search/query/parser'
import { KeywordKind } from '@sourcegraph/shared/src/search/query/token'
import { resolveFilterMemoized } from '@sourcegraph/shared/src/search/query/utils'

import { decoratedTokens, parsedQuery } from './parsedQuery'

// Defines decorators for syntax highlighting
const tokenHoverDecoration = Decoration.mark({ class: 'sg-decorated-token-hover' })

const hoverStyle = [
    // Overwrite styles for built-in hover element
    EditorView.theme({
        '.cm-tooltip': {
            padding: '0.25rem',
            color: 'var(--search-query-text-color)',
            backgroundColor: 'var(--color-bg-1)',
            border: '1px solid var(--border-color)',
            borderRadius: 'var(--border-radius)',
            boxShadow: 'var(--box-shadow)',
            maxWidth: '50vw',
        },

        '.cm-tooltip p:last-child': {
            marginBottom: 0,
        },

        '.cm-tooltip code': {
            backgroundColor: 'rgba(220, 220, 220, 0.4)',
            borderRadius: 'var(--border-radius)',
            padding: '0 0.4em',
        },

        '.cm-tooltip-section': {
            paddingBottom: '0.25rem',
            borderTopColor: 'var(--border-color)',
        },

        '.cm-tooltip-section:last-child': {
            paddingTop: '0.25rem',
            paddingBottom: 0,
        },
        '.cm-tooltip-section:last-child:first-child': {
            padding: 0,
        },
    }),
    // Base style for custom classes
    EditorView.baseTheme({
        '.sg-decorated-token-hover': {
            backgroundColor: 'var(--gray-02)',
        },
        '&dark .sg-decorated-token-hover': {
            backgroundColor: 'var(--gray-08)',
        },
    }),
]

/**
 * Extension for providing token information. This includes showing a popover
 * on hover and highlighting the hovered token.
 */
export function tokenInfo(): Extension {
    const setHighlighedTokenPosition = StateEffect.define<number | null>()
    const highlightedTokenPosition = StateField.define<number | null>({
        create() {
            return null
        },
        update(position, transaction) {
            // Hide the highlight when the document changes. This replicates
            // Monaco's behavior.
            if (transaction.docChanged) {
                return null
            }
            const effect = transaction.effects.find((effect): effect is StateEffect<number | null> =>
                effect.is(setHighlighedTokenPosition)
            )
            if (effect) {
                position = effect?.value
            }
            if (position !== null) {
                // Mapping the position might not be necessary since we clear
                // the highlight when the document changes anyway, but this is
                // the safer way.
                // MapMode.TrackDel causes mapPos to return null if content at
                // this position was deleted (in which case we want to remove
                // the highlight)
                return transaction.changes.mapPos(position, 0, MapMode.TrackDel)
            }
            return position
        },
        provide(field) {
            return EditorView.decorations.compute([field, decoratedTokens], state => {
                const position = state.field(field)
                if (!position) {
                    return Decoration.none
                }

                const tooltipInfo = getTokensTooltipInformation(state.facet(decoratedTokens), position)
                if (!tooltipInfo) {
                    return Decoration.none
                }
                let { range } = tooltipInfo

                const token = tooltipInfo.tokensAtCursor[0]
                switch (token.type) {
                    case 'keyword': {
                        // Find operator (AND and OR are supported) and
                        // highlight its operands too if possible
                        const operator = findOperatorNode(position, state.facet(parsedQuery))
                        if (operator) {
                            range = operator.groupRange ?? operator.range
                        }
                        // Highlight operator keyword only
                        break
                    }
                }

                return Decoration.set([tokenHoverDecoration.range(range.start, range.end)])
            })
        },
    })

    return [
        hoverStyle,
        highlightedTokenPosition,
        // Highlights the hovered token
        EditorView.domEventHandlers({
            mousemove(event, view) {
                const position = view.posAtCoords(event)
                if (position && position !== view.state.field(highlightedTokenPosition)) {
                    view.dispatch({ effects: setHighlighedTokenPosition.of(position) })
                }
            },
            mouseleave(_event, view) {
                if (view.state.field(highlightedTokenPosition) !== null) {
                    view.dispatch({ effects: setHighlighedTokenPosition.of(null) })
                }
            },
        }),
        // Shows information about the hovered token
        hoverTooltip(
            (view, position) => {
                const tooltipInfo = getTokensTooltipInformation(view.state.facet(decoratedTokens), position)
                if (!tooltipInfo) {
                    return null
                }

                return {
                    pos: tooltipInfo.range.start,
                    // tooltipInfo.range.end is exclusive, but this needs to be
                    // inclusive to correctly hide the tooltip when the cursor
                    // moves to the next token
                    end: tooltipInfo.range.end - 1,
                    // Show token info above the text by default to avoid
                    // interfering with autcompletion (otherwise this could show
                    // the token info *below* the autocompletion popover, which
                    // looks bad)
                    above: true,
                    create(): TooltipView {
                        const dom = document.createElement('div')
                        dom.innerHTML = renderMarkdown(tooltipInfo.value)
                        return {
                            dom,
                        }
                    },
                }
            },
            {
                hoverTime: 100,
                // Hiding the tooltip when the document changes replicates
                // Monaco's behavior and also "feels right" because it removes
                // "clutter" from the input.
                hideOnChange: true,
            }
        ),
    ]
}

function getTokensTooltipInformation(
    tokens: readonly DecoratedToken[],
    position: number
): { tokensAtCursor: readonly DecoratedToken[]; range: { start: number; end: number }; value: string } | null {
    const tokensAtCursor = tokens.filter(token => {
        let { start, end } = token.range
        switch (token.type) {
            case 'field':
                // +1 to include field separator :
                end += 1
                break
        }
        return start <= position && end > position
    })

    if (tokensAtCursor?.length === 0) {
        return null
    }
    const values: string[] = []
    let range: { start: number; end: number } | undefined

    // Copied and adapted from getHoverResult (hover.ts)
    for (const token of tokensAtCursor) {
        switch (token.type) {
            case 'field': {
                const resolvedFilter = resolveFilterMemoized(token.value)
                if (resolvedFilter) {
                    values.push(
                        'negated' in resolvedFilter
                            ? resolvedFilter.definition.description(resolvedFilter.negated)
                            : resolvedFilter.definition.description
                    )
                    // +1 to include field separator :
                    range = { start: token.range.start, end: token.range.end + 1 }
                }
                break
            }
            case 'pattern':
            case 'metaRevision':
            case 'metaRepoRevisionSeparator':
            case 'metaSelector':
                values.push(toHover(token))
                range = token.range
                break
            case 'metaRegexp':
            case 'metaStructural':
            case 'metaPredicate':
                values.push(toHover(token))
                range = token.groupRange ? token.groupRange : token.range
                break
            case 'keyword':
                switch (token.kind) {
                    case KeywordKind.And:
                        values.push('Find results which match both the left and the right expression.')
                        range = token.range
                        break
                    case KeywordKind.Or:
                        values.push('Find results which match the left or the right expression.')
                        range = token.range
                        break
                }
        }
    }

    if (!range) {
        return null
    }
    return { tokensAtCursor, range, value: values.join('') }
}

function findOperatorNode(position: number, node: Node | null): Extract<Node, { type: 'operator' }> | null {
    if (!node || node.type !== 'operator' || node.range.start >= position || node.range.end <= position) {
        return null
    }
    for (const operand of node.operands) {
        const result = findOperatorNode(position, operand)
        if (result) {
            return result
        }
    }
    return node
}
