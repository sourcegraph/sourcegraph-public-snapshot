
import BookClosed from '@sourcegraph/icons/lib/BookClosed';
import Document from '@sourcegraph/icons/lib/Document';
import escapeRegexp = require('escape-string-regexp');
import * as React from 'react';
import AutoSuggest = require('react-autosuggest');
import { Subject } from 'rxjs';
import { Subscription } from 'rxjs/Subscription';
import { queryGraphQL } from 'sourcegraph/backend/graphql';
import { getSearchPath } from './index';

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

/** Object representation of a typed query */
interface Query {
    /** The type:value filters typed in */
    filters: Filter[];
    /** The typed text in the input that is not a filter */
    text: string;
}

/** Takes a string query and parses it into an object */
function parseQuery(query: string): Query {
    const words = query.split(/\s+/).filter(s => s);
    let text = words.pop() || '';
    if (text.includes(':')) {
        words.push(text);
        text = '';
    }
    const filters: Filter[] = [];
    for (const word of words) {
        const [type, value] = word.split(':') as [FilterType | undefined, string | undefined];
        if ((type !== FilterType.File && type !== FilterType.Repo) || !value) {
            console.warn(`Invalid query filter: ${word}`);
            continue;
        }
        filters.push({ type, value });
    }
    return { filters, text };
}

/** Takes a parsed query and encodes it into a string */
function stringifyQuery(query: Query): string {
    return [...query.filters.map(f => f.type + ':' + f.value), query.text].filter(s => s).join(' ');
}

interface Props {}

interface State {
    /** The suggestions shown to the user */
    suggestions: Filter[];
    /** The whole content of the search input (a stringified query) */
    query: string;
}

export class SearchBox extends React.Component<Props, State> {

    public state: State = {
        suggestions: [],
        query: ''
    };

    /** Subscriptions to unsubscribe from on component unmount */
    private subscriptions = new Subscription();

    /** Emits queries to fetch suggestions for */
    private suggestionsFetchRequests = new Subject<string>();

    constructor(props: Props) {
        super(props);
        this.subscriptions.add(
            this.suggestionsFetchRequests
                .map(parseQuery)
                .debounceTime(200)
                .switchMap(({ filters, text }) => {
                    if (text.length <= 1) {
                        this.setState({ suggestions: [] });
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
                        query: text,
                        repositories: filters.filter(f => f.type === FilterType.Repo).map(f => f.value)
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
                    this.setState({ suggestions });
                }, err => {
                    console.error(err);
                })
        );
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe();
    }

    public render(): JSX.Element | null {
        return (
            <form className='search-box' onSubmit={this.onSubmit}>
                <AutoSuggest
                    suggestions={this.state.suggestions}
                    onSuggestionsFetchRequested={this.onSuggestionsFetchRequested}
                    onSuggestionsClearRequested={this.onSuggestionsClearRequested}
                    getSuggestionValue={this.getSuggestionValue}
                    renderSuggestion={this.renderSuggestionValue}
                    shouldRenderSuggestions={this.shouldRenderSuggestion}
                    highlightFirstSuggestion={false}
                    onSuggestionSelected={this.onSuggestionSelected}
                    inputProps={{
                        value: this.state.query,
                        onChange: this.onChange,
                        placeholder: 'Search',
                        spellCheck: false
                    }}
                />
            </form>
        );
    }

    private onSuggestionsClearRequested = () => {
        this.setState({ suggestions: [] });
    }

    private shouldRenderSuggestion = (query: string): boolean => parseQuery(query).text.length > 1;

    /** Called whenever the suggestions should be refetched */
    private onSuggestionsFetchRequested = ({ value }: { value: string }): void => this.suggestionsFetchRequests.next(value);

    private getSuggestionValue = (suggestion: Filter) => suggestion.value;

    /**
     * Returns a React element for a suggestion
     * Applies a class to the typed-in text
     */
    private renderSuggestionValue = ({ value, type }: Filter, { query }: { query: string }): JSX.Element => {
        const toHighlight = parseQuery(query).text.toLowerCase();
        const parts = value.split(new RegExp(`(${escapeRegexp(toHighlight)})`, 'gi'));
        const Icon = SUGGESTION_ICONS[type];
        return (
            <span>
                <Icon />
                {parts.map((part, i) => <span key={i} className={part.toLowerCase() === toHighlight ? 'search-box-highlighted-query' : ''}>{part}</span>)}
            </span>
        );
    }

    /**
     * Called when a suggestion is selected by pressing enter or clicking it
     */
    private onSuggestionSelected = (event: React.SyntheticEvent<any>, { suggestion, method }: { suggestion: Filter, method: 'enter' | 'click' }): void => {
        const { filters } = parseQuery(this.state.query);
        filters.push(suggestion);
        this.setState({ query: stringifyQuery({ filters, text: '' }) + ' ', suggestions: [] });
        if (method === 'enter') {
            // Prevent a form submit
            event.preventDefault();
        }
    }

    /**
     * Called when the input changes to update the state
     */
    private onChange = (event: any, { newValue, method }: { newValue: string, method: 'down' | 'up' | 'enter' | 'click' | 'escape' | 'type' }): void => {
        switch (method) {
            // Don't override the query when the user only highlights the selection
            case 'down':
            case 'up':
                return;
            default:
                this.setState({ query: newValue });
        }
    }

    /**
     * Called when the user submits the form (by pressing Enter)
     * If only one repo was selected and no query typed, redirects to the repo page
     * Otherwise redirects to the search results page
     */
    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        event.preventDefault();
        const { filters, text } = parseQuery(this.state.query);
        if (text && filters.length > 0) {
            // Go to search results
            const path = getSearchPath({
                q: text,
                repos: filters.filter(f => f.type === FilterType.Repo).map(f => f.value).join(','),
                files: filters.filter(f => f.type === FilterType.File).map(f => f.value).join(','),
                matchCase: false,
                matchRegex: false,
                matchWord: false
            });
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
