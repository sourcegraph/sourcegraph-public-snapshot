/**
 * This extension exends CodeMirro's own search extension with a custom search
 * UI.
 */

import {
    combineConfig,
    Compartment,
    EditorState,
    Extension,
    Facet,
    RangeSetBuilder,
    StateEffect,
    StateField,
    Text,
} from '@codemirror/state'
import {
    Command,
    Decoration,
    DecorationSet,
    EditorView,
    getPanel,
    KeyBinding,
    keymap,
    Panel,
    PanelConstructor,
    runScopeHandlers,
    showPanel,
    ViewPlugin,
    ViewUpdate,
} from '@codemirror/view'
import { mdiChevronDown, mdiChevronUp, mdiFormatLetterCase, mdiInformationOutline, mdiRegex } from '@mdi/js'
import { History } from 'history'
import { createRoot, Root } from 'react-dom/client'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { QueryInputToggle } from '@sourcegraph/search-ui'
import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { shortcutDisplayName } from '@sourcegraph/shared/src/keyboardShortcuts'
import { Button, Icon, Input, Label, Tooltip } from '@sourcegraph/wildcard'

import { Keybindings } from '../../../components/KeyboardShortcutsHelp/KeyboardShortcutsHelp'
import { createElement } from '../../../util/dom'

import { Container } from './react-interop'

import { blobPropsFacet } from '.'
import escapeRegExp from 'lodash/escapeRegExp'

interface RegexpSearchConfig {
    createPanel: PanelConstructor
}

interface SearchQueryConfig {
    /**
     * The query input.
     */
    search: string
    caseSensitive: boolean
    regexp: boolean
}

const EMPTY_REGEXP = new RegExp('')

class SearchQuery {
    /**
     * Query is non empty and, if regexp, a valid regular expression.
     */
    public valid: boolean

    private matchRegexp: RegExp = EMPTY_REGEXP
    public matches: { from: number; to: number }[] = []
    public current: number = 0

    constructor(private config: SearchQueryConfig) {
        this.valid = config.search.length > 0

        let flags = 'g'
        if (!config.caseSensitive) {
            flags += 'i'
        }

        if (this.valid) {
            if (config.regexp) {
                try {
                    this.matchRegexp = new RegExp(config.search, flags)
                } catch (error) {
                    this.valid = false
                }
            } else {
                this.matchRegexp = new RegExp(escapeRegExp(config.search), flags)
            }
        }
    }

    get regexp() {
        return this.config.regexp
    }

    get caseSensitive() {
        return this.config.caseSensitive
    }

    get search() {
        return this.config.search
    }

    eq(other: SearchQuery): boolean {
        return (
            this.config.search === other.config.search &&
            this.config.caseSensitive === other.config.caseSensitive &&
            this.config.regexp === other.config.regexp
        )
    }

    exec(input: string): { from: number; to: number }[] {
        const matches: { from: number; to: number }[] = []

        let match: RegExpMatchArray | null
        while ((match = this.matchRegexp.exec(input))) {
            if (match.index !== undefined) {
                matches.push({ from: match.index, to: match.index + match[0].length })
            }
        }
        return matches
    }
}

class SearchMatch {
    public matches: { from: number; to: number }[] = []
    public current: number = 0

    constructor(query: SearchQuery, text: Text) {
        if (query.valid) {
            this.matches = query.exec(text.sliceString(0))
        }
    }

    public get count() {
        return this.matches.length
    }
}

const searchConfig = Facet.define<RegexpSearchConfig, Required<RegexpSearchConfig>>({
    combine(configs) {
        return combineConfig(configs, { createPanel: view => new SearchPanel(view) })
    },
})

const createSearchPanel: PanelConstructor = view => view.state.facet(searchConfig).createPanel(view)

function getSearchQuery(state: EditorState): SearchQuery {
    return state.field(searchState).query
}

const setSearchQuery = StateEffect.define<SearchQuery>()
const togglePanel = StateEffect.define<boolean>()
const setSelectedMatch = StateEffect.define<number>()

interface SearchState {
    query: SearchQuery
    matches: SearchMatch | null
    selectedMatch: number
    panel: PanelConstructor | null
}

function defaultQuery(state: EditorState, fallback?: SearchQuery): SearchQuery {
    const selection = state.selection.main
    const selectedText =
        selection.empty || selection.to > selection.from + 100 ? '' : state.sliceDoc(selection.from, selection.to)
    return new SearchQuery({
        search: selectedText,
        caseSensitive: fallback?.caseSensitive || false,
        regexp: false,
    })
}

