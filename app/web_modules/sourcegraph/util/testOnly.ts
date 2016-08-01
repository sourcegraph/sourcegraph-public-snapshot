export default function(): void {
	if (!global.it) {
		throw new Error("this function can only be used tests");
	}
}
