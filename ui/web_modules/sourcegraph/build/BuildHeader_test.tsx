// tslint:disable: typedef ordered-imports

import {autotest} from "sourcegraph/util/autotest";

import * as React from "react";

import {BuildHeader} from "sourcegraph/build/BuildHeader";

import testdataInitial from "sourcegraph/build/testdata/BuildHeader-initial.json";

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
