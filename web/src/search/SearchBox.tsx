
import BookClosed from '@sourcegraph/icons/lib/BookClosed';
import CloseIcon from '@sourcegraph/icons/lib/Close';
import Document from '@sourcegraph/icons/lib/Document';
import SearchIcon from '@sourcegraph/icons/lib/Search';
import escapeRegexp = require('escape-string-regexp');
import * as React from 'react';
import { Subject } from 'rxjs';
import { Subscription } from 'rxjs/Subscription';
import { queryGraphQL } from 'sourcegraph/backend/graphql';
import { events } from 'sourcegraph/tracking/events';
import { isSearchResultsPage, parseBlob } from 'sourcegraph/util/url';
import { getSearchParamsFromURL, getSearchPath, parseRepoList } from './index';

/** The type of type:value filters a user can enter */
enum FilterType {
    Repo = 'repo',
    File = 'file'
}

/** The icons used to show a suggestion for a filter */
const SUGGESTION_ICONS = {
    [FilterType.Repo]: BookClosed,
    [FilterType.File]: Document
};

/** Object representation of a suggestion in the suggestion list */
interface Filter {
    /** The label that is displayed to the user. The `query.text` part will be highlighted inside it */
    value: string;
    /** The type of suggestion (as indicated by an icon) */
    type: FilterType;
}

interface Props {}

interface State {

    /** The selected filters (shown as chips) */
    filters: Filter[];

    /** The suggestions shown to the user */
    suggestions: Filter[];

    /** Index of the currently selected suggestion (-1 if none selected) */
    selectedSuggestion: number;

    /** The text typed in */
    query: string;
}

export class SearchBox extends React.Component<Props, State> {

    /** Subscriptions to unsubscribe from on component unmount */
    private subscriptions = new Subscription();

    /** Emits new input values */
    private inputValues = new Subject<string>();

    /** Only used for selection and focus management */
    private inputElement: HTMLInputElement;

    constructor(props: Props) {
        super(props);
        // Fill text input from URL info
        const filters: Filter[] = [];
        let query = '';
        const parsedUrl = parseBlob();
        if (isSearchResultsPage()) {
            // Search results page, show query
            const params = getSearchParamsFromURL(location.href);
            for (const uri of parseRepoList(params.repos)) {
                filters.push({ type: FilterType.Repo, value: uri });
            }
            if (params.files) {
                filters.push({ type: FilterType.File, value: params.files });
            }
            query = params.q;
        } else if (parsedUrl.uri) {
            // Repo page, add repo filter
            filters.push({ type: FilterType.Repo, value: parsedUrl.uri });
            if (parsedUrl.path) {
                // Blob page, add file filter
                filters.push({ type: FilterType.File, value: parsedUrl.path });
            }
        }
        this.state = { filters, suggestions: [], query, selectedSuggestion: -1 };
        this.subscriptions.add(
            this.inputValues
                .do(query => this.setState({ query }))
                .debounceTime(200)
                .switchMap(query => {
                    if (query.length <= 1) {
                        this.setState({ suggestions: [], selectedSuggestion: -1 });
                        return [];
                    }
                    // Search repos
                    return queryGraphQL(`
                        query SearchRepos($query: String!) {
                            root {
                                search(query: $query, repositories: $repositories) {
                                    ... on Repository {
                                        __typename
                                        uri
                                    }
                                    ... on File {
                                        __typename
                                        name
                                    }
                                }
                            }
                        }
                    `, {
                        query,
                        repositories: this.state.filters.filter(f => f.type === FilterType.Repo).map(f => f.value)
                    })
                        .catch(err => {
                            console.error(err);
                            return [];
                        });
                })
                .subscribe(result => {
                    console.error(...result.errors || []);
                    const suggestions = result.data!.root.search.map(item => {
                        switch (item.__typename) {
                            case 'Repository': return { value: item.uri, type: FilterType.Repo };
                            case 'File':       return { value: item.name, type: FilterType.File };
                        }
                    }).filter(s => s) as Filter[];
                    this.setState({ suggestions, selectedSuggestion: Math.min(suggestions.length - 1, 0) });
                }, err => {
                    console.error(err);
                })
        );
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe();
    }

