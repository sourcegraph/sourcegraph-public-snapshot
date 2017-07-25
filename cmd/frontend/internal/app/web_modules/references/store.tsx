import * as Rx from "@sourcegraph/rxjs";
import * as types from "app/util/types";
import * as immutable from "immutable";

// reference implementation: http://rudiyardley.com/redux-single-line-of-code-rxjs/

export interface Location {
	uri: string;
	rev: string;
	path: string;
	line: number;
	char: number;
}

export interface ReferencesContext {
	loc: Location;
	word?: string;
}

export interface ReferencesState {
	context?: ReferencesContext;
	refsByLoc: immutable.Map<string, types.Reference[]>;
}

const initMap = immutable.Map<any, any>({});

const initState: ReferencesState = { refsByLoc: initMap };
const actionSubject = new Rx.Subject<ReferencesState>();

const reducer = (state, action) => { // TODO(john): use immutable data structure
	switch (action.type) {
		case "SET_REFERENCES":
			return action.payload;
		default:
			return state;
	}
};

export const store = new Rx.BehaviorSubject<ReferencesState>(initState);
actionSubject.startWith(initState).scan(reducer).subscribe(store);

const actionDispatcher = (func) => (...args) => actionSubject.next(func(...args));

export const setReferences: (t: ReferencesState) => void = actionDispatcher((payload) => ({
	type: "SET_REFERENCES",
	payload,
}));

export function locKey(loc: Location): string {
	return `${loc.uri}@${loc.rev}/${loc.path}#${loc.line}:${loc.char}`;
}

export function addReferences(loc: Location, refs: types.Reference[]): void {
	const next = { ...store.getValue() };
	next.refsByLoc = next.refsByLoc.update(locKey(loc), (_refs) => (_refs || []).concat(refs));
	setReferences(next);
}
