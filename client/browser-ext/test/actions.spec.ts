import * as actions from "../app/actions";
import * as types from "../app/constants/ActionTypes";
import {keyFor} from "../app/reducers/helpers";
import {expect} from "chai";
import * as fetchMock from "fetch-mock";
import configureMockStore from "redux-mock-store";
import thunk from "redux-thunk";

const middlewares = [thunk];
const mockStore = configureMockStore(middlewares);

describe("actions", () => {
	afterEach(fetchMock.restore);

	const repo = "github.com/gorilla/mux";
	const rev = "master";
	const path = "path";
	const query = "query";
	const resolvedRev = "resolvedRev";

	it("setAccessToken", () => {
		expect(actions.setAccessToken("token")).to.eql({type: types.SET_ACCESS_TOKEN, token: "token"});
	});

	// TODO(john): add proper typings
	function assertAsyncActionsDispatched(action: any, initStore: any, expectedActions: any): any {
		const store = mockStore(initStore);
		return store.dispatch(action).then(() => expect(store.getActions()).to.eql(expectedActions));
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
			fetchMock.mock(annotationsAPI, "GET", {Annotations: []} as any); // TODO(john): why is response type invalid?

            // TODO(john): remove undefined
			return assertAsyncActionsDispatched(actions.getAnnotations(repo, rev, path), {
				resolvedRev: {content: {[keyFor(repo, undefined, undefined, undefined)]: {authRequired: false, cloneInProgress: false, respCode: 200}, [keyFor(repo, rev, undefined, undefined)]: {json: {CommitID: resolvedRev}}}},
				annotations: {content: {}},
			}, [
				{type: types.FETCHED_ANNOTATIONS, repo, relRev: rev, rev: resolvedRev, path, xhrResponse: {body: {Annotations: []}, head: undefined, status: 200}},
			]);
		});

		it("404s", () => {
			fetchMock.mock(annotationsAPI, "GET", 404);

			return assertAsyncActionsDispatched(actions.getAnnotations(repo, rev, path), {
				resolvedRev: {content: {[keyFor(repo, undefined, undefined, undefined)]: {authRequired: false, cloneInProgress: false, respCode: 200}, [keyFor(repo, rev, undefined, undefined)]: {json: {CommitID: resolvedRev}}}},
				annotations: {content: {}},
			}, [
				{type: types.FETCHED_ANNOTATIONS, repo, relRev: rev, rev: resolvedRev, path, xhrResponse: {body: null, head: undefined, status: 404}},
			]);
		});

		it("noops when annotations are cached", () => {
			return assertAsyncActionsDispatched(actions.getAnnotations(repo, rev, path), {
				resolvedRev: {content: {[keyFor(repo, undefined, undefined, undefined)]: {authRequired: false, cloneInProgress: false, respCode: 200}, [keyFor(repo, rev, undefined, undefined)]: {json: {CommitID: resolvedRev}}}},
				defs: {content: {[keyFor(repo, resolvedRev, path, query)]: {Annotations: []}}},
			}, []);
		});
	});
});
