import autotest from "./util/autotest";

import React from "react";

import DiscussionsList from "./DiscussionsList";

import testdataSmall from "./testdata/DiscussionsList-small.json";
import testdataLarge from "./testdata/DiscussionsList-large.json";

describe("DiscussionsList", () => {
	it("should render small list", () => {
		autotest(testdataSmall, `${__dirname}/testdata/DiscussionsList-small.json`,
			<DiscussionsList
				defURL="/someURL"
				discussions={[{ID: 42, Title: "foo", Description: "bar", Author: {Login: "me"}, Comments: [1, 2, 3]}]}
				small={true} />
		);
	});

	it("should render large list", () => {
		autotest(testdataLarge, `${__dirname}/testdata/DiscussionsList-large.json`,
			<DiscussionsList
				defURL="/someURL"
				discussions={[{ID: 42, Title: "foo", Description: "bar", Author: {Login: "me"}, Comments: [1, 2, 3]}]} />
		);
	});
});
