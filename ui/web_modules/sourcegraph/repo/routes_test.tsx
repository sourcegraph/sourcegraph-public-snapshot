import expect from "expect.js";
import { rootRoute } from "sourcegraph/app/App";
import { repoRoutes } from "sourcegraph/app/routes/repoRoutes";
import { urlWithRev } from "sourcegraph/repo/routes";

describe("urlWithRev", () => {
	let tests = [
		{ want: "/r@newRev", routes: [rootRoute, repoRoutes[1]], routeParams: { splat: "r" } },
		{ want: "/r@newRev", routes: [rootRoute, repoRoutes[1]], routeParams: { splat: "r@v" } },
		{ want: "/r@newRev", routes: [rootRoute, repoRoutes[1]], routeParams: { splat: "r@v1/v2" } },
		{ want: "/r@newRev/-/a/q/p", routes: [rootRoute, repoRoutes[0], { path: "a/*" }], routeParams: { splat: ["r", "q/p"] } },
		{ want: "/r@newRev/-/a/q/p", routes: [rootRoute, repoRoutes[0], { path: "a/*" }], routeParams: { splat: ["r@v", "q/p"] } },
		{ want: "/r@newRev/-/a/q/p", routes: [rootRoute, repoRoutes[0], { path: "a/*" }], routeParams: { splat: ["r@v1/v2", "q/p"] } },
	];
	tests.forEach((test) => {
		it(`should produce ${test.want}`, () => {
			expect(urlWithRev(test.routes, test.routeParams, "newRev")).to.be(test.want);
		});
	});
});
