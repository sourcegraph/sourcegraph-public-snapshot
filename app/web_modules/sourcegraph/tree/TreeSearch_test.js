import autotest from "sourcegraph/util/autotest";

import React from "react";
import Dispatcher from "sourcegraph/Dispatcher";

import TreeSearch from "sourcegraph/tree/TreeSearch";
import TreeStore from "sourcegraph/tree/TreeStore";
import * as TreeActions from "sourcegraph/tree/TreeActions";

import testdataFiles from "sourcegraph/tree/testdata/TreeSearch-files.json";
import testdataLoading from "sourcegraph/tree/testdata/TreeSearch-loading.json";

describe("TreeSearch", () => {
	it("should show a loading indicator", () => {
		autotest(testdataLoading, `${__dirname}/testdata/TreeSearch-loading.json`,
			<TreeSearch repo="repo" rev="rev" commitID="c" prefetch={true} currPath={[]} overlay={true} />
		);
	});

	it("should render files", () => {
		Dispatcher.directDispatch(TreeStore, new TreeActions.FileListFetched("repo", "rev", "c", {Files: ["a", "b"]}));
		autotest(testdataFiles, `${__dirname}/testdata/TreeSearch-files.json`,
			<TreeSearch repo="repo" rev="rev" commitID="c" prefetch={true} currPath={[]} overlay={true} />
		);
	});
});
