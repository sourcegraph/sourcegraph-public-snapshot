var jsdom = require("jsdom").jsdom;
var sinon = require("sinon");

global.document = jsdom("<html><body></body></html>");
global.window = document.parentWindow;
global.navigator = {
	userAgent: "test",
};
var React = require("react");

var sandbox = sinon.sandbox.create();

sandbox.reactContainers = [];
sandbox.renderComponent = function(instance) {
	var container = global.document.createElement("div");
	sandbox.reactContainers.push(container);
	return React.render(instance, container);
};

afterEach(function() {
	sandbox.reactContainers.forEach(React.unmountComponentAtNode);
	sandbox.reactContainers = [];

	sandbox.restore();
});

module.exports = sandbox;
