export default function() {
	if (global.it === undefined) {
		throw new Error("this function can only be used tests");
	}
}
