var $ = require("jquery");

$(function() {
	var historyMethods = ["back", "forward", "go", "pushState", "replaceState"];
	historyMethods.forEach(function(method) {
		var orig = window.history[method];
		window.history[method] = function() {
			Reflect.apply(orig, this, arguments);
			window.dispatchEvent(new Event("sg:"+method));
		};
	});

	var onpopstate_ = window.onpopstate;
	window.onpopstate = function() {
		if (onpopstate_) {
			Reflect.apply(onpopstate_, this, arguments);
		}
		window.dispatchEvent(new Event("sg:popState"));
	};
});
