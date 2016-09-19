import expect from "expect.js";
import * as React from "react";

import {ResultCategories} from "sourcegraph/search/modal/SearchComponent";
import {renderToString} from "sourcegraph/util/testutil/componentTestUtils";

const data = {
	categories: [{
		Title: "Category 1",
		Results: [{
			title: "Result 1",
			description: "this is result 1",
		}],
	}],
	selected: [0, 0],
	delegate: null,
};

describe("ResultCategories", () => {
	it("should show categories", () => {
		const o = renderToString(<ResultCategories categories={data.categories} selection={data.selected} delegate={data.delegate} />);
		let shouldContain = [
			"this is result 1",
			"Result 1",
			"floalsdkfj",
		];
		shouldContain.forEach((s: string) => {
			expect(o).to.contain(s);
		});
	});
});
