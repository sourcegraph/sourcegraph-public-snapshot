import { Facet } from '@codemirror/state'
import { type EditorView, type PluginValue, ViewPlugin, type ViewUpdate } from '@codemirror/view'

import type { RecentSearch } from '@sourcegraph/shared/src/settings/temporary/recentSearches'

/**
 * Transactions which modify the input because a history item was selected are
 * tagged with this user event.
 */
export const HISTORY_USER_EVENT = 'input.sg-history'

/**
 * This view plugin binds arrow up/down keyboard event handlers and loads
 * the provided search history entries into the input.
 * The plugin does _not_ "wrap around" to the beginning or end of the history.
 * If the input is changed through any other means (mainly via user input), this
 * plugin does not load history entries anymore.
 * If the input is cleared the plugin will load history entries again.
 */
const historyView = ViewPlugin.fromClass(
    class implements PluginValue {
        private historyEntries: readonly RecentSearch[]
        private currentHistoryEntry = -1

        constructor(private view: EditorView) {
            this.historyEntries = this.view.state.facet(searchHistory)

            if (this.view.state.doc.length === 0) {
                this.view.dom.addEventListener('keydown', this.onKeyDown)
            }
        }

        public update(update: ViewUpdate): void {
            if (update.docChanged) {
                if (update.state.doc.length === 0) {
                    this.view.dom.addEventListener('keydown', this.onKeyDown)
                    this.currentHistoryEntry = -1
                } else if (update.transactions.some(transaction => !transaction.isUserEvent(HISTORY_USER_EVENT))) {
                    this.view.dom.removeEventListener('keydown', this.onKeyDown)
                }
            }

            const historyEntries = this.view.state.facet(searchHistory)
            if (this.historyEntries !== historyEntries) {
                this.historyEntries = historyEntries
                // Probably not necessary to update the view
            }
        }

        public destroy(): void {
            this.view.dom.removeEventListener('keydown', this.onKeyDown)
        }

        private updateInput(): void {
            // Revert to empty input if there is jnot selected history entry
            const query = this.currentHistoryEntry >= 0 ? this.historyEntries[this.currentHistoryEntry].query : ''
            this.view.dispatch({
                changes: { from: 0, to: this.view.state.doc.length, insert: query },
                selection: { anchor: query.length },
                userEvent: HISTORY_USER_EVENT,
            })
        }

        private onKeyDown = (event: KeyboardEvent): void => {
            switch (event.key) {
                case 'ArrowUp': {
                    {
                        event.preventDefault()
                        const nextHistoryEntry = this.currentHistoryEntry + 1
                        if (nextHistoryEntry < this.historyEntries.length) {
                            this.currentHistoryEntry = nextHistoryEntry
                            this.updateInput()
                        }
                    }
                    break
                }
                case 'ArrowDown': {
                    {
                        event.preventDefault()
                        const previousHistoryEntry = this.currentHistoryEntry - 1
                        this.currentHistoryEntry = previousHistoryEntry > -1 ? previousHistoryEntry : -1
                        this.updateInput()
                    }
                    break
                }
            }
        }
    }
)

/**
 * If provided this facet enables command-line style history cycling.
 */
export const searchHistory = Facet.define<RecentSearch[], RecentSearch[]>({
    combine(searches) {
        return searches
            .flat()
            .filter(search => search.query.trim() !== '')
            .sort((searchA, searchB) => searchB.timestamp.localeCompare(searchA.timestamp))
    },
    enables: historyView,
})
