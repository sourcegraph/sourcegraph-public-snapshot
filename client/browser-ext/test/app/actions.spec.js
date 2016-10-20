import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";
import fetchMock from "fetch-mock";

import {expect} from "chai";
import * as types from "../../app/constants/ActionTypes";
import * as actions from "../../app/actions";
import {keyFor} from "../../app/reducers/helpers";

const middlewares = [thunk];
const mockStore = configureMockStore(middlewares);

describe("actions", () => {
	afterEach(fetchMock.restore);

	const repo = "github.com/gorilla/mux";
	const rev = "master";
	const head = "head";
	const base = "base";
	const path = "path";
	const defPath = "defPath";
	const query = "query";

	const resolvedRev = "resolvedRev";

	function errorResponse(status, url) {
		return {response: {status, url}};
	}

	it("setAccessToken", () => {
		expect(actions.setAccessToken("token")).to.eql({type: types.SET_ACCESS_TOKEN, token: "token"});
	});

	function assertAsyncActionsDispatched(action, initStore, expectedActions) {
		let store = mockStore(initStore);
		return store.dispatch(action)
			.then(() => expect(store.getActions()).to.eql(expectedActions));
	}

	describe("ensureRepoExists", () => {
		const repoCreateAPI = `https://sourcegraph.com/.api/repos?AcceptAlreadyExists=true`;

		it("200s", () => {
			fetchMock.mock(repoCreateAPI, "POST", 200);
			return assertAsyncActionsDispatched(actions.ensureRepoExists(repo), {}, []);
		});
	});

	describe("getAnnotations", () => {
		const annotationsAPI = `https://sourcegraph.com/.api/repos/${repo}@${resolvedRev}/-/tree/${path}?ContentsAsString=false&NoSrclibAnns=true`;

		it("200s", () => {
			fetchMock.mock(annotationsAPI, "GET", {Annotations: []});

			return assertAsyncActionsDispatched(actions.getAnnotations(repo, rev, path), {
				resolvedRev: {content: {[keyFor(repo, rev)]: {CommitID: resolvedRev}}},
				annotations: {content: {}},
			}, [
				{type: types.FETCHED_ANNOTATIONS, repo, rev: resolvedRev, path, json: {Annotations: []}},
			]);
		});

		it("404s", () => {
			fetchMock.mock(annotationsAPI, "GET", 404);

			return assertAsyncActionsDispatched(actions.getAnnotations(repo, rev, path), {
				resolvedRev: {content: {[keyFor(repo, rev)]: {CommitID: resolvedRev}}},
				annotations: {content: {}},
			}, [
				{type: types.FETCHED_ANNOTATIONS, repo, rev: resolvedRev, path, err: errorResponse(404, annotationsAPI)},
			]);
		});

		it("noops when annotations are cached", () => {
			return assertAsyncActionsDispatched(actions.getAnnotations(repo, rev, path), {
				resolvedRev: {content: {[keyFor(repo, rev)]: {CommitID: resolvedRev}}},
				defs: {content: {[keyFor(repo, resolvedRev, path, query)]: {Annotations: []}}},
			}, []);
		});
	});
});
