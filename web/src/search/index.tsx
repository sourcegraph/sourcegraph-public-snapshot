import * as querystring from 'querystring';
import { ActiveRepoResults } from 'sourcegraph/util/types';
import { openFromJS } from 'sourcegraph/util/url';
import * as URI from 'urijs';

export interface SearchParams {
    q: string;
    repos: string;
    files: string;
    matchCase: boolean;
    matchWord: boolean;
    matchRegex: boolean;
}

export function handleSearchInput(e: any, params: SearchParams): void {
    const query = e.target.value;
    if ((e.key !== 'Enter' && e.keyCode !== 13) || !query) {
        return;
    }

    params.q = query;
    openFromJS(getSearchPath(params), e);
}

export function getSearchPath(params: SearchParams): string {
    // Build query string of the string representation of all truthy values
    const query = new URLSearchParams(Object.entries(params).filter(([, value]) => value) as any); // https://github.com/Microsoft/TypeScript/issues/15338
    return '/search?' + query.toString();
}

export function getSearchParamsFromURL(url: string): SearchParams {
    const query: { [key: string]: string } = querystring.parse(URI.parse(url).query);
    return {
        q: query.q || '',
        repos: query.repos || 'active',
        files: query.files || '',
        matchCase: query.matchCase === 'true',
        matchWord: query.matchWord === 'true',
        matchRegex: query.matchRegex === 'true'
    };
}

export function getSearchParamsFromLocalStorage(): SearchParams {
    return {
        q: window.localStorage.getItem('searchQuery') || '',
        repos: window.localStorage.getItem('searchRepoScope') || 'active',
        files: window.localStorage.getItem('searchFileScope') || '',
        matchCase: window.localStorage.getItem('searchMatchCase') === 'true',
        matchWord: window.localStorage.getItem('searchMatchWord') === 'true',
        matchRegex: window.localStorage.getItem('searchMatchRegex') === 'true'
    };
}

export function parseRepoList(repos: string): string[] {
    return repos.split(/\s*,\s*/).map(repo => repo.trim()).filter(repo => repo !== '');
}

export function expandActiveInactive(repos: string[], groups: ActiveRepoResults): string[] {
    const res: string[] = [];
    for (const repo of repos) {
        if (repo === 'active') {
            for (const r of groups.active) {
                res.push(r);
            }
        } else if (repo === 'inactive') {
            for (const r of groups.inactive) {
                res.push(r);
            }
        } else {
            res.push(repo);
        }
    }
    return res;
}
