exports.base = function(uri) {
	var parts = uri.split("/");
	return parts[parts.length-1];
};

exports.label = function(uri) {
	return uri.replace(/^(github|sourcegraph).com\//, "");
};
