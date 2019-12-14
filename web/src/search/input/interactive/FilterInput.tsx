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
    startWith,
    share,
    delay,
} from 'rxjs/operators'
import { createSuggestion, Suggestion, SuggestionItem, FiltersSuggestionTypes } from '../Suggestion'
import { fetchSuggestions } from '../../backend'
import { ComponentSuggestions, noSuggestions, typingDebounceTime, focusQueryInput } from '../QueryInput'
import { isDefined } from '../../../../../shared/src/util/types'
import Downshift from 'downshift'
import { generateFiltersQuery } from '../helpers'
import { QueryState, interactiveFormatQueryForFuzzySearch } from '../../helpers'
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
    /**
     * Whether the input is currently focused. Used for styling.
     */
    inputFocused: boolean
    /** Only show suggestions if search input is focused */
    showSuggestions: boolean
    /** The suggestions shown to the user */
    suggestions: typeof LOADING | ComponentSuggestions
}

/**
 * The filter chips for each filter added to the query in interactive mode. The filter input can be either editable or non-editable.
 * If it's in an editable state, it consists of a text input field, with suggestions, and a close button. Otherwise, it's a simple
 * button with a close button.
 */
export default class FilterInput extends React.Component<Props, State> {
    private subscriptions = new Subscription()
    private inputValues = new Subject<string>()
    private componentUpdates = new Subject<Props>()
    private inputEl = React.createRef<HTMLInputElement>()
    /** Emits when the suggestions are hidden */
    private suggestionsHidden = new Subject<void>()

    constructor(props: Props) {
        super(props)

        this.state = {
            inputFocused: document.activeElement === this.inputEl.current,
            suggestions: noSuggestions,
            showSuggestions: false,
        }

        this.subscriptions.add(this.inputValues.subscribe(query => this.props.onFilterEdited(this.props.mapKey, query)))

        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    debounceTime(typingDebounceTime),
                    distinctUntilChanged(
                        (previous, current) => dedupeWhitespace(previous.value) === dedupeWhitespace(current.value)
                    ),
                    switchMap(props => {
                        if (props.value.length === 0) {
                            return [{ suggestions: noSuggestions }]
                        }
                        const filterType = props.filterType
                        let fullQuery = `${props.navbarQuery.query} ${generateFiltersQuery({
                            ...props.filtersInQuery,
                        })}`

                        fullQuery = interactiveFormatQueryForFuzzySearch(fullQuery, filterType, props.value)
                        const suggestions = fetchSuggestions(fullQuery).pipe(
                            map(createSuggestion),
                            filter(isDefined),
                            map((suggestion): Suggestion => ({ ...suggestion, fromFuzzySearch: true })),
                            filter(suggestion => suggestion.type === filterType),
                            toArray(),
                            map(suggestions => ({
                                suggestions: {
                                    values: suggestions,
                                    cursorPosition: this.props.value.length,
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

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }
    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private onClickSelected = (): void => {
        if (this.inputEl.current) {
            this.inputEl.current.focus()
        }
        this.props.toggleFilterEditable(this.props.mapKey)
    }

    private onClickDelete = (): void => {
        this.props.onFilterDeleted(this.props.mapKey)
    }

    private onInputUpdate: React.ChangeEventHandler<HTMLInputElement> = e => {
        this.inputValues.next(e.target.value)
    }

    private onSubmitInput = (e: React.FormEvent<HTMLFormElement>): void => {
        e.preventDefault()
        e.stopPropagation()

        this.props.onSubmit(e)
        focusQueryInput.next()
    }

    private onSuggestionSelect = (suggestion: Suggestion | undefined): void => {
        // Insert value into filter input. For any suggestion selected, the whole value should be updated,
        // not just appended.
        if (suggestion) {
            this.inputValues.next(suggestion.value)
        }

        this.setState({ suggestions: noSuggestions, showSuggestions: false }, () => this.suggestionsHidden.next())
    }

    private onInputFocus = (): void => this.setState({ inputFocused: true })

    private onInputBlur = (e: React.FocusEvent<HTMLDivElement>): void => {
        const focusIsNotChildElement = this.focusInCurrentTarget(e)
        if (focusIsNotChildElement) {
            return
        }

        if (this.props.value === '') {
            // Don't allow empty filters
            this.onClickDelete()
            return
        }

        this.props.toggleFilterEditable(this.props.mapKey)
        this.setState({ inputFocused: false, suggestions: noSuggestions })
    }

    /**
     * Checks that the newly focused element is not a child of the previously focused element.
     * Prevents onBlur from firing if we are clicking inside the filter input chip.
     * https://stackoverflow.com/a/38019906
     */
    private focusInCurrentTarget = (e: React.FocusEvent<HTMLDivElement>): boolean => {
        const { relatedTarget, currentTarget } = e
        if (relatedTarget === null) {
            return false
        }

        let node = (relatedTarget as HTMLElement).parentNode

        while (node !== null) {
            if (node === currentTarget) {
                return true
            }
            node = node.parentNode
        }

        return false
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
    }

    private downshiftItemToString = (suggestion?: Suggestion): string => (suggestion ? suggestion.value : '')

    public render(): JSX.Element | null {
        const suggestionsAreLoading = this.state.suggestions === LOADING
        const showSuggestions =
            this.state.showSuggestions &&
            ((this.state.suggestions !== LOADING && this.state.suggestions.values.length > 0) || suggestionsAreLoading)

        return (
            <div
                className={`filter-input ${this.state.inputFocused ? 'filter-input--active' : ''} e2e-filter-input-${
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
                                                    value={this.props.value}
                                                    onChange={this.onInputUpdate}
                                                    placeholder={`${startCase(this.props.filterType)} filter`}
                                                    onKeyDown={event => {
                                                        this.onInputKeyDown(event)
                                                        onKeyDown(event)
                                                    }}
                                                    autoFocus={true}
                                                    onFocus={this.onInputFocus}
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
                                                onClick={this.onClickDelete}
                                                className="btn btn-icon icon-inline"
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
                            className="filter-input__button-text btn text-nowrap"
                            onClick={this.onClickSelected}
                            tabIndex={0}
                        >
                            {this.props.filterType}:{this.props.value}
                        </button>
                        <button
                            type="button"
                            onClick={this.onClickDelete}
                            className={`btn btn-icon icon-inline e2e-filter-input__close-button-${this.props.mapKey}`}
                        >
                            <CloseIcon />
                        </button>
                    </div>
                )}
            </div>
        )
    }
}
