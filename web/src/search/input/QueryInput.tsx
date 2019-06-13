import * as H from 'history'
import * as React from 'react'
import { fromEvent, merge, Observable, of, Subject, Subscription } from 'rxjs'
import {
    catchError,
    debounceTime,
    delay,
    distinctUntilChanged,
    filter,
    map,
    publishReplay,
    refCount,
    repeat,
    startWith,
    switchMap,
    takeUntil,
    tap,
    toArray,
} from 'rxjs/operators'
import { Key } from 'ts-key-enum'
import { eventLogger } from '../../tracking/eventLogger'
import { scrollIntoView } from '../../util'
import { fetchSuggestions } from '../backend'
import { QueryInput3 } from './QueryInput3'
import { createSuggestion, Suggestion, SuggestionItem } from './Suggestion'

/**
 * Whether the experimental token text query input field is enabled.
 *
 * To enable this, run `localStorage.tokenTextQueryInputExp=true;location.reload()` in your browser's JavaScript
 * console.
 */
const USE_TOKEN_TEXT_QUERY_INPUT = localStorage.getItem('tokenTextQueryInputExp') !== null

/**
 * The query input field is clobbered and updated to contain this subject's values, as
 * they are received. This is used to trigger an update; the source of truth is still the URL.
 */
export const queryUpdates = new Subject<string>()

export interface QueryInputProps {
    location: H.Location
    history: H.History

    /** The value of the query input */
    value: string

    /** Called when the value changes */
    onChange: (newValue: string) => void

    /**
     * A string that is appended to the query input's query before
     * fetching suggestions.
     */
    prependQueryForSuggestions?: string

    /** Whether the input should be autofocused (and the behavior thereof) */
    autoFocus?: true | 'cursor-at-end'

    /** The input placeholder, if different from the default is desired. */
    placeholder?: string

    /**
     * Whether this input should behave like the global query input: (1)
     * pressing the '/' key focuses it and (2) other components contribute a
     * query to it with their context (such as the repository area contributing
     * 'repo:foo@bar' for the current repository and revision).
     *
     * At most one query input per page should have this behavior.
     */
    hasGlobalQueryBehavior?: boolean
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

class QueryInput2 extends React.Component<QueryInputProps, State> {
    private static SUGGESTIONS_QUERY_MIN_LENGTH = 2

    private componentUpdates = new Subject<QueryInputProps>()

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

    /** Only used to keep track if the user has typed a single character into the input field so we can log an event once. */
    private hasLoggedFirstInput = false

