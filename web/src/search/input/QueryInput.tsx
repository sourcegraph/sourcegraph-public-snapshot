import * as H from 'history'
import * as React from 'react'
import { fromEvent, Subject, Subscription } from 'rxjs'
import {
    debounceTime,
    distinctUntilChanged,
    filter,
    repeat,
    startWith,
    switchMap,
    takeUntil,
    tap,
} from 'rxjs/operators'
import { eventLogger } from '../../tracking/eventLogger'
import { scrollIntoView } from '../../util'
import { Suggestion, SuggestionItem, SuggestionTypes } from './Suggestion'
import RegexpToggle from './RegexpToggle'
import { SearchPatternType } from '../../../../shared/src/graphql/schema'
import { PatternTypeProps } from '..'
import Downshift from 'downshift'
import { getSearchFilterSuggestions, SearchFilterSuggestions } from '../getSearchFilterSuggestions'
import { Key } from 'ts-key-enum'

/**
 * The query input field is clobbered and updated to contain this subject's values, as
 * they are received. This is used to trigger an update; the source of truth is still the URL.
 */
export const queryUpdates = new Subject<string>()

interface Props extends PatternTypeProps {
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
    /** The suggestions shown to the user */
    suggestions: Suggestion[]

    /** All suggestions (some static and some fetched on page load) */
    searchFilterSuggestions: SearchFilterSuggestions | null

    onChangeCursorPosition: number
}

// TODO: format suggestion with createSuggestion
// TODO: check outside click listener to hide suggestions
// TODO: check something that was removed, compare code
export class QueryInput extends React.Component<Props, State> {
    private static FILTER_SEPARATOR = ':'

    private componentUpdates = new Subject<Props>()

    /** Subscriptions to unsubscribe from on component unmount */
    private subscriptions = new Subscription()

    /** Emits new input values */
    private inputValues = new Subject<[string, number]>()

    /** Emits when the suggestions are hidden */
    private suggestionsHidden = new Subject<void>()

    /** Only used for selection and focus management */
    private inputElement?: HTMLInputElement

    /** Used for scrolling suggestions into view while scrolling with keyboard */
    private containerElement?: HTMLDivElement

    /** Only used to keep track if the user has typed a single character into the input field so we can log an event once. */
    private hasLoggedFirstInput = false

    public state: State = {
        suggestions: [],
        searchFilterSuggestions: null,
        onChangeCursorPosition: 0,
    }

