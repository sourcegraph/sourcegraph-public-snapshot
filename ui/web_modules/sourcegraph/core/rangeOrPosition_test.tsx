import expect from "expect.js";
import { RangeOrPosition } from "sourcegraph/core/rangeOrPosition";

describe("RangeOrPosition", () => {
	describe("parse and toString", () => {
		[
			["1", "1"],
			["1:2", "1:2"],
			["1-1", "1"],
			["1-2", "1-2"],
			["1:2-3", "1-3"],
			["1:2-1", "1"],
			["1-2:3", "1-2"],
			["1:2-3:4", "1:2-3:4"],
			["1:2-1:2", "1:2"],
		].forEach(test => {
			it(JSON.stringify(test), () => {
				const r = RangeOrPosition.parse(test[0]);
				expect(r && r.toString()).to.eql(test[1]);
			});
		});
	});

	describe("fromZeroIndexed", () => {
		[
			{input: {startLine: 1}, want: {startLine: 1}},
			{input: {startLine: 1, startCol: 2}, want: "same"},
			{input: {startLine: 1, endLine: 1}, want: {startLine: 1}},
			{input: {startLine: 1, endLine: 2}, want: "same"},
			{input: {startLine: 1, startCol: 2, endLine: 3}, want: {startLine: 1, endLine: 3}},
			{input: {startLine: 1, startCol: 2}, want: "same"},
			{input: {startLine: 1, endLine: 2, endCol: 3}, want: {startLine: 1, endLine: 2}},
			{input: {startLine: 1, startCol: 2, endLine: 3, endCol: 4}, want: "same"},
			{input: {startLine: 1, startCol: 2}, want: "same"},
		].forEach((test: any) => {
			it(JSON.stringify(test.input), () => {
				const r = RangeOrPosition.fromZeroIndexed(test.input.startLine, test.input.startCol, test.input.endLine, test.input.endCol);
				expect(r && r.zeroIndexed()).to.eql(test.want === "same" ? test.input : test.want);
			});
		});
	});
});
