// @flow weak

import React from "react";
import expect from "expect.js";
import withRepoBuild from "sourcegraph/build/withRepoBuild";
import {render} from "sourcegraph/util/renderTestUtils";
import BuildStore from "sourcegraph/build/BuildStore";
import * as BuildActions from "sourcegraph/build/BuildActions";

const C = withRepoBuild((props) => null);

describe("withRepoBuild", () => {
	it("should render initially", () => {
		render(<C repo="r" />);
	});

	it("should render with commit ID", () => {
		let res = render(<C repo="r" commitID="c" />);
		expect(res.actions).to.eql([new BuildActions.WantNewestBuildForCommit("r", "c", false)]);
	});

	it("should render if the build and rev exist", () => {
		BuildStore.directDispatch(new BuildActions.BuildsFetchedForCommit("r", "c", [{ID: 1}]));
		render(<C repo="r" commitID="c" />);
	});

	it("should render if the build does not exist", () => {
		BuildStore.directDispatch(new BuildActions.BuildsFetchedForCommit("r", "c", []));
		render(<C repo="r" commitID="c" />);
	});
});
