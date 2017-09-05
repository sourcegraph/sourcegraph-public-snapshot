
import CloseIcon from '@sourcegraph/icons/lib/Close';
import FileIcon from '@sourcegraph/icons/lib/File';
import FileGlobIcon from '@sourcegraph/icons/lib/FileGlob';
import RepoIcon from '@sourcegraph/icons/lib/Repo';
import RepoGroup from '@sourcegraph/icons/lib/RepoGroup';
import SearchIcon from '@sourcegraph/icons/lib/Search';
import escapeRegexp = require('escape-string-regexp');
import * as H from 'history';
import * as React from 'react';
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/debounceTime';
import 'rxjs/add/operator/do';
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/mergeMap';
import 'rxjs/add/operator/switchMap';
import 'rxjs/add/operator/toArray';
import { Subject } from 'rxjs/Subject';
import { Subscription } from 'rxjs/Subscription';
import { queryGraphQL } from 'sourcegraph/backend/graphql';
import { events } from 'sourcegraph/tracking/events';
import { isSearchResultsPage, parseBlob } from 'sourcegraph/util/url';
import { getSearchParamsFromURL, getSearchPath, parseRepoList } from './index';

/** The type of type:value filters a user can enter */
enum FilterType {
    Repo = 'repo',
    RepoGroup = 'repogroup',
    File = 'file',
    FileGlob = 'fileglob'
}

/** The icons used to show a suggestion for a filter */
const SUGGESTION_ICONS = {
    [FilterType.Repo]: RepoIcon,
    [FilterType.RepoGroup]: RepoGroup,
    [FilterType.File]: FileIcon,
    [FilterType.FileGlob]: FileGlobIcon
};

/** Object representation of a suggestion in the suggestion list */
interface Filter {
    /** The label that is displayed to the user. The `query.text` part will be highlighted inside it */
    value: string;
    /** The type of suggestion (as indicated by an icon) */
    type: FilterType;
}

interface Props {
    history: H.History;
}

interface State {

    /** The selected filters (shown as chips) */
    filters: Filter[];

    /** The suggestions shown to the user */
    suggestions: Filter[];

    /** Index of the currently selected suggestion (-1 if none selected) */
    selectedSuggestion: number;

    /** The text typed in */
    query: string;

