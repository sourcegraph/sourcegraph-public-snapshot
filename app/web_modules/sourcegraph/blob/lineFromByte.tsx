// tslint:disable

import * as utf8 from "utf8";

export default function(contents: string | string[], bytePos: number) {
	const lines = typeof contents === "string" ? contents.split("\n") : contents;
	let pos = 0;
	for (let i = 0; i < lines.length; i++) {
			// Encode the line using utf8 to account for multi-byte unicode characters.
		let endPos = pos + utf8.encode(lines[i]).length + 1; // add 1 to account for newline (stripped by split)
		if (bytePos >= pos && bytePos < endPos) {
			return i + 1; // line numbers start with 1
		}
		pos = endPos;
	}
	throw new Error(`Byte ${bytePos} is out of bounds (file length: ${contents.length})`);
}

// createLineFromByteFunc returns a function that can quickly compute the line
// for any byte, for the single file's contents that are specified.
export function createLineFromByteFunc(contents: string): (byte: number) => number {
	const lsb = computeLineStartBytes(contents.split("\n"));
	return (byte: number) => {
		if (lsb.length === 1) return 1;
		for (let i = 1; i < lsb.length; i++) {
			if (byte < lsb[i]) return i;
		}
		if (byte <= contents.length) return lsb.length;
		throw new Error(`Byte ${byte} is out of bounds`);
	};
}

export function computeLineStartBytes(lines: string[]): number[] {
	let pos: number = 0;
	return lines.map((line) => {
		let start = pos;
		// Encode the line using utf8 to account for multi-byte unicode characters.
		pos += utf8.encode(line).length + 1; // add 1 to account for newline
		return start;
	});
}
