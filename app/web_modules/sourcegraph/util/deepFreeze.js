function deepFreeze(o) {
	if (process.env.NODE_ENV === "production" || Object.isFrozen(o)) {
		return o;
	}

	Object.freeze(o);
	Object.keys(o).forEach(function(prop) {
		deepFreeze(o[prop]);
	});
	return o;
}

export default deepFreeze;
