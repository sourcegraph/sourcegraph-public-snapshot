var sandbox = require("../testSandbox");
var expect = require("expect.js");
var sinon = require("sinon");

var $ = require("jquery");
var React = require("react");
var ReactDOM = require("react-dom");
var TestUtils = require("react-addons-test-utils");
var DiscussionModel = require("../stores/models/DiscussionModel");
var DiscussionView = require("./DiscussionView");

describe("components/DiscussionView", () => {
	var withCommentsData = {
		ID: 1,
		Title: "title",
		CreatedAt: new Date(),
		Author: {Login: "gbbr"},
		Description: "description",
		Ratings: ["gbbr", "shurcooL"],
		Comments: [{
			ID: 11,
			Author: {Login: "shurcooL"},
			CreatedAt: new Date(),
			Body: "comment_body\n",
		}],
	};

	it("should render a discussion with comments and description correctly", () => {
		var model = new DiscussionModel(withCommentsData);
		var view = sandbox.renderComponent(
			<DiscussionView defName={{__html: "name"}} defKey="key" model={model} />
		);

		var $root = $(ReactDOM.findDOMNode(view));
		var $header = $root.find("header");

		expect($header.find("h1").text()).to.be("title #1");
		expect($header.find(".stats").text()).to.be(" 1 ");
		expect($header.find(".subtitle .author").text()).to.be("@gbbr");
		expect($header.find(".subtitle .date").text()).to.be(" a few seconds ago");
		expect($header.find(".subtitle .subject").text()).to.be(" on name");
		expect($root.find("main.body").text()).to.be("description");

		var $comment = $root.find("ul.thread-comments li").first();

		expect($comment.find(".signature").text()).to.be("@shurcooL replied a few seconds ago");
		expect($comment.find(".markdown-view").text()).to.be("comment_body\n");
	});

	it("should render a discussion without comments and description correctly", () => {
		var model = new DiscussionModel({
			ID: 1,
			Title: "title",
			Author: {Login: "gbbr"},
			Ratings: [],
		});
		var view = sandbox.renderComponent(
			<DiscussionView defName={{__html: "name"}} defKey="key" model={model} />
		);

		var $root = $(ReactDOM.findDOMNode(view));

		expect($root.find("main.body").length).to.be(0);
		expect($root.find("ul.thread-comments").length).to.be(0);
	});

	it("should call correct callbacks when actions occur", () => {
		var model = new DiscussionModel(withCommentsData);
		var onList = sinon.stub();
		var onCreate = sinon.stub();
		var onComment = sinon.stub();
		var view = sandbox.renderComponent(
			<DiscussionView
				defName={{__html: "name"}}
				defKey="key"
				onCreate={onCreate}
				onList={onList}
				onComment={onComment}
				model={model} />
		);

		sinon.stub(view.refs.commentTextarea, "value").returns("body");

		TestUtils.Simulate.click(view.refs.commentBtn);
		expect(onComment.callCount).to.be(1);
		expect(onComment.firstCall.args[0]).to.be(1);
		expect(onComment.firstCall.args[1]).to.be("body");

		TestUtils.Simulate.click(view.refs.listBtn);
		expect(onList.callCount).to.be(1);

		TestUtils.Simulate.click(view.refs.createBtn);
		expect(onCreate.callCount).to.be(1);
	});
});
