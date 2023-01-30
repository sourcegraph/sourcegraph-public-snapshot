import { Extension, Prec } from '@codemirror/state'
import { Decoration, EditorView, ViewPlugin, WidgetType } from '@codemirror/view'
import { mdiClockOutline } from '@mdi/js'
import { formatDistanceToNow, parseISO } from 'date-fns'
import { Fzf, FzfOptions } from 'fzf'

import { pluralize } from '@sourcegraph/common'
import { RecentSearch } from '@sourcegraph/shared/src/settings/temporary/recentSearches'
import { createSVGIcon } from '@sourcegraph/shared/src/util/dom'

import { queryRenderer } from '../optionRenderer'
import {
    clearMode,
    getSelectedMode,
    setMode,
    Source,
    suggestionSources,
    Option,
    ModeDefinition,
    suggestionModes,
} from '../suggestionsExtension'

const theme = EditorView.theme({
    '.sg-history-button': {
        border: 'none',
        backgroundColor: 'transparent',
        padding: 0,
        marginRight: '0.5rem',
        width: 'var(--icon-inline-size)',
        height: 'var(--icon-inline-size)',
        color: 'var(--icon-color)',
        verticalAlign: 'text-top',
    },
    '.sg-history-button > svg': {
        // Setting this simplifies event handling for the history button widget
        pointerEvents: 'none',
        // This is overwritten to 'middle' which doesn't look right
        verticalAlign: 'initial',
    },
    '.sg-mode-History .sg-history-button': {
        marginRight: '0.125rem',
        color: 'var(--logo-purple)',
    },
})

function toggleHistoryMode(event: MouseEvent | KeyboardEvent, view: EditorView): void {
    if ((event.target as HTMLElement).classList.contains('sg-history-button')) {
        event.preventDefault()
        const selectedMode = getSelectedMode(view.state)
        if (selectedMode?.name === 'History') {
            clearMode(view)
        } else {
            setMode(view, 'History')
        }
        view.focus()
    }
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
                    public toDOM(_view: EditorView): HTMLElement {
                        const button = document.createElement('button')
                        button.className = 'sg-history-button'
                        button.type = 'button'
                        const icon = createSVGIcon(mdiClockOutline)
                        button.append(icon)
                        return button
                    }
                    public ignoreEvent(): boolean {
                        return false
                    }
                })(),
            }).range(0)
        ),
    }),
    {
        decorations: plugin => plugin.decorations,
        eventHandlers: {
            // Click doesn't work because it moves the cursor to the beginning
            // of the input.
            mousedown: toggleHistoryMode,
            keydown: (event, view) => {
                if (event.key === ' ') {
                    toggleHistoryMode(event, view)
                }
            },
        },
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
                        description: `• ${item.resultCount}${item.limitHit ? '+' : ''} ${pluralize(
                            'result',
                            item.resultCount
                        )} • ${formatDistanceToNow(parseISO(item.timestamp), formatTimeOptions)}`,
                        action: {
                            type: 'command',
                            name: 'Run',
                            apply: applySuggestion,
                        },
                        alternativeAction: {
                            type: 'completion',
                            name: 'Edit query',
                            insertValue: item.query + ' ',
                            from: 0,
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
        suggestionModes.of([config.mode]),
        theme,
        Prec.highest(historyButton),
        suggestionSources.of({
            query: createHistorySuggestionSource(config.source, config.submitQuery),
            mode: config.mode.name,
        }),
    ]
}
