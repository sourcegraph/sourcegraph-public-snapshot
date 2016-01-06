import autotest from "sourcegraph/util/autotest";

import React from "react";

import * as BuildActions from "sourcegraph/build/BuildActions";
import BuildStore from "sourcegraph/build/BuildStore";
import BuildContainer from "sourcegraph/build/BuildContainer";
import Dispatcher from "sourcegraph/Dispatcher";

import testdataInitial from "sourcegraph/build/testdata/BuildContainer-initial.json";
import testdataAvailable from "sourcegraph/build/testdata/BuildContainer-available.json";
import testdataBuildSuccess from "sourcegraph/build/testdata/BuildContainer-buildSuccess.json";

describe("BuildContainer", () => {
	it("should initially render empty and want build and tasks", () => {
		autotest(testdataInitial, `${__dirname}/testdata/BuildContainer-initial.json`,
			<BuildContainer
				build={{ID: 123}}
				commit={{ID: "abc"}} />
		);
	});

	it("should render tasks", () => {
		Dispatcher.directDispatch(BuildStore, new BuildActions.TasksFetched("aRepo", 123, {BuildTasks: [{ID: 456}]}));
		autotest(testdataAvailable, `${__dirname}/testdata/BuildContainer-available.json`,
			<BuildContainer
				build={{ID: 123, Repo: "aRepo"}}
				commit={{ID: "abc"}} />
		);
	});

	it("should render updated build", () => {
		Dispatcher.directDispatch(BuildStore, new BuildActions.BuildFetched("aRepo", 123, {ID: 456, Success: true}));
		autotest(testdataBuildSuccess, `${__dirname}/testdata/BuildContainer-buildSuccess.json`,
			<BuildContainer
				build={{ID: 123, Repo: "aRepo"}}
				commit={{ID: "abc"}} />
		);
	});
});
