/**
 * This extension extends CodeMirror's own search extension with a custom search
 * UI.
 */

import {
    findNext,
    findPrevious,
    getSearchQuery,
    openSearchPanel,
    closeSearchPanel,
    search as codemirrorSearch,
    searchKeymap,
    SearchQuery,
    setSearchQuery,
} from '@codemirror/search'
import {
    Compartment,
    type Extension,
    StateEffect,
    type TransactionSpec,
    type Text as CodeMirrorText,
    type SelectionRange,
} from '@codemirror/state'
import {
    EditorView,
    type KeyBinding,
    keymap,
    type Panel,
    runScopeHandlers,
    ViewPlugin,
    type ViewUpdate,
} from '@codemirror/view'
import classNames from 'classnames'
import { Subject, Subscription } from 'rxjs'
import { debounceTime, distinctUntilChanged, startWith, tap } from 'rxjs/operators'

import { createUpdateableField } from '@sourcegraph/shared/src/components/codemirror/utils'
import { createElement } from '@sourcegraph/shared/src/util/dom'

import { SearchPanelViewMode } from './types'

import styles from './search.module.scss'

// Match 'from' position -> 1-based serial number (index) of this match in the document.
type SearchMatches = Map<number, number>

export const BLOB_SEARCH_CONTAINER_ID = 'blob-search-container'

const focusSearchInput = StateEffect.define<boolean>()

export interface SearchPanelConfig {
    searchValue: string
    regexp: boolean
    caseSensitive: boolean
    mode: SearchPanelViewMode
}

export interface SearchPanelState {
    searchQuery: SearchQuery
    // The input value is usually derived from searchQuery. But we are
    // debouncing updating the searchQuery and without tracking the input value
    // separately user input would be lossing characters and feel laggy.
    inputValue: string
    overrideBrowserSearch: boolean
    matches: SearchMatches
    // Currently selected 1-based match index.
    currentMatchIndex: number | null
    mode: SearchPanelViewMode
}

export interface SearchPanelViewCreationOptions {
    root: HTMLElement
    initialState: SearchPanelState
    onSearch: (search: string) => void
    findNext: () => void
    findPrevious: () => void
    setCaseSensitive: (caseSensitive: boolean) => void
    setRegexp: (regexp: boolean) => void
    setOverrideBrowserSearch: (override: boolean) => void
    close: () => void
}

export interface SearchPanelView {
    input: HTMLInputElement | null
    update(state: SearchPanelState): void
    destroy(): void
}

class SearchPanel implements Panel {
    public dom: HTMLElement
    public top = true

    private state: SearchPanelState
    private panel: SearchPanelView | null = null
    private searchTerm = new Subject<string>()
    private subscriptions = new Subscription()

    constructor(
        private view: EditorView,
        private createPanelView: (options: SearchPanelViewCreationOptions) => SearchPanelView,
        config?: SearchPanelConfig
    ) {
        this.dom = createElement('div', {
            className: classNames('cm-sg-search-container', styles.root),
            id: BLOB_SEARCH_CONTAINER_ID,
            onkeydown: this.onkeydown,
        })

        const searchQuery = getSearchQuery(this.view.state)
        const matches = calculateMatches(searchQuery, view.state.doc)
        this.state = {
            searchQuery: new SearchQuery({
                ...searchQuery,
                caseSensitive: config?.caseSensitive ?? searchQuery.caseSensitive,
                regexp: config?.regexp ?? searchQuery.regexp,
                search: config?.searchValue ?? searchQuery.search,
            }),
            inputValue: config?.searchValue ?? searchQuery.search,
            overrideBrowserSearch: this.view.state.field(overrideBrowserFindInPageShortcut),
            matches,
            currentMatchIndex: getMatchIndexForSelection(matches, view.state.selection.main),
            mode: config?.mode ?? SearchPanelViewMode.FullSearch,
        }

        this.subscriptions.add(
            this.searchTerm
                .pipe(
                    startWith(this.state.searchQuery.search),
                    distinctUntilChanged(),
                    // Immediately update input for fast feedback
                    tap(value => {
                        this.state = { ...this.state, inputValue: value }
                        this.panel?.update(this.state)
                    }),
                    debounceTime(100)
                )
                .subscribe(searchTerm => this.commit({ search: searchTerm }))
        )
    }

    public update(update: ViewUpdate): void {
        let newState = this.state

        const searchQuery = getSearchQuery(update.state)
        const searchQueryChanged = !searchQuery.eq(this.state.searchQuery)
        if (searchQueryChanged) {
            newState = {
                ...newState,
                inputValue: searchQuery.search,
                searchQuery,
                matches: calculateMatches(searchQuery, update.view.state.doc),
            }
        }

        const overrideBrowserSearch = update.state.field(overrideBrowserFindInPageShortcut)
        if (overrideBrowserSearch !== this.state.overrideBrowserSearch) {
            newState = { ...newState, overrideBrowserSearch }
        }

        // It looks like update.SelectionSet is not set when the search query changes
        if (searchQueryChanged || update.selectionSet) {
            newState = {
                ...newState,
                currentMatchIndex: getMatchIndexForSelection(newState.matches, update.view.state.selection.main),
            }
        }

        if (newState !== this.state) {
            this.state = newState
            this.panel?.update(this.state)
        }

        if (
            update.transactions.some(transaction =>
                transaction.effects.some(effect => effect.is(focusSearchInput) && effect.value)
            )
        ) {
            this.panel?.input?.focus()
            this.panel?.input?.select()
        }
    }

