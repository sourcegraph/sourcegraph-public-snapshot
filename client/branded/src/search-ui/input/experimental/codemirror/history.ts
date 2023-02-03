import { Extension, Prec } from '@codemirror/state'
import { Decoration, EditorView, ViewPlugin, WidgetType } from '@codemirror/view'
import { mdiClockOutline } from '@mdi/js'
import { formatDistanceToNow, parseISO } from 'date-fns'
import { Fzf, FzfOptions } from 'fzf'

import { pluralize } from '@sourcegraph/common'
import { RecentSearch } from '@sourcegraph/shared/src/settings/temporary/recentSearches'
import { createSVGIcon } from '@sourcegraph/shared/src/util/dom'

import { clearMode, getSelectedMode, ModeDefinition, modesFacet, setMode } from '../modes'
import { queryRenderer } from '../optionRenderer'
import { Source, suggestionSources, Option } from '../suggestionsExtension'

const theme = EditorView.theme({
    '.sg-history-button': {
        marginRight: '0.25rem',
        paddingRight: '0.25rem',
        borderRight: '1px solid var(--border-color-2)',
        color: 'var(--icon-color)',
    },
    '.sg-history-button button': {
        width: '1rem',
        border: '0',
        backgroundColor: 'transparent',
        padding: 0,
        color: 'inherit',
    },
    '.sg-history-button svg': {
        display: 'inline-block',
        width: 'var(--icon-inline-size)',
        height: 'var(--icon-inline-size)',
        // Setting this simplifies event handling for the history button widget
        pointerEvents: 'none',
    },
    '.sg-mode-History .sg-history-button': {
        color: 'var(--logo-purple)',
        marginRight: '0',
        border: '0',
    },
})

function toggleHistoryMode(event: MouseEvent | KeyboardEvent, view: EditorView): void {
    event.preventDefault()
    const selectedMode = getSelectedMode(view.state)
    if (selectedMode?.name === 'History') {
        clearMode(view)
    } else {
        setMode(view, 'History')
    }
    view.focus()
}

/**
 * This ViewPlugin renders the history button at the beginning of the search
 * input.
 */
const historyButton = ViewPlugin.define(
    () => ({
        decorations: Decoration.set(
            Decoration.widget({
                side: -1,
                widget: new (class extends WidgetType {
                    public toDOM(view: EditorView): HTMLElement {
                        const container = document.createElement('span')
                        container.className = 'sg-history-button'
                        const button = document.createElement('button')
                        button.type = 'button'
                        const icon = createSVGIcon(mdiClockOutline)

                        container.append(button)
                        button.append(icon)
                        button.addEventListener('click', event => toggleHistoryMode(event, view))
                        return container
                    }
                })(),
            }).range(0)
        ),
    }),
    {
        decorations: plugin => plugin.decorations,
    }
)

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
    const applySuggestion = (option: Option): void => {
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
        theme,
        Prec.highest(historyButton),
        suggestionSources.of({
            query: createHistorySuggestionSource(config.source, config.submitQuery),
            mode: config.mode.name,
        }),
    ]
}
