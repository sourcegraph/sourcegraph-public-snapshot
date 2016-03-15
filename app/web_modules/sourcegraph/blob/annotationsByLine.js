// annotationsByLine returns an array with one entry per line. Each line's entry
// is the array of annotations that intersect that line.
//
// Assumes annotations has been sorted by sortAnns.
export default function annotationsByLine(lineStartBytes, annotations, lines) {
	const lineAnns = new Array(lineStartBytes.length);
	for (let i = 0; i < lineStartBytes.length; i++) {
		lineAnns[i] = [];
	}

	const lineEndBytes = lineStartBytes.map((lineStartByte, i) =>
		lineStartBytes[i + 1] || lineStartByte + lines[i].length
	);
	let line = 0; // 0-indexed line number
	for (let i = 0; i < annotations.length; i++) {
		const ann = annotations[i];

		// Advance (if necessary) to the first line that ann intersects.
		if (ann.StartByte >= lineEndBytes[line]) {
			while (line < lines.length && ann.StartByte >= lineEndBytes[line]) {
				line++;
			}
		}
		if (line === lines.length) break;

		// Add the ann to all lines (current and subsequent) it intersects.
		for (let line2 = line; line2 < lines.length; line2++) {
			if (ann.StartByte >= lineEndBytes[line2]) break;
			if (ann.StartByte < lineEndBytes[line2] && ann.EndByte >= lineStartBytes[line2]) {
				lineAnns[line2].push(ann);
			}
		}
	}
	return lineAnns;
}
