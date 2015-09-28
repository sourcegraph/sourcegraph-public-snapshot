var sandbox = require("../testSandbox");

var expect = require("expect.js");
var $ = require("jquery");
var React = require("react/addons");
var TestUtils = React.addons.TestUtils;
var RepoBuildIndicator = require("./RepoBuildIndicator");
var client = require("../client");

describe("RepoBuildIndicator", () => {
	// Render function tests. The object key is also the expected BuildStatus.
	var renderTests = {
		FAILURE: {
			Failure: true,
			EndedAt: "2014-12-20 22:53:11",
			CommitID: "CommID123",
			Attempt: 1,
			expect: {
				cls: "danger",
				txt: "build failed",
				icon: "fa-exclamation-circle",
			},
		},
		BUILT: {
			Success: true,
			EndedAt: "2014-12-20 22:53:11",
			Attempt: 1,
			CommitID: "CSmmID123",
			expect: {
				cls: "success",
				txt: "built",
				icon: "fa-check",
			},
		},
		STARTED: {
			StartedAt: "2014-12-20 22:53:11",
			CommitID: "CTmmID123",
			Attempt: 1,
			expect: {
				cls: "info",
				txt: "started",
				icon: "fa-spin",
			},
		},
		QUEUED: {
			CreatedAt: "2014-12-20 22:53:11",
			CommitID: "CQmmID123",
			Attempt: 1,
			expect: {
				cls: "warning",
				txt: "queued",
				icon: "fa-clock-o",
			},
		},
	};

	Object.keys(renderTests).forEach((name) => {
		it("should have correct classes, attributes and state when build is: " + name, () => {
			var test = renderTests[name];
			sandbox.stub(client, "builds", () => {
				return $.Deferred().resolve({Builds: [test]}).promise();
			});

			var component = sandbox.renderComponent(
				<RepoBuildIndicator btnSize="test-size" RepoURI="test-uri" rev="test-rev" />
			);
			var tag = TestUtils.findRenderedDOMComponentWithTag(component, "a");
			var $node = $(tag.getDOMNode());

			expect(client.builds.callCount).to.be(1);
			expect(component.state.status).to.be(component.BuildStatus[name]);
			expect($node.hasClass("btn-"+test.expect.cls)).to.be(true);
			expect($node.attr("href")).to.be("/test-uri/.builds/"+test.CommitID+"/"+test.Attempt);
			expect($node.attr("title")).to.contain(test.CommitID.slice(0, 6)+" "+test.expect.txt);

			tag = TestUtils.findRenderedDOMComponentWithTag(component, "i");
			expect($(tag.getDOMNode()).hasClass(test.expect.icon)).to.be(true);
		});
	}, this);

	it("should display build link when one is not available", () => {
		sandbox.stub(client, "builds", function() {
			return $.Deferred().resolve([]).promise();
		});

		var component = sandbox.renderComponent(
			<RepoBuildIndicator btnSize="test-size" RepoURI="test-uri" rev="test-rev" />
		);

		var tag = TestUtils.findRenderedDOMComponentWithTag(component, "a");
		var $node = $(tag.getDOMNode());
		expect($node.attr("title")).to.contain("Not yet built.");
	});

	it("should not display label if label prop is not set", () => {
		var component = sandbox.renderComponent(
			<RepoBuildIndicator LastBuild={renderTests.STARTED} btnSize="test-size" RepoURI="test-uri" rev="test-rev" />
		);
		var $el = $(component.getDOMNode());
		expect($el.html()).not.to.contain("Build started");
	});

	it("should display label if label prop is set to \"yes\"", () => {
		var component = sandbox.renderComponent(
			<RepoBuildIndicator LastBuild={renderTests.STARTED} Label="yes" btnSize="test-size" RepoURI="test-uri" rev="test-rev" />
		);
		var $el = $(component.getDOMNode());
		expect($el.html()).to.contain("Build started");
	});

	it("should request a new build and change cache when clicked with no build available", () => {
		sandbox.stub(client, "builds", () => $.Deferred().resolve([]).promise());
		sandbox.stub(client, "createRepoBuild", () => $.Deferred().resolve([]).promise());

		var component = sandbox.renderComponent(
			<RepoBuildIndicator btnSize="test-size" RepoURI="test-uri" rev="test-rev" Buildable={true} />
		);
		expect(component.state.status).to.be(component.BuildStatus.NA);

		TestUtils.Simulate.click(TestUtils.findRenderedDOMComponentWithTag(component, "a"));
		expect(client.createRepoBuild.callCount).to.be(1);
		expect(component.state.noCache).to.be(true);
	});

	it("should render a span with class 'text-danger' on error", () => {
		sandbox.stub(client, "builds", () => $.Deferred().reject().promise());

		var component = sandbox.renderComponent(
			<RepoBuildIndicator btnSize="test-size" RepoURI="test-uri" rev="test-rev" />
		);
		expect(component.state.status).to.be(component.BuildStatus.ERROR);

		var tag = TestUtils.findRenderedDOMComponentWithTag(component, "span");
		var $node = $(tag.getDOMNode());

		expect($node.hasClass("text-danger")).to.be(true);
		expect($node.text()).to.contain("Error");
	});

	it("should not request build status if build is provided in props", () => {
		sandbox.stub(client, "builds");

		sandbox.renderComponent(
			<RepoBuildIndicator LastBuild={renderTests.FAILURE} btnSize="test-size" RepoURI="test-uri" rev="test-rev" />
		);

		expect(client.builds.callCount).to.be(0);
	});

	it("should clear and start a new poller when build is STARTED", () => {
		sandbox.stub(global, "clearInterval");
		sandbox.stub(global, "setInterval");

		sandbox.renderComponent(
			<RepoBuildIndicator LastBuild={renderTests.STARTED} btnSize="test-size" RepoURI="test-uri" rev="test-rev" />
		);

		expect(clearInterval.callCount).to.be(1);
		expect(setInterval.callCount).to.be(1);
	});

	it("should clear interval and not start a new poller when a build is BUILT", () => {
		sandbox.stub(global, "clearInterval");
		sandbox.stub(global, "setInterval");

		sandbox.renderComponent(
			<RepoBuildIndicator LastBuild={renderTests.BUILT} btnSize="test-size" RepoURI="test-uri" rev="test-rev" />
		);

		expect(clearInterval.callCount).to.be(1);
		expect(setInterval.callCount).to.be(0);
	});

	it("should continue polling for build status QUEUED", () => {
		sandbox.stub(client, "builds", () => $.Deferred().resolve({Builds: [renderTests.QUEUED]}).promise());
		sandbox.useFakeTimers();
		sandbox.spy(global, "setInterval");

		sandbox.renderComponent(
			<RepoBuildIndicator btnSize="test-size" RepoURI="test-uri" rev="test-rev" />
		);

		expect(client.builds.callCount).to.be(1);
		expect(setInterval.callCount).to.be(1);

		sandbox.clock.tick(10000);

		expect(client.builds.callCount).to.be(2);
		expect(setInterval.callCount).to.be(2);
	});

	it("should poll for build status if props is STARTED, and stop after FAILURE", () => {
		sandbox.stub(client, "builds", () => $.Deferred().resolve({Builds: [renderTests.FAILURE]}));
		sandbox.useFakeTimers();
		sandbox.spy(global, "clearInterval");
		sandbox.spy(global, "setInterval");

		sandbox.renderComponent(
			<RepoBuildIndicator LastBuild={renderTests.STARTED} btnSize="test-size" RepoURI="test-uri" rev="test-rev" />
		);

		expect(client.builds.callCount).to.be(0);
		expect(clearInterval.callCount).to.be(1);
		expect(setInterval.callCount).to.be(1);

		sandbox.clock.tick(5000);

		expect(client.builds.callCount).to.be(1);
		expect(clearInterval.callCount).to.be(2);
		expect(setInterval.callCount).to.be(1);
	});

	it("should reload page when succeeding with success-reload on", () => {
		global.location = {reload: sandbox.stub()};
		sandbox.stub(client, "builds", () => $.Deferred().resolve({Builds: [renderTests.BUILT]}));
		sandbox.useFakeTimers();

		sandbox.renderComponent(
			<RepoBuildIndicator LastBuild={renderTests.STARTED} btnSize="test-size" RepoURI="test-uri" rev="test-rev" SuccessReload="on" />
		);

		sandbox.clock.tick(5000);

		expect(location.reload.callCount).to.be(1);
	});

	it("should not reload page when succeeding with success-reload undefined", () => {
		global.location = {reload: sandbox.stub()};
		sandbox.stub(client, "builds", () => $.Deferred().resolve({Builds: [renderTests.BUILT]}));
		sandbox.useFakeTimers();

		sandbox.renderComponent(
			<RepoBuildIndicator LastBuild={renderTests.STARTED} btnSize="test-size" RepoURI="test-uri" rev="test-rev" />
		);

		sandbox.clock.tick(5000);

		expect(location.reload.callCount).to.be(0);
	});
});
