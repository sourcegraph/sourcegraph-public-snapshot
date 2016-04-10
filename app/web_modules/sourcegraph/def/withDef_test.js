// @flow weak

import React from "react";
import expect from "expect.js";
import withDef from "sourcegraph/def/withDef";
import {renderedStatus} from "sourcegraph/app/statusTestUtils";
import DefStore from "sourcegraph/def/DefStore";
import * as DefActions from "sourcegraph/def/DefActions";
import {rel as relPath} from "sourcegraph/app/routePatterns";

const C = withDef((props) => null);

const props = {
	params: {splat: [null, "d"]},
	route: {path: relPath.def},
};

describe("withDef", () => {
	describe("status", () => {
		it("should have no error initially", () => {
			expect(renderedStatus(
				<C repo="r" rev="v" {...props} />
			)).to.eql({error: null});
		});

		it("should have no error if the def and rev exist", () => {
			DefStore.directDispatch(new DefActions.DefFetched("r", "v", "d", {}));
			expect(renderedStatus(
				<C repo="r" rev="v" {...props} />
			)).to.eql({error: null});
		});

		it("should have error if the def does not exist", () => {
			DefStore.directDispatch(new DefActions.DefFetched("r", "v", "d", {Error: true}));
			expect(renderedStatus(
				<C repo="r" rev="v" {...props} />
			)).to.eql({error: true});
		});
	});
});
