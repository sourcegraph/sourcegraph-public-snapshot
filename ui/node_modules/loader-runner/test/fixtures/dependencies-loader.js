module.exports = function(source) {
	this.clearDependencies();
	this.addDependency("a");
	this.addDependency("b");
	this.addContextDependency("c");
	return source;
};
