import * as types from "../constants/ActionTypes";
import {keyFor} from "../reducers/helpers";
import fetch, {useAccessToken} from "./xhr";
import EventLogger from "../analytics/EventLogger";

export function getAuthentication(state) {
	return function (dispatch) {
		return fetch("https://sourcegraph.com/.api/auth-info")
			.then((json) => dispatch({type: types.FETCHED_AUTH_INFO, json}))
			.then((action) => EventLogger.setUserLogin(action.json.Login))
			.catch((err) => dispatch({type: types.FETCHED_AUTH_INFO, err}));
	}
}

export function setAccessToken(token) {
	useAccessToken(token); // for future fetches
	return {type: types.SET_ACCESS_TOKEN, token};
}

// Utility method to fetch the absolute commit id for a branch
function _resolveRev(dispatch, state, repo, rev) {
	const resolvedRev = state.resolvedRev.content[keyFor(repo, rev)];
	if (resolvedRev) return Promise.resolve(resolvedRev);

	return fetch(`https://sourcegraph.com/.api/repos/${repo}${rev ? `@${rev}` : ""}/-/rev`)
		.then((json) => { dispatch({type: types.RESOLVED_REV, repo, rev, json}); return json; })
		.catch((err) => { dispatch({type: types.RESOLVED_REV, repo, rev, err}); throw err; });
}

// This is used to fetch the styling info, which we use to tokenize nodes in DOM
export function getAnnotations(repo, rev, path) {
	return function (dispatch, getState) {
		const state = getState();
		return _resolveRev(dispatch, state, repo, rev).then((json) => {
			rev = json.CommitID;
			if (state.annotations.content[keyFor(repo, rev, path)]) return Promise.resolve(); // nothing to do; already have annotations

			// TODO: Remove NoSrclibAnns when srclib has been purged
			return fetch(`https://sourcegraph.com/.api/repos/${repo}@${rev}/-/tree/${path}?ContentsAsString=false&NoSrclibAnns=true`)
				.then((json) => dispatch({type: types.FETCHED_ANNOTATIONS, repo, rev, path, json}))
				.catch((err) => dispatch({type: types.FETCHED_ANNOTATIONS, repo, rev, path, err}));
		}).catch((err) => {}); // no error handling
	}
}

export function refreshVCS(repo) {
	return function (dispatch) {
		return fetch(`https://sourcegraph.com/.api/repos/${repo}/-/refresh`, {method:  "POST"})
			.then((json) => {})
			.catch((err) => {});
	}
}

export function ensureRepoExists(repo) {
	return function() {
		const body = {
			Op: {
				New: {
					URI: repo,
					CloneURL: `https://${repo}`,
					DefaultBranch: "master",
					Mirror: true,
				},
			},
		};

		return fetch(`https://sourcegraph.com/.api/repos`, {method: "POST", body: JSON.stringify(body)})
			.then((json) => {})
			.catch((err) => {});
	}
}
