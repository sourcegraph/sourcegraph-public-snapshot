var sandbox = require("../../testSandbox");
var expect = require("expect.js");

var HunkModel = require("./HunkModel");

describe("stores/models/HunkModel", () => {
	it("should warn user if initialized without the parse option", () => {
		sandbox.stub(console, "warn");
		new HunkModel(); // eslint-disable-line no-new
		expect(console.warn.callCount).to.be(1);
	});

	it("should throw an error if there are more line prefixes than lines", () => {
		expect(() => {
			new HunkModel({ // eslint-disable-line no-new
				LinePrefixes: "+ -- +",
				BodySource: {
					Lines: [1, 2, 3, 4],
				},
			}, {parse: true});
		}).to.throwException();
	});
});
