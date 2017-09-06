
import CloseIcon from '@sourcegraph/icons/lib/Close';
import FileIcon from '@sourcegraph/icons/lib/File';
import FileGlobIcon from '@sourcegraph/icons/lib/FileGlob';
import RepoIcon from '@sourcegraph/icons/lib/Repo';
import RepoGroupIcon from '@sourcegraph/icons/lib/RepoGroup';
import SearchIcon from '@sourcegraph/icons/lib/Search';
import escapeRegexp = require('escape-string-regexp');
import * as H from 'history';
import * as React from 'react';
import 'rxjs/add/observable/merge';
import 'rxjs/add/operator/catch';
import 'rxjs/add/operator/debounceTime';
import 'rxjs/add/operator/do';
import 'rxjs/add/operator/map';
import 'rxjs/add/operator/mergeMap';
import 'rxjs/add/operator/switchMap';
import 'rxjs/add/operator/toArray';
import { Observable } from 'rxjs/Observable';
import { Subject } from 'rxjs/Subject';
import { Subscription } from 'rxjs/Subscription';
import { queryGraphQL } from 'sourcegraph/backend/graphql';
import { events } from 'sourcegraph/tracking/events';
import { isSearchResultsPage, parseBlob } from 'sourcegraph/util/url';
import { getSearchParamsFromURL, getSearchPath, parseRepoList } from './index';

enum FilterType {
    Repo = 'repo',
    RepoGroup = 'repogroup',
    File = 'file',
    FileGlob = 'fileglob'
}

/** The icons used to show a suggestion for a filter */
const SUGGESTION_ICONS = {
    [FilterType.Repo]: RepoIcon,
    [FilterType.RepoGroup]: RepoGroupIcon,
    [FilterType.File]: FileIcon,
    [FilterType.FileGlob]: FileGlobIcon
};

/** Object representation of a suggestion in the suggestion list */
type Filter = RepoGroupFilter | RepoFilter | FileFilter | FileGlobFilter;

interface BaseFilter {
    label: string;
}

interface RepoFilter extends BaseFilter {
    type: FilterType.Repo;
    repoUri: string;
}

interface FileFilter extends BaseFilter {
    type: FilterType.File;
    filePath: string;
}

interface FileGlobFilter extends BaseFilter {
    type: FilterType.FileGlob;
    glob: string;
}

interface RepoGroupFilter extends BaseFilter {
    type: FilterType.RepoGroup;
    repoUris: string[];
}

interface Props {
    history: H.History;
}

interface State {

    /** The selected filters (shown as chips) */
    filters: Filter[];

    /** Whether suggestions are shown or not */
    suggestionsVisible: boolean;

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

