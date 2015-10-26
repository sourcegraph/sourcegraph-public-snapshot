import autotest from "./util/autotest";

import React from "react";

import DefPopup from "./DefPopup";

import testdataData from "./testdata/DefPopup-data.json";

describe("DefPopup", () => {
	it("should render definition data", () => {
		autotest(testdataData, `${__dirname}/testdata/DefPopup-data.json`,
			<DefPopup def={{URL: "someURL", QualifiedName: "someName", Data: {DocHTML: "someDoc"}}} examples={{test: "examples"}} highlightedDef="otherURL" />
		);
	});
});
