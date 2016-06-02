import {combineReducers} from "redux";
import {keyFor} from "./helpers";
import * as ActionTypes from "../constants/ActionTypes";

const accessToken = function(state = null, action) {
	switch (action.type) {
	case ActionTypes.SET_ACCESS_TOKEN:
		return action.token ? action.token : state;
	default:
		return state;
	}
}

const repo = function(state = "github.com/gorilla/mux", action) {
	switch (action.type) {
	case ActionTypes.SET_REPO_REV:
		return action.repo ? action.repo : state;
	default:
		return state;
	}
}

const rev = function(state = "master", action) {
	switch (action.type) {
	case ActionTypes.SET_REPO_REV:
		return action.rev ? action.rev : state;
	default:
		return state;
	}
}

const path = function(state = null, action) {
	switch (action.type) {
	case ActionTypes.SET_PATH:
		return action.path ? action.path : null;
	default:
		return state;
	}
}

const defPath = function(state = null, action) {
	switch (action.type) {
	case ActionTypes.SET_DEF_PATH:
		return action.defPath ? action.defPath : null;
	default:
		return state;
	}
}

const query = function(state = "", action) {
	switch (action.type) {
	case ActionTypes.SET_QUERY:
		return action.query;
	default:
		return state;
	}
}

const srclibDataVersion = function(state = {content: {}, fetches: {}, timestamps: {}}, action) {
	switch (action.type) {
	case ActionTypes.WANT_SRCLIB_DATA_VERSION:
		return {
			...state,
			fetches: {
				...state.fetches,
				[keyFor(action.repo, action.rev, action.path)]: true,
			}
		};
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
			},
			timestamps: {
				...state.timestamps,
				[keyFor(action.repo, action.rev, action.path)]: action.json ? Date.now() : null,
			}
		};
	case ActionTypes.EXPIRE_SRCLIB_DATA_VERSION:
		return {
			...state,
			content: {
				...state.content,
				[keyFor(action.repo, action.rev, action.path)]: null,
			},
			timestamps: {
				...state.timestamps,
				[keyFor(action.repo, action.rev, action.path)]: null,
			}
		};
	default:
		return state;
	}
}

const def = function(state = {content: {}, fetches: {}, timestamps: {}}, action) {
	switch (action.type) {
	case ActionTypes.WANT_DEF:
		return {
			...state,
			fetches: {
				...state.fetches,
				[keyFor(action.repo, action.rev, action.defPath)]: true,
			}
		};
	case ActionTypes.FETCHED_DEF:
		return {
			...state,
			fetches: {
				...state.fetches,
				[keyFor(action.repo, action.rev, action.defPath)]: action.err ? action.err : false,
			},
			content: {
				...state.content,
				[keyFor(action.repo, action.rev, action.defPath)]: action.json ? action.json : null,
			},
			timestamps: {
				...state.timestamps,
				[keyFor(action.repo, action.rev, action.defPath)]: action.json ? Date.now() : null,
			}
		};
	case ActionTypes.EXPIRE_DEF:
		return {
			...state,
			content: {
				...state.content,
				[keyFor(action.repo, action.rev, action.defPath)]: null,
			},
			timestamps: {
				...state.timestamps,
				[keyFor(action.repo, action.rev, action.defPath)]: null,
			}
		};
	default:
		return state;
	}
}


const defs = function(state = {content: {}, fetches: {}, timestamps: {}}, action) {
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
			},
			timestamps: {
				...state.timestamps,
				[keyFor(action.repo, action.rev, action.path, action.query)]: action.json ? Date.now() : null,
			}
		};
	case ActionTypes.EXPIRE_DEFS:
		return {
			...state,
			content: {
				...state.content,
				[keyFor(action.repo, action.rev, action.path, action.query)]: null,
			},
			timestamps: {
				...state.timestamps,
				[keyFor(action.repo, action.rev, action.path, action.query)]: null,
			}
		};
	default:
		return state;
	}
}

const annotations = function(state = {content: {}, fetches: {}, timestamps: {}}, action) {
	switch (action.type) {
	case ActionTypes.WANT_ANNOTATIONS:
		return {
			...state,
			fetches: {
				...state.fetches,
				[keyFor(action.repo, action.rev, action.path)]: true,
			}
		};
	case ActionTypes.FETCHED_ANNOTATIONS:
		return {
			...state,
			fetches: {
				...state.fetches,
				[keyFor(action.repo, action.rev, action.path)]: action.err ? action.err : false,
			},
			content: {
				...state.content,
				[keyFor(action.repo, action.rev, action.path)]: action.json ? action.json : null,
			},
			timestamps: {
				...state.timestamps,
				[keyFor(action.repo, action.rev, action.path)]: action.json ? Date.now() : null,
			}
		};
	case ActionTypes.EXPIRE_ANNOTATIONS:
		return {
			...state,
			content: {
				...state.content,
				[keyFor(action.repo, action.rev, action.path)]: null,
			},
			timestamps: {
				...state.timestamps,
				[keyFor(action.repo, action.rev, action.path)]: null,
			}
		};
	default:
		return state;
	}
}

export default combineReducers({accessToken, repo, rev, path, defPath, query, srclibDataVersion, def, defs, annotations});