    matchCase: boolean;
    matchRegexp: boolean;
    matchWord: boolean;
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
        let matchCase = false;
        let matchWord = false;
        let matchRegexp = false;
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
            matchCase = params.matchCase;
            matchWord = params.matchWord;
            matchRegexp = params.matchRegex;
        } else if (parsedUrl.uri) {
            // Repo page, add repo filter
            filters.push({ type: FilterType.Repo, value: parsedUrl.uri });
            if (parsedUrl.path) {
                // Blob page, add file filter
                filters.push({ type: FilterType.File, value: parsedUrl.path });
            }
        }
        this.state = {
            filters,
            suggestions: [],
            query,
            selectedSuggestion: -1,
            matchCase,
            matchWord,
            matchRegexp
        };
        this.subscriptions.add(
            this.inputValues
                .do(query => this.setState({ query }))
                .debounceTime(200)
                .switchMap(query => {
                    if (query.length <= 1) {
                        this.setState({ suggestions: [], selectedSuggestion: -1 });
                        return [];
                    }
                    // If query includes a wildcard, suggest a file glob filter
                    // TODO suggest repo glob filter (needs server implementation)
                    // TODO verify that the glob matches something server-side,
                    //      only suggest if it does and show number of matches
                    if (/\*|\?/.exec(this.state.query)) {
                        return [[{ type: FilterType.FileGlob, value: this.state.query }]];
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
                                    ... on SearchProfile {
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
                        .do(result => console.error(...result.errors || []))
                        .mergeMap(result => result.data!.root.search)
                        .map((item: GQL.SearchResult): Filter => {
                            switch (item.__typename) {
                                case 'Repository':    return { value: item.uri, type: FilterType.Repo };
                                case 'SearchProfile': return { value: item.name, type: FilterType.RepoGroup };
                                case 'File':          return { value: item.name, type: FilterType.File };
                            }
                        })
                        .toArray()
                        .catch(err => {
                            console.error(err);
                            return [];
                        });
                })
                .subscribe(suggestions => {
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
                <div className='search-box__query'>
                    <div className='search-box__search-icon'><SearchIcon /></div>
                    <div className='search-box__chips'>
                        {
                            this.state.filters.map((filter, i) => {
                                const Icon = SUGGESTION_ICONS[filter.type];
                                return (
                                    <span key={i} className='search-box__chip'>
                                        <Icon />
                                        <span className='search-box__chip-text'>{filter.value}</span>
                                        <button type='button' className='search-box__chip-remove-button' onClick={() => this.removeFilter(i)}>
                                            <CloseIcon />
                                        </button>
                                    </span>
                                );
                            })
                        }
                        <input
                            type='search'
                            className='search-box__input'
                            value={this.state.query}
                            onChange={this.onInputChange}
                            onKeyDown={this.onInputKeyDown}
                            spellCheck={false}
                            autoCapitalize='off'
                            placeholder='Search'
                            autoFocus={true}
                            ref={ref => this.inputElement = ref!}
                        />
                    </div>
                    <label className='search-box__option' title='Match case'>
                        <input type='checkbox' checked={this.state.matchCase} onChange={e => this.setState({ matchCase: e.currentTarget.checked })} /><span>Aa</span>
                    </label>
                    <label className='search-box__option' title='Match whole word'>
                        <input type='checkbox' checked={this.state.matchWord} onChange={e => this.setState({ matchWord: e.currentTarget.checked })} /><span><u>Ab</u></span>
                    </label>
                    <label className='search-box__option' title='Use regular expression'>
                        <input type='checkbox' checked={this.state.matchRegexp} onChange={e => this.setState({ matchRegexp: e.currentTarget.checked })} /><span>.*</span>
                    </label>
                </div>
                <ul className='search-box__suggestions'>
                    {
                        this.state.suggestions.map((suggestion, i) => {
                            const Icon = SUGGESTION_ICONS[suggestion.type];
                            const parts = suggestion.value.split(splitRegexp);
                            let className = 'search-box__suggestion';
                            if (this.state.selectedSuggestion === i) {
                                className += ' search-box__suggestion--selected';
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
                                    <div className='search-box__suggestion-label'>
                                        {parts.map((part, i) => <span key={i} className={part.toLowerCase() === toHighlight ? 'search-box__highlighted-query' : ''}>{part}</span>)}
                                    </div>
                                    <div className='search-box__suggestion-tip' hidden={this.state.selectedSuggestion !== i}><kbd>tab</kbd> to add as filter</div>
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
                repos: filters.filter(f => f.type === FilterType.Repo || f.type === FilterType.RepoGroup).map(f => f.value).join(','),
                files: filters.filter(f => f.type === FilterType.File || f.type === FilterType.FileGlob).map(f => f.value).join(','),
                matchCase: this.state.matchCase,
                matchRegex: this.state.matchRegexp,
                matchWord: this.state.matchWord
            });
            events.SearchSubmitted.log({ code_search: { pattern: query, repos: filters.filter(f => f.type === FilterType.Repo).map(f => f.value) } });
            this.props.history.push(path);
        } else if (filters[0].type === FilterType.Repo) {
            if (filters.length === 1) {
                // Go to repo
                this.props.history.push(`/${filters[0].value}`);
            } else if (filters[1].type === FilterType.File && filters.length === 2 && !/\?|\*/.exec(filters[1].value)) {
                // Go to file
                this.props.history.push(`/${filters[0].value}/-/blob/${filters[1].value}`);
            }
        }
    }
}
