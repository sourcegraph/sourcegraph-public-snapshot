import autotest from "sourcegraph/util/autotest";
import expect from "expect.js";

import React from "react";
import TestUtils from "react-addons-test-utils";

import Dispatcher from "sourcegraph/Dispatcher";
import BlobRouter from "sourcegraph/blob/BlobRouter";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import * as DefActions from "sourcegraph/def/DefActions";
import * as RepoActions from "sourcegraph/repo/RepoActions";
import RepoStore from "sourcegraph/repo/RepoStore";
import DefStore from "sourcegraph/def/DefStore";
import {GoTo} from "sourcegraph/util/hotLink";

import testdataFile from "sourcegraph/blob/testdata/BlobRouter-file.json";
import testdataDotfile from "sourcegraph/blob/testdata/BlobRouter-dotfile.json";
import testdataLineSelection from "sourcegraph/blob/testdata/BlobRouter-lineSelection.json";
import testdataDefSelection from "sourcegraph/blob/testdata/BlobRouter-defSelection.json";
import testdataDefinition from "sourcegraph/blob/testdata/BlobRouter-definition.json";

describe("BlobRouter", () => {
	it("should handle file URLs", () => {
		autotest(testdataFile, `${__dirname}/testdata/BlobRouter-file.json`,
			<BlobRouter location="http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go" />
		);
	});

	it("should handle dotfile URLs", () => {
		autotest(testdataDotfile, `${__dirname}/testdata/BlobRouter-dotfile.json`,
			<BlobRouter location="http://localhost:3080/github.com/gorilla/mux@master/.tree/.travis.yml" />
		);
	});

	it("should handle line selection URLs", () => {
		autotest(testdataLineSelection, `${__dirname}/testdata/BlobRouter-lineSelection.json`,
			<BlobRouter location="http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L40-53" _isMounted={true} />
		);
	});

	it("should handle definition selection URLs", () => {
		autotest(testdataDefSelection, `${__dirname}/testdata/BlobRouter-defSelection.json`,
			<BlobRouter location="http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#def-someDef" />
		);
	});

	it("should handle definition URLs", () => {
		autotest(testdataDefinition, `${__dirname}/testdata/BlobRouter-definition.json`,
			<BlobRouter location="http://localhost:3080/github.com/gorilla/mux@master/.GoPackage/github.com/gorilla/mux/.def/Router" />
		);
	});

	it("should consult the RepoStore for the repo's default branch if the rev is empty", () => {
		Dispatcher.directDispatch(RepoStore, new RepoActions.FetchedRepo("myrepo", {DefaultBranch: "mybranch"}));
		[
			"http://localhost:3080/myrepo",
			"http://localhost:3080/myrepo/.tree/file.txt",
			"http://localhost:3080/myrepo/.GoPackage/u/.def/p",
		].forEach((url) => {
			let renderer = TestUtils.createRenderer();
			renderer.render(<BlobRouter location={url} />);
			expect(renderer._instance._instance.state.rev).to.be("mybranch");
		});
	});

	it("should handle DefActions.SelectDef and trigger WantDef when no def is in store", () => {
		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go",
			new DefActions.SelectDef("someURL"),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go"
		);
	});

	it("should handle DefActions.SelectDef and go to def when the def is in store", () => {
		Dispatcher.directDispatch(DefStore, new DefActions.DefFetched("someURL", {}));
		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go",
			new DefActions.SelectDef("someURL"),
			"http://localhost:3080/someURL"
		);
	});

	it("should handle DefActions.SelectDef and NOT go to def when the def is errored", () => {
		Dispatcher.directDispatch(DefStore, new DefActions.DefFetched("someURL", {Error: "x"}));
		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go",
			new DefActions.SelectDef("someURL"),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go"
		);
	});

	it("should ignore standalone DefActions.DefFetched actions for defs that are not its active def", () => {
		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go",
			new DefActions.DefFetched("someURL", {}),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go"
		);
	});

	it("should handle BlobActions.SelectLine", () => {
		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go",
			new BlobActions.SelectLine(null, null, null, 42),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L42"
		);

		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L20-60",
			new BlobActions.SelectLine(null, null, null, 42),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L42"
		);
	});

	it("should handle BlobActions.SelectLineRange", () => {
		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go",
			new BlobActions.SelectLineRange(null, null, null, 42),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L42"
		);

		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L20",
			new BlobActions.SelectLineRange(null, null, null, 42),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L20-42"
		);

		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L50",
			new BlobActions.SelectLineRange(null, null, null, 42),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L42-50"
		);
	});

	it("should handle GoTo", () => {
		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L42",
			new GoTo("/github.com/gorilla/mux@master/.GoPackage/github.com/gorilla/mux/.def/Router"),
			"http://localhost:3080/github.com/gorilla/mux@master/.GoPackage/github.com/gorilla/mux/.def/Router"
		);
	});
});

function testAction(uri, action, expectedURI) {
	let renderer = TestUtils.createRenderer();
	renderer.render(<BlobRouter location={uri} navigate={(newURI) => { uri = newURI; }} _isMounted={true} />);
	Dispatcher.directDispatch(renderer._instance._instance, action);
	expect(uri).to.be(expectedURI);
}
