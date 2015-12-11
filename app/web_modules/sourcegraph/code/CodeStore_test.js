import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import CodeStore from "sourcegraph/code/CodeStore";
import * as CodeActions from "sourcegraph/code/CodeActions";

describe("CodeStore", () => {
	it("should handle FileFetched", () => {
		Dispatcher.directDispatch(CodeStore, new CodeActions.FileFetched("aRepo", "aRev", "aTree", "someContent"));
		expect(CodeStore.files.get("aRepo", "aRev", "aTree")).to.be("someContent");
	});
});
