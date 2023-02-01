import React from 'react'

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
import { createRoot, Root } from 'react-dom/client'

import { compatNavigate, HistoryOrNavigate } from '@sourcegraph/common'

import { Suggestions } from './Suggestions'

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

export type CustomRenderer = (option: Option) => React.ReactElement

export interface Option {
    /**
     * The label the input is matched against and shown in the UI.
     */
    label: string
    /**
     * What to do when this option is applied (via Enter)
     */
    action: Action
    /**
     * Options can have perform an alternative action when applied via
     * Shift+Enter.
     */
    alternativeAction?: Action
    /**
     * A short description of the option, shown next to the label.
     */
    description?: string
    /**
     * The SVG path of the icon to use for this option.
     */
    icon?: string
    /**
     * If present the provided component will be used to render the label of the
     * option.
     */
    render?: CustomRenderer
    /**
     * If present this component is rendered as footer.
     */
    info?: CustomRenderer
    /**
     * A set of character indexes. If provided the characters of at these
     * positions in the label will be highlighted as matches.
     */
    matches?: Set<number>
}

export interface CommandAction {
    type: 'command'
    apply: (view: EditorView) => void
    name?: string
}
export interface GoToAction {
    type: 'goto'
    url: string
    name?: string
}
export interface CompletionAction {
    type: 'completion'
    from: number
    name?: string
    to?: number
    insertValue?: string
}
export type Action = CommandAction | GoToAction | CompletionAction

export interface Group {
    title: string
    options: Option[]
}

class SuggestionView {
    private container: HTMLDivElement
    private root: Root

    private onSelect = (option: Option): void => {
        applyAction(this.view, option.action, option)
        // Query input looses focus when option is selected via
        // mousedown/click. This is a necessary hack to re-focus the query
        // input.
        window.requestAnimationFrame(() => this.view.contentDOM.focus())
    }

    constructor(private readonly id: string, public readonly view: EditorView, public parent: HTMLDivElement) {
        const state = view.state.field(suggestionsStateField)
        this.container = document.createElement('div')
        this.root = createRoot(this.container)
        parent.append(this.container)

        // We need to delay the initial render otherwise React complains that
        // wer are rendering a component while already rendering another one
        // (the query input component)
        setTimeout(() => {
            this.root.render(
                React.createElement(Suggestions, {
                    id,
                    results: state.result.groups,
                    activeRowIndex: state.selectedOption,
                    open: state.open,
                    onSelect: this.onSelect,
                })
            )
        }, 0)
    }

    public update(update: ViewUpdate): void {
        const state = update.state.field(suggestionsStateField)

        if (state !== update.startState.field(suggestionsStateField)) {
            this.updateResults(state)
        }
    }

    private updateResults(state: SuggestionsState): void {
        this.root.render(
            React.createElement(Suggestions, {
                id: this.id,
                results: state.result.groups,
                activeRowIndex: state.selectedOption,
                open: state.open,
                onSelect: this.onSelect,
            })
        )
    }

    public destroy(): void {
        this.container.remove()

        // We need to delay unmounting the root otherwise React complains about
        // synchronsouly unmounting multiple components.
        setTimeout(() => {
            this.root.unmount()
        }, 0)
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
    private allOptions: Option[]

    constructor(
        public readonly groups: Group[],
        public valid: (state: EditorState, position: number) => boolean = () => false
    ) {
        this.allOptions = groups.flatMap(group => group.options)
    }

    // eslint-disable-next-line id-length
    public at(index: number): Option | undefined {
        return this.allOptions[index]
    }

    public groupRowIndex(index: number): [number, number] | undefined {
        const option = this.allOptions[index]

        if (!option) {
            return undefined
        }

        for (let groupIndex = 0; groupIndex < this.groups.length; groupIndex++) {
            const options = this.groups[groupIndex].options
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
        return this.allOptions.length
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
        // TODO: We probably don't want to trigger fetches on every doc changed
        if (isUserInput(transaction) || transaction.docChanged) {
            return this.query(transaction.state)
        }

        if (transaction.selection) {
            if (!transaction.selection.main.empty) {
                // Hide suggestions when the user selects a range in the input
                return new RegisteredSource(this.source, RegisteredSourceState.Inactive, this.result)
            }
            if (this.result.valid(transaction.state, transaction.newSelection.main.head)) {
                return this
            }
            return this.query(transaction.state)
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
    return (
        transaction.isUserEvent('input.type') ||
        transaction.isUserEvent('input.paste') ||
        transaction.isUserEvent('delete')
    )
}

interface Config {
    id: string
    historyOrNavigate?: HistoryOrNavigate
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

function applyAction(view: EditorView, action: Action, option: Option): void {
    switch (action.type) {
        case 'completion':
            {
                const text = action.insertValue ?? option.label
                view.dispatch({
                    ...view.state.changeByRange(range => {
                        if (range === view.state.selection.main) {
                            return {
                                changes: {
                                    from: action.from,
                                    to: action.to ?? view.state.selection.main.head,
                                    insert: text,
                                },
                                range: EditorSelection.cursor(action.from + text.length),
                            }
                        }
                        return { range }
                    }),
                })
            }
            break
        case 'command':
            action.apply(view)
            break
        case 'goto':
            {
                const historyOrNavigate = view.state.facet(suggestionsConfig).historyOrNavigate
                if (historyOrNavigate) {
                    compatNavigate(historyOrNavigate, action.url)
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
                        applyAction(view, option.action, option)
                        return true
                    },
                    shift(view) {
                        const state = view.state.field(suggestionsStateField)
                        const option = state.result.at(state.selectedOption)
                        if (!state.open || !option || !option.alternativeAction) {
                            return false
                        }
                        applyAction(view, option.alternativeAction, option)
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

export const suggestions = (
    id: string,
    parent: HTMLDivElement,
    source: Source,
    historyOrNavigate: HistoryOrNavigate
): Extension => [
    suggestionsConfig.of({ historyOrNavigate, id }),
    suggestionSource.of(source),
    ViewPlugin.define(view => new SuggestionView(id, view, parent)),
]
