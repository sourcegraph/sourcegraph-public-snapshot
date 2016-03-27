export class WantCommit {
	constructor(repo, rev, path) {
		this.repo = repo;
		this.rev = rev;
		this.path = path;
	}
}

export class CommitFetched {
	constructor(repo, rev, path, commit) {
		this.repo = repo;
		this.rev = rev;
		this.path = path;
		this.commit = commit;
	}
}

export class WantFileList {
	constructor(repo, rev, commitID) {
		this.repo = repo;
		this.rev = rev;
		this.commitID = commitID;
	}
}

export class FileListFetched {
	constructor(repo, rev, commitID, fileList) {
		this.repo = repo;
		this.rev = rev;
		this.commitID = commitID;
		this.fileList = fileList;
	}
}

export class WantSrclibDataVersion {
	constructor(repo, rev, commitID, pathOrNull) {
		this.repo = repo;
		this.rev = rev;
		this.commitID = commitID;
		this.path = pathOrNull || null;
	}
}

export class FetchedSrclibDataVersion {
	constructor(repo, rev, commitID, pathOrNull, versionOrNull) {
		this.repo = repo;
		this.rev = rev;
		this.commitID = commitID;
		this.path = pathOrNull || null;
		this.version = versionOrNull || null;
	}
}

export class UpDirectory {
	constructor() {}
}

export class DownDirectory {
	constructor(part) {
		this.part = part;
	}
}
