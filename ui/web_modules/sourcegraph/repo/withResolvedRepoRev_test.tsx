import * as React from "react";
import {Location} from "sourcegraph/Location";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import {RepoStore} from "sourcegraph/repo/RepoStore";
import {withResolvedRepoRev} from "sourcegraph/repo/withResolvedRepoRev";
import {render} from "sourcegraph/util/testutil/renderTestUtils";

const C = withResolvedRepoRev((props) => null);

describe("withResolvedRepoRev", () => {
	it("should render initially", () => {
		render(<C params={{splat: "r"}} location={{} as Location} />);
	});

	it("should render if the repo and rev exist", () => {
		RepoStore.directDispatch(new RepoActions.FetchedRepo("r", {DefaultBranch: "v"}));
		render(<C params={{splat: "r"}}  location={{} as Location}/>);
	});

	it("should render if the repo is cloning", () => {
		RepoStore.directDispatch(new RepoActions.FetchedRepo("r", {DefaultBranch: "v"}));
		RepoStore.directDispatch(new RepoActions.RepoCloning("r", true));
		render(<C params={{splat: "r"}}  location={{} as Location}/>);
	});

	it("should render if the repo does not exist", () => {
		RepoStore.directDispatch(new RepoActions.FetchedRepo("r", {Error: true} as any));
		render(<C params={{splat: "r"}}  location={{} as Location}/>);
	});
});
