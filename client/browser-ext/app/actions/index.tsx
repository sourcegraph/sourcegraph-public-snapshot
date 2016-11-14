import * as types from "../constants/ActionTypes";
import {ReducerState} from "../reducers";
import {keyFor} from "../reducers/helpers";
import {doFetch as fetch, useAccessToken} from "./xhr";
import {Dispatch, ThunkAction} from "redux";

export const allActions = {setAccessToken, getAnnotations, ensureRepoExists};

export function setAccessToken(token: string): types.SetAccessTokenAction {
	useAccessToken(token); // for future fetches
	return {type: types.SET_ACCESS_TOKEN, token};
}

// Utility method to fetch the absolute commit id for a branch
const resolveRevOnce = new Map<string, Promise<any>>();
function _resolveRev(dispatch: Dispatch<ReducerState>, state: ReducerState, repo: string, rev: string): Promise<any> {
	const resolvedRep = state.resolvedRev.content[keyFor(repo)];
	const resolvedRev = state.resolvedRev.content[keyFor(repo, rev)];

	// If we have a successful fetch, return that data since rev
	// won't change in a matter of a few seconds between the fetches.
	if (resolvedRev && resolvedRep && resolvedRep.respCode === 200) {
		return Promise.resolve(resolvedRev);
	}

	// If a fetch is in flight, return that same promise
	if (resolveRevOnce.has(keyFor(repo, rev))) {
		return resolveRevOnce.get(keyFor(repo, rev)) as Promise<any>;
	}

	// Createa a new fetch request
	const revPromise = fetch(`https://sourcegraph.com/.api/repos/${repo}${rev ? `@${rev}` : ""}/-/rev`)
		.then((resp) => {
			resolveRevOnce.delete(keyFor(repo, rev));

			return resp.json()
				.then((json) => {
					const xhrResponse = {
						status: resp.status,
						head: (resp.headers as any).map, // TODO(john): figure out if we need this
						json: json,
					};
					dispatch({type: types.RESOLVED_REV, repo, rev, xhrResponse});
					return xhrResponse;
				})
				.catch((err) => {
					const xhrResponse = {
						status: resp.status,
						head: (resp.headers as any).map, // TODO(john): figure out if we need this
						json: null,
					};
					dispatch({type: types.RESOLVED_REV, repo, rev, xhrResponse});
					return xhrResponse;
				});
		})
		.catch(() => ({}));

	resolveRevOnce.set(keyFor(repo, rev), revPromise);
	return revPromise;
}

// This is used to fetch the styling info, which we use to tokenize nodes in DOM
export function getAnnotations(repo: string, rev: string, path: string): ThunkAction<Promise<any>, ReducerState, {}> {
	return function(dispatch: Dispatch<ReducerState>, getState: () => ReducerState): Promise<any> {
		const state = getState();
		return _resolveRev(dispatch, state, repo, rev).then((resolvedRevResp) => {
			const resolvedRepoRev = resolvedRevResp.json.CommitID;

			if (state.annotations.content[keyFor(repo, resolvedRepoRev, path)]) {
				return Promise.resolve(); // nothing to do; already have annotations
			}

			// TODO: Remove NoSrclibAnns when srclib has been purged
			return fetch(`https://sourcegraph.com/.api/repos/${repo}@${resolvedRepoRev}/-/tree/${path}?ContentsAsString=false&NoSrclibAnns=true`)
				.then((resp) => {
					return resp.json()
						.then((json) => {
							const xhrResponse = {
								status: resp.status,
								head: (resp.headers as any).map,
								body: json,
							};
							dispatch({type: types.FETCHED_ANNOTATIONS, repo, relRev: rev, rev: resolvedRepoRev, path, xhrResponse});
							return xhrResponse;
						})
						.catch((err) => {
							const xhrResponse = {
								status: resp.status,
								head: (resp.headers as any).map,
								body: null,
							};
							dispatch({type: types.FETCHED_ANNOTATIONS, repo, relRev: rev, rev: resolvedRepoRev, path, xhrResponse});
							return xhrResponse;
						});
				})
				.catch((err) => ({}));
		})
		.catch((err) => ({})); // no error handling
	};
}

// TODO(john): this doesn't need to be wrapped in Redux...
const createdRepoOnce = new Map<string, Promise<null>>();
export function ensureRepoExists(repo: string): ThunkAction<Promise<null>, ReducerState, any> {
	return function(): Promise<null> { // no dispatch or state necessary...
		if (createdRepoOnce.has(repo)) {
			return Promise.resolve(null);
		}

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
			.then((json) => null)
			.catch((err) => null);
		createdRepoOnce.set(repo, p);
		return p;
	};
}
