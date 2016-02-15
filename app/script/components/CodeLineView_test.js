var sandbox = require("../testSandbox");
var expect = require("expect.js");

var $ = require("jquery");
var React = require("react");
var CodeLineView = require("./CodeLineView");
var CodeLineModel = require("../stores/models/CodeLineModel");

describe("components/CodeLineView", () => {
	it("should register its node with the model", () => {
		var model = new CodeLineModel();

		sandbox.renderComponent(<table><tbody><CodeLineView model={model} /></tbody></table>);
		expect(model.__node).not.to.be(null);
	});

	it("should register line numbers by default", () => {
		var model = new CodeLineModel({number: 5});
		var component = sandbox.renderComponent(<table><tbody><CodeLineView model={model} /></tbody></table>);

		expect($(component.querySelector(".line-number")).data("line")).to.be(5);
	});

	it("should not register line numbers if specifically disabled via props", () => {
		var model = new CodeLineModel();
		var component = sandbox.renderComponent(
			<table><tbody><CodeLineView lineNumbers={false} model={model} /></tbody></table>
		);

		expect(component.querySelectorAll(".line-number").length).to.be(0);
	});

	it("should render 1 whitespace if the line is empty", () => {
		var model = new CodeLineModel();
		var component = sandbox.renderComponent(<table><tbody><CodeLineView model={model} /></tbody></table>);

		expect(component.querySelector(".line-content").textContent).to.be(" ");
	});

	it("should render plain text correctly", () => {
		var model = new CodeLineModel({contents: "hello"});
		var component = sandbox.renderComponent(<table><tbody><CodeLineView model={model} /></tbody></table>);

		var td = component.querySelector(".line-content");

		expect($(td).children().length).to.be(1);
		expect($(td).children("span")[0].innerHTML).to.be("hello");
	});

	it("should apply main-byte-range class when line model is highlighted", () => {
		var model = new CodeLineModel({highlight: true});
		var component = sandbox.renderComponent(<table><tbody><CodeLineView model={model} /></tbody></table>);

		var tok = component.querySelector("tr");

		expect($(tok).hasClass("main-byte-range")).to.be(true);
		expect($(tok).hasClass("line")).to.be(true);
	});

	it("should not apply main-byte-range class when line model is not highlighted", () => {
		var model = new CodeLineModel();
		var component = sandbox.renderComponent(<table><tbody><CodeLineView model={model} /></tbody></table>);

		var tok = component.querySelector("tr");

		expect($(tok).hasClass("main-byte-range")).not.to.be(true);
		expect($(tok).hasClass("line")).to.be(true);
	});
});
