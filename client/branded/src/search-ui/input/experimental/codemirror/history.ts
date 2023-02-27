import type { Extension } from '@codemirror/state'
import { EditorView } from '@codemirror/view'
import { mdiClockOutline } from '@mdi/js'
import { formatDistanceToNow, parseISO } from 'date-fns'
import { Fzf, type FzfOptions } from 'fzf'

import { pluralize } from '@sourcegraph/common'
import { type RecentSearch } from '@sourcegraph/shared/src/settings/temporary/recentSearches'

import { type ModeDefinition, modesFacet, setMode } from '../modes'
import { queryRenderer } from '../optionRenderer'
import { type Source, suggestionSources, type Option } from '../suggestionsExtension'

const fzfOptions: FzfOptions<RecentSearch> = {
    selector: search => search.query,
    // match: extendedMatch,
    // fuzzy: false,
}

const formatTimeOptions = {
    addSuffix: true,
}

function createHistorySuggestionSource(
    source: () => RecentSearch[],
    submitQuery: (query: string) => void
): Source['query'] {
    const applySuggestion = (option: Option, view: EditorView): void => {
        setMode(view, null)
        view.dispatch({ changes: { from: 0, to: view.state.doc.length, insert: option.label } })
        submitQuery(option.label)
    }

    return state => {
        const query = state.sliceDoc()
        const fzf = new Fzf(source(), fzfOptions)
        const results = fzf.find(query)
        return {
            result: [
                {
                    title: 'History',
                    options: results.map(({ item, positions }) => ({
                        label: item.query,
                        icon: mdiClockOutline,
                        matches: positions,
                        action: {
                            type: 'command',
                            name: `${item.resultCount}${item.limitHit ? '+' : ''} ${pluralize(
                                'result',
                                item.resultCount
                            )} • ${formatDistanceToNow(parseISO(item.timestamp), formatTimeOptions)}`,
                            apply: applySuggestion,
                            info: 'run the query',
                        },
                        render: queryRenderer,
                    })),
                },
            ],
        }
    }
}

export function searchHistoryExtension(config: {
    mode: ModeDefinition
    source: () => RecentSearch[]
    submitQuery: (query: string) => void
}): Extension {
    return [
        modesFacet.of([config.mode]),
        suggestionSources.of({
            query: createHistorySuggestionSource(config.source, config.submitQuery),
            mode: config.mode.name,
        }),
    ]
}
