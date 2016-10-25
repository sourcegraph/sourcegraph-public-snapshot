import * as types from "../constants/ActionTypes";
import {keyFor} from "../reducers/helpers";
import fetch, {useAccessToken} from "./xhr";
import EventLogger from "../analytics/EventLogger";

export function setAccessToken(token) {
	useAccessToken(token); // for future fetches
	return {type: types.SET_ACCESS_TOKEN, token};
}

// Utility method to fetch the absolute commit id for a branch
const resolveRevOnce = new Map();
function _resolveRev(dispatch, state, repo, rev) {
	const resolvedRep = state.resolvedRev.content[keyFor(repo)];
	const resolvedRev = state.resolvedRev.content[keyFor(repo, rev)];

	// If we have a successful fetch, return that data since rev won't change w.r.t time
	if (resolvedRev && resolvedRep && resolvedRep.respCode === 200) return Promise.resolve(resolvedRev);

	// If a fetch is in flight, return that same promise
	if (resolveRevOnce.has(keyFor(repo, rev))) return resolveRevOnce.get(keyFor(repo, rev));

	// Createa a new fetch request
	const revPromise = fetch(`https://sourcegraph.com/.api/repos/${repo}${rev ? `@${rev}` : ""}/-/rev`)
		.then((resp) => {
			resolveRevOnce.delete(keyFor(repo, rev));

			return resp.json()
				.then((json) => {
					const xhrResponse = Object.assign({status: resp.status}, {head: resp.headers.map}, {body: json});
					dispatch({type: types.RESOLVED_REV, repo, rev, xhrResponse});
					return xhrResponse;
				})
				.catch((err) => {
					const xhrResponse = Object.assign({status: resp.status}, {head: resp.headers.map}, {body: null});
					dispatch({type: types.RESOLVED_REV, repo, rev, xhrResponse});
					return xhrResponse;
				});
		})
		.catch(() => {});

	resolveRevOnce.set(keyFor(repo, rev), revPromise);
	return revPromise;
}

// This is used to fetch the styling info, which we use to tokenize nodes in DOM
export function getAnnotations(repo, rev, path) {
	return function (dispatch, getState) {
		const state = getState();
		return _resolveRev(dispatch, state, repo, rev).then((xhrResponse) => {
			const resolvedRepoRev = xhrResponse.body.CommitID;

			if (state.annotations.content[keyFor(repo, resolvedRepoRev, path)]) return Promise.resolve(); // nothing to do; already have annotations

			// TODO: Remove NoSrclibAnns when srclib has been purged
			return fetch(`https://sourcegraph.com/.api/repos/${repo}@${resolvedRepoRev}/-/tree/${path}?ContentsAsString=false&NoSrclibAnns=true`)
				.then((resp) => {
					return resp.json()
						.then((json) => {
							const xhrResponse = Object.assign({status: resp.status}, {head: resp.headers.map}, {body: json});
							dispatch({type: types.FETCHED_ANNOTATIONS, repo, rev: resolvedRepoRev, path, xhrResponse});
							return xhrResponse;
						})
						.catch((err) => {
							const xhrResponse = Object.assign({status: resp.status}, {head: resp.headers.map}, {body: null});
							dispatch({type: types.FETCHED_ANNOTATIONS, repo, rev: resolvedRepoRev, path, xhrResponse});
							return xhrResponse;
						});
				})
				.catch((err) => {});
		})
		.catch((err) => {}); // no error handling
	}
}

const createdRepoOnce = new Map();
export function ensureRepoExists(repo) {
	return function () {
		if (createdRepoOnce.has(repo)) return Promise.resolve(null);

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

		const p = fetch(`https://sourcegraph.com/.api/repos?AcceptAlreadyExists=true`, {method: "POST", body: JSON.stringify(body)})
			.then((json) => {})
			.catch((err) => {});
		createdRepoOnce.set(repo, p);
		return p;
	}
}
