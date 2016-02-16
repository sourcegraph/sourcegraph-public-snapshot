// fileLines returns the lines of the file.
export default function fileLines(contents) {
	let lines = contents.split("\n");
	if (lines.length > 0 && lines[lines.length - 1] === "") {
		lines.pop();
	}
	return lines;
}
