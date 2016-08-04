// tslint:disable

import {repoPath, repoRev} from "sourcegraph/repo/index";
import expect from "expect.js";

describe("repoRev and repoPath", () => {
	let tests = {
		"a": ["a", null],
		"a/b": ["a/b", null],
		"a@": ["a", null],
		"a@b": ["a", "b"],
		"a@b/c": ["a", "b/c"],
		"a/b@c/d": ["a/b", "c/d"],
	};
	Object.keys(tests).forEach((repoRevRouteVar) => {
		const want = tests[repoRevRouteVar];
		it(`should parse '${repoRevRouteVar}'`, () => {
			expect(repoPath(repoRevRouteVar)).to.eql(want[0]);
			expect(repoRev(repoRevRouteVar)).to.eql(want[1]);
		});
	});
});
