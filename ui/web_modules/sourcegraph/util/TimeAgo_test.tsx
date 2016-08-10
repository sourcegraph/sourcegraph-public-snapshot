// tslint:disable: typedef ordered-imports

import expect from "expect.js";
import {formatDuration} from "sourcegraph/util/TimeAgo";

describe("TimeAgo.formatDuration", () => {
	const tests = [
		{sec: 0, want: "0s"},
		{sec: 1, want: "1s"},
		{sec: 30, want: "30s"},
		{sec: 60, want: "1m"},
		{sec: 60 * 2 + 30, want: "2m 30s"},
		{sec: 60 * 60, want: "1h"},
		{sec: 60 * 60 + 30, want: "1h"},
		{sec: 60 * 60 * 24, want: "1d"},
		{sec: 60 * 60 * 24 + 60 * 30, want: "1d"},
		{sec: 60 * 60 * 24 * 30, want: "1mth"},
		{sec: 60 * 60 * 24 * 30 + 60 * 60 * 12, want: "1mth"},
		{sec: 60 * 60 * 24 * 30 * 12, want: "1yr"},
		{sec: 60 * 60 * 24 * 30 * 12 + 60 * 60 * 24 * 30 * 3, want: "1yr 3mth"},
	];
	tests.forEach((test) => {
		it(`formats ${test.sec}sec`, () => {
			expect(formatDuration(test.sec * 1000)).to.be(test.want);
		});
	});
});
