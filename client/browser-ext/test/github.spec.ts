import * as utils from "../app/utils";
import * as annotations from "../app/utils/annotations";
import * as github from "../app/utils/github";
import { expect } from "chai";
import "fetch-mock";
import * as jsdom from "jsdom";

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
		const url = "https://github.com/gorilla/mux/blob/757bef944d0f21880861c2dd9c871ca543023cba/mux.go";
		before(setupDOM(url));

		it("should have 1 file container", () => {
			expect(github.getFileContainers()).to.have.length(1);
		});

		it("should get blob from file container", () => {
			Array.from(github.getFileContainers()).forEach((file) => expect(github.tryGetBlobElement(file)).to.be.ok);
		});

		it("should create blob annotator mount", () => {
			Array.from(github.getFileContainers()).forEach((file) => expect(github.createBlobAnnotatorMount(file)).to.be.ok);
			expect(document.getElementsByClassName("sourcegraph-app-annotator")).to.have.length(1);
		});

		it("should not be private repo", () => {
			expect(github.isPrivateRepo()).to.be.false;
		});

		it("should not be split diff", () => {
			expect(github.isSplitDiff()).to.be.false;
		});

		it("should not parse base/head rev", () => {
			expect(github.getDeltaRevs()).to.not.be.ok;
		});

		it("should not parse delta info", () => {
			expect(github.getDeltaInfo()).to.not.be.ok;
		});

		it("should not register diff expand handlers", () => {
			github.registerExpandDiffClickHandler(() => ({}));
			expect(document.getElementsByClassName("sg-diff-expander")).to.have.length(0);
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

		describe("annotations", () => {
			let _codeCells: github.CodeCell[];
			before(() => {
				const file = github.getFileContainers()[0];
				_codeCells = github.getCodeCellsForAnnotation(github.getCodeTable(file), { isDelta: false, isSplitDiff: false, isBase: false });
			});

			it("should get code cells", () => {
				expect(_codeCells).to.have.length(542);
			});

			it("should convert element node", () => {
				// convert line 5, "package mux"
				const convertedNode = annotations.convertElementNode(_codeCells[4].cell, 1, 5, false);
				expect(convertedNode.bytesConsumed).to.eql(11);
				expect(convertedNode.resultNode.textContent).to.eql("package mux");
				expect(convertedNode.resultNode.innerHTML).to.eql(`<span data-byteoffset="1" class="pl-k"><span id="text-node-wrapper-5-1"><span id="text-node-5-1" data-byteoffset="1">package</span></span></span><span id="text-node-wrapper-5-8"><span id="text-node-5-8" data-byteoffset="8"> </span><span id="text-node-5-9" data-byteoffset="9">mux</span></span>`);
			});

			it("should convert text node", () => {
				// convert line 18, " &Router..."
				const convertedNode = annotations.convertTextNode(_codeCells[17].cell.childNodes[2], 8, 18, false);
				expect(convertedNode.bytesConsumed).to.eql(22);
				expect(convertedNode.resultNode.textContent).to.eql(" &Router{namedRoutes: ");
				expect(convertedNode.resultNode.innerHTML).to.eql(`<span id="text-node-18-8" data-byteoffset="8"> </span><span id="text-node-18-9" data-byteoffset="9">&amp;</span><span id="text-node-18-10" data-byteoffset="10">Router</span><span id="text-node-18-16" data-byteoffset="16">{</span><span id="text-node-18-17" data-byteoffset="17">namedRoutes</span><span id="text-node-18-28" data-byteoffset="28">:</span><span id="text-node-18-29" data-byteoffset="29"> </span>`);
			});

			it("should convert node (stress test)", () => {
				// convert complete line 18 code cell
				const convertedNode = annotations.convertElementNode(_codeCells[17].cell, 1, 18, false);
				expect(convertedNode.bytesConsumed).to.eql(73);
				expect(convertedNode.resultNode.textContent).to.eql("\treturn &Router{namedRoutes: make(map[string]*Route), KeepContext: false}");
				expect(convertedNode.resultNode.innerHTML).to.eql(`<span id="text-node-wrapper-18-1"><span id="text-node-18-1" data-byteoffset="1">\t</span></span><span data-byteoffset="2" class="pl-k"><span id="text-node-wrapper-18-2"><span id="text-node-18-2" data-byteoffset="2">return</span></span></span><span id="text-node-wrapper-18-8"><span id="text-node-18-8" data-byteoffset="8"> </span><span id="text-node-18-9" data-byteoffset="9">&amp;</span><span id="text-node-18-10" data-byteoffset="10">Router</span><span id="text-node-18-16" data-byteoffset="16">{</span><span id="text-node-18-17" data-byteoffset="17">namedRoutes</span><span id="text-node-18-28" data-byteoffset="28">:</span><span id="text-node-18-29" data-byteoffset="29"> </span></span><span data-byteoffset="30" class="pl-c1"><span id="text-node-wrapper-18-30"><span id="text-node-18-30" data-byteoffset="30">make</span></span></span><span id="text-node-wrapper-18-34"><span id="text-node-18-34" data-byteoffset="34">(</span></span><span data-byteoffset="35" class="pl-k"><span id="text-node-wrapper-18-35"><span id="text-node-18-35" data-byteoffset="35">map</span></span></span><span id="text-node-wrapper-18-38"><span id="text-node-18-38" data-byteoffset="38">[</span></span><span data-byteoffset="39" class="pl-k"><span id="text-node-wrapper-18-39"><span id="text-node-18-39" data-byteoffset="39">string</span></span></span><span id="text-node-wrapper-18-45"><span id="text-node-18-45" data-byteoffset="45">]</span><span id="text-node-18-46" data-byteoffset="46">*</span><span id="text-node-18-47" data-byteoffset="47">Route</span><span id="text-node-18-52" data-byteoffset="52">)</span><span id="text-node-18-53" data-byteoffset="53">,</span><span id="text-node-18-54" data-byteoffset="54"> </span><span id="text-node-18-55" data-byteoffset="55">KeepContext</span><span id="text-node-18-66" data-byteoffset="66">:</span><span id="text-node-18-67" data-byteoffset="67"> </span></span><span data-byteoffset="68" class="pl-c1"><span id="text-node-wrapper-18-68"><span id="text-node-18-68" data-byteoffset="68">false</span></span></span><span id="text-node-wrapper-18-73"><span id="text-node-18-73" data-byteoffset="73">}</span></span>`);
			});
		});
	});

	describe("commit view", () => {
		const url = "https://github.com/gorilla/mux/commit/0a192a193177452756c362c20087ddafcf6829c4";
		before(setupDOM(url));

		it("should have 5 file containers", () => {
			expect(github.getFileContainers()).to.have.length(5);
		});

		it("should get blob from file containers", () => {
			Array.from(github.getFileContainers()).forEach((file) => expect(github.tryGetBlobElement(file)).to.be.ok);
		});

		it("should create blob annotator mounts", () => {
			Array.from(github.getFileContainers()).forEach((file) => expect(github.createBlobAnnotatorMount(file)).to.be.ok);
			expect(document.getElementsByClassName("sourcegraph-app-annotator")).to.have.length(5);
		});

		it("should get file names of containers", () => {
			expect(Array.from(github.getFileContainers()).map(github.getDeltaFileName)).to.eql([
				"mux.go", "mux_test.go", "old_test.go", "regexp.go", "route.go",
			]);
		});

		it("should not be private repo", () => {
			expect(github.isPrivateRepo()).to.be.false;
		});

		it("should not be split diff", () => {
			expect(github.isSplitDiff()).to.be.false;
		});

		it("should parse base/head rev", () => {
			const deltaRevs = github.getDeltaRevs();
			expect(deltaRevs).to.have.property("base", "0b13a922203ebdbfd236c818efcd5ed46097d690");
			expect(deltaRevs).to.have.property("head", "0a192a193177452756c362c20087ddafcf6829c4");
		});

		it("should parse deltaInfo", () => {
			const deltaInfo = github.getDeltaInfo();
			expect(deltaInfo).to.have.property("baseBranch", "master");
			expect(deltaInfo).to.have.property("headBranch", "master");
			expect(deltaInfo).to.have.property("baseURI", "github.com/gorilla/mux");
			expect(deltaInfo).to.have.property("headURI", "github.com/gorilla/mux");
		});

		it("should register diff expand handlers", () => {
			github.registerExpandDiffClickHandler(() => ({}));
			expect(document.getElementsByClassName("sg-diff-expander")).to.have.length(25);
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

		describe("annotations", () => {
			it("should get code cells", () => {
				const files = Array.from(github.getFileContainers());
				const baseCells = files.map((file) => github.getCodeCellsForAnnotation(github.getCodeTable(file), { isDelta: true, isSplitDiff: false, isBase: true }));
				const headCells = files.map((file) => github.getCodeCellsForAnnotation(github.getCodeTable(file), { isDelta: true, isSplitDiff: false, isBase: false }));

				baseCells.forEach((group) => group.forEach((cell) => expect(cell.isDeletion).to.be.true));
				headCells.forEach((group) => group.forEach((cell) => expect(cell.isDeletion).to.be.false));

				expect(baseCells[0]).to.have.length(2);
				expect(baseCells[1]).to.have.length(10);
				expect(baseCells[2]).to.have.length(1);
				expect(baseCells[3]).to.have.length(11);
				expect(baseCells[4]).to.have.length(1);

				expect(headCells[0]).to.have.length(46);
				expect(headCells[1]).to.have.length(92);
				expect(headCells[2]).to.have.length(7);
				expect(headCells[3]).to.have.length(51);
				expect(headCells[4]).to.have.length(15);

				expect(headCells[0].filter((cell) => cell.isAddition)).to.have.length(22);
				expect(headCells[1].filter((cell) => cell.isAddition)).to.have.length(44);
				expect(headCells[2].filter((cell) => cell.isAddition)).to.have.length(1);
				expect(headCells[3].filter((cell) => cell.isAddition)).to.have.length(21);
				expect(headCells[4].filter((cell) => cell.isAddition)).to.have.length(3);
			});
		});
	});

	describe("PR unified diff", () => {
		const url = "https://github.com/gorilla/mux/pull/190/files?diff=unified";
		before(setupDOM(url));

		it("should have 5 file containers", () => {
			expect(github.getFileContainers()).to.have.length(5);
		});

		it("should get blob from file containers", () => {
			Array.from(github.getFileContainers()).forEach((file) => expect(github.tryGetBlobElement(file)).to.be.ok);
		});

		it("should create blob annotator mounts", () => {
			Array.from(github.getFileContainers()).forEach((file) => expect(github.createBlobAnnotatorMount(file)).to.be.ok);
			expect(document.getElementsByClassName("sourcegraph-app-annotator")).to.have.length(5);
		});

		it("should get file names of containers", () => {
			expect(Array.from(github.getFileContainers()).map(github.getDeltaFileName)).to.eql([
				"mux.go", "mux_test.go", "old_test.go", "regexp.go", "route.go",
			]);
		});

		it("should not be private repo", () => {
			expect(github.isPrivateRepo()).to.be.false;
		});

		it("should not be split diff", () => {
			expect(github.isSplitDiff()).to.be.false;
		});

		it("should parse base/head rev", () => {
			const deltaRevs = github.getDeltaRevs();
			expect(deltaRevs).to.have.property("base", "0b13a922203ebdbfd236c818efcd5ed46097d690");
			expect(deltaRevs).to.have.property("head", "0f3e78049d1980ca91e916b5ff90f711310c651f");
		});

		it("should parse deltaInfo", () => {
			const deltaInfo = github.getDeltaInfo();
			expect(deltaInfo).to.have.property("baseBranch", "master");
			expect(deltaInfo).to.have.property("headBranch", "use-encoded-path-option");
			expect(deltaInfo).to.have.property("baseURI", "github.com/gorilla/mux");
			expect(deltaInfo).to.have.property("headURI", "github.com/kushmansingh/mux");
		});

		it("should register diff expand handlers", () => {
			github.registerExpandDiffClickHandler(() => ({}));
			expect(document.getElementsByClassName("sg-diff-expander")).to.have.length(25);
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

		describe("annotations", () => {

			it("should get code cells for each snippet", () => {
				const files = Array.from(github.getFileContainers());
				const baseCells = files.map((file) => github.getCodeCellsForAnnotation(github.getCodeTable(file), { isDelta: true, isSplitDiff: false, isBase: true }));
				const headCells = files.map((file) => github.getCodeCellsForAnnotation(github.getCodeTable(file), { isDelta: true, isSplitDiff: false, isBase: false }));

				baseCells.forEach((group) => group.forEach((cell) => expect(cell.isDeletion).to.be.true));
				headCells.forEach((group) => group.forEach((cell) => expect(cell.isDeletion).to.be.false));

				expect(baseCells[0]).to.have.length(2);
				expect(baseCells[1]).to.have.length(10);
				expect(baseCells[2]).to.have.length(1);
				expect(baseCells[3]).to.have.length(11);
				expect(baseCells[4]).to.have.length(1);

				expect(headCells[0]).to.have.length(46);
				expect(headCells[1]).to.have.length(92);
				expect(headCells[2]).to.have.length(7);
				expect(headCells[3]).to.have.length(51);
				expect(headCells[4]).to.have.length(15);

				expect(headCells[0].filter((cell) => cell.isAddition)).to.have.length(22);
				expect(headCells[1].filter((cell) => cell.isAddition)).to.have.length(44);
				expect(headCells[2].filter((cell) => cell.isAddition)).to.have.length(1);
				expect(headCells[3].filter((cell) => cell.isAddition)).to.have.length(21);
				expect(headCells[4].filter((cell) => cell.isAddition)).to.have.length(3);
			});

			function getPRCells(isBase: boolean): github.CodeCell[] {
				const file = github.getFileContainers()[0];
				const codeCells = github.getCodeCellsForAnnotation(github.getCodeTable(file), { isDelta: true, isSplitDiff: false, isBase });
				return codeCells;
			}

			it("should convert deletion node (stress test)", () => {
				// first red line of mux.go
				const codeCells = getPRCells(true);
				const convertedNode = annotations.convertElementNode(codeCells[0].cell, 1, 80, false);
				expect(convertedNode.bytesConsumed).to.eql(23);
				expect(convertedNode.resultNode.textContent).to.eql("-\t\tpath := getPath(req)");
				expect(convertedNode.resultNode.innerHTML).to.eql(`<span id="text-node-wrapper-80-1"><span id="text-node-80-1" data-byteoffset="1">-</span><span id="text-node-80-2" data-byteoffset="2">\t</span><span id="text-node-80-3" data-byteoffset="3">\t</span></span><span data-byteoffset="4" class="pl-smi"><span id="text-node-wrapper-80-4"><span id="text-node-80-4" data-byteoffset="4">path</span></span></span><span id="text-node-wrapper-80-8"><span id="text-node-80-8" data-byteoffset="8"> </span></span><span data-byteoffset="9" class="pl-k"><span id="text-node-wrapper-80-9"><span id="text-node-80-9" data-byteoffset="9">:</span><span id="text-node-80-10" data-byteoffset="10">=</span></span></span><span id="text-node-wrapper-80-11"><span id="text-node-80-11" data-byteoffset="11"> </span></span><span data-byteoffset="12" class="pl-c1"><span id="text-node-wrapper-80-12"><span id="text-node-80-12" data-byteoffset="12">getPath</span></span></span><span id="text-node-wrapper-80-19"><span id="text-node-80-19" data-byteoffset="19">(</span><span id="text-node-80-20" data-byteoffset="20">req</span><span id="text-node-80-23" data-byteoffset="23">)</span></span>`);
			});

			it("should convert addition node (stress test)", () => {
				// second green line of mux.go
				const codeCells = getPRCells(false);
				const convertedNode = annotations.convertElementNode(codeCells[4].cell, 1, 57, false);
				expect(convertedNode.bytesConsumed).to.eql(21);
				expect(convertedNode.resultNode.textContent).to.eql("+\tuseEncodedPath bool");
				expect(convertedNode.resultNode.innerHTML).to.eql(`<span id="text-node-wrapper-57-1"><span id="text-node-57-1" data-byteoffset="1">+</span><span id="text-node-57-2" data-byteoffset="2">\t</span><span id="text-node-57-3" data-byteoffset="3">useEncodedPath</span><span id="text-node-57-17" data-byteoffset="17"> </span></span><span data-byteoffset="18" class="pl-k"><span id="text-node-wrapper-57-18"><span id="text-node-57-18" data-byteoffset="18">bool</span></span></span>`);
			});
		});
	});

	describe("PR split diff", () => {
		const url = "https://github.com/gorilla/mux/pull/190/files?diff=split";
		before(setupDOM(url));

		it("should have 5 file containers", () => {
			expect(github.getFileContainers()).to.have.length(5);
		});

		it("should get blob from file containers", () => {
			Array.from(github.getFileContainers()).forEach((file) => expect(github.tryGetBlobElement(file)).to.be.ok);
		});

		it("should create blob annotator mounts", () => {
			Array.from(github.getFileContainers()).forEach((file) => expect(github.createBlobAnnotatorMount(file)).to.be.ok);
			expect(document.getElementsByClassName("sourcegraph-app-annotator")).to.have.length(5);
		});

		it("should get file names of containers", () => {
			expect(Array.from(github.getFileContainers()).map(github.getDeltaFileName)).to.eql([
				"mux.go", "mux_test.go", "old_test.go", "regexp.go", "route.go",
			]);
		});

		it("should not be private repo", () => {
			expect(github.isPrivateRepo()).to.be.false;
		});

		it("should be split diff", () => {
			expect(github.isSplitDiff()).to.be.true;
		});

		it("should parse base/head rev", () => {
			const deltaRevs = github.getDeltaRevs();
			expect(deltaRevs).to.have.property("base", "0b13a922203ebdbfd236c818efcd5ed46097d690");
			expect(deltaRevs).to.have.property("head", "0f3e78049d1980ca91e916b5ff90f711310c651f");
		});

		it("should parse deltaInfo", () => {
			const deltaInfo = github.getDeltaInfo();
			expect(deltaInfo).to.have.property("baseBranch", "master");
			expect(deltaInfo).to.have.property("headBranch", "use-encoded-path-option");
			expect(deltaInfo).to.have.property("baseURI", "github.com/gorilla/mux");
			expect(deltaInfo).to.have.property("headURI", "github.com/kushmansingh/mux");
		});

		it("should register diff expand handlers", () => {
			github.registerExpandDiffClickHandler(() => ({}));
			expect(document.getElementsByClassName("sg-diff-expander")).to.have.length(25);
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

		describe("annotations", () => {
			it("should get code cells for each snippet", () => {
				const files = Array.from(github.getFileContainers());
				const baseCells = files.map((file) => github.getCodeCellsForAnnotation(github.getCodeTable(file), { isDelta: true, isSplitDiff: true, isBase: true }));
				const headCells = files.map((file) => github.getCodeCellsForAnnotation(github.getCodeTable(file), { isDelta: true, isSplitDiff: true, isBase: false }));

				baseCells.forEach((group) => group.forEach((cell) => expect(cell.isAddition).to.be.false));
				headCells.forEach((group) => group.forEach((cell) => expect(cell.isDeletion).to.be.false));

				expect(baseCells[0]).to.have.length(26);
				expect(baseCells[1]).to.have.length(58);
				expect(baseCells[2]).to.have.length(7);
				expect(baseCells[3]).to.have.length(41);
				expect(baseCells[4]).to.have.length(13);

				expect(baseCells[0].filter((cell) => cell.isDeletion)).to.have.length(2);
				expect(baseCells[1].filter((cell) => cell.isDeletion)).to.have.length(10);
				expect(baseCells[2].filter((cell) => cell.isDeletion)).to.have.length(1);
				expect(baseCells[3].filter((cell) => cell.isDeletion)).to.have.length(11);
				expect(baseCells[4].filter((cell) => cell.isDeletion)).to.have.length(1);

				expect(headCells[0]).to.have.length(46);
				expect(headCells[1]).to.have.length(92);
				expect(headCells[2]).to.have.length(7);
				expect(headCells[3]).to.have.length(51);
				expect(headCells[4]).to.have.length(15);

				expect(headCells[0].filter((cell) => cell.isAddition)).to.have.length(22);
				expect(headCells[1].filter((cell) => cell.isAddition)).to.have.length(44);
				expect(headCells[2].filter((cell) => cell.isAddition)).to.have.length(1);
				expect(headCells[3].filter((cell) => cell.isAddition)).to.have.length(21);
				expect(headCells[4].filter((cell) => cell.isAddition)).to.have.length(3);
			});
		});
	});
});
