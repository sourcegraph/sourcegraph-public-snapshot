import autotest from "sourcegraph/util/autotest";

import React from "react";

import TreeSearch from "sourcegraph/tree/TreeSearch";
import TreeStore from "sourcegraph/tree/TreeStore";
import * as TreeActions from "sourcegraph/tree/TreeActions";

import testdataFiles from "sourcegraph/tree/testdata/TreeSearch-files.json";

describe("TreeSearch", () => {
	it("should render files", () => {
		TreeStore.directDispatch(new TreeActions.FileListFetched("repo", "rev", "c", {Files: ["a", "b"]}));
		autotest(testdataFiles, `${__dirname}/testdata/TreeSearch-files.json`,
			<TreeSearch repo="repo" rev="rev" commitID="c" prefetch={true} currPath={[]} overlay={true} />
		);
	});
});
