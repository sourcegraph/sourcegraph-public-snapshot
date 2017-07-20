import * as Rx from "@sourcegraph/rxjs";
import * as types from "app/util/types";

// reference implementation: http://rudiyardley.com/redux-single-line-of-code-rxjs/

export interface ReferencesContext {
	path: string;
	repoRevSpec: types.RepoRevSpec;
	coords: {
		line: number;
		char: number;
		word: string;
	};
}

export interface ReferencesState {
	docked?: boolean;
	context?: ReferencesContext;
	data?: types.ReferencesData;
}

const initState: ReferencesState = {};
const actionSubject = new Rx.Subject<ReferencesState>();

const reducer = (state, action) => { // TODO(john): use immutable data structure
	switch (action.type) {
		case "SET_REFERENCES":
			return action.payload;
		default:
			return state;
	}
};

export const store = new Rx.BehaviorSubject<ReferencesState>({});
actionSubject.startWith(initState).scan(reducer).subscribe(store);

const actionDispatcher = (func) => (...args) => actionSubject.next(func(...args));

export const setReferences: (t: ReferencesState) => void = actionDispatcher((payload) => ({
	type: "SET_REFERENCES",
	payload,
}));
