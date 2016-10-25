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
		if (!state.content[keyFor(action.repo, action.rev)] || action.xhrResponse.status === 200) {
			return {
				...state,
				content: {
					...state.content,
					[keyFor(action.repo)]: {
						respCode: action.xhrResponse.status,
						authRequired: action.xhrResponse.status === 404,
						cloneInProgress: action.xhrResponse.status === 202,
					},
					[keyFor(action.repo, action.rev)]: action.xhrResponse.status === 200 ? action.xhrResponse.body : null,
				}
			};
		}
		// Fall through
		// As a result, we serve the result of the last
		// successful request (one that returned HTTP 200).
	default:
		return state; // no update needed; avoid re-rending components
	}
}

const annotations = function(state = {content: {}}, action) {
	switch (action.type) {
	case ActionTypes.FETCHED_ANNOTATIONS:
		if (!state.content[keyFor(action.repo, action.rev, action.path)] || action.xhrResponse.status === 200) {
			return {
				...state,
				content: {
					...state.content,
					[keyFor(action.repo, action.rev, action.path)]: action.xhrResponse.status === 200 ? action.xhrResponse.body : null,
				}
			};
		}
		// Fall through
		// As a result, we serve the result of the last
		// successful request (one that returned HTTP 200).
	default:
		return state;
	}
}

export default combineReducers({accessToken, resolvedRev, annotations});
