var jsdom = require("jsdom").jsdom;
var sinon = require("sinon");
var expect = require("expect.js");

global.document = jsdom("<html><body></body></html>");
global.window = document.defaultView;
global.navigator = {
	userAgent: "test",
};
var ReactDOM = require("react-dom");
var TestUtils = require("react-addons-test-utils");

var sandbox = sinon.sandbox.create();
sandbox.reactContainers = [];
sandbox.renderComponent = function(instance, container) {
	container = container || global.document.createElement("div");
	sandbox.reactContainers.push(container);
	return ReactDOM.render(instance, container);
};
sandbox.renderAndExpect = function(instance) {
	var renderer = TestUtils.createRenderer();
	renderer.render(instance);
	return expect(renderer.getRenderOutput());
};

var Dispatcher = require("./new/Dispatcher");
sandbox.dispatched = [];
Dispatcher.dispatch = function(action) {
	sandbox.dispatched.push(action);
};

afterEach(function() {
	sandbox.dispatched = [];

	sandbox.reactContainers.forEach(ReactDOM.unmountComponentAtNode);
	sandbox.reactContainers = [];

	sandbox.restore();
});

module.exports = sandbox;
