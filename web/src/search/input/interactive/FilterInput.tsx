import * as React from 'react'
import { Form } from '../../../components/Form'
import CloseIcon from 'mdi-react/CloseIcon'
import { Subscription, Subject, merge, of } from 'rxjs'
import {
    distinctUntilChanged,
    switchMap,
    map,
    filter,
    toArray,
    catchError,
    debounceTime,
    takeUntil,
    repeat,
    share,
    delay,
} from 'rxjs/operators'
import { createSuggestion, Suggestion, SuggestionItem, FiltersSuggestionTypes } from '../Suggestion'
import { fetchSuggestions } from '../../backend'
import { ComponentSuggestions, noSuggestions, typingDebounceTime } from '../QueryInput'
import { isDefined } from '../../../../../shared/src/util/types'
import Downshift from 'downshift'
import { generateFiltersQuery } from '../helpers'
import { QueryState, formatInteractiveQueryForFuzzySearch } from '../../helpers'
import { dedupeWhitespace } from '../../../../../shared/src/util/strings'
import { FiltersToTypeAndValue } from '../../../../../shared/src/search/interactive/util'
import { SuggestionTypes } from '../../../../../shared/src/search/suggestions/util'
import { startCase } from 'lodash'
import { searchFilterSuggestions } from '../../searchFilterSuggestions'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { CheckButton } from './CheckButton'

interface Props {
    /**
     * The filters currently added to the query.
     */
    filtersInQuery: FiltersToTypeAndValue

    /**
     * The query in the main query input.
     */
    navbarQuery: QueryState

    /**
     * The key representing this filter in the top-level filtersInQuery map.
     */
    mapKey: string

    /**
     * The value for this filter.
     */
    value: string

    /**
     * The search filter type, as available in {@link SuggstionTypes}
     */
    filterType: SuggestionTypes

    /**
     * Whether or not this FilterInput is currently editable.
     *
     * This is passed as a prop rather than being a state field because
     * this component is unaware whether to render as editable or uneditable
     * on initial mount.
     */
    editable: boolean

    /**
     * Callback that handles a filter input being submitted. Triggers a search
     * with the new query value.
     */
    onSubmit: (e: React.FormEvent<HTMLFormElement>) => void

    /**
     * Callback to handle the filter's value being updated.
     */
    onFilterEdited: (filterKey: string, value: string) => void

    /**
     * Callback to handle the filter chip being deleted.
     */
    onFilterDeleted: (filterKey: string) => void

    /**
     * Callback to handle the editable state of this filter.
     */
    toggleFilterEditable: (filterKey: string) => void
}

const LOADING: 'loading' = 'loading'

interface State {
    /** Only show suggestions if search input is focused */
    showSuggestions: boolean
    /** The suggestions shown to the user */
    suggestions: typeof LOADING | ComponentSuggestions

    /** The value being typed in the filter's input*/
    inputValue: string
}

/**
 * The filter chips for each filter added to the query in interactive mode. The filter input can be either editable or non-editable.
 * If it's in an editable state, it consists of a text input field, with suggestions, and a close button. Otherwise, it's a simple
 * button with a close button.
 */
export class FilterInput extends React.Component<Props, State> {
    private subscriptions = new Subscription()
    private inputValues = new Subject<string>()
    private inputEl = React.createRef<HTMLInputElement>()
    /** Emits when the suggestions are hidden */
    private suggestionsHidden = new Subject<void>()

