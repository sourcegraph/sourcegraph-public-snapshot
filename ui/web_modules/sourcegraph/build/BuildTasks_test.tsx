import * as React from "react";
import {BuildTasks} from "sourcegraph/build/BuildTasks";
import testdataInitial from "sourcegraph/build/testdata/BuildTasks-initial.json";
import {autotest} from "sourcegraph/util/autotest";

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
	it("should render", () => {
		autotest(testdataInitial, "sourcegraph/build/testdata/BuildTasks-initial.json",
			<BuildTasks
				tasks={sampleTasks}
				logs={{get(): any { return null; }}} />
		);
	});
});
