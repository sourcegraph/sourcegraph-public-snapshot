var sandbox = require("../testSandbox");
var expect = require("expect.js");

var AppDispatcher = require("../dispatchers/AppDispatcher");
var TokenPopupStore = require("./TokenPopupStore");
var globals = require("../globals");
require("../routing/CodeFileRouter");

describe("stores/TokenPopupStore", () => {
	it("should correctly set error, closed and data while triggering only 1 change event on RECEIVED_POPUP", () => {
		var changeFn = sandbox.stub();
		var data = {a: 2};

		TokenPopupStore.on("change", changeFn);

		AppDispatcher.handleServerAction({
			type: globals.Actions.RECEIVED_POPUP,
			data: data,
		});

		expect(changeFn.callCount).to.be(1);
		expect(TokenPopupStore.get("error")).to.be(false);
		expect(TokenPopupStore.get("closed")).to.be(false);
		expect(TokenPopupStore.get("a")).to.be(2);
	});

	xit("should set closed to true, type to REF correctly on TOKEN_SELECT", () => {
		AppDispatcher.handleViewAction({
			type: globals.Actions.TOKEN_SELECT,
			token: null,
		});

		expect(TokenPopupStore.get("closed")).to.be(true);
		expect(TokenPopupStore.get("type")).to.be(globals.TokenType.REF);
	});

	xit("should set popup type to DEF when receiving SHOW_DEFINITION action", () => {
		AppDispatcher.handleViewAction({
			type: globals.Actions.SHOW_DEFINITION,
		});

		expect(TokenPopupStore.get("type")).to.be(globals.TokenType.DEF);
	});
});
