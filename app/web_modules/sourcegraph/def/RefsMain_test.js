// @flow weak

import React from "react";
import RefsMain from "sourcegraph/def/RefsMain";
import {render} from "sourcegraph/util/renderTestUtils";

describe("RefsMain", () => {
	it("should render initially", () => {
		render(<RefsMain />);
	});

	it("should render if the refs failed ", () => {
		render(<RefsMain defObj={{File: "foo.go"}} refs={{Error: true}} />);
	});

	it("should render if the def and refs loaded", () => {
		render(<RefsMain defObj={{}} refs={[]} />);
	});
});
