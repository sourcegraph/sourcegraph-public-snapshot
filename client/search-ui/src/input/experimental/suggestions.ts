import {
    EditorSelection,
    EditorState,
    Extension,
    Facet,
    Prec,
    StateEffect,
    StateField,
    Transaction,
} from '@codemirror/state'
import { Command as CodeMirrorCommand, EditorView, keymap, ViewPlugin, ViewUpdate } from '@codemirror/view'
import { History } from 'history'
import { SvelteComponentTyped } from 'svelte'

import Suggestions from './Suggestions.svelte'
export { default as FilterSuggestion } from './FilterSuggestion.svelte'
export { default as SearchQueryOption } from './SearchQueryOption.svelte'

// Temporary solution to make some editor settings available to other extensions
interface EditorConfig {
    onSubmit: () => void
}
export const editorConfigFacet = Facet.define<EditorConfig, EditorConfig>({
    combine(configs) {
        return configs[0] ?? { onSubmit: () => {} }
    },
})
export function getEditorConfig(state: EditorState): EditorConfig {
    return state.facet(editorConfigFacet)
}

/**
 * A source for completion/suggestion results
 */
export type Source = (state: EditorState, position: number) => SuggestionResult

export interface SuggestionResult {
    /**
     * Initial/synchronous result.
     */
    result: Group[]
    /**
     * Function to be called to load additional results if necessary.
     */
    next?: () => Promise<SuggestionResult>
    /**
     * Determines whether this result is invalidated by the new editor state.
     */
    valid?: (state: EditorState, position: number) => boolean
}

export type CustomRenderer = typeof SvelteComponentTyped<{ option: Option }>

export interface Command {
    type: 'command'
    value: string
    apply: (view: EditorView) => void
    matches?: Set<number>
    // svg path
    icon?: string
    render?: CustomRenderer
    description?: string
    note?: string
}
export interface Target {
    type: 'target'
    value: string
    url: string
    matches?: Set<number>
    // svg path
    icon?: string
    render?: CustomRenderer
    description?: string
}
export interface Completion {
    type: 'completion'
    from: number
    to?: number
    value: string
    insertValue?: string
    matches?: Set<number>
    // svg path
    icon?: string
    render?: CustomRenderer
    description?: string
}
export type Option = Command | Target | Completion

export interface Group {
    title: string
    entries: Option[]
}

class SuggestionView {
    private instance: Suggestions
    private root: HTMLElement

    constructor(id: string, public view: EditorView, public parent: HTMLDivElement) {
        const state = view.state.field(suggestionsStateField)
        this.root = document.createElement('div')
        this.instance = new Suggestions({
            target: parent,
            props: {
                id,
                results: state.result.groups,
                activeRowIndex: state.selectedOption,
                open: state.open,
            },
        })
        this.instance.$on('select', event => {
            applyOption(this.view, event.detail)
            // Query input looses focus when option is selected via
            // mousedown/click. This is a necessary hack to re-focus the query
            // input.
            window.requestAnimationFrame(() => view.contentDOM.focus())
        })
        this.view.dom.append(this.root)
    }

    public update(update: ViewUpdate): void {
        const state = update.state.field(suggestionsStateField)

        if (state !== update.startState.field(suggestionsStateField)) {
            this.updateResults(state)
        }
    }

    private updateResults(state: SuggestionsState): void {
        this.instance.$set({ results: state.result.groups, activeRowIndex: state.selectedOption, open: state.open })
    }

    public destroy(): void {
        this.instance.$destroy()
        this.root.remove()
    }
}

const completionPlugin = ViewPlugin.fromClass(
    class {
        private running: RunningQuery | null = null

        constructor(public readonly view: EditorView) {
            this.startQuery(view.state.field(suggestionsStateField).source)
        }

        public update(update: ViewUpdate): void {
            if (update.view.hasFocus) {
                this.startQuery(update.state.field(suggestionsStateField).source)
            }
        }

        private startQuery(source: RegisteredSource): void {
            if (
                source.state === RegisteredSourceState.Pending &&
                (!this.running || this.running.timestamp !== source.timestamp)
            ) {
                const query = (this.running = new RunningQuery(source))
                query.source
                    .run()
                    ?.then(result => {
                        if (this.running === query) {
                            this.view.dispatch({ effects: updateResultEffect.of({ source, result }) })
                        }
                    })
                    .catch(() => {})
            } else if (source.state === RegisteredSourceState.Inactive) {
                this.running = null
            }
        }
    }
)

class RunningQuery {
    constructor(public readonly source: RegisteredSource) {}

    public get timestamp(): number {
        return this.source.timestamp
    }
}

/**
 * Wrapper class to make operating on groups of options easier.
 */
