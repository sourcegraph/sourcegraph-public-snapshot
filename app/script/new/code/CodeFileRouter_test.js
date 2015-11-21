import autotest from "../util/autotest";
import expect from "expect.js";

import React from "react";
import TestUtils from "react-addons-test-utils";

import Dispatcher from "../Dispatcher";
import CodeFileRouter from "./CodeFileRouter";
import * as CodeActions from "./CodeActions";
import * as DefActions from "../def/DefActions";
import {GoTo} from "../util/hotLink";

import testdataFile from "./testdata/CodeFileRouter-file.json";
import testdataDotfile from "./testdata/CodeFileRouter-dotfile.json";
import testdataLineSelection from "./testdata/CodeFileRouter-lineSelection.json";
import testdataDefSelection from "./testdata/CodeFileRouter-defSelection.json";
import testdataDefinition from "./testdata/CodeFileRouter-definition.json";

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
			<CodeFileRouter location="http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go?startline=40&endline=53" />
		);
	});

	it("should handle definition selection URLs", () => {
		autotest(testdataDefSelection, `${__dirname}/testdata/CodeFileRouter-defSelection.json`,
			<CodeFileRouter location="http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go?seldef=someDef" />
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
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go?seldef=someURL"
		);
	});

	it("should handle CodeActions.SelectLine", () => {
		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go",
			new CodeActions.SelectLine(42),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go?startline=42&endline=42"
		);

		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go?startline=20&endline=60",
			new CodeActions.SelectLine(42),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go?startline=42&endline=42"
		);
	});

	it("should handle CodeActions.SelectRange", () => {
		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go",
			new CodeActions.SelectRange(42),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go?startline=42&endline=42"
		);

		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go?startline=20&endline=20",
			new CodeActions.SelectRange(42),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go?startline=20&endline=42"
		);

		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go?startline=50&endline=50",
			new CodeActions.SelectRange(42),
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go?startline=42&endline=50"
		);
	});

	it("should handle GoTo", () => {
		testAction(
			"http://localhost:3080/github.com/gorilla/mux@master/.tree/mux.go?startline=42&endline=42&seldef=someURL",
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
