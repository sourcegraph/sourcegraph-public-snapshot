// tslint:disable: typedef ordered-imports

export function mockTimeout(f) {
	let orig = (global as any).setTimeout;
	let callbacks: any[] = [];
	(global as any).setTimeout = function(callback, delay) {
		callbacks.push(callback);
	};
	try {
		f();
	} finally {
		while (callbacks.length > 0) {
			callbacks.shift()();
		}
		(global as any).setTimeout = orig;
	}
}
