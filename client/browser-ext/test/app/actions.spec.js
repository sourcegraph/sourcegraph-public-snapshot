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
	const dataVer = "dataVer";
	const srclibDataVersionAPI = `https://sourcegraph.com/.api/repos/${repo}@${resolvedRev}/-/srclib-data-version?Path=${path}`;
	const srclibDataVersion = {CommitID: dataVer};

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
		const repoCreateAPI = `https://sourcegraph.com/.api/repos`;

		it("200s", () => {
			fetchMock.mock(repoCreateAPI, "POST", 200);
		    return assertAsyncActionsDispatched(actions.ensureRepoExists(repo), {}, []);
		});

		it("409s", () => {
			fetchMock.mock(repoCreateAPI, "POST", 409);
		    return assertAsyncActionsDispatched(actions.ensureRepoExists(repo), {}, []);
		});
	});

	describe("getAnnotations", () => {
		const annotationsAPI = `https://sourcegraph.com/.api/annotations?Entry.RepoRev.Repo=${repo}&Entry.RepoRev.CommitID=${dataVer}&Entry.Path=${path}&Range.StartByte=0&Range.EndByte=0`;

		it("200s", () => {
			fetchMock.mock(srclibDataVersionAPI, "GET", srclibDataVersion).mock(annotationsAPI, "GET", {Annotations: []});

		    return assertAsyncActionsDispatched(actions.getAnnotations(repo, rev, path), {
		    	resolvedRev: {content: {[keyFor(repo, rev)]: {CommitID: resolvedRev}}},
		    	annotations: {content: {}},
		    	srclibDataVersion: {content: {}},
		    }, [
		    	{type: types.FETCHED_SRCLIB_DATA_VERSION, repo, rev: resolvedRev, path, json: srclibDataVersion},
		    	{type: types.FETCHED_ANNOTATIONS, repo, rev: dataVer, path, json: {Annotations: []}},
		    ]);
		});

		it("404s", () => {
			fetchMock.mock(srclibDataVersionAPI, "GET", srclibDataVersion).mock(annotationsAPI, "GET", 404);

		    return assertAsyncActionsDispatched(actions.getAnnotations(repo, rev, path), {
		    	resolvedRev: {content: {[keyFor(repo, rev)]: {CommitID: resolvedRev}}},
		    	annotations: {content: {}},
		    	srclibDataVersion: {content: {}},
		    }, [
		    	{type: types.FETCHED_SRCLIB_DATA_VERSION, repo, rev: resolvedRev, path, json: srclibDataVersion},
		    	{type: types.FETCHED_ANNOTATIONS, repo, rev: dataVer, path, err: errorResponse(404, annotationsAPI)},
		    ]);
		});

		it("noops when annotations are cached", () => {
			fetchMock.mock(srclibDataVersionAPI, "GET", srclibDataVersion);

			return assertAsyncActionsDispatched(actions.getAnnotations(repo, rev, path), {
		    	resolvedRev: {content: {[keyFor(repo, rev)]: {CommitID: resolvedRev}}},
				defs: {content: {[keyFor(repo, dataVer, path, query)]: {Annotations: []}}},
		    	srclibDataVersion: {content: {}},
			}, [
		    	{type: types.FETCHED_SRCLIB_DATA_VERSION, repo, rev: resolvedRev, path, json: srclibDataVersion},
			]);
		});
	});
});
