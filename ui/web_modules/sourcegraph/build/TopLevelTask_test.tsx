// tslint:disable

import {autotest} from "sourcegraph/util/autotest";

import * as React from "react";

import {TopLevelTask} from "sourcegraph/build/TopLevelTask";

import testdataEmpty from "sourcegraph/build/testdata/TopLevelTask-empty.json";
import testdataSteps from "sourcegraph/build/testdata/TopLevelTask-steps.json";

const sampleTask = {
	ID: 456,
	Build: {Repo: {URI: "aRepo"}, ID: 123},
};

describe("TopLevelTask", () => {
	it("should render empty", () => {
		autotest(testdataEmpty, "sourcegraph/build/testdata/TopLevelTask-empty.json",
			<TopLevelTask
				task={sampleTask}
				subtasks={[]}
				logs={{get() { return null; }}} />
		);
	});

	it("should render steps", () => {
		autotest(testdataSteps, "sourcegraph/build/testdata/TopLevelTask-steps.json",
			<TopLevelTask
				task={sampleTask}
				subtasks={[sampleTask, sampleTask]}
				logs={{get() { return {log: "a"}; }}} />
		);
	});
});
