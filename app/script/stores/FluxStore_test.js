var sandbox = require("../testSandbox");
var expect = require("expect.js");

var FluxStore = require("./FluxStore");
var testDispatcher = new (require("flux").Dispatcher)();

var globals = require("../globals");
globals.Actions.ACTION_A = "ACTION_A";
globals.Actions.ACTION_B = "ACTION_B";
globals.Actions.ACTION_C = "ACTION_C";

describe("stores/FluxStore", () => {
	function getStoreWithSetup(obj) {
		obj.dispatcher = testDispatcher;
		return FluxStore(obj);
	}

	it("should let user know if any actions are non-existent", () => {
		sandbox.stub(console, "warn");

		getStoreWithSetup({
			actions: {
				ACTION_A: "noop",
				ACTION_B: "noop",
				ACTION_X: "noop",
			},
			noop() {},
		});

		expect(console.warn.callCount).to.be(1);
		expect(console.warn.firstCall.args[0]).to.contain("ACTION_X");
	});

	it("should let users know if any registered methods are inexistent", () => {
		sandbox.stub(console, "warn");

		getStoreWithSetup({
			actions: {
				ACTION_A: "noop",
				ACTION_B: "undefined1",
				ACTION_C: "undefined2",
			},
			noop() {},
		});

		expect(console.warn.callCount).to.be(2);
		expect(console.warn.firstCall.args[0]).to.contain("undefined1");
		expect(console.warn.secondCall.args[0]).to.contain("undefined2");
	});

	it("should dispatch to the correct callback and pass action along with source", () => {
		var store = getStoreWithSetup({
			actions: {ACTION_A: "callback"},
			callback: sandbox.stub(),
		});

		var actionPayload = {
			type: "ACTION_A",
			params: 123,
		};

		testDispatcher.dispatch({
			source: "source",
			action: actionPayload,
		});

		expect(store.callback.callCount).to.be(1);
		expect(store.callback.firstCall.args[0]).to.be(actionPayload);
		expect(store.callback.firstCall.args[1]).to.be("source");
	});

	it("should also call parent initialize", () => {
		var init = sandbox.stub();
		getStoreWithSetup({initialize: init});

		expect(init.callCount).to.be(1);
	});

	it("should also call parent destroy and unregister from dispatcher", () => {
		var store = getStoreWithSetup({destroy: sandbox.stub()});

		testDispatcher.unregister = sandbox.stub();
		store.destroy();

		expect(testDispatcher.unregister.callCount).to.be(1);
	});
});
