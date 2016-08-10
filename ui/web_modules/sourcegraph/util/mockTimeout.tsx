// tslint:disable: typedef ordered-imports curly

export function mockTimeout(f) {
	let orig = global.setTimeout;
	let callbacks: any[] = [];
	global.setTimeout = function(callback, delay) {
		callbacks.push(callback);
	};
	try {
		f();
	} finally {
		while (callbacks.length > 0) {
			callbacks.shift()();
		}
		global.setTimeout = orig;
	}
}