    constructor(props: Props) {
        super(props)

        this.state = {
            suggestions: noSuggestions,
            showSuggestions: false,
            inputValue: props.value || '',
        }

        this.subscriptions.add(this.inputValues.subscribe(query => this.setState({ inputValue: query })))

        this.subscriptions.add(
            this.inputValues
                .pipe(
                    debounceTime(typingDebounceTime),
                    distinctUntilChanged(
                        (previous, current) => dedupeWhitespace(previous) === dedupeWhitespace(current)
                    ),
                    switchMap(inputValue => {
                        if (inputValue.length === 0) {
                            return [{ suggestions: noSuggestions }]
                        }
                        const filterType = props.filterType
                        const newFiltersQuery = { ...props.filtersInQuery }
                        if (newFiltersQuery[props.mapKey]) {
                            newFiltersQuery[props.mapKey].value = inputValue
                        } else {
                            newFiltersQuery[props.mapKey] = {
                                type: props.filterType,
                                value: inputValue,
                                editable: true,
                            }
                        }

                        let fullQuery = `${props.navbarQuery.query} ${generateFiltersQuery(newFiltersQuery)}`

                        fullQuery = formatInteractiveQueryForFuzzySearch(fullQuery, filterType, inputValue)
                        const suggestions = fetchSuggestions(fullQuery).pipe(
                            map(createSuggestion),
                            filter(isDefined),
                            map((suggestion): Suggestion => ({ ...suggestion, fromFuzzySearch: true })),
                            filter(suggestion => suggestion.type === filterType),
                            toArray(),
                            map(suggestions => ({
                                suggestions: {
                                    values: suggestions,
                                    cursorPosition: fullQuery.length,
                                },
                            })),
                            catchError(error => {
                                console.error(error)
                                return [{ suggestions: noSuggestions }]
                            }),
                            share()
                        )

                        return merge(
                            of({ suggestions: LOADING }).pipe(delay(1000), takeUntil(suggestions)),
                            suggestions
                        )
                    }),
                    takeUntil(this.suggestionsHidden),
                    repeat()
                )
                .subscribe(state => this.setState({ ...state, showSuggestions: true }))
        )
    }

