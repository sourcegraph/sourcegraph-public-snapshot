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
	constructor(startLine, startCol, endLine, endCol) {
		this.startLine = startLine;
		this.startCol = startCol;
		this.endLine = endLine;
		this.endCol = endCol;
	}
}
