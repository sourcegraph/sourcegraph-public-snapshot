var sandbox = require("../testSandbox");
var expect = require("expect.js");

var React = require("react");

var CodeFileController = require("./CodeFileController");
var CodeListing = require("./CodeListing");
var CodeActions = require("./CodeActions");
var CodeStore = require("./CodeStore");

describe("CodeFileController", () => {
	it("should handle unavailable file", () => {
		CodeStore.handle(new CodeActions.FileFetched("aRepo", "aRev", "aTree", undefined));
		sandbox.renderAndExpect(<CodeFileController repo="aRepo" rev="aRev" tree="aTree" />).to.eql(null);
		expect(sandbox.dispatched).to.eql([new CodeActions.WantFile("aRepo", "aRev", "aTree")]);
	});

	it("should handle available file", () => {
		CodeStore.handle(new CodeActions.FileFetched("aRepo", "aRev", "aTree", {Entry: {SourceCode: {Lines: ["someLine"]}}}));
		sandbox.renderAndExpect(<CodeFileController repo="aRepo" rev="aRev" tree="aTree" />).to.eql(
			<CodeListing lines={["someLine"]} />
		);
	});
});
