var sandbox = require("../../testSandbox");
var expect = require("expect.js");

var React = require("react");
var ReactDOM = require("react-dom");
var Backbone = require("backbone");
var ModelPropWatcherMixin = require("./ModelPropWatcherMixin");

describe("ModelPropWatcherMixin", () => {
	function getDummyComponent() {
		return React.createClass({
			displayName: "dummyComponent",
			mixins: [ModelPropWatcherMixin],
			render: () => <div key="dummy"></div>,
		});
	}

	it("should throw an exception if there is no 'model' prop", () => {
		var Item = getDummyComponent();
		var thrower = sandbox.renderComponent.bind(Item, <Item />);

		expect(thrower).to.throwException();
	});

	it("should initialize state with model attributes and listen to changes", () => {
		var Item = getDummyComponent();
		var attrs = {a: 2};
		var model = new Backbone.Model(attrs);

		model.on = sandbox.mock();
		var component = sandbox.renderComponent(<Item model={model} />);

		expect(component.state).to.eql(attrs);
		expect(model.on.callCount).to.be(1);
		expect(model.on.firstCall.args[0]).to.be("add remove change");
	});

	it("should set state with new attributes on change", () => {
		var Item = getDummyComponent();
		var attrs = {a: 2};
		var model = new Backbone.Model(attrs);
		var component = sandbox.renderComponent(<Item model={model} />);

		expect(component.state.a).to.be(2);
		model.set("a", 5);
		expect(component.state.a).to.be(5);
	});

	it("should unbind itself from model changes when component is unmounted", () => {
		var Item = getDummyComponent();
		var attrs = {a: 2};
		var model = new Backbone.Model(attrs);

		model.off = sandbox.mock();
		var component = sandbox.renderComponent(<Item model={model} />);

		ReactDOM.unmountComponentAtNode(ReactDOM.findDOMNode(component).parentNode);
		expect(model.off.callCount).to.be(1);
	});
});
