import * as React from "react";
import {BuildHeader} from "sourcegraph/build/BuildHeader";
import testdataInitial from "sourcegraph/build/testdata/BuildHeader-initial.json";
import {autotest} from "sourcegraph/util/testutil/autotest";

const sampleBuild = {
	ID: 123,
	CreatedAt: "",
};

describe("BuildHeader", () => {
	it("should render", () => {
		autotest(testdataInitial, "sourcegraph/build/testdata/BuildHeader-initial.json",
			<BuildHeader build={sampleBuild} />
		);
	});
});
