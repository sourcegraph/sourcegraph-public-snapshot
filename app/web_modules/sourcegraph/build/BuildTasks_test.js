import autotest from "sourcegraph/util/autotest";

import * as React from "react";

import BuildTasks from "sourcegraph/build/BuildTasks";

import testdataInitial from "sourcegraph/build/testdata/BuildTasks-initial.json";
import testdataActive from "sourcegraph/build/testdata/BuildTasks-active.json";

const sampleTasks = [
	{
		ID: 456,
		Build: {Repo: {URI: "aRepo"}, ID: 123},
	},
	{
		ID: 567,
		Build: {Repo: {URI: "aRepo"}, ID: 234},
	},
];

describe("BuildTasks", () => {
	it("should initially render empty", () => {
		autotest(testdataInitial, `${__dirname}/testdata/BuildTasks-initial.json`,
			<BuildTasks
				location=""
				tasks={sampleTasks}
				logs={{get() { return null; }}} />
		);
	});

	it("should render task based on URL", () => {
		autotest(testdataActive, `${__dirname}/testdata/BuildTasks-active.json`,
			<BuildTasks
				location="#T567"
				tasks={sampleTasks}
				logs={{get() { return null; }}} />
		);
	});
});