const matchDecoration = Decoration.mark({ class: 'cm-searchMatch' })
const selectedMatchDecoration = Decoration.mark({ class: 'cm-searchMatch cm-searchMatch-selected' })

const searchState = StateField.define<SearchState>({
    create(state) {
        return { query: defaultQuery(state), matches: null, selectedMatch: 0, panel: null }
    },

    update(value, transaction) {
        for (const effect of transaction.effects) {
            if (effect.is(setSearchQuery)) {
                const query = effect.value
                value = { ...value, query, matches: new SearchMatch(query, transaction.newDoc) }
            } else if (effect.is(togglePanel)) {
                value = { ...value, panel: createSearchPanel }
            } else if (effect.is(setSelectedMatch)) {
                value = { ...value, selectedMatch: effect.value }
            }
        }
        return value
    },

    provide(field) {
        return [showPanel.from(field, value => value.panel)]
    },
})

const searchHighlighter = ViewPlugin.fromClass(
    class {
        decorations: DecorationSet

        constructor(readonly view: EditorView) {
            this.decorations = this.highlight(view.state.field(searchState))
        }

        update(update: ViewUpdate): void {
            const state = update.state.field(searchState)
            if (
                state !== update.startState.field(searchState) ||
                update.docChanged ||
                update.selectionSet ||
                update.viewportChanged
            ) {
                this.decorations = this.highlight(state)
            }
        }

        highlight({ panel, matches, selectedMatch }: SearchState): DecorationSet {
            if (!panel || !matches) {
                return Decoration.none
            }

            const builder = new RangeSetBuilder<Decoration>()
            let index = 0
            for (const { from, to } of this.view.visibleRanges) {
                while (index < matches.matches.length) {
                    const selected = index === selectedMatch
                    const match = matches.matches[index++]

                    if (from > match.to) {
                        // Skip matches before the visible range
                        continue
                    }

                    if (match.from > to) {
                        // Continue with the next visible range
                        break
                    }

                    builder.add(match.from, match.to, selected ? selectedMatchDecoration : matchDecoration)
                }
            }
            return builder.finish()
        }
    },
    { decorations: plugin => plugin.decorations }
)

export const openSearchPanel: Command = view => {
    const state = view.state.field(searchState, false)
    if (state && state.panel) {
        const panel = getPanel(view, createSearchPanel)
        if (!panel) return false
        const searchInput = panel.dom.querySelector('[main-field]') as HTMLInputElement | null
        if (searchInput && searchInput != view.root.activeElement) {
            const query = defaultQuery(view.state, state.query)
            if (query.valid) view.dispatch({ effects: setSearchQuery.of(query) })
            searchInput.focus()
            searchInput.select()
        }
    } else {
        view.dispatch({ effects: [togglePanel.of(true), setSearchQuery.of(defaultQuery(view.state, state?.query))] })
    }
    return true
}

const closeSearchPanel: Command = view => {
    let state = view.state.field(searchState, false)
    if (!state || !state.panel) return false
    let panel = getPanel(view, createSearchPanel)
    if (panel && panel.dom.contains(view.root.activeElement)) view.focus()
    view.dispatch({ effects: togglePanel.of(false) })
    return true
}

const findNext: Command = view => {
    const state = view.state.field(searchState)
    if (state.query.valid && state.matches) {
        const matchCount = state.matches.count
        const nextIndex = (state.selectedMatch + 1) % matchCount
        const nextMatch = state.matches.matches[nextIndex]

        view.dispatch({
            selection: { anchor: nextMatch.from, head: nextMatch.to },
            scrollIntoView: true,
            effects: setSelectedMatch.of(nextIndex),
        })
    }
    return true
}

const findPrevious: Command = view => {
    const state = view.state.field(searchState)
    if (state.query.valid && state.matches) {
        const matchCount = state.matches.count
        const previousIndex = state.selectedMatch === 0 ? matchCount - 1 : state.selectedMatch - 1
        const previousMatch = state.matches.matches[previousIndex]

        view.dispatch({
            selection: { anchor: previousMatch.from, head: previousMatch.to },
            scrollIntoView: true,
            effects: setSelectedMatch.of(previousIndex),
        })
    }
    return true
}

function regexpSearch(config: RegexpSearchConfig): Extension {
    return [searchConfig.of(config), searchState, searchHighlighter]
}

const searchKeybinding = <Keybindings keybindings={[{ held: ['Mod'], ordered: ['F'] }]} />

