import * as Rx from "@sourcegraph/rxjs";
import { getPathExtension } from "app/util";
import * as types from "app/util/types";

// reference implementation: http://rudiyardley.com/redux-single-line-of-code-rxjs/

export interface TooltipContext {
	path: string;
	repoRevSpec: types.RepoRevSpec;
	coords?: {
		line: number;
		char: number;
	};
	selectedText?: string;
}

export interface TooltipState {
	target?: HTMLElement;
	docked?: boolean;
	context?: TooltipContext;
	data?: types.TooltipData;
}

const initState: TooltipState = {};
const actionSubject = new Rx.Subject<TooltipState>();

const reducer = (state, action) => { // TODO(john): use immutable data structure
	switch (action.type) {
		case "SET_TOOLTIP":
			return action.payload;
		default:
			return state;
	}
};

export const store = new Rx.BehaviorSubject<TooltipState>({});
actionSubject.startWith(initState).scan(reducer).subscribe(store);

const actionDispatcher = (func) => (...args) => actionSubject.next(func(...args));

export const setTooltip: (t: TooltipState) => void = actionDispatcher((payload) => ({
	type: "SET_TOOLTIP",
	payload,
}));

export const clearTooltip = (target?: HTMLElement) => setTooltip({ target });

export function getTooltipEventProperties(data: types.TooltipData, context: TooltipContext): any {
	// TODO: bring back isPullRequest, isCommit
	return {
		repo: context.repoRevSpec.repoURI,
		path: context.path,
		language: getPathExtension(context.path),
		isLoading: Boolean(data.loading),
		hasData: Boolean(data.title),
	};
}
