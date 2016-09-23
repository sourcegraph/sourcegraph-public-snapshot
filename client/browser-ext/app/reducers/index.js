import {combineReducers} from "redux";
import {keyFor} from "./helpers";
import * as ActionTypes from "../constants/ActionTypes";

const accessToken = function(state = null, action) {
	switch (action.type) {
	case ActionTypes.SET_ACCESS_TOKEN:
		return action.token ? action.token : null;
	default:
		return state;
	}
}

const resolvedRev = function(state = {content: {}}, action) {
	switch (action.type) {
	case ActionTypes.RESOLVED_REV:
		if (!action.json && !state.content[keyFor(action.repo, action.rev)]) return state; // no update needed; avoid re-rending components

		return {
			...state,
			content: {
				...state.content,
				[keyFor(action.repo, action.rev)]: action.json ? action.json : null,
			}
		};
	default:
		return state;
	}
}

const annotations = function(state = {content: {}}, action) {
	switch (action.type) {
	case ActionTypes.FETCHED_ANNOTATIONS:
		if (!action.json && !state.content[keyFor(action.repo, action.rev, action.path)]) return state; // no update needed; avoid re-rending components

		return {
			...state,
			content: {
				...state.content,
				[keyFor(action.repo, action.rev, action.path)]: action.json ? action.json : null,
			}
		};
	default:
		return state;
	}
}

export default combineReducers({accessToken, resolvedRev, annotations});
