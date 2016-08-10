// tslint:disable: typedef ordered-imports curly

import * as React from "react";
import {withDef} from "sourcegraph/def/withDef";
import {render} from "sourcegraph/util/renderTestUtils";
import {DefStore} from "sourcegraph/def/DefStore";
import * as DefActions from "sourcegraph/def/DefActions";
import {rel as relPath} from "sourcegraph/app/routePatterns";

const C = withDef((props) => null);

const props = {
	params: {splat: [null, "d"]},
	route: {path: relPath.def},
};

describe("withDef", () => {
	it("should render initially", () => {
		render(<C repo="r" rev="v" {...props} />);
	});

	it("should render if the def and rev exist", () => {
		DefStore.directDispatch(new DefActions.DefFetched("r", "v", "d", {CommitID: "c"} as any));
		render(<C repo="r" rev="v" {...props} />);
	});

	it("should render if the def does not exist", () => {
		DefStore.directDispatch(new DefActions.DefFetched("r", "v", "d", {Error: true} as any));
		render(<C repo="r" rev="v" {...props} />);
	});
});
