import autotest from "sourcegraph/util/autotest";

import React from "react";

import RefLocationsList from "sourcegraph/def/RefLocationsList";

import testdataData from "sourcegraph/def/testdata/RefLocationsList-data.json";
import testdataEmpty from "sourcegraph/def/testdata/RefLocationsList-empty.json";

describe("RefLocationsList", () => {
	it("should render definition data", () => {
		autotest(testdataData, `${__dirname}/testdata/RefLocationsList-data.json`,
			<RefLocationsList
				repo="r" rev="v" path="p"
				def={{Repo: "r", CommitID: "c"}}
				refLocations={[{Repo: "r", Files: [{Path: "f", Count: 2}]}]} />
		);
	});

	it("should render empty", () => {
		autotest(testdataEmpty, `${__dirname}/testdata/RefLocationsList-empty.json`,
			<RefLocationsList
				repo="r" rev="v" path="p"
				def={{Repo: "r", CommitID: "c"}}
				refLocations={[]} />
		);
	});
});
