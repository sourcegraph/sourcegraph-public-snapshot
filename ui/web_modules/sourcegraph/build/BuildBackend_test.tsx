// tslint:disable: typedef ordered-imports curly

import expect from "expect.js";

import * as Dispatcher from "sourcegraph/Dispatcher";
import {BuildBackend} from "sourcegraph/build/BuildBackend";
import {BuildStore} from "sourcegraph/build/BuildStore";
import * as BuildActions from "sourcegraph/build/BuildActions";
import {immediateSyncPromise} from "sourcegraph/util/immediateSyncPromise";

describe("BuildBackend", () => {
	it("should handle WantBuild", () => {
		let action = {
			repo: "aRepo",
			buildID: 123,
		};
		let expectedURI = `/.api/repos/${action.repo}/-/builds/${action.buildID}`;

		BuildBackend.fetch = function(url, options) {
			expect(url).to.be(expectedURI);
			return immediateSyncPromise({status: 200, json: () => ({ID: 123})});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			BuildBackend.__onDispatch(new BuildActions.WantBuild(action.repo, action.buildID));
		})).to.eql([new BuildActions.BuildFetched(action.repo, action.buildID, {ID: 123})]);
	});

	it("should handle WantLog", () => {
		let action = {
			repo: "aRepo",
			buildID: 123,
			taskID: 456,
		};
		let expectedURI = `/.api/repos/${action.repo}/-/builds/${action.buildID}/tasks/${action.taskID}/log`;

		BuildBackend.fetch = function(url, options) {
			expect(url).to.be(expectedURI);
			return immediateSyncPromise({status: 200, text: () => immediateSyncPromise("a"), headers: {get() { return "789"; }}});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			BuildBackend.__onDispatch(new BuildActions.WantLog(action.repo, action.buildID, action.taskID));
		})).to.eql([new BuildActions.LogFetched(action.repo, action.buildID, action.taskID, 0, 789, "a")]);
	});

	it("should reuse the MaxID to allow efficient tailing", () => {
		let action = {
			repo: "aRepo",
			buildID: 123,
			taskID: 456,
		};

		// Mock the previous fetch as having returned a maxID of 12.
		BuildStore.logs = {get() { return {maxID: 12, log: "a\n"}; }};

		// Trigger "second" fetch, which should reuse MaxID from
		// initial fetch as MinID of this fetch.
		let expectedURI = `/.api/repos/${action.repo}/-/builds/${action.buildID}/tasks/${action.taskID}/log?MinID=12`;
		BuildBackend.fetch = function(url, options) {
			expect(url).to.be(expectedURI);
			return immediateSyncPromise({status: 200, text: () => immediateSyncPromise("c"), headers: {get() { return 34; }}});
		};

		expect(Dispatcher.Stores.catchDispatched(() => {
			BuildBackend.__onDispatch(new BuildActions.WantLog(action.repo, action.buildID, action.taskID));
		})).to.eql([new BuildActions.LogFetched(action.repo, action.buildID, action.taskID, 12, 34, "c")]);
	});

	it("should handle WantTasks", () => {
		let action = {
			repo: "aRepo",
			buildID: 123,
		};
		let expectedURI = `/.api/repos/${action.repo}/-/builds/${action.buildID}/tasks?PerPage=1000`;

		BuildBackend.fetch = function(url, options) {
			expect(url).to.be(expectedURI);
			return immediateSyncPromise({status: 200, json: () => [{ID: 456}]});
		};
		expect(Dispatcher.Stores.catchDispatched(() => {
			BuildBackend.__onDispatch(new BuildActions.WantTasks(action.repo, action.buildID));
		})).to.eql([new BuildActions.TasksFetched(action.repo, action.buildID, [{ID: 456}])]);
	});
});
