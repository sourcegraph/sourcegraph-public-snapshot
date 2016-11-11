import {expect} from "chai";
import {Nightmare} from "nightmare";

const nightmare = new (require("nightmare"))({show: true}) as Nightmare<any>;

describe("GitHub DOM", () => {
	describe("blob view", () => {
		it("should have a single file", () => nightmare
			.goto("https://github.com/gorilla/mux/blob/757bef944d0f21880861c2dd9c871ca543023cba/mux.go")
			.evaluate(() => Array.from(document.getElementsByClassName("file")))
			.end()
			.then((elems) => expect(elems).to.have.length(1))
		);
	});
});
