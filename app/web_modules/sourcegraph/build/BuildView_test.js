import autotest from "sourcegraph/util/autotest";

import React from "react";

import * as BuildActions from "sourcegraph/build/BuildActions";
import BuildStore from "sourcegraph/build/BuildStore";
import BuildView from "sourcegraph/build/BuildView";
import Dispatcher from "sourcegraph/Dispatcher";

import testdataInitial from "sourcegraph/build/testdata/BuildView-initial.json";
import testdataAvailable from "sourcegraph/build/testdata/BuildView-available.json";
import testdataBuildSuccess from "sourcegraph/build/testdata/BuildView-buildSuccess.json";

describe("BuildView", () => {
	it("should initially render empty and want build and tasks", () => {
		autotest(testdataInitial, `${__dirname}/testdata/BuildView-initial.json`,
			<BuildView
				build={{ID: 123}}
				commit={{ID: "abc"}} />
		);
	});

	it("should render tasks", () => {
		Dispatcher.directDispatch(BuildStore, new BuildActions.TasksFetched("aRepo", 123, {BuildTasks: [{ID: 456}]}));
		autotest(testdataAvailable, `${__dirname}/testdata/BuildView-available.json`,
			<BuildView
				build={{ID: 123, Repo: "aRepo"}}
				commit={{ID: "abc"}} />
		);
	});

	it("should render updated build", () => {
		Dispatcher.directDispatch(BuildStore, new BuildActions.BuildFetched("aRepo", 123, {ID: 456, Success: true}));
		autotest(testdataBuildSuccess, `${__dirname}/testdata/BuildView-buildSuccess.json`,
			<BuildView
				build={{ID: 123, Repo: "aRepo"}}
				commit={{ID: "abc"}} />
		);
	});
});