    public mount(): void {
        this.panel = this.createPanelView({
            root: this.dom,
            initialState: this.state,
            onSearch: search => this.searchTerm.next(search),
            findNext: this.findNext,
            findPrevious: this.findPrevious,
            setCaseSensitive: caseSensitive => this.commit({ caseSensitive }),
            setRegexp: regexp => this.commit({ regexp }),
            setOverrideBrowserSearch: this.setOverrideBrowserSearch,
            close: () => closeSearchPanel(this.view),
        })
    }

    public destroy(): void {
        this.subscriptions.unsubscribe()
        this.panel?.destroy()
    }

    private setOverrideBrowserSearch = (override: boolean): void =>
        this.view.dispatch({
            effects: setOverrideBrowserFindInPageShortcut.of(override),
        })

    private findNext = (): void => {
        findNext(this.view)
        // Scroll the selection into the middle third of the view
        this.view.dispatch({
            effects: EditorView.scrollIntoView(this.view.state.selection.main.from, {
                y: 'nearest',
                yMargin: this.view.dom.getBoundingClientRect().height / 3,
            }),
        })
    }

    private findPrevious = (): void => {
        findPrevious(this.view)
        // Scroll the selection into the middle third of the view
        this.view.dispatch({
            effects: EditorView.scrollIntoView(this.view.state.selection.main.from, {
                y: 'nearest',
                yMargin: this.view.dom.getBoundingClientRect().height / 3,
            }),
        })
    }

    // Taken from CodeMirror's default search panel implementation. This is
    // necessary so that pressing Meta+F (and other CodeMirror keybindings) will
    // trigger the configured event handlers and not just fall back to the
    // browser's default behavior.
    private onkeydown = (event: KeyboardEvent): void => {
        if (runScopeHandlers(this.view, event, 'search-panel')) {
            event.preventDefault()
        } else if (event.key === 'Enter' && event.target === this.panel?.input) {
            event.preventDefault()
            if (event.shiftKey) {
                this.findPrevious()
            } else {
                this.findNext()
            }
        }
    }

    private commit = ({
        search,
        caseSensitive,
        regexp,
    }: {
        search?: string
        caseSensitive?: boolean
        regexp?: boolean
    }): void => {
        const query = new SearchQuery({
            search: search ?? this.state.searchQuery.search,
            caseSensitive: caseSensitive ?? this.state.searchQuery.caseSensitive,
            regexp: regexp ?? this.state.searchQuery.regexp,
        })

        if (!query.eq(this.state.searchQuery)) {
            let transactionSpec: TransactionSpec = {}
            const effects: StateEffect<any>[] = [setSearchQuery.of(query)]

            if (query.search) {
                // The following code scrolls next match into view if there is no
                // match in the visible viewport. This is done by searching for the
                // text from the currently top visible line and determining whether
                // the next match is in the current viewport

                const { scrollTop } = this.view.scrollDOM

                // Get top visible line. More than half of the line must be visible.
                // We don't use `view.viewportLineBlocks` because that also includes
                // lines that are rendered but not actually visible.
                let topLineBlock = this.view.lineBlockAtHeight(scrollTop)
                if (Math.abs(topLineBlock.bottom - scrollTop) <= topLineBlock.height / 2) {
                    topLineBlock = this.view.lineBlockAtHeight(scrollTop + topLineBlock.height)
                }

                if (query.regexp && !query.valid) {
                    return
                }

                let result = query.getCursor(this.view.state.doc, topLineBlock.from).next()
                if (result.done) {
                    // No match in the remainder of the document, wrap around
                    result = query.getCursor(this.view.state.doc).next()
                }

                if (!result.done) {
                    // Taken from the original `findPrevious` and `findNext` CodeMirror implementation:
                    // https://github.com/codemirror/search/blob/affb772655bab706e08f99bd50a0717bfae795f5/src/search.ts#L385-L416

                    transactionSpec = {
                        selection: { anchor: result.value.from, head: result.value.to },
                        scrollIntoView: true,
                        userEvent: 'select.search',
                    }
                    effects.push(announceMatch(this.view, result.value))
                }
                // Search term is not in the document, nothing to do
            }

            this.view.dispatch({
                ...transactionSpec,
                effects,
            })
        }
    }
}

function calculateMatches(query: SearchQuery, document: CodeMirrorText): SearchMatches {
    const newSearchMatches: SearchMatches = new Map()

    if (!query.valid) {
        return newSearchMatches
    }

    let index = 1
    const matches = query.getCursor(document)
    let result = matches.next()

    while (!result.done) {
        if (result.value.from !== result.value.to) {
            newSearchMatches.set(result.value.from, index++)
        }

        result = matches.next()
    }

    return newSearchMatches
}

