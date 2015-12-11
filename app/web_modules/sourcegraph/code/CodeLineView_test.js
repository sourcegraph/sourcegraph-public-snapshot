import autotest from "sourcegraph/util/autotest";

import React from "react";

import CodeLineView from "sourcegraph/code/CodeLineView";

import testdataTokens from "sourcegraph/code/testdata/CodeLineView-tokens.json";
import testdataEmpty from "sourcegraph/code/testdata/CodeLineView-empty.json";
import testdataLineNumber from "sourcegraph/code/testdata/CodeLineView-lineNumber.json";
import testdataLineSelection from "sourcegraph/code/testdata/CodeLineView-selection.json";

describe("CodeLineView", () => {
	it("should render tokens", () => {
		autotest(testdataTokens, `${__dirname}/testdata/CodeLineView-tokens.json`,
			<CodeLineView tokens={[
				{Label: "foo"},
				{Label: "bar", Class: "b"},
				{Label: "baz", Class: "c"},
				{Label: "ref", Class: "d", URL: ["someURL"]},
				{Label: "def", Class: "e", URL: ["firstURL", "secondURL"], IsDef: true},
			]} selectedDef="someURL" highlightedDef="secondURL" />
		);
	});

	it("should render empty token list", () => {
		autotest(testdataEmpty, `${__dirname}/testdata/CodeLineView-empty.json`,
			<CodeLineView tokens={[]} />
		);
	});

	it("should render line number", () => {
		autotest(testdataLineNumber, `${__dirname}/testdata/CodeLineView-lineNumber.json`,
			<CodeLineView lineNumber={42} tokens={[{Label: "foo"}]} />
		);
	});

	it("should render selection", () => {
		autotest(testdataLineSelection, `${__dirname}/testdata/CodeLineView-selection.json`,
			<CodeLineView tokens={[{Label: "foo"}]} selected={true} />
		);
	});
});
