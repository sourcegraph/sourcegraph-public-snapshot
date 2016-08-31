import * as React from "react";
import testdataEmpty from "sourcegraph/build/testdata/TopLevelTask-empty.json";
import testdataSteps from "sourcegraph/build/testdata/TopLevelTask-steps.json";
import {TopLevelTask} from "sourcegraph/build/TopLevelTask";
import {autotest} from "sourcegraph/util/testutil/autotest";

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
				logs={{get(): any { return null; }}} />
		);
	});

	it("should render steps", () => {
		autotest(testdataSteps, "sourcegraph/build/testdata/TopLevelTask-steps.json",
			<TopLevelTask
				task={sampleTask}
				subtasks={[sampleTask, sampleTask]}
				logs={{get(): any { return {log: "a"}; }}} />
		);
	});
});
