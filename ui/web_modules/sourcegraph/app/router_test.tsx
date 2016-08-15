import expect from "expect.js";
import matchRoutes from "react-router/lib/matchRoutes";
import {routes as repoRoutes} from "sourcegraph/repo/routes";

describe("repo route", () => {
	const tests = {
		"/a": {params: {splat: "a"}},
		"/a/b": {params: {splat: "a/b"}},
		"/a/b/-/def/d/d/-/d": {params: {splat: ["a/b", "d/d/-/d"]}},
	};
	Object.keys(tests).forEach((pathname) => {
		it(`matches '${pathname}'`, (done) => {
			matchRoutes(repoRoutes, {pathname: pathname}, (err, state) => {
				expect(err).to.be(null);
				if (!err) {
					expect(state.params).to.eql(tests[pathname].params);
				}
				done();
			});
		});
	});
});
