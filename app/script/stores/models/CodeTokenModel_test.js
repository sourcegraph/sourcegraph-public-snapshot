var expect = require("expect.js");

var globals = require("../../globals");
var CodeTokenModel = require("./CodeTokenModel");

describe("stores/models/CodeTokenModel", () => {
	it("should correctly parse server style payload for definitions", () => {
		var token = new CodeTokenModel({
			URL: "token_url",
			IsDef: true,
			Class: "cls",
			Label: "ABC",
		}, {parse: true});

		expect(token.attributes.syntax).to.be("cls");
		expect(token.attributes.html).to.be("ABC");
		expect(token.attributes.type).to.be(globals.TokenType.DEF);
		expect(token.attributes.url).to.be("token_url");
	});

	it("should correctly parse server style payload for references", () => {
		var token = new CodeTokenModel({
			URL: "token_url",
			IsDef: false,
			Class: "cls",
			Label: "ABC",
		}, {parse: true});

		expect(token.attributes.syntax).to.be("cls");
		expect(token.attributes.html).to.be("ABC");
		expect(token.attributes.type).to.be(globals.TokenType.REF);
		expect(token.attributes.url).to.be("token_url");
	});

	it("should correctly parse server style payload for text tokens", () => {
		var token = new CodeTokenModel({
			Class: "cls",
			Label: "ABC",
		}, {parse: true});

		expect(token.attributes.syntax).to.be("cls");
		expect(token.attributes.html).to.be("ABC");
		expect(token.attributes.type).to.be(globals.TokenType.SPAN);
	});

	it("should correctly parse server style payload for whitespace", () => {
		var token = new CodeTokenModel({Label: "   "}, {parse: true});

		expect(token.attributes.html).to.be("   ");
		expect(token.attributes.type).to.be(globals.TokenType.STRING);
	});
});
