import expect from "expect.js";

import CodeStore from "./CodeStore";
import * as CodeActions from "./CodeActions";

describe("CodeStore", () => {
	it("should handle FileFetched", () => {
		CodeStore.handle(new CodeActions.FileFetched("aRepo", "aRev", "aTree", "someContent"));
		expect(CodeStore.files.get("aRepo", "aRev", "aTree")).to.be("someContent");
	});
});
