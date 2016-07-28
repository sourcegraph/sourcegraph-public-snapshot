function deepFreeze<T>(o: T): T {
	if (process.env.NODE_ENV === "production" || Object.isFrozen(o)) {
		return o;
	}

	Object.freeze(o);
	Object.keys(o).forEach(function(prop: string): void {
		deepFreeze(o[prop]);
	});
	return o;
}

export default deepFreeze;
