// @flow

export class WantFile {
	repo: string;
	commitID: ?string;
	path: string;

	constructor(repo: string, commitID: ?string, path: string) {
		this.repo = repo;
		this.commitID = commitID;
		this.path = path;
	}
}

export class FileFetched {
	repo: string;
	commitID: ?string;
	path: string;
	file: any;
	eventName: string;

	constructor(repo: string, commitID: ?string, path: string, file: any) {
		this.repo = repo;
		this.commitID = commitID;
		this.path = path;
		this.file = file;
		this.eventName = "FileFetched";
	}
}

export class WantAnnotations {
	repo: string;
	commitID: string;
	path: string;
	startByte: ?number;
	endByte: ?number;

	constructor(repo: string, commitID: string, path: string, startByte: ?number, endByte: ?number) {
		this.repo = repo;
		this.commitID = commitID;
		this.path = path;
		this.startByte = startByte;
		this.endByte = endByte;
	}
}

export class AnnotationsFetched {
	repo: string;
	commitID: string;
	path: string;
	startByte: ?number;
	endByte: ?number;
	annotations: any;

	constructor(repo: string, commitID: string, path: string, startByte: ?number, endByte: ?number, annotations: any) {
		this.repo = repo;
		this.commitID = commitID;
		this.path = path;
		this.startByte = startByte;
		this.endByte = endByte;
		this.annotations = annotations;
	}
}

export class SelectLine {
	repo: string;
	rev: string;
	path: string;
	line: ?number;
	eventName: string;

	constructor(repo: string, rev: string, path: string, line: ?number) {
		this.repo = repo;
		this.rev = rev;
		this.path = path;
		this.line = line;
		this.eventName = "SelectLine";
	}
}

export class SelectLineRange {
	repo: string;
	rev: string;
	path: string;
	line: number;
	eventName: string;

	constructor(repo: string, rev: string, path: string, line: number) {
		this.repo = repo;
		this.rev = rev;
		this.path = path;
		this.line = line;
		this.eventName = "SelectLineRange";
	}
}

export class SelectCharRange {
	repo: string;
	rev: string;
	path: string;
	startLine: ?number;
	startCol: ?number;
	startByte: ?number;
	endLine: ?number;
	endCol: ?number;
	endByte: ?number;
	eventName: string;

	// startByte and endByte are absolute in the file. It is redundant to specify both
	// the line+col and the byte, but we need both in various places, and it's easier
	// to compute them and pass them together.
	//
	// If startLine is null, then all other fields' values are ignored and the action
	// is interpreted as "deselect the current selection."
	constructor(repo: string, rev: string, path: string, startLine: ?number, startCol: ?number, startByte: ?number, endLine: ?number, endCol: ?number, endByte: ?number) {
		this.repo = repo;
		this.rev = rev;
		this.path = path;
		this.startLine = startLine;
		this.startCol = startCol;
		this.startByte = startByte;
		this.endLine = endLine;
		this.endCol = endCol;
		this.endByte = endByte;
		this.eventName = "SelectCharRange";
	}
}
