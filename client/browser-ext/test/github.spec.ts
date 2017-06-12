import { expect } from "chai";
import "fetch-mock";
import * as jsdom from "jsdom";
import * as utils from "../app/utils";
import * as annotations from "../app/utils/annotations";
import * as github from "../app/utils/github";
import { CodeCell, GitHubBlobUrl } from "../app/utils/types";

function setupDOM(url: string): (done: any) => void {
	return (done) => jsdom.env(url, (err, window) => {
		if (err) {
			done(err);
		}
		global.window = window;
		global.document = window.document;
		global.navigator = window.navigator;
		global.Node = (window as any).Node;
		done();
	});
}

describe("GitHub DOM", () => {
	describe("blob view", () => {
		const url = "https://github.com/gorilla/mux/blob/master/mux.go";
		before(setupDOM(url));

		it("should parse branch name from button", () => {
			const gitHubState = github.getGitHubState(global.window.location.href);
			expect(gitHubState).to.not.be.null;
			expect((gitHubState as GitHubBlobUrl).rev).to.equal("master");
		});

	});
});
