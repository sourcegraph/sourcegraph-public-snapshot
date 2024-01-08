import { Facet } from '@codemirror/state'
import { type DecorationSet, EditorView } from '@codemirror/view'

/**
 * Facet to specify a range in which all decorations registered via
 * {@link codeIntelDecorations} should be ignored. This really is
 * only used by {@link selectedToken}, but a separate facet is provided
 * to avoid circular dependencies between modules.
 */
export const ignoreDecorations = Facet.define<{ from: number; to: number } | null, { from: number; to: number } | null>(
    {
        combine(value) {
            return value[0] ?? null
        },
    }
)

/**
 * We can't add/remove any decorations inside the selected token, because
 * that causes the node to be recreated and lose focus, which breaks
 * token keyboard navigation.
 * This facet should be used by all codeIntel extensions to ensure that any
 * conflicting decoration is removed.
 */
export const codeIntelDecorations = Facet.define<DecorationSet>({
    enables: self =>
        EditorView.decorations.computeN([self, ignoreDecorations], state => {
            const decorationSets = state.facet(self)
            const range = state.facet(ignoreDecorations)
            const filter = range ? { filterFrom: range.from, filterTo: range.to, filter: () => false } : null
            return filter ? decorationSets.map(set => set.update(filter)) : decorationSets
        }),
})
