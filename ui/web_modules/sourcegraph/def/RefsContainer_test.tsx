import * as React from "react";
import {RefsContainer} from "sourcegraph/def/RefsContainer";
import {render} from "sourcegraph/util/testutil/renderTestUtils";

const context = {
	eventLogger: {logEvent: () => null},
	user: null,
};

const location = {
			hash: "",
			key: "",
			pathname: "",
			search: "",
			action: "",
			query: {},
			state: {},
};

describe("RefsContainer", () => {
	it("should render initially", () => {
		render(<RefsContainer location={location} repoRefs={{Repo: "github.com/gorilla/mux", Files: []}} />, context);
	});

	it("should render if the refs failed ", () => {
		render(<RefsContainer location={location} repoRefs={{Repo: "github.com/gorilla/mux", Files: []}} defObj={{File: "foo.go"}} refs={{Error: true}} />, context);
	});

	it("should render if the def and refs loaded", () => {
		render(<RefsContainer location={location} repoRefs={{Repo: "github.com/gorilla/mux", Files: []}} defObj={{}} refs={[{Repo: "repo", CommitID: "commit"}]} />, context);
	});
});
