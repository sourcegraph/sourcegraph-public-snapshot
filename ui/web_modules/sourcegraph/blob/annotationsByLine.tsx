// tslint:disable: typedef ordered-imports curly

// annotationsByLine returns an array with one entry per line. Each line's entry
// is the array of annotations that intersect that line.
//
// Assumes annotations has been sorted by sortAnns.
//
// NOTE: This must stay in sync with blob.go annotationsByLine.
export function annotationsByLine(lineStartBytes, annotations, lines) {
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

		// Optimization: add the ann to this line (if it intersects);
		if (ann.StartByte < lineEndBytes[line] && ann.EndByte >= lineStartBytes[line]) {
			lineAnns[line].push(ann);
		}

		// Add the ann to all lines (current and subsequent) it intersects.
		if (ann.EndByte <= lineEndBytes[line]) continue;
		for (let line2 = line + 1; line2 < lines.length; line2++) {
			if (ann.StartByte >= lineEndBytes[line2]) break;
			if (ann.StartByte < lineEndBytes[line2] && ann.EndByte >= lineStartBytes[line2]) {
				lineAnns[line2].push(ann);
			}
		}
	}
	return lineAnns;
}
