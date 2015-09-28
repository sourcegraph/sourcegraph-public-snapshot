var sandbox = require("../testSandbox");
var expect = require("expect.js");
var sinon = require("sinon");

var $ = require("jquery");
var React = require("react/addons");
var TestUtils = React.addons.TestUtils;
var DiscussionCollection = require("../stores/collections/DiscussionCollection");
var DiscussionSnippet = require("./DiscussionSnippet");

describe("components/DiscussionSnippet", () => {
	var withCommentsAndRatings = id => {
		return {
			ID: id,
			Title: "title1",
			Description: "description1",
			Author: {Login: "gbbr"},
			Ratings: ["gbbr", "sqs", "dmitri"],
			Comments: [{
				ID: id+"1",
				Author: {Login: "gbbr"},
				CreatedAt: new Date(),
				Body: "comment_body",
			}, {
				ID: id+"2",
				Author: {Login: "sqs"},
				CreatedAt: new Date(),
				Body: "comment_body",
			}],
		};
	};

	var withNoCommentNorRatings = id => {
		return {
			ID: id,
			Title: "title2",
			Description: "description2",
			Author: {Login: "gbbr"},
			Ratings: [],
			Comments: [],
		};
	};

	it("should render message with a link to create a discussion when no discussions exist", () => {
		var collection = new DiscussionCollection([]);
		var view = sandbox.renderComponent(
			<DiscussionSnippet defKey="key" model={collection} onClick={() => {}} />
		);

		var $root = $(React.findDOMNode(view));
		expect($root.find(".no-discussions").length).to.be(1);
		expect($root.find(".no-discussions a").length).to.be(1);
	});

	it("should allow creating new discussions when no discussions exist", () => {
		var collection = new DiscussionCollection([]);
		var onCreate = sinon.stub();
		var view = sandbox.renderComponent(
			<DiscussionSnippet defKey="key" model={collection} onClick={() => {}} onCreate={onCreate} />
		);

		TestUtils.Simulate.click(view.refs.createBtn);
		expect(onCreate.callCount).to.be(1);
	});

	it("should render correctly with discussions and toolbar", () => {
		var collection = new DiscussionCollection([
			withCommentsAndRatings(1),
			withNoCommentNorRatings(2),
		]);

		var view = sandbox.renderComponent(
			<DiscussionSnippet defKey="key" toolbar={true} onClick={() => {}} model={collection} />
		);

		var $root = $(React.findDOMNode(view));
		var lis = $root.find("ul.list li");
		expect(lis.length).to.be(2);

		expect($(lis[0]).find("a.title").text()).to.be("title1");
		expect($(lis[0]).find(".stats").text()).to.contain("2");
		expect($(lis[0]).find(".stats").text()).to.contain("3");
		expect($(lis[0]).find("p.body").text()).to.contain("description1");

		expect($(lis[1]).find("a.title").text()).to.be("title2");
		expect($(lis[1]).find(".stats").text()).to.contain("0");
		expect($(lis[1]).find("p.body").text()).to.contain("description2");

		expect($root.find("footer").length).to.be(1);
		expect($root.find("footer a").length).to.be(2);
	});

	it("should not render more than 4 discussions in the snippet", () => {
		var collection = new DiscussionCollection([
			withCommentsAndRatings(1),
			withCommentsAndRatings(2),
			withCommentsAndRatings(3),
			withCommentsAndRatings(4),
			withNoCommentNorRatings(5),
		]);

		var view = sandbox.renderComponent(
			<DiscussionSnippet defKey="key" model={collection} onClick={() => {}} />
		);

		var lis = $(React.findDOMNode(view)).find("ul.list li");
		expect(lis.length).to.be(4);
	});

	it("should correctly use given callbacks for viewing, creating and listing", () => {
		var collection = new DiscussionCollection([withCommentsAndRatings(1)]);
		var onCreate = sinon.stub();
		var onList = sinon.stub();
		var onClick = sinon.stub();
		var view = sandbox.renderComponent(
			<DiscussionSnippet
				defKey="key"
				toolbar={true}
				model={collection}
				onCreate={onCreate}
				onClick={onClick}
				onList={onList} />
		);

		TestUtils.Simulate.click(view.refs.listBtn);
		expect(onList.callCount).to.be(1);
		expect(onList.firstCall.args[0]).to.be("key");

		TestUtils.Simulate.click(view.refs.createBtn);
		expect(onCreate.callCount).to.be(1);

		var discussion = TestUtils.findRenderedDOMComponentWithClass(view, "discussion");
		var discussionTitle = TestUtils.findRenderedDOMComponentWithTag(discussion, "a");

		TestUtils.Simulate.click(discussionTitle);
		expect(onClick.callCount).to.be(1);
	});
});
