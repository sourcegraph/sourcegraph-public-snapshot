// tslint:disable: typedef ordered-imports

import * as React from "react";
import {RefsContainer} from "sourcegraph/def/RefsContainer";
import {render} from "sourcegraph/util/renderTestUtils";

const context = {
	eventLogger: {logEvent: () => null},
	user: null,
};

describe("RefsContainer", () => {
	it("should render initially", () => {
		render(<RefsContainer repoRefs={{Repo: "github.com/gorilla/mux", Files: []}} />, context);
	});

	it("should render if the refs failed ", () => {
		render(<RefsContainer repoRefs={{Repo: "github.com/gorilla/mux", Files: []}} defObj={{File: "foo.go"}} refs={{Error: true}} />, context);
	});

	it("should render if the def and refs loaded", () => {
		render(<RefsContainer repoRefs={{Repo: "github.com/gorilla/mux", Files: []}} defObj={{}} refs={[{Repo: "repo", CommitID: "commit"}]} />, context);
	});
});
