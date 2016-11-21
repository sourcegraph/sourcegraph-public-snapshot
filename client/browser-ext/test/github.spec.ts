import * as utils from "../app/utils";
import * as annotations from "../app/utils/annotations";
import * as github from "../app/utils/github";
import * as testData from "./data/annotations";
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
			Array.from(github.getFileContainers()).forEach((file) => expect(github.getBlobElement(file)).to.be.ok);
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
			let _annsByStartByte: annotations.AnnotationsByByte;
			let _annsByEndByte: annotations.AnnotationsByByte;
			let _lineStartBytes: annotations.StartBytesByLine;
			let _codeCells: github.CodeCell[];
			before(() => {
				const {annsByStartByte, annsByEndByte} = annotations.indexAnnotations(testData.blobAnnotations.IncludedAnnotations.Annotations);
				_annsByStartByte = annsByStartByte;
				_annsByEndByte = annsByEndByte;
				_lineStartBytes = annotations.indexLineStartBytes(testData.blobAnnotations.IncludedAnnotations.LineStartBytes);

				const file = github.getFileContainers()[0];
				_codeCells = github.getCodeCellsForAnnotation(github.getCodeTable(file), {isDelta: false, isSplitDiff: false, isBase: false});
			});

			it("should index annotations", () => {
				expect(_annsByStartByte[173]).to.be.ok; // from line 5, "package mux" (the mux token)
				expect(_annsByStartByte[172]).to.not.be.ok;
			});

			it("should index line start bytes", () => {
				expect(_lineStartBytes[0]).to.be.not.ok;
				expect(_lineStartBytes[1]).to.eql(0);
				expect(_lineStartBytes[2]).to.eql(60);
				expect(_lineStartBytes[3]).to.eql(114);
				expect(_lineStartBytes[4]).to.eql(164);
				expect(_lineStartBytes[5]).to.eql(165);
			});

			it("should get code cells", () => {
				expect(_codeCells).to.have.length(542);
			});

			it("should convert element node", () => {
				// convert line 5, "package mux"
				const convertedNode = annotations.convertElementNode(_codeCells[4].cell, _annsByStartByte, _lineStartBytes[5], _lineStartBytes[5], false);
				expect(convertedNode.bytesConsumed).to.eql(11);
				expect(convertedNode.resultNode.textContent).to.eql("package mux");
				expect(convertedNode.resultNode.innerHTML).to.eql(`<span data-byteoffset="1" class="pl-k"><span id="text-node-wrapper-165"><span id="text-node-165-1" data-byteoffset="1">package</span></span></span><span id="text-node-wrapper-172"><span id="text-node-172-8" data-byteoffset="8"> </span><span id="text-node-172-9" data-byteoffset="9">mux</span></span>`);
			});

			it("should convert text node", () => {
				// convert line 18, " &Router..."
				const convertedNode = annotations.convertTextNode(_codeCells[17].cell.childNodes[2], _annsByStartByte, _lineStartBytes[18] + 7, _lineStartBytes[18], false);
				expect(convertedNode.bytesConsumed).to.eql(22);
				expect(convertedNode.resultNode.textContent).to.eql(" &Router{namedRoutes: ");
				expect(convertedNode.resultNode.innerHTML).to.eql(`<span id="text-node-326-8" data-byteoffset="8"> &amp;</span><span id="text-node-326-10" data-byteoffset="10">Router</span><span id="text-node-326-16" data-byteoffset="16">{</span><span id="text-node-326-17" data-byteoffset="17">namedRoutes</span><span id="text-node-326-28" data-byteoffset="28">: </span>`);
			});

			it("should convert node (stress test)", () => {
				// convert complete line 18 code cell
				const convertedNode = annotations.convertElementNode(_codeCells[17].cell, _annsByStartByte, _lineStartBytes[18], _lineStartBytes[18], false);
				expect(convertedNode.bytesConsumed).to.eql(73);
				expect(convertedNode.resultNode.textContent).to.eql("\treturn &Router{namedRoutes: make(map[string]*Route), KeepContext: false}");
				expect(convertedNode.resultNode.innerHTML).to.eql(`<span id="text-node-wrapper-319"><span id="text-node-319-1" data-byteoffset="1">\t</span></span><span data-byteoffset="2" class="pl-k"><span id="text-node-wrapper-320"><span id="text-node-320-2" data-byteoffset="2">return</span></span></span><span id="text-node-wrapper-326"><span id="text-node-326-8" data-byteoffset="8"> &amp;</span><span id="text-node-326-10" data-byteoffset="10">Router</span><span id="text-node-326-16" data-byteoffset="16">{</span><span id="text-node-326-17" data-byteoffset="17">namedRoutes</span><span id="text-node-326-28" data-byteoffset="28">: </span></span><span data-byteoffset="30" class="pl-c1"><span id="text-node-wrapper-348"><span id="text-node-348-30" data-byteoffset="30">make</span></span></span><span id="text-node-wrapper-352"><span id="text-node-352-34" data-byteoffset="34">(</span></span><span data-byteoffset="35" class="pl-k"><span id="text-node-wrapper-353"><span id="text-node-353-35" data-byteoffset="35">map</span></span></span><span id="text-node-wrapper-356"><span id="text-node-356-38" data-byteoffset="38">[</span></span><span data-byteoffset="39" class="pl-k"><span id="text-node-wrapper-357"><span id="text-node-357-39" data-byteoffset="39">string</span></span></span><span id="text-node-wrapper-363"><span id="text-node-363-45" data-byteoffset="45">]*</span><span id="text-node-363-47" data-byteoffset="47">Route</span><span id="text-node-363-52" data-byteoffset="52">), </span><span id="text-node-363-55" data-byteoffset="55">KeepContext</span><span id="text-node-363-66" data-byteoffset="66">: </span></span><span data-byteoffset="68" class="pl-c1"><span id="text-node-wrapper-386"><span id="text-node-386-68" data-byteoffset="68">false</span></span></span><span id="text-node-wrapper-391"><span id="text-node-391-73" data-byteoffset="73">}</span></span>`);
			});

			it("should detect comment node", () => {
				expect(annotations.isCommentNode(_codeCells[0].cell.firstChild)).to.be.true;
				expect(annotations.isCommentNode(_codeCells[6].cell.firstChild)).to.be.false;
			});

			it("should detect string node", () => {
				expect(annotations.isStringNode(_codeCells[7].cell.childNodes[1])).to.be.true;
				expect(annotations.isStringNode(_codeCells[0].cell.firstChild)).to.be.false;
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
			Array.from(github.getFileContainers()).forEach((file) => expect(github.getBlobElement(file)).to.be.ok);
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
				const baseCells = files.map((file) => github.getCodeCellsForAnnotation(github.getCodeTable(file), {isDelta: true, isSplitDiff: false, isBase: true}));
				const headCells = files.map((file) => github.getCodeCellsForAnnotation(github.getCodeTable(file), {isDelta: true, isSplitDiff: false, isBase: false}));

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

	describe("PR conversation view w/ snippets", () => {
		const url = "https://github.com/gorilla/mux/pull/190";
		before(setupDOM(url));

		it("should have 3 file containers", () => {
			expect(github.getFileContainers()).to.have.length(3);
		});

		it("should get blob from file containers", () => {
			Array.from(github.getFileContainers()).forEach((file) => expect(github.getBlobElement(file)).to.be.ok);
		});

		it("should not create blob annotator mounts", () => {
			Array.from(github.getFileContainers()).forEach((file) => expect(github.createBlobAnnotatorMount(file)).to.not.be.ok);
		});

		it("should get file names of containers", () => {
			expect(Array.from(github.getFileContainers()).map(github.getDeltaFileName)).to.eql([
				"mux_test.go", "mux.go", "mux.go",
			]);
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

		it("should not parse deltaInfo", () => {
			const deltaInfo = github.getDeltaInfo();
			expect(deltaInfo).to.have.property("baseBranch", "master");
			expect(deltaInfo).to.have.property("headBranch", "use-encoded-path-option");
			expect(deltaInfo).to.have.property("baseURI", "github.com/gorilla/mux");
			expect(deltaInfo).to.have.property("headURI", "github.com/kushmansingh/mux");
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

		it("should get blob from file containers", () => {
			Array.from(github.getFileContainers()).forEach((file) => expect(github.getBlobElement(file)).to.be.ok);
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
				const baseCells = files.map((file) => github.getCodeCellsForAnnotation(github.getCodeTable(file), {isDelta: true, isSplitDiff: false, isBase: true}));
				const headCells = files.map((file) => github.getCodeCellsForAnnotation(github.getCodeTable(file), {isDelta: true, isSplitDiff: false, isBase: false}));

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

			function getPRTestData(isBase: boolean): {annsByStartByte: annotations.AnnotationsByByte, lineStartBytes: annotations.StartBytesByLine, codeCells: github.CodeCell[]} {
				const anns = isBase ? testData.pullRequestBaseAnnotations.IncludedAnnotations : testData.pullRequestHeadAnnotations.IncludedAnnotations;
				const {annsByStartByte} = annotations.indexAnnotations(anns.Annotations);
				const lineStartBytes = annotations.indexLineStartBytes(anns.LineStartBytes);
				const file = github.getFileContainers()[0];
				const codeCells = github.getCodeCellsForAnnotation(github.getCodeTable(file), {isDelta: true, isSplitDiff: false, isBase});
				return {annsByStartByte, lineStartBytes, codeCells};
			}

			it("should convert deletion node (stress test)", () => {
				// first red line of mux.go
				const {annsByStartByte, lineStartBytes, codeCells} = getPRTestData(true);
				const convertedNode = annotations.convertElementNode(codeCells[0].cell, annsByStartByte, lineStartBytes[80], lineStartBytes[80], false);
				expect(convertedNode.bytesConsumed).to.eql(34); // TODO(john): number doesn't make sense
				expect(convertedNode.resultNode.textContent).to.eql("-\t\tpath := getPath(req)");
				expect(convertedNode.resultNode.innerHTML).to.eql(`<span id="text-node-wrapper-2136"><span id="text-node-2136-1" data-byteoffset="1">-\t</span><span id="text-node-2136-3" data-byteoffset="3">\t</span></span><span data-byteoffset="7" class="pl-smi"><span id="text-node-wrapper-2142"><span id="text-node-2142-7" data-byteoffset="7">path</span></span></span><span id="text-node-wrapper-2146"><span id="text-node-2146-11" data-byteoffset="11"> </span></span><span data-byteoffset="18" class="pl-k"><span id="text-node-wrapper-2153"><span id="text-node-2153-18" data-byteoffset="18">:</span><span id="text-node-2153-19" data-byteoffset="19">=</span></span></span><span id="text-node-wrapper-2157"><span id="text-node-2157-22" data-byteoffset="22"> </span></span><span data-byteoffset="23" class="pl-c1"><span id="text-node-wrapper-2158"><span id="text-node-2158-23" data-byteoffset="23">getPath</span></span></span><span id="text-node-wrapper-2165"><span id="text-node-2165-30" data-byteoffset="30">(req)</span></span>`);
			});

			it("should convert addition node (stress test)", () => {
				// first green line of mux.go
				const {annsByStartByte, lineStartBytes, codeCells} = getPRTestData(false);
				const convertedNode = annotations.convertElementNode(codeCells[4].cell, annsByStartByte, lineStartBytes[57], lineStartBytes[57], false);
				expect(convertedNode.bytesConsumed).to.eql(24); // TODO(john): number doesn't make sense
				expect(convertedNode.resultNode.textContent).to.eql("+\tuseEncodedPath bool");
				expect(convertedNode.resultNode.innerHTML).to.eql(`<span id="text-node-wrapper-1580"><span id="text-node-1580-1" data-byteoffset="1">+</span><span id="text-node-1580-2" data-byteoffset="2">\tuseEncodedPat</span><span id="text-node-1580-16" data-byteoffset="16">h</span><span id="text-node-1580-17" data-byteoffset="17"> </span></span><span data-byteoffset="21" class="pl-k"><span id="text-node-wrapper-1600"><span id="text-node-1600-21" data-byteoffset="21">bool</span></span></span>`);
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
			Array.from(github.getFileContainers()).forEach((file) => expect(github.getBlobElement(file)).to.be.ok);
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
				const baseCells = files.map((file) => github.getCodeCellsForAnnotation(github.getCodeTable(file), {isDelta: true, isSplitDiff: true, isBase: true}));
				const headCells = files.map((file) => github.getCodeCellsForAnnotation(github.getCodeTable(file), {isDelta: true, isSplitDiff: true, isBase: false}));

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
