// tslint:disable: typedef ordered-imports

import expect from "expect.js";
import * as React from "react";

import {Category, SearchDelegate} from "sourcegraph/search/modal/SearchContainer";
import {ResultCategories} from "sourcegraph/search/modal/SearchComponent";
import {renderToString} from "sourcegraph/util/testutil/componentTestUtils";

interface Case {
	categories: Category[];
	selected: number[];
	delegate: SearchDelegate;
	shouldContain: string[]; // strings the rendered element is expected to contain
}

const testDelegate = {
	dismiss: function(): void {/* do nothing */},
	select: function(category: number, row: number): void {/* do nothing */},
};

describe("ResultCategories", () => {
	it("should show categories", () => {
		const cases: Case[] = [{
			categories: [{
				Title: "Category 1",
				Results: [{
					title: "Result 1",
					description: "this is result 1",
					URLPath: "",
				}],
				IsLoading: false,
			}],
			selected: [0, 0],
			delegate: testDelegate,
			shouldContain: ["this is result 1", "Result 1"],
		}, {
			categories: [{
				Title: "Category 1",
				Results: [{
					title: "Result 1",
					description: "this is result 1",
					URLPath: "",
				}],
				IsLoading: false,
			}, {
				Title: "Category 2",
				Results: [{
					title: "Result 2-a",
					description: "this is result 2-a",
					URLPath: "",
				}, {
					title: "Result 2-b",
					description: "this is result 2-b",
					URLPath: "",
				}],
				IsLoading: false,
			}, {
				Title: "Category 3",
				IsLoading: true,
			}],
			selected: [1, 1],
			delegate: testDelegate,
			shouldContain: [
				"this is result 1",
				"Result 1",
				"this is result 2-a",
				"this is result 2-b",
				"loading...",
			],
		}];
		cases.forEach((t) => {
			const o = renderToString(<ResultCategories categories={t.categories} selection={t.selected} delegate={t.delegate} />);
			t.shouldContain.forEach((s: string) => {
				expect(o).to.contain(s);
			});
		});
	});
});
