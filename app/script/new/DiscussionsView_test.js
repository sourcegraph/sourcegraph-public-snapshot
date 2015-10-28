import autotest from "./util/autotest";

import React from "react";

import DiscussionsView from "./DiscussionsView";

import testdataNoDiscussions from "./testdata/DiscussionsView-noDiscussions.json";
import testdataDiscussions from "./testdata/DiscussionsView-discussions.json";

describe("DiscussionsView", () => {
	it("should render no discussions", () => {
		autotest(testdataNoDiscussions, `${__dirname}/testdata/DiscussionsView-noDiscussions.json`,
			<DiscussionsView defURL="/someURL" discussions={[]} />
		);
	});

	it("should render discussions", () => {
		autotest(testdataDiscussions, `${__dirname}/testdata/DiscussionsView-discussions.json`,
			<DiscussionsView defURL="/someURL" discussions={[{ID: 42, Title: "foo", Description: "bar", Comments: [1, 2, 3]}]} />
		);
	});
});
