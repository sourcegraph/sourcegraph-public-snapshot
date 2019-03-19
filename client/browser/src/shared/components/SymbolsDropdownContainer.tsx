import { DownshiftState, StateChangeOptions } from 'downshift'
import { inRange, isEqual } from 'lodash'
import * as React from 'react'
import { concat, fromEvent, merge, Observable, of, Subject, Subscription } from 'rxjs'
import {
    catchError,
    debounceTime,
    delay,
    distinctUntilChanged,
    map,
    mapTo,
    share,
    switchMap,
    takeUntil,
} from 'rxjs/operators'
import getCaretCoordinates from 'textarea-caret'
import { Key } from 'ts-key-enum'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { getContext } from '../backend/context'
import { asError, ErrorLike, isErrorLike } from '../backend/errors'
import { fetchSymbols } from '../backend/search'
import { repoUrlCache, sourcegraphUrl } from '../util/context'
import { SymbolsDropdown } from './SymbolsDropdown'

export const LOADING: 'LOADING' = 'LOADING'
export type SymbolFetchResult = typeof LOADING | ErrorLike | GQL.ISymbol[]

function isSuccessfulFetch(result: SymbolFetchResult): result is GQL.ISymbol[] {
    return result !== LOADING && !isErrorLike(result)
}

interface TextBoxState {
    /**
     * The text contents of the text box
     */
    contents: string

    /**
     * The position of the caret (index in string) inside the text Box
     */
    caretPosition: number
}

/**
 * A user can activate inline symbol search by
 * typing a string that looks like '!<queryText>' inside of
 * a GitHub PR comment.
 */
interface SymbolQuery {
    /**
     * The text of the symbol query, minus the '!' trigger text
     * before it.
     *
     * For example, queryText for `!hover` is 'hover'.
     */
    queryText: string

    /**
     * The index for where the string for the current query (including the trigger)
     * starts in the GitHub PR comment box.
     *
     * ...!<queryText>...
     *    ^
     *    (startIndex)
     */
    startIndex: number

    /**
     * The index for where the string for the current query (including the trigger)
     * ends in the GitHub PR comment box.
     *
     * ...!<queryText>...
     *                ^
     *                (endIndex)
     */
    endIndex: number
}

// In order to handle keyboard events correctly, we need to explictly
// manage/control some of the state of Downshift since we have no control
// over the underlying TextArea element.
//
// (See https://github.com/paypal/downshift#control-props for more information.)
export type ManagedDownShiftState = Pick<DownshiftState<GQL.ISymbol>, 'highlightedIndex' | 'selectedItem'>

interface State extends ManagedDownShiftState {
    /**
     * The currently active/selected symbol query, or undefined if
     * no query is active.
     *
     * A query is considered active when the caret is in
     * [activeQuery.startIndex, activeQuery.endIndex].
     */
    activeQuery?: SymbolQuery

    /**
     * The current state of the symbol fetch request for activeQuery,
     * or undefined if there is no activeQuery.
     */
    symbolsOrError?: SymbolFetchResult

    /**
     * Whether or not the user hid the dropdown by clicking outside of it
     */
    hidden: boolean
}

interface Props {
    /**
     * A reference to the text box DOM node that this autocomplete instance
     * is watching
     */
    textBoxRef: HTMLTextAreaElement
}

export class SymbolsDropdownContainer extends React.Component<Props, State> {
    private subscriptions: Subscription

    /**
     * Emits whenever the user clicks on a symbol result in the dropdown menu
     */
    private selectedSymbolUpdates: Subject<GQL.ISymbol>

