import shallowRender from "./util/shallowRender";
import expect from "expect.js";

import React from "react";

import Dispatcher from "./Dispatcher";
import CodeFileRouter from "./CodeFileRouter";
import CodeFileContainer from "./CodeFileContainer";
import * as DefActions from "./DefActions";

describe("CodeFileRouter", () => {
	it("should handle file URLs", () => {
		global.window = {location: {href: "http://localhost:3000/github.com/gorilla/mux@master/.tree/mux.go"}};
		shallowRender(
			<CodeFileRouter />
		).compare(
			<CodeFileContainer
				repo="github.com/gorilla/mux"
				rev="master"
				tree="mux.go"
				startLine={undefined}
				endLine={undefined}
				selectedDef={undefined} />
		);
	});

	it("should handle line selection URLs", () => {
		global.window = {location: {href: "http://localhost:3000/github.com/gorilla/mux@master/.tree/mux.go?startline=40&endline=53"}};
		shallowRender(
			<CodeFileRouter />
		).compare(
			<CodeFileContainer
				repo="github.com/gorilla/mux"
				rev="master"
				tree="mux.go"
				startLine={40}
				endLine={53}
				selectedDef={undefined} />
		);
	});

	it("should handle definition selection URLs", () => {
		global.window = {location: {href: "http://localhost:3000/github.com/gorilla/mux@master/.tree/mux.go?seldef=someDef"}};
		shallowRender(
			<CodeFileRouter />
		).compare(
			<CodeFileContainer
				repo="github.com/gorilla/mux"
				rev="master"
				tree="mux.go"
				startLine={undefined}
				endLine={undefined}
				selectedDef={"someDef"} />
		);
	});

	it("should handle definition URLs", () => {
		global.window = {location: {href: "http://localhost:3000/github.com/gorilla/mux@master/.GoPackage/github.com/gorilla/mux/.def/Router"}};
		shallowRender(
			<CodeFileRouter />
		).compare(
			<CodeFileContainer
				repo="github.com/gorilla/mux"
				rev="master"
				unitType="GoPackage"
				unit="github.com/gorilla/mux"
				def="Router"
				example={undefined} />
		);
	});

	it("should handle example URLs", () => {
		global.window = {location: {href: "http://localhost:3000/github.com/gorilla/mux@master/.GoPackage/github.com/gorilla/mux/.def/Router/.examples/4"}};
		shallowRender(
			<CodeFileRouter />
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
		let r = new CodeFileRouter();
		r._navigate = function(path, query) {
			expect(path).to.be(null);
			expect(query).to.eql({seldef: "someURL"});
		};
		Dispatcher.directDispatch(r, new DefActions.SelectDef("someURL"));
	});
});
