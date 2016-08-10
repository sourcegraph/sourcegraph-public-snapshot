// tslint:disable: typedef ordered-imports

import expect from "expect.js";

import {lineFromByte} from "sourcegraph/blob/lineFromByte";
import {createLineFromByteFunc} from "sourcegraph/blob/lineFromByte";

const testCases = { // eslint-disable-line quote-props
	"ab": [
		{byte: 0, wantLine: 1},
		{byte: 1, wantLine: 1},
		{byte: 2, wantLine: 1},
	],
	"ab\n": [
		{byte: 0, wantLine: 1},
		{byte: 1, wantLine: 1},
		{byte: 2, wantLine: 1},
		{byte: 3, wantLine: 2},
	],
	"a\nb\n": [
		{byte: 0, wantLine: 1},
		{byte: 1, wantLine: 1},
		{byte: 2, wantLine: 2},
		{byte: 3, wantLine: 2},
		{byte: 4, wantLine: 3},
	],
};

Object.keys(testCases).forEach((contents) => {
	const tests = testCases[contents];

	describe(`createLineFromByteFunc ${JSON.stringify(contents)}`, () => {
		const fn = createLineFromByteFunc(contents);
		tests.forEach(({byte, wantLine}) => {
			it(`byte ${byte}`, () => {
				const got = fn(byte);
				expect(got).to.eql(wantLine);
			});
		});
	});

	describe(`lineFromByte ${JSON.stringify(contents)}`, () => {
		tests.forEach(({byte, wantLine}) => {
			it(`byte ${byte}`, () => {
				const got = lineFromByte(contents, byte);
				expect(got).to.eql(wantLine);
			});
		});
	});
});
