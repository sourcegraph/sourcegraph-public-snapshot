import * as immutable from "immutable";
import * as Rx from "rxjs";
import { Reference, ResolvedRepoRevSpec } from "sourcegraph/util/types";

// reference implementation: http://rudiyardley.com/redux-single-line-of-code-rxjs/

export interface Location extends ResolvedRepoRevSpec {
	path: string;
	line: number;
	char: number;
}

export interface ReferencesContext {
	loc: Location;
	word?: string;
}

type FetchStatus = "pending" | "completed";

export interface ReferencesState {
	context?: ReferencesContext;
	refsByLoc: immutable.Map<string, Reference[]>;
	fetches: immutable.Map<string, FetchStatus>;
}

const initMap = immutable.Map<any, any>({});

const initState: ReferencesState = { refsByLoc: initMap, fetches: initMap };
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
	return `${loc.repoURI}@${loc.commitID}/${loc.path}#${loc.line}:${loc.char}`;
}

export function addReferences(loc: Location, refs: Reference[]): void {
	const next = { ...store.getValue() };
	next.refsByLoc = next.refsByLoc.update(locKey(loc), (_refs) => (_refs || []).concat(refs));
	setReferences(next);
}

export function refsFetchKey(loc: Location, local: boolean): string {
	return locKey(loc) + "_" + local;
}

function setRefsHelper(key: string, status: FetchStatus): void {
	const next = { ...store.getValue() };
	next.fetches = next.fetches.set(key, status);
	setReferences(next);
}

export function setReferencesLoad(loc: Location, status: FetchStatus): void {
	setRefsHelper(refsFetchKey(loc, true), status);
}

export function setXReferencesLoad(loc: Location, status: FetchStatus): void {
	setRefsHelper(refsFetchKey(loc, false), status);
}
