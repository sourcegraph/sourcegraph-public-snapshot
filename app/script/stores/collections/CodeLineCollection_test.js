var expect = require("expect.js");

var CodeTokenCollection = require("./CodeLineCollection");

describe("stores/collections/CodeLineCollection", () => {
	function getLineCollection() {
		return new CodeTokenCollection([
			{number: 1, start: 0, end: 50, highlight: false},
			{number: 2, start: 51, end: 100, highlight: false},
			{number: 3, start: 101, end: 150, highlight: false},
			{number: 4, start: 151, end: 200, highlight: false},
		]);
	}

	it("should correctly set highlight by byte range", () => {
		[
			{start: 0, end: 45, range: [1]},
			{start: 0, end: 51, range: [1, 2]},
			{start: 0, end: 100, range: [1, 2]},
			{start: 60, end: 160, range: [2, 3, 4]},
			{start: 160, end: 168, range: [4]},
		].forEach(test => {
			var coll = getLineCollection();
			coll.highlightByteRange(test.start, test.end);
			test.range.forEach(i => expect(coll.get(i).get("highlight")).to.be(true));
		});
	});

	it("should correctly highlight by line range", () => {
		[
			{start: 1, end: 1, range: [1]},
			{start: 2, end: 4, range: [2, 3, 4]},
			{start: 3, end: 4, range: [3, 4]},
		].forEach(test => {
			var coll = getLineCollection();
			coll.highlightRange(test.start, test.end);
			test.range.forEach(i => expect(coll.get(i).get("highlight")).to.be(true));
		});
	});

	it("should correctly clear highlighted highlights", () => {
		var coll = getLineCollection();

		coll.highlightRange(2, 4);
		[2, 3, 4].forEach(i => expect(coll.get(i).get("highlight")).to.be(true));

		coll.clearHighlighted(2, 4);
		[2, 3, 4].forEach(i => expect(coll.get(i).get("highlight")).to.be(false));
	});
});
