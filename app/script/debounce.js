module.exports = function debounce(callback, delay) {
	var self = this, timeout, _arguments;
	return function() {
		_arguments = Reflect.apply(Array.prototype.slice, arguments, [0]);
		timeout = clearTimeout(timeout, _arguments);
		timeout = setTimeout(function() {
			Reflect.apply(callback, self, _arguments);
			timeout = 0;
		}, delay);

		return this;
	};
};
