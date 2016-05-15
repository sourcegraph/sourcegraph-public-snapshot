// @flow weak

import React from "react";
import RefsContainer from "sourcegraph/def/RefsContainer";
import {render} from "sourcegraph/util/renderTestUtils";

describe("RefsContainer", () => {
	it("should render initially", () => {
		render(<RefsContainer repoRefs={{Repo: "github.com/gorilla/mux", Files: []}} />);
	});

	it("should render if the refs failed ", () => {
		render(<RefsContainer repoRefs={{Repo: "github.com/gorilla/mux", Files: []}} defObj={{File: "foo.go"}} refs={{Error: true}} />);
	});

	it("should render if the def and refs loaded", () => {
		render(<RefsContainer repoRefs={{Repo: "github.com/gorilla/mux", Files: []}} defObj={{}} refs={[]} />);
	});
});
