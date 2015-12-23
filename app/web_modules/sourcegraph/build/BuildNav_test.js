import autotest from "sourcegraph/util/autotest";

import React from "react";

import BuildNav from "sourcegraph/build/BuildNav";

import testdataInitial from "sourcegraph/build/testdata/BuildNav-initial.json";
import testdataAvailable from "sourcegraph/build/testdata/BuildNav-available.json";

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

describe("BuildNav", () => {
	it("should initially render empty", () => {
		autotest(testdataInitial, `${__dirname}/testdata/BuildNav-initial.json`,
			<BuildNav
				location=""
				tasks={sampleTasks}
				logs={{get() { return null; }}} />
		);
	});

	it("should render items", () => {
		autotest(testdataAvailable, `${__dirname}/testdata/BuildNav-available.json`,
			<BuildNav
				location="#T567"
				tasks={sampleTasks}
				logs={{get() { return null; }}} />
		);
	});
});
