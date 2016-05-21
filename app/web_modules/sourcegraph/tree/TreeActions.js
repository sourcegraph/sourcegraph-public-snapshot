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
	constructor(repo, commitID) {
		this.repo = repo;
		this.commitID = commitID;
	}
}

export class FileListFetched {
	constructor(repo, commitID, fileList) {
		this.repo = repo;
		this.commitID = commitID;
		this.fileList = fileList;
	}
}

export class WantSrclibDataVersion {
	constructor(repo, commitID, pathOrNull) {
		this.repo = repo;
		this.commitID = commitID;
		this.path = pathOrNull || null;
	}
}

export class FetchedSrclibDataVersion {
	constructor(repo, commitID, pathOrNull, versionOrNull) {
		this.repo = repo;
		this.commitID = commitID;
		this.path = pathOrNull || null;
		this.version = versionOrNull || null;
	}
}
