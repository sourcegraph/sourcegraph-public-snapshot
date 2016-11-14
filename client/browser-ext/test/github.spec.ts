import * as utils from "../app/utils";
import * as github from "../app/utils/github";
import {expect} from "chai";
import * as jsdom from "jsdom";

function setupDOM(url: string): (done: any) => void {
	return (done) => jsdom.env(url, (err, window) => {
		if (err) {
			done(err);
		}
		global.window = window;
		global.document = window.document;
		global.navigator = window.navigator;
		done();
	});
}

describe("GitHub DOM", () => {
	describe("blob view", () => {
		const url = "https://github.com/gorilla/mux/blob/757bef944d0f21880861c2dd9c871ca543023cba/mux.go";
		before(setupDOM(url));

		it("should have 1 file container", () => {
			expect(github.getFileContainers()).to.have.length(1);
		});

		it("should parse url", () => {
			const data = utils.parseURL(window.location);
			expect(data.user).to.eql("gorilla");
			expect(data.repo).to.eql("mux");
			expect(data.repoURI).to.eql("github.com/gorilla/mux");
			expect(data.rev).to.eql("757bef944d0f21880861c2dd9c871ca543023cba");
			expect(data.path).to.eql("mux.go");
			expect(data.isDelta).to.not.be.ok;
			expect(data.isCommit).to.not.be.ok;
			expect(data.isPullRequest).to.not.be.ok;
		});
	});

	describe("commit view", () => {
		const url = "https://github.com/gorilla/mux/commit/0a192a193177452756c362c20087ddafcf6829c4";
		before(setupDOM(url));

		it("should have 5 file containers", () => {
			expect(github.getFileContainers()).to.have.length(5);
		});

		it("should get file names of containers", () => {
			expect(Array.from(github.getFileContainers()).map(github.getDeltaFileName)).to.eql([
				"mux.go", "mux_test.go", "old_test.go", "regexp.go", "route.go",
			]);
		});

		it("should parse url", () => {
			const data = utils.parseURL(window.location);
			expect(data.user).to.eql("gorilla");
			expect(data.repo).to.eql("mux");
			expect(data.repoURI).to.eql("github.com/gorilla/mux");
			expect(data.rev).to.eql("0a192a193177452756c362c20087ddafcf6829c4");
			expect(data.path).to.not.be.ok;
			expect(data.isDelta).to.be.true;
			expect(data.isCommit).to.be.true;
			expect(data.isPullRequest).to.be.false;
		});
	});

	describe("PR conversation view w/ snippets", () => {
		const url = "https://github.com/gorilla/mux/pull/190";
		before(setupDOM(url));

		it("should have 3 file containers", () => {
			expect(github.getFileContainers()).to.have.length(3);
		});

		it("should get file names of containers", () => {
			expect(Array.from(github.getFileContainers()).map(github.getDeltaFileName)).to.eql([
				"mux_test.go", "mux.go", "mux.go",
			]);
		});

		it("should parse url", () => {
			const data = utils.parseURL(window.location);
			expect(data.user).to.eql("gorilla");
			expect(data.repo).to.eql("mux");
			expect(data.repoURI).to.eql("github.com/gorilla/mux");
			expect(data.rev).to.not.be.ok;
			expect(data.path).to.not.be.ok;
			expect(data.isDelta).to.be.true;
			expect(data.isCommit).to.be.false;
			expect(data.isPullRequest).to.be.true;
		});
	});

	describe("PR unified diff", () => {
		const url = "https://github.com/gorilla/mux/pull/190/files?diff=unified";
		before(setupDOM(url));

		it("should have 5 file containers", () => {
			expect(github.getFileContainers()).to.have.length(5);
		});

		it("should get file names of containers", () => {
			expect(Array.from(github.getFileContainers()).map(github.getDeltaFileName)).to.eql([
				"mux.go", "mux_test.go", "old_test.go", "regexp.go", "route.go",
			]);
		});

		it("should parse url", () => {
			const data = utils.parseURL(window.location);
			expect(data.user).to.eql("gorilla");
			expect(data.repo).to.eql("mux");
			expect(data.repoURI).to.eql("github.com/gorilla/mux");
			expect(data.rev).to.not.be.ok;
			expect(data.path).to.not.be.ok;
			expect(data.isDelta).to.be.true;
			expect(data.isCommit).to.be.false;
			expect(data.isPullRequest).to.be.true;
		});
	});

	describe("PR split diff", () => {
		const url = "https://github.com/gorilla/mux/pull/190/files?diff=split";
		before(setupDOM(url));

		it("should have 5 file containers", () => {
			expect(github.getFileContainers()).to.have.length(5);
		});

		it("should get file names of containers", () => {
			expect(Array.from(github.getFileContainers()).map(github.getDeltaFileName)).to.eql([
				"mux.go", "mux_test.go", "old_test.go", "regexp.go", "route.go",
			]);
		});

		it("should parse url", () => {
			const data = utils.parseURL(window.location);
			expect(data.user).to.eql("gorilla");
			expect(data.repo).to.eql("mux");
			expect(data.repoURI).to.eql("github.com/gorilla/mux");
			expect(data.rev).to.not.be.ok;
			expect(data.path).to.not.be.ok;
			expect(data.isDelta).to.be.true;
			expect(data.isCommit).to.be.false;
			expect(data.isPullRequest).to.be.true;
		});
	});
});
