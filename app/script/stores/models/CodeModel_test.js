var sandbox = require("../../testSandbox");
var expect = require("expect.js");

var CodeModel = require("./CodeModel");

describe("stores/models/CodeModel", () => {
	function getSourceCodeMock() {
		var token = [
			"whitespace",
			{
				URL: ["url_1"],
				IsDef: true,
				Class: "cls_1",
				Label: "ABC_1",
			}, {
				URL: ["url_2"],
				IsDef: false,
				Class: "cls_2",
				Label: "ABC_2",
			}, {
				Class: "pun_1",
				Label: "DEF_1",
			},
			"whitespace",
			{
				Class: "pun_2",
				Label: "DEF_2",
			},
		];

		return {
			entry: {
				StartLine: 1,
				SourceCode: {
					Lines: [
						{
							StartByte: 0,
							EndByte: 5,
							Tokens: [token[0], token[1], token[2]],
						}, {
							StartByte: 5,
							EndByte: 6,
						}, {
							StartByte: 6,
							EndByte: 10,
							Tokens: [token[3], token[4], token[5]],
						},
					],
				},
			},
			entryTokens: token,
		};
	}

	xit("should correctly create line and token collections from server payload", () => {
		var model = new CodeModel();
		var serverMock = getSourceCodeMock();

		model.load(serverMock.entry);

		expect(model.tokens).not.to.be(null);
		expect(model.get("lines").length).to.be(3);

		serverMock.entry.SourceCode.Lines.forEach((line, k) => {
			var codeLine = model.get("lines").at(k);

			expect(codeLine.get("start")).to.eql(line.StartByte);
			expect(codeLine.get("end")).to.eql(line.EndByte);
			expect(codeLine.get("number")).to.eql(serverMock.entry.StartLine+k);
			expect(codeLine.get("tokens").length).to.be(line.Tokens ? line.Tokens.length : 0);

			var lineTokens = codeLine.get("tokens");

			(line.Tokens || []).forEach((token, j) => {
				var t = lineTokens(j);
				expect(t.get("syntax")).to.be(token.Class);
				expect(t.get("html")).to.be(token === "string" ? token : token.Label);
				expect(t.get("url")).to.be(token.URL);
			});
		});

		expect(model.tokens.length).to.be(2);
	});

	it("should propagate getDefinition to TokenCollection", () => {
		var model = new CodeModel();
		var serverMock = getSourceCodeMock();

		model.load(serverMock.entry);
		sandbox.spy(model.tokens, "getDefinition");
		model.getDefinition("url_2");

		expect(model.tokens.getDefinition.callCount).to.be(1);
		expect(model.tokens.getDefinition.firstCall.args[0]).to.be("url_2");
	});

	it("should propagate line highlighting to line collection property", () => {
		var model = new CodeModel(),
			serverMock = getSourceCodeMock();

		model.load(serverMock.entry);
		sandbox.spy(model.get("lines"), "highlightRange");
		model.highlightLineRange(1, 2);

		expect(model.get("lines").highlightRange.callCount).to.be(1);
		expect(model.get("lines").highlightRange.firstCall.args[0]).to.be(1);
		expect(model.get("lines").highlightRange.firstCall.args[1]).to.be(2);
	});

	it("should propagate byte highlighting to line collection property", () => {
		var model = new CodeModel();
		var serverMock = getSourceCodeMock();

		model.load(serverMock.entry);
		sandbox.spy(model.get("lines"), "highlightByteRange");
		model.highlightByteRange(1, 2);

		expect(model.get("lines").highlightByteRange.callCount).to.be(1);
		expect(model.get("lines").highlightByteRange.firstCall.args[0]).to.be(1);
		expect(model.get("lines").highlightByteRange.firstCall.args[1]).to.be(2);
	});

	it("should propagate highlight clearing for tokens to tokens collection", () => {
		var model = new CodeModel();
		var serverMock = getSourceCodeMock();

		model.load(serverMock.entry);
		sandbox.spy(model.tokens, "clearHighlighted");
		model.clearHighlightedTokens();

		expect(model.tokens.clearHighlighted.callCount).to.be(1);
	});

	it("should propagate selection clearing for tokens to tokens collection", () => {
		var model = new CodeModel();
		var serverMock = getSourceCodeMock();

		model.load(serverMock.entry);
		sandbox.spy(model.tokens, "clearSelected");
		model.clearSelectedTokens();

		expect(model.tokens.clearSelected.callCount).to.be(1);
	});

	it("should propagate highlight clearing for lines to line collection property", () => {
		var model = new CodeModel();
		var serverMock = getSourceCodeMock();

		model.load(serverMock.entry);
		sandbox.spy(model.get("lines"), "clearHighlighted");
		model.clearHighlightedLines();

		expect(model.get("lines").clearHighlighted.callCount).to.be(1);
	});

	it("should propagate token highlight to token collection", () => {
		var model = new CodeModel();
		var serverMock = getSourceCodeMock();

		model.load(serverMock.entry);
		sandbox.spy(model.tokens, "highlight");
		model.highlightToken("URL");

		expect(model.tokens.highlight.callCount).to.be(1);
		expect(model.tokens.highlight.firstCall.args[0]).to.be("URL");
	});

	it("should propagate token selection to token collection", () => {
		var model = new CodeModel();
		var serverMock = getSourceCodeMock();

		model.load(serverMock.entry);
		sandbox.spy(model.tokens, "select");
		model.selectToken("URL");

		expect(model.tokens.select.callCount).to.be(1);
		expect(model.tokens.select.firstCall.args[0]).to.be("URL");
	});
});
