var sandbox = require("../testSandbox");
var expect = require("expect.js");

var $ = require("jquery");
var React = require("react");
var ReactDOM = require("react-dom");
var TestUtils = require("react-addons-test-utils");
var CodeTokenView = require("./CodeTokenView");
var CodeTokenModel = require("../stores/models/CodeTokenModel");
var globals = require("../globals");

describe("components/CodeTokenView", () => {
	it("should correct apply classes for token type REF", () => {
		var model = new CodeTokenModel({
			type: globals.TokenType.REF,
			syntax: "pln",
			html: "abc",
		});

		var component = sandbox.renderComponent(<CodeTokenView model={model} />);
		var a = TestUtils.findRenderedDOMComponentWithTag(component, "a");
		var span = TestUtils.findRenderedDOMComponentWithTag(component, "span");

		expect($(ReactDOM.findDOMNode(a)).hasClass("ref")).to.be(true);
		expect($(ReactDOM.findDOMNode(span)).hasClass("pln")).to.be(true);
		expect($(ReactDOM.findDOMNode(span)).html()).to.be("abc");
	});

	it("should correctly apply classes for token type DEF", () => {
		var model = new CodeTokenModel({
			type: globals.TokenType.DEF,
			syntax: "pln",
			html: "abc",
		});

		var component = sandbox.renderComponent(<CodeTokenView model={model} />);
		var a = TestUtils.findRenderedDOMComponentWithTag(component, "a");
		var span = TestUtils.findRenderedDOMComponentWithTag(component, "span");

		expect($(ReactDOM.findDOMNode(a)).hasClass("ref")).to.be(true);
		expect($(ReactDOM.findDOMNode(a)).hasClass("def")).to.be(true);
		expect($(ReactDOM.findDOMNode(span)).hasClass("pln")).to.be(true);
		expect($(ReactDOM.findDOMNode(span)).html()).to.be("abc");
	});

	it("should correctly apply classes for highlighted tokens", () => {
		var model = new CodeTokenModel({
			type: globals.TokenType.DEF,
			syntax: "pln",
			html: "abc",
			highlighted: true,
		});

		var component = sandbox.renderComponent(<CodeTokenView model={model} />);
		var a = TestUtils.findRenderedDOMComponentWithTag(component, "a");

		expect($(ReactDOM.findDOMNode(a)).hasClass("highlight-secondary")).to.be(true);
	});

	it("should correctly apply classes for selected tokens", () => {
		var model = new CodeTokenModel({
			type: globals.TokenType.DEF,
			syntax: "pln",
			html: "abc",
			selected: true,
		});

		var component = sandbox.renderComponent(<CodeTokenView model={model} />);
		var a = TestUtils.findRenderedDOMComponentWithTag(component, "a");

		expect($(ReactDOM.findDOMNode(a)).hasClass("highlight-primary")).to.be(true);
	});

	it("the selected property should override the highlighted property", () => {
		var model = new CodeTokenModel({
			type: globals.TokenType.DEF,
			syntax: "pln",
			html: "abc",
			selected: true,
			highlighted: true,
		});

		var component = sandbox.renderComponent(<CodeTokenView model={model} />);
		var a = TestUtils.findRenderedDOMComponentWithTag(component, "a");

		expect($(ReactDOM.findDOMNode(a)).hasClass("highlight-primary")).to.be(true);
		expect($(ReactDOM.findDOMNode(a)).hasClass("highlight-secondary")).to.be(false);
	});

	it("should call the set function for click events", () => {
		var clickFn = sandbox.spy();
		var model = new CodeTokenModel();
		var component = sandbox.renderComponent(
			<CodeTokenView model={model} onTokenClick={clickFn} />
		);
		var a = TestUtils.findRenderedDOMComponentWithTag(component, "a");

		TestUtils.Simulate.click(a);
		expect(clickFn.callCount).to.be(1);
		expect(clickFn.firstCall.args[0]).to.be(model);
	});

	it("should not call the set function for click events if component is loading", () => {
		var clickFn = sandbox.spy();
		var model = new CodeTokenModel();
		var component = sandbox.renderComponent(
			<CodeTokenView model={model} onTokenClick={clickFn} loading={true} />
		);
		var a = TestUtils.findRenderedDOMComponentWithTag(component, "a");

		TestUtils.Simulate.click(a);
		expect(clickFn.called).to.be(false);
	});

	// TODO(gbbr): Find a way to simulate mouseenter/mouseleave
});
