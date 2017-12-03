import * as H from 'history'
import * as React from 'react'
import { Observable } from 'rxjs/Observable'
import { fromEvent } from 'rxjs/observable/fromEvent'
import { merge } from 'rxjs/observable/merge'
import { of } from 'rxjs/observable/of'
import { catchError } from 'rxjs/operators/catchError'
import { debounceTime } from 'rxjs/operators/debounceTime'
import { delay } from 'rxjs/operators/delay'
import { distinctUntilChanged } from 'rxjs/operators/distinctUntilChanged'
import { filter } from 'rxjs/operators/filter'
import { map } from 'rxjs/operators/map'
import { publishReplay } from 'rxjs/operators/publishReplay'
import { refCount } from 'rxjs/operators/refCount'
import { repeat } from 'rxjs/operators/repeat'
import { switchMap } from 'rxjs/operators/switchMap'
import { takeUntil } from 'rxjs/operators/takeUntil'
import { tap } from 'rxjs/operators/tap'
import { toArray } from 'rxjs/operators/toArray'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { eventLogger } from '../tracking/eventLogger'
import { scrollIntoView } from '../util'
import { fetchSuggestions } from './backend'
import { SearchOptions } from './index'
import { createSuggestion, Suggestion, SuggestionItem } from './Suggestion'

interface Props {
    history: H.History

    /** The value of the query input */
    value: string

    /** Called when the value changes */
    onChange: (newValue: string) => void

    /** The query provided by the active search scope */
    scopeQuery?: string

    /**
     * A string that is appended to the query input's query before
     * fetching suggestions.
     */
    prependQueryForSuggestions?: string

    /** Whether the input should be autofocused (and the behavior thereof) */
    autoFocus?: 'cursor-at-end'
}

interface State {
    /** Whether the query input is focused */
    inputFocused: boolean

    /** Whether suggestions are shown or not */
    hideSuggestions: boolean

    /** The suggestions shown to the user */
    suggestions: Suggestion[]

    /** Index of the currently selected suggestion (-1 if none selected) */
    selectedSuggestion: number

    /** Whether suggestions are currently being fetched */
    loading: boolean
}

export class QueryInput extends React.Component<Props, State> {
    private static SUGGESTIONS_QUERY_MIN_LENGTH = 2

    /** Subscriptions to unsubscribe from on component unmount */
    private subscriptions = new Subscription()

    /** Emits on keydown events in the input field */
    private inputKeyDowns = new Subject<React.KeyboardEvent<HTMLInputElement>>()

    /** Emits new input values */
    private inputValues = new Subject<string>()

    /** Emits when the input field is clicked */
    private inputFocuses = new Subject<void>()

    /** Emits when the suggestions are hidden */
    private suggestionsHidden = new Subject<void>()

    /** Only used for selection and focus management */
    private inputElement?: HTMLInputElement

    /** Only used for scroll state management */
    private suggestionListElement?: HTMLElement

    /** Only used for scroll state management */
    private selectedSuggestionElement?: HTMLElement

    constructor(props: Props) {
        super(props)

        this.state = {
            hideSuggestions: false,
            inputFocused: false,
            loading: false,
            selectedSuggestion: -1,
            suggestions: [],
        }

        this.subscriptions.add(
            // Trigger new suggestions every time the input field is typed into
            this.inputValues
                .pipe(
                    tap(query => this.props.onChange(query)),
                    distinctUntilChanged(),
                    debounceTime(200),
                    switchMap(query => {
                        if (query.length < QueryInput.SUGGESTIONS_QUERY_MIN_LENGTH) {
                            return [{ suggestions: [], selectedSuggestion: -1, loading: false }]
                        }
                        const options: SearchOptions = {
                            query: [this.props.prependQueryForSuggestions, this.props.value].filter(s => !!s).join(' '),
                            scopeQuery: this.props.scopeQuery || '',
                        }
                        const suggestionsFetch = fetchSuggestions(options).pipe(
                            map(createSuggestion),
                            toArray(),
                            map((suggestions: Suggestion[]) => ({
                                suggestions,
                                selectedSuggestion: -1,
                                hideSuggestions: false,
                                loading: false,
                            })),
                            catchError((err: Error) => {
                                console.error(err)
                                return []
                            }),
                            publishReplay(),
                            refCount()
                        )
                        return merge(
                            suggestionsFetch,
                            // Show a loader if the fetch takes longer than 100ms
                            of({ loading: true }).pipe(delay(100), takeUntil(suggestionsFetch))
                        )
                    }),
                    // Abort suggestion display on route change or suggestion hiding
                    takeUntil(this.suggestionsHidden),
                    // But resubscribe afterwards
                    repeat()
                )
                .subscribe(
                    state => {
                        this.setState(state as State)
                    },
                    err => {
                        console.error(err)
                    }
                )
        )

        // Quick-Open hotkeys
        this.subscriptions.add(
            fromEvent<KeyboardEvent>(window, 'keydown')
                .pipe(
                    filter(
                        event =>
                            // Slash shortcut (if no input element is focused)
                            (event.key === '/' && !['INPUT', 'TEXTAREA'].includes(document.activeElement.nodeName)) ||
                            // Cmd/Ctrl+P shortcut
                            ((event.metaKey || event.ctrlKey) && event.key === 'p') ||
                            // Cmd/Ctrl+Shift+F shortcut
                            ((event.metaKey || event.ctrlKey) && event.shiftKey && event.key === 'f')
                    ),
                    switchMap(event => {
                        event.preventDefault()
                        // Use selection as query
                        const selection = window.getSelection().toString()
                        if (selection) {
                            return new Observable<void>(observer =>
                                this.setState(
                                    {
                                        // query: selection, TODO(sqs): add back this behavior
                                        suggestions: [],
                                        selectedSuggestion: -1,
                                    },
                                    () => {
                                        observer.next()
                                        observer.complete()
                                    }
                                )
                            )
                        }
                        return [undefined]
                    })
                )
                .subscribe(() => {
                    if (this.inputElement) {
                        // Select all input
                        this.inputElement.focus()
                        this.inputElement.setSelectionRange(0, this.inputElement.value.length)
                    }
                })
        )
    }

