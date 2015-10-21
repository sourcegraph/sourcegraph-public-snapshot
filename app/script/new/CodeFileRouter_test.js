import shallowRender from "./util/shallowRender";
import expect from "expect.js";

import React from "react";

import Dispatcher from "./Dispatcher";
import CodeFileRouter from "./CodeFileRouter";
import CodeFileContainer from "./CodeFileContainer";
import * as DefActions from "./DefActions";

describe("CodeFileRouter", () => {
	it("should handle file URLs", () => {
		shallowRender(
			<CodeFileRouter location="http://localhost:3000/github.com/gorilla/mux@master/.tree/mux.go" />
		).compare(
			<CodeFileContainer
				repo="github.com/gorilla/mux"
				rev="master"
				tree="mux.go"
				startLine={null}
				endLine={null}
				selectedDef={null}
				def={null} />
		);
	});

	it("should handle line selection URLs", () => {
		shallowRender(
			<CodeFileRouter location="http://localhost:3000/github.com/gorilla/mux@master/.tree/mux.go?startline=40&endline=53" />
		).compare(
			<CodeFileContainer
				repo="github.com/gorilla/mux"
				rev="master"
				tree="mux.go"
				startLine={40}
				endLine={53}
				selectedDef={null}
				def={null} />
		);
	});

	it("should handle definition selection URLs", () => {
		shallowRender(
			<CodeFileRouter location="http://localhost:3000/github.com/gorilla/mux@master/.tree/mux.go?seldef=someDef" />
		).compare(
			<CodeFileContainer
				repo="github.com/gorilla/mux"
				rev="master"
				tree="mux.go"
				startLine={null}
				endLine={null}
				selectedDef={"someDef"}
				def={null} />
		);
	});

	it("should handle definition URLs", () => {
		shallowRender(
			<CodeFileRouter location="http://localhost:3000/github.com/gorilla/mux@master/.GoPackage/github.com/gorilla/mux/.def/Router" />
		).compare(
			<CodeFileContainer
				repo="github.com/gorilla/mux"
				rev="master"
				unitType="GoPackage"
				unit="github.com/gorilla/mux"
				def="Router"
				example={null} />
		);
	});

	it("should handle example URLs", () => {
		shallowRender(
			<CodeFileRouter location="http://localhost:3000/github.com/gorilla/mux@master/.GoPackage/github.com/gorilla/mux/.def/Router/.examples/4" />
		).compare(
			<CodeFileContainer
				repo="github.com/gorilla/mux"
				rev="master"
				unitType="GoPackage"
				unit="github.com/gorilla/mux"
				def="Router"
				example={4} />
		);
	});

	it("should handle DefActions.SelectDef", () => {
		let uri = "http://localhost:3000/github.com/gorilla/mux@master/.tree/mux.go";
		let r = new CodeFileRouter({location: uri, navigate(newURI) { uri = newURI; }});
		Dispatcher.directDispatch(r, new DefActions.SelectDef("someURL"));
		expect(uri).to.be("http://localhost:3000/github.com/gorilla/mux@master/.tree/mux.go?seldef=someURL");
	});
});
