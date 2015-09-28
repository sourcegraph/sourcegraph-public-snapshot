var sandbox = require("../testSandbox");
var expect = require("expect.js");

var $ = require("jquery");
var React = require("react/addons");
var TestUtils = React.addons.TestUtils;
var CodeView = require("./CodeView");
var CodeLineView = require("./CodeLineView");
var CodeModel = require("../stores/models/CodeModel");
var CodeLineModel = require("../stores/models/CodeLineModel");

describe("components/CodeView", () => {
	it("should show a loader while there are no lines to display", () => {
		var model = new CodeModel();
		var component = sandbox.renderComponent(
			<CodeView model={model} />
		);

		var tag = TestUtils.findRenderedDOMComponentWithTag(component, "i");
		expect($(tag.getDOMNode()).hasClass("file-loader")).to.be(true);

		var children = TestUtils.scryRenderedComponentsWithType(component, CodeLineView);
		expect(children.length).to.be(0);
	});

	it("should render rows for each line in CodeModel", () => {
		var lines = [
			new CodeLineModel(),
			new CodeLineModel(),
			new CodeLineModel(),
		];
		var model = new CodeModel({lines: lines});
		var component = sandbox.renderComponent(<CodeView model={model} />);

		TestUtils.findRenderedDOMComponentWithTag(component, "table");
		TestUtils.findRenderedDOMComponentWithTag(component, "tbody");

		var children = TestUtils.scryRenderedComponentsWithType(component, CodeLineView);
		expect(children.length).to.be(3);
		expect(children[0].props.model).to.be(lines[0]);
		expect(children[1].props.model).to.be(lines[1]);
		expect(children[2].props.model).to.be(lines[2]);
	});
});
