/**
 * This extension exends CodeMirro's own search extension with a custom search
 * UI.
 */

import {
    findNext,
    findPrevious,
    getSearchQuery,
    openSearchPanel,
    search as codemirrorSearch,
    searchKeymap,
    SearchQuery,
    setSearchQuery,
} from '@codemirror/search'
import { Compartment, Extension } from '@codemirror/state'
import { EditorView, KeyBinding, keymap, Panel, runScopeHandlers, ViewPlugin, ViewUpdate } from '@codemirror/view'
import { mdiChevronDown, mdiChevronUp, mdiFormatLetterCase, mdiInformationOutline, mdiRegex } from '@mdi/js'
import { History } from 'history'
import { createRoot, Root } from 'react-dom/client'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { QueryInputToggle } from '@sourcegraph/search-ui'
import { createUpdateableField } from '@sourcegraph/shared/src/components/CodeMirrorEditor'
import { Button, Icon, Input, Label, Tooltip } from '@sourcegraph/wildcard'

import { Keybindings, renderShortcutKey } from '../../../components/KeyboardShortcutsHelp/KeyboardShortcutsHelp'
import { createElement } from '../../../util/dom'

import { Container } from './react-interop'

import { blobPropsFacet } from '.'

const searchKeybinding = <Keybindings keybindings={[{ held: ['Mod'], ordered: ['F'] }]} />

const platformKeycombo = renderShortcutKey('Mod') + '+F'
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

    private state: { searchQuery: SearchQuery; overrideBrowserSearch: boolean; history: History }
    private root: Root | null = null
    private input: HTMLInputElement | null = null

    constructor(private view: EditorView) {
        this.dom = createElement('div', {
            className: 'cm-sg-search-container d-flex align-items-center',
            onkeydown: this.onkeydown,
        })

        this.state = {
            searchQuery: getSearchQuery(this.view.state),
            overrideBrowserSearch: this.view.state.field(overrideBrowserFindInPageShortcut),
            history: this.view.state.facet(blobPropsFacet).history,
        }
    }

    public update(update: ViewUpdate): void {
        let newState = this.state

        const searchQuery = getSearchQuery(update.state)
        if (!searchQuery.eq(this.state.searchQuery)) {
            newState = { ...newState, searchQuery }
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
        overrideBrowserSearch,
        history,
    }: {
        searchQuery: SearchQuery
        overrideBrowserSearch: boolean
        history: History
    }): void {
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
        }
    }
}

interface SearchConfig {
    overrideBrowserFindInPageShortcut: boolean
    onOverrideBrowserFindInPageToggle: (enabled: boolean) => void
}

const [overrideBrowserFindInPageShortcut, , setOverrideBrowserFindInPageShortcut] = createUpdateableField(true)

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
        codemirrorSearch({
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
