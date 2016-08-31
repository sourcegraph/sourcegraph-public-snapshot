import * as React from "react";
import {DefTooltip} from "sourcegraph/def/DefTooltip";
import testdataData from "sourcegraph/def/testdata/DefTooltip-data.json";
import {autotest} from "sourcegraph/util/testutil/autotest";

const fmtStrings = {DefKeyword: "a", NameAndTypeSeparator: "s", Name: {ScopeQualified: "n"}, Type: {ScopeQualified: "t"}};

describe("DefTooltip", () => {
	it("should render definition data", () => {
		let testPos = {repo: "foo", commit: "aaa", file: "bar", line: 42, character: 3};
		let hoverInfos = {
			get(pos: any): any {
				return pos === testPos ? {def: {URL: "someURL", FmtStrings: fmtStrings, DocHTML: "someDoc", Repo: "someRepo"}} : null;
			},
		};
		autotest(testdataData, "sourcegraph/def/testdata/DefTooltip-data.json",
			<DefTooltip hoverPos={testPos} hoverInfos={hoverInfos} />
		);
	});
});
