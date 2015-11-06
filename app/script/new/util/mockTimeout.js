export default function(f) {
	let orig = global.setTimeout;
	let callbacks = [];
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
