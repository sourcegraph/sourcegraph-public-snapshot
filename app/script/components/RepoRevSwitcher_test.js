var sandbox = require("../testSandbox");
var expect = require("expect.js");

var $ = require("jquery");
var React = require("react");
var TestUtils = require("react-addons-test-utils");
var RepoRevSwitcher = require("./RepoRevSwitcher");
var router = require("../routing/router");

describe("components/RepoRevSwitcher", () => {
	it("should display rev dropdown menu on click", () => {
		var props = {
			path: ".",
			repoSpec: "sourcegraph/milton",
			rev: "master",
		};

		var component = sandbox.renderComponent(<RepoRevSwitcher {...props} />);

		var btn = TestUtils.findRenderedDOMComponentWithClass(component, "repo-rev-switcher");
		TestUtils.Simulate.click(btn);

		expect($(React.findDOMNode(btn)).find(".dropdown-menu").html()).to.be.ok();
	});

	it("should link to file URLs for available branches", () => {
		var props = {
			path: "woof.go",
			repoSpec: "sourcegraph/milton",
			rev: "master",
		};

		var component = sandbox.renderComponent(<RepoRevSwitcher {...props} />);
		component.setState({
			branches: {
				Branches: [
					{
						Head: "bc22ed45cb8ce3e767417fe88e9d1005d350d85f",
						Name: "master",
					},
					{
						Head: "d080ef6774ffbf9eae234361a69f792cadfb0d7b",
						Name: "hotfix",
					},
				],
			},
		});

		var btn = TestUtils.findRenderedDOMComponentWithClass(component, "repo-rev-switcher");
		var menu = TestUtils.findRenderedDOMComponentWithClass(component, "dropdown-menu");
		TestUtils.Simulate.click(btn);

		var branchHrefs = Reflect.apply(Array.prototype.map, menu.querySelectorAll("a"), [(a) => $(a).attr("href")]);
		expect(branchHrefs).to.contain(router.fileURL(props.repoSpec, "master", props.path));
		expect(branchHrefs).to.contain(router.fileURL(props.repoSpec, "hotfix", props.path));
	});

	it("should link to a specified route URL when one is supplied", () => {
		var props = {
			path: ".",
			repoSpec: "sourcegraph/milton",
			rev: "master",
			route: "commits",
		};

		var component = sandbox.renderComponent(<RepoRevSwitcher {...props} />);
		component.setState({
			branches: {
				Branches: [
					{
						Head: "bc22ed45cb8ce3e767417fe88e9d1005d350d85f",
						Name: "master",
					},
				],
			},
		});

		var btn = TestUtils.findRenderedDOMComponentWithClass(component, "repo-rev-switcher");
		var menu = TestUtils.findRenderedDOMComponentWithClass(component, "dropdown-menu");
		TestUtils.Simulate.click(btn);

		var branchHrefs = Reflect.apply(Array.prototype.map, menu.querySelectorAll("a"), [(a) => $(a).attr("href")]);
		expect(branchHrefs).to.contain(router.commitsURL(props.repoSpec, "master"));
	});
});