const platformKeycombo = shortcutDisplayName('Mod+F')
const tooltipContent = `When enabled, ${platformKeycombo} searches the file only. Disable to search the page, and press ${platformKeycombo} for changes to apply.`
const searchKeybindingTooltip = (
    <Tooltip content={tooltipContent}>
        <Icon
            className="cm-sg-search-info ml-1 align-textbottom"
            svgPath={mdiInformationOutline}
            aria-label="Search keybinding information"
        />
    </Tooltip>
)

class SearchPanel implements Panel {
    public dom: HTMLElement
    public top = true

    private state: {
        searchQuery: SearchQuery
        overrideBrowserSearch: boolean
        history: History
        selectedMatch: number
        matchCount: number
    }
    private root: Root | null = null
    private input: HTMLInputElement | null = null

    constructor(private view: EditorView) {
        this.dom = createElement('div', {
            className: 'cm-sg-search-container d-flex align-items-center',
            onkeydown: this.onkeydown,
        })

        const { query, selectedMatch, matches } = this.view.state.field(searchState)

        this.state = {
            searchQuery: query,
            selectedMatch,
            matchCount: matches?.count ?? 0,
            overrideBrowserSearch: this.view.state.field(overrideBrowserFindInPageShortcut),
            history: this.view.state.facet(blobPropsFacet).history,
        }
    }

    public update(update: ViewUpdate): void {
        let newState = this.state

        const { query, selectedMatch, matches } = this.view.state.field(searchState)
        const matchCount = matches?.count ?? 0

        if (!query.eq(this.state.searchQuery)) {
            newState = { ...newState, searchQuery: query }
        }

        if (selectedMatch !== this.state.selectedMatch) {
            newState = { ...newState, selectedMatch }
        }

        if (matchCount !== this.state.matchCount) {
            newState = { ...newState, matchCount }
        }

        const overrideBrowserSearch = update.state.field(overrideBrowserFindInPageShortcut)
        if (overrideBrowserSearch !== this.state.overrideBrowserSearch) {
            newState = { ...newState, overrideBrowserSearch }
        }

        const history = update.state.facet(blobPropsFacet).history
        if (history !== this.state.history) {
            newState = { ...newState, history }
        }

        if (newState !== this.state) {
            this.state = newState
            this.render(newState)
        }
    }

    public mount(): void {
        this.render(this.state)
    }

    private render({
        searchQuery,
        selectedMatch,
        matchCount,
        overrideBrowserSearch,
        history,
    }: SearchPanel['state']): void {
        if (!this.root) {
            this.root = createRoot(this.dom)
        }

        this.root.render(
            <Container
                history={history}
                onMount={() => {
                    this.input?.focus()
                    this.input?.select()
                }}
            >
                <div className="cm-sg-search-input d-flex align-items-center pr-2 mr-2">
                    <Input
                        ref={element => (this.input = element)}
                        name="search"
                        variant="small"
                        placeholder="Find..."
                        value={searchQuery.search}
                        autoComplete="off"
                        onChange={() => this.commit()}
                        onKeyUp={() => this.commit()}
                        main-field="true"
                        role="search"
                    />
                    <QueryInputToggle
                        isActive={searchQuery.caseSensitive}
                        onToggle={() => this.commit({ caseSensitive: !searchQuery.caseSensitive })}
                        iconSvgPath={mdiFormatLetterCase}
                        title="Case sensitivity"
                        className="test-blob-view-search-case-sensitive"
                    />
                    <QueryInputToggle
                        isActive={searchQuery.regexp}
                        onToggle={() => this.commit({ regexp: !searchQuery.regexp })}
                        iconSvgPath={mdiRegex}
                        title="Regular expression"
                        className="test-blob-view-search-regexp"
                    />
                </div>
                <Button
                    className="mr-2"
                    type="button"
                    size="sm"
                    outline={true}
                    variant="secondary"
                    onClick={() => findPrevious(this.view)}
                    data-testid="blob-view-search-previous"
                >
                    <Icon svgPath={mdiChevronUp} aria-hidden={true} />
                    Previous
                </Button>

                <Button
                    className="mr-3"
                    type="button"
                    size="sm"
                    outline={true}
                    variant="secondary"
                    onClick={() => findNext(this.view)}
                    data-testid="blob-view-search-next"
                >
                    <Icon svgPath={mdiChevronDown} aria-hidden={true} />
                    Next
                </Button>

                <div>
                    <Label className="mb-0">
                        <Toggle
                            className="mr-1 align-text-bottom"
                            value={overrideBrowserSearch}
                            onToggle={this.setOverrideBrowserSearch}
                        />
                        {searchKeybinding} searches file
                    </Label>
                    {searchKeybindingTooltip}
                </div>
                {searchQuery.valid && (
                    <span className="ml-auto">
                        {selectedMatch + (matchCount > 0 ? 1 : 0)} / {matchCount}
                    </span>
                )}
            </Container>
        )
    }

