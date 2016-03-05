import autotest from "sourcegraph/util/autotest";
import expect from "expect.js";

import React from "react";
import TestUtils from "react-addons-test-utils";

import Dispatcher from "sourcegraph/Dispatcher";
import CodeFileRouter from "sourcegraph/code/CodeFileRouter";
import * as CodeActions from "sourcegraph/code/CodeActions";
import * as DefActions from "sourcegraph/def/DefActions";
import {GoTo} from "sourcegraph/util/hotLink";

import testdataFile from "sourcegraph/code/testdata/CodeFileRouter-file.json";
import testdataDotfile from "sourcegraph/code/testdata/CodeFileRouter-dotfile.json";
import testdataLineSelection from "sourcegraph/code/testdata/CodeFileRouter-lineSelection.json";
import testdataDefSelection from "sourcegraph/code/testdata/CodeFileRouter-defSelection.json";
import testdataDefinition from "sourcegraph/code/testdata/CodeFileRouter-definition.json";

describe("CodeFileRouter", () => {
	it("should handle file URLs", () => {
		autotest(testdataFile, `${__dirname}/testdata/CodeFileRouter-file.json`,
			<CodeFileRouter location="http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go" />
		);
	});

	it("should handle dotfile URLs", () => {
		autotest(testdataDotfile, `${__dirname}/testdata/CodeFileRouter-dotfile.json`,
			<CodeFileRouter location="http://localhost:3080/github.com/gorilla/mux@master/.tree/.travis.yml" />
		);
	});

	it("should handle line selection URLs", () => {
		autotest(testdataLineSelection, `${__dirname}/testdata/CodeFileRouter-lineSelection.json`,
			<CodeFileRouter location="http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L40-53" />
		);
	});

	it("should handle definition selection URLs", () => {
		autotest(testdataDefSelection, `${__dirname}/testdata/CodeFileRouter-defSelection.json`,
			<CodeFileRouter location="http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#def-someDef" />
		);
	});

	it("should handle definition URLs", () => {
		autotest(testdataDefinition, `${__dirname}/testdata/CodeFileRouter-definition.json`,
			<CodeFileRouter location="http://localhost:3080/github.com/gorilla/mux@master/.GoPackage/github.com/gorilla/mux/.def/Router" />
		);
	});

	it("should handle DefActions.SelectDef", () => {
		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go",
			new DefActions.SelectDef("someURL"),
			"http://localhost:3080/someURL"
		);
	});

	it("should handle CodeActions.SelectLine", () => {
		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go",
			new CodeActions.SelectLine(42),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L42"
		);

		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L20-60",
			new CodeActions.SelectLine(42),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L42"
		);
	});

	it("should handle CodeActions.SelectLineRange", () => {
		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go",
			new CodeActions.SelectLineRange(42),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L42"
		);

		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L20",
			new CodeActions.SelectLineRange(42),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L20-42"
		);

		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go#L50",
			new CodeActions.SelectLineRange(42),
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
	renderer.render(<CodeFileRouter location={uri} navigate={(newURI) => { uri = newURI; }} />);
	Dispatcher.directDispatch(renderer._instance._instance, action);
	expect(uri).to.be(expectedURI);
}
