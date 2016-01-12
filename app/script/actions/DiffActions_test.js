var sandbox = require("../testSandbox");
var expect = require("expect.js");

var $ = require("jquery");
var React = require("react");
var ReactDOM = require("react-dom");
var TestUtils = require("react-addons-test-utils");
var Backbone = require("backbone");
var DiffActions = require("./DiffActions");
var HunkModel = require("../stores/models/HunkModel");
var FileDiffModel = require("../stores/models/FileDiffModel");
var DiffStore = require("../stores/DiffStore");

describe("actions/DiffActions", () => {
	it("should expand upward", () => {
    DiffStore.set("RepoRevSpec", {URI: "foo/bar", Head: {Rev: "feature-branch"}});
    DiffStore.set("DeltaSpec", {Head: {Rev: "feature-branch"}});

    var fileDiff = new FileDiffModel({
      NewName: "abc",
      Hunks: new Backbone.Collection(
        new HunkModel({}, {parse: false})
      ),
    }, {parse: false});
    var hunk = new HunkModel({
      NewStartLine: 10,
      Parent: fileDiff,
    }, {parse: false});

    var l = DiffActions.expandHunkUp(hunk);
    console.log(l);
    expect(l.startLine).to.be(1);
    expect(l.endLine).to.be(8);
	});
});
