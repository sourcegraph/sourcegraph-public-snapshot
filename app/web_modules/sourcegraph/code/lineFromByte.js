export default function(contents, bytePos) {
	let lines = typeof contents === "string" ? contents.split("\n") : contents;
	let pos = 0;
	for (let i = 0; i < lines.length; i++) {
		let endPos = pos + lines[i].length + 1; // add 1 to account for newline (stripped by split)
		if (bytePos >= pos && bytePos < endPos) {
			return i + 1; // line numbers start with 1
		}
		pos = endPos;
	}
	throw new Error(`Byte ${bytePos} is out of bounds (file length: ${contents.length})`);
}
