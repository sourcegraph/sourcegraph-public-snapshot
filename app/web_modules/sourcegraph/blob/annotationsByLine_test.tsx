// tslint:disable

import annotationsByLine from "sourcegraph/blob/annotationsByLine";
import expect from "expect.js";

function labels(annsByLine) {
	return annsByLine.map((lineAnns) => lineAnns.map((ann) => ann.Label));
}

// NOTE: This must be kept in sync with blob_test.go TestAnnotationsByLine.
describe("annotationsByLine", () => {
	it("empty", () => {
		expect(annotationsByLine([], [], [""])).to.eql([]);
		expect(annotationsByLine([], [], [])).to.eql([]);
	});
	it("one line", () => {
		expect(annotationsByLine([0], [], [""])).to.eql([[]]);
	});
	it("one line, one ann", () => {
		expect(labels(
			annotationsByLine([0], [{StartByte: 0, EndByte: 2, Label: "a"}], ["aaa"])
		)).to.eql([["a"]]);
	});
	it("multiple lines, no cross-line span", () => {
		expect(labels(
			annotationsByLine([0, 4], [
				{StartByte: 0, EndByte: 2, Label: "a"},
				{StartByte: 2, EndByte: 3, Label: "b"},
				{StartByte: 4, EndByte: 6, Label: "c"},
				{StartByte: 5, EndByte: 7, Label: "d"},
			], ["aaa", "aaa"])
		)).to.eql([["a", "b"], ["c", "d"]]);
	});
	it("multiple lines, empty line, no cross-line span", () => {
		expect(labels(
			annotationsByLine([0, 4, 8], [
				{StartByte: 0, EndByte: 2, Label: "a"},
				{StartByte: 2, EndByte: 3, Label: "b"},
				{StartByte: 8, EndByte: 10, Label: "c"},
				{StartByte: 9, EndByte: 11, Label: "d"},
			], ["aaa", "aaa", "aaa"])
		)).to.eql([["a", "b"], [], ["c", "d"]]);
	});
	it("cross-line span", () => {
		expect(labels(
			annotationsByLine([0, 4], [
				{StartByte: 0, EndByte: 3, Label: "a"},
				{StartByte: 1, EndByte: 7, Label: "b"},
				{StartByte: 2, EndByte: 5, Label: "c"},
				{StartByte: 4, EndByte: 5, Label: "d"},
			], ["aaa", "aaa"])
		)).to.eql([["a", "b", "c"], ["b", "c", "d"]]);
	});
	it("cross-line span complex", () => {
		expect(labels(
			annotationsByLine([0, 4, 8], [
				{StartByte: 0, EndByte: 11, Label: "a"},
				{StartByte: 1, EndByte: 5, Label: "b"},
				{StartByte: 4, EndByte: 5, Label: "c"},
				{StartByte: 4, EndByte: 10, Label: "d"},
			], ["aaa", "aaa", "aaa"])
		)).to.eql([["a", "b"], ["a", "b", "c", "d"], ["a", "d"]]);
	});
});
