var sandbox = require("../testSandbox");

var React = require("react");
var CodeFileView = require("./CodeFileView");
var CodeFileRouter = require("../routing/CodeFileRouter");
var globals = require("../globals");

describe("components/CodeFileView", () => {
	it("should render without error", () => {
		globals.Features = {Discussions: true};
		sandbox.stub(CodeFileRouter, "start");
		sandbox.renderComponent(<CodeFileView />);
	});
});
