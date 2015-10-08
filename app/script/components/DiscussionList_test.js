var sandbox = require("../testSandbox");
var expect = require("expect.js");
var sinon = require("sinon");

var $ = require("jquery");
var React = require("react");
var ReactDOM = require("react-dom");
var TestUtils = require("react-addons-test-utils");
var DiscussionCollection = require("../stores/collections/DiscussionCollection");
var DiscussionList = require("./DiscussionList");

describe("components/DiscussionList", () => {
	var withCommentsAndRatings = id => {
		return {
			ID: id,
			Title: "title1",
			Description: "description1",
			Author: {Login: "gbbr"},
			Ratings: ["gbbr", "sqs", "dmitri"],
			CreatedAt: new Date(),
			Comments: [{
				ID: `${id}1`,
				Author: {Login: "gbbr"},
				CreatedAt: new Date(),
				Body: "comment_body",
			}, {
				ID: `${id}2`,
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
			CreatedAt: new Date(),
			Ratings: [],
			Comments: [],
		};
	};

	it("should render a list of discussions", () => {
		var collection = new DiscussionCollection([
			withCommentsAndRatings(1),
			withCommentsAndRatings(2),
			withCommentsAndRatings(3),
			withCommentsAndRatings(4),
			withNoCommentNorRatings(5),
		]);
		var view = sandbox.renderComponent(
			<DiscussionList defKey="key" defName={{__html: "name"}} model={collection} />
		);

		// checkAgainst checks that a DOM node contains correct HTML
		// to render the given props.
		var checkAgainst = (node, attrs) => {
			var $header = $(node).find("header");
			expect($header.find("h1 .contents").text()).to.be(`${attrs.Title} #${attrs.ID}`);
			expect($header.find(".stats").text()).to.be(` ${attrs.Comments.length} `);
			expect($header.find(".subtitle").text()).to.be(`@${attrs.Author.Login} a few seconds ago`);
			expect($(node).find("p.body").text()).to.contain(attrs.Description);
		};

		var $root = $(ReactDOM.findDOMNode(view));
		var all = $root.find(".discussions-list li");

		expect(all.length).to.be(5);
		all.each((i, d) => checkAgainst(d, collection.models[i].attributes));
	});

	it("should allow creating new discussions", () => {
		var collection = new DiscussionCollection([]);
		var onCreate = sinon.stub();
		var view = sandbox.renderComponent(
			<DiscussionList defKey="key" defName={{__html: "name"}} model={collection} onCreate={onCreate} />
		);

		TestUtils.Simulate.click(view.refs.createBtn);
		expect(onCreate.callCount).to.be(1);
	});

	it("should allow clicking listed discussions", () => {
		var collection = new DiscussionCollection([withCommentsAndRatings(1)]);
		var onClick = sinon.stub();
		var view = sandbox.renderComponent(
			<DiscussionList defKey="key" defName={{__html: "name"}} model={collection} onClick={onClick} />
		);

		TestUtils.Simulate.click(TestUtils.findRenderedDOMComponentWithTag(view, "li"));
		expect(onClick.callCount).to.be(1);
	});
});