function getMatchIndexForSelection(matches: SearchMatches, range: SelectionRange): number | null {
    return range.empty ? null : matches.get(range.from) ?? null
}

// Announce the current match to screen readers.
// Taken from original the CodeMirror implementation:
// https://github.com/codemirror/search/blob/affb772655bab706e08f99bd50a0717bfae795f5/src/search.ts#L694-L717
const announceMargin = 30
const breakRegex = /[\s!,.:;?]/
function announceMatch(view: EditorView, { from, to }: { from: number; to: number }): StateEffect<string> {
    const line = view.state.doc.lineAt(from)
    const lineEnd = view.state.doc.lineAt(to).to
    const start = Math.max(line.from, from - announceMargin)
    const end = Math.min(lineEnd, to + announceMargin)
    let text = view.state.sliceDoc(start, end)
    if (start !== line.from) {
        for (let index = 0; index < announceMargin; index++) {
            if (!breakRegex.test(text[index + 1]) && breakRegex.test(text[index])) {
                text = text.slice(index)
                break
            }
        }
    }
    if (end !== lineEnd) {
        for (let index = text.length - 1; index > text.length - announceMargin; index--) {
            if (!breakRegex.test(text[index - 1]) && breakRegex.test(text[index])) {
                text = text.slice(0, index)
                break
            }
        }
    }

    return EditorView.announce.of(
        `${view.state.phrase('current match')}. ${text} ${view.state.phrase('on line')} ${line.number}.`
    )
}

const theme = EditorView.theme({
    '.cm-sg-search-container': {
        backgroundColor: 'var(--code-bg)',
        padding: '0.5rem 0.5rem',
    },
    '.cm-sg-search-input': {
        borderRadius: 'var(--border-radius)',
        border: '1px solid var(--input-border-color)',

        '&:focus-within': {
            boxShadow: 'var(--input-focus-box-shadow)',
        },

        '& input': {
            borderColor: 'transparent',
            '&:focus': {
                boxShadow: 'none',
            },
        },
    },
    '.search-container > input.form-control': {
        width: '15rem',
        height: '1.0rem',
    },
    '.cm-searchMatch': {
        backgroundColor: 'var(--mark-bg)',
    },
    '.cm-searchMatch-selected': {
        backgroundColor: 'var(--oc-orange-3)',
    },
    '.cm-sg-search-info': {
        color: 'var(--body-color)',
    },
    '.cm-search-results': {
        color: 'var(--gray-06)',
        fontFamily: 'var(--font-family-base)',
        marginLeft: '2rem',
    },
    '.cm-search-toggle': {
        color: 'var(--gray-06)',
    },
})

interface SearchConfig {
    overrideBrowserFindInPageShortcut: boolean
    onOverrideBrowserFindInPageToggle: (enabled: boolean) => void
    createPanel: (options: SearchPanelViewCreationOptions) => SearchPanelView
    initialState?: SearchPanelConfig
}

const [overrideBrowserFindInPageShortcut, , setOverrideBrowserFindInPageShortcut] = createUpdateableField(true)

export function search(config: SearchConfig): Extension {
    const keymapCompartment = new Compartment()

    function getKeyBindings(override: boolean): readonly KeyBinding[] {
        if (override) {
            return searchKeymap.map(binding =>
                binding.key === 'Mod-f'
                    ? {
                          ...binding,
                          run: view => {
                              // By default pressing Mod+f when the search input is already focused won't select
                              // the input value, unlike browser's built-in search feature.
                              // We are overwriting the keybinding here to ensure that the input value is always
                              // selected.
                              const result = binding.run?.(view)
                              if (result) {
                                  view.dispatch({ effects: focusSearchInput.of(true) })
                                  return true
                              }
                              return false
                          },
                      }
                    : binding
            )
        }
        return searchKeymap.filter(binding => binding.key !== 'Mod-f' && binding.key !== 'Escape')
    }

    return [
        overrideBrowserFindInPageShortcut.init(() => config.overrideBrowserFindInPageShortcut),
        EditorView.updateListener.of(update => {
            const override = update.state.field(overrideBrowserFindInPageShortcut)
            if (update.startState.field(overrideBrowserFindInPageShortcut) !== override) {
                config.onOverrideBrowserFindInPageToggle(override)
                update.view.dispatch({ effects: keymapCompartment.reconfigure(keymap.of(getKeyBindings(override))) })
            }
        }),
        theme,
        keymapCompartment.of(keymap.of(getKeyBindings(config.overrideBrowserFindInPageShortcut))),
        codemirrorSearch({
            createPanel: view => new SearchPanel(view, config.createPanel, config.initialState),
        }),
        ViewPlugin.define(view => {
            // If we have some initial state for the search bar this means we want
            // to render it by default
            if (!config.overrideBrowserFindInPageShortcut || config.initialState) {
                window.requestAnimationFrame(() => openSearchPanel(view))
            }
            return {}
        }),
    ]
}
