// tslint:disable

export function lineCol(line: number | string, col?: number): string {
	if (typeof col === "undefined") {
		return line.toString();
	}
	return `${line}:${col}`;
}

export function lineRange(startLineCol: string | number, endLineCol?: string | number): string {
	if (typeof endLineCol === "undefined" || startLineCol === endLineCol) {
		return startLineCol.toString();
	}
	return `${startLineCol}-${endLineCol}`;
}

export function parseLineRange(range: string): {startLine: number, startCol?: number, endLine: number, endCol?: number} | null {
	let lineMatch = range.match(/^(\d+)(?:-(\d+))?$/);
	if (lineMatch) {
		return {
			startLine: parseInt(lineMatch[1], 10),
			endLine: parseInt(lineMatch[2] || lineMatch[1], 10),
		};
	}
	let lineColMatch = range.match(/^(\d+):(\d+)-(\d+):(\d+)$/);
	if (lineColMatch) {
		return {
			startLine: parseInt(lineColMatch[1], 10),
			startCol: parseInt(lineColMatch[2], 10),
			endLine: parseInt(lineColMatch[3], 10),
			endCol: parseInt(lineColMatch[4], 10),
		};
	}
	return null;
}
