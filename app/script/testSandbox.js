var jsdom = require("jsdom").jsdom;
var sinon = require("sinon");

global.document = jsdom("<html><body></body></html>");
global.window = document.defaultView;
global.navigator = {
	userAgent: "test",
};
var ReactDOM = require("react-dom");

var sandbox = sinon.sandbox.create();
sandbox.reactContainers = [];
sandbox.renderComponent = function(instance, container) {
	container = container || global.document.createElement("div");
	sandbox.reactContainers.push(container);
	return ReactDOM.render(instance, container);
};

afterEach(function() {
	sandbox.reactContainers.forEach(ReactDOM.unmountComponentAtNode);
	sandbox.reactContainers = [];

	sandbox.restore();
});

module.exports = sandbox;
