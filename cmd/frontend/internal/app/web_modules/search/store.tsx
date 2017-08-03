import { getSearchParamsFromURL, SearchParams } from "app/search";
import * as Rx from "rxjs";

export interface State extends SearchParams {
	showAdvancedSearch?: boolean;
	showAutocomplete?: boolean;
}

const initState: State = getSearchParamsFromURL(window.location.href);
const actionSubject = new Rx.Subject<State>();

const reducer = (state, action) => { // TODO(john): use immutable data structure
	switch (action.type) {
		case "SET":
			return action.payload;
		default:
			return state;
	}
};

export const store = new Rx.BehaviorSubject<State>(initState);
actionSubject.startWith(initState).scan(reducer).subscribe(store);

const actionDispatcher = (func) => (...args) => actionSubject.next(func(...args));

export const setState: (t: State) => void = actionDispatcher((payload) => ({
	type: "SET",
	payload,
}));
