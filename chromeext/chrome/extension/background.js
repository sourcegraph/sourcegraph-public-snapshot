const bluebird = require("bluebird");
global.Promise = bluebird;

function promisifier(method) {
	// return a function
	return function promisified(...args) {
		// which returns a promise
		return new Promise((resolve, reject) => {
			args.push(resolve);
			method.apply(this, args);
		});
	};
}

function promisifyAll(obj, list) {
	list.forEach(api => bluebird.promisifyAll(obj[api], {promisifier}));
}

// let chrome extension api support Promise
promisifyAll(chrome, [
	"tabs",
	"windows",
	"browserAction",
	"contextMenus"
]);
promisifyAll(chrome.storage, [
	"local",
]);

require("./background/contextMenus");
require("./background/inject");
require("./background/badge");
