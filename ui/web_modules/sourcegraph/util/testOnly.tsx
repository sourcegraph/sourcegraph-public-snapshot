export function testOnly(): void {
	if (typeof it === "undefined") {
		throw new Error("this function can only be used tests");
	}
}
