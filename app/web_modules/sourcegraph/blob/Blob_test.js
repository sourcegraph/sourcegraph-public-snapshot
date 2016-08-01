import autotest from "sourcegraph/util/autotest";

import * as React from "react";

import Blob from "sourcegraph/blob/Blob";

import testdataLines from "sourcegraph/blob/testdata/Blob-lines.json";
import testdataNoLineNumbers from "sourcegraph/blob/testdata/Blob-noLineNumbers.json";

const context = {
	eventLogger: {logEvent: () => null},
};

describe("Blob", () => {
	it("should render lines", () => {
		autotest(testdataLines, "sourcegraph/blob/testdata/Blob-lines.json",
			<Blob contents={"hello\nworld"} lineNumbers={true} startLine={1} endLine={2} highlightedDef="otherDef" />,
			context
		);
	});

	it("should not render line numbers by default", () => {
		autotest(testdataNoLineNumbers, "sourcegraph/blob/testdata/Blob-noLineNumbers.json",
			<Blob contents={"hello\nworld"} highlightedDef={null} />,
			context
		);
	});
});
