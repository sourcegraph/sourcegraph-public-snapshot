import autotest from "sourcegraph/util/autotest";

import React from "react";

import BuildHeader from "sourcegraph/build/BuildHeader";

import testdataInitial from "sourcegraph/build/testdata/BuildHeader-initial.json";

const sampleBuild = {
	ID: 123,
	CreatedAt: "",
};

const sampleCommit = {
	ID: "abc",
	Message: "msg",
	Author: {Date: ""},
	AuthorPerson: {AvatarURL: "http://example.com/avatar.png"},
};

describe("BuildHeader", () => {
	it("should render", () => {
		autotest(testdataInitial, `${__dirname}/testdata/BuildHeader-initial.json`,
			<BuildHeader
				build={sampleBuild}
				commit={sampleCommit} />
		);
	});
});
