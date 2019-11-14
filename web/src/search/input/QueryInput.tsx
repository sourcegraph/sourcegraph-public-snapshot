import * as H from 'history'
import * as React from 'react'
import { fromEvent, Subject, Subscription } from 'rxjs'
import {
    debounceTime,
    distinctUntilChanged,
    filter,
    repeat,
    startWith,
    takeUntil,
    tap,
    switchMap,
    map,
    toArray,
    catchError,
} from 'rxjs/operators'
import { eventLogger } from '../../tracking/eventLogger'
import { scrollIntoView } from '../../util'
import { Suggestion, SuggestionItem, SuggestionTypes, createSuggestion, fuzzySearchFilters } from './Suggestion'
import RegexpToggle from './RegexpToggle'
import { SearchPatternType } from '../../../../shared/src/graphql/schema'
import { PatternTypeProps } from '..'
import Downshift from 'downshift'
import { searchFilterSuggestions } from '../searchFilterSuggestions'
import {
    QueryState,
    filterStaticSuggestions,
    insertSuggestionInQuery,
    isFuzzyWordSearch,
    lastFilterAndValueBeforeCursor,
    formatQueryForFuzzySearch,
} from '../helpers'
import { fetchSuggestions } from '../backend'
import { isDefined } from '../../../../shared/src/util/types'

/**
 * The query input field is clobbered and updated to contain this subject's values, as
 * they are received. This is used to trigger an update; the source of truth is still the URL.
 */
export const queryUpdates = new Subject<string>()

interface Props extends PatternTypeProps {
    location: H.Location
    history: H.History

    /** The value of the query input */
    value: QueryState

    /** Called when the value changes */
    onChange: (newValue: QueryState) => void

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

/**
 * The search suggestions and cursor position of where the last character was inserted.
 * Cursor position is used to correctly insert the suggestion when it's selected.
 */
interface ComponentSuggestions {
    values: Suggestion[]
    cursorPosition: number
}

interface State {
    /** The suggestions shown to the user */
    suggestions: ComponentSuggestions
}

const noSuggestions: State['suggestions'] = { values: [], cursorPosition: 0 }

export class QueryInput extends React.Component<Props, State> {
    private componentUpdates = new Subject<Props>()

    /** Subscriptions to unsubscribe from on component unmount */
    private subscriptions = new Subscription()

    /** Emits new input values */
    private inputValues = new Subject<QueryState>()

    /** Emits when the suggestions are hidden */
    private suggestionsHidden = new Subject<void>()

    /** Only used for selection and focus management */
    private inputElement = React.createRef<HTMLInputElement>()

    /** Used for scrolling suggestions into view while scrolling with keyboard */
    private containerElement = React.createRef<HTMLDivElement>()

    /** Only used to keep track if the user has typed a single character into the input field so we can log an event once. */
    private hasLoggedFirstInput = false

    public state: State = {
        suggestions: {
            cursorPosition: 0,
            values: [],
        },
    }

