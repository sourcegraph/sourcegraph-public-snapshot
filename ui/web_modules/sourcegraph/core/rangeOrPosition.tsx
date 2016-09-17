// RangeOrPosition represents a range or position.
export class RangeOrPosition {
	static fromZeroIndexed(startLine: number, startCol?: number, endLine?: number, endCol?: number): RangeOrPosition {
		const r = new RangeOrPosition();
		r.startLine = startLine;
		if ((endLine === undefined && endCol === undefined) || (typeof endLine === "number" && typeof endCol === "number")) {
			r.startCol = startCol;
		}

		// Consider line ranges such as "line 1 to line 1" (without
		// cols) as only having a start line and not an explicit end
		// line. Technically, the end col would be the last character
		// in the line, but we don't know its index here.
		if (endLine !== startLine || typeof endCol === "number") {
			r.endLine = endLine;
		}
		if (typeof endCol === "number" && typeof r.startCol === "number") {
			r.endCol = endCol;
		}
		return r;
	}

	static fromOneIndexed(startLine: number, startCol?: number, endLine?: number, endCol?: number): RangeOrPosition {
		return RangeOrPosition.fromZeroIndexed(
			startLine - 1,
			typeof startCol === "number" ? startCol - 1 : undefined,
			typeof endLine === "number" ? endLine - 1 : undefined,
			typeof endCol === "number" ? endCol - 1 : undefined
		);
	}

	static fromMonacoPosition(r: monaco.IPosition): RangeOrPosition {
		return RangeOrPosition.fromOneIndexed(r.lineNumber, r.column);
	}

	static fromMonacoRange(r: monaco.IRange): RangeOrPosition {
		return RangeOrPosition.fromOneIndexed(r.startLineNumber, r.startColumn, r.endLineNumber, r.endColumn);
	}

	// parse parses a string like "1-2", "1:2", "1:2-3", "1-2:3", or
	// "1:2-3:4". It assumes that line and column numbers are
	// 1-indexed.
	static parse(range: string): RangeOrPosition | null {
		let m = range.match(/^(\d+)(?::(\d+))?(?:-(\d+)(?::(\d+))?)?$/);
		if (m) {
			return RangeOrPosition.fromOneIndexed(
				parseInt(m[1], 10),
				typeof m[2] === "string" ? parseInt(m[2], 10) : undefined,
				typeof m[3] === "string" ? parseInt(m[3], 10) : undefined,
				typeof m[4] === "string" ? parseInt(m[4], 10) : undefined
			);
		}
		return null;
	}

	private startLine: number;
	private startCol?: number;
	private endLine?: number;
	private endCol?: number;

	zeroIndexed(): {startLine: number, startCol?: number, endLine?: number, endCol?: number} {
		return this.removeUndefined({
			startLine: this.startLine,
			startCol: this.startCol,
			endLine: this.endLine,
			endCol: this.endCol,
		});
	}

	oneIndexed(): {startLine: number, startCol?: number, endLine?: number, endCol?: number} {
		return this.removeUndefined({
			startLine: this.startLine + 1,
			startCol: typeof this.startCol === "number" ? this.startCol + 1 : undefined,
			endLine: typeof this.endLine === "number" ? this.endLine + 1 : undefined,
			endCol: typeof this.endCol === "number" ? this.endCol + 1 : undefined,
		});
	}

	toMonacoPosition(): monaco.IPosition {
		if (this.startCol === undefined) {
			throw new Error("converting to monaco position requires start column");
		}
		if (this.endLine !== undefined || this.endCol !== undefined) {
			throw new Error("can't convert range to position (without loss of information)");
		}
		return {
			lineNumber: this.startLine + 1,
			column: this.startCol + 1,
		};
	}

	toMonacoRange(): monaco.IRange {
		if (this.startCol === undefined) {
			throw new Error("converting to monaco range requires start column");
		}
		if (this.endLine === undefined) {
			throw new Error("converting to monaco range requires end line");
		}
		if (this.endCol === undefined) {
			throw new Error("converting to monaco range requires end column");
		}
		// We've guaranteed it won't be empty.
		return this.toMonacoRangeAllowEmpty();
	}

	toMonacoRangeAllowEmpty(): monaco.IRange {
		const startColumn = typeof this.startCol === "number" ? this.startCol + 1 : 1;
		let endColumn: number | undefined;
		if (typeof this.endCol === "number") {
			endColumn = this.endCol + 1;
		} else if (this.endLine === undefined || this.startLine === this.endLine) {
			endColumn = startColumn;
		} else {
			endColumn = 1;
		}
		return {
			startLineNumber: this.startLine + 1,
			startColumn,
			endLineNumber: typeof this.endLine === "number" ? this.endLine + 1 : this.startLine + 1,
			endColumn,
		};
	}

	// toString returns the string representation of the range or
	// position using 1-indexed numbers. The string representation can
	// be parsed using RangeOrPosition.parse.
	toString(): string {
		let s = (this.startLine + 1).toString();
		if (typeof this.startCol === "number") {
			s += `:${this.startCol + 1}`;
		}
		if (this.endLine !== this.startLine || this.endCol !== this.startCol) {
			if (typeof this.endLine === "number") {
				s += `-${this.endLine + 1}`;
			}
			if (typeof this.endCol === "number") {
				s += `:${this.endCol + 1}`;
			}
		}
		return s;
	}

	private removeUndefined(o: {startLine: number, startCol?: number, endLine?: number, endCol?: number}): {startLine: number, startCol?: number, endLine?: number, endCol?: number} {
		if (o.startCol === undefined) {
			delete o.startCol;
		}
		if (o.endLine === undefined) {
			delete o.endLine;
		}
		if (o.endCol === undefined) {
			delete o.endCol;
		}
		return o;
	}
}
