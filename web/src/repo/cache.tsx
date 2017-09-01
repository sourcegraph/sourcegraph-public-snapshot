import * as immutable from 'immutable';
import * as Rx from 'rxjs';

/**
 * RepoContext identifies a repository resource (similar to a URI), such as
 *   - the repository itself: `{repoPath: 'github.com/gorilla/mux'}`
 *   - the repository at a particular revision: `{repoPath: 'github.com/gorilla/mux', rev: 'branch'}`
 *   - a file in a repository at an immutable revision: `{repoPath: 'github.com/gorilla/mux', commitID: '<40-char-SHA>', filePath: 'mux.go'}`
 *
 * NEEDS DISCUSSION: Should we use the same resource URI scheme as we do for the native app?
 */
export interface RepoContext {
    repoPath: string;
    rev?: string;
    commitID?: string;
    filePath?: string;
    line?: number;
}

/**
 * A contextKey is used to retrieve contents that are fetched and cached in the store.
 * It identifies a repository, with an optional revision, filepath, and file location.
 * To get the files that were fetched for repo 'github.com/gorilla/mux' at rev 'foo', you would use
 * `contextKey({repoPath: 'github.com/gorilla/mux', rev: 'B'})`.
 */
export function contextKey(ctx: RepoContext): string {
    return `${ctx.repoPath}@${ctx.rev}_${ctx.commitID}/${ctx.filePath}#L$${ctx.line}`;
}

/**
 * This factory takes a fetching function (e.g. "get the files for this repo using graphql")
 * and returns a new function that is guaranteed not to fetch twice. While a fetch is in flight,
 * a Promise of the in flight value is returned. When the first flight resolves, the value is
 * cached and all subsequent fetches for the same resource will resolve immediately with the cached
 * value. Resources are keyed by contextKey().
 */
export function singleflightCachedFetch<C extends RepoContext, T>(fetch: (ctx: C) => Promise<T>, pick: (ctx: C) => C): (ctx: C, force?: boolean) => Promise<T> {
    const cache = new Map<string, Promise<T>>();
    return (ctx: C, force?: boolean) => {
        const key = contextKey(pick(ctx));
        const hit = cache.get(key);
        if (!force && hit) {
            return hit;
        }
        const p = fetch(ctx);
        cache.set(key, p);
        return p;
    };
}

export interface CacheState {
    /**
     * files is a map from contextKey to a list of paths, like ['foo', 'foo/bar', 'foo/bar/baz.go'].
     * An example contextKey is {uri: 'github.com/gorilla/mux', rev: 'master'}.
     */
    files: immutable.Map<string, string[]>;
    /**
     * highlightedContents is a map from contextKey to highlightedContents (an HTML string).
     * An example contextKey is {uri: 'github.com/gorilla/mux', rev: 'master', path: 'mux.go'}.
     */
    highlightedContents: immutable.Map<string, string>;
    /**
     * resolvedRevs is a map from contextKey to resolved revision (a 40 character commit SHA).
     */
    resolvedRevs: immutable.Map<string, string>;
}

const initMap = immutable.Map<any, any>({});

const initState: CacheState = { files: initMap, highlightedContents: initMap, resolvedRevs: initMap };
const actionSubject = new Rx.Subject<CacheState>();

const reducer = (state, action) => { // TODO(john): use immutable data structure
    switch (action.type) {
        case 'SET':
            return action.payload;
        default:
            return state;
    }
};

export const repoCache = new Rx.BehaviorSubject<CacheState>(initState);
actionSubject.startWith(initState).scan(reducer).subscribe(repoCache);

const actionDispatcher = func => (...args) => actionSubject.next(func(...args));

export const setRepoCache: (t: CacheState) => void = actionDispatcher(payload => ({
    type: 'SET',
    payload
}));
