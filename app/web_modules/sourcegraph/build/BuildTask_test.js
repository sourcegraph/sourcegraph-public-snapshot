import autotest from "sourcegraph/util/autotest";

import React from "react";

import BuildTask from "sourcegraph/build/BuildTask";

import testdataInitial from "sourcegraph/build/testdata/BuildTask-initial.json";
import testdataFetchedLog from "sourcegraph/build/testdata/BuildTask-fetchedLog.json";

const sampleTask = {
	ID: 456,
	Build: {Repo: {URI: "aRepo"}, ID: 123},
};

describe("BuildTask", () => {
	it("should initially render empty and want log", () => {
		autotest(testdataInitial, `${__dirname}/testdata/BuildTask-initial.json`,
			<BuildTask
				task={sampleTask}
				logs={{get() { return null; }}} />
		);
	});

	it("should render log", () => {
		autotest(testdataFetchedLog, `${__dirname}/testdata/BuildTask-fetchedLog.json`,
			<BuildTask
				task={sampleTask}
				logs={{get() { return {log: "a"}; }}} />
		);
	});
});
