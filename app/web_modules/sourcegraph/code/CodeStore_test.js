import expect from "expect.js";

import Dispatcher from "../Dispatcher";
import CodeStore from "./CodeStore";
import * as CodeActions from "./CodeActions";

describe("CodeStore", () => {
	it("should handle FileFetched", () => {
		Dispatcher.directDispatch(CodeStore, new CodeActions.FileFetched("aRepo", "aRev", "aTree", "someContent"));
		expect(CodeStore.files.get("aRepo", "aRev", "aTree")).to.be("someContent");
	});
});
