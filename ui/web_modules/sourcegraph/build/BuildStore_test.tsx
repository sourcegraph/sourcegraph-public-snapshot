// tslint:disable: typedef ordered-imports

import expect from "expect.js";

import {BuildStore} from "sourcegraph/build/BuildStore";
import * as BuildActions from "sourcegraph/build/BuildActions";

describe("BuildStore", () => {
	it("should handle BuildFetched", () => {
		BuildStore.directDispatch(new BuildActions.BuildFetched("aRepo", 123, {ID: 123}));
		expect(BuildStore.builds.get("aRepo", 123)).to.eql({ID: 123});
	});

	it("should handle LogFetched", () => {
		BuildStore.directDispatch(new BuildActions.LogFetched("aRepo", 123, 456, 12, 34, "a"));
		expect(BuildStore.logs.get("aRepo", 123, 456)).to.eql({maxID: 34, log: "a"});
	});

	it("should append cleanly to the log when handling a sequential LogFetched", () => {
		BuildStore.directDispatch(new BuildActions.LogFetched("bRepo", 123, 456, 0, 34, "a"));
		BuildStore.directDispatch(new BuildActions.LogFetched("bRepo", 123, 456, 34, 56, "b"));
		expect(BuildStore.logs.get("bRepo", 123, 456)).to.eql({maxID: 56, log: "ab"});
	});

	it("should handle TasksFetched", () => {
		BuildStore.directDispatch(new BuildActions.TasksFetched("aRepo", 123, [{ID: 456}]));
		expect(BuildStore.tasks.get("aRepo", 123)).to.eql([{ID: 456}]);
	});
});
