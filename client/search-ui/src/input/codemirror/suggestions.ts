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
import { Command, EditorView, keymap, ViewPlugin, ViewUpdate } from '@codemirror/view'
import { History } from 'history'
import { SvelteComponentTyped } from 'svelte'
import Suggestions from './Suggestions.svelte'
export { default as FilterSuggestion } from './FilterSuggestion.svelte'

export interface SuggestionResult {
    result: Group[]
    asyncResult?: Promise<{ result: Group[] }>
    valid?: (state: EditorState, position: number) => boolean
}

/**
 * A source for completion/suggestion results
 */
export interface Source {
    (state: EditorState, position: number): Promise<SuggestionResult> | SuggestionResult
}

enum RegisteredSourceState {
    Inactive,
    Pending,
    Fetching,
    Complete,
}

export type CustomRenderer = typeof SvelteComponentTyped<{ option: Option }>

export type Option =
    | {
          type: 'command'
          value: string
          apply: (view: EditorView) => void
          matches?: Set<number>
          // svg path
          icon?: string
          render?: CustomRenderer
          description?: string
      }
    | {
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
    | {
          type: 'target'
          value: string
          url: string
          matches?: Set<number>
          // svg path
          icon?: string
          render?: CustomRenderer
          description?: string
      }

export type EntryOf<T> = Extract<Option, { type: T }>

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
            },
        })
        this.view.dom.appendChild(this.root)
    }

    update(update: ViewUpdate): void {
        const state = update.state.field(suggestionsStateField)

        if (state !== update.startState.field(suggestionsStateField)) {
            this.updateResults(state)
        }
    }

    updateResults(state: SuggestionsState) {
        this.instance.$set({ results: state.result.groups, activeRowIndex: state.selectedOption })
    }

    destroy() {
        this.instance.$destroy()
        this.root.remove()
    }
}

const completionPlugin = ViewPlugin.fromClass(
    class {
        running: RunningQuery | null = null

        constructor(public readonly view: EditorView) {
            this.startQuery(view.state.field(suggestionsStateField).source)
        }

        update(update: ViewUpdate): void {
            if (update.view.hasFocus) {
                this.startQuery(update.state.field(suggestionsStateField).source)
            }
        }

        async startQuery(source: RegisteredSource) {
            const { state } = this.view
            if (
                source.state === RegisteredSourceState.Pending &&
                (!this.running || this.running.timestamp !== source.timestamp)
            ) {
                const query = new RunningQuery(source)
                this.running = query
                Promise.resolve(source.run(state, state.selection.main.anchor)).then(result => {
                    if (this.running === query) {
                        this.running = null
                        this.view.dispatch({ effects: updateResultEffect.of({ source, result }) })
                    }
                })
            } else if (source.state === RegisteredSourceState.Inactive) {
                this.running = null
            }
        }
    }
)

class RunningQuery {
    constructor(public readonly source: RegisteredSource) {}

    get timestamp(): number {
        return this.source.timestamp
    }
}

const defaultValid = () => false

class Result {
    private entries: Option[]
    constructor(
        readonly groups: Group[],
        public valid: (state: EditorState, position: number) => boolean = defaultValid
    ) {
        this.entries = groups.flatMap(group => group.entries)
    }

    at(index: number): Option | undefined {
        return this.entries[index]
    }

    empty(): boolean {
        return this.length === 0
    }

    get length(): number {
        return this.entries.length
    }
}

const emptyResult = new Result([])

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
        private readonly nextResults?: Promise<SuggestionResult>
    ) {
        switch (state) {
            case RegisteredSourceState.Pending:
                this.timestamp = Date.now()
                break
            default:
                this.timestamp = -1
        }
    }

    update(transaction: Transaction): RegisteredSource {
        if (isUserInput(transaction)) {
            return new RegisteredSource(this.source, RegisteredSourceState.Pending, emptyResult)
        }
        if (transaction.selection) {
            if (this.result.valid(transaction.state, transaction.newSelection.main.head)) {
                return this
            }
            return new RegisteredSource(this.source, RegisteredSourceState.Pending, emptyResult)
        }
        if (transaction.docChanged) {
            return new RegisteredSource(this.source, RegisteredSourceState.Inactive, emptyResult)
        }
        for (const effect of transaction.effects) {
            if (
                effect.is(updateResultEffect) &&
                effect.value.source.source === this.source &&
                this.state === RegisteredSourceState.Pending
            ) {
                const nextState = effect.value.result.asyncResult
                    ? RegisteredSourceState.Pending
                    : RegisteredSourceState.Complete
                return new RegisteredSource(
                    this.source,
                    nextState,
                    new Result(effect.value.result.result, effect.value.result.valid),
                    effect.value.result.asyncResult
                )
            }
            if (effect.is(startCompletion)) {
                return new RegisteredSource(this.source, RegisteredSourceState.Pending, emptyResult)
            }
        }
        return this
    }

    run(...args: Parameters<Source>): ReturnType<Source> {
        return this.nextResults ?? this.source(...args)
    }
}

