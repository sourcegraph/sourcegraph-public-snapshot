import { Fzf } from 'fzf'
import { FILTERS, FilterType } from '@sourcegraph/shared/src/search/query/filters'
import { Source } from './suggestions'

const FILTER_SUGGESTIONS = new Fzf(
    (Object.keys(FILTERS) as FilterType[]).map(filter => ({ value: filter })),
    { selector: item => item.value }
)

export const filterSuggestions: Source = (state, pos) => {
    const word = state.wordAt(pos)
    if (!word) {
        return []
    }
    return FILTER_SUGGESTIONS.find(state.sliceDoc(word.from, word.to)).map(entry => entry.item)
}
