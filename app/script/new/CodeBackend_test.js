var sandbox = require("../testSandbox");
var expect = require("expect.js");

var CodeBackend = require("./CodeBackend");
var CodeActions = require("./CodeActions");

describe("CodeBackend", () => {
	it("should handle WantFile", () => {
		CodeBackend.xhr = function(options, callback) {
			expect(options.uri).to.be("/ui/aRepo@aRev/.tree/aTree");
			callback(null, null, "someFile");
		};
		CodeBackend.handle(new CodeActions.WantFile("aRepo", "aRev", "aTree"));
		expect(sandbox.dispatched).to.eql([new CodeActions.SetFile("aRepo", "aRev", "aTree", "someFile")]);
	});
});
