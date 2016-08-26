import * as React from "react";
import testdataFiles from "sourcegraph/tree/testdata/TreeList-files.json";
import testdataNotFound from "sourcegraph/tree/testdata/TreeList-notFound.json";
import * as TreeActions from "sourcegraph/tree/TreeActions";
import {TreeList} from "sourcegraph/tree/TreeList";
import {TreeStore} from "sourcegraph/tree/TreeStore";
import {autotest} from "sourcegraph/util/autotest";

describe("TreeList", () => {
	it("should render files", () => {
		TreeStore.directDispatch(new TreeActions.FileListFetched("repo", "c", {Files: ["p1/p2/f3", "p1/f2"]}));
		autotest(testdataFiles, "sourcegraph/tree/testdata/TreeList-files.json",
			<TreeList repo="repo" rev="rev" commitID="c" path="p1/p2" location={{query: {q: ""}}} />,
			{router: {}, status: {}, user: {}},
		);
	});

	it("should display 404 for not found directory", () => {
		TreeStore.directDispatch(new TreeActions.FileListFetched("repo", "c", {Files: ["p1/p2/f3", "p1/f2"]}));
		autotest(testdataNotFound, "sourcegraph/tree/testdata/TreeList-notFound.json",
			<TreeList repo="repo" rev="rev" commitID="c" path="p1/notfound" location={{query: {q: ""}}} />,
			{router: {}, status: {}, user: {}},
		);
	});
});
