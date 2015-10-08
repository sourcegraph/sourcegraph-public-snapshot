var sandbox = require("../testSandbox");
var expect = require("expect.js");
var sinon = require("sinon");

var React = require("react/addons");
var TestUtils = React.addons.TestUtils;

var Pagination = require("./Pagination");

describe("components/Pagination", () => {
	it("displays the correct number of page links", () => {
		var props = {
			currentPage: 1,
			totalPages: 10,
			pageRange: 10,
			onPageChange: () => {},
		};

		var component = sandbox.renderComponent(<Pagination {...props} />);
		var pageLinks = TestUtils.scryRenderedDOMComponentsWithClass(component, "num-page-link");

		expect(pageLinks.length).to.be(props.totalPages);

		var pageLink;
		for (var i=0; i < props.totalPages; i++) {
			pageLink = pageLinks[i];
			expect(pageLink).to.be.ok();
			expect(React.findDOMNode(pageLink).textContent).to.be((i+1).toString());
		}
	});

	it("is bounded by the total number of page links", () => {
		var props = {
			currentPage: 1,
			totalPages: 5,
			pageRange: 10,
			onPageChange: () => {},
		};

		var component = sandbox.renderComponent(<Pagination {...props} />);
		var pageLinks = TestUtils.scryRenderedDOMComponentsWithClass(component, "num-page-link");

		expect(pageLinks.length).to.be(props.totalPages);
	});

	it("is bounded by the total number of page links on the last page", () => {
		var props = {
			currentPage: 42,
			totalPages: 42,
			pageRange: 10,
			onPageChange: () => {},
		};

		var component = sandbox.renderComponent(<Pagination {...props} />);
		var pageLinks = TestUtils.scryRenderedDOMComponentsWithClass(component, "num-page-link");
		var lastPageLink = pageLinks[pageLinks.length-1];

		expect(React.findDOMNode(lastPageLink).textContent).to.be(props.totalPages.toString());
	});

	it("calls the onPageChange callback when a new page is selected", () => {
		var props = {
			currentPage: 1,
			totalPages: 10,
			pageRange: 10,
			onPageChange: sinon.stub(),
		};

		var component = sandbox.renderComponent(<Pagination {...props} />);
		var pageLinks = TestUtils.scryRenderedDOMComponentsWithClass(component, "num-page-link");

		var newPage = 7;
		var newPageLink = pageLinks[newPage-1];
		TestUtils.Simulate.click(newPageLink);

		expect(props.onPageChange.callCount).to.be(1);
		expect(props.onPageChange.firstCall.args[0]).to.be(newPage);
	});
});
