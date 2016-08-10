// tslint:disable: typedef ordered-imports curly

import expect from "expect.js";
import * as utf8 from "utf8";

import {annotate, sortAnns} from "sourcegraph/blob/Annotations";

const testCases = {
	"empty and unannotated": {
		input: "",
		anns: [],
		want: "",
	},
	"unannotated": {
		input: "a⌘b",
		anns: [],
		want: "a⌘b",
	},
	"zero-length annotations": {
		input: "aaaa",
		anns: [
			{StartByte: 0, EndByte: 0, Left: "<b>", Right: "</b>"},
			{StartByte: 0, EndByte: 0, Left: "<i>", Right: "</i>"},
			{StartByte: 2, EndByte: 2, Left: "<x>", Right: "</x>"},
		],
		want: "<b></b><i></i>aa<x></x>aa",
	},
	"1 annotation": {
		input: "a",
		anns: [{StartByte: 0, EndByte: 1, Left: "[", Right: "]"}],
		want: "[a]",
	},
	"2 annotations": {
		input: "a b",
		anns: [
			{StartByte: 0, EndByte: 1, Left: "[", Right: "]"},
			{StartByte: 2, EndByte: 3, Left: "<", Right: ">"},
		],
		want: "[a] <b>",
	},
	"nested": {
		input: "abc",
		anns: [
			{StartByte: 0, EndByte: 3, Left: "[", Right: "]"},
			{StartByte: 1, EndByte: 2, Left: "<", Right: ">"},
		],
		want: "[a<b>c]",
	},
	"nested 1": {
		input: "abcd",
		anns: [
			{StartByte: 0, EndByte: 4, Left: "<1>", Right: "</1>"},
			{StartByte: 1, EndByte: 3, Left: "<2>", Right: "</2>"},
			{StartByte: 2, EndByte: 2, Left: "<3>", Right: "</3>"},
		],
		want: "<1>a<2>b<3></3>c</2>d</1>",
	},
	"same range": {
		input: "ab",
		anns: [
			{StartByte: 0, EndByte: 2, Left: "[", Right: "]"},
			{StartByte: 0, EndByte: 2, Left: "<", Right: ">"},
		],
		want: "[<ab>]",
	},
	"same range (with WantInner)": {
		input: "ab",
		anns: [
			{StartByte: 0, EndByte: 2, Left: "[", Right: "]", WantInner: 1},
			{StartByte: 0, EndByte: 2, Left: "<", Right: ">", WantInner: 0},
		],
		want: "<[ab]>",
	},
	"remainder": {
		input: "xyz",
		anns: [
			{StartByte: 0, EndByte: 2, Left: "<b>", Right: "</b>"},
			{StartByte: 0, EndByte: 1, Left: "<c>", Right: "</c>"},
		],
		want: "<b><c>x</c>y</b>z",
	},
	"overlap simple": {
		input: "abc",
		anns: [
			{StartByte: 0, EndByte: 2, Left: "<X>", Right: "</X>"},
			{StartByte: 1, EndByte: 3, Left: "<Y>", Right: "</Y>"},
		],
		// Without re-opening overlapped annotations, we'd get
		// "<X>a<Y>b</X>c</Y>".
		want: "<X>a<Y>b</Y></X><Y>c</Y>",
	},
	"overlap simple double": {
		input: "abc",
		anns: [
			{StartByte: 0, EndByte: 2, Left: "<X1>", Right: "</X1>"},
			{StartByte: 0, EndByte: 2, Left: "<X2>", Right: "</X2>"},
			{StartByte: 1, EndByte: 3, Left: "<Y1>", Right: "</Y1>"},
			{StartByte: 1, EndByte: 3, Left: "<Y2>", Right: "</Y2>"},
		],
		want: "<X1><X2>a<Y1><Y2>b</Y2></Y1></X2></X1><Y1><Y2>c</Y2></Y1>",
	},
	"overlap triple complex": {
		input: "abcd",
		anns: [
			{StartByte: 0, EndByte: 2, Left: "<X>", Right: "</X>"},
			{StartByte: 1, EndByte: 3, Left: "<Y>", Right: "</Y>"},
			{StartByte: 2, EndByte: 4, Left: "<Z>", Right: "</Z>"},
		],
		want: "<X>a<Y>b</Y></X><Y><Z>c</Z></Y><Z>d</Z>",
	},
	"overlap same start": {
		input: "abcd",
		anns: [
			{StartByte: 0, EndByte: 2, Left: "<X>", Right: "</X>"},
			{StartByte: 0, EndByte: 3, Left: "<Y>", Right: "</Y>"},
			{StartByte: 1, EndByte: 4, Left: "<Z>", Right: "</Z>"},
		],
		want: "<Y><X>a<Z>b</Z></X><Z>c</Z></Y><Z>d</Z>",
	},
	"overlap (infinite loop regression #1)": {
		input: "abcde",
		anns: [
			{StartByte: 0, EndByte: 3, Left: "<X>", Right: "</X>"},
			{StartByte: 1, EndByte: 5, Left: "<Y>", Right: "</Y>"},
			{StartByte: 1, EndByte: 2, Left: "<Z>", Right: "</Z>"},
		],
		want: "<X>a<Y><Z>b</Z>c</Y></X><Y>de</Y>",
	},
	"start oob": {
		input: "a",
		anns: [
			{StartByte: -1, EndByte: 1, Left: "<", Right: ">"},
		],
		want: "<a>",
	},
	"start oob (multiple)": {
		input: "a",
		anns: [
			{StartByte: -3, EndByte: 1, Left: "1", Right: ""},
			{StartByte: -3, EndByte: 1, Left: "2", Right: ""},
			{StartByte: -1, EndByte: 1, Left: "3", Right: ""},
		],
		want: "123a",
	},
	"end oob": {
		input: "a",
		anns: [{StartByte: 0, EndByte: 3, Left: "<", Right: ">"}],
		want: "<a>",
	},
	"end oob overlapping": {
		input: "aaa",
		anns: [
			{StartByte: 0, EndByte: 3, Left: "<", Right: ">"},
			{StartByte: 2, EndByte: 10, Left: "{", Right: "}"},
		],
		want: "<aa{a}>",
	},
	"end oob overlapping 2": {
		input: "aaaaaa",
		anns: [
			{StartByte: -5, EndByte: 2, Left: "{", Right: "}"},
			{StartByte: 0, EndByte: 3, Left: "<", Right: ">"},
		],
		want: "{<aa>}<a>aaa",
	},
	"end oob (multiple)": {
		input: "ab",
		anns: [
			{StartByte: 0, EndByte: 3, Left: "", Right: "1"},
			{StartByte: 1, EndByte: 3, Left: "", Right: "2"},
			{StartByte: 0, EndByte: 5, Left: "", Right: "3"},
		],
		want: "ab213",
	},
};

function render(ann, content) {
	return ann.Left + (typeof content === "undefined" ? "" : content.join("")) + ann.Right;
}

describe("Annotations.annotate", () => {
	Object.keys(testCases).forEach((key) => {
		it(key, () => {
			const test = testCases[key];
			const anns = sortAnns(test.anns);
			const out = utf8.decode(annotate(test.input, 0, anns, render).join(""));
			expect(out).to.eql(test.want);
		});
	});
});