    constructor(props: Props) {
        super(props)

        this.subscriptions.add(
            // Trigger new suggestions every time the input field is typed into.
            this.inputValues
                .pipe(
                    tap(queryState => this.props.onChange(queryState)),
                    distinctUntilChanged(),
                    debounceTime(200),
                    switchMap(queryState => {
                        if (queryState.query.length === 0) {
                            return [{ suggestions: noSuggestions }]
                        }

                        // A filter value (in "archive:yes", "archive" is the filter and "yes" is the value)
                        // can either be from `searchFilterSuggestions` or from the fuzzy-search.

                        // First get static suggestions
                        const staticSuggestions = {
                            cursorPosition: queryState.cursorPosition,
                            values: filterStaticSuggestions(queryState, searchFilterSuggestions),
                        }

                        // Used to know if a filter value, and not just a separate word, is being typed
                        const filterAndValueBeforeCursor = lastFilterAndValueBeforeCursor(queryState)

                        // If a filter value is being typed but selected filter does not use
                        // fuzzy-search suggestions, then return only static suggestions
                        if (
                            filterAndValueBeforeCursor &&
                            !fuzzySearchFilters.includes(filterAndValueBeforeCursor.filter)
                        ) {
                            return [{ suggestions: staticSuggestions }]
                        }

                        const queryForFuzzySearch = formatQueryForFuzzySearch(queryState)
                        const fullQuery = (this.props.prependQueryForSuggestions ?? '') + ' ' + queryForFuzzySearch

                        // Dos a fuzzy search and formats suggestions for display.
                        return fetchSuggestions(fullQuery).pipe(
                            map(createSuggestion),
                            filter(isDefined),
                            map((suggestion): Suggestion => ({ ...suggestion, fromFuzzySearch: true })),
                            filter(suggestion => {
                                // Only show fuzzy-suggestions that are relevant to the typed filter
                                if (filterAndValueBeforeCursor?.filter) {
                                    switch (filterAndValueBeforeCursor.filter) {
                                        case SuggestionTypes.repohasfile:
                                            return suggestion.type === SuggestionTypes.file
                                        default:
                                            return suggestion.type === filterAndValueBeforeCursor.filter
                                    }
                                }
                                return true
                            }),
                            toArray(),
                            map(suggestions => ({
                                suggestions: {
                                    cursorPosition: queryState.cursorPosition,
                                    values: staticSuggestions.values.concat(suggestions),
                                },
                            })),
                            catchError(error => {
                                console.error(error)
                                // If fuzzy-search is not capable of returning suggestions for the query
                                // or there is an internal error, then at least return the static suggestions
                                return [{ suggestions: staticSuggestions }]
                            })
                        )
                    }),
                    // Abort suggestion display on route change or suggestion hiding
                    takeUntil(this.suggestionsHidden),
                    // But resubscribe afterwards
                    repeat()
                )
                .subscribe(
                    state => {
                        this.setState(state)
                    },
                    err => {
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
                                // Cmd/Ctrl+Shift+F
                                (event.metaKey || event.ctrlKey) &&
                                event.shiftKey &&
                                event.key.toLowerCase() === 'f' &&
                                !!document.activeElement &&
                                !['INPUT', 'TEXTAREA'].includes(document.activeElement.nodeName)
                        )
                    )
                    .subscribe(() => {
                        const selection = String(window.getSelection() || '')
                        this.props.onChange({ query: selection, cursorPosition: selection.length })
                        if (this.inputElement.current) {
                            this.inputElement.current.focus()
                            // Select whole input text
                            this.inputElement.current.setSelectionRange(0, this.inputElement.current.value.length)
                        }
                    })
            )

            // Allow other components to update the query (e.g., to be relevant to what the user is
            // currently viewing).
            this.subscriptions.add(
                queryUpdates.pipe(distinctUntilChanged()).subscribe(query =>
                    this.props.onChange({
                        query,
                        cursorPosition: query.length,
                    })
                )
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

        this.subscriptions.add(
            // hide suggestions when clicking outside search input
            fromEvent<MouseEvent>(window, 'click').subscribe(event => {
                if (
                    this.state.suggestions.values.length > 0 && // prevent unnecessary render
                    (!this.containerElement.current || !this.containerElement.current.contains(event.target as Node))
                ) {
                    this.hideSuggestions()
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

    public componentDidUpdate(prevProps: Props): void {
        if (this.props.value.cursorPosition && prevProps.value.cursorPosition !== this.props.value.cursorPosition) {
            this.focusInputAndPositionCursor(this.props.value.cursorPosition)
        }
        this.componentUpdates.next(this.props)
    }

    public render(): JSX.Element | null {
        const showSuggestions = this.state.suggestions.values.length > 0
        // If last typed word is not a filter type,
        // suggestions should show url label and redirect on select.
        const showUrlLabel = isFuzzyWordSearch({
            query: this.props.value.query,
            cursorPosition: this.state.suggestions.cursorPosition,
        })
        return (
            <Downshift
                scrollIntoView={this.downshiftScrollIntoView}
                onSelect={this.onSuggestionSelect}
                itemToString={this.downshiftItemToString}
            >
                {({ getInputProps, getItemProps, getMenuProps, highlightedIndex }) => {
                    const { onChange: downshiftChange, onKeyDown } = getInputProps()
                    return (
                        <div className="query-input2">
                            <div ref={this.containerElement}>
                                <input
                                    className="form-control query-input2__input rounded-left e2e-query-input"
                                    value={this.props.value.query}
                                    autoFocus={this.props.autoFocus === true}
                                    onChange={(event: React.ChangeEvent<HTMLInputElement>) => {
                                        downshiftChange(event)
                                        this.onInputChange(event)
                                    }}
                                    onKeyDown={event => {
                                        this.onInputKeyDown(event)
                                        onKeyDown(event)
                                    }}
                                    spellCheck={false}
                                    autoCapitalize="off"
                                    placeholder={
                                        this.props.placeholder === undefined ? 'Search code...' : this.props.placeholder
                                    }
                                    ref={this.inputElement}
                                    name="query"
                                    autoComplete="off"
                                />
                                {showSuggestions && (
                                    <ul className="query-input2__suggestions e2e-query-suggestions" {...getMenuProps()}>
                                        {this.state.suggestions.values.map((suggestion, index) => {
                                            const isSelected = highlightedIndex === index
                                            const key = `${index}-${suggestion}`
                                            return (
                                                <SuggestionItem
                                                    key={key}
                                                    {...getItemProps({
                                                        key,
                                                        index,
                                                        item: suggestion,
                                                    })}
                                                    suggestion={suggestion}
                                                    isSelected={isSelected}
                                                    showUrlLabel={showUrlLabel}
                                                    defaultLabel="add to query"
                                                />
                                            )
                                        })}
                                    </ul>
                                )}
                                <RegexpToggle
                                    {...this.props}
                                    toggled={this.props.patternType === SearchPatternType.regexp}
                                    navbarSearchQuery={this.props.value.query}
                                />
                            </div>
                        </div>
                    )
                }}
            </Downshift>
        )
    }

    private downshiftItemToString = (suggestion?: Suggestion): string => (suggestion ? suggestion.value : '')

    private downshiftScrollIntoView = (node: HTMLElement, menuNode: HTMLElement): void => {
        scrollIntoView(menuNode, node)
    }

    private onInputKeyDown = (event: React.KeyboardEvent<HTMLInputElement>): void => {
        // Ctrl+Space to show all available filter type suggestions
        if (!this.props.value.query && event.ctrlKey && event.key === ' ') {
            this.setState({
                suggestions: {
                    cursorPosition: event.currentTarget.selectionStart ?? 0,
                    values: searchFilterSuggestions.filters.values,
                },
            })
        }
    }

    private hideSuggestions = (): void => {
        this.suggestionsHidden.next()
        this.setState({ suggestions: noSuggestions })
    }

    /**
     * if query only has one word and selected suggestion is not a filter: redirect to suggestion URL
     * else: add selected suggestion to query
     */
    private onSuggestionSelect = (suggestion: Suggestion | undefined): void => {
        this.setState((state, props) => {
            if (!suggestion) {
                return {
                    suggestions: noSuggestions,
                }
            }

            // 🚨 PRIVACY: never provide any private data in { code_search: { suggestion: { type } } }.
            eventLogger.log('SearchSuggestionSelected', {
                code_search: {
                    suggestion: {
                        type: suggestion.type,
                    },
                },
            })

            // if separate word is being typed and suggestion with url is selected
            if (
                isFuzzyWordSearch({
                    query: props.value.query,
                    cursorPosition: state.suggestions.cursorPosition,
                }) &&
                suggestion.url
            ) {
                this.props.history.push(suggestion.url)
                return { suggestions: noSuggestions }
            }

            const isValueSuggestion = suggestion.type !== SuggestionTypes.filters

            const { cursorPosition } = state.suggestions
            const { query: newQuery, cursorPosition: newCursorPosition } = insertSuggestionInQuery(
                props.value.query,
                // Add regex end of string boundary to limit future suggestion results
                !suggestion.fromFuzzySearch
                    ? suggestion
                    : {
                          ...suggestion,
                          value: suggestion.value + '$',
                      },
                cursorPosition
            )

            props.onChange({
                query: newQuery,
                cursorPosition: newCursorPosition,
            })

            // If a filter was selected, show filter value suggestions
            if (!isValueSuggestion) {
                this.inputValues.next({
                    cursorPosition: newCursorPosition,
                    query: newQuery,
                })
            }

            return { suggestions: noSuggestions }
        })
    }

    private focusInputAndPositionCursor(cursorPosition: number): void {
        if (this.inputElement.current) {
            this.inputElement.current.focus()
            this.inputElement.current.setSelectionRange(cursorPosition, cursorPosition)
        }
    }

    private focusInputAndPositionCursorAtEnd(): void {
        if (this.inputElement.current) {
            this.focusInputAndPositionCursor(this.inputElement.current.value.length)
        }
    }

    private onInputChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        if (!this.hasLoggedFirstInput) {
            eventLogger.log('SearchInitiated')
            this.hasLoggedFirstInput = true
        }
        this.inputValues.next({
            query: event.currentTarget.value,
            cursorPosition: event.currentTarget.selectionStart || 0,
        })
    }
}
