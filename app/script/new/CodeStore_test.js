var expect = require("expect.js");

var CodeStore = require("./CodeStore");
var CodeActions = require("./CodeActions");

describe("CodeStore", () => {
	it("should handle FileFetched", () => {
		CodeStore.handle(new CodeActions.FileFetched("aRepo", "aRev", "aTree", "someContent"));
		expect(CodeStore.files.get("aRepo", "aRev", "aTree")).to.be("someContent");
	});
});
