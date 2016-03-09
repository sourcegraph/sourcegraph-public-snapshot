import autotest from "sourcegraph/util/autotest";

import React from "react";

import DiffStatScale from "sourcegraph/delta/DiffStatScale";

import testdataInitial from "sourcegraph/delta/testdata/DiffStatScale-initial.json";

const sampleStats = {
	Added: 5,
	Changed: 6,
	Deleted: 7,
};

describe("DiffStatScale", () => {
	it("should render", () => {
		autotest(testdataInitial, `${__dirname}/testdata/DiffStatScale-initial.json`,
			<DiffStatScale Stat={sampleStats} />
		);
	});
});
