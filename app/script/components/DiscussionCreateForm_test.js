var sandbox = require("../testSandbox");
var expect = require("expect.js");
var sinon = require("sinon");

var $ = require("jquery");
var React = require("react/addons");
var TestUtils = React.addons.TestUtils;
var DiscussionCreateForm = require("./DiscussionCreateForm");

describe("components/DiscussionCreateForm", () => {
	it("should render the component", () => {
		var view = sandbox.renderComponent(
			<DiscussionCreateForm
				defName={{__html: "defName"}}
				onCancel={() => {}}
				onCreate={() => {}} />
		);

		expect($(React.findDOMNode(view)).find("p").text()).to.contain("defName");
	});

	it("should correctly submit or cancel a discussion creation", () => {
		var view = sandbox.renderComponent(
			<DiscussionCreateForm
				defName={{__html: "defName"}}
				onCancel={sinon.stub()}
				onCreate={sinon.stub()} />
		);

		TestUtils.Simulate.click(view.refs.cancelBtn);
		expect(view.props.onCancel.callCount).to.be(1);

		$(React.findDOMNode(view.refs.titleText)).val("discussion-title");
		sinon.stub(view.refs.bodyText, "value").returns("discussion-body");
		TestUtils.Simulate.click(view.refs.createBtn);
		expect(view.props.onCreate.callCount).to.be(1);
		expect(view.props.onCreate.firstCall.args[0]).to.be("discussion-title");
		expect(view.props.onCreate.firstCall.args[1]).to.be("discussion-body");
	});
});
