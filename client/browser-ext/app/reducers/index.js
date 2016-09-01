import {combineReducers} from "redux";
import {keyFor} from "./helpers";
import * as ActionTypes from "../constants/ActionTypes";

const authInfo = function(state = null, action) {
  	switch (action.type) {
  	case ActionTypes.FETCHED_AUTH_INFO:
  		return action.json ? action.json : state;
  	default:
  		return state;
 	}
 }

const accessToken = function(state = null, action) {
	switch (action.type) {
	case ActionTypes.SET_ACCESS_TOKEN:
		return action.token;
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

const build = function(state = {content: {}}, action) {
	switch (action.type) {
	case ActionTypes.FETCHED_BUILD:
	case ActionTypes.CREATED_BUILD:
		return {
			...state,
			content: {
				...state.content,
				[keyFor(action.repo, action.commitID)]: action.json ? action.json : null,
			}
		};
	default:
		return state;
	}
}

const srclibDataVersion = function(state = {content: {}, fetches: {}}, action) {
	switch (action.type) {
	case ActionTypes.FETCHED_SRCLIB_DATA_VERSION:
		return {
			...state,
			fetches: {
				...state.fetches,
				[keyFor(action.repo, action.rev, action.path)]: action.err ? action.err : false,
			},
			content: {
				...state.content,
				[keyFor(action.repo, action.rev, action.path)]: action.json ? action.json : null,
			}
		};
	default:
		return state;
	}
}

const def = function(state = {content: {}}, action) {
	switch (action.type) {
	case ActionTypes.FETCHED_DEF:
		if (!action.json && !state.content[keyFor(action.repo, action.rev, action.defPath)]) return state; // no update needed; avoid re-rending components

		return {
			...state,
			content: {
				...state.content,
				[keyFor(action.repo, action.rev, action.defPath)]: action.json ? action.json : null,
			}
		};
	default:
		return state;
	}
}


const defs = function(state = {content: {}, fetches: {}}, action) {
	switch (action.type) {
	case ActionTypes.WANT_DEFS:
		return {
			...state,
			fetches: {
				...state.fetches,
				[keyFor(action.repo, action.rev, action.path, action.query)]: true,
			}
		};
	case ActionTypes.FETCHED_DEFS:
		return {
			...state,
			fetches: {
				...state.fetches,
				[keyFor(action.repo, action.rev, action.path, action.query)]: action.err ? action.err : false,
			},
			content: {
				...state.content,
				[keyFor(action.repo, action.rev, action.path, action.query)]: action.json ? action.json : null,
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

export default combineReducers({authInfo, accessToken, resolvedRev, srclibDataVersion, build, def, defs, annotations});
