import autotest from "sourcegraph/util/autotest";

import React from "react";
import {btoa} from "abab";

import Hunk from "sourcegraph/delta/Hunk";

import testdataInitial from "sourcegraph/delta/testdata/Hunk-initial.json";

const sampleHunk = {
	OrigStartLine: 5,
	OrigLines: 20,
	NewStartLine: 5,
	NewLines: 20,
	Section: "mysection",
	Body: btoa("+ a\n- b\n c\n- d\n- e"),
};

describe("Hunk", () => {
	it("should render", () => {
		autotest(testdataInitial, `${__dirname}/testdata/Hunk-initial.json`,
			<Hunk hunk={sampleHunk} />
		);
	});
});
