import autotest from "sourcegraph/util/autotest";

import React from "react";

import DiffFileList from "sourcegraph/delta/DiffFileList";

import testdataInitial from "sourcegraph/delta/testdata/DiffFileList-initial.json";

const sampleFiles = [
	{
		OrigName: "a",
		NewName: "b",
	},
	{
		OrigName: "/dev/null",
		NewName: "b",
	},
	{
		OrigName: "a",
		NewName: "/dev/null",
	},
];

const sampleStats = {
	Added: 5,
	Changed: 6,
	Deleted: 7,
};

describe("DiffFileList", () => {
	it("should render", () => {
		autotest(testdataInitial, `${__dirname}/testdata/DiffFileList-initial.json`,
			<DiffFileList files={sampleFiles} stats={sampleStats} />
		);
	});
});
