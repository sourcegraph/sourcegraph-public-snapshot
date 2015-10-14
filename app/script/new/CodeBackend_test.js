import sandbox from "../testSandbox";
import expect from "expect.js";

import CodeBackend from "./CodeBackend";
import * as CodeActions from "./CodeActions";

describe("CodeBackend", () => {
	it("should handle WantFile", () => {
		CodeBackend.xhr = function(options, callback) {
			expect(options.uri).to.be("/ui/aRepo@aRev/.tree/aTree");
			callback(null, null, "someFile");
		};
		CodeBackend.handle(new CodeActions.WantFile("aRepo", "aRev", "aTree"));
		expect(sandbox.dispatched).to.eql([new CodeActions.FileFetched("aRepo", "aRev", "aTree", "someFile")]);
	});
});
