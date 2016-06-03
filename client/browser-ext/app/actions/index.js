import * as types from "../constants/ActionTypes";
import {keyFor} from "../reducers/helpers";
import fetch, {useAccessToken} from "./xhr";
import {defCache} from "../../chrome/extension/annotations";

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

export function setDefPath(defPath) {
	return {type: types.SET_DEF_PATH, defPath};
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

export function getDef(repo, rev, defPath) {
	return function (dispatch, getState) {
		const state = getState();
		if (state.def.content[keyFor(repo, rev, defPath)]) return Promise.resolve(); // nothing to do; already have def

		// HACK: share def data with annotations.js. This violates the redux
		// boundaries but it means that in many cases you can click on a ref
		// and immediately go there instead of going via the repo homepage.
		//
		// NOTE: Need to keep this in sync with the defCache key structure.
		const cacheKey = `https://sourcegraph.com/.api/repos/${repo}/-/def/${defPath}?ComputeLineRange=true&Doc=true`;
		if (defCache[cacheKey]) {
			// Dispatch FETCHED_DEF so it gets added to the normal def.content
			// for next time.
			dispatch({type: types.FETCHED_DEF, repo, rev, defPath, json: defCache[cacheKey]})
			return Promise.resolve();
		}

		dispatch({type: types.WANT_DEF, repo, rev, defPath})
		return fetch(`https://sourcegraph.com/.api/repos/${repo}@${rev}/-/def/${defPath}?ComputeLineRange=true`)
			.then((json) => dispatch({type: types.FETCHED_DEF, repo, rev, defPath, json}))
			.catch((err) => dispatch({type: types.FETCHED_DEF, repo, rev, defPath, err}));
	}
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
			return fetch(`https://sourcegraph.com/.api/annotations?Entry.RepoRev.Repo=${encodeURIComponent(repo)}&Entry.RepoRev.CommitID=${encodeURIComponent(rev)}&Entry.Path=${encodeURIComponent(path)}&Range.StartByte=0&Range.EndByte=0`)
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

export function expireDef(repo, rev, defPath) {
	return {type: types.EXPIRE_DEF, repo, rev, defPath};
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
