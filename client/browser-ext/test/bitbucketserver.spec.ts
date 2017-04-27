import { expect } from "chai";
import "fetch-mock";
import * as jsdom from "jsdom";
import * as utils from "../app/utils";
import * as annotations from "../app/utils/annotations";
import * as bitbucket from "../app/utils/bitbucket";
import { CodeCell, BitbucketBrowseUrl, BitbucketMode, BitbucketUrl } from "../app/utils/types";
import { BitbucketBlobAnnotator } from "../app/components/BitbucketBlobAnnotator";

function setupDOM(url: string): (done: any) => void {
	return (done) => jsdom.env(url, (err, window) => {
		if (err) {
			done(new Error("Could not build DOM from URL. Is there a Bitbucket Server running at " + url + "?\n" + err));
		}
		global.window = window;
		global.document = window.document;
		global.navigator = window.navigator;
		global.Node = (window as any).Node;
		done();
	});
}

describe("Bitbucket Server DOM", () => {
	describe("blob view", () => {
		const url = "http://localhost:7990/projects/GORILLA/repos/mux/browse/mux.go";
		before(setupDOM(url));

		it("should parse state", () => {
			const bitbucketState = bitbucket.getBitbucketState(global.window.location);
			expect(bitbucketState).to.not.be.null;
			expect((bitbucketState as BitbucketBrowseUrl).repo).to.equal("mux");
			expect((bitbucketState as BitbucketBrowseUrl).path).to.equal("mux.go");
			expect((bitbucketState as BitbucketBrowseUrl).projectCode).to.equal("GORILLA");
			expect((bitbucketState as BitbucketBrowseUrl).rev).to.equal("master");
		});
	});
});
