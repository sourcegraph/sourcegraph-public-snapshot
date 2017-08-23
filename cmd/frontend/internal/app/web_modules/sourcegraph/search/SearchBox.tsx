
import escapeRegexp = require('escape-string-regexp');
import * as React from 'react';
import AutoSuggest = require('react-autosuggest');
import { Subject } from 'rxjs';
import { Subscription } from 'rxjs/Subscription';
import { queryGraphQL } from 'sourcegraph/backend/graphql';
import { getSearchPath } from './index';

/** Object representation of a suggestion in the suggestion list */
interface Suggestion {
    /** The label that is displayed to the user. The `query.text` part will be highlighted inside it */
    label: string;
}

/** Object representation of a typed query */
interface Query {
    /** The filters, e.g. `['repo:gorilla/mux']` */
    filters: string[];
    /** The typed text in the input that is not a filter */
    text: string;
}

/** Takes a string query and parses it into an object */
function parseQuery(query: string): Query {
    const words = query.split(' ').filter(s => s);
    const text = words.pop() || '';
    return { filters: words, text };
}

/** Takes a parsed query and encodes it into a string */
function stringifyQuery(query: Query): string {
    return [...query.filters, query.text].filter(s => s).join(' ');
}

interface Props {}

interface State {
    /** The suggestions shown to the user */
    suggestions: Suggestion[];
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
                .map(query => parseQuery(query).text)
                .debounceTime(200)
                .switchMap(text => {
                    const [filterText/*, filterType*/] = text.split(':').reverse();
                    if (filterText.length <= 1) {
                        this.setState({ suggestions: [] });
                        return [];
                    }
                    // Search repos
                    return queryGraphQL(`
                        query SearchRepos($query: String!) {
                            root {
                                repositories(query: $query, fast: true) {
                                    uri
                                    description
                                    private
                                    fork
                                    starsCount
                                    forksCount
                                    language
                                    pushedAt
                                }
                            }
                        }
                    `, { query: text });
                })
                .subscribe(result => {
                    console.error(...result.errors || []);
                    const suggestions: Suggestion[] = [
                        ...result.data!.root.repositories.map(repo => ({
                            label: repo.uri.split('/').slice(1).join('/')
                        }))
                    ];
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
                        placeholder: 'Search'
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

    private getSuggestionValue = (suggestion: Suggestion) => suggestion.label;

    /**
     * Returns a React element for a suggestion
     * Applies a class to the typed-in text
     */
    private renderSuggestionValue = ({ label }: Suggestion, { query }: { query: string }): JSX.Element => {
        const toHighlight = parseQuery(query).text.toLowerCase();
        const parts = label.split(new RegExp(`(${escapeRegexp(toHighlight)})`, 'gi'));
        return <span>{parts.map((part, i) => <span key={i} className={part.toLowerCase() === toHighlight ? 'search-box-highlighted-query' : ''}>{part}</span>)}</span>;
    }

    /**
     * Called when a suggestion is selected by pressing enter or clicking it
     */
    private onSuggestionSelected = (event: React.SyntheticEvent<any>, { suggestion, method }: { suggestion: Suggestion, method: 'enter' | 'click' }): void => {
        const { filters } = parseQuery(this.state.query);
        filters.push('repo:' + suggestion.label);
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
     * Redirects to the search results page
     */
    private onSubmit = (event: React.FormEvent<HTMLFormElement>): void => {
        const { filters, text } = parseQuery(this.state.query);
        const path = getSearchPath({
            q: text,
            repos: filters.map(f => 'github.com/' + f.split(':')[1]).join(','),
            files: '',
            matchCase: false,
            matchRegex: false,
            matchWord: false
        });
        event.preventDefault();
        location.href = path;
    }
}
