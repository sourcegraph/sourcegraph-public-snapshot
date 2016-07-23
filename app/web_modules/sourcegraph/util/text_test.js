// @flow weak

import expect from "expect.js";
import {getSnippets, leftIndexOf} from "sourcegraph/util/text";

describe("getSnippets", () => {
	it("should ignore case and correctly handle padding when results are next to each other", () => {
		let o = getSnippets("(you|use)", "You should use Sourcegraph for everything.", 0, 1);
		expect(o).to.equal("You should use Sourcegraph...");
	});
	it("should corretly handle two exact matches that are right next to each other", () => {
		let o = getSnippets("(you|should)", "I think you should use Sourcegraph.", 0, 1);
		expect(o).to.equal("...think you should use...");
	});
	it("should correctly handle having one of the tokens be part of the initial string ", () => {
		let o = getSnippets("(you|use)", "You should use Sourcegraph for everything.", 1, 1);
		expect(o).to.equal("You should use Sourcegraph...");
	});
	it("should correctly handle having the tokens be separated by some distance", () => {
		let o = getSnippets("(think|for)", "I think you should use Sourcegraph for everything.", 1, 1);
		expect(o).to.equal("I think you...Sourcegraph for everything.");
	});
});

describe("leftIndexOf", () => {
	it("should correctly handle a search term that's present when reading forwards, but not backwards", () => {
		let o = leftIndexOf("bar", "ar", 1);
		expect(o).to.equal(-1);
	});
	it("should correctly handle a search term that's present when reading backwards, but not forwards", () => {
		let o = leftIndexOf("foobar", "oob", 3);
		expect(o).to.equal(1);
	});
});