    public componentDidMount(): void {
        if (this.inputEl.current) {
            this.inputEl.current.focus()
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private onInputUpdate: React.ChangeEventHandler<HTMLInputElement> = e => {
        this.inputValues.next(e.target.value)
    }

    private onSubmitInput = (e: React.FormEvent<HTMLFormElement>): void => {
        e.preventDefault()
        e.stopPropagation()

        if (this.state.inputValue !== '') {
            // Don't allow empty filters.
            // Update the top-level filtersInQueryMap with the new value for this filter.
            this.props.onFilterEdited(this.props.mapKey, this.state.inputValue)
        }
    }

    private onSuggestionSelect = (suggestion: Suggestion | undefined): void => {
        // Insert value into filter input. For any suggestion selected, the whole value should be updated,
        // not just appended.
        if (suggestion) {
            this.inputValues.next(suggestion.value)
        }

        this.setState({ suggestions: noSuggestions, showSuggestions: false }, () => this.suggestionsHidden.next())
    }

    /**
     * Handles clicking a filter chip in an uneditable state. Makes the filter editable
     * and focuses the input field.
     */
    private onClickFilterChip = (): void => {
        if (this.inputEl.current) {
            this.inputEl.current.focus()
        }
        this.props.toggleFilterEditable(this.props.mapKey)
    }

    /** Handles clicking the delete button on an uneditable filter chip. */
    private onClickDelete = (): void => {
        this.props.onFilterDeleted(this.props.mapKey)
    }

    /**
     * Handles the input field in the filter input becoming unfocused.
     */
    private onInputBlur: React.FocusEventHandler<HTMLDivElement> = e => {
        const focusIsNotChildElement = this.focusInCurrentTarget(e)
        if (focusIsNotChildElement) {
            return
        }

        this.handleDiscard()
    }

    /**
     * Handles discarding while editing a filter.
     *
     * If the filter had no value, and no new value was submitted, the filter is deleted.
     * If the filter had an old value, and no new value was submitted, the inputValue is reverted
     * to the initial value, and the filter becomes uneditable.
     * Any suggestions get hidden.
     */
    private handleDiscard = (): void => {
        if (this.props.value === '') {
            // Don't allow empty filters
            this.onClickDelete()
            return
        }

        this.props.toggleFilterEditable(this.props.mapKey)

        // Revert to the last locked-in value for this filter, since the user didn't submit their new value.
        this.setState({ suggestions: noSuggestions, inputValue: this.props.value })
    }

    /**
     * Checks that the newly focused element is not a child of the previously focused element.
     * Prevents onBlur from firing if we are clicking inside the filter input chip.
     */
    private focusInCurrentTarget = (e: React.FocusEvent<HTMLDivElement>): boolean => {
        const { relatedTarget, currentTarget } = e
        if (relatedTarget === null) {
            return false
        }

        const node = (relatedTarget as HTMLElement).parentNode
        return currentTarget.contains(node)
    }

    private onInputKeyDown = (event: React.KeyboardEvent<HTMLInputElement>): void => {
        // Ctrl+Space to show all available filter type suggestions
        if (event.ctrlKey && event.key === ' ') {
            this.setState({
                suggestions: {
                    cursorPosition: event.currentTarget.selectionStart ?? 0,
                    values: searchFilterSuggestions[this.props.filterType as FiltersSuggestionTypes].values,
                },
            })
        }

        // Escape to cancel editing a filter
        if (event.key === 'Escape' && this.props.editable) {
            this.handleDiscard()
        }
    }

    private downshiftItemToString = (suggestion?: Suggestion): string => (suggestion ? suggestion.value : '')

    public render(): JSX.Element | null {
        const suggestionsAreLoading = this.state.suggestions === LOADING
        const showSuggestions =
            this.state.showSuggestions &&
            ((this.state.suggestions !== LOADING && this.state.suggestions.values.length > 0) || suggestionsAreLoading)

        return (
            <div
                className={`filter-input ${this.props.editable ? 'filter-input--active' : ''} e2e-filter-input-${
                    this.props.mapKey
                }`}
                onBlur={this.onInputBlur}
            >
                {this.props.editable ? (
                    <Form onSubmit={this.onSubmitInput}>
                        <Downshift onSelect={this.onSuggestionSelect} itemToString={this.downshiftItemToString}>
                            {({ getInputProps, getItemProps, getMenuProps, highlightedIndex }) => {
                                const { onKeyDown } = getInputProps()
                                return (
                                    <div>
                                        <div className="filter-input__form">
                                            <div className="filter-input__input-wrapper">
                                                <input
                                                    ref={this.inputEl}
                                                    className={`form-control filter-input__input-field e2e-filter-input__input-field-${this.props.mapKey}`}
                                                    value={this.state.inputValue}
                                                    onChange={this.onInputUpdate}
                                                    placeholder={`${startCase(this.props.filterType)} filter`}
                                                    onKeyDown={event => {
                                                        this.onInputKeyDown(event)
                                                        onKeyDown(event)
                                                    }}
                                                    autoFocus={true}
                                                />
                                                {showSuggestions && (
                                                    <ul
                                                        className="filter-input__suggestions e2e-filter-input__suggestions"
                                                        {...getMenuProps()}
                                                    >
                                                        {this.state.suggestions === LOADING ? (
                                                            <li className="suggestion suggestion--selected">
                                                                <LoadingSpinner className="icon-inline" />
                                                                <div className="suggestion__description">Loading</div>
                                                            </li>
                                                        ) : (
                                                            this.state.suggestions.values.map((suggestion, index) => {
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
                                                                        showUrlLabel={false}
                                                                    />
                                                                )
                                                            })
                                                        )}
                                                    </ul>
                                                )}
                                            </div>
                                            <CheckButton />
                                            <button
                                                type="button"
                                                onClick={this.handleDiscard}
                                                className={`btn btn-icon icon-inline e2e-filter-input__cancel-button-${this.props.mapKey}`}
                                                aria-label="Cancel"
                                                data-tooltip="Cancel"
                                            >
                                                <CloseIcon />
                                            </button>
                                        </div>
                                    </div>
                                )
                            }}
                        </Downshift>
                    </Form>
                ) : (
                    <div className="filter-input--uneditable d-flex">
                        <button
                            type="button"
                            className={`filter-input__button-text btn text-nowrap e2e-filter-input__button-text-${this.props.mapKey}`}
                            onClick={this.onClickFilterChip}
                            data-tooltip="Edit filter"
                            aria-label="Edit filter"
                            tabIndex={0}
                        >
                            {this.props.filterType}:{this.state.inputValue}
                        </button>
                        <button
                            type="button"
                            onClick={this.onClickDelete}
                            className={`btn btn-icon icon-inline e2e-filter-input__delete-button-${this.props.mapKey}`}
                            aria-label="Delete filter"
                            data-tooltip="Delete filter"
                        >
                            <CloseIcon />
                        </button>
                    </div>
                )}
            </div>
        )
    }
}
