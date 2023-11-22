import { type Extension, Facet, RangeSetBuilder } from '@codemirror/state'
import { Decoration, EditorView, WidgetType } from '@codemirror/view'

import type { Filter } from '@sourcegraph/shared/src/search/query/token'
import { resolveFilterMemoized } from '@sourcegraph/shared/src/search/query/utils'

import { queryTokens } from './parsedQuery'

const activeFilterFacet = Facet.define<Filter>()
const activeFilterExtension = activeFilterFacet.computeN([queryTokens, 'selection'], state => {
    // Do not mark a token as active if the user is selecting text. This avoids
    // conflicts with the selection color.
    if (!state.selection.main.empty) {
        return []
    }
    const query = state.facet(queryTokens)
    const position = state.selection.main.head
    return query.tokens.filter(
        (token): token is Filter =>
            // Inclusive end so that the filter is selected when
            // the cursor is positioned directly after the value
            token.type === 'filter' && token.range.start <= position && token.range.end >= position
    )
})

const activeFilterDecoration = Decoration.mark({ class: 'sg-query-filter-active' })

/**
 * An extension that adds the class .sg-query-filter-active to the filter token
 * "touched" by the cursor.
 */
export const decorateActiveFilter: Extension = [
    activeFilterExtension,
    EditorView.decorations.compute([activeFilterFacet], state => {
        const selectedFilters = state.facet(activeFilterFacet)
        if (selectedFilters.length === 0) {
            return Decoration.none
        }

        const decorations = new RangeSetBuilder<Decoration>()
        for (const filter of selectedFilters) {
            decorations.add(filter.range.start, filter.range.end, activeFilterDecoration)
        }

        return decorations.finish()
    }),
]

class PlaceholderWidget extends WidgetType {
    constructor(private placeholder: string) {
        super()
    }

    public eq(other: PlaceholderWidget): boolean {
        return this.placeholder === other.placeholder
    }

    public toDOM(): HTMLElement {
        const span = document.createElement('span')
        span.className = 'sg-query-filter-placeholder'
        span.textContent = this.placeholder
        return span
    }
}

/**
 * An extension that shows the preconfigured placeholder (if available) for the active
 * filter. The placeholder is given the class .sg-query-filter-placeholder
 */
export const filterPlaceholder: Extension = [
    activeFilterExtension,
    EditorView.decorations.compute([activeFilterFacet], state => {
        const selectedFilters = state.facet(activeFilterFacet)
        if (selectedFilters.length === 0) {
            return Decoration.none
        }

        const decorations = new RangeSetBuilder<Decoration>()
        for (const filter of selectedFilters) {
            if (!filter.value?.value) {
                const resolvedFilter = resolveFilterMemoized(filter.field.value)
                if (resolvedFilter?.definition.placeholder) {
                    decorations.add(
                        filter.range.end,
                        filter.range.end,
                        Decoration.widget({
                            widget: new PlaceholderWidget(resolvedFilter.definition.placeholder),
                            side: 1, // show after the cursor
                        })
                    )
                }
            }
        }

        return decorations.finish()
    }),
    EditorView.baseTheme({
        '.sg-query-filter-placeholder': {
            pointerEvents: 'none',
        },
    }),
]