    /** Emits when the input field is clicked */
    private inputClicks = new Subject<void>();

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
            for (const repoUri of parseRepoList(params.repos)) {
                filters.push({ type: FilterType.Repo, label: repoUri, repoUri });
            }
            if (params.files) {
                filters.push({ type: FilterType.File, label: params.files, filePath: params.files });
            }
            query = params.q;
            matchCase = params.matchCase;
            matchWord = params.matchWord;
            matchRegexp = params.matchRegex;
        } else if (parsedUrl.uri) {
            // Repo page, add repo filter
            filters.push({ type: FilterType.Repo, label: parsedUrl.uri, repoUri: parsedUrl.uri });
            if (parsedUrl.path) {
                // Blob page, add file filter
                filters.push({ type: FilterType.File, label: parsedUrl.path, filePath: parsedUrl.path });
            }
        }
        this.state = {
            filters,
            suggestionsVisible: false,
            suggestions: [],
            query,
            selectedSuggestion: -1,
            matchCase,
            matchWord,
            matchRegexp
        };
        this.subscriptions.add(
            Observable.merge(
                // Trigger new suggestions every time the input field is typed into
                this.inputValues
                    .do(query => this.setState({ query }))
                    .debounceTime(200),
                // Trigger new suggestions every time the input field is clicked
                this.inputClicks
                    .map(() => this.inputElement.value)
            )
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
                        const fileGlobFilter: FileGlobFilter = {
                            type: FilterType.FileGlob,
                            label: this.state.query,
                            glob: this.state.query
                        };
                        return [[fileGlobFilter]];
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
                                        repositories {
                                            uri
                                        }
                                    }
                                }
                            }
                        }
                    `, {
                        query,
                        repositories: this.state.filters.filter(f => f.type === FilterType.Repo).map((f: RepoFilter) => f.repoUri)
                    })
                        .do(result => console.error(...result.errors || []))
                        .mergeMap(result => result.data!.root.search)
                        .map((item: GQL.SearchResult): Filter => {
                            switch (item.__typename) {
                                case 'Repository':    return { label: item.uri, type: FilterType.Repo, repoUri: item.uri };
                                case 'SearchProfile': return { label: item.name, type: FilterType.RepoGroup, repoUris: item.repositories.map(r => r.uri) };
                                case 'File':          return { label: item.name, type: FilterType.File, filePath: item.name };
                            }
                        })
                        .toArray()
                        .catch(err => {
                            console.error(err);
                            return [];
                        });
                })
                .subscribe(suggestions => {
                    this.setState({ suggestions, selectedSuggestion: Math.min(suggestions.length - 1, 0), suggestionsVisible: true });
                }, err => {
                    console.error(err);
                })
        );
    }

    public componentDidMount(): void {
        // Focus the input element and set cursor to the end
        this.inputElement.focus();
        this.inputElement.setSelectionRange(this.inputElement.value.length, this.inputElement.value.length);
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
                                        <span className='search-box__chip-text'>{filter.label}</span>
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
                            onBlur={this.onInputBlur}
                            onClick={this.onInputClick}
                            spellCheck={false}
                            autoCapitalize='off'
                            placeholder='Search'
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
                <ul className='search-box__suggestions' style={this.state.suggestionsVisible ? {} : { height: 0 }}>
                    {
                        this.state.suggestions.map((suggestion, i) => {
                            const Icon = SUGGESTION_ICONS[suggestion.type];
                            const parts = suggestion.label.split(splitRegexp);
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
        this.inputValues.next(event.currentTarget.value);
    }

    private onInputBlur: React.FocusEventHandler<HTMLInputElement> = event => {
        this.setState({ suggestionsVisible: false });
    }

    private onInputClick: React.MouseEventHandler<HTMLInputElement> = event => {
        this.inputClicks.next();
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
        this.setState({ suggestionsVisible: false });
        const { filters, query } = this.state;
        if (query && filters.length > 0) {
            // Go to search results
            const path = getSearchPath({
                q: query,
                repos: filters.filter(f => f.type === FilterType.Repo).map((f: RepoFilter) => f.repoUri)
                    .concat(...filters.filter(f => f.type === FilterType.RepoGroup).map((f: RepoGroupFilter) => f.repoUris))
                    .join(','),
                files: [
                    ...filters.filter(f => f.type === FilterType.File).map((f: FileFilter) => f.filePath),
                    ...filters.filter(f => f.type === FilterType.FileGlob).map((f: FileGlobFilter) => f.glob)
                ].join(','),
                matchCase: this.state.matchCase,
                matchRegex: this.state.matchRegexp,
                matchWord: this.state.matchWord
            });
            events.SearchSubmitted.log({ code_search: { pattern: query, repos: filters.filter(f => f.type === FilterType.Repo).map((f: RepoFilter) => f.repoUri) } });
            this.props.history.push(path);
        } else if (filters[0].type === FilterType.Repo) {
            if (filters.length === 1) {
                // Go to repo
                this.props.history.push(`/${(filters[0] as RepoFilter).repoUri}`);
            } else if (filters[1].type === FilterType.File && filters.length === 2) {
                // Go to file
                this.props.history.push(`/${(filters[0] as RepoFilter).repoUri}/-/blob/${(filters[1] as FileFilter).filePath}`);
            }
        }
    }
}
