import * as immutable from 'immutable';
import * as Rx from 'rxjs';

/**
 * A contextKey is used to retrieve contents that are fetched and cached in the store.
 * It identifies a repository, with an optional revision and filepath.
 * To get the files that were fetched for repo A at rev B, you would use
 * contextKey('A', 'B').
 */
export function contextKey(uri: string, rev?: string, path?: string): string {
    return `${uri}@${rev}/${path}`;
}

export interface State {
    /**
     * files is a map from contextKey to a list of paths, like ['foo', 'foo/bar', 'foo/bar/baz.go'].
     * An example contextKey is {uri: 'github.com/gorilla/mux', rev: 'master'}.
     */
    files: immutable.Map<string, string[]>;
    /**
     * files is a map from contextKey to a list of paths, like ['foo', 'foo/bar', 'foo/bar/baz.go'].
     * An example contextKey is {uri: 'github.com/gorilla/mux', rev: 'master', path: 'mux.go'}.
     */
    highlightedContents: immutable.Map<string, string>;
}

const initMap = immutable.Map<any, any>({});

const initState: State = { files: initMap, highlightedContents: initMap };
const actionSubject = new Rx.Subject<State>();

const reducer = (state, action) => { // TODO(john): use immutable data structure
    switch (action.type) {
        case 'SET':
            return action.payload;
        default:
            return state;
    }
};

export const store = new Rx.BehaviorSubject<State>(initState);
actionSubject.startWith(initState).scan(reducer).subscribe(store);

const actionDispatcher = func => (...args) => actionSubject.next(func(...args));

export const setState: (t: State) => void = actionDispatcher(payload => ({
    type: 'SET',
    payload
}));