    public render(): JSX.Element | null {
        const toHighlight = this.state.query.toLowerCase();
        const splitRegexp = new RegExp(`(${escapeRegexp(toHighlight)})`, 'gi');
        return (
            <form className='search-box' onSubmit={this.onSubmit}>
                <div className='search-box-query'>
                    <div className='search-box-query-search-icon'><SearchIcon /></div>
                    <div className='search-box-query-chips'>
                        {
                            this.state.filters.map((filter, i) => {
                                const Icon = SUGGESTION_ICONS[filter.type];
                                return (
                                    <span key={i} className='search-box-query-chips-chip'>
                                        <Icon />
                                        <span className='search-box-query-chips-chip-text'>{filter.value}</span>
                                        <button type='button' className='search-box-query-chips-chip-remove-button' onClick={() => this.removeFilter(i)}>
                                            <CloseIcon />
                                        </button>
                                    </span>
                                );
                            })
                        }
                        <input
                            type='search'
                            className='search-box-query-chips-input'
                            value={this.state.query}
                            onChange={this.onInputChange}
                            onKeyDown={this.onInputKeyDown}
                            spellCheck={false}
                            autoCapitalize='off'
                            placeholder='Search'
                            ref={ref => this.inputElement = ref!}
                        />
                    </div>
                    <button className='search-box-option'>Aa</button>
                    <button className='search-box-option'><u>Ab</u></button>
                    <button className='search-box-option'>.*</button>
                </div>
                <ul className='search-box-suggestions'>
                    {
                        this.state.suggestions.map((suggestion, i) => {
                            const Icon = SUGGESTION_ICONS[suggestion.type];
                            const parts = suggestion.value.split(splitRegexp);
                            let className = 'search-box-suggestions-item';
                            if (this.state.selectedSuggestion === i) {
                                className += ' search-box-suggestions-item-selected';
                            }
                            return (
                                <li key={i} className={className} onClick={() => {
                                    this.setState({
                                        filters: this.state.filters.concat(suggestion),
                                        suggestions: [],
                                        selectedSuggestion: -1,
                                        query: ''
                                    });
                                }}>
                                    <Icon />
                                    <div className='search-box-suggestions-item-label'>
                                        {parts.map((part, i) => <span key={i} className={part.toLowerCase() === toHighlight ? 'search-box-highlighted-query' : ''}>{part}</span>)}
                                    </div>
                                    <div className='search-box-suggestions-item-tip' hidden={this.state.selectedSuggestion !== i}><kbd>tab</kbd> to add as filter</div>
                                </li>
                            );
                        })
                    }
                </ul>
            </form>
        );
    }

    private removeFilter(index: number): void {
        const { filters } = this.state;
        filters.splice(index, 1);
        this.setState({ filters });
    }

    private onInputChange: React.ChangeEventHandler<HTMLInputElement> = event => {
        this.inputValues.next(event.target.value);
    }

    private onInputKeyDown: React.KeyboardEventHandler<HTMLInputElement> = event => {
        switch (event.key) {
            case 'ArrowDown': {
                event.preventDefault();
                this.moveSelection(1);
                break;
            }
            case 'ArrowUp': {
                event.preventDefault();
                this.moveSelection(-1);
                break;
            }
            case 'Tab': {
                if (this.state.selectedSuggestion > -1) {
                    event.preventDefault();
                    this.setState({
                        filters: this.state.filters.concat(this.state.suggestions[this.state.selectedSuggestion]),
                        suggestions: [],
                        selectedSuggestion: -1,
                        query: ''
                    });
                }
                break;
            }
            case 'Backspace': {
                if (this.inputElement.selectionStart === 0 && this.inputElement.selectionEnd === 0) {
                    this.setState({ filters: this.state.filters.slice(0, -1) });
                }
                break;
            }
        }
    }

    private moveSelection(steps: number): void {
        this.setState({ selectedSuggestion: Math.max(Math.min(this.state.selectedSuggestion + steps, this.state.suggestions.length - 1), -1) });
    }

    /**
     * Called when the user submits the form (by pressing Enter)
     * If only one repo was selected and no query typed, redirects to the repo page
     * Otherwise redirects to the search results page
     */
    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault();
        const { filters, query } = this.state;
        if (query && filters.length > 0) {
            // Go to search results
            const path = getSearchPath({
                q: query,
                repos: filters.filter(f => f.type === FilterType.Repo).map(f => f.value).join(','),
                files: filters.filter(f => f.type === FilterType.File).map(f => f.value).join(','),
                matchCase: false,
                matchRegex: false,
                matchWord: false
            });
            events.SearchSubmitted.log({ code_search: { pattern: query, repos: filters.filter(f => f.type === FilterType.Repo).map(f => f.value) } });
            location.href = path;
        } else if (filters[0].type === FilterType.Repo) {
            if (filters.length === 1) {
                // Go to repo
                location.href = `/${filters[0].value}`;
            } else if (filters[1].type === FilterType.File && filters.length === 2 && !/\?|\*/.exec(filters[1].value)) {
                // Go to file
                location.href = `/${filters[0].value}/-/blob/${filters[1].value}`;
            }
        }
    }
}
