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
	const path = "path";
	const defPath = "defPath";
	const query = "query";

	const dataVer = "dataVer";
	const srclibDataVersionAPI = `https://sourcegraph.com/.api/repos/${repo}@${rev}/-/srclib-data-version?Path=${path}`;
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

	describe("refreshVCS", () => {
		const refreshVCSAPI = `https://sourcegraph.com/.api/repos/${repo}/-/refresh`;

		it("200s", () => {
			fetchMock.mock(refreshVCSAPI, "POST", 200);
		    return assertAsyncActionsDispatched(actions.refreshVCS(repo), {}, [{type: types.REFRESH_VCS}]);
		});

		it("404s", () => {
			fetchMock.mock(refreshVCSAPI, "POST", 404);
		    return assertAsyncActionsDispatched(actions.refreshVCS(repo), {}, [{type: types.REFRESH_VCS}]);
		});
	});

	describe("ensureRepoExists", () => {
		const repoCreateAPI = `https://sourcegraph.com/.api/repos`;

		it("200s", () => {
			fetchMock.mock(repoCreateAPI, "POST", 200);
		    return assertAsyncActionsDispatched(actions.ensureRepoExists(repo), {createdRepos: {}}, [{type: types.CREATED_REPO, repo}]);
		});

		it("409s", () => {
			fetchMock.mock(repoCreateAPI, "POST", 409);
		    return assertAsyncActionsDispatched(actions.ensureRepoExists(repo), {createdRepos: {}}, [{type: types.CREATED_REPO, repo}]);
		});

		it("noops when status is cached", () => {
			return assertAsyncActionsDispatched(actions.ensureRepoExists(repo), {createdRepos: {[repo]: true}}, []);
		});
	});

	describe("getSrclibDataVersion", () => {

		it("200s", () => {
			fetchMock.mock(srclibDataVersionAPI, "GET", srclibDataVersion);

		    return assertAsyncActionsDispatched(actions.getSrclibDataVersion(repo, rev, path), {
		    	srclibDataVersion: {content: {}, fetches: {}, timestamps: {}}
		    }, [
		    	{type: types.WANT_SRCLIB_DATA_VERSION, repo, rev, path},
		    	{type: types.FETCHED_SRCLIB_DATA_VERSION, repo, rev, path, json: srclibDataVersion},
		    ]);
		});

		it("404s", () => {
			fetchMock.mock(srclibDataVersionAPI, "GET", 404);

		    return assertAsyncActionsDispatched(actions.getSrclibDataVersion(repo, rev, path), {
		    	srclibDataVersion: {content: {}, fetches: {}, timestamps: {}}
		    }, [
		    	{type: types.WANT_SRCLIB_DATA_VERSION, repo, rev, path},
		    	{type: types.FETCHED_SRCLIB_DATA_VERSION, repo, rev, path, err: errorResponse(404, srclibDataVersionAPI)},
		    ]);
		});

		it("noops when dataVer is cached", () => {
			return assertAsyncActionsDispatched(actions.getSrclibDataVersion(repo, rev, path), {
				srclibDataVersion: {content: {[keyFor(repo, rev, path)]: srclibDataVersion}, fetches: {}, timestamps: {}}
			}, []);
		});
	});

	describe("getDefs", () => {
		const defsAPI = `https://sourcegraph.com/.api/defs?RepoRevs=${repo}@${dataVer}&Nonlocal=true&Query=${query}&FilePathPrefix=${path}`;

		it("200s", () => {
			fetchMock.mock(srclibDataVersionAPI, "GET", srclibDataVersion).mock(defsAPI, "GET", {Defs: []});

		    return assertAsyncActionsDispatched(actions.getDefs(repo, rev, path, query), {
		    	defs: {content: {}, fetches: {}, timestamps: {}},
		    	srclibDataVersion: {content: {}, fetches: {}, timestamps: {}},
		    }, [
		    	{type: types.WANT_SRCLIB_DATA_VERSION, repo, rev, path},
		    	{type: types.FETCHED_SRCLIB_DATA_VERSION, repo, rev, path, json: srclibDataVersion},
		    	{type: types.WANT_DEFS, repo, rev: dataVer, path, query},
		    	{type: types.FETCHED_DEFS, repo, rev: dataVer, path, query, json: {Defs: []}},
		    ]);
		});

		it("404s", () => {
			fetchMock.mock(srclibDataVersionAPI, "GET", srclibDataVersion).mock(defsAPI, "GET", 404);

		    return assertAsyncActionsDispatched(actions.getDefs(repo, rev, path, query), {
		    	defs: {content: {}, fetches: {}, timestamps: {}},
		    	srclibDataVersion: {content: {}, fetches: {}, timestamps: {}},
		    }, [
		    	{type: types.WANT_SRCLIB_DATA_VERSION, repo, rev, path},
		    	{type: types.FETCHED_SRCLIB_DATA_VERSION, repo, rev, path, json: srclibDataVersion},
		    	{type: types.WANT_DEFS, repo, rev: dataVer, path, query},
		    	{type: types.FETCHED_DEFS, repo, rev: dataVer, path, query, err: errorResponse(404, defsAPI)},
		    ]);
		});

		it("noops when defs are cached", () => {
			fetchMock.mock(srclibDataVersionAPI, "GET", srclibDataVersion);

			return assertAsyncActionsDispatched(actions.getDefs(repo, rev, path, query), {
				defs: {content: {[keyFor(repo, dataVer, path, query)]: {Defs: []}}, fetches: {}, timestamps: {}},
		    	srclibDataVersion: {content: {}, fetches: {}, timestamps: {}},
			}, [
		    	{type: types.WANT_SRCLIB_DATA_VERSION, repo, rev, path},
		    	{type: types.FETCHED_SRCLIB_DATA_VERSION, repo, rev, path, json: srclibDataVersion},
			]);
		});
	});

	describe("getDef", () => {
		const defAPI = `https://sourcegraph.com/.api/repos/${repo}@${rev}/-/def/${defPath}?ComputeLineRange=true`;

		it("200s", () => {
			fetchMock.mock(defAPI, "GET", {});

		    return assertAsyncActionsDispatched(actions.getDef(repo, rev, defPath), {
		    	def: {content: {}, fetches: {}, timestamps: {}},
		    }, [
		    	{type: types.WANT_DEF, repo, rev, defPath},
		    	{type: types.FETCHED_DEF, repo, rev, defPath, json: {}},
		    ]);
		});

		it("404s", () => {
			fetchMock.mock(defAPI, "GET", 404);

		    return assertAsyncActionsDispatched(actions.getDef(repo, rev, defPath), {
		    	def: {content: {}, fetches: {}, timestamps: {}},
		    }, [
		    	{type: types.WANT_DEF, repo, rev, defPath},
		    	{type: types.FETCHED_DEF, repo, rev, defPath, err: errorResponse(404, defAPI)},
		    ]);
		});

		it("noops when def is cached", () => {
			return assertAsyncActionsDispatched(actions.getDef(repo, rev, defPath), {
				def: {content: {[keyFor(repo, rev, defPath)]: {}}, fetches: {}, timestamps: {}},
			}, []);
		});
	});

	describe("getAnnotations", () => {
		const annotationsAPI = `https://sourcegraph.com/.api/annotations?Entry.RepoRev.Repo=${repo}&Entry.RepoRev.CommitID=${dataVer}&Entry.Path=${path}&Range.StartByte=0&Range.EndByte=0`;

		it("200s", () => {
			fetchMock.mock(srclibDataVersionAPI, "GET", srclibDataVersion).mock(annotationsAPI, "GET", {Annotations: []});

		    return assertAsyncActionsDispatched(actions.getAnnotations(repo, rev, path), {
		    	annotations: {content: {}, fetches: {}, timestamps: {}},
		    	srclibDataVersion: {content: {}, fetches: {}, timestamps: {}},
		    }, [
		    	{type: types.WANT_SRCLIB_DATA_VERSION, repo, rev, path},
		    	{type: types.FETCHED_SRCLIB_DATA_VERSION, repo, rev, path, json: srclibDataVersion},
		    	{type: types.WANT_ANNOTATIONS, repo, rev: dataVer, path},
		    	{type: types.FETCHED_ANNOTATIONS, repo, rev: dataVer, path, json: {Annotations: []}},
		    ]);
		});

		it("404s", () => {
			fetchMock.mock(srclibDataVersionAPI, "GET", srclibDataVersion).mock(annotationsAPI, "GET", 404);

		    return assertAsyncActionsDispatched(actions.getAnnotations(repo, rev, path), {
		    	annotations: {content: {}, fetches: {}, timestamps: {}},
		    	srclibDataVersion: {content: {}, fetches: {}, timestamps: {}},
		    }, [
		    	{type: types.WANT_SRCLIB_DATA_VERSION, repo, rev, path},
		    	{type: types.FETCHED_SRCLIB_DATA_VERSION, repo, rev, path, json: srclibDataVersion},
		    	{type: types.WANT_ANNOTATIONS, repo, rev: dataVer, path},
		    	{type: types.FETCHED_ANNOTATIONS, repo, rev: dataVer, path, err: errorResponse(404, annotationsAPI)},
		    ]);
		});

		it("noops when annotations are cached", () => {
			fetchMock.mock(srclibDataVersionAPI, "GET", srclibDataVersion);

			return assertAsyncActionsDispatched(actions.getAnnotations(repo, rev, path), {
				defs: {content: {[keyFor(repo, dataVer, path, query)]: {Annotations: []}}, fetches: {}, timestamps: {}},
		    	srclibDataVersion: {content: {}, fetches: {}, timestamps: {}},
			}, [
		    	{type: types.WANT_SRCLIB_DATA_VERSION, repo, rev, path},
		    	{type: types.FETCHED_SRCLIB_DATA_VERSION, repo, rev, path, json: srclibDataVersion},
			]);
		});
	});
});
