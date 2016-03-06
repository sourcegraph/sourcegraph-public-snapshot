export class WantFile {
	constructor(repo, rev, tree) {
		this.repo = repo;
		this.rev = rev;
		this.tree = tree;
	}
}

export class FileFetched {
	constructor(repo, rev, tree, file) {
		this.repo = repo;
		this.rev = rev;
		this.tree = tree;
		this.file = file;
	}
}

export class WantAnnotations {
	constructor(repo, rev, path, startByte, endByte) {
		this.repo = repo;
		this.rev = rev;
		this.path = path;
		this.startByte = startByte;
		this.endByte = endByte;
	}
}

export class AnnotationsFetched {
	constructor(repo, rev, path, startByte, endByte, annotations) {
		this.repo = repo;
		this.rev = rev;
		this.path = path;
		this.startByte = startByte;
		this.endByte = endByte;
		this.annotations = annotations;
	}
}

export class SelectLine {
	constructor(line) {
		this.line = line;
	}
}

export class SelectLineRange {
	constructor(line) {
		this.line = line;
	}
}

export class SelectCharRange {
	// startByte and endByte are absolute in the file. It is redundant to specify both
	// the line+col and the byte, but we need both in various places, and it's easier
	// to compute them and pass them together.
	//
	// If startLine is null, then all other fields' values are ignored and the action
	// is interpreted as "deselect the current selection."
	constructor(startLine, startCol, startByte, endLine, endCol, endByte) {
		this.startLine = startLine;
		this.startCol = startCol;
		this.startByte = startByte;
		this.endLine = endLine;
		this.endCol = endCol;
		this.endByte = endByte;
	}
}