    private setOverrideBrowserSearch = (override: boolean): void =>
        this.view.dispatch({
            effects: setOverrideBrowserFindInPageShortcut.of(override),
        })

    // Taken from CodeMirror's default serach panel implementation. This is
    // necessary so that pressing Meta+F (and other CodeMirror keybindings) will
    // trigger the configured event handlers and not just fall back to the
    // browser's default behavior.
    private onkeydown = (event: KeyboardEvent): void => {
        if (runScopeHandlers(this.view, event, 'search-panel')) {
            event.preventDefault()
        } else if (event.keyCode === 13 && event.target === this.input) {
            event.preventDefault()
            if (event.shiftKey) {
                findPrevious(this.view)
            } else {
                findNext(this.view)
            }
        }
    }

    private commit = ({ caseSensitive, regexp }: { caseSensitive?: boolean; regexp?: boolean } = {}): void => {
        const query = new SearchQuery({
            search: this.input?.value ?? '',
            caseSensitive: caseSensitive ?? this.state.searchQuery.caseSensitive,
            regexp: regexp ?? this.state.searchQuery.regexp,
        })

        if (!query.eq(this.state.searchQuery)) {
            this.view.dispatch({ effects: setSearchQuery.of(query) })

            /*
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

                let result = query.getCursor(this.view.state.doc, topLineBlock.from).next()
                if (result.done) {
                    // No match in the remainder of the document, wrap around
                    result = query.getCursor(this.view.state.doc).next()
                    if (result.done) {
                        // Search term is not in the document, nothing to do
                        return
                    }
                }

                const matchLineBlock = this.view.lineBlockAt(result.value.from)
                const matchLineCenter = matchLineBlock.top + matchLineBlock.height / 2

                if (matchLineCenter < scrollTop || matchLineCenter > scrollTop + this.view.scrollDOM.clientHeight) {
                    this.view.dispatch({
                        effects: EditorView.scrollIntoView(result.value.from, {
                            y: 'center',
                        }),
                    })
                }
            }
             */
        }
    }
}

interface SearchConfig {
    overrideBrowserFindInPageShortcut: boolean
    onOverrideBrowserFindInPageToggle: (enabled: boolean) => void
}

const [overrideBrowserFindInPageShortcut, , setOverrideBrowserFindInPageShortcut] = createUpdateableField(true)

const searchKeymap: KeyBinding[] = [
    {
        key: 'Mod-f',
        run: openSearchPanel,
    },
    {
        key: 'Esc',
        run: closeSearchPanel,
    },
    {
        key: 'Mod-g',
        run: findNext,
        shift: findPrevious,
    },
]

export function search(config: SearchConfig): Extension {
    const keymapCompartment = new Compartment()

    function getKeyBindings(override: boolean): readonly KeyBinding[] {
        if (override) {
            return searchKeymap
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
        EditorView.theme({
            '.cm-sg-search-container': {
                backgroundColor: 'var(--code-bg)',
                padding: '0.375rem 1rem',
            },
            '.cm-sg-search-input': {
                borderRadius: 'var(--border-radius)',
                border: '1px solid var(--input-border-color)',
            },
            '.cm-sg-search-input:focus-within': {
                borderColor: 'var(--inpt-focus-border-color)',
                boxShadow: 'var(--input-focus-box-shadow)',
            },
            '.cm-sg-search-input input': {
                borderColor: 'transparent',
            },
            '.cm-sg-search-input input:focus': {
                boxShadow: 'none',
            },
            '.search-container > input.form-control': {
                width: '15rem',
            },
            '.cm-searchMatch': {
                backgroundColor: 'var(--mark-bg)',
            },
            '.cm-searchMatch-selected': {
                backgroundColor: 'var(--oc-orange-3)',
            },
            '.cm-sg-search-info': {
                color: 'var(--gray-06)',
            },
        }),
        keymapCompartment.of(keymap.of(getKeyBindings(config.overrideBrowserFindInPageShortcut))),
        regexpSearch({
            createPanel: view => new SearchPanel(view),
        }),
        ViewPlugin.define(view => {
            if (!config.overrideBrowserFindInPageShortcut) {
                window.requestAnimationFrame(() => openSearchPanel(view))
            }
            return {}
        }),
    ]
}
