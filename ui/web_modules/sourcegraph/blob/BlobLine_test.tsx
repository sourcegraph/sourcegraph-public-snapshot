// tslint:disable: typedef ordered-imports

import {autotest} from "sourcegraph/util/autotest";

import * as React from "react";

import {BlobLine} from "sourcegraph/blob/BlobLine";

import testdataContents from "sourcegraph/blob/testdata/BlobLine-contents.json";
import testdataEmpty from "sourcegraph/blob/testdata/BlobLine-empty.json";
import testdataLineNumber from "sourcegraph/blob/testdata/BlobLine-lineNumber.json";
import testdataLineSelection from "sourcegraph/blob/testdata/BlobLine-selection.json";

const context = {
	eventLogger: {logEvent: () => null},
};
const common = {
	location: {
			hash: "",
			key: "",
			pathname: "",
			search: "",
			action: "",
			query: {},
			state: {},
		},
	startByte: 0,
	highlightedDefObj: null,
	activeDef: null,
	activeDefRepo: null,
};

describe("BlobLine", () => {
	it("should render", () => {
		autotest(testdataContents, "sourcegraph/blob/testdata/BlobLine-contents.json",
			<BlobLine {...common} contents={"hello\nworld"} highlightedDef="secondURL" />,
			context
		);
	});

	it("should render empty", () => {
		autotest(testdataEmpty, "sourcegraph/blob/testdata/BlobLine-empty.json",
			<BlobLine {...common} contents={"hello\nworld"} highlightedDef={null} />,
			context
		);
	});

	it("should render line number", () => {
		autotest(testdataLineNumber, "sourcegraph/blob/testdata/BlobLine-lineNumber.json",
			<BlobLine {...common} lineNumber={42} repo="r" rev="v" path="p" contents={"hello\nworld"} highlightedDef={null} />,
			context
		);
	});

	it("should render selection", () => {
		autotest(testdataLineSelection, "sourcegraph/blob/testdata/BlobLine-selection.json",
			<BlobLine {...common} contents={"hello\nworld"} selected={true} highlightedDef={null} />,
			context
		);
	});
});
