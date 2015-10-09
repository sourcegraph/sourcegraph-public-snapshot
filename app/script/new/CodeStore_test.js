var expect = require("expect.js");

var CodeStore = require("./CodeStore");
var CodeActions = require("./CodeActions");

describe("CodeStore", () => {
	it("should handle SetFile", () => {
		CodeStore.handle(new CodeActions.SetFile("aRepo", "aRev", "aTree", "someContent"));
		expect(CodeStore.files.get("aRepo", "aRev", "aTree")).to.be("someContent");
	});
});
