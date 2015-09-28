var sandbox = require("../testSandbox");
var expect = require("expect.js");

var $ = require("jquery");
var CodeFileActions = require("./CodeFileActions");
var Dispatcher = require("../dispatchers/AppDispatcher");
var CodeUtil = require("../util/CodeUtil");
var CodeTokenModel = require("../stores/models/CodeTokenModel");
var CodeFileStore = require("../stores/CodeFileStore");
var globals = require("../globals");
var router = require("../routing/router");

describe("actions/CodeFileActions", () => {
	function getTokenWithProps(props) {
		return new CodeTokenModel(props);
	}

	xit("deselect tokens should dispatch TOKEN_CLEAR", () => {
		CodeFileActions.deselectTokens();

		expect(Dispatcher.handleViewAction.callCount).to.be(1);
		expect(Dispatcher.handleViewAction.firstCall.args[0].type).to.be(globals.Actions.TOKEN_CLEAR);
	});

	it("focus token action should not proceed if token is selected", () => {
		var token = getTokenWithProps({
			url: ["tokenUrl"],
			selected: true,
		});
		sandbox.stub(CodeUtil, "fetchPopover");
		sandbox.spy(Dispatcher, "handleViewAction");

		CodeFileActions.focusToken(token, null);

		expect(Dispatcher.handleViewAction.callCount).to.be(0);
		expect(CodeUtil.fetchPopover.callCount).to.be(0);
	});

	xit("blurring tokens should dispatch TOKEN_BLUR and abort popover xhr", () => {
		sandbox.spy(Dispatcher, "handleViewAction");
		CodeFileActions.blurTokens();

		expect(Dispatcher.handleViewAction.callCount).to.be(1);
		expect(Dispatcher.handleViewAction.firstCall.args[0].type).to.be(globals.Actions.TOKEN_BLUR);

		expect(CodeUtil.abortPopoverXhr.callCount).to.be(1);
	});

	it("show definition action creator should dispatch SHOW_DEFINITION and pass params if same file", () => {
		var def = {File: "filename.ext"};

		sandbox.stub(CodeFileStore, "isSameFile", () => true);
		sandbox.stub(CodeUtil, "fetchFile");
		sandbox.spy(Dispatcher, "handleViewAction");

		CodeFileActions.navigateToDefinition(def);

		expect(CodeFileStore.isSameFile.callCount).to.be(1);
		expect(Dispatcher.handleViewAction.callCount).to.be(1);
		expect(Dispatcher.handleViewAction.firstCall.args[0].type).to.be(globals.Actions.SHOW_DEFINITION);
		expect(Dispatcher.handleViewAction.firstCall.args[0].params).to.be(def);

		expect(CodeUtil.fetchFile.callCount).to.be(0);
	});

	xit("show definition action creator should load correct file if necessarry", () => {
		var def = {File: "filename.ext", URL: "path/to/file"};

		sandbox.stub(CodeFileStore, "isSameFile", () => false);
		sandbox.stub(CodeUtil, "fetchFile", () => $.Deferred().resolve("dummyData"));
		sandbox.spy(Dispatcher, "handleViewAction");

		CodeFileActions.navigateToDefinition(def);

		expect(CodeFileStore.isSameFile.callCount).to.be(1);

		expect(Dispatcher.handleDependentAction.callCount).to.be(1);
		expect(Dispatcher.handleDependentAction.firstCall.args[0].type).to.be(globals.Actions.FETCH_FILE);
		expect(Dispatcher.handleDependentAction.firstCall.args[0].url).to.be("path/to/file");

		expect(CodeUtil.fetchFile.callCount).to.be(1);
		expect(CodeUtil.fetchFile.firstCall.args[0]).to.be("path/to/file");

		expect(Dispatcher.handleViewAction.callCount).to.be(1);
		expect(Dispatcher.handleViewAction.firstCall.args[0].type).to.be(globals.Actions.SHOW_DEFINITION);
		expect(Dispatcher.handleViewAction.firstCall.args[0].params).to.be(def);
	});

	it("fetching examples shoudl dispatch FETCH_EXAMPLE and call correct API function", () => {
		sandbox.stub(CodeUtil, "fetchExample", () => {
			return $.Deferred().resolve("dummyData");
		});
		sandbox.spy(Dispatcher, "handleViewAction");

		CodeFileActions.selectExample("/path/to/def", 1);

		expect(Dispatcher.handleViewAction.callCount).to.be(1);
		expect(Dispatcher.handleViewAction.firstCall.args[0].type).to.be(globals.Actions.FETCH_EXAMPLE);

		expect(CodeUtil.fetchExample.callCount).to.be(1);
		expect(CodeUtil.fetchExample.firstCall.args[0]).to.be("/path/to/def");
		expect(CodeUtil.fetchExample.firstCall.args[1]).to.be(1);
	});

	xit("show snippet should create correct action and not load file if same", () => {
		sandbox.stub(CodeFileStore, "isSameFile", () => true);
		sandbox.spy(Dispatcher, "handleViewAction");

		CodeFileActions.changeState("file", 0, 1, "path/to/def");

		expect(Dispatcher.handleViewAction.callCount).to.be(1);
		expect(Dispatcher.handleViewAction.firstCall.args[0].type).to.be(globals.Actions.SHOW_SNIPPET);
		expect(Dispatcher.handleViewAction.firstCall.args[0].params.file).to.be("file");
		expect(Dispatcher.handleViewAction.firstCall.args[0].params.startLine).to.be(0);
		expect(Dispatcher.handleViewAction.firstCall.args[0].params.endLine).to.be(1);
		expect(Dispatcher.handleViewAction.firstCall.args[0].params.defUrl).to.be("path/to/def");

		expect(CodeUtil.fetchFile.callCount).to.be(0);
	});

	xit("show snippet should load correct file if needed", () => {
		var file = {
			Path: "a/b/c",
			RepoRev: {
				URI: "uri",
				CommitID: "123abc",
			},
		};
		var fileUrl = router.fileURL(file.RepoRev.URI, file.RepoRev.CommitID, file.Path);

		sandbox.stub(CodeFileStore, "isSameFile", () => false);
		sandbox.stub(CodeUtil, "fetchFile", () => $.Deferred().resolve("dummyData"));
		sandbox.spy(Dispatcher, "handleViewAction");

		CodeFileActions.changeState(file, 0, 1, "path/to/def");

		expect(Dispatcher.handleDependentAction.callCount).to.be(1);
		expect(Dispatcher.handleDependentAction.firstCall.args[0].type).to.be(globals.Actions.FETCH_FILE);
		expect(Dispatcher.handleDependentAction.firstCall.args[0].url).to.be(fileUrl);

		expect(Dispatcher.handleViewAction.callCount).to.be(1);
		expect(Dispatcher.handleViewAction.firstCall.args[0].type).to.be(globals.Actions.SHOW_SNIPPET);
		expect(Dispatcher.handleViewAction.firstCall.args[0].params.file).to.be(file);
		expect(Dispatcher.handleViewAction.firstCall.args[0].params.startLine).to.be(0);
		expect(Dispatcher.handleViewAction.firstCall.args[0].params.endLine).to.be(1);
		expect(Dispatcher.handleViewAction.firstCall.args[0].params.defUrl).to.be("path/to/def");

		expect(CodeUtil.fetchFile.callCount).to.be(1);
		expect(CodeUtil.fetchFile.firstCall.args[0]).to.be(fileUrl);
	});
});
