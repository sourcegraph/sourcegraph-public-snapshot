import * as immutable from 'immutable';
import * as Rx from 'rxjs';

export interface State {
    files: immutable.Map<string, GQL.IFile[]>;
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

export function contextKey(uri: string, rev?: string, path?: string): string {
    return `${uri}@${rev}/${path}`;
}
