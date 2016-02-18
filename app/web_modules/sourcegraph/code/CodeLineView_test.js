import autotest from "sourcegraph/util/autotest";

import React from "react";

import CodeLineView from "sourcegraph/code/CodeLineView";

import testdataContents from "sourcegraph/code/testdata/CodeLineView-contents.json";
import testdataEmpty from "sourcegraph/code/testdata/CodeLineView-empty.json";
import testdataLineNumber from "sourcegraph/code/testdata/CodeLineView-lineNumber.json";
import testdataLineSelection from "sourcegraph/code/testdata/CodeLineView-selection.json";

describe("CodeLineView", () => {
	it("should render", () => {
		autotest(testdataContents, `${__dirname}/testdata/CodeLineView-contents.json`,
			<CodeLineView contents={"hello\nworld"} highlightedDef="secondURL" />
		);
	});

	it("should render empty", () => {
		autotest(testdataEmpty, `${__dirname}/testdata/CodeLineView-empty.json`,
			<CodeLineView contents={"hello\nworld"} />
		);
	});

	it("should render line number", () => {
		autotest(testdataLineNumber, `${__dirname}/testdata/CodeLineView-lineNumber.json`,
			<CodeLineView lineNumber={42} contents={"hello\nworld"} />
		);
	});

	it("should render selection", () => {
		autotest(testdataLineSelection, `${__dirname}/testdata/CodeLineView-selection.json`,
			<CodeLineView contents={"hello\nworld"} selected={true} />
		);
	});
});
