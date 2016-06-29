import autotest from "sourcegraph/util/autotest";

import React from "react";

import DiffFileList from "sourcegraph/delta/DiffFileList";

import testdataInitial from "sourcegraph/delta/testdata/DiffFileList-initial.json";

const sampleStats = {
	Added: 5,
	Changed: 6,
	Deleted: 7,
};

const sampleFiles = [
	{
		OrigName: "a",
		NewName: "b",
		Stats: sampleStats,
	},
	{
		OrigName: "/dev/null",
		NewName: "b",
		Stats: sampleStats,
	},
	{
		OrigName: "a",
		NewName: "/dev/null",
		Stats: sampleStats,
	},
];

describe("DiffFileList", () => {
	it("should render", () => {
		autotest(testdataInitial, `${__dirname}/testdata/DiffFileList-initial.json`,
			<DiffFileList files={sampleFiles} stats={sampleStats} />
		);
	});
});