class Result {
    private entries: Option[]

    constructor(
        public readonly groups: Group[],
        public valid: (state: EditorState, position: number) => boolean = () => false
    ) {
        this.entries = groups.flatMap(group => group.entries)
    }

    // eslint-disable-next-line id-length
    public at(index: number): Option | undefined {
        return this.entries[index]
    }

    public groupRowIndex(index: number): [number, number] | undefined {
        const option = this.entries[index]

        if (!option) {
            return undefined
        }

        for (let groupIndex = 0; groupIndex < this.groups.length; groupIndex++) {
            const options = this.groups[groupIndex].entries
            for (let rowIndex = 0; rowIndex < options.length; rowIndex++) {
                if (options[rowIndex] === option) {
                    return [groupIndex, rowIndex]
                }
            }
        }

        return undefined
    }

    public empty(): boolean {
        return this.length === 0
    }

    public get length(): number {
        return this.entries.length
    }
}

const emptyResult = new Result([])

enum RegisteredSourceState {
    Inactive,
    Pending,
    Complete,
}

/**
 * Internal wrapper around a provided source. Keeps track of the sources state
 * and results.
 */
class RegisteredSource {
    public timestamp: number

    constructor(
        public readonly source: Source,
        public readonly state: RegisteredSourceState,
        public readonly result: Result,
        private readonly next?: () => Promise<SuggestionResult>
    ) {
        switch (state) {
            case RegisteredSourceState.Pending:
                this.timestamp = Date.now()
                break
            default:
                this.timestamp = -1
        }
    }

    public update(transaction: Transaction): RegisteredSource {
        if (isUserInput(transaction)) {
            return this.query(transaction.state)
        }

        if (transaction.selection) {
            if (this.result.valid(transaction.state, transaction.newSelection.main.head)) {
                return this
            }
            return this.query(transaction.state)
        }

        // Handles "external" changes to the query input
        if (transaction.docChanged) {
            return new RegisteredSource(this.source, RegisteredSourceState.Inactive, emptyResult)
        }

        for (const effect of transaction.effects) {
            if (
                effect.is(updateResultEffect) &&
                effect.value.source.source === this.source &&
                this.state === RegisteredSourceState.Pending
            ) {
                const { result } = effect.value
                return new RegisteredSource(
                    this.source,
                    result.next ? RegisteredSourceState.Pending : RegisteredSourceState.Complete,
                    new Result(result.result, result.valid),
                    result.next
                )
            }

            if (effect.is(startCompletion)) {
                return this.query(transaction.state)
            }
        }

        return this
    }

    private query(state: EditorState): RegisteredSource {
        const result = this.source(state, state.selection.main.head)
        const nextState = result.next ? RegisteredSourceState.Pending : RegisteredSourceState.Complete
        return new RegisteredSource(this.source, nextState, new Result(result.result, result.valid), result.next)
    }

    public run(): Promise<SuggestionResult> | null {
        return this.next?.() ?? null
    }

    public get inactive(): boolean {
        return this.state === RegisteredSourceState.Inactive
    }
}

/**
 * Main suggestions state. Mangages of data source and selected option.
 */
class SuggestionsState {
    constructor(
        public readonly source: RegisteredSource,
        public readonly open: boolean,
        public readonly selectedOption: number
    ) {}

    public update(transaction: Transaction): SuggestionsState {
        // Aliasing makes it easier to update the state
        // eslint-disable-next-line @typescript-eslint/no-this-alias,unicorn/no-this-assignment
        let state: SuggestionsState = this

        const source = transaction.state.facet(suggestionSource)
        let registeredSource =
            source === state.source.source
                ? state.source
                : new RegisteredSource(source, RegisteredSourceState.Inactive, emptyResult)
        registeredSource = registeredSource.update(transaction)
        if (registeredSource !== state.source) {
            state = new SuggestionsState(
                registeredSource,
                !registeredSource.inactive,
                state.source.state === RegisteredSourceState.Inactive ||
                state.source.state === RegisteredSourceState.Complete
                    ? 0
                    : state.selectedOption
            )
        }

        if (state.selectedOption > -1 && transaction.newDoc.length === 0) {
            state = new SuggestionsState(state.source, !state.source.inactive, -1)
        }

        for (const effect of transaction.effects) {
            if (effect.is(setSelectedEffect)) {
                state = new SuggestionsState(state.source, state.open, effect.value)
            }
            if (effect.is(hideCompletion)) {
                state = new SuggestionsState(state.source, false, state.selectedOption)
            }
        }

        return state
    }

    public get result(): Result {
        return this.source.result
    }
}

function isUserInput(transaction: Transaction): boolean {
    return transaction.isUserEvent('input.type') || transaction.isUserEvent('delete.backward')
}

