import autotest from "sourcegraph/util/autotest";

import React from "react";

import Commit from "sourcegraph/vcs/Commit";

import testdataInitial from "sourcegraph/vcs/testdata/Commit-initial.json";
import testdataAvailable from "sourcegraph/vcs/testdata/Commit-available.json";
import testdataNoAuthorPerson from "sourcegraph/vcs/testdata/Commit-noAuthorPerson.json";

const sampleCommit = {
	ID: "abc",
	Message: "msg",
	Author: {Date: ""},
	AuthorPerson: {AvatarURL: "http://example.com/avatar.png"},
};

const sampleRepo = "sourcegraph.com/sourcegraph";

describe("Commit", () => {
	it("should initially render empty", () => {
		autotest(testdataInitial, `${__dirname}/testdata/Commit-initial.json`,
			<Commit commit={sampleCommit} repo={sampleRepo} full={false} />
		);
	});

	it("should render commit", () => {
		autotest(testdataAvailable, `${__dirname}/testdata/Commit-available.json`,
			<Commit commit={sampleCommit} repo={sampleRepo} full={false} />
		);
	});

	it("should render commit without author information", () => {
		autotest(testdataNoAuthorPerson, `${__dirname}/testdata/Commit-noAuthorPerson.json`,
			<Commit commit={{ID: "abc", Message: "msg", Author: {Date: ""}, AuthorPerson: null}} repo={sampleRepo} full={false} />
		);
	});
});
