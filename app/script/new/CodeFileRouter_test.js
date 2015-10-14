import sandbox from "../testSandbox";

import React from "react";

import CodeFileRouter from "./CodeFileRouter";
import CodeFileController from "./CodeFileController";

describe("CodeFileRouter", () => {
	it("should handle file URLs", () => {
		window.location.href = "http://localhost:3000/github.com/gorilla/mux@master/.tree/mux.go";
		sandbox.renderAndExpect(<CodeFileRouter />).to.eql(
			<CodeFileController
				repo="github.com/gorilla/mux"
				rev="master"
				tree="mux.go"
				startline={undefined}
				endline={undefined}
				token={undefined} />
		);
	});

	it("should handle selection URLs", () => {
		window.location.href = "https://sourcegraph.com/github.com/gorilla/mux@master/.tree/mux.go?startline=40&endline=53";
		sandbox.renderAndExpect(<CodeFileRouter />).to.eql(
			<CodeFileController
				repo="github.com/gorilla/mux"
				rev="master"
				tree="mux.go"
				startline={40}
				endline={53}
				token={undefined} />
		);
	});

	it("should handle token URLs", () => {
		window.location.href = "http://localhost:3000/github.com/gorilla/mux@master/.tree/mux.go/.token/42";
		sandbox.renderAndExpect(<CodeFileRouter />).to.eql(
			<CodeFileController
				repo="github.com/gorilla/mux"
				rev="master"
				tree="mux.go"
				startline={undefined}
				endline={undefined}
				token={42} />
		);
	});

	it("should handle definition URLs", () => {
		window.location.href = "http://localhost:3000/github.com/gorilla/mux@master/.GoPackage/github.com/gorilla/mux/.def/Router";
		sandbox.renderAndExpect(<CodeFileRouter />).to.eql(
			<CodeFileController
				repo="github.com/gorilla/mux"
				rev="master"
				unitType="GoPackage"
				unit="github.com/gorilla/mux"
				def="Router"
				example={undefined} />
		);
	});

	it("should handle example URLs", () => {
		window.location.href = "http://localhost:3000/github.com/gorilla/mux@master/.GoPackage/github.com/gorilla/mux/.def/Router/.examples/4";
		sandbox.renderAndExpect(<CodeFileRouter />).to.eql(
			<CodeFileController
				repo="github.com/gorilla/mux"
				rev="master"
				unitType="GoPackage"
				unit="github.com/gorilla/mux"
				def="Router"
				example={4} />
		);
	});
});