    constructor(props: Props) {
        super(props)

        this.subscriptions.add(
            // Trigger new suggestions every time the input field is typed into
            this.inputValues
                .pipe(
                    tap(([query]) => this.props.onChange(query)),
                    distinctUntilChanged(),
                    debounceTime(200),
                    switchMap(values => {
                        this.setState({ onChangeCursorPosition: values[1] })
                        return [{ suggestions: this.getSuggestions(values) }]
                    }),
                    // Abort suggestion display on route change or suggestion hiding
                    takeUntil(this.suggestionsHidden),
                    // But resubscribe afterwards
                    repeat()
                )
                .subscribe((partial: Partial<State>) => {
                    this.setState(state => ({ ...state, ...partial }))
                }, console.error.bind(console))
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
                        this.props.onChange(String(window.getSelection() || ''))
                        if (this.inputElement) {
                            this.inputElement.focus()
                            // Select whole input text
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

        this.subscriptions.add(
            fromEvent<MouseEvent>(window, 'click').subscribe(event => {
                if (!this.containerElement || !this.containerElement.contains(event.target as Node)) {
                    this.hideSuggestions()
                }
            })
        )

        this.subscriptions.add(
            getSearchFilterSuggestions().subscribe(searchFilterSuggestions =>
                this.setState({ searchFilterSuggestions })
            )
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

    public componentDidUpdate(_: Props, prevState: State): void {
        if (prevState.onChangeCursorPosition !== this.state.onChangeCursorPosition) {
            this.focusInputAndPositionCursor(this.state.onChangeCursorPosition)
        }
        this.componentUpdates.next(this.props)
    }

    private scrollIntoView = (node: HTMLElement, menuNode: HTMLElement) => {
        scrollIntoView(menuNode, node)
    }

    private itemToString = (suggestion: Suggestion) => !!suggestion && suggestion.title

    public render(): JSX.Element | null {
        const showSuggestions = !!this.state.suggestions.length

        return (
            <Downshift
                scrollIntoView={this.scrollIntoView}
                onSelect={this.onSuggestionSelect}
                itemToString={this.itemToString}
            >
                {({ getInputProps, getItemProps, getMenuProps, highlightedIndex }) => {
                    const { onChange: downshiftChange, onKeyDown } = getInputProps()
                    return (
                        <div className="query-input2">
                            <div ref={ref => (this.containerElement = ref!)}>
                                <input
                                    className="form-control query-input2__input rounded-left e2e-query-input"
                                    value={this.props.value}
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
                                    ref={ref => (this.inputElement = ref!)}
                                    name="query"
                                    autoComplete="off"
                                />
                                {showSuggestions && (
                                    <ul className="query-input2__suggestions" {...getMenuProps()}>
                                        {this.state.suggestions.map((suggestion, index) => {
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
                                                />
                                            )
                                        })}
                                    </ul>
                                )}
                                <RegexpToggle
                                    {...this.props}
                                    toggled={this.props.patternType === SearchPatternType.regexp}
                                    navbarSearchQuery={this.props.value}
                                />
                            </div>
                        </div>
                    )
                }}
            </Downshift>
        )
    }

    private onInputKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
        // ArrowDown to show all available suggestions
        if (!this.props.value && event.key === Key.ArrowDown) {
            this.setState({ suggestions: this.getSuggestions() })
        }
    }

    private getSuggestions = ([query, cursorPosition]: [string, number] = ['', 0]): Suggestion[] => {
        const { searchFilterSuggestions } = this.state

        if (!searchFilterSuggestions) {
            return []
        }

        const textTillCursor = query.substring(0, cursorPosition)
        const [lastWord] = textTillCursor.match(/([^\s]+)$/) || ['']
        const [_filter, valueSearch] = lastWord.split(':')
        const filter = _filter.replace('-', '') as SuggestionTypes

        if (filter !== SuggestionTypes.filters && (valueSearch || lastWord.endsWith(':'))) {
            const suggestionsToShow = searchFilterSuggestions[filter] || []
            return suggestionsToShow.values.filter(
                suggestion => suggestion.title.slice(0, valueSearch.length) === valueSearch
            )
        }

        return searchFilterSuggestions.filters.values
            .filter(({ title }) => title.slice(0, filter.length) === filter)
            .map(suggestion => ({
                ...suggestion,
                type: SuggestionTypes.filters,
            }))
    }

    private hideSuggestions = () => {
        this.suggestionsHidden.next()
        this.setState({ suggestions: [] })
    }

    private onSuggestionSelect = (suggestion: Suggestion | undefined) => {
        if (typeof suggestion === 'undefined') {
            return this.hideSuggestions()
        }

        // 🚨 PRIVACY: never provide any private data in { code_search: { suggestion: { type } } }.
        /* eventLogger.log('SearchSuggestionSelected', {
            code_search: {
                suggestion: {
                    type: suggestion.type,
                    url: suggestion.url,
                },
            },
        }) */

        console.log(this.state.onChangeCursorPosition)

        // divides input text, adds suggestion, joins new text and sets new cursor position
        const firstPart = this.props.value.substring(0, this.state.onChangeCursorPosition)
        const lastPart = this.props.value.substring(firstPart.length)
        const isValueSuggestion = suggestion.type !== SuggestionTypes.filters
        const separatorIndex = firstPart.lastIndexOf(isValueSuggestion ? QueryInput.FILTER_SEPARATOR : ' ')

        const newFirstPart =
            firstPart.substring(0, separatorIndex + 1) +
            suggestion.title +
            (!isValueSuggestion ? QueryInput.FILTER_SEPARATOR : '')

        const newValue = newFirstPart + lastPart
        const newCursorPosition = newFirstPart.length

        this.props.onChange(newValue)

        this.setState({ onChangeCursorPosition: newCursorPosition })

        if (isValueSuggestion) {
            this.hideSuggestions()
        } else {
            this.setState({ suggestions: this.getSuggestions([newValue, newCursorPosition]) })
        }
    }

    private focusInputAndPositionCursor(cursorPosition: number): void {
        if (this.inputElement) {
            this.inputElement.focus()
            this.inputElement.setSelectionRange(cursorPosition, cursorPosition)
        }
    }

    private focusInputAndPositionCursorAtEnd(): void {
        if (this.inputElement) {
            this.focusInputAndPositionCursor(this.inputElement.value.length)
        }
    }

    private onInputChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        if (!this.hasLoggedFirstInput) {
            eventLogger.log('SearchInitiated')
            this.hasLoggedFirstInput = true
        }
        this.inputValues.next([event.currentTarget.value, event.currentTarget.selectionStart || 0])
    }
}
