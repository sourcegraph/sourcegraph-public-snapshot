var sandbox = require("../testSandbox");
var expect = require("expect.js");

var $ = require("jquery");
var React = require("react");
var ReactDOM = require("react-dom");
var ContextMenu = require("./ContextMenu");
var ContextMenuModel = require("../stores/models/ContextMenuModel");

describe("components/ContextMenu", () => {
	it("should render correctly when given 2 defs", () => {
		var defs = [
			{URL: "/user/srclib-testcases@9884e89566c25c72e6ca85aa8e201d4bf29e7ec3/.GoPackage/user/srclib-testcases/.def/ES/S"},
			{URL: "/user/srclib-testcases@9884e89566c25c72e6ca85aa8e201d4bf29e7ec3/.GoPackage/user/srclib-testcases/.def/ES/S2"},
		];
		var options = defs.map(def => {
			return {
				label: def.QualifiedName,
				data: def,
			};
		});
		var model = new ContextMenuModel({options: options, closed: false});
		var component = sandbox.renderComponent(<ContextMenu model={model} />);

		var $root = $(ReactDOM.findDOMNode(component));
		var lis = $root.find("ul li");
		expect(lis.length).to.be(2);
	});
});
