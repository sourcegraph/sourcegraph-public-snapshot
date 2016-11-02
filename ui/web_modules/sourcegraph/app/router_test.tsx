import expect from "expect.js";
import matchRoutes from "react-router/lib/matchRoutes";
import { repoRoutes } from "sourcegraph/app/routes/repoRoutes";

describe("repo route", () => {
	const tests = {
		"/a": { params: { splat: "a" } },
		"/a/b": { params: { splat: "a/b" } },
		"/a/b/-/blob/f": { params: { splat: ["a/b", "f"] } },
	};
	Object.keys(tests).forEach((pathname) => {
		it(`matches '${pathname}'`, (done) => {
			matchRoutes(repoRoutes, { pathname: pathname }, (err, state) => {
				expect(err).to.be(null);
				if (!err) {
					expect(state.params).to.eql(tests[pathname].params);
				}
				done();
			});
		});
	});
});