class SuggestionsState {
    constructor(readonly id: string, readonly source: RegisteredSource, readonly selectedOption: number) {}

    update(transaction: Transaction): SuggestionsState {
        let state: SuggestionsState = this

        const source = transaction.state.facet(suggestionSource)
        let registeredSource =
            source === this.source.source
                ? this.source
                : new RegisteredSource(source, RegisteredSourceState.Inactive, emptyResult)
        registeredSource = registeredSource.update(transaction)
        if (registeredSource !== this.source) {
            state = new SuggestionsState(
                state.id,
                registeredSource,
                registeredSource.timestamp === this.source.timestamp
                    ? this.selectedOption
                    : this.selectedOption === 0
                    ? this.selectedOption
                    : -1
            )
        }

        for (const effect of transaction.effects) {
            if (effect.is(setSelectedEffect)) {
                state = new SuggestionsState(state.id, state.source, effect.value)
            }
        }

        return state
    }

    get result(): Result {
        return this.source.result
    }
}

function isUserInput(transaction: Transaction): boolean {
    return transaction.isUserEvent('input.type') || transaction.isUserEvent('delete.backward')
}

function arraysAreEqual<T>(a: T[], b: T[]): boolean {
    return a.length === b.length && a.every((item, index) => item === b[index])
}

const none: any[] = []
const setSelectedEffect = StateEffect.define<number>()
const startCompletion = StateEffect.define<void>()
const hideCompletion = StateEffect.define<void>()
const updateResultEffect = StateEffect.define<{ source: RegisteredSource; result: SuggestionResult }>()
const suggestionsStateField = StateField.define<SuggestionsState>({
    create() {
        return new SuggestionsState(
            'suggestions-' + Math.floor(Math.random() * 2e6).toString(36),
            new RegisteredSource(() => ({ result: [] }), RegisteredSourceState.Inactive, emptyResult),
            -1
        )
    },

    update(state, transaction) {
        return state.update(transaction)
    },

    provide(field) {
        return EditorView.contentAttributes.compute([field], state => {
            const result = state.field(field).result
            return {
                'aria-expanded': result.empty() ? 'false' : 'true',
            }
        })
    },
})

function moveSelection(direction: 'forward' | 'backward'): Command {
    const forward = direction === 'forward'
    return view => {
        const state = view.state.field(suggestionsStateField, false)
        if (!state || state.source.result.empty()) {
            return false
        }
        const { length } = state?.source.result
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
                    changes: { from: option.from, to: option.to ?? view.state.selection.main.head, insert: text },
                    // Move cursor to the end of the inserted text
                    selection: EditorSelection.cursor(option.from + text.length),
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
                        if (!view.state.field(suggestionsStateField).open) {
                            view.dispatch({ effects: startCompletion.of() })
                            return true
                        }
                        return false
                    },
                },
                {
                    key: 'Enter',
                    run(view) {
                        const state = view.state.field(suggestionsStateField)
                        const option = state.result.at(state.selectedOption)
                        if (option) {
                            applyOption(view, option)
                            return true
                        }
                        return false
                    },
                },
                {
                    key: 'Tab',
                    run(view) {
                        const state = view.state.field(suggestionsStateField)
                        const option = state.result.at(state.selectedOption)
                        if (option) {
                            applyOption(view, option)
                            return true
                        }
                        return false
                    },
                },
                {
                    key: 'Esc',
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

interface Config {
    history?: History
}

const suggestionsConfig = Facet.define<Config, Config>({
    combine(configs) {
        return configs[0] ?? {}
    },
})

// TODO: Indicate current row via active-descendant
// TODO: Fix DOM hierarchy via aria-owns
// TODO: Query data
// TODO: Only handle keybindings when suggestions are open
// TODO:
//  - Content specific headings
//  - Value completion for repo: and file:
//  - fix "open quote problem" (stretch goal)

export const suggestions = (id: string, parent: HTMLDivElement, source: Source, history: History): Extension => {
    return [
        suggestionsConfig.of({ history }),
        suggestionSource.of(source),
        ViewPlugin.define(view => new SuggestionView(id, view, parent)),
    ]
}
