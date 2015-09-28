var expect = require("expect.js");

var CodeTokenCollection = require("./CodeTokenCollection");
var CodeTokenModel = require("../models/CodeTokenModel");
var globals = require("../../globals");

describe("stores/collections/CodeTokenCollection", () => {
	function getCollectionWithTokens(list) {
		var c = new CodeTokenCollection();
		if (Array.isArray(list)) {
			list.forEach(t => c.put(t));
		}
		return c;
	}

	it("should correctly register a new token of type REF in a new group", () => {
		var token = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.REF,
		});
		var coll = getCollectionWithTokens([token]);

		var item = coll.get("url_a");

		expect(item.attributes.refs).not.to.be(false);

		expect(item.get("def").length).to.be(0);
		expect(item.get("refs").length).to.be(1);
		expect(item.get("refs")[0]).to.be(token);
	});

	it("should correctly register a new token of type DEF in a new group", () => {
		var token = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.DEF,
		});
		var coll = getCollectionWithTokens([token]);

		var item = coll.get("url_a");

		expect(item).not.to.be(false);
		expect(item.attributes.refs).not.to.be(false);
		expect(item.attributes.def).not.to.be(null);

		expect(item.get("refs").length).to.be(0);
		expect(item.get("def").length).to.be(1);
		expect(item.get("def")[0]).to.be(token);
	});

	it("should correctly add incremental token ID property .tid", () => {
		var t1 = new CodeTokenModel();
		var t2 = new CodeTokenModel();
		var coll = getCollectionWithTokens([t1, t2]);

		expect(t1.tid).to.be(0);
		expect(t2.tid).to.be(1);
		expect(coll.byId(0)).to.be(t1);
		expect(coll.byId(1)).to.be(t2);
	});

	it("should correctly add DEFs and REFs to already existing groups by concatenating", () => {
		var t1 = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.DEF,
		});
		var t2 = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.REF,
		});
		var t3 = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.REF,
		});

		var coll = getCollectionWithTokens([t1, t2, t3]);
		var item = coll.get("url_a");

		expect(item.get("def").length).to.be(1);
		expect(item.get("def")[0]).to.be(t1);

		expect(item.get("refs").length).to.be(2);
		expect(item.get("refs")[0]).to.be(t2);
		expect(item.get("refs")[1]).to.be(t3);
	});

	it("should correctly obtain definitions", () => {
		var t1 = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.DEF,
		});
		var t2 = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.REF,
		});
		var t3 = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.REF,
		});
		var t4 = new CodeTokenModel({
			url: ["url_b"],
			type: globals.TokenType.REF,
		});

		var coll = getCollectionWithTokens([t1, t2, t3, t4]);

		expect(coll.getDefinition("url_a").length).to.be(1);
		expect(coll.getDefinition("url_a")[0]).to.be(t1);
	});

	it("should correctly select a group of tokens", () => {
		var t1 = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.DEF,
		});
		var t2 = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.REF,
		});
		var t3 = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.REF,
		});
		var t4 = new CodeTokenModel({
			url: ["url_b"],
			type: globals.TokenType.REF,
		});

		var coll = getCollectionWithTokens([t1, t2, t3, t4]);

		coll.select("url_a");

		expect(t1.get("selected")).to.be(true);
		expect(t2.get("selected")).to.be(true);
		expect(t3.get("selected")).to.be(true);
		expect(t4.get("selected")).to.be(undefined);
	});

	it("should correctly clear selected groups", () => {
		var t1 = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.DEF,
		});
		var t2 = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.REF,
		});
		var t3 = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.REF,
		});
		var t4 = new CodeTokenModel({
			url: ["url_b"],
			type: globals.TokenType.REF,
		});

		var coll = getCollectionWithTokens([t1, t2, t3, t4]);

		coll.select("url_a");
		coll.clearSelected();

		expect(t1.get("selected")).to.be(false);
		expect(t2.get("selected")).to.be(false);
		expect(t3.get("selected")).to.be(false);
		expect(t4.get("selected")).to.be(undefined);
	});

	it("should correctly highlight groups of tokens", () => {
		var t1 = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.DEF,
		});
		var t2 = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.REF,
		});
		var t3 = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.REF,
		});
		var t4 = new CodeTokenModel({
			url: ["url_b"],
			type: globals.TokenType.REF,
		});

		var coll = getCollectionWithTokens([t1, t2, t3, t4]);

		coll.highlight("url_a");

		expect(t1.get("highlighted")).to.be(true);
		expect(t2.get("highlighted")).to.be(true);
		expect(t3.get("highlighted")).to.be(true);
		expect(t4.get("highlighted")).to.be(undefined);
	});

	it("should correctly clear highlights", () => {
		var t1 = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.DEF,
		});
		var t2 = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.REF,
		});
		var t3 = new CodeTokenModel({
			url: ["url_a"],
			type: globals.TokenType.REF,
		});
		var t4 = new CodeTokenModel({
			url: ["url_b"],
			type: globals.TokenType.REF,
		});

		var coll = getCollectionWithTokens([t1, t2, t3, t4]);

		coll.highlight("url_a");
		coll.clearHighlighted();

		expect(t1.get("highlighted")).to.be(false);
		expect(t2.get("highlighted")).to.be(false);
		expect(t3.get("highlighted")).to.be(false);
		expect(t4.get("highlighted")).to.be(undefined);
	});
});
