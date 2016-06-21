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

const createdRepos = function(state = {}, action) {
	switch (action.type) {
	case ActionTypes.CREATED_REPO:
		return {
			...state,
			[action.repo]: true,
		};
	default:
		return state;
	}
}

const resolvedRev = function(state = {content: {}, fetches: {}, timestamps: {}}, action) {
	switch (action.type) {
	case ActionTypes.WANT_RESOLVE_REV:
		return {
			...state,
			fetches: {
				...state.fetches,
				[keyFor(action.repo, action.rev)]: true,
			}
		};
	case ActionTypes.RESOLVED_REV:
		return {
			...state,
			fetches: {
				...state.fetches,
				[keyFor(action.repo, action.rev)]: action.err ? action.err : false,
			},
			content: {
				...state.content,
				[keyFor(action.repo, action.rev)]: action.json ? action.json : null,
			},
			timestamps: {
				...state.timestamps,
				[keyFor(action.repo, action.rev)]: action.json ? Date.now() : null,
			}
		};
	default:
		return state;
	}
}

const delta = function(state = {content: {}, fetches: {}, timestamps: {}}, action) {
	switch (action.type) {
	case ActionTypes.WANT_DELTA:
		return {
			...state,
			fetches: {
				...state.fetches,
				[keyFor(action.repo, action.base, action.head)]: true,
			}
		};
	case ActionTypes.FETCHED_DELTA:
		return {
			...state,
			fetches: {
				...state.fetches,
				[keyFor(action.repo, action.base, action.head)]: action.err ? action.err : false,
			},
			content: {
				...state.content,
				[keyFor(action.repo, action.base, action.head)]: action.json ? action.json : null,
			},
			timestamps: {
				...state.timestamps,
				[keyFor(action.repo, action.base, action.head)]: action.json ? Date.now() : null,
			}
		};
	default:
		return state;
	}
}

const build = function(state = {content: {}, fetches: {}, timestamps: {}}, action) {
	switch (action.type) {
	case ActionTypes.WANT_BUILD:
		return {
			...state,
			fetches: {
				...state.fetches,
				[keyFor(action.repo, action.base, action.head)]: true,
			}
		};
	case ActionTypes.FETCHED_BUILD:
		return {
			...state,
			fetches: {
				...state.fetches,
				[keyFor(action.repo, action.commitID)]: action.err ? action.err : false,
			},
			content: {
				...state.content,
				[keyFor(action.repo, action.commitID)]: action.json ? action.json : null,
			},
			timestamps: {
				...state.timestamps,
				[keyFor(action.repo, action.commitID)]: action.json ? Date.now() : null,
			}
		};
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
	default:
		return state;
	}
}

export default combineReducers({accessToken, resolvedRev, srclibDataVersion, delta, build, def, defs, annotations, createdRepos});
