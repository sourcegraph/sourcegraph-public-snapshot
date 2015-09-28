var sandbox = require("../testSandbox");
var expect = require("expect.js");

var $ = require("jquery");
var React = require("react/addons");
var TestUtils = React.addons.TestUtils;
var CodeLineView = require("./CodeLineView");
var CodeTokenView = require("./CodeTokenView");
var CodeLineModel = require("../stores/models/CodeLineModel");
var CodeTokenModel = require("../stores/models/CodeTokenModel");
var CodeFileActions = require("../actions/CodeFileActions");
var globals = require("../globals");

describe("components/CodeLineView", () => {
	it("should register its node with the model", () => {
		var model = new CodeLineModel();

		sandbox.renderComponent(<table><tbody><CodeLineView model={model} /></tbody></table>);
		expect(model.__node).not.to.be(null);
	});

	it("should register line numbers by default", () => {
		var model = new CodeLineModel({number: 5});
		var component = sandbox.renderComponent(<table><tbody><CodeLineView model={model} /></tbody></table>);

		var el = TestUtils.findRenderedDOMComponentWithClass(component, "line-number");
		expect($(el.getDOMNode()).data("line")).to.be(5);
	});

	it("should not register line numbers if specifically disabled via props", () => {
		var model = new CodeLineModel();
		var component = sandbox.renderComponent(
			<table><tbody><CodeLineView lineNumbers={false} model={model} /></tbody></table>
		);

		var td = TestUtils.scryRenderedDOMComponentsWithClass(component, "line-number");
		expect(td.length).to.be(0);
	});

	it("should render 1 whitespace if the line has no tokens", () => {
		var model = new CodeLineModel();
		var component = sandbox.renderComponent(<table><tbody><CodeLineView model={model} /></tbody></table>);

		var td = TestUtils.findRenderedDOMComponentWithClass(component, "line-content");
		expect(td.getDOMNode().textContent).to.be(" ");
	});

	it("should render plain text (STRING) tokens correctly", () => {
		var token = new CodeTokenModel({
			html: "abc",
			type: globals.TokenType.STRING,
		});

		var model = new CodeLineModel({tokens: [token]});
		var component = sandbox.renderComponent(<table><tbody><CodeLineView model={model} /></tbody></table>);

		var td = TestUtils.findRenderedDOMComponentWithClass(component, "line-content");

		expect($(td.getDOMNode()).children().length).to.be(1);
		expect($(td.getDOMNode()).children("span")[0].innerHTML).to.be("abc");
	});

	it("should render code highlighted tokens (SPAN) correctly", () => {
		var token = new CodeTokenModel({
			html: "abc",
			cid: 1,
			type: globals.TokenType.SPAN,
			syntax: "pln",
		});

		var model = new CodeLineModel({tokens: [token]});
		var component = sandbox.renderComponent(<table><tbody><CodeLineView model={model} /></tbody></table>);
		var tok = TestUtils.findRenderedDOMComponentWithClass(component, "pln");

		expect($(tok.getDOMNode()).html()).to.be("abc");
	});

	it("should rendered token component for everything else and pass down parent props (assumed linked, ie: REF & DEF)", () => {
		var ref = new CodeTokenModel({type: globals.TokenType.REF});
		var def = new CodeTokenModel({type: globals.TokenType.DEF});

		var model = new CodeLineModel({tokens: [ref, def]});
		var component = sandbox.renderComponent(<table><tbody><CodeLineView someprop={1} model={model} /></tbody></table>);

		var children = TestUtils.scryRenderedComponentsWithType(component, CodeTokenView);
		expect(children.length).to.be(2);

		expect(children[0].props.model).to.be(ref);
		expect(children[1].props.model).to.be(def);

		expect(children[0].props.someprop).to.be(1);
		expect(children[1].props.someprop).to.be(1);
	});

	it("should apply main-byte-range class when line model is highlighted", () => {
		var model = new CodeLineModel({highlight: true});
		var component = sandbox.renderComponent(<table><tbody><CodeLineView model={model} /></tbody></table>);

		var tok = TestUtils.findRenderedDOMComponentWithTag(component, "tr");

		expect($(tok.getDOMNode()).hasClass("main-byte-range")).to.be(true);
		expect($(tok.getDOMNode()).hasClass("line")).to.be(true);
	});

	it("should not apply main-byte-range class when line model is not highlighted", () => {
		var model = new CodeLineModel();
		var component = sandbox.renderComponent(<table><tbody><CodeLineView model={model} /></tbody></table>);

		var tok = TestUtils.findRenderedDOMComponentWithTag(component, "tr");

		expect($(tok.getDOMNode()).hasClass("main-byte-range")).not.to.be(true);
		expect($(tok.getDOMNode()).hasClass("line")).to.be(true);
	});

	it("should trigger CodeFileActions.selectLines when line number is clicked", () => {
		var model = new CodeLineModel();
		var component = sandbox.renderComponent(<table><tbody><CodeLineView model={model} /></tbody></table>);
		var no = TestUtils.findRenderedDOMComponentWithClass(component, "line-number");

		sandbox.spy(CodeFileActions, "selectLines");
		TestUtils.Simulate.click(no);
		expect(CodeFileActions.selectLines.callCount).to.be(1);
	});
});
