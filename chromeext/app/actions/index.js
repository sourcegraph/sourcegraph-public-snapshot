import * as types from "../constants/ActionTypes";
import {keyFor} from "../reducers/helpers";
import fetch, {useAccessToken} from "./xhr";

export function setAccessToken(token) {
	useAccessToken(token); // for future fetches
	return {type: types.SET_ACCESS_TOKEN, token};
}

export function setRepoRev(repo, rev) {
	return {type: types.SET_REPO_REV, repo, rev};
}

export function setPath(path) {
	return {type: types.SET_PATH, path};
}

export function setQuery(query) {
	return {type: types.SET_QUERY, query};
}

function fetchSrclibDataVersion(dispatch, currJson, repo, rev, path) {
	let promise;
	// TODO: handle inflight / errored fetches of srclibDataVersion.
	if (currJson) {
		if (currJson.CommitID) {
			promise = Promise.resolve(currJson);
		} else {
			promise = Promise.reject(new Error("no srclib data version"));
		}
	} else {
		dispatch({type: types.WANT_SRCLIB_DATA_VERSION, repo, rev, path})
		promise = fetch(`https://sourcegraph.com/.api/repos/${repo}@${rev}/-/srclib-data-version?Path=${path ? encodeURIComponent(path) : ""}`)
			.then((json) => { dispatch({type: types.FETCHED_SRCLIB_DATA_VERSION, repo, rev, path, json}); return json; })
			.catch((err) => { dispatch({type: types.FETCHED_SRCLIB_DATA_VERSION, repo, rev, path, err}); throw err; });
	}
	return promise;
}

export function getDefs(repo, rev, path, query) {
	return function (dispatch, getState) {
		const state = getState();
		// Before fetching defs, get the srclib data version.
		const srclibDataVersion = state.srclibDataVersion.content[keyFor(repo, rev, path)];
		if (srclibDataVersion && srclibDataVersion.CommitID) {
			if (state.defs.content[keyFor(repo, srclibDataVersion.CommitID, path, query)]) return Promise.resolve(); // nothing to do; already have defs
		}

		fetchSrclibDataVersion(dispatch, srclibDataVersion, repo, rev, path).then((json) => {
			rev = json.CommitID;
			if (state.defs.content[keyFor(repo, rev, path, query)]) return Promise.resolve(); // nothing to do; already have defs
			dispatch({type: types.WANT_DEFS, repo, rev, path, query})
			return fetch(`https://sourcegraph.com/.api/defs?RepoRevs=${encodeURIComponent(repo)}@${encodeURIComponent(rev)}&Nonlocal=true&Query=${encodeURIComponent(query)}&FilePathPrefix=${path ? encodeURIComponent(path) : ""}`)
				.then((json) => dispatch({type: types.FETCHED_DEFS, repo, rev, path, query, json}))
				.catch((err) => dispatch({type: types.FETCHED_DEFS, repo, rev, path, query, err}));
		})
	}
}

export function getAnnotations(repo, rev, path) {
	return function (dispatch, getState) {
		const state = getState();
		const srclibDataVersion = state.srclibDataVersion.content[keyFor(repo, rev, path)];
		if (srclibDataVersion && srclibDataVersion.CommitID) {
			if (state.annotations.content[keyFor(repo, srclibDataVersion.CommitID, path)]) return Promise.resolve(); // nothing to do; already have annotations
		}

		fetchSrclibDataVersion(dispatch, srclibDataVersion, repo, rev, path).then((json) => {
			rev = json.CommitID;
			if (state.annotations.content[keyFor(repo, rev, path)]) return Promise.resolve(); // nothing to do; already have annotations
			dispatch({type: types.WANT_ANNOTATIONS, repo, rev, path});
			return fetch(`https://sourcegraph.com/.api/annotations?Entry.RepoRev.URI=${encodeURIComponent(repo)}&Entry.RepoRev.Rev=${encodeURIComponent(rev)}&Entry.RepoRev.CommitID=&Entry.Path=${encodeURIComponent(path)}&Range.StartByte=0&Range.EndByte=0`)
				.then((json) => dispatch({type: types.FETCHED_ANNOTATIONS, repo, rev, path, json}))
				.catch((err) => dispatch({type: types.FETCHED_ANNOTATIONS, repo, rev, path, err}));
		});
	}
}

export function expireAnnotations(repo, rev, path) {
	return {type: types.EXPIRE_ANNOTATIONS, repo, rev, path};
}

export function expireSrclibDataVersion(repo, rev, path) {
	return {type: types.EXPIRE_SRCLIB_DATA_VERSION, repo, rev, path};
}

export function expireDefs(repo, rev, path, query) {
	return {type: types.EXPIRE_DEFS, repo, rev, path, query};
}

// refreshVCS has no UI side effects
export function refreshVCS(repo) {
	return function (dispatch) {
		return fetch(`https://sourcegraph.com/.api/repos/${repo}/-/refresh`, {method: "POST"});
	}
}
