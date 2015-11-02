import autotest from "./util/autotest";

import React from "react";

import DiscussionView from "./DiscussionView";

import testdataRender from "./testdata/DiscussionView-render.json";

describe("DiscussionView", () => {
	it("should render discussion", () => {
		autotest(testdataRender, `${__dirname}/testdata/DiscussionView-render.json`,
			<DiscussionView
				discussion={{ID: 42, Title: "foo", Description: "bar", Author: {Login: "me"}, Comments: [
					{ID: 1, Author: {Login: "you"}, Body: "comment"},
				]}}
				def={{QualifiedName: {__html: "someName"}}} />
		);
	});
});
