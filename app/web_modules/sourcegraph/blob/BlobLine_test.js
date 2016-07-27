// @flow weak

import autotest from "sourcegraph/util/autotest";

import * as React from "react";

import BlobLine from "sourcegraph/blob/BlobLine";

import testdataContents from "sourcegraph/blob/testdata/BlobLine-contents.json";
import testdataEmpty from "sourcegraph/blob/testdata/BlobLine-empty.json";
import testdataLineNumber from "sourcegraph/blob/testdata/BlobLine-lineNumber.json";
import testdataLineSelection from "sourcegraph/blob/testdata/BlobLine-selection.json";

describe("BlobLine", () => {
	it("should render", () => {
		autotest(testdataContents, `${__dirname}/testdata/BlobLine-contents.json`,
			<BlobLine contents={"hello\nworld"} startByte={0} highlightedDef="secondURL" />
		);
	});

	it("should render empty", () => {
		autotest(testdataEmpty, `${__dirname}/testdata/BlobLine-empty.json`,
			<BlobLine contents={"hello\nworld"} startByte={0} />
		);
	});

	it("should render line number", () => {
		autotest(testdataLineNumber, `${__dirname}/testdata/BlobLine-lineNumber.json`,
			<BlobLine lineNumber={42} repo="r" rev="v" path="p" contents={"hello\nworld"} startByte={0} />
		);
	});

	it("should render selection", () => {
		autotest(testdataLineSelection, `${__dirname}/testdata/BlobLine-selection.json`,
			<BlobLine contents={"hello\nworld"} startByte={0} selected={true} />
		);
	});
});
