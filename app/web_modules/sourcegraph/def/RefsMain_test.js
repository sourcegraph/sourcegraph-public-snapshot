// @flow weak

import React from "react";
import expect from "expect.js";
import RefsMain from "sourcegraph/def/RefsMain";
import {renderedStatus} from "sourcegraph/app/statusTestUtils";

describe("RefsMain", () => {
	describe("status", () => {
		it("should have no error initially", () => {
			expect(renderedStatus(
				<RefsMain />
			)).to.eql({error: null});
		});

		it("should have error if the refs failed ", () => {
			expect(renderedStatus(
				<RefsMain defObj={{File: "foo.go"}} refs={{Error: true}} />
			)).to.eql({error: true});
		});

		it("should have error if the def failed", () => {
			expect(renderedStatus(
				<RefsMain defObj={{Error: true}} />
			)).to.eql({error: true});
		});

		it("should have no error if the def and refs loaded", () => {
			expect(renderedStatus(
				<RefsMain defObj={{}} refs={{Refs: []}} />
			)).to.eql({error: null});
		});
	});
});
