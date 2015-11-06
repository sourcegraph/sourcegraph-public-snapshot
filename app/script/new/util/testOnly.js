export default function() {
	if (!global.it) {
		throw new Error("this function can only be used tests");
	}
}
