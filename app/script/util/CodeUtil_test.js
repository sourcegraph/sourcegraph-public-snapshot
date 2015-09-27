var sandbox = require("../testSandbox");
var expect = require("expect.js");

var CodeUtil = require("./CodeUtil");
var $ = require("jquery");

describe("util/CodeUtil", () => {
	it("should correctly set request headers for JSON file requests", () => {
		sandbox.stub($, "ajax", () => {
			return $.Deferred().resolve({});
		});

		CodeUtil.fetchFile("/path/to/file");
		expect($.ajax.firstCall.args[0].contentType).to.be("application/json");
	});

	it("should do a correct AJAX request when fetching popup", () => {
		sandbox.stub($, "ajax", () => {
			return $.Deferred().resolve("dummyData");
		});

		CodeUtil.fetchPopup("/a/b/c");

		expect($.ajax.callCount).to.be(1);
		expect($.ajax.firstCall.args[0].url).to.be("/ui/a/b/c");
		expect($.ajax.firstCall.args[0].headers.hasOwnProperty("X-Definition-Data-Only")).to.be(true);
		expect($.ajax.firstCall.args[0].headers["X-Definition-Data-Only"]).to.be("yes");
	});

	it("should fetch from cache when popup is already available", () => {
		sandbox.stub($, "ajax", () => {
			return $.Deferred().resolve("dummyData");
		});

		CodeUtil.fetchPopup("a/b/c");
		CodeUtil.fetchPopup("a/b/c");
		CodeUtil.fetchPopup("a/b/c");

		// only request from server once - the rest should come from cache
		expect($.ajax.callCount).to.be(1);
	});

	it("should call correct URL for fetching a popover", () => {
		sandbox.stub($, "ajax", () => {
			return $.Deferred().resolve("dummyData");
		});

		CodeUtil.fetchPopover("a/b/c");
		expect($.ajax.callCount).to.be(1);
		expect($.ajax.firstCall.args[0].url).to.be("a/b/c/.popover");
	});

	it("should resort to cache when fetching an already fetched popover", () => {
		sandbox.stub($, "ajax", () => {
			return $.Deferred().resolve("dummyData");
		});

		CodeUtil.fetchPopover("a/b/c/d");
		CodeUtil.fetchPopover("a/b/c/d");

		expect($.ajax.callCount).to.be(1);
	});

	it("should fetch examples using the correct params", () => {
		var testUrl = "/repo@rev/.unitType/unit/.def/path";

		sandbox.stub($, "ajax", () => {
			return $.Deferred().resolve("dummyData");
		});

		CodeUtil.fetchExample(testUrl, 1);

		expect($.ajax.callCount).to.be(1);
		expect($.ajax.firstCall.args[0].url).to.contain(`/ui${testUrl}`);
		expect($.ajax.firstCall.args[0].dataType).to.be("json");
		expect($.ajax.firstCall.args[0].data.Page).to.be(1);
	});
});
