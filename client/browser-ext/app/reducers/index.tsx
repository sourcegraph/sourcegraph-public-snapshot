import * as Actions from "../constants/types";
import {combineReducers} from "redux";

export function keyFor(repo: string, rev?: string, path?: string, query?: string): string {
	return `${repo || null}@${rev || null}@${path || null}@${query || null}`;
}

export type AccessTokenState = string | null;
const accessToken = function(state: AccessTokenState = null, action: Actions.SetAccessTokenAction): AccessTokenState {
	switch (action.type) {
	case Actions.SET_ACCESS_TOKEN:
		return action.token;

	default:
		return state;
	}
};

export type ResolvedRevState = {content: {[key: string]: any}}; // TODO(john): use proper type
const resolvedRev = function(state: ResolvedRevState = {content: {}}, action: Actions.ResolvedRevAction): ResolvedRevState {
	switch (action.type) {
	case Actions.RESOLVED_REV:
		if (!state.content[keyFor(action.repo, action.rev)] || action.xhrResponse.status === 200) {
			return Object.assign({}, state, {
				content: Object.assign({}, state.content, {
					[keyFor(action.repo)]: {
						respCode: action.xhrResponse.status,
						authRequired: action.xhrResponse.status === 404,
						cloneInProgress: action.xhrResponse.status === 202,
					},
					[keyFor(action.repo, action.rev)]: {json: action.xhrResponse.status === 200 ? action.xhrResponse.json : null},
				}),
			});
		}
		// Fall through
		// As a result, we serve the result of the last
		// successful request (one that returned HTTP 200).

	default:
		return state; // no update needed; avoid re-rending components
	}
};

export interface ReducerState {
	accessToken: AccessTokenState;
	resolvedRev: ResolvedRevState;
}
export const rootReducer = combineReducers<ReducerState>({accessToken, resolvedRev});