    public constructor(props: Props) {
        super(props)

        this.state = {
            symbolsOrError: undefined,
            activeQuery: undefined,
            hidden: false,

            // managed state for Downshift
            highlightedIndex: 0, // defaults to '0' so that the first element in the list is selected by default
            selectedItem: null,
        }

        this.subscriptions = new Subscription()

        this.subscriptions.add(
            fromEvent<KeyboardEvent>(this.props.textBoxRef, 'keydown').subscribe(e =>
                this.handleKeyboadNavigationEvents(e)
            )
        )

        /**
         * Emits the current text content and the location of the user's caret inside of
         * the GitHubPR comment box.
         */
        const textBoxStateUpdates: Observable<TextBoxState> = merge(
            fromEvent<KeyboardEvent>(this.props.textBoxRef, 'keyup'),
            fromEvent<KeyboardEvent>(this.props.textBoxRef, 'input')
        ).pipe(
            map(() => ({
                contents: this.props.textBoxRef.value,
                caretPosition: this.props.textBoxRef.selectionStart,
            })),
            debounceTime(50),
            distinctUntilChanged((a, b) => isEqual(a, b))
        )

        // Reset the hidden state whenever the user types anything
        this.subscriptions.add(textBoxStateUpdates.subscribe(() => this.setState({ hidden: false })))

        this.selectedSymbolUpdates = new Subject()

        this.subscriptions.add(
            this.selectedSymbolUpdates.subscribe(selectedSymbol => {
                this.spliceSymbolURL(selectedSymbol)
            })
        )

        // TODO(@ggilmore): refactor symbolQueryUpdates and symbolFetchResult updates into a stream
        // that emits partial state updates (will eliminate the need for both subscriptions)

        /**
         * Emits metadata for the currently active/selected symbol query, or undefined if no query is active
         */
        const symbolQueryUpdates = merge(
            // selectedSymbolUpdates only emits when a user selects a symbol in the dropdown,
            // so there is no an active query since the user just selected a result
            this.selectedSymbolUpdates.pipe(mapTo(undefined)),

            textBoxStateUpdates.pipe(
                map(({ contents, caretPosition }) => this.findContainingQuery(contents, caretPosition)),
                distinctUntilChanged((a, b) => isEqual(a, b))
            )
        ).pipe(share())

        this.subscriptions.add(
            symbolQueryUpdates.subscribe(symbolQuery => {
                this.setState({
                    activeQuery: symbolQuery,

                    // Reset the mangaged Downshift state since the component will be disposed since
                    // the query changed
                    highlightedIndex: 0,
                    selectedItem: null,
                })
            })
        )

        /**
         * Emits the current state of the symbol fetch request for the last value
         * emitted by symbolQueryUpdates, or undefined if there is no active query
         */
        const symbolFetchResultUpdates = symbolQueryUpdates.pipe(
            map(query => query && query.queryText),
            distinctUntilChanged(),
            switchMap(queryText => this.getSymbolResults(queryText))
        )

        this.subscriptions.add(
            symbolFetchResultUpdates.subscribe(symbolFetchResult => {
                this.setState({ symbolsOrError: symbolFetchResult })
            })
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    private onUserSelectedItem = (selectedSymbol: GQL.ISymbol): void => {
        this.selectedSymbolUpdates.next(selectedSymbol)
    }

    /**
     * Returns an observable which emits the symbol results for the provided
     * 'queryText'
     */
    private getSymbolResults(queryText?: string): Observable<SymbolFetchResult | undefined> {
        if (!queryText) {
            // there is no valid result if the query is undefined/empty
            return of(undefined)
        }

        const { repoKey } = getContext()

        const symbolResults = fetchSymbols(`repo:^${repoKey}\$ type:symbol ${queryText}`).pipe(
            catchError(err => {
                console.error(err)
                return [asError(err)]
            }),
            share()
        )

        // 1) emit a 'loading...' message if fetching the symbol results takes more than half a second
        // 2) emit the actual results
        const incrementalResults = merge(
            of(LOADING).pipe(
                delay(500),
                takeUntil(symbolResults)
            ),

            symbolResults
        )

        if (queryText.length === 1) {
            // The user has likely just typed a brand-new autocomplete
            // query. Clear out the results from any previous symbol fetch requests.
            return concat(of(undefined), incrementalResults)
        }

        return incrementalResults
    }

    /**
     * Returns the SymbolQuery in 'text' whose range contains 'targetIndex',
     * or undefined otherwise
     */
    private findContainingQuery(text: string, targetIndex: number): SymbolQuery | undefined {
        for (const query of this.extractSymbolQueries(text)) {
            const { startIndex, endIndex } = query

            if (startIndex <= targetIndex && targetIndex <= endIndex) {
                return query
            }
        }

        return undefined
    }

    /**
     * Returns an array of all the SymbolQueries contained in 'text'
     */
    private extractSymbolQueries(text: string): SymbolQuery[] {
        const out: SymbolQuery[] = []

        const symbolAutoCompleteRegexp = /@!(\w[^\s]*)/g

        let match = symbolAutoCompleteRegexp.exec(text)

        while (match !== null) {
            out.push({
                queryText: match[1],
                startIndex: match.index,
                endIndex: symbolAutoCompleteRegexp.lastIndex,
            })

            match = symbolAutoCompleteRegexp.exec(text)
        }

        return out
    }

    /**
     * Replaces '!<queryText>' inside the text box with selectedSymbol's URL.
     *
     * This is called whenever the user selects a symbol result from the dropdown.
     */
    private spliceSymbolURL(selectedSymbol: GQL.ISymbol): void {
        const { activeQuery } = this.state
        if (!activeQuery) {
            console.error(`Could not splice splice symbol URL for ${selectedSymbol} because activeQuery is undefined.`)
            return
        }

        // `symbol.location.url` looks like '/github.com/sourcegraph/sourcegraph/-/blob/...'
        //                                   ^
        // so, we need to prefix it with the URL for the correct Sourcegraph instance

        const relativeURL = selectedSymbol.location.url
        const baseSourcegraphURL = repoUrlCache[getContext().repoKey] || sourcegraphUrl

        const absoluteURL = new URL(relativeURL, baseSourcegraphURL)
        absoluteURL.searchParams.set('utm_source', 'inline-symbol')

        const symbolMarkdownLink = `[${selectedSymbol.name}](${absoluteURL.toString()})`

        // snip '!<queryText>' from the text box, replace it with the absolute URL that we built above

        const { startIndex, endIndex } = activeQuery

        const textBeforeQuery = this.props.textBoxRef.value.slice(0, startIndex)
        const textAfterQuery = this.props.textBoxRef.value.slice(endIndex)

        this.props.textBoxRef.value = textBeforeQuery + symbolMarkdownLink + textAfterQuery
    }

    private handleKeyboadNavigationEvents(e: KeyboardEvent): void {
        let handled = false

        switch (e.key) {
            case Key.ArrowDown:
                handled = this.scrollDown()
                break
            case Key.ArrowUp:
                handled = this.scrollUp()
                break
            case Key.Enter:
                handled = this.selectHighlightedItem()
                break
            case Key.Escape:
                this.hideDropdown()
                break
        }

        if (handled) {
            // only call e.preventDefault() if the corresponding handler actually
            // did something
            e.preventDefault()
        }
    }

    /**
     * Scrolls the list in the dropdown down by one entry, wrapping around if necessary.
     *
     * Returns 'true' iff the list was actually scrolled (after checking to see if the current state is valid)
     */
    private scrollDown(): boolean {
        return this.scrollDelta(1)
    }

    /**
     * Scrolls the list in the dropdown up by one entry, wrapping around if necessary.
     *
     * Returns 'true' iff the list was actually scrolled (after checking to see if the current state is valid)
     */
    private scrollUp(): boolean {
        return this.scrollDelta(-1)
    }

    /**
     * Scrolls the list in the dropdown by 'delta' entries, wrapping around if necessary.
     *
     * Returns 'true' iff the list was actually scrolled (after checking to see if the current state is valid)
     */
    private scrollDelta(delta: number): boolean {
        const { symbolsOrError, hidden } = this.state

        if (hidden || !symbolsOrError || !isSuccessfulFetch(symbolsOrError)) {
            // There is nothing to scroll to if there aren't symbol results for the query.
            return false
        }

        const oldIndex = this.state.highlightedIndex || 0
        const newIndex = oldIndex + delta

        // wrap around to the other end of the list if necessary
        const wrappedIndex = ((newIndex % symbolsOrError.length) + symbolsOrError.length) % symbolsOrError.length

        this.setState({ highlightedIndex: wrappedIndex })
        return true
    }

    /**
     * Selects the currently highlighted entry in the dropdown.
     *
     * Returns 'true' iff an entry was actually selected (after checking to see if the current state is valid)
     */
    private selectHighlightedItem(): boolean {
        const { symbolsOrError, hidden } = this.state

        if (hidden || !symbolsOrError || !isSuccessfulFetch(symbolsOrError)) {
            // There is nothing to select if there aren't symbol results for the query.
            return false
        }

        const currentIndex = this.state.highlightedIndex || 0

        if (!inRange(currentIndex, 0, symbolsOrError.length)) {
            // There is nothing to select if the index is outside the indicies of
            // the fetched symbol results
            return false
        }

        const selectedItem = symbolsOrError[currentIndex]
        this.setState({ selectedItem })
        this.onUserSelectedItem(selectedItem)

        return true
    }

    private hideDropdown = (): void => {
        this.setState({ hidden: true })
    }

    /**
     * Updates our copy of Downshift's state whenever Downshift itself changes it
     */
    private onDownshiftStateChange = ({ highlightedIndex, selectedItem }: StateChangeOptions<GQL.ISymbol>): void => {
        if (highlightedIndex !== undefined) {
            this.setState({ highlightedIndex })
        }

        if (selectedItem !== undefined) {
            this.setState({ selectedItem })
        }
    }

    public render(): JSX.Element | null {
        const { symbolsOrError, activeQuery, hidden, highlightedIndex, selectedItem } = this.state

        if (hidden || !activeQuery || !symbolsOrError) {
            return null
        }

        const caretCoordinates = getCaretCoordinates(this.props.textBoxRef, activeQuery.startIndex)

        return (
            <div className="symbols-dropdown-container">
                <SymbolsDropdown
                    symbolsOrError={symbolsOrError}
                    onSymbolSelected={this.onUserSelectedItem}
                    onClickOutside={this.hideDropdown}
                    onDownshiftStateChange={this.onDownshiftStateChange}
                    highlightedIndex={highlightedIndex}
                    selectedItem={selectedItem}
                    caretCoordinates={caretCoordinates}
                />
            </div>
        )
    }
}
