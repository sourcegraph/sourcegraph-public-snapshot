// tslint:disable

import {abs as absRoutes, getRouteParams} from "sourcegraph/app/routePatterns";
import expect from "expect.js";

describe("routePatterns", () => {
	let tests = {
		home: {
			"": {},
		},
		repo: {
			"r": {splat: "r"},
			"r/r": {splat: "r/r"},
			"r/r@c": {splat: "r/r@c"},
			"r/r@c/c": {splat: "r/r@c/c"},
		},
		def: {
			"r/-/def/d": {splat: ["r", "d"]},
			"r/r/-/def/d/d": {splat: ["r/r", "d/d"]},
			"r/r@c/-/def/d/d/-/d": {splat: ["r/r@c", "d/d/-/d"]},
		},
		blob: {
			"r/-/blob/b": {splat: ["r", "b"]},
			"r/r/-/blob/b/b": {splat: ["r/r", "b/b"]},
			"r/r@c/-/blob/b/b": {splat: ["r/r@c", "b/b"]},
		},
	};

	Object.keys(tests).forEach((name) => {
		Object.keys(tests[name]).forEach((url) => {
			it(`matches ${name} URL '${url}' against route pattern '${absRoutes[name]}'`, () => {
				expect(getRouteParams(absRoutes[name], url)).to.eql(tests[name][url]);
			});
		});
	});
});
