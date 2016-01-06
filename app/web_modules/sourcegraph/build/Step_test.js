import autotest from "sourcegraph/util/autotest";

import React from "react";

import Step from "sourcegraph/build/Step";

import testdataInitial from "sourcegraph/build/testdata/Step-initial.json";
import testdataFetchedLog from "sourcegraph/build/testdata/Step-fetchedLog.json";

const sampleTask = {
	ID: 456,
	Build: {Repo: {URI: "aRepo"}, ID: 123},
};

describe("Step", () => {
	it("should initially render empty and want log", () => {
		autotest(testdataInitial, `${__dirname}/testdata/Step-initial.json`,
			<Step
				task={sampleTask}
				logs={{get() { return null; }}} />
		);
	});

	it("should render log", () => {
		autotest(testdataFetchedLog, `${__dirname}/testdata/Step-fetchedLog.json`,
			<Step
				task={sampleTask}
				logs={{get() { return {log: "a"}; }}} />
		);
	});
});
