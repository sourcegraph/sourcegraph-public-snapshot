// @flow weak

import React from "react";
import expect from "expect.js";
import RepoMain from "sourcegraph/repo/RepoMain";
import {renderedHTTPStatusCode} from "sourcegraph/app/httpResponseTestUtils";

describe("RepoMain", () => {
	describe("HTTP response", () => {
		it("should be null until the repo and rev are both loaded", () => {
			expect(renderedHTTPStatusCode(
				<RepoMain />
			)).to.be(null);

			expect(renderedHTTPStatusCode(
				<RepoMain repoObj={{}} repo="r" />
			)).to.be(null);

			expect(renderedHTTPStatusCode(
				<RepoMain rev="v" />
			)).to.be(null);
		});

		it("should be 200 if the repo and rev exist", () => {
			expect(renderedHTTPStatusCode(
				<RepoMain repo="r" repoObj={{}} rev="v" />
			)).to.be(200);
		});

		it("should be 404 if the repo does not exist", () => {
			expect(renderedHTTPStatusCode(
				<RepoMain repo="r" repoObj={{Error: true}} />
			)).to.be(404);
		});
	});
});
