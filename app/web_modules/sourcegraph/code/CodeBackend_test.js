import expect from "expect.js";

import Dispatcher from "sourcegraph/Dispatcher";
import CodeBackend from "sourcegraph/code/CodeBackend";
import * as CodeActions from "sourcegraph/code/CodeActions";

describe("CodeBackend", () => {
	it("should handle WantFile", () => {
		CodeBackend.xhr = function(options, callback) {
			expect(options.uri).to.be("/.ui/aRepo@aRev/.tree/aTree");
			callback(null, null, "someFile");
		};
		expect(Dispatcher.catchDispatched(() => {
			Dispatcher.directDispatch(CodeBackend, new CodeActions.WantFile("aRepo", "aRev", "aTree"));
		})).to.eql([new CodeActions.FileFetched("aRepo", "aRev", "aTree", "someFile")]);
	});
});