interface Config {
    id: string
    history?: History
}

const suggestionsConfig = Facet.define<Config, Config>({
    combine(configs) {
        return configs[0] ?? {}
    },
})

const setSelectedEffect = StateEffect.define<number>()
const startCompletion = StateEffect.define<void>()
const hideCompletion = StateEffect.define<void>()
const updateResultEffect = StateEffect.define<{ source: RegisteredSource; result: SuggestionResult }>()
const suggestionsStateField = StateField.define<SuggestionsState>({
    create() {
        return new SuggestionsState(
            new RegisteredSource(() => ({ result: [] }), RegisteredSourceState.Inactive, emptyResult),
            false,
            -1
        )
    },

    update(state, transaction) {
        return state.update(transaction)
    },

    provide(field) {
        return EditorView.contentAttributes.compute([field, suggestionsConfig], state => {
            const id = state.facet(suggestionsConfig).id
            const suggestionState = state.field(field)
            const groupRowIndex = suggestionState.result.groupRowIndex(suggestionState.selectedOption)
            return {
                'aria-expanded': suggestionState.result.empty() ? 'false' : 'true',
                'aria-activedescendant': groupRowIndex ? `${id}-${groupRowIndex[0]}x${groupRowIndex[1]}` : '',
            }
        })
    },
})

function moveSelection(direction: 'forward' | 'backward'): CodeMirrorCommand {
    const forward = direction === 'forward'
    return view => {
        const state = view.state.field(suggestionsStateField, false)
        if (!state?.open || state.result.empty()) {
            return false
        }
        const { length } = state?.result
        let selected = state.selectedOption > -1 ? state.selectedOption + (forward ? 1 : -1) : forward ? 0 : length - 1
        if (selected < 0) {
            selected = length - 1
        } else if (selected >= length) {
            selected = 0
        }
        view.dispatch({ effects: setSelectedEffect.of(selected) })
        return true
    }
}

function applyOption(view: EditorView, option: Option): void {
    switch (option.type) {
        case 'completion':
            {
                const text = option.insertValue ?? option.value
                view.dispatch({
                    ...view.state.changeByRange(range => {
                        if (range === view.state.selection.main) {
                            return {
                                changes: {
                                    from: option.from,
                                    to: option.to ?? view.state.selection.main.head,
                                    insert: text,
                                },
                                range: EditorSelection.cursor(option.from + text.length),
                            }
                        }
                        return { range }
                    }),
                })
            }
            break
        case 'command':
            option.apply(view)
            break
        case 'target':
            {
                const history = view.state.facet(suggestionsConfig).history

                if (history) {
                    history.push(option.url)
                }
            }
            break
    }
}

export const suggestionSource = Facet.define<Source, Source>({
    combine(sources) {
        return sources[0] || (() => {})
    },
    enables: [
        completionPlugin,
        suggestionsStateField,
        EditorView.updateListener.of(update => {
            if (
                update.focusChanged &&
                update.view.hasFocus &&
                update.view.state.field(suggestionsStateField).result.empty()
            ) {
                update.view.dispatch({ effects: startCompletion.of() })
            }
        }),
        Prec.highest(
            keymap.of([
                {
                    key: 'ArrowDown',
                    run: moveSelection('forward'),
                },
                {
                    key: 'ArrowUp',
                    run: moveSelection('backward'),
                },
                {
                    key: 'Mod-Space',
                    run(view) {
                        view.dispatch({ effects: startCompletion.of() })
                        return true
                    },
                },
                {
                    key: 'Enter',
                    run(view) {
                        const state = view.state.field(suggestionsStateField)
                        const option = state.result.at(state.selectedOption)
                        if (!state.open || !option) {
                            return false
                        }
                        applyOption(view, option)
                        return true
                    },
                },
                {
                    key: 'Tab',
                    run(view) {
                        const state = view.state.field(suggestionsStateField)
                        const option = state.result.at(state.selectedOption)
                        if (!state.open || !option) {
                            return false
                        }
                        applyOption(view, option)
                        return true
                    },
                },
                {
                    key: 'Escape',
                    run(view) {
                        if (view.state.field(suggestionsStateField).open) {
                            view.dispatch({ effects: hideCompletion.of() })
                            return true
                        }
                        return false
                    },
                },
            ])
        ),
    ],
})

// TODO:
//  - Only handle keybindings when suggestions are open
//  - fix "open quote problem" (stretch goal)

export const suggestions = (id: string, parent: HTMLDivElement, source: Source, history: History): Extension => [
    suggestionsConfig.of({ history, id }),
    suggestionSource.of(source),
    ViewPlugin.define(view => new SuggestionView(id, view, parent)),
]
