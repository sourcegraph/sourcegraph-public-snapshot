var traceData = {};

/**
 * @description Starts a timer with the given name.
 * @param {string} name - Timer name.
 * @returns {void}
 */
module.exports.start = function(name) {
	traceData[name] = new Date();
	console.info("[" + name + "]: started");
};

/**
 * @description Ends the timer with the given name and reports time elapsed.
 * @param {string} name - Timer to stop.
 * @returns {void}
 */
module.exports.end = function(name) {
	var elapsed = traceData[name] ? (new Date() - traceData[name]) / 1000 : "never started";
	console.info("[" + name + "]: ended (" + elapsed + "s)");
	Reflect.deleteProperty(traceData, name);
};
