describe("test-case", function() {
	it("should run", function(done) {
		setTimeout(done, 1000);
	});
	it("should fail", function(done) {
		setTimeout(function() {
			throw new Error("Fail");
		}, 1000);
	});
	it("should randomly fail", function() {
		if(require("./dep"))
			throw new Error("Random fail");
	});
});