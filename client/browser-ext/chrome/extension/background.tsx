import * as bluebird from "bluebird";
global.Promise = bluebird;

function promisifier(method: any): (...args: any[]) => Promise<any> {
	return (...args) => new Promise((resolve, reject) => {
		args.push(resolve);
		method.apply(this, args);
	});
}

function promisifyAll(obj: Object, list: string[]): void {
	list.forEach(api => bluebird.promisifyAll(obj[api], { promisifier }));
}

// let chrome extension api support Promise
promisifyAll(chrome, ["tabs"]);
promisifyAll(chrome.storage, ["local"]);

// These must be required after promisification completes.
// tslint:disable
require("./background/storage");
require("./background/tracker");
require("./background/inject");
