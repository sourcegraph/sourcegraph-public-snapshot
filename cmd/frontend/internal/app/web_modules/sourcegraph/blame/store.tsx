import * as immutable from "immutable";
import * as moment from "moment";
import * as Rx from "rxjs";
import * as types from "sourcegraph/util/types";

// reference implementation: http://rudiyardley.com/redux-single-line-of-code-rxjs/

export interface BlameContext {
	time: moment.Moment;
	repoURI: string;
	rev: string;
	path: string;
	line: number;
}

export interface BlameState {
	context?: BlameContext;
	hunksByLoc: immutable.Map<string, types.Hunk[]>;
	displayLoading: boolean;
}

const initMap = immutable.Map<any, any>({});

const initState: BlameState = { hunksByLoc: initMap, displayLoading: false };
const actionSubject = new Rx.Subject<BlameState>();

const reducer = (state, action) => { // TODO(john): use immutable data structure
	switch (action.type) {
		case "SET_BLAME":
			return action.payload;
		default:
			return state;
	}
};

export const store = new Rx.BehaviorSubject<BlameState>(initState);
actionSubject.startWith(initState).scan(reducer).subscribe(store);

const actionDispatcher = (func) => (...args) => actionSubject.next(func(...args));

export const setBlame: (t: BlameState) => void = actionDispatcher((payload) => ({
	type: "SET_BLAME",
	payload,
}));

export function contextKey(ctx: BlameContext): string {
	return `${ctx.repoURI}@${ctx.rev}/${ctx.path}#${ctx.line}`;
}

export function addHunks(ctx: BlameContext, hunks: types.Hunk[]): void {
	const next = { ...store.getValue() };
	next.hunksByLoc = next.hunksByLoc.set(contextKey(ctx), hunks);
	setBlame(next);
}
