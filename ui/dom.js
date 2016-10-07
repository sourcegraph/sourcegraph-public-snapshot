var jsdom = require("jsdom");
require("source-map-support").install({
	environment: "node",
});

global.document = jsdom.jsdom('<!doctype html><html><body></body></html>');
global.window = global.document.defaultView;
global.self = global.window;

global.Element = global.window.Element;
global.HTMLElement = global.window.HTMLElement;
global.Node = global.window.Node;
global.navigator = global.window.navigator;
global.XMLHttpRequest = global.window.XMLHttpRequest;//

global.document.queryCommandSupported = function() { return false; };
