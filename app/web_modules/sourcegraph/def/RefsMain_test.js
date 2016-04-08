// @flow weak

import React from "react";
import expect from "expect.js";
import RefsMain from "sourcegraph/def/RefsMain";
import {renderedHTTPStatusCode} from "sourcegraph/app/httpResponseTestUtils";

describe("RefsMain", () => {
	describe("HTTP response", () => {
		it("should be null until the def is loaded", () => {
			expect(renderedHTTPStatusCode(
				<RefsMain />
			)).to.be(null);
		});

		it("should be 200 if the def exists", () => {
			expect(renderedHTTPStatusCode(
				<RefsMain defObj={{File: "foo.go"}} />
			)).to.be(200);
		});

		it("should be 404 if the def does not exist", () => {
			expect(renderedHTTPStatusCode(
				<RefsMain defObj={{Error: true}} />
			)).to.be(404);
		});
	});
});
