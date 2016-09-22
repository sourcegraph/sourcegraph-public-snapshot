// tslint:disable typedef ordered-imports
interface Position {
	line: number;
	character: number;
}

export function toPosition(pos: monaco.IPosition): Position {
	return {line: pos.lineNumber - 1, character: pos.column - 1};
}

interface Range {
	start: Position;
	end: Position;
}

export interface Location {
	uri: string;
	range: Range;
}

export function toMonacoLocation(loc: Location): monaco.languages.Location {
	return {
		uri: monaco.Uri.parse(loc.uri),
		range: toMonacoRange(loc.range),
	};
}

export function toMonacoRange(r: Range): monaco.IRange {
	return new monaco.Range(r.start.line + 1, r.start.character + 1, r.end.line + 1, r.end.character + 1);
}
