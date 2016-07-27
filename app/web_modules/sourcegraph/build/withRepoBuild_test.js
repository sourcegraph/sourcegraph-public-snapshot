// @flow weak

import * as React from "react";
import expect from "expect.js";
import withRepoBuild from "sourcegraph/build/withRepoBuild";
import {render} from "sourcegraph/util/renderTestUtils";
import BuildStore from "sourcegraph/build/BuildStore";
import * as BuildActions from "sourcegraph/build/BuildActions";

const C = withRepoBuild((props) => null);

describe("withRepoBuild", () => {
	it("should render initially with no repo ID", () => {
		render(<C />);
	});

	it("should render with commit ID", () => {
		let res = render(<C repoID={1} commitID="c" />);
		expect(res.actions).to.eql([new BuildActions.WantNewestBuildForCommit(1, "c", false)]);
	});

	it("should render if the build and rev exist", () => {
		BuildStore.directDispatch(new BuildActions.BuildsFetchedForCommit(1, "c", [{ID: 1}]));
		render(<C repoID={1} commitID="c" />);
	});

	it("should render if the build does not exist", () => {
		BuildStore.directDispatch(new BuildActions.BuildsFetchedForCommit(1, "c", []));
		render(<C repoID={1} commitID="c" />);
	});
});
