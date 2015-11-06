import autotest from "./util/autotest";
import expect from "expect.js";

import React from "react";

import Dispatcher from "./Dispatcher";
import CodeFileRouter from "./CodeFileRouter";
import * as DefActions from "./DefActions";

import testdataFile from "./testdata/CodeFileRouter-file.json";
import testdataLineSelection from "./testdata/CodeFileRouter-lineSelection.json";
import testdataDefSelection from "./testdata/CodeFileRouter-defSelection.json";
import testdataDefinition from "./testdata/CodeFileRouter-definition.json";
import testdataExample from "./testdata/CodeFileRouter-example.json";

describe("CodeFileRouter", () => {
	it("should handle file URLs", () => {
		autotest(testdataFile, `${__dirname}/testdata/CodeFileRouter-file.json`,
			<CodeFileRouter location="http://localhost:3000/github.com/gorilla/mux@master/.tree/mux.go" />
		);
	});

	it("should handle line selection URLs", () => {
		autotest(testdataLineSelection, `${__dirname}/testdata/CodeFileRouter-lineSelection.json`,
			<CodeFileRouter location="http://localhost:3000/github.com/gorilla/mux@master/.tree/mux.go?startline=40&endline=53" />
		);
	});

	it("should handle definition selection URLs", () => {
		autotest(testdataDefSelection, `${__dirname}/testdata/CodeFileRouter-defSelection.json`,
			<CodeFileRouter location="http://localhost:3000/github.com/gorilla/mux@master/.tree/mux.go?seldef=someDef" />
		);
	});

	it("should handle definition URLs", () => {
		autotest(testdataDefinition, `${__dirname}/testdata/CodeFileRouter-definition.json`,
			<CodeFileRouter location="http://localhost:3000/github.com/gorilla/mux@master/.GoPackage/github.com/gorilla/mux/.def/Router" />
		);
	});

	it("should handle example URLs", () => {
		autotest(testdataExample, `${__dirname}/testdata/CodeFileRouter-example.json`,
			<CodeFileRouter location="http://localhost:3000/github.com/gorilla/mux@master/.GoPackage/github.com/gorilla/mux/.def/Router/.examples/4" />
		);
	});

	it("should handle DefActions.SelectDef", () => {
		let uri = "http://localhost:3000/github.com/gorilla/mux@master/.tree/mux.go";
		let r = new CodeFileRouter({location: uri, navigate(newURI) { uri = newURI; }});
		Dispatcher.directDispatch(r, new DefActions.SelectDef("someURL"));
		expect(uri).to.be("http://localhost:3000/github.com/gorilla/mux@master/.tree/mux.go?seldef=someURL");
	});
});
