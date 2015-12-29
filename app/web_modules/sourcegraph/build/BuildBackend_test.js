import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import BuildBackend from "sourcegraph/build/BuildBackend";
import {BuildStore} from "sourcegraph/build/BuildStore";
import * as BuildActions from "sourcegraph/build/BuildActions";

beforeEach(() => {
	BuildBackend.buildStore = new BuildStore(Dispatcher);
});

describe("BuildBackend", () => {
	it("should handle WantBuild", () => {
		let action = {
			repo: "aRepo",
			buildID: 123,
		};
		let expectedURI = `/.api/repos/${action.repo}/.builds/${action.buildID}`;

		BuildBackend.xhr = function(options, callback) {
			expect(options.uri).to.be(expectedURI);
			callback(null, null, {ID: 123});
		};
		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(BuildBackend, new BuildActions.WantBuild(action.repo, action.buildID));
		})).to.eql([new BuildActions.BuildFetched(action.repo, action.buildID, {ID: 123})]);
	});
});

describe("BuildBackend", () => {
	it("should handle WantLog", () => {
		let action = {
			repo: "aRepo",
			buildID: 123,
			taskID: 456,
		};
		let expectedURI = `/${action.repo}/.builds/${action.buildID}/tasks/${action.taskID}/log`;

		BuildBackend.xhr = function(options, callback) {
			expect(options.uri).to.be(expectedURI);
			callback(null, {headers: {"x-sourcegraph-log-max-id": 789}}, "a");
		};
		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(BuildBackend, new BuildActions.WantLog(action.repo, action.buildID, action.taskID));
		})).to.eql([new BuildActions.LogFetched(action.repo, action.buildID, action.taskID, 0, 789, "a")]);
	});

	it("should reuse the MaxID to allow efficient tailing", () => {
		let action = {
			repo: "aRepo",
			buildID: 123,
			taskID: 456,
		};

		// Mock the previous fetch as having returned a maxID of 12.
		BuildBackend.buildStore = {logs: {get() { return {maxID: 12, log: "a\n"}; }}};

		// Trigger "second" fetch, which should reuse MaxID from
		// initial fetch as MinID of this fetch.
		let expectedURI = `/${action.repo}/.builds/${action.buildID}/tasks/${action.taskID}/log?MinID=12`;
		BuildBackend.xhr = function(options, callback) {
			expect(options.uri).to.be(expectedURI);
			callback(null, {headers: {"x-sourcegraph-log-max-id": 34}}, "c");
		};

		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(BuildBackend, new BuildActions.WantLog(action.repo, action.buildID, action.taskID));
		})).to.eql([new BuildActions.LogFetched(action.repo, action.buildID, action.taskID, 12, 34, "c")]);
	});
});

describe("BuildBackend", () => {
	it("should handle WantTasks", () => {
		let action = {
			repo: "aRepo",
			buildID: 123,
		};
		let expectedURI = `/.api/repos/${action.repo}/.builds/${action.buildID}/.tasks`;

		BuildBackend.xhr = function(options, callback) {
			expect(options.uri).to.be(expectedURI);
			callback(null, null, [{ID: 456}]);
		};
		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(BuildBackend, new BuildActions.WantTasks(action.repo, action.buildID));
		})).to.eql([new BuildActions.TasksFetched(action.repo, action.buildID, [{ID: 456}])]);
	});
});
