import expect from "expect.js";

import {rankFile} from "sourcegraph/quickopen/Container";

describe("rankFile", () => {
	it("should add perfect matches", () => {
		const file = "foo";
		const query = "foo";
		const r = rankFile([], file, query);
		expect(r[0]).to.eql({title: "foo", score: 1, length: 3, index: 0, description: ""});
	});

	it("should weight prefix matches", () => {
		const file = "foo/bar";
		const query = "foo";
		const r = rankFile([], file, query);
		expect(r[0]).to.eql({title: "foo/bar", score: .8, length: 3, index: 0, description: ""});
	});

	it("should weight contains matches", () => {
		const file = "foo/bar/baz";
		const query = "bar";
		const r = rankFile([], file, query);
		expect(r[0]).to.eql({title: "foo/bar/baz", score: .6, length: 3, index: 4, description: ""});
	});
});
