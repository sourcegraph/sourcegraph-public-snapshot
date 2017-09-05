import * as React from 'react';
import { SearchResult, searchText } from 'sourcegraph/backend';
import { ReferencesGroup } from 'sourcegraph/references/ReferencesWidget';
import { getSearchParamsFromURL, parseRepoList } from 'sourcegraph/search';
import * as activeRepos from 'sourcegraph/util/activeRepos';
import * as URI from 'urijs';

interface Props { }

interface State {
    results: SearchResult[];
    loading: boolean;
    searchDuration?: number;
}

function numberWithCommas(x: any): string {
    return x.toString().replace(/\B(?=(\d{3})+(?!\d))/g, ',');
}

export class SearchResults extends React.Component<Props, State> {

    constructor(props: Props) {
        super(props);
        const params = getSearchParamsFromURL(window.location.href);
        const q = params.q;
        this.state = {
            results: [],
            loading: true
        };

        // Clean the comma delimited input (remove whitespace / duplicate commas).
        //
        // See https://stackoverflow.com/a/13306993
        let repos = params.repos.replace(/^[,\s]+|[,\s]+$/g, '');
        repos = repos.replace(/\s*,\s*/g, ',');

        // Split the list of repos, and create "active" and "inactive"
        // booleans + remove them from the list.
        const repoList: string[] = [];
        let addActive = false;
        let addInactive = false;
        for (const repo of parseRepoList(repos)) {
            if (repo === 'active') {
                addActive = true;
                continue;
            }
            if (repo === 'inactive') {
                addActive = true;
                addInactive = true;
                continue;
            }
            repoList.push(repo);
        }

        const start = Date.now();
        const search = searchReposList => {
            searchText(q, searchReposList.map(repo => ({ repo, rev: '' })), params).then(res => {
                const searchDuration = Date.now() - start;
                if (res.results) {
                    this.setState({ results: res.results, loading: false, searchDuration });
                }
            }).catch(e => {
                // TODO(slimsag): display error in UX
                console.error('failed to search text', e);
            });
        };

        // If we need to add active or inactive repositories to the list, do so
        // inside the promise:
        if (addActive || addInactive) {
            activeRepos.get().then(r => {
                if (addActive) {
                    for (const active of r.active) {
                        repoList.push(active);
                    }
                }
                if (addInactive) {
                    for (const inactive of r.inactive) {
                        repoList.push(inactive);
                    }
                }
                search(repoList);
            }).catch(error => {
                // TODO: actually tell the user about the error.
                console.error('failed to get active repos:', error);
                this.setState({ loading: false });
            });
        } else {
            // Don't need to add active or inactive repositories, so perform
            // our search without waiting for the active repo list.
            search(repoList);

            // But also request it, so that it's cached for later.
            activeRepos.get().catch(_e => { /* ignore it */ });
        }
    }

    public render(): JSX.Element | null {
        if (this.state.loading) {
            return (
                <div className='searchResults'>
                    <div className='search-results__header'>
                        Working...
                    </div>
                </div>
            );
        }
        if (!this.state.results || this.state.results.length === 0) {
            return (
                <div className='searchResults'>
                    <div className='search-results__header'>
                        No results
                    </div>
                </div>
            );
        }
        let totalMatches = 0;
        let totalResults = 0;
        let totalFiles = 0;
        let totalRepos = 0;
        const seenRepos = new Set<string>();
        for (const result of this.state.results) {
            const parsed = URI.parse(result.resource);
            if (!seenRepos.has(parsed.hostname + parsed.path)) {
                seenRepos.add(parsed.hostname + parsed.path);
                totalRepos += 1;
            }
            totalFiles += 1;
            totalResults += result.lineMatches.length;
        }
        const pluralize = (str: string, n: number) => `${str}${n === 1 ? '' : 's'}`;
        return (
            <div className='search-results'>
                <div className='search-results__header'>
                    <div className='search-results__badge'>{numberWithCommas(totalResults)}</div>
                    <div className='search-results__label'>{pluralize('result', totalResults)} in</div>
                    <div className='search-results__badge'>{numberWithCommas(totalFiles)}</div>
                    <div className='search-results__label'>{pluralize('file', totalFiles)}  in</div>
                    <div className='search-results__badge'>{numberWithCommas(totalRepos)}</div>
                    <div className='search-results__label'>{pluralize('repo', totalRepos)} </div>
                </div>
                {this.state.results.map((result, i) => {
                    const prevTotal = totalMatches;
                    totalMatches += result.lineMatches.length;
                    const parsed = URI.parse(result.resource);
                    const refs = result.lineMatches.map(match => {
                        return {
                            range: {
                                start: {
                                    character: match.offsetAndLengths[0][0],
                                    line: match.lineNumber
                                },
                                end: {
                                    character: match.offsetAndLengths[0][0] + match.offsetAndLengths[0][1],
                                    line: match.lineNumber
                                }
                            },
                            uri: result.resource,
                            repoURI: parsed.hostname + parsed.path
                        };
                    });

                    return <ReferencesGroup hidden={prevTotal > 500} uri={parsed.hostname + parsed.path} path={parsed.fragment} key={i} refs={refs} isLocal={false} />;
                })}
            </div>
        );
    }
}
