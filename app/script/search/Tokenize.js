exports.lastTermMightBecomeToken = function(str) {
	str = str.trim();
	if (!str) return true;
	var parts = str.split(/\s+/);
	var lastTerm = parts[parts.length - 1];
	return /[~$:@/?]/.test(lastTerm);
};