    constructor(props: QueryInputProps) {
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
                        if (query.length < QueryInput2.SUGGESTIONS_QUERY_MIN_LENGTH) {
                            return [{ suggestions: [], selectedSuggestion: -1, loading: false }]
                        }
                        const fullQuery = [this.props.prependQueryForSuggestions, this.props.value]
                            .filter(s => !!s)
                            .join(' ')
                        const suggestionsFetch = fetchSuggestions(fullQuery).pipe(
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
                                this.setState({ loading: false })
                                // HACK: if we catchError before 100ms, then the loader will display over us.
                                // This is not a good fix.
                                setTimeout(() => this.setState({ loading: false }), 120)
                                return [{}]
                            }),
                            publishReplay(),
                            refCount()
                        )
                        return merge(
                            suggestionsFetch,
                            // Show a loader if the fetch takes longer than 100ms
                            of({ loading: true }).pipe(
                                delay(100),
                                takeUntil(suggestionsFetch)
                            )
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
                        this.setState({ loading: false })
                        console.error(err)
                    }
                )
        )

        if (this.props.hasGlobalQueryBehavior) {
            // Quick-Open hotkeys
            this.subscriptions.add(
                fromEvent<KeyboardEvent>(window, 'keydown')
                    .pipe(
                        filter(
                            event =>
                                // Slash shortcut (if no input element is focused)
                                (event.key === '/' &&
                                    !!document.activeElement &&
                                    !['INPUT', 'TEXTAREA'].includes(document.activeElement.nodeName)) ||
                                // Cmd/Ctrl+Shift+F shortcut
                                ((event.metaKey || event.ctrlKey) && event.shiftKey && event.key === 'f')
                        ),
                        switchMap(event => {
                            event.preventDefault()
                            // Use selection as query
                            const selection = window.getSelection()
                            if (selection && selection.toString() !== '') {
                                return new Observable<void>(observer =>
                                    this.setState(
                                        {
                                            // query: selection.toString(), TODO(sqs): add back this behavior
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

            // Allow other components to update the query (e.g., to be relevant to what the user is
            // currently viewing).
            this.subscriptions.add(
                queryUpdates.pipe(distinctUntilChanged()).subscribe(query => this.props.onChange(query))
            )

            /** Whenever the URL query has a "focus" property, remove it and focus the query input. */
            this.subscriptions.add(
                this.componentUpdates
                    .pipe(
                        startWith(props),
                        filter(({ location }) => new URLSearchParams(location.search).get('focus') !== null)
                    )
                    .subscribe(props => {
                        this.focusInputAndPositionCursorAtEnd()

                        const q = new URLSearchParams(props.location.search)
                        q.delete('focus')
                        this.props.history.replace({ search: q.toString() })
                    })
            )
        }
    }

    public componentDidMount(): void {
        switch (this.props.autoFocus) {
            case 'cursor-at-end':
                this.focusInputAndPositionCursorAtEnd()
                break
        }
    }

    public componentWillReceiveProps(newProps: QueryInputProps): void {
        this.componentUpdates.next(newProps)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public componentDidUpdate(prevProps: QueryInputProps, prevState: State): void {
        // Check if selected suggestion is out of view
        scrollIntoView(this.suggestionListElement, this.selectedSuggestionElement)
    }

    public render(): JSX.Element | null {
        const showSuggestions =
            this.props.value.length >= QueryInput2.SUGGESTIONS_QUERY_MIN_LENGTH &&
            this.state.inputFocused &&
            !this.state.hideSuggestions &&
            this.state.suggestions.length !== 0

        return (
            <div className="query-input2">
                <input
                    className="form-control query-input2__input rounded-left e2e-query-input"
                    type="text"
                    value={this.props.value}
                    autoFocus={this.props.autoFocus === true}
                    onChange={this.onInputChange}
                    onKeyDown={this.onInputKeyDown}
                    onFocus={this.onInputFocus}
                    onBlur={this.onInputBlur}
                    spellCheck={false}
                    autoCapitalize="off"
                    placeholder={this.props.placeholder === undefined ? 'Search code...' : this.props.placeholder}
                    ref={ref => (this.inputElement = ref!)}
                    name="query"
                    autoComplete="off"
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
        // 🚨 PRIVACY: never provide any private data in { code_search: { suggestion: { type } } }.
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
        if (!this.hasLoggedFirstInput) {
            eventLogger.log('SearchInitiated')
            this.hasLoggedFirstInput = true
        }
        this.inputValues.next(event.currentTarget.value)
    }

    private onInputFocus: React.FocusEventHandler<HTMLInputElement> = event => {
        this.inputFocuses.next()
        this.setState({ inputFocused: true })
    }

    private onInputBlur: React.FocusEventHandler<HTMLInputElement> = event => {
        this.suggestionsHidden.next()
        this.setState({ inputFocused: false, loading: false, hideSuggestions: true })
    }

    private onInputKeyDown: React.KeyboardEventHandler<HTMLInputElement> = event => {
        event.persist()
        this.inputKeyDowns.next(event)
        switch (event.key) {
            case Key.Escape: {
                this.suggestionsHidden.next()
                this.setState({ loading: false, hideSuggestions: true, selectedSuggestion: -1 })
                break
            }
            case Key.ArrowDown: {
                event.preventDefault()
                this.moveSelection(1)
                break
            }
            case Key.ArrowUp: {
                event.preventDefault()
                this.moveSelection(-1)
                break
            }
            case Key.Enter: {
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

export const QueryInput = USE_TOKEN_TEXT_QUERY_INPUT ? QueryInput3 : QueryInput2
