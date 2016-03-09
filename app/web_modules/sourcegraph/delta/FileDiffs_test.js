import autotest from "sourcegraph/util/autotest";

import React from "react";

import FileDiffs from "sourcegraph/delta/FileDiffs";

import testdataInitial from "sourcegraph/delta/testdata/FileDiffs-initial.json";

const sampleFiles = [{OrigName: "a", NewName: "b"}, {OrigName: "c", NewName: "d"}];

describe("FileDiffs", () => {
	it("should render", () => {
		autotest(testdataInitial, `${__dirname}/testdata/FileDiffs-initial.json`,
			<FileDiffs files={sampleFiles} stats={{}} baseRepo="br" baseRev="bv" headRepo="hr" headRev="hv" />
		);
	});
});