    public componentDidMount(): void {
        switch (this.props.autoFocus) {
            case 'cursor-at-end':
                this.focusInputAndPositionCursorAtEnd()
                break
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public componentDidUpdate(prevProps: Props, prevState: State): void {
        // Check if selected suggestion is out of view
        scrollIntoView(this.suggestionListElement, this.selectedSuggestionElement)
    }

    public render(): JSX.Element | null {
        const showSuggestions =
            this.props.value.length >= QueryInput.SUGGESTIONS_QUERY_MIN_LENGTH &&
            this.state.inputFocused &&
            !this.state.hideSuggestions &&
            this.state.suggestions.length !== 0

        return (
            <div className="query-input2">
                <input
                    className="form-control query-input2__input"
                    value={this.props.value}
                    onChange={this.onInputChange}
                    onKeyDown={this.onInputKeyDown}
                    onFocus={this.onInputFocus}
                    onBlur={this.onInputBlur}
                    spellCheck={false}
                    autoCapitalize="off"
                    placeholder="Search code..."
                    ref={ref => (this.inputElement = ref!)}
                />
                {showSuggestions && (
                    <ul className="query-input2__suggestions" ref={this.setSuggestionListElement}>
                        {this.state.suggestions.map((suggestion, i) => {
                            const isSelected = this.state.selectedSuggestion === i
                            const onRef = (ref: HTMLLIElement | null) => {
                                if (isSelected) {
                                    this.selectedSuggestionElement = ref || undefined
                                }
                            }
                            return (
                                <SuggestionItem
                                    key={i}
                                    suggestion={suggestion}
                                    isSelected={isSelected}
                                    // tslint:disable-next-line:jsx-no-lambda
                                    onClick={() => this.selectSuggestion(suggestion)}
                                    liRef={onRef}
                                />
                            )
                        })}
                    </ul>
                )}
            </div>
        )
    }

    private setSuggestionListElement = (ref: HTMLElement | null): void => {
        this.suggestionListElement = ref || undefined
    }

    private selectSuggestion = (suggestion: Suggestion): void => {
        eventLogger.log('SearchSuggestionSelected', {
            code_search: {
                suggestion: {
                    type: suggestion.type,
                    url: suggestion.url,
                },
            },
        })

        this.props.history.push(suggestion.url)

        this.suggestionsHidden.next()
        this.setState({ hideSuggestions: true, selectedSuggestion: -1 })
    }

    private focusInputAndPositionCursorAtEnd(): void {
        if (this.inputElement) {
            // Focus the input element and set cursor to the end
            this.inputElement.focus()
            this.inputElement.setSelectionRange(this.inputElement.value.length, this.inputElement.value.length)
        }
    }

    private onInputChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        this.inputValues.next(event.currentTarget.value)
    }

    private onInputFocus: React.FocusEventHandler<HTMLInputElement> = event => {
        this.inputFocuses.next()
        this.setState({ inputFocused: true })
    }

    private onInputBlur: React.FocusEventHandler<HTMLInputElement> = event => {
        this.suggestionsHidden.next()
        this.setState({ inputFocused: false, hideSuggestions: true })
    }

    private onInputKeyDown: React.KeyboardEventHandler<HTMLInputElement> = event => {
        event.persist()
        this.inputKeyDowns.next(event)
        switch (event.key) {
            case 'Escape': {
                this.suggestionsHidden.next()
                this.setState({ hideSuggestions: true })
                break
            }
            case 'ArrowDown': {
                event.preventDefault()
                this.moveSelection(1)
                break
            }
            case 'ArrowUp': {
                event.preventDefault()
                this.moveSelection(-1)
                break
            }
            case 'Enter': {
                if (this.state.selectedSuggestion === -1) {
                    // Submit form and hide suggestions
                    this.suggestionsHidden.next()
                    this.setState({ hideSuggestions: true })
                    break
                }

                // Select suggestion
                event.preventDefault()
                if (this.state.suggestions.length === 0) {
                    break
                }
                this.selectSuggestion(this.state.suggestions[Math.max(this.state.selectedSuggestion, 0)])
                this.setState({ hideSuggestions: true })
                break
            }
        }
    }

    private moveSelection(steps: number): void {
        this.setState({
            selectedSuggestion: Math.max(
                Math.min(this.state.selectedSuggestion + steps, this.state.suggestions.length - 1),
                -1
            ),
        })
    }
}
