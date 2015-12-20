import autotest from "sourcegraph/util/autotest";

import React from "react";

import ReactionList from "sourcegraph/reactions/ReactionList";

import testdataNoReactions from "sourcegraph/reactions/testdata/ReactionList-noReactions.json";
import testdataReactions from "sourcegraph/reactions/testdata/ReactionList-reactions.json";

describe("ReactionList", () => {
	it("should not render without reactions", () => {
		autotest(testdataNoReactions, `${__dirname}/testdata/ReactionsList-noReactions.json`,
			<ReactionList reactions={[]} onSelect={() => null} currentUser={{Login: "aUser"}}/>
		);
	});

	it("should render reactions", () => {
		autotest(testdataReactions, `${__dirname}/testdata/ReactionList-reactions.json`,
			<ReactionList reactions={[
				{Reaction: "+1", Users: [{Login: "user1"}]},
				{Reaction: "rocket", Users: [{Login: "user2"}]},
				{Reaction: "whale", Users: [{Login: "user1"}, {Login: "user2"}]},
			]} onSelect={() => null} />
		);
	});
});
