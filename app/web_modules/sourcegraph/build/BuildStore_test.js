import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import BuildStore from "sourcegraph/build/BuildStore";
import * as BuildActions from "sourcegraph/build/BuildActions";

afterEach(BuildStore.reset.bind(BuildStore));
beforeEach(BuildStore.reset.bind(BuildStore));

describe("BuildStore", () => {
	it("should handle BuildFetched", () => {
		Dispatcher.directDispatch(BuildStore, new BuildActions.BuildFetched("aRepo", 123, {ID: 123}));
		expect(BuildStore.builds.get("aRepo", 123)).to.eql({ID: 123});
	});
});

describe("BuildStore", () => {
	it("should handle LogFetched", () => {
		Dispatcher.directDispatch(BuildStore, new BuildActions.LogFetched("aRepo", 123, 456, 12, 34, "a"));
		expect(BuildStore.logs.get("aRepo", 123, 456)).to.eql({maxID: 34, log: "a"});
	});

	it("should append cleanly to the log when handling a sequential LogFetched", () => {
		Dispatcher.directDispatch(BuildStore, new BuildActions.LogFetched("bRepo", 123, 456, 0, 34, "a"));
		Dispatcher.directDispatch(BuildStore, new BuildActions.LogFetched("bRepo", 123, 456, 34, 56, "b"));
		expect(BuildStore.logs.get("bRepo", 123, 456)).to.eql({maxID: 56, log: "ab"});
	});
});

describe("BuildStore", () => {
	it("should handle TasksFetched", () => {
		Dispatcher.directDispatch(BuildStore, new BuildActions.TasksFetched("aRepo", 123, [{ID: 456}]));
		expect(BuildStore.tasks.get("aRepo", 123)).to.eql([{ID: 456}]);
	});
});
