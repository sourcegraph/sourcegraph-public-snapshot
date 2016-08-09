export function deepFreeze<T>(o: T): T {
	if (process.env.NODE_ENV === "production" || Object.isFrozen(o)) {
		return o;
	}

	Object.freeze(o);
	Object.keys(o).forEach(function(prop: string): void {
		deepFreeze(o[prop]);
	});
	return o;
}

export function mergeAndDeepFreeze<T>(o1: T, o2: T): T {
	let o = Object.assign({} as T, o1, o2);
	return deepFreeze(o);
}
