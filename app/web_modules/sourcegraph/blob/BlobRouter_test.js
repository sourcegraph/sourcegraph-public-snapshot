import autotest from "sourcegraph/util/autotest";
import expect from "expect.js";

import React from "react";
import TestUtils from "react-addons-test-utils";

import Dispatcher from "sourcegraph/Dispatcher";
import BlobRouter from "sourcegraph/blob/BlobRouter";
import * as BlobActions from "sourcegraph/blob/BlobActions";
import * as DefActions from "sourcegraph/def/DefActions";
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
			<BlobRouter location="http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L40-53" />
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

	it("should handle DefActions.SelectDef", () => {
		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go",
			new DefActions.SelectDef("someURL"),
			"http://localhost:3080/someURL"
		);
	});

	it("should handle BlobActions.SelectLine", () => {
		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go",
			new BlobActions.SelectLine(42),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L42"
		);

		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L20-60",
			new BlobActions.SelectLine(42),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L42"
		);
	});

	it("should handle BlobActions.SelectLineRange", () => {
		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go",
			new BlobActions.SelectLineRange(42),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L42"
		);

		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L20",
			new BlobActions.SelectLineRange(42),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L20-42"
		);

		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L50",
			new BlobActions.SelectLineRange(42),
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
	renderer.render(<BlobRouter location={uri} navigate={(newURI) => { uri = newURI; }} />);
	Dispatcher.directDispatch(renderer._instance._instance, action);
	expect(uri).to.be(expectedURI);
}
