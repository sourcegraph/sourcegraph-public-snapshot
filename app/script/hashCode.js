/**
 * @description Hash code generates a unique hash code of the string.
 * @param {string} str - The string to hash.
 * @returns {string} The generated hash code.
 */
var hashCode = function(str) {
	var hash = 0, i, chr, len;
	if (str.length === 0) return hash;
	for (i = 0, len = str.length; i < len; i++) {
		chr = str.charCodeAt(i);
		hash = ((hash << 5) - hash) + chr;
		hash |= 0; // Convert to 32bit integer
	}
	return hash;
};

module.exports = hashCode;
