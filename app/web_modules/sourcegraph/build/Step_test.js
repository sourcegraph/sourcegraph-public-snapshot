import autotest from "sourcegraph/util/autotest";

import * as React from "react";

import Step from "sourcegraph/build/Step";

import testdataInitial from "sourcegraph/build/testdata/Step-initial.json";
import testdataFetchedLog from "sourcegraph/build/testdata/Step-fetchedLog.json";
import testdataSuccess from "sourcegraph/build/testdata/Step-success.json";
import testdataWarnings from "sourcegraph/build/testdata/Step-warnings.json";
import testdataFailure from "sourcegraph/build/testdata/Step-failure.json";

const sampleTask = {
	ID: 456,
	Build: {Repo: {URI: "aRepo"}, ID: 123},
};

const taskSuccess = {
	ID: 5,
	Build: {
		Repo: 57215,
		ID: 8,
	},
	ParentID: 4,
	Label: "Clone",
	CreatedAt: "2016-07-04T14:37:08.497997Z",
	StartedAt: "2016-07-04T14:37:08.606152Z",
	EndedAt: "2016-07-04T14:37:14.48451Z",
	Success: true,
};

const taskWarnings = {
	ID: 6,
	Build: {
		Repo: 57215,
		ID: 8,
	},
	ParentID: 4,
	Label: "Warning: Can't automatically generate CI build config for C, Shell, C++, D, GAS, Logos, Perl, Nemerle; please create a .sg-drone.yml file",
	CreatedAt: "2016-07-04T14:37:14.49782Z",
	StartedAt: "2016-07-04T14:37:14.592688Z",
	EndedAt: "2016-07-04T14:37:15.147179Z",
	Success: true,
	Warnings: true,
};

const taskFailure = {
	ID: 8,
	Build: {
		Repo: 57215,
		ID: 8,
	},
	ParentID: 4,
	Label: "Bash (indexing)",
	CreatedAt: "2016-07-04T14:40:45.556814Z",
	StartedAt: "2016-07-04T14:40:45.576537Z",
	EndedAt: "2016-07-04T14:40:46.755126Z",
	Success: true,
	Failure: true,
};

describe("Step", () => {
	it("should initially render empty and want log", () => {
		autotest(testdataInitial, "sourcegraph/build/testdata/Step-initial.json",
			<Step
				task={sampleTask}
				logs={{get() { return null; }}} />
		);
	});

	it("should render log", () => {
		autotest(testdataFetchedLog, "sourcegraph/build/testdata/Step-fetchedLog.json",
			<Step
				task={sampleTask}
				logs={{get() { return {log: "a"}; }}} />
		);
	});

	it("should render success", () => {
		autotest(testdataSuccess, "sourcegraph/build/testdata/Step-success.json",
			<Step
				task={taskSuccess}
				logs={{get() { return null; }}} />
		);
	});

	it("should render warnings", () => {
		autotest(testdataWarnings, "sourcegraph/build/testdata/Step-warnings.json",
			<Step
				task={taskWarnings}
				logs={{get() { return null; }}} />
		);
	});

	it("should render failure", () => {
		autotest(testdataFailure, "sourcegraph/build/testdata/Step-failure.json",
			<Step
				task={taskFailure}
				logs={{get() { return null; }}} />
		);
	});
});
