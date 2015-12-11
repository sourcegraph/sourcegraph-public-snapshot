import autotest from "./autotest";

import React from "react";

import Pagination from "./Pagination";

import testdataPageLinks from "./testdata/Pagination-pageLinks.json";
import testdataBounded from "./testdata/Pagination-bounded.json";
import testdataBoundedLastPage from "./testdata/Pagination-boundedLastPage.json";

describe("Pagination", () => {
	it("displays the correct number of page links", () => {
		autotest(testdataPageLinks, `${__dirname}/testdata/Pagination-pageLinks.json`,
			<Pagination currentPage={1} totalPages={10} pageRange={10} onPageChange={() => null}/>
		);
	});

	it("is bounded by the total number of page links", () => {
		autotest(testdataBounded, `${__dirname}/testdata/Pagination-bounded.json`,
			<Pagination currentPage={1} totalPages={5} pageRange={10} onPageChange={() => null}/>
		);
	});

	it("is bounded by the total number of page links on the last page", () => {
		autotest(testdataBoundedLastPage, `${__dirname}/testdata/Pagination-boundedLastPage.json`,
			<Pagination currentPage={42} totalPages={42} pageRange={10} onPageChange={() => null}/>
		);
	});
});
